package agones

import (
	"fmt"
	"strings"
	"time"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"agones-minecraft/config"

	gamev1Model "agones-minecraft/models/v1/game"
)

type Edition string

const (
	// Minecraft java server edition
	JavaEdition Edition = "java"
	// Minecraft bedrock sever edition
	BedrockEdition Edition = "bedrock"

	// GameServer Spec
	DefaultBedrockContainerPort int32 = 19132
	DefaultJavaContainerPort    int32 = 25565

	DefaultGenerateName            string = "mc-server-"
	DefaultGameServerContainerName string = "mc-server"

	DefaultBedrockImage string = "saulmaldonado/minecraft-bedrock-server"
	DefaultJavaImage    string = "itzg/minecraft-server"
	JavaSRVServiceName  string = "minecraft"

	// health

	DefaultInitialDelaySeconds int32 = 60
	DefaultHealthPeriodSeconds int32 = 12

	// Pod Template

	DefaultRCONPort      int32  = 25575
	DefaultRCONPassword  string = "minecraft"
	DefaultDataDirectory string = "/data"

	// mc-monitor

	MCMonitorImageName string = "saulmaldonado/agones-mc"

	// mc-backup

	MCBackupImageName   string = "saulmaldonado/agones-mc"
	DefaultMCBackupCron string = "0 */6 * * *"

	// volumes

	DefaultDataVolumeName string = "world-vol"

	// annotations

	HostnameAnnotation        string = "external-dns.alpha.kubernetes.io/hostname"
	SRVServiceAnnotation      string = "external-dns.alpha.kubernetes.io/gameserver-service"
	CustomSubdomainAnnotation string = "external-dns.alpha.kubernetes.io/gameserver-subdomain"

	// labels

	EditionLabel string = "edition"
	UserIdLabel  string = "userId"
	UUIDLabel    string = "uuid"

	// env names

	initialDelay string = "INITIAL_DELAY"
	host         string = "HOST"
	port         string = "PORT"
	edition      string = "EDITION"
	interval     string = "INTERVAL"
	timeout      string = "TIMEOUT"
	maxAttempts  string = "MAX_ATTEMPTS"
	podName      string = "POD_NAME"
	rconPassword string = "RCON_PASSWORD"
	rconPort     string = "RCON_PORT"
	backupCron   string = "BACKUP_CRON"
	bucketName   string = "BUCKET_NAME"
)

var (
	// health

	DefaultInitialDelay     time.Duration = time.Second * (2)
	DefaultFailureThreshold int32         = 5

	DefaultHealthInterval  time.Duration = (time.Second * time.Duration(DefaultHealthPeriodSeconds)) - (time.Second * 2)
	DefaultTimeoutDuration time.Duration = (time.Second * time.Duration(DefaultHealthPeriodSeconds)) - (time.Second * 2)
)

func NewAddress(subdomain string) string {
	domain := config.GetDNSZone()
	return fmt.Sprintf("%s.%s", subdomain, domain)
}

func GetAddress(gs *agonesv1.GameServer) string {
	domain := gs.Annotations[HostnameAnnotation]
	subdomain := gs.Annotations[CustomSubdomainAnnotation]
	return fmt.Sprintf("%s.%s", subdomain, domain)
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

func SetName(gs *agonesv1.GameServer, userId uuid.UUID, name string) {
	gs.Name = fmt.Sprintf("%s.%s", userId.String(), name)
}

func GetName(gs *agonesv1.GameServer) string {
	return strings.TrimPrefix(strings.TrimLeft(gs.Name, "0123456789abcdef-"), ".")
}

func GetState(gs *agonesv1.GameServer) gamev1Model.GameState {
	if IsOnline(gs) || IsStarting(gs) {
		return gamev1Model.On
	}
	return gamev1Model.Off
}

func GetStatus(gs *agonesv1.GameServer) *agonesv1.GameServerState {
	return &gs.Status.State
}

func SetUUID(gs *agonesv1.GameServer, uuid uuid.UUID) {
	gs.Labels[UUIDLabel] = uuid.String()
}

func GetUUID(gs *agonesv1.GameServer) uuid.UUID {
	return uuid.MustParse(gs.Labels[UUIDLabel])
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

func GetEdition(gs *agonesv1.GameServer) gamev1Model.Edition {
	return gamev1Model.Edition(gs.Annotations[EditionLabel])
}

func SetEdition(gs *agonesv1.GameServer, edition Edition) {
	gs.Annotations[EditionLabel] = string(edition)
}

func GetPort(gs *agonesv1.GameServer) *int32 {
	if IsBeforePodCreated(gs) {
		return nil
	}
	return &gs.Status.Ports[0].Port
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

func newServer() agonesv1.GameServer {
	return agonesv1.GameServer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   metav1.NamespaceDefault,
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		},
		Spec: agonesv1.GameServerSpec{
			Container: DefaultGameServerContainerName,
			Ports: []agonesv1.GameServerPort{{
				Name:       "mc",
				PortPolicy: agonesv1.Dynamic,
			}},
			Health: agonesv1.Health{
				InitialDelaySeconds: DefaultInitialDelaySeconds,
				PeriodSeconds:       DefaultHealthPeriodSeconds,
				FailureThreshold:    DefaultFailureThreshold,
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            DefaultGameServerContainerName,
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
							Args:  []string{"monitor"},

							Env: []corev1.EnvVar{
								{Name: maxAttempts, Value: string(DefaultFailureThreshold)},
								{Name: initialDelay, Value: DefaultInitialDelay.String()},
								{Name: interval, Value: DefaultHealthInterval.String()},
								{Name: timeout, Value: DefaultTimeoutDuration.String()},
							},

							ImagePullPolicy: corev1.PullAlways,
						},
						{
							Name:  "mc-backup",
							Image: MCMonitorImageName,
							Args:  []string{"backup"},
							Env: []corev1.EnvVar{
								{
									Name: podName, ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name:  rconPassword,
									Value: DefaultRCONPassword,
								},
								{Name: initialDelay, Value: DefaultInitialDelay.String()},
								{Name: backupCron, Value: DefaultMCBackupCron},
								{Name: bucketName, Value: config.GetBucketName()},
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
