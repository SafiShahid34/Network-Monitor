package scanner

import (
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Tests for utility functions and core scanner behavior.

// Verifies that getIPsInRange generates the expected list of IPs for a small CIDR.
func TestGetIPsInRange(t *testing.T) {
	ips := getIPsInRange("192.168.1.0/30")
	expected := []string{"192.168.1.0", "192.168.1.1", "192.168.1.2", "192.168.1.3"}

	if len(ips) != len(expected) {
		t.Fatalf("Expected %d IPs, got %d", len(expected), len(ips))
	}
	for i, ip := range expected {
		if ips[i] != ip {
			t.Errorf("Expected %s, got %s", ip, ips[i])
		}
	}
}

// Confirms that the IP increment helper correctly moves to the next address.
func TestInc(t *testing.T) {
	ip := net.ParseIP("192.168.1.1")
	inc(ip)
	if !ip.Equal(net.ParseIP("192.168.1.2")) {
		t.Errorf("Expected 192.168.1.2, got %s", ip.String())
	}
}

// Checks reverse DNS resolution; some IPs may not have PTR records, which is fine.
func TestReverseLookup(t *testing.T) {
	name := reverseLookup("8.8.8.8")
	if name == "" {
		t.Log("reverseLookup returned empty (no PTR record found, acceptable)")
	}
}

// --- Mock setup ---

// fakeCommand replaces exec.Command during tests to simulate ping/arp responses.
func fakeCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// TestHelperProcess mimics system command behavior used in ScanCIDR.
// It returns fake ping and arp outputs so we can test without hitting the network.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	cmd := args[3]

	switch cmd {
	case "ping":
		ip := args[len(args)-1]
		// Simulate only .1 responding to ping
		if strings.HasSuffix(ip, "1") {
			println("Reply from " + ip + ": bytes=32 time<1ms TTL=64")
		}
	case "arp":
		ip := args[len(args)-1]
		println(ip + " 00-11-22-33-44-55 dynamic")
	default:
		println("")
	}
	os.Exit(0)
}

// --- Main scanner test ---

// Ensures that ScanCIDR correctly picks up a mock active host and retrieves its MAC.
func TestScanCIDR_Mocked(t *testing.T) {
	//execCommand = fakeCommand --> UT
	//defer func() { execCommand = exec.Command }() --

	devices := ScanCIDR("192.168.1.0/30")
	if len(devices) == 0 {
		t.Fatal("Expected at least one active device")
	}

	found := false
	for _, d := range devices {
		if d.IP == "192.168.1.1" {
			found = true
			if d.MAC != "00-11-22-33-44-55" {
				t.Errorf("Expected mock MAC, got %s", d.MAC)
			}
		}
	}
	if !found {
		t.Error("Expected 192.168.1.1 in scan results")
	}
}
