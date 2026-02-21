# Unicard Project Roadmap

This document serves as a high-level overview of the features, improvements, and future plans for the Unicard system.

## Phase 1: Foundation & MVP (Minimum Viable Product)

**Goal**: Get the basic system running with essential user management and balance viewing.

- ### [x] Project Initialization
    - [x] Set up Repository & Version Control
    - [x] Configure Database (MySQL/Firebase/MongoDB)
    - [x] Setup basic folder structure
- ### [x] Authentication System
    - [x] User Registration (Sign Up)
    - [x] User Login (Sign In)
    - [x] Forgot Password flow
- ### [ ] Core Wallet Features
    - [ ] Create User Wallet upon registration
    - [ ] View Current Balance
    - [ ] View Digital Card ID/Number

## Phase 2: Transactions & Payments
**Goal**: Allow money to move between accounts or be spent.

- ### [ ] Transaction Logic
    - [ ] Deposit funds (Admin side or Simulation)
    - [ ] Transfer funds between users
    - [ ] Payment simulation (Deduct balance)
- ### [ ] Transaction History
    - [ ] List of recent transactions
    - [ ] Filter by date or type (Credit/Debit)
- ### [ ] QR Code Integration
    - [ ] Generate unique QR code for User ID
    - [ ] Scanner feature to read QR codes

## Phase 3: Admin & Security

**Goal**: Manage the ecosystem and ensure data safety.

- ### [ ] Admin Dashboard
    - [ ] View all registered users
    - [ ] Manually freeze/suspend accounts
    - [ ] System-wide analytics (Total volume, user count)
- ### [ ] Security Enhancements
    - [ ] Input validation & Sanitization
    - [ ] Role-Based Access Control (User vs Admin)
    - [ ] Session management (Auto-logout)

## Phase 4: Future Enhancements (Backlog)

**Goal**: Polish the UI and add advanced features.

- [ ] **Notification System** (Email or In-App alerts for payments)
- [ ] **Dark Mode** support
- [ ] **Profile Settings** (Change avatar, update email)
- [ ] **NFC Support** (Tap-to-pay via mobile)
- [ ] **Export Data** (Download transaction history as PDF/CSV)

---
## Legend:
`[x]` : Completed

`[ ]` : Pending / To Do

---