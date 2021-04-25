package controller

import (
	"fmt"
	"strings"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
)

func getHostnameAnnotation(gs *agonesv1.GameServer) (string, bool) {
	return getAnnotation(HostnameAnnotation, gs)
}

func setExternalDnsAnnotation(recordName string, gs *agonesv1.GameServer) string {
	setAnnotation(ExternalDnsAnnotation, recordName, gs)
	return recordName
}

func findeExternalDnsAnnotation(gs *agonesv1.GameServer) bool {
	_, ok := getAnnotation(ExternalDnsAnnotation, gs)
	return ok
}

func getAnnotation(suffix string, gs *agonesv1.GameServer) (string, bool) {
	annotation := fmt.Sprintf("%s/%s", AnnotationPrefix, suffix)
	hostname, ok := gs.Annotations[annotation]

	if !ok || strings.TrimSpace(hostname) == "" {
		return "", false
	}

	return hostname, true
}

func setAnnotation(suffix string, value string, gs *agonesv1.GameServer) {
	annotation := fmt.Sprintf("%s/%s", AnnotationPrefix, ExternalDnsAnnotation)
	gs.Annotations[annotation] = value
}
