-- updated sql file for unicard

CREATE DATABASE IF NOT EXISTS unicard;
USE unicard;

--
-- Table structure for table `cards`
--

DROP TABLE IF EXISTS `cards`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cards` (
  `card_number` varchar(20) NOT NULL COMMENT 'The visible consumer-facing identifier printed on the physical plastic token',
  `card_uid` varchar(50) NOT NULL COMMENT 'The physical hardware chip unique UID read directly from the MIFARE/RFID sectors',
  `user_id` varchar(50) DEFAULT NULL COMMENT 'Links cardholder account identity via the public users.user_id string identifier',
  `card_type` enum('regular','student','pwd','senior') DEFAULT 'regular' COMMENT 'Drives dynamic discount calculation algorithms across transportation fares',
  `discount_verified` tinyint(1) DEFAULT '0' COMMENT 'Flags whether regulatory documents were verified for fare discount tier eligibility',
  `balance` decimal(10,2) DEFAULT '0.00' COMMENT 'Current secure stored monetary value assigned to the physical card unit',
  `loyalty_points` decimal(10,2) DEFAULT '0.00' COMMENT 'Accrued transaction reward points redeemable at verified retail merchant stations',
  `status` enum('active','inactive','blocked','lost') DEFAULT 'inactive' COMMENT 'Lifecycle block status constraint to instantly freeze stolen or missing tokens',
  `expiry_date` date NOT NULL COMMENT 'Expiration threshold date determining card block lifecycle validations',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp tracking when the physical token record was registered in inventory',
  `linked_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp tracking the exact moment the card was registered to a specific user',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Automatically updates when balances adjust or card status switches occur',
  PRIMARY KEY (`card_number`),
  UNIQUE KEY `card_uid` (`card_uid`),
  KEY `user_id` (`user_id`),
  CONSTRAINT `cards_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`user_id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Ecosystem transit wallet asset tracker maintaining balances, hardware mapping tokens, and fare tier flags';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `merchants`
--

DROP TABLE IF EXISTS `merchants`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchants` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'Internal merchant row index used for fast database indexing',
  `merchant_id` varchar(50) NOT NULL COMMENT 'Custom public identifier for the business entity (e.g., MCH-2026-001)',
  `business_name` varchar(150) NOT NULL COMMENT 'Registered trade or company name of the client merchant',
  `business_type` enum('retail','transportation','food_and_beverage','services','other') NOT NULL COMMENT 'Industry category for transaction filtering and analytics',
  `business_registration_number` varchar(100) DEFAULT NULL COMMENT 'Official government tracking number (e.g., DTI, SEC, or BIR TIN)',
  `business_address` text NOT NULL COMMENT 'Physical location of the main store or corporate headquarters',
  `city` varchar(100) DEFAULT NULL COMMENT 'City or municipality where the primary business is physically located',
  `postal_code` varchar(20) DEFAULT NULL COMMENT 'Postal or ZIP code of the primary business address location',
  `user_id` varchar(50) NOT NULL COMMENT 'Links to the user_id in the users table who owns this business account',
  `owner_name` varchar(100) NOT NULL COMMENT 'Full name of the principal owner or authorized business representative',
  `business_email` varchar(100) NOT NULL COMMENT 'Official company contact email address for corporate updates and billing statements',
  `business_phone` varchar(20) NOT NULL COMMENT 'Official telephone or mobile number for merchant support and emergency updates',
  `commission_rate` decimal(5,2) DEFAULT '2.00' COMMENT 'Percentage cut taken by UniCard per processed card transaction (e.g., 2.50 = 2.5%)',
  `settlement_account_name` varchar(100) DEFAULT NULL COMMENT 'The name on the merchant bank account or mobile wallet for payouts',
  `settlement_account_number` varchar(50) DEFAULT NULL COMMENT 'The actual bank account number or mobile number (GCash/Maya) for payouts',
  `settlement_bank_name` varchar(100) DEFAULT NULL COMMENT 'The target bank or e-wallet company name (e.g., BDO, BPI, GCash, Maya)',
  `status` enum('pending approval','approved','rejected','active','suspended','deleted') DEFAULT 'pending approval',
  `business_document` varchar(255) DEFAULT NULL COMMENT 'Primary business registration document path (DTI or SEC depending on business_structure)',
  `bir_document` varchar(255) DEFAULT NULL COMMENT 'File path for the uploaded BIR registration document',
  `valid_id` varchar(255) DEFAULT NULL COMMENT 'File path for the uploaded government-issued valid ID of the owner',
  `bank_document` varchar(255) DEFAULT NULL COMMENT 'File path for the uploaded bank document or passbook for settlement verification',
  `business_structure` enum('sole_proprietorship','partnership','corporation','cooperative') DEFAULT NULL COMMENT 'Legal structure of the business entity used to determine required registration documents',
  `document_status` enum('pending','approved','rejected') DEFAULT 'pending' COMMENT 'Verification state of the submitted business registration and compliance documents',
  `message` text COMMENT 'Admin-written note or feedback message regarding document review or account status changes',
  `approved_by` varchar(50) DEFAULT NULL COMMENT 'The user_id of the Super Admin who verified and activated this company profile',
  `approved_at` timestamp NULL DEFAULT NULL COMMENT 'The specific date and timestamp when the business was activated',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Auto-generated date and time record of the initial registration request',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Automatically updates whenever any merchant profile field is modified',
  PRIMARY KEY (`id`),
  UNIQUE KEY `merchant_id` (`merchant_id`),
  UNIQUE KEY `business_email` (`business_email`),
  UNIQUE KEY `business_phone` (`business_phone`),
  UNIQUE KEY `business_registration_number` (`business_registration_number`),
  KEY `user_id` (`user_id`),
  KEY `approved_by` (`approved_by`),
  CONSTRAINT `merchants_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`user_id`) ON DELETE RESTRICT,
  CONSTRAINT `merchants_ibfk_2` FOREIGN KEY (`approved_by`) REFERENCES `users` (`user_id`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Enterprise business registry tracking partner tenants, hardware mapping nodes, and financial settlement details';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `system_settings`
--

DROP TABLE IF EXISTS `system_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_settings` (
  `setting_key` varchar(50) NOT NULL COMMENT 'Unique string configuration key acting as the primary look-up token',
  `setting_value` varchar(255) NOT NULL COMMENT 'The active parameter threshold or value parsed directly by the Go backend',
  `description` text COMMENT 'Descriptive documentation notes detailing exactly what system rules or parameters this alters',
  `updated_by` varchar(50) NOT NULL COMMENT 'The public users.user_id of the Super Admin who executed the latest configuration adjustment override',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Auto-generated clock timestamp tracking when this specific configuration parameter was initialized',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Automatically locks the exact clock timestamp whenever this system parameter value is updated',
  PRIMARY KEY (`setting_key`),
  KEY `updated_by` (`updated_by`),
  CONSTRAINT `system_settings_ibfk_1` FOREIGN KEY (`updated_by`) REFERENCES `users` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Global platform configuration matrix driving dynamic fees, operational bounds, and system constants';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `terminal_requests`
--

DROP TABLE IF EXISTS `terminal_requests`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `terminal_requests` (
  `id` int NOT NULL AUTO_INCREMENT,
  `request_id` varchar(50) NOT NULL,
  `merchant_id` varchar(50) NOT NULL,
  `terminal_sn` varchar(50) DEFAULT NULL,
  `status` enum('pending','approved','rejected') DEFAULT 'pending',
  `requested_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `handled_by` varchar(50) DEFAULT NULL,
  `handled_at` timestamp NULL DEFAULT NULL,
  `notes` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `request_id` (`request_id`),
  KEY `merchant_id` (`merchant_id`),
  CONSTRAINT `terminal_requests_ibfk_1` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`merchant_id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Tracks merchant-initiated terminal assignment requests';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `terminals`
--

DROP TABLE IF EXISTS `terminals`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `terminals` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'Internal hardware registry auto-increment row index',
  `terminal_id` varchar(50) NOT NULL COMMENT 'Custom public hardware identifier (e.g., TRM-2026-0001) used in API payloads',
  `terminal_sn` varchar(50) NOT NULL COMMENT 'Physical factory-assigned unique serial number or MAC address of the ESP32 board',
  `merchant_id` varchar(50) DEFAULT NULL COMMENT 'Links to the merchant_id of the managing merchant entity',
  `device_name` varchar(100) NOT NULL COMMENT 'Human-readable descriptor identifying placement (e.g., Counter 1, Jeepney Plate # ABC-123)',
  `location_details` varchar(255) DEFAULT NULL COMMENT 'Optional physical sector data, such as a branch route path or stall number designation',
  `status` enum('active','suspended','inactive') DEFAULT 'inactive' COMMENT 'Operational network connectivity state of the edge node hardware',
  `last_heartbeat` timestamp NULL DEFAULT NULL COMMENT 'Tracks the precise timestamp of the last successful ping packet received from the ESP32 network stack',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Auto-generated clock timestamp tracking initial edge device registration',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Automatically monitors configuration adjustments or state transitions over time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `terminal_id` (`terminal_id`),
  UNIQUE KEY `terminal_sn` (`terminal_sn`),
  KEY `merchant_id` (`merchant_id`),
  CONSTRAINT `terminals_ibfk_1` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`merchant_id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=17 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Hardware node registry tracking deployed physical authentication nodes and network heartbeat states';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `top_ups`
--

DROP TABLE IF EXISTS `top_ups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `top_ups` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'Internal tracking row index scaled to 64-bit headroom to safely accommodate massive historical logging growth',
  `topup_id` varchar(50) NOT NULL COMMENT 'Custom unique public transaction code (e.g., LD-2026-987153) used for user receipts and payment gateway queries',
  `card_number` varchar(20) NOT NULL COMMENT 'Links target token balance injection via cards.card_number',
  `amount` decimal(10,2) NOT NULL COMMENT 'Gross load amount requested by the customer before convenience charges are applied',
  `convenience_fee` decimal(10,2) DEFAULT '0.00' COMMENT 'Ecosystem engine collection fee applied to over-the-air channels like GCash or Maya webhooks',
  `gateway_cost` decimal(10,2) DEFAULT '0.00' COMMENT 'Actual fee incurred from external payment providers (GCash, Maya, Bank) per transaction',
  `net_gateway_fee` decimal(10,2) GENERATED ALWAYS AS ((`convenience_fee` - `gateway_cost`)) STORED COMMENT 'Automatically calculated net revenue kept by the platform after 3rd party costs',
  `total_charged` decimal(10,2) GENERATED ALWAYS AS ((`amount` + `convenience_fee`)) STORED COMMENT 'Automatically calculated column representing the absolute total cash value collected from the external channel source',
  `payment_method` enum('cash','gcash','maya','over_the_counter','xendit') NOT NULL COMMENT 'Drives system tracking to audit cash-drawer liquid positions against programmatic API callbacks',
  `handled_by` varchar(50) DEFAULT NULL COMMENT 'Public user_id string identifier referencing the administrative staff member who manually accepted physical bills if OTC cash-loaded',
  `external_id` varchar(255) DEFAULT NULL COMMENT 'External reference ID from payment gateways like Xendit to prevent double processing',
  `status` enum('pending','completed','failed','expired') DEFAULT 'pending' COMMENT 'pending=invoice created, completed=paid, failed=payment error, expired=unpaid timeout',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Auto-generated clock timestamp mapping exactly when wallet balance credits were finalized',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Tracks chronological life cycle changes, such as a top-up shifting from pending to completed',
  PRIMARY KEY (`id`),
  UNIQUE KEY `topup_id` (`topup_id`),
  UNIQUE KEY `external_id` (`external_id`),
  KEY `card_number` (`card_number`),
  CONSTRAINT `top_ups_ibfk_1` FOREIGN KEY (`card_number`) REFERENCES `cards` (`card_number`)
) ENGINE=InnoDB AUTO_INCREMENT=32 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='High-growth balance loader ledger maintaining immutable compliance auditing for all incoming ecosystem liquidity channels';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `transactions`
--

DROP TABLE IF EXISTS `transactions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `transactions` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'Internal financial primary index scaled to 64-bit headroom to comfortably support billions of platform entries',
  `transaction_id` varchar(50) NOT NULL COMMENT 'Custom unique public reference string (e.g., TXN-2026-104294) printed on digital and paper receipts',
  `card_number` varchar(20) DEFAULT NULL COMMENT 'Links target token balance deduction via cards.card_number',
  `merchant_id` varchar(50) DEFAULT NULL COMMENT 'Identifies vendor company collecting the payment token via merchants.merchant_id',
  `terminal_id` varchar(50) DEFAULT NULL COMMENT 'Identifies physical ESP32 or terminal node hardware unit triggering the capture via terminals.terminal_id',
  `transaction_type` enum('payment','refund','reversal','topup','withdrawal') DEFAULT 'payment' COMMENT 'Categorizes ledger records to process standard deductions or transaction void mappings cleanly',
  `amount` decimal(10,2) DEFAULT NULL COMMENT 'Total Gross fiat amount captured from the card wallet balance tracking column',
  `points_earned` decimal(10,2) DEFAULT NULL COMMENT 'Total points earned from the transaction',
  `service_fee` decimal(10,2) DEFAULT NULL COMMENT 'Platform revenue slice collected by UniCard ecosystem engine per tap processing action',
  `net_merchant_payout` decimal(10,2) GENERATED ALWAYS AS ((`amount` - `service_fee`)) STORED COMMENT 'Automatically calculated column tracking exactly how much money goes to the merchant after our platform cut',
  `processed_by` varchar(50) DEFAULT NULL COMMENT 'Public string identifier users.user_id capturing the identity of the physical staff member operating the payment client terminal',
  `status` enum('pending','completed','failed','expired') DEFAULT 'pending' COMMENT 'pending=invoice created, completed=paid, failed=payment error, expired=unpaid timeout',
  `description` varchar(255) DEFAULT NULL COMMENT 'Optional human-readable note or system-generated label describing the transaction context',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Cryptographic server node timestamp securing exactly when transaction settlement clearing finalized',
  PRIMARY KEY (`id`),
  UNIQUE KEY `transaction_id` (`transaction_id`),
  KEY `card_number` (`card_number`),
  CONSTRAINT `transactions_ibfk_1` FOREIGN KEY (`card_number`) REFERENCES `cards` (`card_number`)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='High-growth financial master ledger capturing all terminal token taps, transaction classifications, and system fees';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'Internal row index optimized for database indexing and fast joins',
  `user_id` varchar(50) NOT NULL COMMENT 'Custom public ID (e.g., UNI-YYMM-minsecxxxx) used in APIs and frontend',
  `username` varchar(50) NOT NULL COMMENT 'Unique handle for admin/staff to log in quickly without an email',
  `name` varchar(100) NOT NULL COMMENT 'Full name of the individual user or client contact person',
  `email` varchar(100) NOT NULL COMMENT 'Primary email address used for consumer logins and notifications',
  `phone_number` varchar(20) DEFAULT NULL COMMENT 'Mobile number (e.g., +639...) for OTPs and SMS transaction alerts',
  `password_hash` varchar(255) NOT NULL COMMENT 'Cryptographically secured password string handled via bcrypt in Go',
  `role` enum('super_admin','merchant_admin','merchant_staff','customer') NOT NULL COMMENT 'Defines application-wide role-based access control',
  `status` enum('active','suspended','inactive') DEFAULT 'active' COMMENT 'Account access state for platform security and compliance checks',
  `pending_email` varchar(100) DEFAULT NULL COMMENT 'Temporary email storage for account recovery and verification changes',
  `email_verification_token` varchar(255) DEFAULT NULL COMMENT 'Token for verifying email address updates',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Auto-generated timestamp of account creation',
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `phone_number` (`phone_number`),
  UNIQUE KEY `pending_email` (`pending_email`),
  UNIQUE KEY `email_verification_token` (`email_verification_token`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Core identity table tracking authentication and tenancy access control levels';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;