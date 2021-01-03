module github.com/snorwin/helm-operator

go 1.15

require (
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.10 // indirect
	github.com/go-logr/logr v0.3.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	helm.sh/helm/v3 v3.4.2
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.7.0
)
