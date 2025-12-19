CREATE TABLE IF NOT EXISTS products
(
    `id`         CHAR(36) NOT NULL,
    `name`       VARCHAR(100) NOT NULL,
    `price`      DECIMAL(10,2) NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_name` (`name`),
    INDEX `idx_price` (`price`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE = InnoDB
  CHARACTER SET = utf8mb4
  COLLATE utf8mb4_unicode_ci;