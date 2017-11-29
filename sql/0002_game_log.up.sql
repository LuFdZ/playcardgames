# CREATE TABLE `users_active_log` (
#   `log_id`          BIGINT       NOT NULL,
#   `user_id`         INT          NOT NULL,
#   `created_at`      DATETIME     NOT NULL,
#   `updated_at`      DATETIME     NOT NULL,
#   `last_login_at`   DATETIME              DEFAULT NULL,
#   `rights`          INT          NOT NULL DEFAULT '0',
#   `sex`             INT          NOT NULL DEFAULT '0',
#   `icon`            VARCHAR(128)          DEFAULT NULL,
#   `invite_user_id`  INT          NOT NULL DEFAULT '0',
#   `mobile_uu_id`    VARCHAR(30)           DEFAULT NULL,
#   `mobile_model`    VARCHAR(20)           DEFAULT NULL,
#   `mobile_net_work` VARCHAR(20)           DEFAULT NULL,
#   `mobile_os`       VARCHAR(20)           DEFAULT NULL,
#   `last_login_ip`   VARCHAR(20)           DEFAULT NULL,
#   `reg_ip`          VARCHAR(20)           DEFAULT NULL,
#   `open_id`         VARCHAR(30)           DEFAULT NULL,
#   `union_id`        VARCHAR(30)           DEFAULT NULL,
#   PRIMARY KEY (`log_id`),
#   UNIQUE KEY `openid_unique` (`open_id`)
# )
#   ENGINE = InnoDB
#   AUTO_INCREMENT = 100000
#   DEFAULT CHARSET = utf8;