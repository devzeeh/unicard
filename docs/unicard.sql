CREATE DATABASE  IF NOT EXISTS `unicard` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;
USE `unicard`;
-- MySQL dump 10.13  Distrib 8.0.44, for Win64 (x86_64)
--
-- Host: 127.0.0.1    Database: unicard
-- ------------------------------------------------------
-- Server version	8.0.44

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `admin_users`
--

DROP TABLE IF EXISTS `admin_users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `admin_users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `username` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `full_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Bcrypt password hash',
  `role` enum('Super admin','Admin','Support','Viewer') COLLATE utf8mb4_unicode_ci DEFAULT 'Admin',
  `status` enum('Active','Inactive','Suspended') COLLATE utf8mb4_unicode_ci DEFAULT 'Active',
  `last_login` timestamp NULL DEFAULT NULL,
  `login_attempts` int DEFAULT '0',
  `locked_until` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`),
  KEY `idx_username` (`username`),
  KEY `idx_email` (`email`),
  KEY `idx_role` (`role`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='System administrator accounts';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `admin_users`
--

LOCK TABLES `admin_users` WRITE;
/*!40000 ALTER TABLE `admin_users` DISABLE KEYS */;
INSERT INTO `admin_users` VALUES (1,'devzeeh','roxas.johnerrol@gmail.com','john errol','$2a$12$tjycPwp4svJajA6cAIywK.61wL/236Eht6E/1PgyyG1zQJM3KnWue','Admin','Active',NULL,0,NULL,'2026-05-20 02:01:11','2026-05-20 02:01:11');
/*!40000 ALTER TABLE `admin_users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `card_reports`
--

DROP TABLE IF EXISTS `card_reports`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `card_reports` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `card_number` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Printed 16-digit card number',
  `reason` enum('Lost','Stolen','Fraud','User request','Admin action') COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `reported_by` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'User ID of the person who filed the report (user or admin)',
  `status` enum('Reported','Resolved','Replaced') COLLATE utf8mb4_unicode_ci NOT NULL,
  `replacement_card_number` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `blocked_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `resolved_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `replacement_card_number` (`replacement_card_number`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_card_number` (`card_number`),
  KEY `idx_status` (`status`),
  CONSTRAINT `card_reports_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`user_id`) ON DELETE CASCADE,
  CONSTRAINT `card_reports_ibfk_2` FOREIGN KEY (`card_number`) REFERENCES `cards` (`card_number`),
  CONSTRAINT `card_reports_ibfk_3` FOREIGN KEY (`replacement_card_number`) REFERENCES `cards` (`card_number`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Blocked and lost card records';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `card_reports`
--

LOCK TABLES `card_reports` WRITE;
/*!40000 ALTER TABLE `card_reports` DISABLE KEYS */;
/*!40000 ALTER TABLE `card_reports` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `cards`
--

DROP TABLE IF EXISTS `cards`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cards` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `card_uid` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'RFID/NFC Tag ID (Internal)',
  `card_holder` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `card_number` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Printed 16-digit card number',
  `card_type` enum('Regular','PWD','Student','Senior') COLLATE utf8mb4_unicode_ci DEFAULT 'Regular',
  `expiry_date` date DEFAULT NULL,
  `cvv` varchar(3) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Security code (optional storage)',
  `initial_amount` decimal(10,2) DEFAULT '0.00' COMMENT 'Amount to credit user upon registration (Set first by ADMIN)',
  `status` enum('Active','Inactive','Blocked','Lost','Expired') COLLATE utf8mb4_unicode_ci DEFAULT 'Inactive',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'When card was added to inventory',
  `is_primary` tinyint(1) DEFAULT '0' COMMENT 'Is this the primary card for the user?',
  `linked_at` timestamp NULL DEFAULT NULL COMMENT 'Card was linked to user',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `card_uid` (`card_uid`),
  UNIQUE KEY `card_number` (`card_number`),
  KEY `user_id` (`user_id`),
  KEY `idx_card_uid` (`card_uid`),
  KEY `idx_card_number` (`card_number`),
  KEY `idx_status` (`status`),
  KEY `idx_card_holder` (`card_holder`),
  CONSTRAINT `cards_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`user_id`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Master inventory of physical cards';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `cards`
--

LOCK TABLES `cards` WRITE;
/*!40000 ALTER TABLE `cards` DISABLE KEYS */;
INSERT INTO `cards` VALUES (1,'840570851385','A346F101','errol roxas','2621020251171655','Regular','2036-04-09',NULL,100.00,'Active','2026-02-14 08:50:13',0,NULL,'2026-04-09 09:33:05'),(3,'173886440021','93C13FED','john errol devss','2621028394671655','Regular','2036-02-21',NULL,100.00,'Active','2026-02-21 08:18:02',0,NULL,'2026-04-09 09:40:23'),(4,'615921401740','7310BCEC','john dev','2621020252341655','Regular','2036-02-21',NULL,100.00,'Active','2026-02-21 08:19:52',0,'2026-04-09 10:24:58','2026-04-09 10:24:58'),(5,'573343578549','53A327EC','maja salvador','2621028700527532','Regular','2036-02-21',NULL,100.00,'Active','2026-02-21 08:33:19',0,'2026-04-09 10:35:51','2026-04-09 10:35:51'),(6,'427079423054','F3BF44EC','julia baretto','2621062102879616','Regular','2036-04-09',NULL,100.00,'Active','2026-02-21 08:37:42',0,'2026-04-09 10:40:59','2026-04-09 10:40:59'),(7,'799817330001','038981EC','frances cruz','2621062102292676','Regular','2036-04-28',NULL,100.00,'Active','2026-02-21 08:41:57',0,'2026-04-28 06:05:53','2026-04-28 06:05:53'),(8,'849645895086','2067BE14','ash cue','2621028885926930','Regular','2036-04-29',NULL,100.00,'Active','2026-02-21 08:45:59',0,'2026-04-29 04:17:44','2026-04-29 04:17:44'),(9,'242070109521','50B32514','mhissy acosta','2621027620974570','Regular','2036-04-29',NULL,100.00,'Active','2026-02-21 08:47:26',0,'2026-04-29 04:31:54','2026-04-29 04:31:54'),(10,'100957470319','43006A19','lauren yen','2621020251771655','Regular','2036-05-01',NULL,100.00,'Active','2026-02-21 08:52:17',0,'2026-05-01 03:50:25','2026-05-01 03:50:25'),(11,NULL,'0005584041',NULL,'2601050762341385','Regular','2036-05-01',NULL,100.00,'Inactive','2026-05-01 04:08:10',0,NULL,'2026-05-01 04:08:10');
/*!40000 ALTER TABLE `cards` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `fare_settings`
--

DROP TABLE IF EXISTS `fare_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `fare_settings` (
  `id` int NOT NULL AUTO_INCREMENT,
  `merchant_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `business_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `base_fare` decimal(10,2) DEFAULT '13.00',
  `per_km_rate` decimal(10,2) DEFAULT '1.50',
  `minimum_fare` decimal(10,2) DEFAULT '13.00',
  `pwd_discount_rate` decimal(5,2) DEFAULT '20.00',
  `student_discount_rate` decimal(5,2) DEFAULT '20.00',
  `senior_discount_rate` decimal(5,2) DEFAULT '20.00',
  `loyalty_rate` decimal(5,4) DEFAULT '0.0020' COMMENT '0.0020 = 0.2% cashback',
  `route_number` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'For transport: route or plate number',
  `effective_from` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `effective_to` timestamp NULL DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT '1',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `merchant_code` (`merchant_code`),
  KEY `idx_route_number` (`route_number`),
  KEY `idx_merchant_code` (`merchant_code`),
  KEY `idx_is_active` (`is_active`),
  CONSTRAINT `fare_settings_ibfk_1` FOREIGN KEY (`merchant_code`) REFERENCES `merchants` (`merchant_code`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Configurable fare and discount settings';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `fare_settings`
--

LOCK TABLES `fare_settings` WRITE;
/*!40000 ALTER TABLE `fare_settings` DISABLE KEYS */;
/*!40000 ALTER TABLE `fare_settings` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `loyalty_redemptions`
--

DROP TABLE IF EXISTS `loyalty_redemptions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `loyalty_redemptions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `transaction_id` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `points_used` decimal(10,2) NOT NULL,
  `value_php` decimal(10,2) NOT NULL COMMENT 'PHP value of redeemed points',
  `status` enum('Pending','Completed','Cancelled') COLLATE utf8mb4_unicode_ci DEFAULT 'Completed',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `transaction_id` (`transaction_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_transaction_id` (`transaction_id`),
  CONSTRAINT `loyalty_redemptions_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`user_id`) ON DELETE CASCADE,
  CONSTRAINT `loyalty_redemptions_ibfk_2` FOREIGN KEY (`transaction_id`) REFERENCES `transactions` (`user_id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Loyalty points redemption history';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `loyalty_redemptions`
--

LOCK TABLES `loyalty_redemptions` WRITE;
/*!40000 ALTER TABLE `loyalty_redemptions` DISABLE KEYS */;
/*!40000 ALTER TABLE `loyalty_redemptions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `merchants`
--

DROP TABLE IF EXISTS `merchants`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchants` (
  `id` int NOT NULL AUTO_INCREMENT,
  `merchant_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Public merchant identifier',
  `business_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `business_type` enum('Transport','Retail','Both') COLLATE utf8mb4_unicode_ci NOT NULL,
  `owner_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `address` text COLLATE utf8mb4_unicode_ci,
  `route_number` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'For transport: route or plate number',
  `store_location` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'For retail: store address',
  `commission_rate` decimal(5,2) DEFAULT '0.00' COMMENT 'Commission percentage',
  `settlement_account` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Bank account for settlements',
  `status` enum('Active','Suspended','Inactive','Removed') COLLATE utf8mb4_unicode_ci DEFAULT 'Active',
  `verified` tinyint(1) DEFAULT '0' COMMENT 'Verified status indicates if the merchant has completed KYC/verification',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `merchant_code` (`merchant_code`),
  KEY `idx_merchant_code` (`merchant_code`),
  KEY `idx_business_type` (`business_type`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Merchant accounts (drivers and store owners)';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `merchants`
--

LOCK TABLES `merchants` WRITE;
/*!40000 ALTER TABLE `merchants` DISABLE KEYS */;
/*!40000 ALTER TABLE `merchants` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `receipts`
--

DROP TABLE IF EXISTS `receipts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `receipts` (
  `id` int NOT NULL AUTO_INCREMENT,
  `transaction_id` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `receipt_number` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `receipt_type` enum('Payment','Topup','Refund') COLLATE utf8mb4_unicode_ci NOT NULL,
  `email_sent` tinyint(1) DEFAULT '0',
  `email_sent_at` timestamp NULL DEFAULT NULL,
  `email_opened` tinyint(1) DEFAULT '0',
  `pdf_generated` tinyint(1) DEFAULT '0',
  `pdf_data` longblob,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `receipt_number` (`receipt_number`),
  KEY `idx_transaction_id` (`transaction_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_receipt_number` (`receipt_number`),
  CONSTRAINT `receipts_ibfk_1` FOREIGN KEY (`transaction_id`) REFERENCES `transactions` (`user_id`) ON DELETE CASCADE,
  CONSTRAINT `receipts_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`user_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='E-receipt generation and tracking';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `receipts`
--

LOCK TABLES `receipts` WRITE;
/*!40000 ALTER TABLE `receipts` DISABLE KEYS */;
/*!40000 ALTER TABLE `receipts` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `transactions`
--

DROP TABLE IF EXISTS `transactions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `transactions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `card_number` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL,
  `transaction_type` enum('Topup','Payment','Refund','Adjustment') COLLATE utf8mb4_unicode_ci NOT NULL,
  `category` enum('Transport','Retail','Other') COLLATE utf8mb4_unicode_ci NOT NULL,
  `amount` decimal(10,2) NOT NULL COMMENT 'Final transaction amount',
  `balance_before` decimal(10,2) NOT NULL,
  `balance_after` decimal(10,2) NOT NULL,
  `discount_amount` decimal(10,2) DEFAULT '0.00' COMMENT 'If amount is 8.00 and discount is 2.00, the original price was 10.00.',
  `discount_type` enum('Regular','PWD','Student','Senior','Promo') COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Reason for discount (PWD, Student, Senior, Promo)',
  `points_earned` decimal(10,2) DEFAULT '0.00',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT 'Optional notes or item details (e.g., "Monthly Pass")',
  `location` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Route number or store name(Human-readable location (e.g., "Main St. Coffee" or "Bus 42")',
  `merchant_code` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Link to the merchant who received payment (Nullable for system adjustments)',
  `terminal_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'ID of the physical device/reader used (for debugging/tracking)',
  `payment_method` enum('Stripe','E-Wallet','Bank','Cash','Points') COLLATE utf8mb4_unicode_ci NOT NULL,
  `reference_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'External/Internal REF ID (e.g. Stripe Charge ID, GCash Ref No.)',
  `transaction_id` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` enum('Pending','Completed','Failed','Refunded') COLLATE utf8mb4_unicode_ci DEFAULT 'Completed',
  `failure_reason` text COLLATE utf8mb4_unicode_ci COMMENT 'Error message if status is "Failed" (e.g., "Insufficient funds")',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `completed_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_card_number` (`card_number`),
  KEY `idx_transaction_type` (`transaction_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_merchant_code` (`merchant_code`),
  KEY `idx_reference_id` (`reference_id`),
  KEY `idx_transaction_id` (`transaction_id`),
  CONSTRAINT `transactions_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`user_id`) ON DELETE CASCADE,
  CONSTRAINT `transactions_ibfk_2` FOREIGN KEY (`merchant_code`) REFERENCES `merchants` (`merchant_code`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='All transaction records (payments, top-ups, refunds)';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `transactions`
--

LOCK TABLES `transactions` WRITE;
/*!40000 ALTER TABLE `transactions` DISABLE KEYS */;
/*!40000 ALTER TABLE `transactions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Public user ID (e.g., account number)',
  `username` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Login username',
  `full_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `password_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Bcrypt password hash',
  `card_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'RFID card unique identifier (internal)',
  `card_number` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '16-digit printed card number',
  `user_type` enum('Regular','PWD','Student','Senior') COLLATE utf8mb4_unicode_ci DEFAULT 'Regular',
  `balance` decimal(10,2) DEFAULT '0.00' COMMENT 'Current card balance in PHP',
  `loyalty_points` decimal(10,2) DEFAULT '0.00' COMMENT 'Accumulated loyalty points',
  `status` enum('Active','Blocked','Inactive') COLLATE utf8mb4_unicode_ci DEFAULT 'Active',
  `id_verified` tinyint(1) DEFAULT '0' COMMENT 'Has user uploaded valid ID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `last_login` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `card_id` (`card_id`),
  UNIQUE KEY `card_number` (`card_number`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_username` (`username`),
  KEY `idx_email` (`email`),
  KEY `idx_card_number` (`card_number`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=52 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User accounts and card information';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (1,'840570851385','user26fxvolv05729','errol roxas','one.devteam.25@gmail.com','09952329743','$2a$10$qpIlzDQLDucvSdNWZy/R3OMdhonkG2bo75WyHZjz8QnOB4NasmbMa','CARD-021426fgxRf9Q','459543887671','Regular',100.00,0.00,'Active',0,'2026-02-13 20:57:29','2026-02-21 04:33:59',NULL),(43,'173886440021','user26ehrzr3e2121','john errol devss','devzeeh@gmail.com','09123456789','$2a$10$1cDLHp3cfBoPtXYzGjkMquqv/whHWE/qlHLFNKgp6kQfZ/IsG8yJi','CARD-040926nzAdIIc','2621028394671655','Regular',100.00,0.00,'Active',0,'2026-04-08 20:21:21','2026-04-17 07:14:57',NULL),(44,'615921401740','user26ilbbt392458','john dev','jjohnroxas06@gmail.com','09123456788','$2a$10$ETcFIEEED8jmkwaK3OicpueNulxYPCEJToeKZjZWTQsA8lSNWHkR.','CARD-040926mcnG6F7','2621020252341655','Regular',100.00,0.00,'Active',0,'2026-04-08 22:24:58','2026-05-20 01:49:20',NULL),(45,'573343578549','user262evjp0v3551','maja salvador','majasalvador@gmail.com','09123123412','$2a$12$u3ySLle7MOBG/2FeOjS96eyCxITcjD10LvYPDOP9pkuvxLe6JHVx2','CARD-040926EvJOB49','2621028700527532','Regular',100.00,0.00,'Active',0,'2026-04-08 22:35:51','2026-04-14 05:08:25',NULL),(47,'427079423054','user26t5tp7ys4059','julia baretto','juliabaretto@gmail.com','0987654321','$2a$10$SnUcFjvugos89TQFqJqDm.6bKCtcrNOdAwvxdLMnvaMU8nrBiMDf.','CARD-040926VQUXMFZ','2621062102879616','Regular',100.00,0.00,'Active',0,'2026-04-08 22:40:59','2026-04-09 10:40:59',NULL),(48,'799817330001','user266iynhsu0553','frances cruz','francescruz@gmail.com','09987654321','$2a$12$xC4wv64kjhMF8Rt6dswkFOleSaU5olPTameBMh5uDX1rGLV6nHEDi','CARD-042826Z2nMhwE','2621062102292676','Regular',100.00,0.00,'Active',0,'2026-04-27 18:05:53','2026-04-28 06:29:20',NULL),(49,'849645895086','user26xx95elc1744','ash cue','jjohnroxas06+dev@gmail.com','09987252723','$2a$10$Hbu8bQ1iCBuIE8U5LCanOeDlqY8WHhE7bNzN6HIHvtSncBMi5fLoC','CARD-042926N1gZYZs','2621028885926930','Regular',100.00,0.00,'Active',0,'2026-04-29 04:17:44','2026-04-29 04:17:44',NULL),(50,'242070109521','user26i89wuxv3154','mhissy acosta','missacosta@gmail.com','09987526323','$2a$10$BimoLj3vyJqFcpquZ5LcauViuR0ZJwwFa0zEG/176yVNbBSTGy7/G','CARD-042926zX5v3kZ','2621027620974570','Regular',100.00,0.00,'Active',0,'2026-04-29 04:31:54','2026-04-29 04:31:54',NULL),(51,'100957470319','user26fw59znn5025','lauren yen','laurenyen@gmail.com','09765434526','$2a$10$/AOys2.oPFXDPOjNQDcBM.vG1WNAF0k3Ej1YvYerqzbIUy5upa8Pa','CARD-050126FFAanLE','2621020251771655','Regular',100.00,0.00,'Active',0,'2026-05-01 03:50:25','2026-05-01 03:50:25',NULL);
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-05-20 10:05:59
