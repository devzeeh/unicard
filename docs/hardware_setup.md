# Hardware Setup Guide

UniCard utilizes ESP32 microcontrollers paired with RC522 RFID modules to create affordable contactless scanning terminals.

## Components Needed
- ESP32 Development Board
- MFRC522 RFID Module
- Jumper wires

## Wiring Diagram

Connect the RC522 to the ESP32 using the SPI interface:

| RC522 Pin | ESP32 Pin | Note |
|-----------|-----------|------|
| SDA / SS  | GPIO 5    | Chip Select |
| SCK       | GPIO 18   | SPI Clock |
| MOSI      | GPIO 23   | Master Out Slave In |
| MISO      | GPIO 19   | Master In Slave Out |
| IRQ       | Unconnected | Not used |
| GND       | GND       | Ground |
| RST       | GPIO 22   | Reset |
| 3.3V      | 3.3V      | **CRITICAL:** Do NOT connect to 5V! |

## Firmware Setup

There are three distinct terminal modes. The source code for each is located in the `hardware/` folder.

1. **Retail Terminal:** `hardware/esp32_retail/esp32_retail.ino`
2. **Transport Terminal:** `hardware/esp32_transport/esp32_transport.ino`
3. **Card Registration:** `hardware/esp32_add_card/esp32_add_card.ino`

### Configuration

Before flashing any of these sketches via the Arduino IDE, you must configure the WiFi and MQTT settings at the top of the file:

```cpp
// --- WiFi Settings ---
const char* ssid = "YOUR_WIFI_SSID";
const char* password = "YOUR_WIFI_PASSWORD";

// --- MQTT Settings ---
const char* mqtt_server = "192.168.x.xxx"; // Replace with your Mosquitto broker IP
```
*Note: Ensure your ESP32 is connected to the same network as your Mosquitto broker.*

### Flashing the ESP32
1. Open the `.ino` file in the [Arduino IDE](https://www.arduino.cc/en/software).
2. Install required libraries via the Library Manager (`Ctrl+Shift+I`):
   - `PubSubClient` by Nick O'Leary
   - `MFRC522` by GithubCommunity
   - `ArduinoJson` by Benoit Blanchon
3. Select your ESP32 board and COM port.
4. Click **Upload**.

Once uploaded, open the Serial Monitor (115200 baud). The ESP32 will connect to WiFi, connect to MQTT, and output "Ready." When you tap an RFID card, it will instantly publish the UID to the configured MQTT topic.

---
**Next Step:** Learn how to use the web application in the [Usage Guide](usage.md).
