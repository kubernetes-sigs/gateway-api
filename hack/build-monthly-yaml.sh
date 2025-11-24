#!/bin/bash

ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "$ROOT" || exit 1

TAG=monthly-$(date +"%Y.%m")

go run ./tools/generator --experimental-only --version=$TAG
bash hack/build-install-yaml.sh --experimental-only

# DON'T commit the generated YAML here.
git restore .

mv release/experimental-install.yaml release/${TAG}-install.yaml
echo "Generated release/${TAG}-install.yaml"

