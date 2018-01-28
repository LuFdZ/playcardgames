CREATE TABLE `doudizhus` (
  `game_id`            INT           NOT NULL AUTO_INCREMENT,
  `room_id`            INT           NOT NULL DEFAULT '0',
  `index`              INT           NOT NULL DEFAULT '0',
  `banker_id`          INT           NOT NULL DEFAULT '0',
  `dizhu_card_str`     VARCHAR(60)   NOT NULL DEFAULT '',
  `user_card_info_str` VARCHAR(1600) NOT NULL DEFAULT '',
  `game_result_str`    VARCHAR(500)  NOT NULL DEFAULT '',
  `game_card_log_str`  VARCHAR(1500) NOT NULL DEFAULT '',
  `get_banker_log_str` VARCHAR(1500) NOT NULL DEFAULT '',
  `restart_times`      INT           NOT NULL DEFAULT '0',
  `status`             INT           NOT NULL DEFAULT '0',
  `base_score`         INT           NOT NULL DEFAULT '0',
  `banker_times`       INT           NOT NULL DEFAULT '0',
  `bomb_times`         INT           NOT NULL DEFAULT '0',
  `op_index`           INT           NOT NULL DEFAULT '0',
  `winer_id`           INT           NOT NULL DEFAULT '0',
  `winer_type`         INT           NOT NULL DEFAULT '0',
  `spring`             INT           NOT NULL DEFAULT '0',
  `common_bomb`        INT           NOT NULL DEFAULT '0',
  `rocket_bomb`        INT           NOT NULL DEFAULT '0',
  `eight_bomb`         INT           NOT NULL DEFAULT '0',
  `op_date_at`         DATETIME      NOT NULL,
  `created_at`         DATETIME      NOT NULL,
  `updated_at`         DATETIME      NOT NULL,
  PRIMARY KEY (`game_id`),
  KEY `idx_status` (`status`),
  KEY `idx_room` (`room_id`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `fourcards` (
  `game_id`          INT           NOT NULL AUTO_INCREMENT,
  `room_id`          INT           NOT NULL DEFAULT '0',
  `index`            INT           NOT NULL DEFAULT '0',
  `banker_id`        INT           NOT NULL DEFAULT '0',
  `game_result_str` VARCHAR(3000)  NOT NULL DEFAULT '',
  `status`           INT           NOT NULL DEFAULT '0',
  `op_date_at`       DATETIME      NOT NULL,
  `created_at`       DATETIME      NOT NULL,
  `updated_at`       DATETIME      NOT NULL,
  PRIMARY KEY (`game_id`),
  KEY `idx_status` (`status`),
  KEY `idx_room` (`room_id`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `mail_infos` (
  `mail_id`          INT           NOT NULL ,
  `mail_name`        VARCHAR(50)   NOT NULL DEFAULT '',
  `mail_title`       VARCHAR(50)   NOT NULL DEFAULT '',
  `mail_content`     VARCHAR(500)  NOT NULL DEFAULT '',
  `mail_type`        INT           NOT NULL DEFAULT '0',
  `status`           INT           NOT NULL DEFAULT '0',
  `item_list`        VARCHAR(500)  NOT NULL DEFAULT '',
  `created_at`       DATETIME      NOT NULL,
  `updated_at`       DATETIME      NOT NULL,
  PRIMARY KEY (`mail_id`),
  KEY `idx_type` (`mail_type`),
  KEY `idx_created`(`created_at`),
  UNIQUE KEY `name_unique` (`mail_name`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;



CREATE TABLE `mail_send_logs` (
  `send_log_id`      INT           NOT NULL AUTO_INCREMENT,
  `mail_id`          INT           NOT NULL DEFAULT '0',
  `send_id`          INT           NOT NULL DEFAULT '0',
  `mail_type`        INT           NOT NULL DEFAULT '0',
  `status`           INT           NOT NULL DEFAULT '0',
  `mail_str`         VARCHAR(1500)  NOT NULL DEFAULT '',
  `send_count`       INT           NOT NULL DEFAULT '0',
  `total_count`      INT           NOT NULL DEFAULT '0',
  `count`            INT           NOT NULL DEFAULT '0',
  `created_at`       DATETIME      NOT NULL,
  `updated_at`       DATETIME      NOT NULL,
  PRIMARY KEY (`send_log_id`),
  KEY `idx_id` (`mail_id`),
  KEY `idx_send` (`send_id`),
  KEY `idx_type` (`mail_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `player_mails` (
  `player_mail_id`   INT           NOT NULL AUTO_INCREMENT,
  `send_log_id`      INT           NOT NULL DEFAULT '0',
  `mail_id`          INT           NOT NULL DEFAULT '0',
  `user_id`          INT           NOT NULL DEFAULT '0',
  `send_id`          INT           NOT NULL DEFAULT '0',
  `mail_type`        INT           NOT NULL DEFAULT '0',
  `have_item`        INT           NOT NULL DEFAULT '0',
  `status`           INT           NOT NULL DEFAULT '0',
  `created_at`       DATETIME      NOT NULL,
  `updated_at`       DATETIME      NOT NULL,
   PRIMARY KEY (`player_mail_id`),
  KEY `idx_id` (`mail_id`),
  KEY `idx_send_log` (`send_log_id`),
  KEY `idx_user` (`user_id`),
  KEY `idx_send` (`send_id`),
  KEY `idx_type` (`mail_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

INSERT INTO mail_infos VALUES (1001,"房间退出通知", "房间退出", "房间%d于%s退出！", 3, "", now(), now());
alter table rooms add shuffle int default 0 not null after giveup;
alter table users add type int default 0 not null after status;
ALTER  TABLE  `users`  ADD  INDEX idx_type (`type`);
