package scanner

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
)

//var execCommand = exec.Command  // <--- For unit tests

type Device struct {
	IP       string
	MAC      string
	Hostname string
}

// ScanCIDR runs a basic ping sweep over a subnet and returns live hosts.
func ScanCIDR(cidr string) []Device {
	ipList := getIPsInRange(cidr)
	var devices []Device
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, ip := range ipList {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			// Ping host once, short timeout
			cmd := exec.Command("ping", "-n", "1", "-w", "200", ip)
			output, err := cmd.CombinedOutput()
			if err == nil && strings.Contains(string(output), "TTL=") {
				hostname := reverseLookup(ip)
				mac := getMac(ip)
				mu.Lock()
				devices = append(devices, Device{
					IP:       ip,
					MAC:      mac,
					Hostname: hostname,
				})
				mu.Unlock()
			}
		}(ip)
	}

	wg.Wait()
	return devices
}

// getIPsInRange returns all IPs within the given subnet.
func getIPsInRange(cidr string) []string {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		fmt.Println("Invalid CIDR:", err)
		return nil
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	return ips
}

// inc increments an IP address by one.
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// reverseLookup tries to resolve the hostname for a given IP.
func reverseLookup(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}

// getMac checks the ARP table for a MAC entry matching the IP.
func getMac(ip string) string {
	cmd := exec.Command("arp", "-a", ip)
	output, _ := cmd.CombinedOutput()
	out := string(output)
	if strings.Contains(out, ip) {
		fields := strings.Fields(out)
		for _, f := range fields {
			if strings.Count(f, "-") == 5 {
				return f
			}
		}
	}
	return ""
}
