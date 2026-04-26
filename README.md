# UniCard

![GitHub Release](https://img.shields.io/github/v/release/devzeeh/unicard)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/devzeeh/unicard)
![Status](https://img.shields.io/badge/Status-Development-blue)
![GitHub Created At](https://img.shields.io/github/created-at/devzeeh/unicard)
![GitHub last commit](https://img.shields.io/github/last-commit/devzeeh/unicard)
![GitHub language count](https://img.shields.io/github/languages/count/devzeeh/unicard)
![GitHub Repo stars](https://img.shields.io/github/stars/devzeeh/unicard)

> A unified contactless payment system for retail stores and public transportation.
>*"The single card for all your retail and fare payments."*

**[Documentation](docs/)** • **[Features](#features)** • **[Quick Start](#quick-start)**  • **[Contributing](docs/CONTRIBUTING.md)**

## Overview

**Unicard** is a cashless payment solution designed for retail stores and transportation. Built with affordable hardware and accessible technology as a school project.
> **Note:** This project is for educational purposes only and not intended for commercial use.

## Why Unicard?

- **Fast Contactless Payments** - Utilizes RFID (ESP32 + RC522) for quick, tap-to-pay functionality.
- **Rewards & Discount Logic** - Includes a proof-of-concept for a 20% fare discounts (e.g., for PWD/Students).
- **Unified System** - A single card system designed to handle both retail (itemized) and transport (fare) transactions.
- **Analytics Dashboard** - A simple dashboard for viewing transaction history and user data.

## Features

- **Card Lifecycle Management** - Core functions to register, activate, load, and block UniCard.
- **Dual Payment Logic** - Handles both itemized billing (for retail) and distance-based fare calculation (for transport). (QR Code & RFID)
- **Reward Points System** - Calculates 0.2% cashback points per transaction.
- **Email Receipts** - Automatically sends transaction details to users registered email after a successful transaction.
- **Web Dashboard** - A simple web interface for users, merchants, and admins to review transaction logs, user data, and card details.
- **Transaction Security** - Ensures secure data transmission, robust session management, and OTP verification to protect user accounts and payment information.

## Tech Stack (MVP Stack)

| Component | Technology |
|-----------|-----------|
| **Backend** | Go 1.25 |
| **Database** | MySQL |
| **Payments** | Stripe |
| **Hardware** | ESP32 + RC522 RFID |
| **Frontend** | HTML, Tailwind CSS, JavaScript |
| **Email** | SMTP |
| **Version Control** | Git & GitHub |

## Quick Start

## Demo
A video demonstration will be available here once the project is finalized.

### Transaction Flow
```
Tap Card -> Validate Balance -> Deduct Payment -> Update Data -> Earn Rewards -> Send Receipt -> Update Balance.
```

### Sample Fare Receipt

![Sample Fare Receipt](frontend/assets/images/Fare%20Receipt.png)

### Sample Retail Receipt

![Sample Retail Receipt](frontend/assets/images/Retail%20Receipt.png)

## Acknowledgements

This project was made possible by the incredible work of the following communities and technologies:

- **The Go Community** - For a robust and efficient backend language.
- **MySQL** - For a reliable and powerful database solution.
- **ESP & Arduino Developers** - For versatile microcontroller support.
- **Tailwind Labs** - For a utility-first CSS framework.
- **Stripe** - For accessible and developer-friendly payment infrastructure.
- **The entire Open Source Community** - For the countless tools and libraries that power modern development.

---

<div align="center">

### Made with ❤️ in the Philippines

<img src="https://flagcdn.com/w40/ph.png" width="30" height="20" alt="Philippine Flag">
<br />

<!--*Built by Filipinos, for Filipinos*-->

**Helping local businesses and campuses build cashless ecosystems**

<small>Copyright © 2025 devzeeh. All Rights Reserved.</small>
<br />
[Back to Top](#UniCard)

</div>
