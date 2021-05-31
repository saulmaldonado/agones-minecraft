package k8s

import (
	"flag"
	"log"
	"path/filepath"

	"agones.dev/agones/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type AppK8sClient struct {
	K8s    *kubernetes.Clientset
	Agones *versioned.Clientset
}

var k8sClient *AppK8sClient

func Init() {
	client, err := New()
	if err != nil {
		log.Fatal(err)
	}
	k8sClient = client
}

func New() (*AppK8sClient, error) {
	var kubeconfig *string
	// kubeconfig path defaults to ~/.kube/config
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	agonesClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &AppK8sClient{
		K8s:    clientset,
		Agones: agonesClient,
	}, nil
}

func Get() *AppK8sClient {
	return k8sClient
}
