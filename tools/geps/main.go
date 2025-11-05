/*
Copyright 2025 The Kubernetes Authors.

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
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"sigs.k8s.io/yaml"

	gep "sigs.k8s.io/gateway-api/pkg/gep"
)

//go:embed templates/*.tmpl
var templateFile embed.FS

var (
	GEPSDir       string
	SkipGEPNumber string
)

// Those are the GEPs that will be included in the final navigation bar
// The order established below will be the order that the statuses will be shown
var includeGEPStatus = []gep.GEPStatus{
	gep.GEPStatusStandard,
	gep.GEPStatusMemorandum,
	gep.GEPStatusExperimental,
	gep.GEPStatusImplementable,
	gep.GEPStatusPrototyping,
	gep.GEPStatusProvisional,
}

type GEPArray []GEPs

type GEPs struct {
	GepType     string
	GepsDetails []*gep.GEPDetail
}

const kindDetails = "GEPDetails"

func main() {
	flag.StringVar(&GEPSDir, "g", "", "Defines the absolute path of the directory containing the GEPs")
	flag.StringVar(&SkipGEPNumber, "s", "696", "Defines GEPs number to be skipped, should be comma-separated")

	flag.Parse()

	if GEPSDir == "" {
		log.Fatal("-g is mandatory arguments")
	}

	if strings.Contains(SkipGEPNumber, " ") {
		log.Fatal("-s flag should not contain spaces")
	}

	skipGep := strings.Split(SkipGEPNumber, ",")

	tmpl, err := template.ParseFS(templateFile, "templates/template.tmpl")
	if err != nil {
		log.Fatalf("error reading mkdocs template: %s", err)
	}

	geps, err := walkGEPs(GEPSDir, skipGep)
	if err != nil {
		log.Fatalf("error walking GEPs: %s", err)
	}

	for _, gep := range geps {
		buf := &bytes.Buffer{}
		fileName := fmt.Sprintf("%s/landing/%s.md", GEPSDir, strings.ToLower(gep.GepType))
		if errTmpl := tmpl.Execute(buf, gep); errTmpl != nil {
			log.Fatalf("error rendering template: %s", errTmpl)
		}

		if errTmpl := os.WriteFile(fileName, buf.Bytes(), 0o600); errTmpl != nil {
			log.Fatalf("error writing file: %s", errTmpl)
		}
	}

	tmplTab, err := template.ParseFS(templateFile, "templates/template-tab.tmpl")
	if err != nil {
		log.Fatalf("error reading mkdocs template: %s", err)
	}
	buf := &bytes.Buffer{}
	fileName := fmt.Sprintf("%s/landing/tab.md", GEPSDir)
	if err := tmplTab.Execute(buf, geps); err != nil {
		log.Fatalf("error rendering template: %s", err)
	}

	if err := os.WriteFile(fileName, buf.Bytes(), 0o600); err != nil {
		log.Fatalf("error writing file: %s", err)
	}
}

func walkGEPs(dir string, skipGEPs []string) (GEPArray, error) {
	gepArray := make(GEPArray, 0)
	tmpMap := make(map[gep.GEPStatus]GEPs)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %s: %w", path, err)
		}
		if d.IsDir() || d.Name() != "metadata.yaml" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		gepDetail := &gep.GEPDetail{}
		log.Printf("checking %s", path)
		if err := yaml.Unmarshal(content, gepDetail); err != nil {
			return err
		}

		if gepDetail.Kind != kindDetails {
			return nil
		}

		// Skip the GEPs types we don't care
		if !slices.Contains(includeGEPStatus, gepDetail.Status) {
			return nil
		}

		// Skip the GEPs numbers we don't care
		if slices.Contains(skipGEPs, strconv.FormatUint(uint64(gepDetail.Number), 10)) {
			return nil
		}

		// Add the GEP to a map indexed by GEP types, so we can provide the sorted array
		// easily later
		_, ok := tmpMap[gepDetail.Status]
		if !ok {
			tmpMap[gepDetail.Status] = GEPs{
				GepType:     string(gepDetail.Status),
				GepsDetails: make([]*gep.GEPDetail, 0),
			}
		}

		item := tmpMap[gepDetail.Status]
		item.GepsDetails = append(item.GepsDetails, gepDetail)
		tmpMap[gepDetail.Status] = item
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Include the GEPs toc on the desired order
	for _, v := range includeGEPStatus {
		if geps, ok := tmpMap[v]; ok {
			gepArray = append(gepArray, geps)
		}
	}

	for i := range gepArray {
		sort.SliceStable(gepArray[i].GepsDetails, func(x, y int) bool {
			return gepArray[i].GepsDetails[x].Number < gepArray[i].GepsDetails[y].Number
		})
	}

	return gepArray, nil
}
