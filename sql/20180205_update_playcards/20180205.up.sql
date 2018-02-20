CREATE TABLE `player_special_game_records` (
  `log_id`             INT           NOT NULL AUTO_INCREMENT,
  `game_id`            INT           NOT NULL DEFAULT '0',
  `room_id`            INT           NOT NULL DEFAULT '0',
  `game_type`          INT           NOT NULL DEFAULT '0',
  `room_type`          INT           NOT NULL DEFAULT '0',
  `user_id`            INT           NOT NULL DEFAULT '0',
  `password`           VARCHAR(16)   NOT NULL,
  `game_result`        VARCHAR(3200) NOT NULL DEFAULT '',
  `created_at`         DATETIME      NOT NULL,
  `updated_at`         DATETIME      NOT NULL,
  PRIMARY KEY (`log_id`),
  KEY `idx_game` (`game_id`),
  KEY `idx_room` (`room_id`),
  KEY `idx_room_type` (`room_type`),
  KEY `idx_game_type` (`game_type`),
  KEY `idx_password` (`password`),
  KEY `idx_user` (`user_id`),
  KEY `idx_created`(`created_at`)
)
  ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8;


alter table rooms add level int default 0 not null after room_type;