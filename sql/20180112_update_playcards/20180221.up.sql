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

INSERT INTO mail_infos VALUES (1001,"充值成功通知", "充值成功", "感谢您的充值，您的【%s】【%s】已到帐，祝您游戏愉快！", 1,110, "", now(), now());
INSERT INTO mail_infos VALUES (1101,"加入俱乐部通知", "加入俱乐部成功", " “【%s】”俱乐部已对您敞开大门，祝您游戏愉快！", 1,110, "", now(), now());
INSERT INTO mail_infos VALUES (1102,"退出俱乐部通知", "退出俱乐部通知", "“【%s】”俱乐部已将您移出，如有疑问请联系俱乐部老板。", 1,110, "", now(), now());
INSERT INTO mail_infos VALUES (1201,"游戏投票解散通知", "游戏解散通知", "您的对局【%s】已经投票解散，对局详情可点击大厅战绩按钮进行查看。", 1,110, "", now(), now());
INSERT INTO mail_infos VALUES (1202,"游戏超时解散通知", "游戏解散通知", "您的对局【%s】因游戏时长超过24小时被系统解散，对局详情可点击大厅战绩按钮进行查看。", 2,110, "", now(), now());
INSERT INTO mail_infos VALUES (1301,"邀请好友绑定获奖通知", "邀请用户奖励通知", "【%s】已成功绑定您作为邀请人，符合邀请奖励条件，请领取您的奖励！", 3,110, '[{"MainType":100,"SubType":2,"ItemID":0,"Count":100}]', now(), now());
INSERT INTO mail_infos VALUES (1302,"好友绑定获奖通知", "好友绑定获奖通知", "已成功绑定【%s】作为您的邀请人，符合邀请奖励条件，请领取您的奖励！", 3,110, '[{"MainType":100,"SubType":2,"ItemID":0,"Count":100}]', now(), now());
INSERT INTO mail_infos VALUES (1303,"分享朋友圈奖励通知", "分享朋友圈奖励通知", "分享成功，请领取您的奖励！", 3,110, '[{"MainType":100,"SubType":2,"ItemID":0,"Count":50}]', now(), now());
INSERT INTO mail_infos VALUES (1304,"邀请好友不符合奖励条件通知", "邀请好友不符合奖励条件通知", "邀请成功，但您的注册时间已超过3天，不能获取奖励！", 2,110, "", now(), now());


INSERT INTO mail_infos VALUES (1103,"俱乐部拒绝加入通知", "俱乐部拒绝加入通知", "“【%s】”俱乐部拒绝了您的加入申请，如有疑问请联系俱乐部老板。", 1,110, "", now(), now());
INSERT INTO mail_infos VALUES (1104,"俱乐部解散通知", "俱乐部解散通知", "“【%s】”俱乐部已解散，如有疑问请联系俱乐部老板。", 1,110, "", now(), now());
