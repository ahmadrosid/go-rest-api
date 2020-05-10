CREATE DATABASE IF NOT EXISTS `go_products`;

USE `go_products`;

CREATE TABLE IF NOT EXISTS `products` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL DEFAULT '0',
  `price` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
);