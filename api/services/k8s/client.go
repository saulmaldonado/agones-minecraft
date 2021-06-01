package k8s

import (
	"flag"
	"path/filepath"

	"go.uber.org/zap"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var config *rest.Config

func InitConfig() {
	cfg, err := NewConfig()
	if err != nil {
		zap.L().Fatal("error initializing k8s config", zap.Error(err))
	}
	config = cfg
}

func NewConfig() (*rest.Config, error) {
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
	return config, nil
}

func GetConfig() *rest.Config {
	return config
}
