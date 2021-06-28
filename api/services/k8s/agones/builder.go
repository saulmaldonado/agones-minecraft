package agones

import (
	"agones-minecraft/config"

	"github.com/google/uuid"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	corev1 "k8s.io/api/core/v1"
)

type MCServerDirector struct {
	builder MCServerBuilder
}

func NewDirector(builder MCServerBuilder) *MCServerDirector {
	return &MCServerDirector{
		builder: builder,
	}
}

func (d *MCServerDirector) BuildServer(name string, subdomain string, uuid uuid.UUID, userId uuid.UUID) *agonesv1.GameServer {
	d.builder.SetName(name)
	d.builder.SetAddress(subdomain)
	d.builder.SetUUID(uuid)
	d.builder.SetUserID(userId)

	return d.builder.GetServer()
}

type MCServerBuilder interface {
	SetName(string)
	SetAddress(string)
	SetUserID(uuid.UUID)
	SetUUID(uuid.UUID)
	GetServer() *agonesv1.GameServer
}

type JavaServerBuilder struct {
	Name    string
	Address string
	UserId  uuid.UUID
	UUID    uuid.UUID
}

func NewJavaServerBuilder() *JavaServerBuilder {
	return &JavaServerBuilder{}
}

func (j *JavaServerBuilder) SetName(name string) {
	j.Name = name
}

func (j *JavaServerBuilder) SetAddress(address string) {
	j.Address = address
}

func (j *JavaServerBuilder) SetUserID(userId uuid.UUID) {
	j.UserId = userId
}

func (j *JavaServerBuilder) SetUUID(uuid uuid.UUID) {
	j.UUID = uuid
}

func (j *JavaServerBuilder) GetServer() *agonesv1.GameServer {
	gs := newServer()
	SetHostname(&gs, config.GetDNSZone(), j.Address)
	SetName(&gs, j.UserId, j.Name)
	SetUserId(&gs, j.UserId)
	SetUUID(&gs, j.UUID)
	SetEdition(&gs, JavaEdition)

	gs.Annotations[SRVServiceAnnotation] = JavaSRVServiceName
	gs.Spec.Ports[0].Protocol = corev1.ProtocolTCP
	gs.Spec.Template.Spec.Containers[0].Image = DefaultJavaImage
	gs.Spec.Ports[0].ContainerPort = DefaultJavaContainerPort

	for _, c := range gs.Spec.Template.Spec.Containers {
		if c.Name == "mc-monitor" {
			c.Env = append(c.Env, corev1.EnvVar{
				Name: edition, Value: string(JavaEdition),
			},
				corev1.EnvVar{
					Name: port, Value: string(DefaultJavaContainerPort),
				},
			)
		}

		if c.Name == "mc-backup" {
			c.Env = append(c.Env, corev1.EnvVar{
				Name: edition, Value: string(JavaEdition),
			})
		}

	}
	return &gs
}

type BedrockServerBuilder struct {
	Name    string
	Address string
	UserId  uuid.UUID
	UUID    uuid.UUID
}

func NewBedrockServerBuilder() *BedrockServerBuilder {
	return &BedrockServerBuilder{}
}

func (j *BedrockServerBuilder) SetName(name string) {
	j.Name = name
}

func (j *BedrockServerBuilder) SetAddress(address string) {
	j.Address = address
}

func (j *BedrockServerBuilder) SetUserID(userId uuid.UUID) {
	j.UserId = userId
}

func (j *BedrockServerBuilder) SetUUID(uuid uuid.UUID) {
	j.UUID = uuid
}

func (j *BedrockServerBuilder) GetServer() *agonesv1.GameServer {
	gs := newServer()
	SetHostname(&gs, config.GetDNSZone(), j.Address)
	SetName(&gs, j.UserId, j.Name)
	SetUserId(&gs, j.UserId)
	SetUUID(&gs, j.UUID)
	SetEdition(&gs, BedrockEdition)

	gs.Spec.Ports[0].Protocol = corev1.ProtocolUDP
	gs.Spec.Template.Spec.Containers[0].Image = DefaultBedrockImage
	gs.Spec.Ports[0].ContainerPort = DefaultBedrockContainerPort

	for _, c := range gs.Spec.Template.Spec.Containers {
		if c.Name == "mc-monitor" {
			c.Env = append(c.Env, corev1.EnvVar{
				Name: edition, Value: string(BedrockEdition),
			},
				corev1.EnvVar{
					Name: port, Value: string(DefaultBedrockContainerPort),
				},
			)
		}

		if c.Name == "mc-backup" {
			c.Env = append(c.Env, corev1.EnvVar{
				Name: edition, Value: string(BedrockEdition),
			})
		}

	}
	return &gs
}
