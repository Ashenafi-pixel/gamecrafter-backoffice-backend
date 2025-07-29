package utils

import (
	"log"
	"net"
)

func ValidateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}

func CheckIPInRanage(ip, startIP, endIP string) bool {
	startIPParsed := net.ParseIP(startIP)
	endIPParsed := net.ParseIP(endIP)
	ipParsed := net.ParseIP(ip)
	if startIPParsed == nil || endIPParsed == nil || ipParsed == nil {
		log.Printf("invalid ip given startIP %s ,endIP %s IP %s", startIP, endIP, ip)
		return false
	}
	if compare(ipParsed, startIPParsed) >= 0 && compare(ipParsed, endIPParsed) <= 0 {
		return true
	}
	return false
}
func compare(ip1, ip2 net.IP) int {
	if len(ip1) != len(ip2) {
		return len(ip1) - len(ip2)
	}
	for i := 0; i < len(ip1); i++ {
		if ip1[i] < ip2[i] {
			return -1
		} else if ip1[i] > ip2[i] {
			return 1
		}
	}
	return 0
}
