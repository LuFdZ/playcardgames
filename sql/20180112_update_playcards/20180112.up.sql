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
  `game_result_str`  VARCHAR(3000) NOT NULL DEFAULT '',
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
  `mail_type`        INT           NOT NULL DEFAULT '110',
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
  `log_id`           INT           NOT NULL AUTO_INCREMENT,
  `mail_id`          INT           NOT NULL DEFAULT '0',
  `send_id`          INT           NOT NULL DEFAULT '0',
  `mail_type`        INT           NOT NULL DEFAULT '0',
  `status`           INT           NOT NULL DEFAULT '110',
  `mail_str`         VARCHAR(1500)  NOT NULL DEFAULT '',
  `send_count`       INT           NOT NULL DEFAULT '0',
  `total_count`      INT           NOT NULL DEFAULT '0',
  `created_at`       DATETIME      NOT NULL,
  `updated_at`       DATETIME      NOT NULL,
  PRIMARY KEY (`log_id`),
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
  `log_id`           INT           NOT NULL AUTO_INCREMENT,
  `send_log_id`      INT           NOT NULL DEFAULT '0',
  `mail_id`          INT           NOT NULL DEFAULT '0',
  `user_id`          INT           NOT NULL DEFAULT '0',
  `send_id`          INT           NOT NULL DEFAULT '0',
  `mail_type`        INT           NOT NULL DEFAULT '0',
  `have_item`        INT           NOT NULL DEFAULT '0',
  `status`           INT           NOT NULL DEFAULT '110',
  `created_at`       DATETIME      NOT NULL,
  `updated_at`       DATETIME      NOT NULL,
   PRIMARY KEY (`log_id`),
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


INSERT INTO mail_infos VALUES (1001,"充值成功通知", "充值成功", "感谢您的充值，您的【%s】【%s】已到帐，祝您游戏愉快！", 3,110, "", now(), now());
INSERT INTO mail_infos VALUES (1101,"加入俱乐部通知", "加入俱乐部成功", " “【%s】”俱乐部已对您敞开大门，祝您游戏愉快！", 3,110, "", now(), now());
INSERT INTO mail_infos VALUES (1102,"退出俱乐部通知", "退出俱乐部通知", "“【%s】”俱乐部已将您移出，如有疑问请联系俱乐部会长。", 3,110, "", now(), now());
INSERT INTO mail_infos VALUES (1201,"游戏投票解散通知", "游戏解散通知", "您的对局【%s】已经投票解散，对局详情可点击大厅战绩按钮进行查看。", 3,110, "", now(), now());
INSERT INTO mail_infos VALUES (1202,"游戏超时解散通知", "游戏解散通知", "您的对局【%s】因游戏时长超过24小时被系统解散，对局详情可点击大厅战绩按钮进行查看。", 3,110, "", now(), now());
INSERT INTO mail_infos VALUES (1301,"邀请好友绑定获奖通知", "邀请用户奖励通知", "【%s】已成功绑定您作为邀请人，符合邀请奖励条件，请领取您的奖励！", 3,110, "", now(), now());
INSERT INTO mail_infos VALUES (1302,"分享朋友圈奖励通知", "邀请用户奖励通知", "【用户昵称】已成功绑定您作为邀请人，符合邀请奖励条件，请领取您的奖励！", 3,110, "", now(), now());
alter table users add type int default 0 not null after status;
ALTER  TABLE  `users`  ADD  INDEX idx_type (`type`);
alter table rooms add shuffle int default 0 not null after giveup;