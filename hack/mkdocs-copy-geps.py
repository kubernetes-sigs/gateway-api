# Copyright 2023 The Kubernetes Authors.
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

import shutil
import logging
from mkdocs import plugins
import yaml
import pandas
from fnmatch import fnmatch
import glob

log = logging.getLogger('mkdocs')

@plugins.event_priority(100)
def on_pre_build(config, **kwargs):
    log.info("copying geps")
    shutil.copytree("geps","site-src/geps", dirs_exist_ok=True)

    # calling to get the conformance reports generated
    yamlReports = getYaml()
    
    generate_conformance_tables(yamlReports)



# NOTE: will have to be updated if new features are added
httproute_extended_conformance_features_list = ['HTTPRouteBackendRequestHeaderModification',"HTTPRouteQueryParamMatching",'HTTPRouteMethodMatching',"HTTPRouteResponseHeaderModification","HTTPRoutePortRedirect","HTTPRouteSchemeRedirect","HTTPRoutePathRedirect","HTTPRouteHostRewrite","HTTPRoutePathRewrite","HTTPRouteRequestMirror","HTTPRouteRequestMultipleMirrors","HTTPRouteRequestTimeout", "HTTPRouteBackendTimeout","HTTPRouteParentRefPort"]

def generate_conformance_tables(reports):

  gateway_http_table = generate_profiles_report(reports,'HTTP')
  # gateway_tls_table = generate_profiles_report(reprots,'TLS') #currently no implementation has extended supported features listed. Can uncomment once a list is needed to keep track
    
  mesh_http_table = generate_profiles_report(reports,'MESH')  


  with open('site-src/implementation-table.md','w') as f:
    f.write("The following tables are populated from the conformance reports uploaded by project implementations. They are separated into the extended features that each project supports listed in their reports.\n\n")

    f.write("# Gateway Profile\n\n")
    f.write("## HTTPRoute\n\n")
    f.write(gateway_http_table.to_markdown()+'\n\n')

    f.write("# Mesh Profile\n\n")
    f.write("## HTTPRoute\n\n")
    f.write(mesh_http_table.to_markdown())

# reports: raw report passed , route: the x in xRoute
def generate_profiles_report(reports, route): 

  http_reports = reports.loc[reports["name"]==route]
  projects = http_reports['organization']
  http_reports.set_index('organization')
  http_reports.sort_values(['organization','version'], inplace=True)
  http_reports.drop_duplicates(subset='organization', inplace=True, keep='last')
  
  http_table = pandas.DataFrame(columns=http_reports['organization'])
  http_table.insert(loc=0, column='Features', value=httproute_extended_conformance_features_list)
  http_reports= http_reports[["organization","extended.supportedFeatures"]] 
    
  http_table.set_index('Features')
  for feat in  httproute_extended_conformance_features_list:
    
    for proj in projects: # for each project, check if the feature is supported

      if feat in http_reports.loc[http_reports["organization"]==proj]['extended.supportedFeatures'].to_list()[0]:
        http_table.loc[http_table['Features']==feat,proj] = ':white_check_mark:'
      else:
        http_table.loc[http_table['Features']==feat,proj] = ':x:'
  
  return http_table.set_index('Features').T



# the path should be changed when there is a new version
conformance_path = "conformance/reports/v1.0.0/**"
def getYaml():
    log.info("parsing conformance reports ============================")
    yamls = []

    # reports must be named according to the following pattern : <API Channel>-<Implementation version>-<mode>-report.yaml

    for p in glob.glob(conformance_path, recursive=True): # getting all the paths in conforamnce

        if fnmatch(p, "*.yaml"):
            
            x = load_yaml(p)
            profiles = pandas.json_normalize(x, record_path='profiles',meta=["implementation"] ) 
            
            implementation = pandas.json_normalize(profiles.implementation)
            yamls.append(pandas.concat([implementation,profiles], axis=1))

    yamls = pandas.concat(yamls)
    return yamls

def load_yaml(name):
    with open(name, 'r') as file:
        x = yaml.safe_load(file)

    return x

