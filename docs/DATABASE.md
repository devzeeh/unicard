PayCard Database Schema Documentation

Database Name: `paycard_playground`
Version: 0.0.1
Author: devzeeh

**[Overview](#overview)** | **[Core concept](#core-concepts)** | **[Table Definitions](#table-definitions)** | **[ER Summary](#entity-relationship-er-summary)** 

## Overview

The PayCard database is designed to support a closed-loop contactless payment system for retail and transport. It follows a normalized structure with a focus on data integrity, audit trails for financial transactions, and a clear separation between user accounts and physical card inventory.

## Core Concepts

Centralized Wallet: User balances are stored in the users table, not on the physical cards.

Card Inventory: Physical cards (cards table) exist independently of users until they are linked. This allows for pre-loading inventory and card replacement.

Immutable Ledger: The transactions table records balance_before and balance_after to ensure a permanent, auditable history of all fund movements.

## Table Definitions

#### table of contents:

- [users](#users-user-accounts)
- [cards](#cards-card-inventory)
- [merchants](#merchants-business-partners)
- [transactions](#transactions-financial-ledger)
- [card_reports](#card_reports-incident-log)
- [receipts](#receipts-digital-receipts)
- [loyalty_redemptions](#loyalty_redemptions-points-history)
- [fare_settings](#fare_settings-dynamic-pricing)
- [admin_users](#admin_users-back-office)

## users (User Accounts)

- Purpose: Stores user identity, authentication, and current financial state.

- Key Columns:

    - id: Internal Primary Key (INT). Used for all foreign key relationships.

    - user_id: Public-facing ID (VARCHAR). Safe to display on receipts/app.

    - password_hash: Bcrypt hash of the user's password.

    - balance: The user's current wallet balance.

    - loyalty_points: Current points balance.

    - Notes: This table does not store card details directly. It links to cards via the cards table.
    
    [back to index](#table-definitions)

## cards (Card Inventory)

- Purpose: The master inventory of all physical cards issued by the system.

- Key Columns:

    - card_uid: The internal RFID/NFC tag ID (read by the scanner).

    - card_number: The 16-digit number printed on the card face.

    - initial_amount: Pre-loaded amount (set by admin) given to the user upon registration.

    - user_id: Links the card to a specific user. If NULL, the card is unassigned (inventory).

    - status: active, inactive (unsold), blocked, lost, expired.

    - Workflow:
    Admin creates card -> status: 'inactive', user_id: NULL. User signs up with card number -> Backend checks initial_amount, adds to user balance, sets 
    - status: 'active', user_id: `USER_ID`.
    
    [back to index](#table-definitions)

## merchants (Business Partners)

- Purpose: Registry of all entities (drivers, stores) that can accept payments.

- Key Columns:

    - merchant_code: Unique public ID (e.g., "BUS-001").

    - business_type: Transport or Retail (determines fare logic).

    - settlement_account: Bank/E-wallet details for payouts.

    - verified: Boolean flag. Must be TRUE for merchant to operate.

    [back to index](#table-definitions)

## transactions (Financial Ledger)

- Purpose: The single source of truth for all money movement.

- Key Columns:

    - transaction_type: Payment (deduction), Topup (addition), Refund, etc.

    - amount: The transaction value.

    - balance_before / balance_after: Critical for auditing. balance_after must always equal balance_before +/- amount.

    - discount_amount / discount_type: Records any discounts applied (PWD/Student).

    - payment_method: For top-ups (Stripe, Cash, etc.).

    - Relationships: Links to users (payer) and merchants (payee).

    [back to index](#table-definitions)

## card_reports (Incident Log)

- Purpose: Logs reported lost/stolen cards.

- Key Columns:

    - reason: Why the card was reported.

    - status: Reported, Resolved, Replaced.

    - replacement_card_number: Links to the new card issued to replace the lost one.

    [back to index](#table-definitions)

## receipts (Digital Receipts)

- Purpose: Stores metadata for generated e-receipts.

- Key Columns:

    - pdf_data: (Optional) Can store the binary PDF or a path to the file.

    - email_sent: Tracks if the user received their copy.

    [back to index](#table-definitions)

## loyalty_redemptions (Points History)

- Purpose: Logs when users spend their loyalty points.

- Key Columns:

    - points_used: Amount deducted.

    - value_php: The monetary value equivalent of those points.

    - transaction_id: Links to a specific purchase if points were used as payment.

    [back to index](#table-definitions)

## fare_settings (Dynamic Pricing)

- Purpose: Configuration table for transport fares and discounts.

- Key Columns:

    - : Links specific settings to a merchant (or null for global defaults).

    - base_fare, per_km_rate: Used by the backend to calculate transport costs.

    - pwd_discount_rate: Percentage discount for special user types.

    [back to index](#table-definitions)

## admin_users (Back Office)

- Purpose: Accounts for system administrators and   support staff.

- Key Columns:

    - role: `super admin`, `admin`, `support`, `viewer`.

    - permissions: JSON blob for fine-grained access control (e.g., {"can_refund": true}).


    [back to index](#table-definitions)

## Entity Relationship (ER) Summary

Users 1 -- **has many** -- Cards (via cards.user_id)

Users 1 -- **makes many** -- Transactions

Merchants 1 -- **receives many** -- Transactions

Transactions 1 -- **has one** -- Receipts

Transactions 1 -- **has one (optional)** -- Loyalty Redemptions

[back to index](#)