# UniCard

> This is a practice ground for my personal project *[PayCard](https://github.com/devzeeh/PayCard)*

[![Go version](https://img.shields.io/badge/Go-v1.25.3-blue)](https://go.dev/)
[![MySQL](https://img.shields.io/badge/MySQL-8.0+-4479A1?style=flat&logo=mysql&logoColor=blue)](https://www.mysql.com/)
![Status](https://img.shields.io/badge/Status-Development-yellow)
[![Created at](https://img.shields.io/badge/Created_at-november_2025-orange.svg?style=flat)](https://github.com/Archival/UniCard)
![Gitea Last Commit](https://img.shields.io/gitea/last-commit/Archival/UniCard)
![Gitea language count](https://img.shields.io/gitea/languages/count/Archival/UniCard)
![Gitea Stars](https://img.shields.io/gitea/stars/Archival/UniCard)

> A unified contactless payment system for retail stores and public transportation in the Philippines.
>*"The single card for all your retail and fare payments."*

**[Documentation](docs/)** • **[Features](#features)** • **[Quick Start](#quick-start)**  • **[Contributing](docs/CONTRIBUTING.md)**

## Overview

**PayCard** is a cashless payment solution designed for retail stores and transportation. Built with affordable hardware and accessible technology as a school project.
> **Note:** This project is for educational purposes only and not intended for commercial use.

## Why PayCard?

- **Fast Contactless Payments** - Utilizes RFID (ESP32 + RC522) for quick, tap-to-pay functionality.
- **Affordable Hardware** - Built using low-cost, readily available components.
- **Rewards & Discount Logic** - Includes a proof-of-concept for a 0.20% cashback and fare discounts (e.g., for PWD/Students).
- **Unified System** - A single card system designed to handle both retail (itemized) and transport (fare) transactions.
- **Analytics Dashboard** - A simple dashboard for viewing transaction history and user data.

## Features

- **Card Lifecycle Management** - Core functions to register, activate, load, and block RFID cards.
- **Dual Payment Logic** - Handles both itemized billing (for retail) and distance-based fare calculation (for transport).
- **Reward Points System** - A proof-of-concept for calculating 0.20% cashback points per transaction.
- **Email Receipts** - Automatically sends transaction details via SMTP (e.g., Gmail) after a payment.
- **Web Dashboard** - A simple web interface for users, merchants, and admins to review transaction logs.
- **Transaction Security** - Implements card-to-server authentication, balance verification, and basic audit logging.

**[View Full Feature List ](docs/FEATURES.md)**


## Tech Stack

| Component | Technology |
|-----------|-----------|
| **Backend** | Go 1.22+ |
| **Database** | MySQL 8.0+ |
| **Payments** | Stripe API |
| **Hardware** | ESP32 + RC522 RFID |
| **Frontend** | HTML, Tailwind CSS, JavaScript |
| **Email** | SMTP (Gmail, etc.) |
| **Version Control** | Git & GitHub |

## Quick Start

### Prerequisites

- Go 1.22+ • MySQL 8.0+ • Arduino IDE
- Stripe Account • SMTP Email

**[Detailed Installation Guide](docs/INSTALLATION.md)**

## Hardware Setup

**Required Components:**
- ESP32 NodeMCU
- RC522 RFID Module
- MIFARE Classic Cards
- Jumper wires
- Breadboard (optional)
- USB cable for ESP32

**Quick Wiring:**
```
RC522  ESP32
SDA    GPIO15
SCK    GPIO14
MOSI   GPIO13
MISO   GPIO12
RST    GPIO0
3.3V   3.3V
GND    GND
```

**Important:** Ensure all connections are secure to avoid communication issues between the ESP32 and RC522 module. **The RC522 operates at 3.3V only!**

**[Complete Hardware Guide ](docs/HARDWARE.md)**


## Documentation

All project documentation is available in the `/docs` folder.

### 1. Getting Started
* **[Installation Guide](docs/INSTALLATION.md)**: Step-by-step setup instructions
* **[Hardware Setup](docs/HARDWARE.md)**: Wiring diagrams and assembly instructions
* **[Usage Guide](docs/USAGE.md)**: How to use for commuters, merchants, and admins

### 2. Project Reference
* **[API Documentation](docs/API.md)**: Complete API reference with examples
* **[Feature List](docs/FEATURES.md)**: Detailed list of features and functionalities

### 3. Community & Development
* **[Contributing Guide](docs/CONTRIBUTING.md)**: How to contribute to the project
* **[Changelog](docs/CHANGELOG.md)**: Version history and updates


## Demo
A video demonstration will be available here once the project is finalized.

### Transaction Flow
```
User taps card ➔ Validates balance ➔ Deducts payment ➔ Earns points ➔ Receipt sent
```

### Sample Fare Receipt

```
════════════════════════════════════
        PAYCARD RECEIPT
════════════════════════════════════
Card ID:        A1B2C3D4
Date:           Nov 1, 2025 3:35PM
Transaction:    Transport - Route 42
────────────────────────────────────
Previous Balance:  ₱150.00
Fare Deducted:     ₱13.00
Discount (PWD):    -₱2.60
────────────────────────────────────
New Balance:       ₱139.60
Points Earned:     +₱0.065
Total Points:      ₱12.35
════════════════════════════════════
Thank you for using PAYCARD!
════════════════════════════════════
```

### Sample Retail Receipt

```
════════════════════════════════════
        PAYCARD RECEIPT
════════════════════════════════════
Card ID:        A1B2C3D4
Date:           Nov 1, 2025 4:15PM
Transaction:    Retail - Store
────────────────────────────────────
Item 1: Rice (₱50.00)
Item 2: Canned Goods (₱30.00)
Item 3: Snacks (₱20.00)
────────────────────────────────────
Subtotal:         ₱100.00
────────────────────────────────────
New Balance:       ₱39.60
Points Earned:     +₱0.50
Total Points:      ₱12.85
════════════════════════════════════
Thank you for paying with PAYCARD!
```

## License

Copyright © 2025 devzeeh. All Rights Reserved.

This project is made available for personal, educational, and academic use only. Commercial use is strictly prohibited.

**You are free to:**
* **Learn:** Use, study, and modify the code for personal learning, academic research, or school projects.
* **Experiment:** Build personal demos and test new features.

**You are NOT allowed to:**
* **Use Commercially:** Use the software in any product, service, or for any profit-generating purpose.
* **Redistribute:** Resell or redistribute the software.

## Support
- **[Discussions](https://github.com/Archival/UniCard/discussions)**
- **[Issues](https://github.com/Archival/UniCard/issues)**
- If this project is helpful, please leave a star! ⭐

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

*Built by Filipinos, for Filipinos*

**Helping local businesses and campuses build cashless ecosystems**

<small>Copyright © 2025 devzeeh. All Rights Reserved.</small>
<br />
[Back to Top](#paycard)

</div>