//go:build js && wasm

/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Controller Matching Wizard - WebAssembly build.
// Build with: GOOS=js GOARCH=wasm go build -o site-src/wizard/main.wasm ./wasm/
// Load from HTML: wasm_exec.js + instantiateStreaming(fetch("main.wasm"), go.importObject).then(r => go.run(r.instance))
package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"syscall/js"
)

type featureDef struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type implementation struct {
	Organization string   `json:"organization"`
	Project      string   `json:"project"`
	URL          string   `json:"url"`
	Version      string   `json:"version"`
	ReportDate   string   `json:"reportDate"`
	Conformance  []string `json:"conformance"`
	Features     []string `json:"features"`
}

type wizardData struct {
	FeatureDefinitions struct {
		HTTPGateway    []featureDef `json:"httpGateway"`
		HTTPRoute      []featureDef `json:"httpRoute"`
		HTTPBackendTLS []featureDef `json:"httpBackendTls"`
		GRPC           []featureDef `json:"grpc"`
		TLS            []featureDef `json:"tls"`
	} `json:"featureDefinitions"`
	Implementations []implementation `json:"implementations"`
	// Version-keyed implementations when multi-version (e.g. v1.5.0, v1.4.0 -> []implementation)
	Versions map[string][]implementation `json:"-"`
}

// maxVersionsInDropdown limits the version selector to the N newest versions (new versions appear at top).
const maxVersionsInDropdown = 3

var (
	doc             js.Value
	impls           []implementation
	allVersionsData map[string][]implementation
	currentVersion  string
	featHTTPGateway []featureDef
	featHTTPRoute   []featureDef
	featHTTPBackend []featureDef
	featGRPC        []featureDef
	featTLS         []featureDef
	featHTTPAll     []featureDef
	radioPrefix = map[string]string{"http": "req-http-", "grpc": "req-grpc-", "tls": "req-tls-"}
)

// Feature definitions come only from the loaded JSON (controller-wizard-data.json from hack/generate-controller-wizard-data.py).
// No hardcoded defaults: tables stay empty until data loads.

func main() {
	doc = js.Global().Get("document")

	// Expose callback for when JS has loaded the JSON data
	js.Global().Set("wizardOnData", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return nil
		}
		jsonStr := args[0].String()
		onDataLoaded(jsonStr)
		return nil
	}))

	// When a requirement checkbox is checked, uncheck the other (Must have / Nice to have are mutually exclusive)
	doc.Get("body").Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return nil
		}
		ev := args[0]
		target := ev.Get("target")
		if target.Get("type").String() != "checkbox" {
			return nil
		}
		if !target.Get("checked").Bool() {
			return nil
		}
		table := target.Call("closest", "table.features")
		if !table.Truthy() {
			return nil
		}
		name := target.Get("name").String()
		group := doc.Call("querySelectorAll", fmt.Sprintf(`input[name="%s"]`, name))
		for i := 0; i < group.Length(); i++ {
			el := group.Index(i)
			if !el.Equal(target) {
				el.Set("checked", false)
			}
		}
		return nil
	}))

	// Recommend button
	recommendBtn := doc.Call("getElementById", "recommend-btn")
	recommendBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		recommend()
		return nil
	}))

	// Reset button
	resetBtn := doc.Call("getElementById", "reset-btn")
	resetBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resetAll()
		return nil
	}))

	// Feature tables are filled only after data loads (onDataLoaded); no default consts.

	// Fetch data via JS fetch (go.run blocks so we do it from Go)
	js.Global().Call("fetch", "data/controller-wizard-data.json").Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resp := args[0]
		if resp.Get("ok").Bool() {
			resp.Call("json").Call("then", js.FuncOf(func(this js.Value, args2 []js.Value) interface{} {
				data := args2[0]
				jsonStr := js.Global().Get("JSON").Call("stringify", data).String()
				onDataLoaded(jsonStr)
				return nil
			}))
		} else {
			doc.Call("getElementById", "wizard-data-status").Set("textContent", "Could not load data/controller-wizard-data.json")
			doc.Call("getElementById", "recommend-btn").Set("disabled", false)
		}
		return nil
	})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		doc.Call("getElementById", "wizard-data-status").Set("textContent", "Could not load data. Run 'make wizard-data' (requires conformance/reports/), then serve from site-src/wizard/.")
		doc.Call("getElementById", "recommend-btn").Set("disabled", false)
		return nil
	}))

	select {}
}

func onDataLoaded(jsonStr string) {
	statusEl := doc.Call("getElementById", "wizard-data-status")
	btn := doc.Call("getElementById", "recommend-btn")
	versionRow := doc.Call("getElementById", "version-row")
	versionSelect := doc.Call("getElementById", "version-select")

	// Top-level array: just implementations
	var implArr []implementation
	if err := json.Unmarshal([]byte(jsonStr), &implArr); err == nil {
		impls = implArr
		allVersionsData = nil
		versionRow.Get("style").Set("display", "none")
		statusEl.Set("textContent", "")
		renderFeatureTablesFiltered()
		btn.Set("disabled", false)
		return
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		statusEl.Set("textContent", "Invalid JSON.")
		btn.Set("disabled", false)
		return
	}

	// Parse feature definitions (single source of truth: JSON from generate-controller-wizard-data.py)
	if fd, ok := raw["featureDefinitions"]; ok {
		var fdStruct struct {
			HTTPGateway    []featureDef `json:"httpGateway"`
			HTTPRoute      []featureDef `json:"httpRoute"`
			HTTPBackendTLS []featureDef `json:"httpBackendTls"`
			GRPC           []featureDef `json:"grpc"`
			TLS            []featureDef `json:"tls"`
		}
		if err := json.Unmarshal(fd, &fdStruct); err == nil {
			featHTTPGateway = fdStruct.HTTPGateway
			featHTTPRoute = fdStruct.HTTPRoute
			featHTTPBackend = fdStruct.HTTPBackendTLS
			featGRPC = fdStruct.GRPC
			featTLS = fdStruct.TLS
		}
	}
	featHTTPAll = append(append(append([]featureDef{}, featHTTPGateway...), featHTTPRoute...), featHTTPBackend...)

	// Single implementations array
	if implArr, ok := raw["implementations"]; ok {
		var list []implementation
		if err := json.Unmarshal(implArr, &list); err == nil {
			impls = list
			allVersionsData = nil
			versionRow.Get("style").Set("display", "none")
			statusEl.Set("textContent", "")
			renderFeatureTablesFiltered()
			btn.Set("disabled", false)
			return
		}
	}

	// Multi-version: keys like v1.4.0, v1.3.0, ...
	allVersionsData = make(map[string][]implementation)
	var versionKeys []string
	for k, v := range raw {
		if k == "featureDefinitions" {
			continue
		}
		var list []implementation
		if err := json.Unmarshal(v, &list); err != nil {
			continue
		}
		allVersionsData[k] = list
		versionKeys = append(versionKeys, k)
	}
	if len(versionKeys) == 0 {
		impls = nil
		statusEl.Set("textContent", "No versions in data file.")
		btn.Set("disabled", false)
		return
	}
	sort.Slice(versionKeys, func(i, j int) bool {
		return versionCompare(versionKeys[j], versionKeys[i]) < 0
	})
	if len(versionKeys) > maxVersionsInDropdown {
		versionKeys = versionKeys[:maxVersionsInDropdown]
	}
	currentVersion = versionKeys[0]
	impls = allVersionsData[currentVersion]

	versionRow.Get("style").Set("display", "block")
	versionSelect.Set("innerHTML", "")
	for _, v := range versionKeys {
		opt := doc.Call("createElement", "option")
		opt.Set("value", v)
		opt.Set("textContent", v)
		versionSelect.Call("appendChild", opt)
	}
	versionSelect.Set("value", currentVersion)
	versionSelect.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		currentVersion = versionSelect.Get("value").String()
		impls = allVersionsData[currentVersion]
		updateVersionLinks(currentVersion)
		renderFeatureTablesFiltered()
		doc.Call("getElementById", "results").Get("classList").Call("remove", "visible")
		statusEl.Set("textContent", "")
		return nil
	}))

	updateVersionLinks(currentVersion)
	renderFeatureTablesFiltered()
	statusEl.Set("textContent", "")
	btn.Set("disabled", false)
}

// versionSegment returns "v1.4" from "v1.4.0" for implementation table URLs.
func versionSegment(versionKey string) string {
	versionKey = strings.TrimPrefix(versionKey, "v")
	parts := strings.SplitN(versionKey, ".", 3)
	major, minor := "1", "0"
	if len(parts) > 0 && parts[0] != "" {
		major = parts[0]
	}
	if len(parts) > 1 && parts[1] != "" {
		minor = parts[1]
	}
	return "v" + major + "." + minor
}

func updateVersionLinks(versionKey string) {
	if versionKey == "" {
		versionKey = currentVersion
	}
	if versionKey == "" {
		versionKey = "v1.0.0"
	}
	seg := versionSegment(versionKey)
	base := "https://gateway-api.sigs.k8s.io/implementations/" + seg + "/"
	introLink := doc.Call("getElementById", "link-intro-conformance")
	if introLink.Truthy() {
		introLink.Set("href", base)
		introLink.Set("textContent", seg+" conformance tables")
	}
	for _, id := range []string{"link-httproute", "link-grpcroute", "link-tlsroute"} {
		el := doc.Call("getElementById", id)
		if el.Truthy() {
			el.Set("href", base+"#"+strings.TrimPrefix(id, "link-"))
		}
	}
}

func versionCompare(a, b string) int {
	parse := func(s string) (major, minor, patch int) {
		s = strings.TrimPrefix(s, "v")
		parts := strings.Split(s, ".")
		if len(parts) > 0 {
			fmt.Sscanf(parts[0], "%d", &major)
		}
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &minor)
		}
		if len(parts) > 2 {
			fmt.Sscanf(parts[2], "%d", &patch)
		}
		return
	}
	ma, mi, pa := parse(a)
	mb, mj, pb := parse(b)
	if ma != mb {
		return ma - mb
	}
	if mi != mj {
		return mi - mj
	}
	return pa - pb
}

func getAvailableFeatureIDs() map[string]bool {
	ids := make(map[string]bool)
	for _, impl := range impls {
		for _, f := range impl.Features {
			ids[f] = true
		}
	}
	return ids
}

func filterFeatures(list []featureDef, available map[string]bool) []featureDef {
	if len(available) == 0 {
		return list
	}
	var out []featureDef
	for _, f := range list {
		if available[f.ID] {
			out = append(out, f)
		}
	}
	return out
}

func renderFeatureTables() {
	renderHTTPRouteWithSpacer("http-route-features", featHTTPGateway, featHTTPRoute, featHTTPBackend, "http")
	renderTable("grpc-features", "grpc", "Gateway", "Gateway ", featGRPC)
	renderTable("tls-features", "tls", "Gateway", "Gateway ", featTLS)
}

func renderFeatureTablesFiltered() {
	avail := getAvailableFeatureIDs()
	renderHTTPRouteWithSpacer("http-route-features", filterFeatures(featHTTPGateway, avail), filterFeatures(featHTTPRoute, avail), filterFeatures(featHTTPBackend, avail), "http")
	renderTable("grpc-features", "grpc", "Gateway", "Gateway ", filterFeatures(featGRPC, avail))
	renderTable("tls-features", "tls", "Gateway", "Gateway ", filterFeatures(featTLS, avail))
}

func stripLabelPrefix(label, prefix string) string {
	if strings.HasPrefix(label, prefix) {
		return label[len(prefix):]
	}
	return label
}

func renderHTTPRouteWithSpacer(tableID string, gateway, route, backend []featureDef, section string) {
	tbody := doc.Call("querySelector", "#"+tableID+" tbody")
	if !tbody.Truthy() {
		return
	}
	var gatewayFirst, rest []featureDef
	for _, f := range gateway {
		if strings.HasPrefix(f.Label, "Gateway ") {
			gatewayFirst = append(gatewayFirst, f)
		} else {
			rest = append(rest, f)
		}
	}
	for _, f := range route {
		if strings.HasPrefix(f.Label, "Gateway ") {
			gatewayFirst = append(gatewayFirst, f)
		} else {
			rest = append(rest, f)
		}
	}
	prefix := radioPrefix[section]
	var html strings.Builder
	writeSubhead := func(subhead string) {
		html.WriteString(fmt.Sprintf(`<tr class="feature-subhead"><th scope="col">%s</th><th scope="col">Requirement</th></tr>`, escapeHTML(subhead)))
	}
	writeRow := func(f featureDef, label string) {
		name := prefix + f.ID
		titleAttr := ""
		if f.Description != "" {
			titleAttr = fmt.Sprintf(` title="%s"`, escapeHTML(f.Description))
		}
		html.WriteString(fmt.Sprintf(`<tr><td%s>%s</td><td>
<label><input type="checkbox" name="%s" value="must" /> Must have</label>
<label><input type="checkbox" name="%s" value="good" /> Nice to have</label>
</td></tr>`, titleAttr, escapeHTML(label), name, name))
	}
	if len(gatewayFirst) > 0 {
		writeSubhead("Gateway")
		for _, f := range gatewayFirst {
			writeRow(f, stripLabelPrefix(f.Label, "Gateway "))
		}
	}
	if len(rest) > 0 {
		writeSubhead("HTTPRoute")
		for _, f := range rest {
			writeRow(f, stripLabelPrefix(f.Label, "HTTPRoute "))
		}
	}
	if len(backend) > 0 {
		writeSubhead("Backend TLS")
		for _, f := range backend {
			writeRow(f, stripLabelPrefix(f.Label, "Backend TLS "))
		}
	}
	tbody.Set("innerHTML", html.String())
}

func renderTable(tableID string, section string, subhead string, labelPrefix string, rows []featureDef) {
	tbody := doc.Call("querySelector", "#"+tableID+" tbody")
	if !tbody.Truthy() {
		return
	}
	prefix := radioPrefix[section]
	var html strings.Builder
	if subhead != "" && len(rows) > 0 {
		html.WriteString(fmt.Sprintf(`<tr class="feature-subhead"><th scope="col">%s</th><th scope="col">Requirement</th></tr>`, escapeHTML(subhead)))
	}
	for _, f := range rows {
		label := f.Label
		if labelPrefix != "" {
			label = stripLabelPrefix(label, labelPrefix)
		}
		name := prefix + f.ID
		titleAttr := ""
		if f.Description != "" {
			titleAttr = fmt.Sprintf(` title="%s"`, escapeHTML(f.Description))
		}
		html.WriteString(fmt.Sprintf(`<tr><td%s>%s</td><td>
<label><input type="checkbox" name="%s" value="must" /> Must have</label>
<label><input type="checkbox" name="%s" value="good" /> Nice to have</label>
</td></tr>`, titleAttr, escapeHTML(label), name, name))
	}
	tbody.Set("innerHTML", html.String())
}

type selection struct {
	Section string
	ID      string
}

func getSelections() (must, good []selection) {
	sections := []struct {
		name  string
		feats []featureDef
	}{
		{"http", featHTTPAll},
		{"grpc", featGRPC},
		{"tls", featTLS},
	}
	for _, s := range sections {
		for _, f := range s.feats {
			name := radioPrefix[s.name] + f.ID
			el := doc.Call("querySelector", fmt.Sprintf(`input[name="%s"]:checked`, name))
			if !el.Truthy() {
				continue
			}
			v := el.Get("value").String()
			if v == "must" {
				must = append(must, selection{Section: s.name, ID: f.ID})
			} else if v == "good" {
				good = append(good, selection{Section: s.name, ID: f.ID})
			}
		}
	}
	mustSet := make(map[string]bool)
	for _, m := range must {
		mustSet[m.Section+"\x00"+m.ID] = true
	}
	var goodFiltered []selection
	for _, sel := range good {
		if !mustSet[sel.Section+"\x00"+sel.ID] {
			goodFiltered = append(goodFiltered, sel)
		}
	}
	return must, goodFiltered
}

func recommend() {
	resultsContent := doc.Call("getElementById", "results-content")
	resultsDiv := doc.Call("getElementById", "results")
	statusEl := doc.Call("getElementById", "wizard-data-status")

	resultsContent.Set("innerHTML", "")
	resultsDiv.Get("classList").Call("remove", "visible")
	if statusEl.Truthy() {
		statusEl.Set("textContent", "")
	}

	if len(impls) == 0 {
		resultsContent.Set("innerHTML", `<p class="no-results">No implementation data loaded. Load the wizard data first.</p>`)
		resultsDiv.Get("classList").Call("add", "visible")
		return
	}

	must, good := getSelections()
	if len(must) == 0 && len(good) == 0 {
		resultsContent.Set("innerHTML", `<p class="no-results">Select at least one requirement as Must have or Nice to have, then click Match.</p>`)
		resultsDiv.Get("classList").Call("add", "visible")
		setStatus(statusEl, 0)
		return
	}

	type scored struct {
		impl       implementation
		mustCount  int
		goodCount  int
		mustTotal  int
		goodTotal  int
		missing    []selection
		reportDate string
	}
	var scoredList []scored
	for _, impl := range impls {
		supp := make(map[string]bool)
		for _, f := range impl.Features {
			supp[f] = true
		}
		mustCount := 0
		var missing []selection
		for _, sel := range must {
			if supp[sel.ID] {
				mustCount++
			} else {
				missing = append(missing, sel)
			}
		}
		goodCount := 0
		for _, sel := range good {
			if supp[sel.ID] {
				goodCount++
			} else {
				missing = append(missing, sel)
			}
		}
		if mustCount >= 1 || (len(must) == 0 && goodCount >= 1) {
			scoredList = append(scoredList, scored{
				impl: impl, mustCount: mustCount, goodCount: goodCount,
				mustTotal: len(must), goodTotal: len(good), missing: missing,
				reportDate: impl.ReportDate,
			})
		}
	}
	if len(scoredList) == 0 {
		resultsContent.Set("innerHTML", `<p class="no-results">No controller supports any of your Must have requirements. Try relaxing to Nice to have or fewer requirements.</p>`)
		resultsDiv.Get("classList").Call("add", "visible")
		setStatus(statusEl, 0)
		return
	}
	sort.Slice(scoredList, func(i, j int) bool {
		if scoredList[j].mustCount != scoredList[i].mustCount {
			return scoredList[j].mustCount < scoredList[i].mustCount
		}
		if scoredList[j].goodCount != scoredList[i].goodCount {
			return scoredList[j].goodCount < scoredList[i].goodCount
		}
		return scoredList[j].reportDate < scoredList[i].reportDate
	})

	featureLabel := func(section, id string) string {
		var label string
		for _, f := range featHTTPAll {
			if f.ID == id {
				label = f.Label
				break
			}
		}
		if label == "" {
			for _, f := range featGRPC {
				if f.ID == id {
					label = f.Label
					break
				}
			}
		}
		if label == "" {
			for _, f := range featTLS {
				if f.ID == id {
					label = f.Label
					break
				}
			}
		}
		if label == "" {
			label = id
		}
		if section == "grpc" {
			return "(GRPC) " + label
		}
		if section == "tls" {
			return "(TLS) " + label
		}
		return label
	}
	var html strings.Builder
	html.WriteString(`<table class="results"><thead><tr><th>Organization</th><th>Project</th><th>Conformance</th><th>Must have</th><th>Nice to have</th><th>Missing</th></tr></thead><tbody>`)
	for _, c := range scoredList {
		conformance := strings.Join(c.impl.Conformance, ", ")
		missingLabels := make([]string, len(c.missing))
		for i, sel := range c.missing {
			missingLabels[i] = featureLabel(sel.Section, sel.ID)
		}
		missingStr := strings.Join(missingLabels, ", ")
		if missingStr == "" {
			missingStr = "—"
		}
		html.WriteString(fmt.Sprintf("<tr><td>%s</td><td><a href=\"%s\" target=\"_blank\" rel=\"noopener\">%s</a> %s</td><td>%s</td><td>%d/%d</td><td>%d/%d</td><td class=\"missing\">%s</td></tr>",
			escapeHTML(c.impl.Organization), escapeHTML(c.impl.URL), escapeHTML(c.impl.Project), escapeHTML(c.impl.Version),
			escapeHTML(conformance), c.mustCount, c.mustTotal, c.goodCount, c.goodTotal, escapeHTML(missingStr)))
	}
	html.WriteString("</tbody></table>")
	resultsContent.Set("innerHTML", html.String())
	resultsDiv.Get("classList").Call("add", "visible")
	resultsDiv.Call("scrollIntoView", map[string]interface{}{"behavior": "smooth", "block": "start"})
	setStatus(statusEl, len(scoredList))
}

func setStatus(el js.Value, n int) {
	if !el.Truthy() {
		return
	}
	if n == 1 {
		el.Set("textContent", "1 controller matches.")
	} else {
		el.Set("textContent", fmt.Sprintf("%d controllers match.", n))
	}
}

func resetAll() {
	sections := []struct {
		section string
		feats   []featureDef
	}{
		{"http", featHTTPAll},
		{"grpc", featGRPC},
		{"tls", featTLS},
	}
	for _, s := range sections {
		prefix := radioPrefix[s.section]
		for _, f := range s.feats {
			name := prefix + f.ID
			group := doc.Call("querySelectorAll", fmt.Sprintf(`input[name="%s"]`, name))
			for i := 0; i < group.Length(); i++ {
				group.Index(i).Set("checked", false)
			}
		}
	}
	doc.Call("getElementById", "results").Get("classList").Call("remove", "visible")
	statusEl := doc.Call("getElementById", "wizard-data-status")
	if statusEl.Truthy() {
		statusEl.Set("textContent", "")
	}
	js.Global().Call("scrollTo", 0, 0)
}

func seen(s []string, x string) bool {
	for _, v := range s {
		if v == x {
			return true
		}
	}
	return false
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
