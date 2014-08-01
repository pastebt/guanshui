CREATE TABLE `gsdata` (
  `postid` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `parent` int(11) unsigned NOT NULL DEFAULT '0',
  `thread` int(11) unsigned NOT NULL DEFAULT '0',
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `visit` int(11) unsigned NOT NULL DEFAULT '0',
  `flag` tinyint(2) unsigned NOT NULL DEFAULT '0',
  `subject` varchar(1024) DEFAULT '',
  `user` smallint(11) unsigned NOT NULL DEFAULT '0',
  `only` smallint(11) unsigned NOT NULL DEFAULT '0',
  `body` text,
  `size` bigint(11) unsigned NOT NULL DEFAULT '0',
  `afile` varchar(1024) DEFAULT '',
  `atype` varchar(255) DEFAULT '',
  PRIMARY KEY (`postid`),
  KEY `t` (`thread`)
) ENGINE=MyISAM AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

insert into gsdata (postid) values (1);     -- reserve space to summary


CREATE TABLE `gsuser` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL DEFAULT 'noname',
  `pwd` varchar(255) NOT NULL DEFAULT '',
  `mail` varchar(255) DEFAULT '',
  `salt` varchar(255) NOT NULL DEFAULT '',
  `last` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  PRIMARY KEY (`id`),
  UNIQUE KEY `n` (`name`),
  KEY `l` (`last`)
) ENGINE=MyISAM AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;


CREATE TABLE `gslast` (
  `thread` int(11) unsigned NOT NULL DEFAULT '0',
  `newest` int(11) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`thread`),
  UNIQUE KEY `n` (`newest`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;    
