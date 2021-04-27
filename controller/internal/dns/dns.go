package dns

import (
	"fmt"
	"strings"
)

const (
	Service  string = "_minecraft"
	Protocol string = "_tcp"
)

func JoinARecordName(domain, podName string) string {
	return fmt.Sprintf("%s.%s", podName, EnsureTrailingDot(domain))
}

func JoinSrvRecordName(domain, podName string) string {
	return fmt.Sprintf("%s.%s.%s.%s", Service, Protocol, podName, EnsureTrailingDot(domain))
}

func JoinSrvRR(srvRecordName string, port uint16, priority int, weight int, aRecordName string) string {
	return fmt.Sprintf("%d %d %d %s", priority, weight, port, aRecordName)
}

func EnsureTrailingDot(hostname string) string {
	return strings.TrimSuffix(hostname, ".") + "."
}
