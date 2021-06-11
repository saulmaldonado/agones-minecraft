package agones

import (
	"context"
	"fmt"
	"time"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"agones.dev/agones/pkg/client/clientset/versioned"
	v1Informers "agones.dev/agones/pkg/client/informers/externalversions/agones/v1"
	"github.com/google/uuid"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	gamev1Resource "agones-minecraft/resource/api/v1/game"
	k8s "agones-minecraft/services/k8s"
)

var agonesClient *AgonesClient

// Agones clientset wrapper
type AgonesClient struct {
	clientSet   *versioned.Clientset
	informer    v1Informers.GameServerInformer
	recordStore GameServerDNSRecordStore
}

// Timeout for connecting to k8s server
var DefaultTimeout time.Duration = time.Second * 30

// Initializes Agones client
func Init() {
	c, err := New(k8s.GetConfig())
	if err != nil {
		zap.L().Fatal("error initializing agones client", zap.Error(err))
	}

	go c.informer.Informer().Run(wait.NeverStop)

	stop := make(chan struct{})

	time.AfterFunc(DefaultTimeout, func() {
		close(stop)
	})

	if !cache.WaitForCacheSync(stop, c.informer.Informer().HasSynced) {
		zap.L().Fatal("error syncing cache")
	}
	agonesClient = c
}

// Gets initialized Agones client
func Client() *AgonesClient {
	return agonesClient
}

// Creates new Agones client
func New(config *rest.Config) (*AgonesClient, error) {
	agonesClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	gameServerInformer := NewGameServerInformer(agonesClient)
	recordStore := NewGameServerDNSRecordStore(gameServerInformer)

	return &AgonesClient{agonesClient, gameServerInformer, recordStore}, nil
}

// Gets a GameServer by name
func (c *AgonesClient) Get(serverName string) (*agonesv1.GameServer, error) {
	return c.informer.Lister().GameServers(metav1.NamespaceDefault).Get(serverName)
}

// Gets all GameServers for default namespace
func (c *AgonesClient) List() ([]*agonesv1.GameServer, error) {
	return c.informer.Lister().GameServers(metav1.NamespaceDefault).List(labels.Everything())
}

// Creates a new GameServer
func (c *AgonesClient) Create(server *agonesv1.GameServer) (*agonesv1.GameServer, error) {
	return c.clientSet.
		AgonesV1().
		GameServers(metav1.NamespaceDefault).
		Create(context.Background(), server, metav1.CreateOptions{})
}

// Creates a new GameServer with dry-run option enabled
func (c *AgonesClient) CreateDryRun(server *agonesv1.GameServer) (*agonesv1.GameServer, error) {
	return c.clientSet.
		AgonesV1().
		GameServers(metav1.NamespaceDefault).
		Create(context.Background(), server, metav1.CreateOptions{
			DryRun: []string{"All"},
		})
}

// Deletes a GameServer by name
func (c *AgonesClient) Delete(serverName string) error {
	return c.clientSet.
		AgonesV1().
		GameServers(metav1.NamespaceDefault).
		Delete(context.Background(), serverName, metav1.DeleteOptions{})
}

func (c *AgonesClient) ListRecords() []string {
	return c.recordStore.List()
}

func (c *AgonesClient) HostnameAvailable(domain, subdomain string) bool {
	hostname := fmt.Sprintf("%s.%s", subdomain, domain)
	_, ok := c.recordStore.Get(hostname)
	return !ok
}

func GetGameStatusByName(game *gamev1Resource.GameStatus, name string) error {
	gs, err := Client().Get(name)
	if err != nil {
		return err
	}
	hostname := GetHostname(gs)
	userId := uuid.MustParse(GetUserId(gs))

	*game = gamev1Resource.GameStatus{
		ID:       GetUID(gs),
		UserID:   userId,
		Name:     gs.Name,
		Status:   GetStatus(gs),
		Edition:  GetEdition(gs),
		Hostname: &hostname,
		Address:  GetAddress(gs),
		Port:     GetPort(gs),
	}

	return nil
}
