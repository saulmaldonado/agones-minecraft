package agones

import (
	"context"
	"fmt"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"agones.dev/agones/pkg/client/clientset/versioned"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

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
)

var agonesClient *AgonesClient

// Agones clientset wrapper
type AgonesClient struct {
	clientSet *versioned.Clientset
}

// Initializes Agones client
func Init() {
	c, err := New(k8s.GetConfig())
	if err != nil {
		zap.L().Fatal("error initializing agones client", zap.Error(err))
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

	return &AgonesClient{agonesClient}, nil
}

// Gets a GameServer by name
func (c *AgonesClient) Get(serverName string) (*agonesv1.GameServer, error) {
	return c.clientSet.
		AgonesV1().
		GameServers(metav1.NamespaceDefault).
		Get(context.Background(), serverName, metav1.GetOptions{})
}

// Gets all GameServers for default namespace
func (c *AgonesClient) List() (*agonesv1.GameServerList, error) {
	return c.clientSet.
		AgonesV1().
		GameServers(metav1.NamespaceDefault).
		List(context.Background(), metav1.ListOptions{})
}

// Creates a new GameServer
func (c *AgonesClient) Create(server *agonesv1.GameServer) (*agonesv1.GameServer, error) {
	return c.clientSet.
		AgonesV1().
		GameServers(metav1.NamespaceDefault).
		Create(context.Background(), server, metav1.CreateOptions{})
}

// Deletes a GameServer by name
func (c *AgonesClient) Delete(serverName string) error {
	return c.clientSet.
		AgonesV1().
		GameServers(metav1.NamespaceDefault).
		Delete(context.Background(), serverName, metav1.DeleteOptions{})
}

// Initializes a new default Java Minecraft server
// Agones v1 GameServer object
func NewJavaServer() *agonesv1.GameServer {
	return &agonesv1.GameServer{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: DefaultGenerateName,
			Namespace:    metav1.NamespaceDefault,
			Labels: map[string]string{
				"edition": JavaEdition,
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
			Labels: map[string]string{
				"edition": BedrockEdition,
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
