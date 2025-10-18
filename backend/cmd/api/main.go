package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/redis/go-redis/v9"
)

type Device struct {
	IP       string `json:"ip"`
	MAC      string `json:"mac"`
	Hostname string `json:"hostname"`
}

var (
	ctx = context.Background()
	rdb *redis.Client
)

func main() {

	// Connect to Redis
	rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	fmt.Println("✅ Connected to Redis")

	// Fiber server
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Scan route
	app.Post("/scan", func(c *fiber.Ctx) error {
		var req struct {
			CIDR string `json:"cidr"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).SendString("Invalid JSON")
		}

		fmt.Println("🚀 Scan started for:", req.CIDR)
		go scanNetwork(req.CIDR)
		return c.JSON(fiber.Map{"status": "scan started", "cidr": req.CIDR})
	})

	// Fetch devices route
	app.Get("/devices", func(c *fiber.Ctx) error {
		keys, _ := rdb.Keys(ctx, "device:*").Result()
		var devices []Device
		for _, k := range keys {
			data, _ := rdb.HGetAll(ctx, k).Result()
			if len(data) > 0 {
				devices = append(devices, Device{
					IP:       data["ip"],
					MAC:      data["mac"],
					Hostname: data["hostname"],
				})
			}
		}
		return c.JSON(devices)
	})

	log.Fatal(app.Listen(":8080"))
}

func scanNetwork(cidr string) {
	ips, err := hosts(cidr)
	if err != nil {
		fmt.Println("❌ Invalid CIDR:", err)
		return
	}

	var wg sync.WaitGroup
	for _, ip := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			cmd := exec.Command("ping", "-n", "1", "-w", "200", ip)
			out, _ := cmd.CombinedOutput()
			if strings.Contains(strings.ToLower(string(out)), "ttl=") {
				fmt.Println("✔ Found device:", ip)
				rdb.HSet(ctx, "device:"+ip, map[string]interface{}{
					"ip":       ip,
					"mac":      "unknown",
					"hostname": resolveHostname(ip),
				})
			}
		}(ip)
	}
	wg.Wait()
	fmt.Println("✅ Scan complete.")
}

func resolveHostname(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return "unknown"
	}
	return strings.TrimSuffix(names[0], ".")
}

func hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	if len(ips) > 2 {
		return ips[1 : len(ips)-1], nil
	}
	return ips, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
