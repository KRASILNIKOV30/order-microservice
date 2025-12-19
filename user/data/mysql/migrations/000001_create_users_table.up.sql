CREATE TABLE IF NOT EXISTS users
(
    `id`         CHAR(36) NOT NULL,
    `login`      VARCHAR(50) NOT NULL,
    `email`      VARCHAR(100) NOT NULL,
    `tg`         VARCHAR(50) NULL DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_login` (`login`),
    UNIQUE INDEX `idx_email` (`email`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE = InnoDB
  CHARACTER SET = utf8mb4
  COLLATE utf8mb4_unicode_ci;