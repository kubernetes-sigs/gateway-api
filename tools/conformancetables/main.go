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

// conformancetables generates the implementation conformance tables
// for the site from the reports in conformance/reports.
package main

import (
	"flag"
	"log"
)

func main() {
	reports := flag.String("r", "conformance/reports", "directory containing the conformance reports")
	out := flag.String("o", "site/content/en/docs/implementations/versions", "directory to write the generated pages to")
	flag.Parse()

	if err := WritePages(*reports, *out); err != nil {
		log.Fatalf("generating conformance tables: %v", err)
	}
}
