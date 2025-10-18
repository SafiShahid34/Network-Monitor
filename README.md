🧭 Network Monitor Dashboard

A simple full-stack app that scans your local network for active devices and shows the results in a web dashboard.
The backend is built in Go, the frontend in React, and Redis is used to cache scan results for faster responses.

🚀 Features

Scans local networks using ping and arp

Displays IP, MAC, and hostname for each active device

Caches scan results for 30 seconds using Redis

Clean and responsive web dashboard built with React

⚙️ How It Works

The React frontend calls the Go backend to start a scan.

The backend runs a fast, concurrent ping sweep using goroutines.

Results (IP, MAC, hostname) are cached in Redis for quick reloads.

The frontend shows all connected devices in a table.

🛠️ Setup
1. Clone the repo
git clone https://github.com/SafiShahid34/network-monitor.git
cd network-monitor

2. Start Redis
docker run -d --name redis -p 6379:6379 redis:latest

3. Run the backend
cd backend
go mod tidy
go run main.go

4. Run the frontend
cd frontend
npm install
npm start

🧪 Tests

The backend includes Go unit tests for:

IP range generation

Incrementing IPs

Reverse DNS

Simulated scan behavior

Run tests:

cd backend
go test ./... -v
