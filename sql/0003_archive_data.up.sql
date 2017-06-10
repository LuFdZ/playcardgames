CREATE TABLE `client_error_logs` (
  `id`             BIGINT        NOT NULL AUTO_INCREMENT,
  `user_id`        INT           NOT NULL DEFAULT 0,
  `client_address` VARCHAR(64)   NOT NULL DEFAULT '',
  `message`        VARCHAR(2048) NOT NULL DEFAULT '',
  `condition`      VARCHAR(1024) NOT NULL DEFAULT '',
  `stack_trace`    VARCHAR(2048) NOT NULL DEFAULT '',
  `system_info`    JSON          NOT NULL,
  `created_at`     DATETIME      NOT NULL,
  PRIMARY KEY (`id`)
)
  ENGINE = ARCHIVE
  AUTO_INCREMENT = 100000000
  DEFAULT CHARSET = utf8;

CREATE TABLE `client_login_logs` (
  `id`             BIGINT      NOT NULL AUTO_INCREMENT,
  `user_id`        INT         NOT NULL DEFAULT 0,
  `client_address` VARCHAR(64) NOT NULL DEFAULT '',
  `system_info`    JSON        NOT NULL,
  `created_at`     DATETIME    NOT NULL,
  PRIMARY KEY (`id`)
)
  ENGINE = ARCHIVE
  AUTO_INCREMENT = 200000000
  DEFAULT CHARSET = utf8;

CREATE TABLE `client_report_logs` (
  `id`             BIGINT        NOT NULL AUTO_INCREMENT,
  `user_id`        INT           NOT NULL DEFAULT 0,
  `client_address` VARCHAR(64)   NOT NULL DEFAULT '',
  `message`        VARCHAR(1024) NOT NULL DEFAULT '',
  `system_info`    JSON          NOT NULL,
  `created_at`     DATETIME      NOT NULL,
  PRIMARY KEY (`id`)
)
  ENGINE = ARCHIVE
  AUTO_INCREMENT = 300000000
  DEFAULT CHARSET = utf8;
