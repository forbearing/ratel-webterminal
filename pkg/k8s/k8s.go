package k8s

import (
	"net/http"

	"github.com/forbearing/ratel-webterminal/pkg/args"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	restConfig      *rest.Config
	httpClient      *http.Client
	restClient      *rest.RESTClient
	clientset       *kubernetes.Clientset
	dynamicClient   dynamic.Interface
	discoveryClient *discovery.DiscoveryClient
	informerFactory informers.SharedInformerFactory
)

func Init() {
	var err error

	kubeconfig := args.GetKubeConfigFile()

	// creates rest config
	if len(kubeconfig) != 0 {
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
	}
	restConfig.APIPath = "api"
	restConfig.GroupVersion = &corev1.SchemeGroupVersion
	//restConfig.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	restConfig.NegotiatedSerializer = scheme.Codecs

	// create a http client for the given config.
	if httpClient, err = rest.HTTPClientFor(restConfig); err != nil {
		log.Fatal(err)
	}

	// create a RESTClient for the given config and http client.
	if restClient, err = rest.RESTClientForConfigAndClient(restConfig, httpClient); err != nil {
		log.Fatal(err)
	}

	// create a Clientset for the given config and http client.
	if clientset, err = kubernetes.NewForConfigAndClient(restConfig, httpClient); err != nil {
		log.Fatal(err)
	}
	// create a dynamic client for the given config and http client.
	if dynamicClient, err = dynamic.NewForConfigAndClient(restConfig, httpClient); err != nil {
		log.Fatal(err)
	}
	// create a DiscoveryClient for the given config and http client.
	if discoveryClient, err = discovery.NewDiscoveryClientForConfigAndClient(restConfig, httpClient); err != nil {
		log.Fatal(err)
	}
}

// RESTConfig returns the underlying rest config
func RESTConfig() *rest.Config {
	return restConfig
}

// RESTClient returns the underlying rest client.
func RESTClient() *rest.RESTClient {
	return restClient
}

// Clientset returns the underlying clientset.
func Clientset() *kubernetes.Clientset {
	return clientset
}

// DynamicClient returns the underlying dynamic client.
func DynamicClient() dynamic.Interface {
	return dynamicClient
}

// DiscoveryClient returns the underlying discovery client.
func DiscoveryClient() *discovery.DiscoveryClient {
	return discoveryClient
}
