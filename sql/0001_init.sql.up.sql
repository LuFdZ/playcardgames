CREATE TABLE `users` (
  `user_id`       INT          NOT NULL AUTO_INCREMENT,
  `username`      VARCHAR(64)  NOT NULL,
  `password`      VARCHAR(64)  NOT NULL,
  `nickname`      VARCHAR(64)           DEFAULT NULL,
  `mobile`        VARCHAR(16)           DEFAULT NULL,
  `email`         VARCHAR(128) NOT NULL,
  `avatar`        VARCHAR(128)          DEFAULT NULL,
  `status`        INT          NOT NULL DEFAULT '0',
  `channel`       VARCHAR(64)           DEFAULT NULL,
  `login_type`    INT                   DEFAULT NULL,
  `created_at`    DATETIME     NOT NULL,
  `updated_at`    DATETIME     NOT NULL,
  `last_login_at` DATETIME              DEFAULT NULL,
  `rights`        INT          NOT NULL DEFAULT '0',
  `sex`           INT          NOT NULL DEFAULT NULL,
  `icon`          VARCHAR(128) NOT NULL DEFAULT NULL,
  PRIMARY KEY (`user_id`),
  UNIQUE KEY `email_unique` (`email`),
  UNIQUE KEY `username_unique` (`username`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 100000
  DEFAULT CHARSET = utf8;

CREATE TABLE `balances` (
  `user_id`        INT      NOT NULL,
  `deposit`        BIGINT   NOT NULL DEFAULT '0',
  `freeze`         BIGINT   NOT NULL DEFAULT '0',
  `gold`           BIGINT   NOT NULL DEFAULT '0',
  `diamond`        BIGINT   NOT NULL DEFAULT '0',
  `amount`         BIGINT   NOT NULL DEFAULT '0',
  `gold_profit`    BIGINT   NOT NULL DEFAULT '0',
  `diamond_profit` BIGINT   NOT NULL DEFAULT '0',
  `created_at`     DATETIME NOT NULL,
  `updated_at`     DATETIME NOT NULL,
  PRIMARY KEY (`user_id`),
  KEY `created_at_index` (`created_at`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `journals` (
  `id`         INT      NOT NULL AUTO_INCREMENT,
  `user_id`    INT      NOT NULL,
  `gold`       BIGINT   NOT NULL DEFAULT '0',
  `diamond`    BIGINT   NOT NULL DEFAULT '0',
  `amount`     BIGINT   NOT NULL DEFAULT '0',
  `type`       INT      NOT NULL,
  `foreign`    BIGINT   NOT NULL,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  KEY `created_at_index` (`created_at`),
  UNIQUE KEY `foreign_type_index` (`type`, `foreign`),
  PRIMARY KEY (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `deposits` (
  `id`         INT      NOT NULL AUTO_INCREMENT,
  `user_id`    INT      NOT NULL,
  `amount`     BIGINT   NOT NULL DEFAULT '0',
  `created_at` DATETIME NOT NULL,
  `type`       INT      NOT NULL,
  PRIMARY KEY (`id`),
  KEY `created_at_index` (`created_at`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;


CREATE TABLE `configs` (
  `id`          INT          NOT NULL,
  `name`        VARCHAR(64)  NOT NULL,
  `description` VARCHAR(128) NOT NULL,
  `value`       VARCHAR(512) NOT NULL,
  `created_at`  DATETIME     NOT NULL,
  `updated_at`  DATETIME     NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name_unique` (`name`),
  KEY `created_at_index` (`created_at`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `activity_configs` (
  `config_id`           INT          NOT NULL,
  `description`         VARCHAR(60)  NOT NULL,
  `parameter`           VARCHAR(400) NOT NULL,
  `last_modify_user_id` INT          NOT NULL,
  `created_at`          DATETIME     NOT NULL,
  `updated_at`          DATETIME     NOT NULL,
  PRIMARY KEY (`config_id`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `rooms` (
  `room_id`        INT          NOT NULL AUTO_INCREMENT,
  `password`       VARCHAR(16)  NOT NULL,
  `user_list`      VARCHAR(255) NOT NULL DEFAULT '',
  `player_max_num` INT          NOT NULL DEFAULT '0',
  `status`         INT          NOT NULL DEFAULT '0',
  `game_type`      INT          NOT NULL DEFAULT '0',
  `created_at`     DATETIME     NOT NULL,
  `updated_at`     DATETIME     NOT NULL,
  PRIMARY KEY (`room_id`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `thirteen` (
  `game_id`         INT          NOT NULL AUTO_INCREMENT,
  `room_id`         INT          NOT NULL DEFAULT '0',
  `user_id_list` VARCHAR(255) NOT NULL DEFAULT '',
  `status`          INT          NOT NULL DEFAULT '0',
  `user_score_list` VARCHAR(255) NOT NULL DEFAULT '',
  `created_at`      DATETIME     NOT NULL,
  `updated_at`      DATETIME     NOT NULL,
  PRIMARY KEY (`game_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `thirteen_user_log` (
  `game_id`         INT          NOT NULL AUTO_INCREMENT,
  `user_id` VARCHAR(255) NOT NULL DEFAULT '',
  `room_id`         INT          NOT NULL DEFAULT '0',
  `user_card_list` VARCHAR(255) NOT NULL DEFAULT '',
  `score`         INT          NOT NULL DEFAULT '0',
  `status`          INT          NOT NULL DEFAULT '0',
  `created_at`      DATETIME     NOT NULL,
  `updated_at`      DATETIME     NOT NULL,
  PRIMARY KEY (`game_id`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;
