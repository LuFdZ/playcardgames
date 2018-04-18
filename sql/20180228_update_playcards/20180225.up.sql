INSERT INTO mail_infos VALUES (1103,"俱乐部拒绝加入通知", "俱乐部拒绝加入通知", "“【%s】”俱乐部拒绝了您的加入申请，如有疑问请联系俱乐部老板。", 1,110, "", now(), now());
INSERT INTO mail_infos VALUES (1104,"俱乐部解散通知", "俱乐部解散通知", "“【%s】”俱乐部已解散，如有疑问请联系俱乐部老板。", 1,110, "", now(), now());

CREATE TABLE `vip_room_settings` (
  `id`               INT           NOT NULL AUTO_INCREMENT,
  `name`             VARCHAR(50)   NOT NULL DEFAULT '',
  `club_id`          INT           NOT NULL DEFAULT '0',
  `user_id`          INT           NOT NULL DEFAULT '0',
  `room_type`        INT           NOT NULL DEFAULT '0',
  `game_type`        INT           NOT NULL DEFAULT '0',
  `max_number`       INT           NOT NULL DEFAULT '0',
  `round_number`     INT           NOT NULL DEFAULT '0',
  `sub_room_type`    INT           NOT NULL DEFAULT '0',
  `game_param`       VARCHAR(255)  NOT NULL DEFAULT '',
  `setting_param`    VARCHAR(255)  NOT NULL DEFAULT '',
  `status`              INT           NOT NULL DEFAULT '0',
  `room_advance_options` VARCHAR(200)  NOT NULL DEFAULT '',
  `created_at`       DATETIME      NOT NULL,
  `updated_at`       DATETIME      NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_room` (`club_id`),
  KEY `idx_user` (`user_id`),
  KEY `club_id` (`club_id`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;

--
alter table clubs add club_coin BIGINT default 0 not null after diamond;
alter table clubs add notice VARCHAR(200) NOT NULL DEFAULT '' after club_param;
alter table clubs add setting_param VARCHAR(200) NOT NULL DEFAULT '' after notice;

alter table club_members add club_coin BIGINT default 0 not null after status;

alter table club_journals add status INT default 110 not null after `foreign`;

alter table rooms add sub_room_type INT default 0 not null after cost_type;
alter table rooms add setting_param VARCHAR(200) NOT NULL DEFAULT '' after sub_room_type;
alter table rooms add start_max_number INT default 0 not null after sub_room_type;
alter table rooms add vip_room_setting_id INT default 0 not null after sub_room_type;

-- alter table rooms add join_type INT default 0 not null after sub_room_type;
-- alter table vip_room_settings add join_type INT default 0 not null after status;

alter table rooms add room_param VARCHAR(255)  NOT NULL DEFAULT '' after sub_room_type;
alter table vip_room_settings add room_param VARCHAR(255)  NOT NULL DEFAULT '' after status;

alter table users add register_channel VARCHAR(200) NOT NULL DEFAULT '' after last_login_ip;
alter table users add proxy_id INT default 0 not null after last_login_ip;

alter table clubs drop key name_unique;

alter table thirteens modify column user_cards VARCHAR(1200);
alter table thirteens modify column user_submit_cards VARCHAR(1200);
alter table thirteens modify column game_results VARCHAR(3000);




-- rtype int32, srtype int32, gtype int32, maxNum int32, roundNum int32, gParam string, setting string, user *mduser.User, pwd string