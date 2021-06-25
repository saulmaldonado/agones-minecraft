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
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	k8s "agones-minecraft/services/k8s"
)

var agonesClient *AgonesClient

// Agones clientset wrapper
type AgonesClient struct {
	clientSet *versioned.Clientset
	informer  v1Informers.GameServerInformer
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

	return &AgonesClient{agonesClient, gameServerInformer}, nil
}

func (c *AgonesClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	res, err := c.clientSet.RESTClient().Get().AbsPath("/healthz").DoRaw(ctx)
	if err != nil {
		return err
	}

	content := string(res)
	if content != "ok" {
		return fmt.Errorf("not ok, error: %s", content)
	}

	return nil
}

// Gets a GameServer by name
func (c *AgonesClient) Get(serverName string) (*agonesv1.GameServer, error) {
	return c.informer.Lister().GameServers(metav1.NamespaceDefault).Get(serverName)
}

// Get gameserver by name and check if the provided userId matches the userId label on the resource
// If user does not match label on resourece it returns a not found k8s api error
func (c *AgonesClient) GetForUser(serverName string, userId uuid.UUID) (*agonesv1.GameServer, error) {
	gs, err := c.Get(serverName)

	if GetUserId(gs) != userId.String() {
		return nil, k8sErrors.NewNotFound(agonesv1.Resource("GameServer"), serverName)
	}

	return gs, err
}

// Gets all GameServers for default namespace
func (c *AgonesClient) List() ([]*agonesv1.GameServer, error) {
	return c.informer.Lister().GameServers(metav1.NamespaceDefault).List(labels.Everything())
}

func (c *AgonesClient) ListGamesForUser(userId string) ([]*agonesv1.GameServer, error) {
	req, err := labels.NewRequirement(UserIdLabel, selection.Equals, []string{userId})
	if err != nil {
		return nil, err
	}

	return c.informer.Lister().
		GameServers(metav1.NamespaceDefault).
		List(labels.NewSelector().Add(*req))
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
