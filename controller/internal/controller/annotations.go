package controller

import (
	"fmt"
	"strings"

	"github.com/saulmaldonado/agones-minecraft/controller/internal/dns"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AnnotationPrefix      string = "agones-mc"
	DomainAnnotation      string = "domain"
	ExternalDnsAnnotation string = "externalDNS"
)

func getDomainAnnotationOrLabel(obj client.Object) (string, bool) {
	if domain, found := getAnnotation(DomainAnnotation, obj); found && dns.IsDnsName(domain) {
		return dns.EnsureTrailingDot(domain), found
	}

	if domain, found := getLabel(DomainAnnotation, obj); found {
		return dns.EnsureTrailingDot(domain), found
	}

	return "", false
}

func setExternalDnsAnnotation(recordName string, obj client.Object) string {
	recordName = dns.EnsureTrailingDot(recordName)
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

	domain, ok := annotations[key]

	if !ok || strings.TrimSpace(domain) == "" {
		return "", false
	}

	return domain, true
}

func getLabel(suffix string, obj client.Object) (string, bool) {
	key := fmt.Sprintf("%s/%s", AnnotationPrefix, suffix)
	labels := obj.GetLabels()

	domain, ok := labels[key]

	if !ok || strings.TrimSpace(domain) == "" {
		return "", false
	}

	return domain, true
}

func setAnnotation(suffix string, value string, obj client.Object) {

	key := fmt.Sprintf("%s/%s", AnnotationPrefix, ExternalDnsAnnotation)
	annotations := obj.GetAnnotations()
	annotations[key] = value
}
