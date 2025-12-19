CREATE TABLE IF NOT EXISTS payments
(
    `id`             CHAR(36) NOT NULL,
    `order_id`       CHAR(36) NOT NULL,
    `user_id`        CHAR(36) NOT NULL,
    `amount`         DECIMAL(10,2) NOT NULL,
    `status`         INT NOT NULL,
    `failure_reason` VARCHAR(255) NULL DEFAULT NULL,
    `created_at`     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_order_id` (`order_id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE = InnoDB
  CHARACTER SET = utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS wallets
(
    `id`         CHAR(36) NOT NULL,
    `user_id`    CHAR(36) NOT NULL,
    `balance`    DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_user_id` (`user_id`),
    INDEX `idx_balance` (`balance`)
) ENGINE = InnoDB
  CHARACTER SET = utf8mb4
  COLLATE utf8mb4_unicode_ci;