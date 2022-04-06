#!/usr/bin/env bash

# Copyright 2020 The Kubernetes Authors.
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

set -o errexit
set -o nounset
set -o pipefail

thisyear=`date +"%Y"`

mkdir -p install

# Make clean files with boilerplate
cat hack/boilerplate/boilerplate.sh.txt > install/experimental.yaml
sed -i "s/YEAR/$thisyear/g" install/experimental.yaml
cat << EOF >> install/experimental.yaml
#
# Gateway API Experimental channel install
#
EOF

cat hack/boilerplate/boilerplate.sh.txt > install/stable.yaml
sed -i "s/YEAR/$thisyear/g" install/stable.yaml
cat << EOF >> install/stable.yaml
#
# Gateway API Stable channel install
#
EOF

for file in `ls config/webhook/*.yaml config/crd/experimental/*.yaml`
do
    echo "---" >> install/experimental.yaml
    echo "#" >> install/experimental.yaml
    echo "# $file" >> install/experimental.yaml
    echo "#" >> install/experimental.yaml
    cat $file >> install/experimental.yaml
done

for file in `ls config/webhook/*.yaml config/crd/stable/*.yaml`
do
    echo "---" >> install/stable.yaml
    echo "#" >> install/stable.yaml
    echo "# $file" >> install/stable.yaml
    echo "#" >> install/stable.yaml
    cat $file >> install/stable.yaml
done

