module sigs.k8s.io/service-apis

go 1.15

require (
	github.com/ahmetb/gen-crd-api-reference-docs v0.2.0
	github.com/go-logr/logr v0.2.1 // indirect
	github.com/onsi/ginkgo v1.13.0 // indirect
	golang.org/x/net v0.0.0-20200904194848-62affa334b73 // indirect
	golang.org/x/tools v0.0.0-20200904185747-39188db58858 // indirect
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/code-generator v0.19.2
	k8s.io/klog/v2 v2.3.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200831175022-64514a1d5d59 // indirect
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-tools v0.4.0
)

replace github.com/ahmetb/gen-crd-api-reference-docs => github.com/jpeach/gen-crd-api-reference-docs v0.2.1-0.20201214045921-2511762f1bee
