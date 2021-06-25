package agones

import (
	"agones.dev/agones/pkg/client/clientset/versioned"
	"agones.dev/agones/pkg/client/informers/externalversions"
	v1Informers "agones.dev/agones/pkg/client/informers/externalversions/agones/v1"
)

type GameServerDNSRecordStore interface {
	Get(domain string) (string, bool)
	Set(domain string, serverName string)
	Delete(domain string)
	List() []string
}

func NewGameServerInformer(clientset *versioned.Clientset) v1Informers.GameServerInformer {
	return externalversions.NewSharedInformerFactory(clientset, 0).Agones().V1().GameServers()
}
