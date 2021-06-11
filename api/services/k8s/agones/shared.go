package agones

import (
	"fmt"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/google/uuid"

	"agones-minecraft/config"
	"agones-minecraft/resource/api/v1/game"

	gamev1Model "agones-minecraft/models/v1/game"
)

const (

	// GameServer Spec

	DefaultGenerateName            string = "mc-server-"
	DefaultGameServerContainerName string = "mc-server"

	// Health

	DefaultInitialDelay     int32 = 60
	DefaultPeriodSeconds    int32 = 12
	DefaultFailureThreshold int32 = 5

	// Pod Template

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
	CustomSubdomainAnnotation string = "external-dns.alpha.kubernetes.io/gameserver-subdomain"

	// labels

	EditionLabel string = "edition"
	UserIdLabel  string = "userId"
)

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

func GetUID(gs *agonesv1.GameServer) uuid.UUID {
	return uuid.MustParse(string(gs.UID))
}

func GetEdition(gs *agonesv1.GameServer) gamev1Model.Edition {
	return gamev1Model.Edition(gs.Annotations[EditionLabel])
}

func GetAddress(gs *agonesv1.GameServer) *string {
	if IsOnline(gs) {
		return &gs.Status.Address
	}
	return nil
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
