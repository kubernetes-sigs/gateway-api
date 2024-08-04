# Copyright 2024 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GIT_COMMIT := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)

BIN_DIR := bin

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

deps:
	@echo "Installing dependencies..."
	@go version
	@go mod tidy
	@go mod vendor

build: deps
	@echo "Building gwctl..."
	@echo "GIT_COMMIT=$(GIT_COMMIT)"
	@echo "BUILD_DATE=$(BUILD_DATE)"
	@go build -ldflags="-X sigs.k8s.io/gateway-api/gwctl/pkg/version.gitCommit=$(GIT_COMMIT) -X sigs.k8s.io/gateway-api/gwctl/pkg/version.buildDate=$(BUILD_DATE)" -o bin/gwctl main.go
	@echo "Done"

clean:
	@echo "Cleaning up..."
	@rm -rf $(BIN_DIR)

.DEFAULT_GOAL := build