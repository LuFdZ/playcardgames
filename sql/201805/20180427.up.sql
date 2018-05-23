CREATE TABLE `runcards` (
  `game_id`          INT           NOT NULL AUTO_INCREMENT,
  `room_id`          INT           NOT NULL DEFAULT '0',
  `index`            INT           NOT NULL DEFAULT '0',
  `game_result_str`  VARCHAR(5000) NOT NULL DEFAULT '',
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