module github.com/mritd/goadmission

go 1.18

require (
	github.com/gorilla/mux v1.8.1
	github.com/json-iterator/go v1.1.12
	github.com/spf13/cobra v1.7.0
	go.uber.org/zap v1.26.0
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
)

require (
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/klog/v2 v2.30.0 // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
	sigs.k8s.io/json v0.0.0-20211020170558-c049b76a60c6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.23.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.23.5
)

// common replace
//replace (
//	k8s.io/api => k8s.io/api v0.22.2
//	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.2
//	k8s.io/apimachinery => k8s.io/apimachinery v0.22.2
//	k8s.io/apiserver => k8s.io/apiserver v0.22.2
//	k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.2
//	k8s.io/client-go => k8s.io/client-go v0.22.2
//	k8s.io/cloud-provider => k8s.io/cloud-provider v0.22.2
//	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.22.2
//	k8s.io/code-generator => k8s.io/code-generator v0.22.2
//	k8s.io/component-base => k8s.io/component-base v0.22.2
//	k8s.io/cri-api => k8s.io/cri-api v0.22.2
//	k8s.io/csi-api => k8s.io/csi-api v0.22.2
//	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.22.2
//	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.22.2
//	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.22.2
//	k8s.io/kube-proxy => k8s.io/kube-proxy v0.22.2
//	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.22.2
//	k8s.io/kubectl => k8s.io/kubectl v0.22.2
//	k8s.io/kubelet => k8s.io/kubelet v0.22.2
//	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.22.2
//	k8s.io/metrics => k8s.io/metrics v0.22.2
//	k8s.io/node-api => k8s.io/node-api v0.22.2
//	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.22.2
//	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.22.2
//	k8s.io/sample-controller => k8s.io/sample-controller v0.22.2
//)
