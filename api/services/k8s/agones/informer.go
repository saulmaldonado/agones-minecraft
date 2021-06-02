package agones

import (
	"agones.dev/agones/pkg/client/clientset/versioned"
	"agones.dev/agones/pkg/client/informers/externalversions"
	v1 "agones.dev/agones/pkg/client/informers/externalversions/agones/v1"
)

func NewGameServerInformer(clientset *versioned.Clientset) v1.GameServerInformer {
	return externalversions.NewSharedInformerFactory(clientset, 0).Agones().V1().GameServers()
}
