package dns

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

const (
	Service  string = "_minecraft"
	Protocol string = "_tcp"
	DnsName  string = `([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`
)

var rxDnsName = regexp.MustCompile(DnsName)

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

func IsDnsName(name string) bool {
	if name == "" || len(strings.Replace(name, ".", "", -1)) > 255 {
		return false
	}

	return (net.ParseIP(name) == nil) && rxDnsName.MatchString(name)
}
