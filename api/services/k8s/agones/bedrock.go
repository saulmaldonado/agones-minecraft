package agones

import (
	"fmt"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// GameServer labels

	BedrockEdition string = "bedrock"

	// GameServer spec

	DefaultBedrockContainerPort int32 = 19132

	// Pod template

	DefaultBedrockImage string = "saulmaldonado/minecraft-bedrock-server"
)

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
