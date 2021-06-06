package agones

import (
	"context"
	"fmt"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"agones.dev/agones/pkg/client/clientset/versioned"
	v1Informers "agones.dev/agones/pkg/client/informers/externalversions/agones/v1"
	"github.com/google/uuid"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"agones-minecraft/config"
	gamev1Model "agones-minecraft/models/v1/game"
	"agones-minecraft/resource/api/v1/game"
	k8s "agones-minecraft/services/k8s"
)

const (
	// GameServer labels

	JavaEdition    string = "java"
	BedrockEdition string = "bedrock"

	// GameServer Spec

	DefaultGenerateName            string = "mc-server-"
	DefaultGameServerContainerName string = "mc-server"
	DefaultJavaContainerPort       int32  = 25565
	DefaultBedrockContainerPort    int32  = 19132

	// Health

	DefaultInitialDelay     int32 = 60
	DefaultPeriodSeconds    int32 = 12
	DefaultFailureThreshold int32 = 5

	// Pod Template

	DefaultJavaImage     string = "itzg/minecraft-server"
	DefaultBedrockImage  string = "saulmaldonado/minecraft-bedrock-server"
	DefaultRCONPort      int32  = 25575
	DefaultRCONPassword  string = "minecraft"
	DefaultDataDirectory string = "/data"

	// mc-monitor

	MCMonitorImageName string = "saulmaldonado/agones-mc"

	// mc-backup

	MCBackupImageName         string = "saulmaldonado/agones-mc"
	MCBackupDefaultBucketName string = "agones-minecraft-mc-worlds"
	DefaultMCBackupCron       string = "0 */6 * * *"

	// volumes

	DefaultDataVolumeName string = "world-vol"

	// annotations

	HostnameAnnotation        string = "external-dns.alpha.kubernetes.io/hostname"
	SRVServiceAnnotation      string = "external-dns.alpha.kubernetes.io/gameserver-service"
	JavaSRVServiceName        string = "minecraft"
	CustomSubdomainAnnotation string = "external-dns.alpha.kubernetes.io/gameserver-subdomain"

	// labels

	EditionLabel string = "edition"
	UserIdLabel  string = "userId"
)

var agonesClient *AgonesClient

// Agones clientset wrapper
type AgonesClient struct {
	clientSet   *versioned.Clientset
	informer    v1Informers.GameServerInformer
	recordStore GameServerDNSRecordStore
}

// Initializes Agones client
func Init() {
	c, err := New(k8s.GetConfig())
	if err != nil {
		zap.L().Fatal("error initializing agones client", zap.Error(err))
	}
	go c.informer.Informer().Run(wait.NeverStop)
	cache.WaitForCacheSync(wait.NeverStop, c.informer.Informer().HasSynced)
	agonesClient = c
}

// Gets initialized Agones client
func Client() *AgonesClient {
	return agonesClient
}

// Creates new Agones client
func New(config *rest.Config) (*AgonesClient, error) {
	agonesClient, err := versioned.NewForConfig(config)
	gameServerInformer := NewGameServerInformer(agonesClient)
	recordStore := NewGameServerDNSRecordStore(gameServerInformer)
	if err != nil {
		return nil, err
	}

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

func (c *AgonesClient) HostnameAvailable(domain, subdomain string) bool {
	hostname := fmt.Sprintf("%s.%s", subdomain, domain)
	_, ok := c.recordStore.Get(hostname)
	return !ok
}

func (c *AgonesClient) ListRecords() []string {
	return c.recordStore.List()
}

func GetEdition(gs *agonesv1.GameServer) gamev1Model.Edition {
	l := gs.GetLabels()
	return gamev1Model.Edition(l[EditionLabel])
}

func SetHostname(gs *agonesv1.GameServer, domain, subdomain string) {
	anno := gs.GetAnnotations()
	anno[CustomSubdomainAnnotation] = subdomain
	anno[HostnameAnnotation] = domain
	gs.SetAnnotations(anno)
}

func GetDomainName(gs *agonesv1.GameServer) string {
	return gs.Annotations[HostnameAnnotation]
}

func GetSubdomain(gs *agonesv1.GameServer) (string, bool) {
	v, ok := gs.Annotations[CustomSubdomainAnnotation]
	return v, ok
}

func GetHostname(gs *agonesv1.GameServer) string {
	subdomain := gs.Name
	if domain, ok := gs.Annotations[CustomSubdomainAnnotation]; ok {
		subdomain = domain
	}
	domain := gs.Annotations[HostnameAnnotation]
	return fmt.Sprintf("%s.%s", subdomain, domain)
}

func GetStatus(gs *agonesv1.GameServer) game.Status {
	if IsOnline(gs) {
		return game.Online
	} else if IsStarting(gs) {
		return game.Starting
	}
	return game.Stopping
}

func GetDNSZone() string {
	return config.GetDNSZone()
}

func GetUserId(gs *agonesv1.GameServer) string {
	return gs.Labels[UserIdLabel]
}

func SetUserId(gs *agonesv1.GameServer, userId uuid.UUID) {
	gs.Labels[UserIdLabel] = userId.String()
}

func GetPort(gs *agonesv1.GameServer) int32 {
	if IsBeforePodCreated(gs) {
		return 0
	}
	return gs.Status.Ports[0].Port
}

func IsStarting(gs *agonesv1.GameServer) bool {
	state := gs.Status.State
	return IsBeforePodCreated(gs) ||
		state == agonesv1.GameServerStateScheduled ||
		state == agonesv1.GameServerStateRequestReady
}

func IsOnline(gs *agonesv1.GameServer) bool {
	state := gs.Status.State
	return state == agonesv1.GameServerStateReady || state == agonesv1.GameServerStateAllocated
}

func IsBeforePodCreated(gs *agonesv1.GameServer) bool {
	state := gs.Status.State
	return state == agonesv1.GameServerStatePortAllocation ||
		state == agonesv1.GameServerStateCreating ||
		state == agonesv1.GameServerStateStarting
}

// Initializes a new default Java Minecraft server
// Agones v1 GameServer object
func NewJavaServer() *agonesv1.GameServer {
	return &agonesv1.GameServer{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: DefaultGenerateName,
			Namespace:    metav1.NamespaceDefault,
			Annotations: map[string]string{
				HostnameAnnotation:   GetDNSZone(),
				SRVServiceAnnotation: JavaSRVServiceName,
			},
			Labels: map[string]string{
				EditionLabel: JavaEdition,
			},
		},
		Spec: agonesv1.GameServerSpec{
			Container: DefaultGameServerContainerName,
			Ports: []agonesv1.GameServerPort{{
				ContainerPort: DefaultJavaContainerPort,
				Name:          "mc",
				PortPolicy:    agonesv1.Dynamic,
				Protocol:      corev1.ProtocolTCP,
			}},
			Health: agonesv1.Health{
				InitialDelaySeconds: DefaultInitialDelay,
				PeriodSeconds:       DefaultPeriodSeconds,
				FailureThreshold:    DefaultFailureThreshold,
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            DefaultGameServerContainerName,
							Image:           DefaultJavaImage,
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{Name: "EULA", Value: "TRUE"},
							},
							VolumeMounts: []corev1.VolumeMount{
								{MountPath: DefaultDataDirectory, Name: DefaultDataVolumeName},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 25575},
							},
						},
						{
							Name:  "mc-monitor",
							Image: MCMonitorImageName,
							Args: []string{
								"monitor",
								"--attempts=" + fmt.Sprint(DefaultFailureThreshold),
								fmt.Sprintf("--initial-delay=%ds", DefaultInitialDelay),
								fmt.Sprintf("--interval=%ds", DefaultPeriodSeconds-2),
								fmt.Sprintf("--timeout=%ds", DefaultPeriodSeconds-2),
							},
							ImagePullPolicy: corev1.PullAlways,
						},
						{
							Name:  "mc-backup",
							Image: MCMonitorImageName,
							Args: []string{
								"backup",
								"--gcp-bucket-name=" + MCBackupDefaultBucketName,
								"--backup-cron=" + DefaultMCBackupCron,
								fmt.Sprintf("--initial-delay=%ds", DefaultInitialDelay),
							},
							Env: []corev1.EnvVar{
								{
									Name: "NAME", ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name:  "RCON_PASSWORD",
									Value: DefaultRCONPassword,
								},
							},
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: DefaultDataDirectory,
									Name:      DefaultDataVolumeName,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: DefaultDataVolumeName,
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
}

func NewBedrockServer() *agonesv1.GameServer {
	return &agonesv1.GameServer{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: DefaultGenerateName,
			Namespace:    metav1.NamespaceDefault,
			Annotations: map[string]string{
				HostnameAnnotation:   GetDNSZone(),
				SRVServiceAnnotation: JavaSRVServiceName,
			},
			Labels: map[string]string{
				EditionLabel: BedrockEdition,
			},
		},
		Spec: agonesv1.GameServerSpec{
			Container: DefaultGameServerContainerName,
			Ports: []agonesv1.GameServerPort{{
				ContainerPort: DefaultBedrockContainerPort,
				Name:          "mc",
				PortPolicy:    agonesv1.Dynamic,
				Protocol:      corev1.ProtocolUDP,
			}},
			Health: agonesv1.Health{
				InitialDelaySeconds: DefaultInitialDelay,
				PeriodSeconds:       DefaultPeriodSeconds,
				FailureThreshold:    DefaultFailureThreshold,
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            DefaultGameServerContainerName,
							Image:           DefaultBedrockImage,
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{Name: "EULA", Value: "TRUE"},
							},
							VolumeMounts: []corev1.VolumeMount{
								{MountPath: DefaultDataDirectory, Name: DefaultDataVolumeName},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 25575},
							},
						},
						{
							Name:  "mc-monitor",
							Image: MCMonitorImageName,
							Args: []string{
								"monitor",
								"--attempts=" + fmt.Sprint(DefaultFailureThreshold),
								fmt.Sprintf("--initial-delay=%ds", DefaultInitialDelay),
								fmt.Sprintf("--interval=%ds", DefaultPeriodSeconds-2),
								fmt.Sprintf("--timeout=%ds", DefaultPeriodSeconds-2),
								"--port=" + fmt.Sprint(DefaultBedrockContainerPort),
								"--edition=" + BedrockEdition,
							},
							ImagePullPolicy: corev1.PullAlways,
						},
						{
							Name:  "mc-backup",
							Image: MCMonitorImageName,
							Args: []string{
								"backup",
								"--gcp-bucket-name=" + MCBackupDefaultBucketName,
								"--backup-cron=" + DefaultMCBackupCron,
								fmt.Sprintf("--initial-delay=%ds", DefaultInitialDelay),
								"--edition=" + BedrockEdition,
							},
							Env: []corev1.EnvVar{
								{
									Name: "NAME", ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name:  "RCON_PASSWORD",
									Value: DefaultRCONPassword,
								},
							},
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: DefaultDataDirectory,
									Name:      DefaultDataVolumeName,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: DefaultDataVolumeName,
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
}
