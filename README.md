# UniCard

> This is a practice ground for my personal project *[PayCard](https://github.com/devzeeh/PayCard)*

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/devzeeh/unicard)
[![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?style=flat&logo=mysql&logoColor=blue)](https://www.mysql.com/)
![Status](https://img.shields.io/badge/Status-Development-blue)
![GitHub Created At](https://img.shields.io/github/created-at/devzeeh/unicard)
![GitHub last commit](https://img.shields.io/github/last-commit/devzeeh/unicard)
![GitHub language count](https://img.shields.io/github/languages/count/devzeeh/unicard)
![GitHub Repo stars](https://img.shields.io/github/stars/devzeeh/unicard)

> A unified contactless payment system for retail stores and public transportation.
>*"The single card for all your retail and fare payments."*

**[Documentation](docs/)** • **[Features](#features)** • **[Quick Start](#quick-start)**  • **[Contributing](docs/CONTRIBUTING.md)**

## Overview

**PayCard** is a cashless payment solution designed for retail stores and transportation. Built with affordable hardware and accessible technology as a school project.
> **Note:** This project is for educational purposes only and not intended for commercial use.

## Why PayCard?

- **Fast Contactless Payments** - Utilizes RFID (ESP32 + RC522) for quick, tap-to-pay functionality.
- **Affordable Hardware** - Built using low-cost, readily available components.
- **Rewards & Discount Logic** - Includes a proof-of-concept for a 20% fare discounts (e.g., for PWD/Students).
- **Unified System** - A single card system designed to handle both retail (itemized) and transport (fare) transactions.
- **Analytics Dashboard** - A simple dashboard for viewing transaction history and user data.

## Features

- **Card Lifecycle Management** - Core functions to register, activate, load, and block RFID cards.
- **Dual Payment Logic** - Handles both itemized billing (for retail) and distance-based fare calculation (for transport).
- **Reward Points System** - A proof-of-concept for calculating 0.2% cashback points per transaction.
- **Email Receipts** - Automatically sends transaction details via SMTP (e.g., Gmail) after a payment.
- **Web Dashboard** - A simple web interface for users, merchants, and admins to review transaction logs.
- **Transaction Security** - Implements card-to-server authentication, balance verification, and basic audit logging.

## Tech Stack (MVP Stack)

| Component | Technology |
|-----------|-----------|
| **Backend** | Go 1.22+ |
| **Database** | MySQL 8.0+ |
| **Payments** | Stripe API |
| **Hardware** | ESP32 + RC522 RFID |
| **Frontend** | HTML, Tailwind CSS, JavaScript |
| **Email** | SMTP (Gmail) |
| **Version Control** | Git & GitHub |

## Quick Start

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
Points Earned:     +₱0.052
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
Points Earned:     +₱0.20
Total Points:      ₱12.85
════════════════════════════════════
Thank you for paying with PAYCARD!
```

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
[Back to Top](#paycard)

</div>
