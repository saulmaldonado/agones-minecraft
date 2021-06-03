package agones

import (
	"sync"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"agones.dev/agones/pkg/client/clientset/versioned"
	"agones.dev/agones/pkg/client/informers/externalversions"
	v1Informers "agones.dev/agones/pkg/client/informers/externalversions/agones/v1"
	"k8s.io/client-go/tools/cache"
)

type GameServerDNSRecordStore interface {
	Get(domain string) (string, bool)
	Set(domain string, serverName string)
	Delete(domain string)
	List() []string
}

type gameServerDNSRecordStore struct {
	records map[string]string
	lock    sync.RWMutex
}

func NewGameServerInformer(clientset *versioned.Clientset) v1Informers.GameServerInformer {
	return externalversions.NewSharedInformerFactory(clientset, 0).Agones().V1().GameServers()
}

func NewGameServerDNSRecordStore(informer v1Informers.GameServerInformer) GameServerDNSRecordStore {
	store := gameServerDNSRecordStore{
		lock:    sync.RWMutex{},
		records: make(map[string]string),
	}

	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			gs := obj.(*agonesv1.GameServer)
			fullDomain := GetHostname(gs)
			store.Set(fullDomain, gs.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newGs := oldObj.(*agonesv1.GameServer)
			oldGs := newObj.(*agonesv1.GameServer)
			newDomain := GetHostname(newGs)
			oldDomain := GetHostname(oldGs)
			if oldDomain != newDomain {
				store.Delete(oldDomain)
				store.Set(newDomain, newGs.Name)
			}
		},
		DeleteFunc: func(obj interface{}) {
			gs := obj.(*agonesv1.GameServer)
			store.Delete(GetHostname(gs))
		},
	})

	return &store
}

func (s *gameServerDNSRecordStore) Get(domain string) (string, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	v, ok := s.records[domain]
	return v, ok
}

func (s *gameServerDNSRecordStore) Set(domain string, serverName string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.records[domain] = serverName
}

func (s *gameServerDNSRecordStore) Delete(domain string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.records, domain)
}

func (s *gameServerDNSRecordStore) List() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	list := make([]string, 0, len(s.records))
	for rec := range s.records {
		list = append(list, rec)
	}
	return list
}
