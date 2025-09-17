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
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"sigs.k8s.io/yaml"

	gep "sigs.k8s.io/gateway-api/pkg/gep"
)

var (
	GEPSDir        string
	MKDocsTemplate string
	SkipGEPNumber  string
)

// Those are the GEPs that will be included in the final navigation bar
// The order established below will be the order that the statuses will be shown
var includeGEPStatus = []gep.GEPStatus{
	gep.GEPStatusImplementable,
	gep.GEPStatusExperimental,
	gep.GEPStatusStandard,
	gep.GEPStatusMemorandum,
}

type GEPArray []GEPList

type GEPList struct {
	GepType string
	Geps    []uint
}

type TemplateData struct {
	GEPData GEPArray
}

const kindDetails = "GEPDetails"

func main() {
	flag.StringVar(&GEPSDir, "g", "", "Defines the absolute path of the directory containing the GEPs")
	flag.StringVar(&MKDocsTemplate, "t", "", "Defines the absolute path of mkdocs.yaml file")
	flag.StringVar(&SkipGEPNumber, "s", "696", "Defines GEPs number to be skipped, should be comma-separated")
	flag.Parse()

	if GEPSDir == "" || MKDocsTemplate == "" {
		log.Fatal("-g and -c are mandatory arguments")
	}

	if strings.Contains(SkipGEPNumber, " ") {
		log.Fatal("-s flag should not contain spaces")
	}

	skipGep := strings.Split(SkipGEPNumber, ",")

	geps, err := walkGEPs(GEPSDir, skipGep)
	if err != nil {
		panic(err)
	}

	tmpl, err := template.ParseFiles(MKDocsTemplate)
	if err != nil {
		log.Fatalf("error reading mkdocs template: %s", err)
	}

	tmplData := TemplateData{
		GEPData: geps,
	}

	buf := &bytes.Buffer{}

	if err := tmpl.Execute(buf, tmplData); err != nil {
		panic(err)
	}
	fmt.Print(buf.String())
}

func walkGEPs(dir string, skipGEPs []string) (GEPArray, error) {
	gepArray := make(GEPArray, 0)
	tmpMap := make(map[gep.GEPStatus]GEPList)

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
			tmpMap[gepDetail.Status] = GEPList{
				GepType: string(gepDetail.Status),
				Geps:    make([]uint, 0),
			}
		}

		item := tmpMap[gepDetail.Status]
		item.Geps = append(item.Geps, gepDetail.Number)
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
		sort.SliceStable(gepArray[i].Geps, func(x, y int) bool {
			return gepArray[i].Geps[x] < gepArray[i].Geps[y]
		})
	}

	return gepArray, nil
}
