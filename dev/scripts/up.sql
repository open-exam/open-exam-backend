CREATE DATABASE  IF NOT EXISTS `open_exam` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;
USE `open_exam`;
-- MySQL dump 10.13  Distrib 8.0.27, for Linux (x86_64)
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
  CONSTRAINT `exam_scopes_scope_group` FOREIGN KEY (`scope`) REFERENCES `groups` (`id`),
  CONSTRAINT `exam_scopes_scope_org` FOREIGN KEY (`scope`) REFERENCES `organizations` (`id`),
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
  CONSTRAINT `exam_template_scopes_group` FOREIGN KEY (`scope`) REFERENCES `groups` (`id`),
  CONSTRAINT `exam_template_scopes_org` FOREIGN KEY (`scope`) REFERENCES `organizations` (`id`),
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
  `start_time` bigint unsigned NOT NULL,
  `end_time` bigint unsigned NOT NULL,
  `duration` int unsigned NOT NULL,
  `created_by` varchar(64) NOT NULL,
  `template` varchar(64) NOT NULL,
  `org` bigint unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  KEY `exams_template_idx` (`template`),
  KEY `exams_org_linked_idx` (`org`),
  CONSTRAINT `exams_created_by` FOREIGN KEY (`id`) REFERENCES `users` (`id`),
  CONSTRAINT `exams_linked_template` FOREIGN KEY (`template`) REFERENCES `exam_template` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `exams_org_linked` FOREIGN KEY (`org`) REFERENCES `organizations` (`id`)
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
  CONSTRAINT `org_id` FOREIGN KEY (`org_id`) REFERENCES `organizations` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
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
  `resource` varchar(64) NOT NULL,
  `description` varchar(256) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `organizations`
--

DROP TABLE IF EXISTS `organizations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `organizations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `plugins`
--

DROP TABLE IF EXISTS `plugins`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `plugins` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `url` varchar(512) NOT NULL,
  `version` varchar(32) NOT NULL,
  `organization` bigint unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `plugins_org_idx` (`organization`),
  CONSTRAINT `plugins_org` FOREIGN KEY (`organization`) REFERENCES `organizations` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pool`
--

DROP TABLE IF EXISTS `pool`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pool` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pool_questions`
--

DROP TABLE IF EXISTS `pool_questions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pool_questions` (
  `pool_id` bigint unsigned NOT NULL,
  `question_id` bigint unsigned NOT NULL,
  KEY `pool_questions_pool_id_idx` (`pool_id`),
  KEY `pool_questions_question_id` (`question_id`),
  CONSTRAINT `pool_questions_pool_id` FOREIGN KEY (`pool_id`) REFERENCES `pool` (`id`),
  CONSTRAINT `pool_questions_question_id` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pool_scopes`
--

DROP TABLE IF EXISTS `pool_scopes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pool_scopes` (
  `pool_id` bigint unsigned NOT NULL,
  `scope` bigint unsigned NOT NULL,
  `scope_type` int unsigned NOT NULL,
  KEY `pool_scopes_pool_id_idx` (`pool_id`),
  KEY `pool_scopes_org_idx` (`scope`),
  CONSTRAINT `pool_scopes_group` FOREIGN KEY (`scope`) REFERENCES `groups` (`id`),
  CONSTRAINT `pool_scopes_org` FOREIGN KEY (`scope`) REFERENCES `organizations` (`id`),
  CONSTRAINT `pool_scopes_pool_id` FOREIGN KEY (`pool_id`) REFERENCES `pool` (`id`),
  CONSTRAINT `pool_scopes_team` FOREIGN KEY (`scope`) REFERENCES `teams` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pooled_set`
--

DROP TABLE IF EXISTS `pooled_set`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pooled_set` (
  `set_id` bigint unsigned NOT NULL,
  `pool_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`set_id`,`pool_id`),
  KEY `pooled_set_pool_id_idx` (`pool_id`),
  CONSTRAINT `pooled_set_pool_id` FOREIGN KEY (`pool_id`) REFERENCES `pool` (`id`),
  CONSTRAINT `pooled_set_set_id` FOREIGN KEY (`set_id`) REFERENCES `sets` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `question_scopes`
--

DROP TABLE IF EXISTS `question_scopes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `question_scopes` (
  `question_id` bigint unsigned NOT NULL,
  `scope` bigint unsigned NOT NULL,
  `scope_type` int unsigned NOT NULL,
  KEY `question_scopes_question_id` (`question_id`),
  KEY `question_scopes_org_idx` (`scope`),
  CONSTRAINT `question_scopes_group` FOREIGN KEY (`scope`) REFERENCES `groups` (`id`),
  CONSTRAINT `question_scopes_org` FOREIGN KEY (`scope`) REFERENCES `organizations` (`id`),
  CONSTRAINT `question_scopes_question_id` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`),
  CONSTRAINT `question_scopes_team` FOREIGN KEY (`scope`) REFERENCES `teams` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `questions`
--

DROP TABLE IF EXISTS `questions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `questions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `plugin_id` bigint unsigned NOT NULL,
  `title` varchar(64) NOT NULL,
  `display_data` json NOT NULL,
  `answer_data` json NOT NULL,
  `files` json NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
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
  `oper_id` bigint unsigned NOT NULL,
  `scope` bigint unsigned NOT NULL,
  `scope_type` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  KEY `user_id_rbac_idx` (`user_id`),
  KEY `rbac_oper_id_idx` (`oper_id`),
  KEY `rbac_scope_org_idx` (`scope`),
  CONSTRAINT `rbac_oper_id` FOREIGN KEY (`oper_id`) REFERENCES `operations` (`id`),
  CONSTRAINT `rbac_scope_group` FOREIGN KEY (`scope`) REFERENCES `groups` (`id`),
  CONSTRAINT `rbac_scope_org` FOREIGN KEY (`scope`) REFERENCES `organizations` (`id`),
  CONSTRAINT `rbac_scope_team` FOREIGN KEY (`scope`) REFERENCES `teams` (`id`),
  CONSTRAINT `user_id_rbac` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `set_questions`
--

DROP TABLE IF EXISTS `set_questions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `set_questions` (
  `set_id` bigint unsigned NOT NULL,
  `question_id` bigint unsigned NOT NULL,
  KEY `set_questions_set_id_idx` (`set_id`),
  KEY `set_questions_question_id` (`question_id`),
  CONSTRAINT `set_questions_question_id` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`),
  CONSTRAINT `set_questions_set_id` FOREIGN KEY (`set_id`) REFERENCES `sets` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `set_scopes`
--

DROP TABLE IF EXISTS `set_scopes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `set_scopes` (
  `set_id` bigint unsigned NOT NULL,
  `scope` bigint unsigned NOT NULL,
  `scope_type` int unsigned NOT NULL,
  KEY `set_scopes_set_id` (`set_id`),
  KEY `set_scopes_org_idx` (`scope`),
  CONSTRAINT `set_scopes_group` FOREIGN KEY (`scope`) REFERENCES `groups` (`id`),
  CONSTRAINT `set_scopes_org` FOREIGN KEY (`scope`) REFERENCES `organizations` (`id`),
  CONSTRAINT `set_scopes_set_id` FOREIGN KEY (`set_id`) REFERENCES `sets` (`id`),
  CONSTRAINT `set_scopes_team` FOREIGN KEY (`scope`) REFERENCES `teams` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sets`
--

DROP TABLE IF EXISTS `sets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sets` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  PRIMARY KEY (`id`)
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
  KEY `standard_users_scope_organizations_idx` (`scope`),
  CONSTRAINT `standard_users_scope_groups` FOREIGN KEY (`scope`) REFERENCES `groups` (`id`),
  CONSTRAINT `standard_users_scope_organizations` FOREIGN KEY (`scope`) REFERENCES `organizations` (`id`),
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
  `id` varchar(64) NOT NULL,
  `team_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `team_id_idx` (`team_id`),
  CONSTRAINT `team_id` FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `user_id` FOREIGN KEY (`id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
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
  `display_name` varchar(45) NOT NULL,
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
  `name` varchar(128) NOT NULL,
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

-- Dump completed on 2021-11-02 15:01:06
