CREATE TABLE `doudizhus` (
  `game_id`           INT           NOT NULL AUTO_INCREMENT,
  `room_id`           INT           NOT NULL DEFAULT '0',
  `index`             INT           NOT NULL DEFAULT '0',
  `banker_type`       INT           NOT NULL DEFAULT '0',
  `banker_id`         INT           NOT NULL DEFAULT '0',
  `user_cards_remain` VARCHAR(1100) NOT NULL DEFAULT '',
  `user_cards`        VARCHAR(1100) NOT NULL DEFAULT '',
  `last_op_id`        INT           NOT NULL DEFAULT '0',
  `game_results`      VARCHAR(1500) NOT NULL DEFAULT '',
  `status`            INT           NOT NULL DEFAULT '0',
  `bomb_times`        INT           NOT NULL DEFAULT '0',
  `op_date_at`        DATETIME      NOT NULL,
  `created_at`        DATETIME      NOT NULL,
  `updated_at`        DATETIME      NOT NULL,
  PRIMARY KEY (`game_id`),
  KEY `idx_status` (`status`),
  KEY `idx_room` (`room_id`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;


CREATE TABLE `doudizhus_log` (
  `log_id`        INT           NOT NULL AUTO_INCREMENT,
  `room_id`       INT           NOT NULL DEFAULT '0',
  `game_id`       INT           NOT NULL DEFAULT '0',
  `user_id`       INT           NOT NULL DEFAULT '0',
  `user_card_log` VARCHAR(1500) NOT NULL DEFAULT '',
  `created_at`    DATETIME      NOT NULL,
  `updated_at`    DATETIME      NOT NULL,
  PRIMARY KEY (`log_id`),
  KEY `idx_game` (`game_id`),
  KEY `idx_room` (`room_id`),
  KEY `idx_user` (`user_id`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;