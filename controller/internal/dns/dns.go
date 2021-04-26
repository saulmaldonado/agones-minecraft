package dns

import "fmt"

func JoinARecordName(domain, podName string) string {
	return fmt.Sprintf("%s.%s.", podName, domain)
}

func JoinSrvRecordName(domain, podName string) string {
	return fmt.Sprintf("_minecraft._tcp.%s.%s.", podName, domain)
}

func JoinSrvRR(srvRecordName string, port uint16, priority int, weight int, aRecordName string) string {
	return fmt.Sprintf("%d %d %d %s", priority, weight, port, aRecordName)
}
