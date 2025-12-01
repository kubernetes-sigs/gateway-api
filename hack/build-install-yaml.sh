#!/usr/bin/env bash

# Copyright 2022 The Kubernetes Authors.
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

CHANNELS=(standard experimental)

if [ "$1" == "--experimental-only" ]; then
    CHANNELS=(experimental)
    shift
fi

readonly YEAR=$(date +"%Y")

mkdir -p release/

for CHANNEL in "${CHANNELS[@]}"; do
    echo $CHANNEL
    # Make clean files with boilerplate
    cat hack/boilerplate/boilerplate.sh.txt > release/${CHANNEL}-install.yaml

    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        sed -i "s/YEAR/${YEAR}/g" release/${CHANNEL}-install.yaml
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s/YEAR/${YEAR}/g" release/${CHANNEL}-install.yaml
    else
        echo "Unsupported OS"
        exit 1
    fi

    cat << EOF >> release/${CHANNEL}-install.yaml
#
# Gateway API ${CHANNEL^} channel install
#
EOF

    for file in config/crd/${CHANNEL}/gateway*.yaml
    do
        echo "---" >> release/${CHANNEL}-install.yaml
        echo "#" >> release/${CHANNEL}-install.yaml
        echo "# $file" >> release/${CHANNEL}-install.yaml
        echo "#" >> release/${CHANNEL}-install.yaml
        cat "$file" >> release/${CHANNEL}-install.yaml
    done

done

echo "Generated:" release/*-install.yaml
