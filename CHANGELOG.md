# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Prepared repository for upcoming changes and release notes.

### Changed
- —

### Deprecated
- —

### Removed
- —

### Fixed
- —

### Security
- —

## [0.0.2] - 2026-08-08

### Added
- Admin Dashboard Analytics: Added admin-facing analytics and reporting for transactions and merchant income.
- Merchant Features: Implemented merchant signup, enhanced merchant dashboard, income and transaction history endpoints, and terminal request functionality.
- Documentation: Added comprehensive setup, hardware, usage and testing documentation.
- Hardware Integration: Integrated MQTT-based hardware communication for ESP32/RC522 terminals and added card validation flows.
- Object Storage: Added support for Cloudflare R2 object storage handlers.

### Changed
- Web UI: Redesigned frontend UI and improved templates for customer and merchant dashboards and admin views.
- Payments: Continued integration and hardening of Xendit payment workflows.
- Dependencies: Routine dependency bumps (AWS SDK v2, golang.org/x/* packages) and GitHub Actions upgrades via Dependabot.

### Deprecated
- —

### Removed
- Removed legacy Stripe integration and replaced with Xendit where applicable.

### Fixed
- Concurrency: Fixed a race condition in otpStore by adding a mutex to guard read/write access.
- Multiple small bug fixes across transaction handling and notification flows.

### Security
- Improved JWT handling and RequireAuth middleware to support JWT extraction from Authorization headers and improved error logging and rate-limit middleware.

## [0.0.1-alpha] - 2026-07-31

### Added
- **Core Architecture**: Established initial Go backend and MySQL database schema.
- **Card Lifecycle Management**: Built end-to-end flows to register, activate, reload, and block UniCards.
- **Dual Payment Engine**: Developed distinct transaction logic for retail billing and distance-based public transport fares.
- **Automated Discounts & Rewards**:
  - Implemented an automatic 20% discount system for eligible user profiles (e.g., PWD, Students, Seniors).
  - Implemented a 0.2% cashback reward system applied per payment transaction.
- **Hardware Integration**: Added robust REST endpoints to process payloads from ESP32 & RC522 RFID terminals.
- **Web Dashboard MVP**: Delivered a server-rendered admin interface using Go Templates, HTML, JavaScript, and Tailwind CSS for monitoring transaction logs, user data, and card statuses.
- **Payment Processing**: Integrated Xendit for handling secure digital payments and payouts.
- **Notifications**: Configured an SMTP service to send receipts to users upon successful transactions.
- **API Authentication**: Implemented JWT-based authentication and role-based access control for API endpoints.
- **R2 Storage Integration**: Integrated Cloudflare R2 for storage of uploaded documents and images.

### Changed
- None

### Deprecated
- None

### Removed
- None

### Fixed
- None

### Security
- Basic transaction security for MVP and session management.

---

### References
- This changelog follows Keep a Changelog conventions: https://keepachangelog.com

[Unreleased]: https://github.com/devzeeh/unicard/compare/v0.0.2...HEAD
[0.0.2]: https://github.com/devzeeh/unicard/releases/tag/v0.0.2
[0.0.1-alpha]: https://github.com/devzeeh/unicard/releases/tag/v0.0.1-alpha
