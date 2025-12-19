CREATE TABLE IF NOT EXISTS notifications
(
    `id`             CHAR(36) NOT NULL,
    `recipient_id`   CHAR(36) NOT NULL,
    `channel`        INT NOT NULL,
    `message`        TEXT NOT NULL,
    `status`         INT NOT NULL,
    `failure_reason` VARCHAR(255) NULL DEFAULT NULL,
    `created_at`     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `sent_at`        DATETIME NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_recipient_id` (`recipient_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE = InnoDB
  CHARACTER SET = utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS recipients
(
    `user_id`  CHAR(36) NOT NULL,
    `email`    VARCHAR(100) NULL DEFAULT NULL,
    `tg`       VARCHAR(50) NULL DEFAULT NULL,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`user_id`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE = InnoDB
  CHARACTER SET = utf8mb4
  COLLATE utf8mb4_unicode_ci;