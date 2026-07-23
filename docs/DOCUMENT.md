# UniCard Documentation Hub

Welcome to the official documentation for **UniCard**, a unified contactless payment system for retail stores and public transportation. 

This document serves as the master index. Click on the links below to navigate to the specific, detailed guides for setting up, using, and testing the system.

## Table of Contents

### 1. [Getting Started](getting_started.md)
The foundational guide to setting up your environment.
- Software prerequisites (Go, MySQL, Mosquitto).
- Cloning the repository and importing the database.
- Running the Go backend server.

### 2. [Hardware Setup Guide](hardware_setup.md)
Everything you need to build the physical RFID scanners.
- Required components (ESP32, RC522).
- Complete wiring diagram and pin-outs.
- How to flash the firmware for Retail, Transport, and Registration terminals.

### 3. [Usage Guide](usage.md)
A breakdown of the different user roles and how they interact with the web platform.
- **Super Admin:** Managing cards, terminals, and merchants.
- **Merchant:** Tracking income, managing transactions, and requesting withdrawals.
- **Customer:** Topping up via Xendit, viewing transaction history, and earning rewards.

### 4. [Testing Guide](testing.md)
How to validate the system's logic without physical hardware.
- Using **MQTT Explorer** to mock hardware scans and test WebSocket connectivity.
- Using the built-in **Web Terminal Simulator** to process simulated retail and fare transactions.

---

### Key Technical Concepts

- **Dual Payment Logic:** The system determines the transaction fee/type depending on which MQTT topic the ESP32 publishes to (`unicard/retail/scan` vs `unicard/transport/scan`).
- **Real-Time UI:** The Go backend validates the hardware scans and publishes the verified data back to the frontend via WebSockets (`unicard/frontend/scan`), creating an instant, seamless experience.
- **Security:** The backend enforces strict database constraints, OTP verification, and JWT session management. 
