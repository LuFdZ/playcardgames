CREATE TABLE `users` (
  `user_id`         INT          NOT NULL AUTO_INCREMENT,
  `username`        VARCHAR(64)  NOT NULL,
  `password`        VARCHAR(64)  NOT NULL,
  `nickname`        VARCHAR(64)  NOT NULL,
  `mobile`          VARCHAR(16)           DEFAULT NULL,
  `email`           VARCHAR(128) NOT NULL,
  `avatar`          VARCHAR(128)          DEFAULT NULL,
  `status`          INT          NOT NULL DEFAULT '0',
  `channel`         VARCHAR(64)           DEFAULT NULL,
  `version`         VARCHAR(64)           DEFAULT NULL,
  `login_type`      INT                   DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL,
  `updated_at`      DATETIME     NOT NULL,
  `last_login_at`   DATETIME              DEFAULT NULL,
  `rights`          INT          NOT NULL DEFAULT '0',
  `sex`             INT          NOT NULL DEFAULT '0',
  `icon`            VARCHAR(128)          DEFAULT NULL,
  `invite_user_id`  INT          NOT NULL DEFAULT '0',
  `mobile_uu_id`    VARCHAR(30)           DEFAULT NULL,
  `mobile_model`    VARCHAR(20)           DEFAULT NULL,
  `mobile_net_work` VARCHAR(20)           DEFAULT NULL,
  `mobile_os`       VARCHAR(20)           DEFAULT NULL,
  `last_login_ip`   VARCHAR(20)           DEFAULT NULL,
  `reg_ip`          VARCHAR(20)           DEFAULT NULL,
  `open_id`         VARCHAR(30)           DEFAULT NULL,
  `union_id`        VARCHAR(30)           DEFAULT NULL,
  PRIMARY KEY (`user_id`),
#   UNIQUE KEY `email_unique` (`email`),
#   UNIQUE KEY `username_unique` (`username`),
  UNIQUE KEY `openid_unique` (`open_id`)

)
  ENGINE = InnoDB
  AUTO_INCREMENT = 100000
  DEFAULT CHARSET = utf8;

insert into users values(0,"admin@xnhd","67bad3e758b4d324381586f209fee08bca0701396a606f12029425f31cd29ce8","YWRtaW5AeG5oZA==","","","",0,"","",0,now(),now(),now(),131071,1,"",0,"","","","","","","","");

CREATE TABLE `balances` (
  `user_id`                INT      NOT NULL,
  `deposit`                BIGINT   NOT NULL DEFAULT '0',
  `freeze`                 BIGINT   NOT NULL DEFAULT '0',
  `gold`                   BIGINT   NOT NULL DEFAULT '0',
  `diamond`                BIGINT   NOT NULL DEFAULT '0',
  `amount`                 BIGINT   NOT NULL DEFAULT '0',
  `gold_profit`            BIGINT   NOT NULL DEFAULT '0',
  `diamond_profit`         BIGINT   NOT NULL DEFAULT '0',
  `cumulative_recharge`    BIGINT   NOT NULL DEFAULT '0',
  `cumulative_consumption` BIGINT   NOT NULL DEFAULT '0',
  `cumulative_gift`        BIGINT   NOT NULL DEFAULT '0',
  `created_at`             DATETIME NOT NULL,
  `updated_at`             DATETIME NOT NULL,
  PRIMARY KEY (`user_id`),
  KEY `created_at_index` (`created_at`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `journals` (
  `id`         INT          NOT NULL AUTO_INCREMENT,
  `user_id`    INT          NOT NULL,
  `gold`       BIGINT       NOT NULL DEFAULT '0',
  `diamond`    BIGINT       NOT NULL DEFAULT '0',
  `amount`     BIGINT       NOT NULL DEFAULT '0',
  `type`       INT          NOT NULL,
  `foreign`    VARCHAR(128) NOT NULL,
  `channel`    VARCHAR(64)  NOT NULL,
  `created_at` DATETIME     NOT NULL,
  `updated_at` DATETIME     NOT NULL,
  `op_user_id` INT          NOT NULL,
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
  `room_id`          INT           NOT NULL AUTO_INCREMENT,
  `password`         VARCHAR(16)   NOT NULL,
  `user_list`        VARCHAR(3200)  NOT NULL DEFAULT '',
  `max_number`       INT           NOT NULL DEFAULT '0',
  `round_number`     INT           NOT NULL DEFAULT '0',
  `round_now`        INT           NOT NULL DEFAULT '0',
  `status`           INT           NOT NULL DEFAULT '0',
  `giveup`           INT           NOT NULL DEFAULT '0',
  `room_type`        INT           NOT NULL DEFAULT '0',
  `payer_id`         INT           NOT NULL DEFAULT '0',
  `game_type`        INT           NOT NULL DEFAULT '0',
  `game_param`       VARCHAR(255)  NOT NULL DEFAULT '',
  `game_user_result` VARCHAR(2000) NOT NULL DEFAULT '',
  `created_at`       DATETIME      NOT NULL,
  `updated_at`       DATETIME      NOT NULL,
  `giveup_at`        DATETIME      NOT NULL,
  `pre_item_1`       VARCHAR(255)  NOT NULL DEFAULT '',
  `pre_item_2`       VARCHAR(255)  NOT NULL DEFAULT '',
  PRIMARY KEY (`room_id`),
  KEY `idx_status` (`status`),
  KEY `idx_gtype` (`game_type`),
  KEY `idx_rtype` (`room_type`),
  KEY `idx_payer` (`payer_id`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `thirteens` (
  `game_id`           INT           NOT NULL AUTO_INCREMENT,
  `room_id`           INT           NOT NULL DEFAULT '0',
  `banker_id`         INT           NOT NULL DEFAULT '0',
  `index`             INT           NOT NULL DEFAULT '0',
  `user_cards`        VARCHAR(1000) NOT NULL DEFAULT '',
  `user_submit_cards` VARCHAR(900)  NOT NULL DEFAULT '',
  `game_results`      VARCHAR(1500) NOT NULL DEFAULT '',
  `status`            INT           NOT NULL DEFAULT '0',
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


CREATE TABLE `notices` (
  `notice_id`      INT          NOT NULL AUTO_INCREMENT,
  `notice_type`    INT          NOT NULL DEFAULT '0',
  `notice_content` TEXT,
  `channels`       VARCHAR(500) NOT NULL DEFAULT '',
  `versions`       VARCHAR(500) NOT NULL DEFAULT '',
  `status`         INT          NOT NULL DEFAULT '0',
  `description`    VARCHAR(50)  NOT NULL DEFAULT '',
  `param`          VARCHAR(255) NOT NULL DEFAULT '',
  `start_at`       DATETIME     NOT NULL,
  `end_at`         DATETIME     NOT NULL,
  `created_at`     DATETIME     NOT NULL,
  `updated_at`     DATETIME     NOT NULL,
  PRIMARY KEY (`notice_id`),
  KEY `idx_status` (`status`),
  KEY `idx_type` (`notice_type`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;


CREATE TABLE `feedbacks` (
  `feedback_id`     INT      NOT NULL AUTO_INCREMENT,
  `user_id`         INT      NOT NULL DEFAULT '0',
  `channel`         VARCHAR(20)       DEFAULT NULL,
  `version`         VARCHAR(20)       DEFAULT NULL,
  `content`         VARCHAR(500)      DEFAULT NULL,
  `mobile_model`    VARCHAR(20)       DEFAULT NULL,
  `mobile_net_work` VARCHAR(20)       DEFAULT NULL,
  `mobile_os`       VARCHAR(20)       DEFAULT NULL,
  `login_ip`        VARCHAR(20)       DEFAULT NULL,
  `created_at`      DATETIME NOT NULL,
  `updated_at`      DATETIME NOT NULL,
  PRIMARY KEY (`feedback_id`),
  KEY `idx_channel` (`channel`),
  KEY `idx_version` (`version`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;


CREATE TABLE `player_rooms` (
  `log_id`     INT      NOT NULL AUTO_INCREMENT,
  `user_id`    INT      NOT NULL DEFAULT '0',
  `room_id`    INT               DEFAULT NULL DEFAULT '0',
  `game_type`  INT               DEFAULT NULL DEFAULT '0',
  `play_times` INT               DEFAULT NULL DEFAULT '0',
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  PRIMARY KEY (`log_id`),
  KEY `idx_user`(`user_id`),
  KEY `idx_room`(`room_id`),
  KEY `idx_game`(`game_type`),
  KEY `idx_created`(`created_at`),
  UNIQUE KEY `unique_index` (`user_id`, `room_id`, game_type)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;


CREATE TABLE `player_shares` (
  `user_id`        INT      NOT NULL DEFAULT '0',
  `share_times`    INT               DEFAULT NULL DEFAULT '0',
  `total_diamonds` BIGINT            DEFAULT NULL DEFAULT '0',
  `created_at`     DATETIME NOT NULL,
  `updated_at`     DATETIME NOT NULL,
  PRIMARY KEY (`user_id`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `niunius` (
  `game_id`           INT           NOT NULL AUTO_INCREMENT,
  `room_id`           INT           NOT NULL DEFAULT '0',
  `index`             INT           NOT NULL DEFAULT '0',
  `banker_type`       INT           NOT NULL DEFAULT '0',
  `banker_id`         INT           NOT NULL DEFAULT '0',
  `user_cards`        VARCHAR(1000) NOT NULL DEFAULT '',
  `game_results`      VARCHAR(1500) NOT NULL DEFAULT '',
  `status`            INT           NOT NULL DEFAULT '0',
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

CREATE TABLE `configs` (
  `config_id`         INT           NOT NULL AUTO_INCREMENT,
  `channel`           VARCHAR(20)   DEFAULT NULL,
  `version`           VARCHAR(20)   DEFAULT NULL,
  `mobile_os`         VARCHAR(20)   DEFAULT NULL,
  `item_id`           INT           DEFAULT NULL,
  `item_value`        VARCHAR(20)   DEFAULT NULL,
  `status`            INT           NOT NULL DEFAULT '1',
  `description`       VARCHAR(200)  DEFAULT NULL,
  `created_at`        DATETIME      NOT NULL,
  `updated_at`        DATETIME      NOT NULL,
  PRIMARY KEY (`config_id`),
  KEY `idx_channel` (`channel`),
  KEY `idx_version` (`version`),
  KEY `idx_os` (`mobile_os`),
  KEY `idx_created`(`created_at`),
  UNIQUE KEY `config_open_unique` (`channel`,`version`,`mobile_os`,`item_id`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

insert into configs values(0,"","","",100001,"1",1,"全局默认充值开关",now(),now());
insert into configs values(0,"","","",110001,"100",1,"全局默认消费开关",now(),now());