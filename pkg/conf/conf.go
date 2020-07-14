package conf

var (
	Cert string
	Key  string
	Addr string
)

var ImageRename []string
var DefaultImageRenameRules = []string{
	"k8s.gcr.io/=gcrxio/k8s.gcr.io_",
	"gcr.io/kubernetes-helm/=gcrxio/gcr.io_kubernetes-helm_",
	"gcr.io/istio-release/=gcrxio/gcr.io_istio-release_",
	"gcr.io/linkerd-io/=gcrxio/gcr.io_linkerd-io_",
	"gcr.io/spinnaker-marketplace/=gcrxio/gcr.io_spinnaker-marketplace_",
	"gcr.io/distroless/=gcrxio/gcr.io_distroless_",
	"gcr.io/google-samples/=gcrxio/gcr.io_google-samples_",
	"gcr.io/knative-releases/=gcrxio/gcr.io_knative-releases_",
}

var ForceDeployLabel string
var DefaultForceDeployLabel = "force-deploy.mritd.me"

var AllowDeployTime []string
var DefaultAllowDeployTime = []string{
	"05:00~10:00",
	"14:00~15:00",
}
