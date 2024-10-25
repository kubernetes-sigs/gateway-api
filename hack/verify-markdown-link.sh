#!/bin/bash  
# Copyright 2014 The Kubernetes Authors.  
# Licensed under the Apache License, Version 2.0 (the "License");  
# you may not use this file except in compliance with the License.  
# You may obtain a copy of the License at  
#     http://www.apache.org/licenses/LICENSE-2.0  
# Unless required by applicable law or agreed to in writing, software  
# distributed under the License is distributed on an "AS IS" BASIS,  
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  
# See the License for the specific language governing permissions and  
# limitations under the License.  
  
set -o errexit  
set -o nounset  
set -o pipefail  
  
log_error() {  
    echo "ERROR: $*" 1>&2  
}  
  
log_info() {  
    echo "INFO: $*" 1>&2  
}  
  
extract_external_links() {  
    local file_path="$1"  
    local -n links_array=$2  # Use nameref to pass array by reference  
  
    # Check if the file exists and is readable  
    if [[ -f "$file_path" && -r "$file_path" ]]; then  
        # Read the content of the file  
        local content=$(<"$file_path")  
        # Use grep with a refined regex pattern to extract links  
        while IFS= read -r link; do  
            # Trim trailing characters like ')' or ']'  
            link=$(echo "$link" | sed 's/[)"]$//')  
            links_array+=("$link")  
        done < <(echo "$content" | grep -oP 'https?://[^\s")]+')  
    else  
        log_error "File does not exist or is not readable: $file_path"  
        return 1  
    fi  
}  
  
validate_url() {  
    local url="$1"  
    local http_status=$(curl -o /dev/null -s -w "%{http_code}\n" "$url")  
  
    # Check if the status code indicates success (200-399)  
    if [[ "$http_status" -ge 200 && "$http_status" -lt 400 ]]; then  
        log_info "Success: $url returned status code $http_status"  
        return 0  
    else  
        log_error "Error: $url returned status code $http_status"  
        return 1  
    fi  
}  
  
# Check for broken links in markdown files within a specified directory  
check_broken_links_in_directory() {  
    local directory=$1  
    local exit_code=0  
  
    # Find all .md files in the specified directory and its subdirectories  
    for markdown_file in $(find "$directory" -type f -name "*.md"); do  
        log_info "Checking $markdown_file for broken links"  
          
        declare -a links  
        extract_external_links "$markdown_file" links  
  
        for link in "${links[@]}"; do  
            validate_url "$link"  
            result=$?  
            if [[ $result -eq 1 ]]; then  
                exit_code=1  
            fi  
        done  
    done  
  
    return $exit_code  
}  
  
ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..  
  
if [[ ! -d "$ROOT_DIR" ]]; then  
    log_error "The specified directory $ROOT_DIR does not exist"  
    exit 1  
fi  
  
# Check for broken links in markdown files  
check_broken_links_in_directory "$ROOT_DIR"  
exit_code=$?  
  
exit $exit_code  
