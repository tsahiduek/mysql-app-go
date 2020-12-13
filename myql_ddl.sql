DROP TABLE IF EXISTS `tickets`;
CREATE TABLE IF NOTE EXISTS `tickets` (
  `id` int(6) unsigned NOT NULL AUTO_INCREMENT,
  `user` varchar(8) NOT NULL,
  `description` varchar(30) NOT NULL,
  `details` varchar(300) NOT NULL,
  `cc_list` varchar(72) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=UTF8;

