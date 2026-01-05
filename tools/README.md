This directory contains the tools used on diverse CI pipelines, Makefile targets, etc.

It is intended to be a separate directory with a separate go.mod to avoid adding
dependencies to main Gateway API.

## Common workflows
Any workflow here should be executed from the repo root directory.

### Adding a new tool:

Adding a new tools means the tool will be added to the specific `go.mod` file.
It is highly recommended that a version is used/pinned, the example below will pick 
the latest tagged version and add to the tools file.

`go get -tool -modfile=tools/go.mod golang.org/x/vuln/cmd/govulncheck@latest`

### Executing a tool:
Executing a tool means the same tool pinned on the `go.mod` file will be 
built and cached on `$XDG_CONFIG/.cache/go-build` the first time it is called, and 
then the binary will be executed normally. 

`go tool -modfile=tools/go.mod govulncheck`