CREATE DATABASE IF NOT EXISTS unicardv3;
USE unicardv3;

-- =========================================================================
-- CORE IDENTITY & AUTHENTICATION TABLES
-- =========================================================================

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY COMMENT 'Internal row index optimized for database indexing and fast joins',
    user_id VARCHAR(50) NOT NULL UNIQUE COMMENT 'Custom public ID (e.g., UNI-YYMM-minsecxxxx) used in APIs and frontend',
    username VARCHAR(50) NOT NULL UNIQUE COMMENT 'Unique handle for admin/staff to log in quickly without an email',
    name VARCHAR(100) NOT NULL COMMENT 'Full name of the individual user or client contact person',
    email VARCHAR(100) UNIQUE NOT NULL COMMENT 'Primary email address used for consumer logins and notifications',
    phone_number VARCHAR(20) NULL UNIQUE COMMENT 'Mobile number (e.g., +639...) for OTPs and SMS transaction alerts',
    password_hash VARCHAR(255) NOT NULL COMMENT 'Cryptographically secured password string handled via bcrypt in Go',
    role ENUM('super_admin', 'merchant_admin', 'merchant_staff', 'customer') NOT NULL COMMENT 'Defines application-wide role-based access control',
    status ENUM('active', 'suspended', 'inactive') DEFAULT 'active' COMMENT 'Account access state for platform security and compliance checks',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Auto-generated timestamp of account creation'
) COMMENT='Core identity table tracking authentication and tenancy access control levels';


CREATE TABLE system_settings (
    setting_key VARCHAR(50) PRIMARY KEY COMMENT 'Unique string configuration key acting as the primary look-up token',
    setting_value VARCHAR(255) NOT NULL COMMENT 'The active parameter threshold or value parsed directly by the Go backend',
    description TEXT NULL COMMENT 'Descriptive documentation notes detailing exactly what system rules or parameters this alters',
    updated_by VARCHAR(50) NOT NULL COMMENT 'The public users.user_id of the Super Admin who executed the latest configuration adjustment override',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Auto-generated clock timestamp tracking when this specific configuration parameter was initialized',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Automatically locks the exact clock timestamp whenever this system parameter value is updated',
    FOREIGN KEY (updated_by) REFERENCES users(user_id)
) COMMENT='Global platform configuration matrix driving dynamic fees, operational bounds, and system constants';


-- =========================================================================
-- MERCHANT TENANCY & HARDWARE REGISTRY TABLES
-- =========================================================================

CREATE TABLE merchants (
    id INT AUTO_INCREMENT PRIMARY KEY COMMENT 'Internal merchant row index used for fast database indexing',
    merchant_id VARCHAR(50) NOT NULL UNIQUE COMMENT 'Custom public identifier for the business entity (e.g., MCH-2026-001)',
    business_name VARCHAR(150) NOT NULL COMMENT 'Registered trade or company name of the client merchant',
    business_type ENUM('retail', 'transportation', 'food_and_beverage', 'services', 'other') NOT NULL COMMENT 'Industry category for transaction filtering and analytics',
    business_registration_number VARCHAR(100) NULL UNIQUE COMMENT 'Official government tracking number (e.g., DTI, SEC, or BIR TIN)',
    business_address TEXT NOT NULL COMMENT 'Physical location of the main store or corporate headquarters',
    user_id VARCHAR(50) NOT NULL COMMENT 'Links to the user_id in the users table who owns this business account',
    owner_name VARCHAR(100) NOT NULL COMMENT 'Full name of the principal owner or authorized business representative',
    business_email VARCHAR(100) NOT NULL UNIQUE COMMENT 'Official company contact email address for corporate updates and billing statements',
    business_phone VARCHAR(20) NOT NULL UNIQUE COMMENT 'Official telephone or mobile number for merchant support and emergency updates',
    commission_rate DECIMAL(5, 2) DEFAULT 2.00 COMMENT 'Percentage cut taken by UniCard per processed card transaction (e.g., 2.50 = 2.5%)',
    settlement_account_name VARCHAR(100) NULL COMMENT 'The name on the merchant bank account or mobile wallet for payouts',
    settlement_account_number VARCHAR(50) NULL COMMENT 'The actual bank account number or mobile number (GCash/Maya) for payouts',
    settlement_bank_name VARCHAR(100) NULL COMMENT 'The target bank or e-wallet company name (e.g., BDO, BPI, GCash, Maya)',
    status ENUM('pending approval', 'approved', 'rejected', 'active', 'suspended') DEFAULT 'pending approval' COMMENT 'Operational state of the merchant ecosystem tenancy',
    dti_document VARCHAR(255) NULL COMMENT 'File path for the uploaded DTI registration document',
    bir_document VARCHAR(255) NULL COMMENT 'File path for the uploaded BIR registration document',
    other_document VARCHAR(255) NULL COMMENT 'File path for any other uploaded business documents',
    approved_by VARCHAR(50) NULL COMMENT 'The user_id of the Super Admin who verified and activated this company profile',
    approved_at TIMESTAMP NULL COMMENT 'The specific date and timestamp when the business was activated',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Auto-generated date and time record of the initial registration request',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Automatically updates whenever any merchant profile field is modified',
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE RESTRICT,
    FOREIGN KEY (approved_by) REFERENCES users(user_id) ON DELETE SET NULL
) COMMENT='Enterprise business registry tracking partner tenants, hardware mapping nodes, and financial settlement details';


CREATE TABLE terminals (
    id INT AUTO_INCREMENT PRIMARY KEY COMMENT 'Internal hardware registry auto-increment row index',
    terminal_id VARCHAR(50) NOT NULL UNIQUE COMMENT 'Custom public hardware identifier (e.g., TRM-2026-0001) used in API payloads',
    terminal_sn VARCHAR(50) UNIQUE NOT NULL COMMENT 'Physical factory-assigned unique serial number or MAC address of the ESP32 board',
    merchant_id VARCHAR(50) NULL COMMENT 'Links to the merchant_id of the managing merchant entity',
    device_name VARCHAR(100) NOT NULL COMMENT 'Human-readable descriptor identifying placement (e.g., Counter 1, Jeepney Plate # ABC-123)',
    location_details VARCHAR(255) NULL COMMENT 'Optional physical sector data, such as a branch route path or stall number designation',
    status ENUM('active', 'suspended', 'inactive') DEFAULT 'inactive' COMMENT 'Operational network connectivity state of the edge node hardware',
    last_heartbeat TIMESTAMP NULL COMMENT 'Tracks the precise timestamp of the last successful ping packet received from the ESP32 network stack',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Auto-generated clock timestamp tracking initial edge device registration',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Automatically monitors configuration adjustments or state transitions over time',
    -- Fixed: was merchants(user_id), now correctly points to merchants(merchant_id)
    FOREIGN KEY (merchant_id) REFERENCES merchants(merchant_id) ON DELETE CASCADE
) COMMENT='Hardware node registry tracking deployed physical authentication nodes and network heartbeat states';


-- =========================================================================
-- UTILITY & USER TRANSACTION LOGS TABLES (HIGH GROWING DATASETS)
-- =========================================================================

CREATE TABLE cards (
    card_number VARCHAR(20) PRIMARY KEY COMMENT 'The visible consumer-facing identifier printed on the physical plastic token',
    card_uid VARCHAR(50) NOT NULL UNIQUE COMMENT 'The physical hardware chip unique UID read directly from the MIFARE/RFID sectors',
    user_id VARCHAR(50) NULL COMMENT 'Links cardholder account identity via the public users.user_id string identifier',
    card_type ENUM('regular', 'student', 'pwd', 'senior') DEFAULT 'regular' COMMENT 'Drives dynamic discount calculation algorithms across transportation fares',
    discount_verified BOOLEAN DEFAULT FALSE COMMENT 'Flags whether regulatory documents were verified for fare discount tier eligibility',
    balance DECIMAL(10, 2) DEFAULT 0.00 COMMENT 'Current secure stored monetary value assigned to the physical card unit',
    loyalty_points DECIMAL(10, 2) DEFAULT 0.00 COMMENT 'Accrued transaction reward points redeemable at verified retail merchant stations',
    status ENUM('active', 'inactive', 'blocked', 'lost') DEFAULT 'inactive' COMMENT 'Lifecycle block status constraint to instantly freeze stolen or missing tokens',
    expiry_date DATE NOT NULL COMMENT 'Expiration threshold date determining card block lifecycle validations',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp tracking when the physical token record was registered in inventory',
    linked_at TIMESTAMP NULL DEFAULT NULL COMMENT 'Timestamp tracking the exact moment the card was registered to a specific user',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Automatically updates when balances adjust or card status switches occur',
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE SET NULL
) COMMENT='Ecosystem transit wallet asset tracker maintaining balances, hardware mapping tokens, and fare tier flags';


CREATE TABLE transactions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT 'Internal financial primary index scaled to 64-bit headroom to comfortably support billions of platform entries',
    transaction_id VARCHAR(50) NOT NULL UNIQUE COMMENT 'Custom unique public reference string (e.g., TXN-2026-104294) printed on digital and paper receipts',
    card_number VARCHAR(20) NOT NULL COMMENT 'Links target token balance deduction via cards.card_number',
    merchant_id VARCHAR(50) NULL COMMENT 'Identifies vendor company collecting the payment token via merchants.merchant_id',
    terminal_id VARCHAR(50) NULL COMMENT 'Identifies physical ESP32 or terminal node hardware unit triggering the capture via terminals.terminal_id',
    transaction_type ENUM('payment', 'refund', 'reversal', 'topup') DEFAULT 'payment' COMMENT 'Categorizes ledger records to process standard deductions or transaction void mappings cleanly',
    amount DECIMAL(10, 2) NOT NULL COMMENT 'Total Gross fiat amount captured from the card wallet balance tracking column',
    service_fee DECIMAL(10, 2) DEFAULT 0.00 COMMENT 'Platform revenue slice collected by UniCard ecosystem engine per tap processing action',
    net_merchant_payout DECIMAL(10, 2) GENERATED ALWAYS AS (amount - service_fee) STORED COMMENT 'Automatically calculated column tracking exactly how much money goes to the merchant after our platform cut',
    processed_by VARCHAR(50) NULL COMMENT 'Public string identifier users.user_id capturing the identity of the physical staff member operating the payment client terminal',
    status ENUM('pending', 'completed', 'failed') DEFAULT 'completed' COMMENT 'Lifecycle state of the transaction for tracking settlement',
    description VARCHAR(255) NULL COMMENT 'Optional human-readable note or system-generated label describing the transaction context',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Cryptographic server node timestamp securing exactly when transaction settlement clearing finalized',
    FOREIGN KEY (card_number) REFERENCES cards(card_number),
    -- Fixed: was merchants(user_id), now correctly points to merchants(merchant_id)
    FOREIGN KEY (merchant_id) REFERENCES merchants(merchant_id),
    FOREIGN KEY (terminal_id) REFERENCES terminals(terminal_id),
    FOREIGN KEY (processed_by) REFERENCES users(user_id)
) COMMENT='High-growth financial master ledger capturing all terminal token taps, transaction classifications, and system fees';

CREATE TABLE top_ups (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT 'Internal tracking row index scaled to 64-bit headroom to safely accommodate massive historical logging growth',
    topup_id VARCHAR(50) NOT NULL UNIQUE COMMENT 'Custom unique public transaction code (e.g., LD-2026-987153) used for user receipts and payment gateway queries',
    card_number VARCHAR(20) NOT NULL COMMENT 'Links target token balance injection via cards.card_number',
    amount DECIMAL(10, 2) NOT NULL COMMENT 'Gross load amount requested by the customer before convenience charges are applied',
    convenience_fee DECIMAL(10, 2) DEFAULT 0.00 COMMENT 'Ecosystem engine collection fee applied to over-the-air channels like GCash or Maya webhooks',
    gateway_cost DECIMAL(10, 2) DEFAULT 0.00 COMMENT 'Actual fee incurred from external payment providers (GCash, Maya, Bank) per transaction',
    net_gateway_fee DECIMAL(10, 2) GENERATED ALWAYS AS (convenience_fee - gateway_cost) STORED COMMENT 'Automatically calculated net revenue kept by the platform after 3rd party costs',
    total_charged DECIMAL(10, 2) GENERATED ALWAYS AS (amount + convenience_fee) STORED COMMENT 'Automatically calculated column representing the absolute total cash value collected from the external channel source',
    payment_method ENUM('cash', 'gcash', 'maya', 'over_the_counter', 'stripe') NOT NULL COMMENT 'Drives system tracking to audit cash-drawer liquid positions against programmatic API callbacks',
    handled_by VARCHAR(50) NULL COMMENT 'Public user_id string identifier referencing the administrative staff member who manually accepted physical bills if OTC cash-loaded',
    status ENUM('pending', 'completed', 'failed') DEFAULT 'completed' COMMENT 'State pipeline tracker handling payment gateway processing exceptions or drops',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Auto-generated clock timestamp mapping exactly when wallet balance credits were finalized',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Tracks chronological life cycle changes, such as a top-up shifting from pending to completed',
    FOREIGN KEY (card_number) REFERENCES cards(card_number),
    FOREIGN KEY (handled_by) REFERENCES users(user_id)
) COMMENT='High-growth balance loader ledger maintaining immutable compliance auditing for all incoming ecosystem liquidity channels';