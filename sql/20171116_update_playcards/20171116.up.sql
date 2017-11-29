CREATE TABLE `clubs` (
  `club_id`        INT          NOT NULL AUTO_INCREMENT,
  `club_name`      VARCHAR(128) NOT NULL,
  `status`         VARCHAR(128) NOT NULL
  COMMENT '俱乐部状态 -1审核中 1正常 2冻结',
  `creater_user`   INT          NOT NULL,
  `creater_proxy`  INT          NOT NULL DEFAULT '0',
  `club_remark`    VARCHAR(200),
  `icon`           VARCHAR(128)          DEFAULT NULL,
  `niu_ratio`      INT          NOT NULL DEFAULT '0'
  COMMENT '牛牛抽成',
  `thirteen_ratio` INT          NOT NULL DEFAULT '0'
  COMMENT '十三张抽成',
  `created_at`     DATETIME     NOT NULL,
  `updated_at`     DATETIME     NOT NULL,
  PRIMARY KEY (`club_id`),
  UNIQUE KEY `name_unique` (`club_name`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `club_members` (
  `member_id`  INT      NOT NULL AUTO_INCREMENT,
  `club_id`    INT      NOT NULL,
  `user_id`    INT      NOT NULL,
  `status`     INT      NOT NULL
  COMMENT '成员状态 1正常 2禁用',
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,

  PRIMARY KEY (`member_id`),
  UNIQUE KEY `members_unique` (`club_id`, `user_id`),
  KEY `idx_created`(`created_at`),
  KEY `idx_club`(`club_id`),
  KEY `idx_user`(`user_id`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `black_lists` (
  `black_id`   INT      NOT NULL AUTO_INCREMENT,
  `type`       INT      NOT NULL
  COMMENT '黑名单类型 1俱乐部',
  `origin_id`  INT      NOT NULL
  COMMENT '发起对象id',
  `target_id`  INT      NOT NULL
  COMMENT '被屏蔽对象id',
  `op_id`      INT      NOT NULL
  COMMENT '操作人id',
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,

  PRIMARY KEY (`black_id`),
  UNIQUE KEY `black_unique` (`type`, `origin_id`, `target_id`),
  KEY `idx_created`(`created_at`),
  KEY `idx_created`(`type`),
  KEY `idx_created`(`origin_id`),
  KEY `idx_created`(`target_id`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

CREATE TABLE `examines` (
  `examine_id`   INT      NOT NULL AUTO_INCREMENT,
  `type`         INT      NOT NULL
  COMMENT '审核申请类型 1俱乐部',
  `applicant_id` INT      NOT NULL
  COMMENT '申请人id',
  `auditor_id`   INT      NOT NULL
  COMMENT '审核人id',
  `status`       INT      NOT NULL
  COMMENT '-1 未处理 1同意 2拒绝',
  `created_at`   DATETIME NOT NULL,
  `updated_at`   DATETIME NOT NULL,

  PRIMARY KEY (`examine_id`),
  UNIQUE KEY `black_unique` (`type`, `applicant_id`, `auditor_id`),
  KEY `idx_created`(`created_at`),
  KEY `idx_created`(`type`),
  KEY `idx_created`(`applicant_id`),
  KEY `idx_created`(`auditor_id`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;


ALTER TABLE `users`
  ADD club_id INT NOT NULL DEFAULT '0';
ALTER TABLE `balances`
  ADD type INT NOT NULL DEFAULT '0' COMMENT '1玩家 2俱乐部';

ALTER TABLE `balances` DROP gold_profit;
ALTER TABLE `balances` DROP diamond_profit;
ALTER TABLE `balances` DROP cumulative_recharge;
ALTER TABLE `balances` DROP cumulative_consumption;
ALTER TABLE `balances` DROP cumulative_gift;
ALTER TABLE `journals` CHANGE `foreign` `foreign` BIGINT;