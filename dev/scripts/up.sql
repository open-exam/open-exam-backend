DROP DATABASE IF EXISTS `open_exam`;
CREATE DATABASE  IF NOT EXISTS `open_exam` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;
USE `open_exam`;
-- MySQL dump 10.13  Distrib 8.0.26, for Linux (x86_64)
--
-- Host: localhost    Database: open_exam
-- ------------------------------------------------------
-- Server version	8.0.26

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
-- Table structure for table `custom_teams`
--

DROP TABLE IF EXISTS `custom_teams`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `custom_teams` (
  `id` bigint unsigned NOT NULL,
  `user_id` varchar(64) NOT NULL,
  KEY `custom_teams_id_idx` (`id`),
  KEY `custom_teams_user_id_idx` (`user_id`),
  CONSTRAINT `custom_teams_id` FOREIGN KEY (`id`) REFERENCES `custom_teams_list` (`id`),
  CONSTRAINT `custom_teams_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `custom_teams_list`
--

DROP TABLE IF EXISTS `custom_teams_list`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `custom_teams_list` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `exam_client_access`
--

DROP TABLE IF EXISTS `exam_client_access`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `exam_client_access` (
  `id` varchar(128) NOT NULL,
  `user_id` varchar(64) NOT NULL,
  `exam_id` varchar(64) NOT NULL,
  `expiry` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  KEY `exam_id_idx` (`exam_id`),
  CONSTRAINT `exam_id` FOREIGN KEY (`exam_id`) REFERENCES `exams` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `exam_scopes`
--

DROP TABLE IF EXISTS `exam_scopes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `exam_scopes` (
  `exam_id` varchar(64) NOT NULL,
  `scope` bigint unsigned NOT NULL,
  `scope_type` int unsigned NOT NULL,
  KEY `exam_scopes_scope_custom_idx` (`scope`),
  KEY `exam_scopes_exam_id_idx` (`exam_id`),
  CONSTRAINT `exam_scopes_exam_id` FOREIGN KEY (`exam_id`) REFERENCES `exams` (`id`),
  CONSTRAINT `exam_scopes_scope_custom` FOREIGN KEY (`scope`) REFERENCES `custom_teams_list` (`id`),
  CONSTRAINT `exam_scopes_scope_group` FOREIGN KEY (`scope`) REFERENCES `groups` (`id`),
  CONSTRAINT `exam_scopes_scope_org` FOREIGN KEY (`scope`) REFERENCES `organisations` (`id`),
  CONSTRAINT `exam_scopes_scope_team` FOREIGN KEY (`scope`) REFERENCES `teams` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `exam_template`
--

DROP TABLE IF EXISTS `exam_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `exam_template` (
  `id` varchar(64) NOT NULL,
  `name` varchar(128) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `exam_template_scopes`
--

DROP TABLE IF EXISTS `exam_template_scopes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `exam_template_scopes` (
  `exam_template_id` varchar(64) NOT NULL,
  `scope` bigint unsigned NOT NULL,
  `scope_type` int unsigned NOT NULL,
  KEY `exam_template_id_idx` (`exam_template_id`),
  KEY `exam_template_scopes_org_idx` (`scope`),
  CONSTRAINT `exam_template_id` FOREIGN KEY (`exam_template_id`) REFERENCES `exam_template` (`id`),
  CONSTRAINT `exam_template_scopes_custom` FOREIGN KEY (`scope`) REFERENCES `custom_teams_list` (`id`),
  CONSTRAINT `exam_template_scopes_group` FOREIGN KEY (`scope`) REFERENCES `groups` (`id`),
  CONSTRAINT `exam_template_scopes_org` FOREIGN KEY (`scope`) REFERENCES `organisations` (`id`),
  CONSTRAINT `exam_template_scopes_team` FOREIGN KEY (`scope`) REFERENCES `teams` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `exams`
--

DROP TABLE IF EXISTS `exams`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `exams` (
  `id` varchar(64) NOT NULL,
  `name` varchar(128) NOT NULL,
  `start_time` bigint NOT NULL,
  `end_time` bigint NOT NULL,
  `duration` int NOT NULL,
  `template` varchar(64) NOT NULL,
  `created_by` varchar(64) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  KEY `exams_template_idx` (`template`),
  CONSTRAINT `exams_created_by` FOREIGN KEY (`id`) REFERENCES `users` (`id`),
  CONSTRAINT `exams_linked_template` FOREIGN KEY (`template`) REFERENCES `exam_template` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `groups`
--

DROP TABLE IF EXISTS `groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `groups` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned NOT NULL,
  `name` varchar(128) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  KEY `org_id_idx` (`org_id`),
  CONSTRAINT `org_id` FOREIGN KEY (`org_id`) REFERENCES `organisations` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `operations`
--

DROP TABLE IF EXISTS `operations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `operations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `operation` varchar(64) NOT NULL,
  `description` varchar(256) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `organisations`
--

DROP TABLE IF EXISTS `organisations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `organisations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `question_bank`
--

DROP TABLE IF EXISTS `question_bank`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `question_bank` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(64) NOT NULL,
  `display_data` json NOT NULL,
  `answer_data` json NOT NULL,
  `files` json NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `question_bank_scopes`
--

DROP TABLE IF EXISTS `question_bank_scopes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `question_bank_scopes` (
  `question_bank_id` bigint unsigned NOT NULL,
  `scope` bigint unsigned NOT NULL,
  `scope_type` bigint unsigned NOT NULL,
  KEY `question_bank_scopes_org_idx` (`scope`),
  KEY `question_bank_scopes_id` (`question_bank_id`),
  CONSTRAINT `question_bank_scopes_group` FOREIGN KEY (`scope`) REFERENCES `groups` (`id`),
  CONSTRAINT `question_bank_scopes_id` FOREIGN KEY (`question_bank_id`) REFERENCES `question_bank` (`id`),
  CONSTRAINT `question_bank_scopes_org` FOREIGN KEY (`scope`) REFERENCES `organisations` (`id`),
  CONSTRAINT `question_bank_scopes_team` FOREIGN KEY (`scope`) REFERENCES `teams` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `rbac`
--

DROP TABLE IF EXISTS `rbac`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `rbac` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` varchar(64) NOT NULL,
  `role_id` bigint unsigned NOT NULL,
  `item_id` bigint unsigned NOT NULL,
  `type` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  KEY `role_id_rbac_idx` (`role_id`),
  KEY `role_item_id_rbac_idx` (`item_id`),
  KEY `role_type_rbac_idx` (`type`),
  KEY `user_id_rbac_idx` (`user_id`),
  CONSTRAINT `role_id_rbac` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`),
  CONSTRAINT `role_item_id_rbac` FOREIGN KEY (`item_id`) REFERENCES `roles` (`item_id`),
  CONSTRAINT `role_type_rbac` FOREIGN KEY (`type`) REFERENCES `roles` (`type`),
  CONSTRAINT `user_id_rbac` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `roles`
--

DROP TABLE IF EXISTS `roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `roles` (
  `id` bigint unsigned NOT NULL,
  `item_id` bigint unsigned NOT NULL,
  `type` int unsigned NOT NULL,
  PRIMARY KEY (`id`,`type`,`item_id`),
  KEY `item_id_idx` (`item_id`),
  KEY `type_idx` (`type`),
  CONSTRAINT `custom_id_r` FOREIGN KEY (`item_id`) REFERENCES `custom_teams_list` (`id`),
  CONSTRAINT `group_id_r` FOREIGN KEY (`item_id`) REFERENCES `groups` (`id`),
  CONSTRAINT `oper_id_r` FOREIGN KEY (`id`) REFERENCES `operations` (`id`),
  CONSTRAINT `org_id_r` FOREIGN KEY (`item_id`) REFERENCES `organisations` (`id`),
  CONSTRAINT `team_id_r` FOREIGN KEY (`item_id`) REFERENCES `teams` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `standard_users`
--

DROP TABLE IF EXISTS `standard_users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `standard_users` (
  `user_id` varchar(64) NOT NULL,
  `scope` bigint unsigned NOT NULL,
  `scope_type` int unsigned NOT NULL,
  KEY `standard_users_user_id_idx` (`user_id`),
  KEY `standard_users_scope_organisations_idx` (`scope`),
  CONSTRAINT `standard_users_scope_groups` FOREIGN KEY (`scope`) REFERENCES `groups` (`id`),
  CONSTRAINT `standard_users_scope_organisations` FOREIGN KEY (`scope`) REFERENCES `organisations` (`id`),
  CONSTRAINT `standard_users_scope_teams` FOREIGN KEY (`scope`) REFERENCES `teams` (`id`),
  CONSTRAINT `standard_users_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `students`
--

DROP TABLE IF EXISTS `students`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `students` (
  `user_id` varchar(64) NOT NULL,
  `team_id` bigint unsigned NOT NULL,
  `name` varchar(45) DEFAULT NULL,
  `email` varchar(45) DEFAULT NULL,
  `phone` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`user_id`,`team_id`),
  KEY `team_id_idx` (`team_id`),
  CONSTRAINT `team_id` FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `teams`
--

DROP TABLE IF EXISTS `teams`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `teams` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `group_id` bigint unsigned NOT NULL,
  `name` varchar(128) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  KEY `group_id_idx` (`group_id`),
  CONSTRAINT `group_id` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` varchar(64) NOT NULL,
  `email` varchar(320) NOT NULL,
  `type` int unsigned NOT NULL,
  `password` varchar(128) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2021-10-16 17:48:25
