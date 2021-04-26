package controller

import (
	"fmt"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getHostnameAnnotation(obj client.Object) (string, bool) {
	return getAnnotation(HostnameAnnotation, obj)
}

func setExternalDnsAnnotation(recordName string, obj client.Object) string {
	setAnnotation(ExternalDnsAnnotation, recordName, obj)
	return recordName
}

func findExternalDnsAnnotation(obj client.Object) bool {
	_, ok := getAnnotation(ExternalDnsAnnotation, obj)
	return ok
}

func getAnnotation(suffix string, obj client.Object) (string, bool) {
	key := fmt.Sprintf("%s/%s", AnnotationPrefix, suffix)
	annotations := obj.GetAnnotations()

	hostname, ok := annotations[key]

	if !ok || strings.TrimSpace(hostname) == "" {
		return "", false
	}

	return hostname, true
}

func setAnnotation(suffix string, value string, obj client.Object) {
	key := fmt.Sprintf("%s/%s", AnnotationPrefix, ExternalDnsAnnotation)
	annotations := obj.GetAnnotations()
	annotations[key] = value
}
