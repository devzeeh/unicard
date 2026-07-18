#include <WiFi.h>
#include <PubSubClient.h>
#include <SPI.h>
#include <MFRC522.h>
#include <ArduinoJson.h>

// --- WiFi Settings ---
const char* ssid = "YOUR_WIFI_SSID";
const char* password = "YOUR_WIFI_PASSWORD";

// --- MQTT Settings ---
const char* mqtt_server = "192.168.x.xxx"; // Replace with your Mosquitto broker IP
const int mqtt_port = 1883;
const char* mqtt_topic = "unicard/transport/scan"; // MQTT topic for transport
const char* terminal_id = "TRANSPORT_01"; // Unique identifier for this terminal

WiFiClient espClient;
PubSubClient client(espClient);

// --- RFID Settings ---
#define SS_PIN 5
#define RST_PIN 22
MFRC522 rfid(SS_PIN, RST_PIN);

unsigned long lastReconnectAttempt = 0;

void setup_wifi() {
  delay(10);
  Serial.println();
  Serial.print("Connecting to ");
  Serial.println(ssid);

  WiFi.begin(ssid, password);

  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }

  Serial.println("");
  Serial.println("WiFi connected");
  Serial.println("IP address: ");
  Serial.println(WiFi.localIP());
}

boolean reconnect() {
  if (client.connect(terminal_id)) {
    Serial.println("Connected to MQTT Broker");
  }
  return client.connected();
}

void setup() {
  Serial.begin(115200);
  
  // Initialize SPI bus and MFRC522 reader
  SPI.begin();
  rfid.PCD_Init();
  Serial.println("RFID Transport Terminal Ready.");

  setup_wifi();
  
  client.setServer(mqtt_server, mqtt_port);
  
  lastReconnectAttempt = 0;
}

void loop() {
  if (!client.connected()) {
    long now = millis();
    if (now - lastReconnectAttempt > 5000) {
      lastReconnectAttempt = now;
      if (reconnect()) {
        lastReconnectAttempt = 0;
      }
    }
  } else {
    client.loop();
  }

  // Look for new cards
  if (!rfid.PICC_IsNewCardPresent() || !rfid.PICC_ReadCardSerial()) {
    return;
  }

  // A card was read
  String uidString = "";
  for (byte i = 0; i < rfid.uid.size; i++) {
    if (rfid.uid.uidByte[i] < 0x10) uidString += "0";
    uidString += String(rfid.uid.uidByte[i], HEX);
  }
  uidString.toUpperCase();

  Serial.print("Card Scanned! UID: ");
  Serial.println(uidString);

  // Stop reading
  rfid.PICC_HaltA();
  rfid.PCD_StopCrypto1();

  // Create JSON Payload
  StaticJsonDocument<200> doc;
  doc["uid"] = uidString;
  doc["terminal_id"] = terminal_id;

  char jsonBuffer[256];
  serializeJson(doc, jsonBuffer);

  // Publish via MQTT
  if (client.connected()) {
    Serial.print("Publishing message: ");
    Serial.println(jsonBuffer);
    client.publish(mqtt_topic, jsonBuffer);
    
    // Add a delay to prevent duplicate scans
    delay(2000); 
  } else {
    Serial.println("MQTT not connected, cannot publish.");
  }
}
