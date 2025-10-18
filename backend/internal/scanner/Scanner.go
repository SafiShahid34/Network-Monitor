package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis client and context
var ctx = context.Background()
var rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

// execCommand can be overridden in tests
var execCommand = exec.Command

type Device struct {
	IP       string `json:"ip"`
	MAC      string `json:"mac"`
	Hostname string `json:"hostname"`
}

// ScanCIDR performs a ping sweep and caches results in Redis for 30 seconds.
func ScanCIDR(cidr string) []Device {
	cacheKey := fmt.Sprintf("scan:%s", cidr)

	// Try to load from cache first
	if cached, err := rdb.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
		var devices []Device
		if err := json.Unmarshal([]byte(cached), &devices); err == nil {
			fmt.Println("Loaded scan results from Redis cache")
			return devices
		}
	}

	// Otherwise, perform full network scan
	ipList := getIPsInRange(cidr)
	var devices []Device
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, ip := range ipList {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			cmd := execCommand("ping", "-n", "1", "-w", "200", ip)
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

	// Cache results for 30 seconds
	if len(devices) > 0 {
		data, _ := json.Marshal(devices)
		rdb.Set(ctx, cacheKey, data, 30*time.Second)
	}

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

// reverseLookup resolves a hostname for a given IP (if available).
func reverseLookup(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}

// getMac checks the ARP table for a MAC address matching the given IP.
func getMac(ip string) string {
	cmd := execCommand("arp", "-a", ip)
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
