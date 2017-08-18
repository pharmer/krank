package cloud

import (
	"fmt"
	"io"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"
)

const (
	ProviderName = "krank"
)

func init() {
	cloudprovider.RegisterCloudProvider(ProviderName, newKrankCloudProvider)
}

type KrankCloudProvider struct {
	client *kubernetes.Clientset
	conf   io.Reader
}

func newKrankCloudProvider(reader io.Reader) (cloudprovider.Interface, error) {
	cfg, err := rest.InClusterConfig()

	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes client config: %s", err.Error())
	}

	ct, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes client: %s", err.Error())
	}

	return &KrankCloudProvider{ct, reader}, nil

}

func (k *KrankCloudProvider) Initialize(clientBuilder controller.ControllerClientBuilder) {}

func (k *KrankCloudProvider) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return nil, false
}

func (k *KrankCloudProvider) Instances() (cloudprovider.Instances, bool) {
	return nil, false
}

func (k *KrankCloudProvider) Zones() (cloudprovider.Zones, bool) {
	return k, true
}

func (k *KrankCloudProvider) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

func (k *KrankCloudProvider) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

func (k *KrankCloudProvider) ProviderName() string {
	return ProviderName
}

func (k *KrankCloudProvider) ScrubDNS(nameservers, searches []string) (nsOut, srchOut []string) {
	return nameservers, searches
}

func (k *KrankCloudProvider) GetZone() (cloudprovider.Zone, error) {
	return cloudprovider.Zone{
		FailureDomain: "FailureDomain",
		Region:        "Region",
	}, nil
}
