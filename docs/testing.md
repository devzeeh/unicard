# Testing Guide

You can fully test the UniCard system without needing physical hardware by using software simulation tools like MQTT Explorer and the built-in Web Terminal Simulator.

## 1. Using MQTT Explorer

MQTT Explorer allows you to simulate the JSON payloads that the ESP32 would normally send when an RFID card is tapped.

### Setup
1. Open [MQTT Explorer](http://mqtt-explorer.com/).
2. Connect to your local broker:
   - **Host:** `localhost`
   - **Port:** `1883`
   - **Protocol:** `mqtt://`

### Simulating a "New Card Registration"
1. Open the **Add Card** page in your browser.
2. In MQTT Explorer, publish to the topic: `unicard/card/add`
3. Use the following JSON payload:
   ```json
   {
     "uid": "TEST-CARD-999",
     "terminal_id": "REGISTRATION_01"
   }
   ```
4. Click **Publish**. The browser should instantly auto-fill the UID input field.

### Simulating a "Retail/Fare Scan"
1. Open the **Terminal Simulator** page in your browser.
2. In MQTT Explorer, publish to `unicard/retail/scan` or `unicard/transport/scan`.
3. Use a JSON payload with a `uid` that *already exists* in your database:
   ```json
   {
     "uid": "YOUR_EXISTING_UID",
     "terminal_id": "SIM_TERM_01"
   }
   ```
4. Click **Publish**. The backend will validate the UID, look up the Card Number, and forward it to the browser, which will auto-fill the form and select the appropriate transaction type.

## 2. Using the Web Terminal Simulator

The Web Terminal Simulator is a built-in UI tool for admins to mock transactions.

1. Navigate to the **Terminal Simulator** in the admin dashboard (`/terminal-sim`).
2. Enter a registered **Card Number**.
3. Select a **Merchant** from the dropdown.
4. Choose the **Transaction Type** (Fare, Retail Payment, or Refund).
5. Enter the **Amount**.
6. Click **Process Scan**.

The system will:
- Deduct the balance from the card.
- Deduct the merchant's service fee commission.
- Add Loyalty Points (0.2% cashback) to the customer's account.
- Add a new record to the `transactions` table.
- Provide a success/failure message on the screen.
