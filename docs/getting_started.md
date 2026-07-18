# Getting Started

This guide walks you through setting up the UniCard backend and database on your local machine.

## Prerequisites

Before starting, ensure you have the following installed:
- [Go (v1.25+)](https://golang.org/doc/install)
- [MySQL Server](https://dev.mysql.com/downloads/)
- [Mosquitto MQTT Broker](https://mosquitto.org/download/)

## 1. Clone the Repository

Clone the UniCard repository to your local machine:
```bash
git clone https://github.com/devzeeh/unicard.git
cd unicard
```

## 2. Database Setup

You need to create a MySQL database and import the schema.
1. Open your MySQL client or terminal.
2. Create a new database, for example, `unicard_db`.
3. Import the provided schema:
   ```bash
   mysql -u root -p unicard_db < docs/unicard.sql
   ```

## 3. Environment Variables

The application relies on a `.env` file for configuration.
1. Copy the example file:
   ```bash
   cp .env.example .env
   ```
2. Open `.env` and configure your settings:
   - `DB_USER`, `DB_PASSWORD`, `DB_NAME`
   - `MQTT_BROKER` (e.g., `tcp://localhost:1883`)
   - `XENDIT_API_KEY` (if using payment gateways)

## 4. MQTT Broker Configuration

Your Mosquitto broker must be configured to support WebSockets, as the frontend communicates directly with it.
Edit your `mosquitto.conf` to include:
```text
listener 1883
listener 9001
protocol websockets
```
Restart the Mosquitto service after applying the changes.

## 5. Run the Application

Start the Go backend server. Note that because Windows Defender might block executables built in the temporary folder by `go run`, it is safer to build it first:
```bash
go build -o main.exe backend/cmd/app/main.go
.\main.exe
```
If successful, you will see `Server started on: http://localhost:3001` and `Connected to MQTT Broker!` in your terminal.

---
**Next Step:** See the [Hardware Setup Guide](hardware_setup.md) to build your ESP32 scanners.
