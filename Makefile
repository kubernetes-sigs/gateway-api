# Copyright 2019 The Kubernetes Authors.
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

all: controller docs

# Kubebuilder driven custom resource definitions.
.PHONY: controller
controller:
	make -f kubebuilder.mk

# Build the documentation.
.PHONY: docs
docs:
	make -f docs.mk

.PHONY: clean
clean:
	make -f docs.mk clean