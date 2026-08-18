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

package main

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"golang.org/x/mod/semver"
	"sigs.k8s.io/yaml"

	v1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
)

//go:embed templates/*
var templateFiles embed.FS

var (
	ImplsDir        string
	IntegrationsDir string
	ReportsDir      string
	OutFile         string
)

// TODO(youngnick): Figure out some way to make this more centralized.
// Don't open a PR about this without talking with youngnick first.
var (
	CurrentVersion     = "v1.6"
	ConformantVersions = []string{CurrentVersion, "v1.5"}
	StaleVersions      = []string{"v1.4"}
)

// FullNameToHeadingLink transforms an implementation's full name
// into the canonical Markdown heading link format.
func FullNameToHeadingLink(fullName string) string {
	// Spaces to dash
	out := strings.ReplaceAll(fullName, " ", "-")
	// Remove all hashes
	out = strings.ReplaceAll(out, "#", "")
	// Add the hash prefix
	out = "#" + out
	// Make sure it's lower case
	return strings.ToLower(out)
}

func main() {
	flag.StringVar(&ImplsDir, "d", "conformance/list/implementations", "Defines the path relative to the repo root of the list directory containing the details files for implementations")
	flag.StringVar(&IntegrationsDir, "i", "conformance/list/integrations", "Defines the path relative to the repo root of the list directory containing the details files for integrations")
	flag.StringVar(&ReportsDir, "r", "conformance/reports", "Defines the path relative to the repo root of the list directory containing the conformance reports")
	flag.StringVar(&OutFile, "o", "site/content/en/docs/implementations/list.md", "Defines the path relative to the repo root of the output file")

	flag.Parse()

	log.SetOutput(os.Stdout)

	implDetails, err := readDetails(ImplsDir)
	if err != nil {
		log.Fatalf("Couldn't read implementation details: %s", err)
	}
	log.Printf("Found %d implementation details entries", len(implDetails))

	confReports, err := readReports(ReportsDir)
	if err != nil {
		log.Fatalf("Couldn't read conformance reports: %s", err)
	}
	log.Printf("Found %d candidate conformance reports", len(confReports))

	templateOut := NewTemplateOutput()
	integrations, err := readDetails(IntegrationsDir)
	if err != nil {
		log.Fatalf("Couldn't read integration details: %s", err)
	}
	templateOut.Integrations = integrations

	templateOut.FilterConformantReports(confReports, implDetails)

	funcMap := template.FuncMap{
		"FullNameToHeadingLink": FullNameToHeadingLink,
		"linkDetails":           templateOut.LinkDetails,
		"pathEscape":            url.PathEscape,
	}

	tmpl, err := template.New("impl-list.tmpl").Funcs(funcMap).ParseFS(templateFiles, "templates/impl-list.tmpl")
	if err != nil {
		log.Fatalf("error reading mkdocs template: %s", err)
	}

	templateBuffer := &bytes.Buffer{}
	if errTmpl := tmpl.Execute(templateBuffer, templateOut); errTmpl != nil {
		log.Fatalf("error rendering template: %s", errTmpl)
	}

	headerBytes, err := templateFiles.ReadFile("templates/impl-list-header.md")
	if err != nil {
		log.Fatalf("Couldn't read header from templates/impl-list-header.md")
	}

	footerBytes, err := templateFiles.ReadFile("templates/impl-list-footer.md")
	if err != nil {
		log.Fatalf("Couldn't read footer from templates/impl-list-footer.md")
	}

	outputBytes := bytes.Join([][]byte{headerBytes, templateBuffer.Bytes(), footerBytes}, []byte("\n"))

	if errTmpl := os.WriteFile(OutFile, outputBytes, 0o600); errTmpl != nil {
		log.Fatalf("error writing file: %s", errTmpl)
	}
}

func readDetails(dir string) (DetailsMap, error) {
	out := make(DetailsMap)

	rootFS := os.DirFS(dir)

	log.Printf("Loading implementation details from %s", dir)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %s: %w", path, err)
		}
		if d.IsDir() || d.Name() != "details.yaml" {
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		content, err := fs.ReadFile(rootFS, relPath)
		if err != nil {
			return err
		}

		implListDetail := &v1.ImplementationListDetail{}
		log.Printf("loading %s", path)
		if err := yaml.Unmarshal(content, implListDetail); err != nil {
			return err
		}

		out[implKey(implListDetail.Implementation)] = *implListDetail
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func readReports(dir string) (ReportsMap, error) {
	out := make(ReportsMap)
	rootFS := os.DirFS(dir)

	log.Printf("Loading Conformance reports from %s", dir)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %s: %w", path, err)
		}
		if d.IsDir() {
			return nil
		}

		if filepath.Ext(d.Name()) != ".yaml" {
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		reportVersion := strings.Split(relPath, "/")[0]

		if slices.Contains(ConformantVersions, semver.MajorMinor(reportVersion)) &&
			slices.Contains(StaleVersions, semver.MajorMinor(reportVersion)) {
			return nil
		}

		log.Printf("Found a candidate conformance report: %s", relPath)

		content, err := fs.ReadFile(rootFS, relPath)
		if err != nil {
			return err
		}

		confReport := &v1.ConformanceReport{}
		if err := yaml.Unmarshal(content, confReport); err != nil {
			return err
		}

		out[relPath] = *confReport
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func implKey(i v1.Implementation) string {
	return i.Organization + "-" + i.Project
}

// DetailsMap is a map of the full name to the relevant ImplementationListDetail
type DetailsMap map[string]v1.ImplementationListDetail

// ReportsMap maps the path of a conformance report to its contents.
// Used for reading in all the conformance reports before processing.
// The path is used for generating the links inside the badges.
type ReportsMap map[string]v1.ConformanceReport

// ConformanceDetails holds all the information needed to print the snipped
// about each implementation in the combined list.
type ConformanceDetails struct {
	Version  string
	Path     string
	Error    string                      `json:"error,omitempty"`
	ImplList v1.ImplementationListDetail `json:"implDetail"`
}

// ConformanceMap is a map of the full name to the version it is conformant in.
type ConformanceMap map[string]ConformanceDetails

// UpdateConformanceDetails makes sure that the most recent conformance report is
// updated over older versions, only.
func (c ConformanceMap) UpdateConformanceDetails(cd ConformanceDetails, implKey string) {
	existing, ok := c[implKey]
	if !ok {
		c[implKey] = cd
		return
	}
	if semver.Compare(cd.Version, existing.Version) > 0 {
		// cd is newer, so update
		c[implKey] = cd
	}
}

// ConformantImpls maps the full name of the implementation to any errors in
// its best report. Used for generating the list of various classes of implementation.
type ConformantImpls map[string]string

type TemplateOutput struct {
	GatewayConformantImpls ConformantImpls
	GatewayPartialImpls    ConformantImpls
	MeshConformantImpls    ConformantImpls
	MeshPartialImpls       ConformantImpls
	AllImplementations     ConformanceMap
	Integrations           DetailsMap
}

func NewTemplateOutput() TemplateOutput {
	var t TemplateOutput
	t.GatewayConformantImpls = make(ConformantImpls)
	t.GatewayPartialImpls = make(ConformantImpls)
	t.MeshConformantImpls = make(ConformantImpls)
	t.MeshPartialImpls = make(ConformantImpls)
	t.AllImplementations = make(ConformanceMap)
	t.Integrations = make(DetailsMap)

	return t
}

// ImplLinkDetails is used in the report output to generate the link badges.
type ImplLinkDetails struct {
	Color  string
	Result string
}

// LinkDetails looks up an implementation by its fullName
// and returns a level and a color for the badge in the implementations list.
func (t *TemplateOutput) LinkDetails(fullName, trafficType string) ImplLinkDetails {
	switch strings.ToLower(trafficType) {
	case "gateway":
		if _, ok := t.GatewayConformantImpls[fullName]; ok {
			return ImplLinkDetails{
				Color:  "green",
				Result: "Conformant",
			}
		}

		if _, ok := t.GatewayPartialImpls[fullName]; ok {
			return ImplLinkDetails{
				Color:  "orange",
				Result: "Partially Conformant",
			}
		}
	case "mesh":
		if _, ok := t.MeshConformantImpls[fullName]; ok {
			return ImplLinkDetails{
				Color:  "green",
				Result: "Conformant",
			}
		}

		if _, ok := t.MeshPartialImpls[fullName]; ok {
			return ImplLinkDetails{
				Color:  "orange",
				Result: "Partially Conformant",
			}
		}
	}

	return ImplLinkDetails{
		Color:  "red",
		Result: "Not Conformant",
	}
}

func (t *TemplateOutput) FilterConformantReports(confReports ReportsMap, implDetails DetailsMap) {
	// Make sure we are deterministically ranging across confReports
	paths := slices.Sorted(maps.Keys(confReports))
	for _, path := range paths {
		report := confReports[path]
		resultGateway, resultMesh, err := report.IsConformant(CurrentVersion)
		reportFor := implKey(report.Implementation)
		// Store any errors with the report as the value in the conformance result tables
		result := ""

		if !slices.Contains(ConformantVersions, semver.MajorMinor(report.GatewayAPIVersion)) {
			// If the report is not in the last two versions, it can only be Partial at maximum.
			if resultGateway == v1.Success {
				resultGateway = v1.Partial
				err = errors.New("report is more than two versions old")
			}
			if resultMesh == v1.Success {
				resultMesh = v1.Partial
				err = errors.New("report is more than two versions old")
			}
		}

		if err != nil {
			result = err.Error()
			log.Printf("Report for %s, version %s, result %s, from %s, reason:%s", reportFor, report.GatewayAPIVersion, resultGateway, path, err)
		} else {
			log.Printf("Report for %s, version %s, result %s, from %s", reportFor, report.GatewayAPIVersion, resultGateway, path)
		}

		detailsFor, ok := implDetails[reportFor]
		if !ok {
			// This should cause the program to end with an erorr so that CI fails eventually
			// Maybe this function should return an error?
			log.Printf("ERROR: Report for %s does not have an implementation details file in conformance/list/implementations", reportFor)
			continue
		}

		fullName := detailsFor.FullName

		cd := ConformanceDetails{
			Version:  report.GatewayAPIVersion,
			Path:     path,
			ImplList: detailsFor,
		}

		switch resultGateway {
		case v1.Partial:
			t.GatewayPartialImpls[fullName] = result
		case v1.Success:
			t.GatewayConformantImpls[fullName] = result
		}

		switch resultMesh {
		case v1.Partial:
			t.MeshPartialImpls[fullName] = result
		case v1.Success:
			t.MeshConformantImpls[fullName] = result
		}

		if resultGateway != v1.Failure || resultMesh != v1.Failure {
			t.AllImplementations.UpdateConformanceDetails(cd, fullName)
		}
	}

	removeDuplicateConformanceReports(t.GatewayPartialImpls, t.GatewayConformantImpls)
	removeDuplicateConformanceReports(t.MeshPartialImpls, t.MeshConformantImpls)
}

func removeDuplicateConformanceReports(partialMap ConformantImpls, conformantMap ConformantImpls) {
	for conformantKey := range conformantMap {
		for partialKey := range partialMap {
			if conformantKey != partialKey {
				continue
			}
			// If there's a conformant and a partial report for the same implementation,
			// only the conformant one counts.
			//
			// NOTE FOR REVIEW:
			//
			// This means that, if you submit a conformance report for an implementation
			// that's partial, but newer, the older one will count.
			delete(partialMap, partialKey)
		}
	}
}
