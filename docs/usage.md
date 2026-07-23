# Usage Guide

UniCard features three primary user roles: Super Admin, Merchant, and Customer. Here is how each role interacts with the system.

## 1. Super Admin

The Super Admin is responsible for the overarching management of the UniCard ecosystem.

### Managing Cards
- Navigate to **Card Inventory**. Here you can view all registered cards, block lost cards, or deactivate them.
- To **Provision a New Card**, navigate to the **Add Card** page. 
  - Ensure the **Card Registration ESP32 Terminal** is connected.
  - Tap a new RFID card on the scanner. The UID will automatically populate in the browser.
  - Enter the initial balance and save.

### Managing Merchants
- Navigate to **Merchants** to review pending merchant applications.
- Approve or reject merchant documents (e.g., Business Permits, IDs).

### Managing Terminals
- Navigate to **Terminals** to add new hardware terminals to the registry and assign them to specific merchants.

## 2. Merchant

Merchants are business owners (Retail shops or Transport drivers) who accept UniCard payments.

- **Registration:** Merchants must register and upload valid business documents for approval.
- **Dashboard:** Once approved, merchants can view their daily/monthly income and transaction volume.
- **Transactions:** Merchants can view a detailed ledger of all payments received from customers.
- **Withdrawals:** Merchants can request withdrawals (payouts) to their bank accounts.

## 3. Customer

Customers are the end-users who hold the physical UniCard RFID cards.

- **Registration:** Customers sign up using an OTP sent to their email. During signup, they link their physical UniCard by entering the Card Number.
- **Dashboard:** Customers can view their current balance and accumulated loyalty points.
- **Top-Up:** Customers can reload their card balance using the integrated Xendit payment gateway (e-wallets, cards, bank transfers).
- **Transactions:** View real-time transaction history (fares deducted, retail purchases made).
- **Discounts:** Eligible customers (e.g., PWD, Students) automatically receive a 20% discount on transport fares.

---
**Next Step:** Learn how to test the system without physical hardware in the [Testing Guide](testing.md).
