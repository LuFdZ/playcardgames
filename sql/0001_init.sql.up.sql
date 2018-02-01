CREATE TABLE `users` (
  `user_id`         INT          NOT NULL AUTO_INCREMENT,
  `username`        VARCHAR(64)  NOT NULL,
  `password`        VARCHAR(64)  NOT NULL,
  `nickname`        VARCHAR(64)  NOT NULL,
  `mobile`          VARCHAR(16)           DEFAULT NULL,
  `email`           VARCHAR(128) NOT NULL,
  `avatar`          VARCHAR(128)          DEFAULT NULL,
  `status`          INT          NOT NULL DEFAULT '0',
  `type`            INT          NOT NULL DEFAULT '0',
  `channel`         VARCHAR(64)           DEFAULT NULL,
  `version`         VARCHAR(64)           DEFAULT NULL,
  `login_type`      INT                   DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL,
  `updated_at`      DATETIME     NOT NULL,
  `last_login_at`   DATETIME              DEFAULT NULL,
  `rights`          INT          NOT NULL DEFAULT '0',
  `sex`             INT          NOT NULL DEFAULT '0',
  `icon`            VARCHAR(256)          DEFAULT NULL,
  `invite_user_id`  INT          NOT NULL DEFAULT '0',
  `club_id`         INT          NOT NULL DEFAULT '0',
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
  UNIQUE KEY `username_unique` (`username`),
  UNIQUE KEY `openid_unique` (`open_id`),
  UNIQUE KEY `invite_unique` (`invite_user_id`),
  UNIQUE KEY `invite_unique` (`invite_user_id`),
  UNIQUE KEY `type_unique` (`type`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 100000
  DEFAULT CHARSET = utf8;

CREATE TABLE `balances` (
  `id`         INT      NOT NULL AUTO_INCREMENT,
  `user_id`    INT      NOT NULL,
  `deposit`    BIGINT   NOT NULL DEFAULT '0',
  `freeze`     BIGINT   NOT NULL DEFAULT '0',
  `coin_type`  INT      NOT NULL DEFAULT '0',
  `amount`     BIGINT   NOT NULL DEFAULT '0',
  `balance`    BIGINT   NOT NULL DEFAULT '0',
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `balance_unique` (`user_id`, `coin_type`),
  KEY `created_at_index` (`created_at`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `journals` (
  `id`            INT         NOT NULL AUTO_INCREMENT,
  `user_id`       INT         NOT NULL,
  `coin_type`     INT         NOT NULL DEFAULT '0',
  `amount`        BIGINT      NOT NULL DEFAULT '0',
  `amount_before` BIGINT      NOT NULL DEFAULT '0',
  `amount_after`  BIGINT      NOT NULL DEFAULT '0',
  `type`          INT         NOT NULL,
  `foreign`       VARCHAR(64) NOT NULL DEFAULT '0',
  `channel`       VARCHAR(64) NOT NULL,
  `created_at`    DATETIME    NOT NULL,
  `updated_at`    DATETIME    NOT NULL,
  `op_user_id`    INT         NOT NULL,
  KEY `created_at_index` (`created_at`),
  #UNIQUE KEY `foreign_type_index` (`coin_type`, `type`, `foreign`),
  PRIMARY KEY (`id`),
  #KEY `idx_amount_type`(`coin_type`),
  KEY `idx_opid`(`op_user_id`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `rooms` (
  `room_id`          INT           NOT NULL AUTO_INCREMENT,
  `password`         VARCHAR(16)   NOT NULL,
  `user_list`        VARCHAR(3200) NOT NULL DEFAULT '',
  `max_number`       INT           NOT NULL DEFAULT '0',
  `round_number`     INT           NOT NULL DEFAULT '0',
  `round_now`        INT           NOT NULL DEFAULT '0',
  `status`           INT           NOT NULL DEFAULT '0',
  `giveup`           INT           NOT NULL DEFAULT '0',
  `Shuffle`          INT           NOT NULL DEFAULT '0',
  `room_type`        INT           NOT NULL DEFAULT '0',
  `payer_id`         INT           NOT NULL DEFAULT '0',
  `game_type`        INT           NOT NULL DEFAULT '0',
  `game_param`       VARCHAR(255)  NOT NULL DEFAULT '',
  `game_user_result` VARCHAR(3000) NOT NULL DEFAULT '',
  `created_at`       DATETIME      NOT NULL,
  `updated_at`       DATETIME      NOT NULL,
  `giveup_at`        DATETIME      NOT NULL,
  `flag`             INT           NOT NULL DEFAULT '0',
  `club_id`          INT           NOT NULL DEFAULT '0',
  `cost`             BIGINT        NOT NULL DEFAULT '0',
  `cost_type`        INT           NOT NULL DEFAULT '0',
  `pre_item_1`       VARCHAR(255)  NOT NULL DEFAULT '',
  `pre_item_2`       VARCHAR(255)  NOT NULL DEFAULT '',
  PRIMARY KEY (`room_id`),
  KEY `idx_status` (`status`),
  KEY `idx_gtype` (`game_type`),
  KEY `idx_rtype` (`room_type`),
  KEY `idx_payer` (`payer_id`),
  KEY `idx_club` (`club_id`),
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
  `user_cards`        VARCHAR(600)  NOT NULL DEFAULT '',
  `user_submit_cards` VARCHAR(600)  NOT NULL DEFAULT '',
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
  `room_type`  INT               DEFAULT NULL DEFAULT '0',
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
  `game_id`      INT           NOT NULL AUTO_INCREMENT,
  `room_id`      INT           NOT NULL DEFAULT '0',
  `index`        INT           NOT NULL DEFAULT '0',
  `banker_type`  INT           NOT NULL DEFAULT '0',
  `banker_id`    INT           NOT NULL DEFAULT '0',
  `user_cards`   VARCHAR(600)  NOT NULL DEFAULT '',
  `game_results` VARCHAR(3000) NOT NULL DEFAULT '',
  `status`       INT           NOT NULL DEFAULT '0',
  `op_date_at`   DATETIME      NOT NULL,
  `created_at`   DATETIME      NOT NULL,
  `updated_at`   DATETIME      NOT NULL,
  PRIMARY KEY (`game_id`),
  KEY `idx_status` (`status`),
  KEY `idx_room` (`room_id`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `configs` (
  `config_id`   INT      NOT NULL AUTO_INCREMENT,
  `channel`     VARCHAR(20)       DEFAULT NULL,
  `version`     VARCHAR(20)       DEFAULT NULL,
  `mobile_os`   VARCHAR(20)       DEFAULT NULL,
  `item_id`     INT               DEFAULT NULL,
  `item_value`  VARCHAR(20)       DEFAULT NULL,
  `status`      INT      NOT NULL DEFAULT '1',
  `description` VARCHAR(200)      DEFAULT NULL,
  `created_at`  DATETIME NOT NULL,
  `updated_at`  DATETIME NOT NULL,
  PRIMARY KEY (`config_id`),
  KEY `idx_channel` (`channel`),
  KEY `idx_version` (`version`),
  KEY `idx_os` (`mobile_os`),
  KEY `idx_created`(`created_at`),
  UNIQUE KEY `config_open_unique` (`channel`, `version`, `mobile_os`, `item_id`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `clubs` (
  `club_id`       INT          NOT NULL AUTO_INCREMENT,
  `club_name`     VARCHAR(128) NOT NULL,
  `status`        VARCHAR(128) NOT NULL, #'俱乐部状态 -1审核中 1正常 2冻结',
  `creator_id`    INT          NOT NULL,
  `creator_proxy` INT          NOT NULL DEFAULT '0',
  `club_remark`   VARCHAR(200),
  `icon`          VARCHAR(128)          DEFAULT NULL,
  `gold`          BIGINT       NOT NULL DEFAULT '0',
  `diamond`       BIGINT       NOT NULL DEFAULT '0',
  `club_param`    VARCHAR(128) NOT NULL,
  `created_at`    DATETIME     NOT NULL,
  `updated_at`    DATETIME     NOT NULL,
  PRIMARY KEY (`club_id`),
  UNIQUE KEY `name_unique` (`club_name`),
  KEY `idx_created`(`created_at`),
  KEY `idx_status`(`status`),
  KEY `idx_name`(`club_name`),
  KEY `idx_creator`(`creator_id`),
  KEY `idx_proxy`(`creator_proxy`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 10000
  DEFAULT CHARSET = utf8;

CREATE TABLE `club_journals` (
  `id`            INT      NOT NULL AUTO_INCREMENT,
  `club_id`       INT      NOT NULL,
  `amount_type`   BIGINT   NOT NULL DEFAULT '0',
  `amount`        BIGINT   NOT NULL DEFAULT '0',
  `amount_before` BIGINT   NOT NULL DEFAULT '0',
  `amount_after`  BIGINT   NOT NULL DEFAULT '0',
  `type`          INT      NOT NULL,
  `foreign`       BIGINT   NOT NULL DEFAULT '0',
  `created_at`    DATETIME NOT NULL,
  `updated_at`    DATETIME NOT NULL,
  `op_user_id`    INT      NOT NULL,
  KEY `created_at_index` (`created_at`),
  #UNIQUE KEY `foreign_type_index` (`type`, `foreign`),
  PRIMARY KEY (`id`),
  KEY `idx_amount_type`(`amount_type`),
  KEY `idx_opid`(`op_user_id`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `club_members` (
  `member_id`  INT      NOT NULL AUTO_INCREMENT,
  `club_id`    INT      NOT NULL,
  `user_id`    INT      NOT NULL,
  `role`       INT      NOT NULL,
  `status`     INT      NOT NULL, #'成员状态 1正常 2禁用 3主动退出 4被动退出 5黑名单',
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,

  PRIMARY KEY (`member_id`),
  #UNIQUE KEY `members_unique` (`club_id`, `user_id`),
  KEY `idx_created`(`created_at`),
  KEY `idx_club`(`club_id`),
  KEY `idx_user`(`user_id`),
  KEY `idx_status`(`status`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `black_lists` (
  `black_id`   INT      NOT NULL AUTO_INCREMENT,
  `type`       INT      NOT NULL, #'黑名单类型 1俱乐部'
  `origin_id`  INT      NOT NULL, #'发起对象id'
  `target_id`  INT      NOT NULL, #'被屏蔽对象id'
  `status`     INT      NOT NULL, #'1生效 2失效'
  `op_id`      INT      NOT NULL, #'操作人id'
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,

  PRIMARY KEY (`black_id`),
  #UNIQUE KEY `black_unique` (`type`, `origin_id`, `target_id`),
  KEY `idx_created`(`created_at`),
  KEY `idx_type`(`type`),
  KEY `idx_origin`(`origin_id`),
  KEY `idx_target`(`target_id`),
  KEY `idx_status`(`status`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `examines` (
  `examine_id`   INT      NOT NULL AUTO_INCREMENT,
  `type`         INT      NOT NULL, #'审核申请类型 1俱乐部'
  `applicant_id` INT      NOT NULL, #'申请人id'
  `auditor_id`   INT      NOT NULL, #'审核人id'
  `status`       INT      NOT NULL, #'-1 未处理 1同意 2拒绝'
  `op_id`        INT      NOT NULL, #'操作人id'
  `created_at`   DATETIME NOT NULL,
  `updated_at`   DATETIME NOT NULL,

  PRIMARY KEY (`examine_id`),
  # UNIQUE KEY `black_unique` (`type`, `applicant_id`, `auditor_id`),
  KEY `idx_created`(`created_at`),
  KEY `idx_type`(`type`),
  KEY `idx_applicant`(`applicant_id`),
  KEY `idx_auditor`(`auditor_id`),
  KEY `idx_status`(`status`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;



INSERT INTO users VALUES
  (0, "admin@xnhd", "67bad3e758b4d324381586f209fee08bca0701396a606f12029425f31cd29ce8", "YWRtaW5AeG5oZA==", "", "", "",
      0, "", "", 0, now(), now(), now(), 2097151, 1, "", 0, 0, "", "", "", "", "", "", "", "");
INSERT INTO balances VALUES (0, 100000, 0, 0, 1, 100000000, 100000000, now(), now());
INSERT INTO balances VALUES (0, 100000, 0, 0, 2, 100000000, 100000000, now(), now());
INSERT INTO configs VALUES (0, "", "", "", 100001, "1", 1, "全局默认充值开关", now(), now());
INSERT INTO configs VALUES (0, "", "", "", 110001, "100", 1, "全局默认消费开关", now(), now());
