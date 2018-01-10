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
  `game_result_str` VARCHAR(3000) NOT NULL DEFAULT '',
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

