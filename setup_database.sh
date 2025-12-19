#!/bin/bash

echo "🔧 Создание базы данных MySQL..."

mysql -u root -p << EOF
CREATE DATABASE IF NOT EXISTS order_microservice CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS user_microservice CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS payment_microservice CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS product_microservice CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS notification_microservice CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Создаем пользователя для всех баз (опционально)
-- CREATE USER IF NOT EXISTS 'microservice_user'@'localhost' IDENTIFIED BY 'secure_password';
-- GRANT ALL PRIVILEGES ON order_microservice.* TO 'microservice_user'@'localhost';
-- GRANT ALL PRIVILEGES ON user_microservice.* TO 'microservice_user'@'localhost';
-- GRANT ALL PRIVILEGES ON payment_microservice.* TO 'microservice_user'@'localhost';
-- GRANT ALL PRIVILEGES ON product_microservice.* TO 'microservice_user'@'localhost';
-- GRANT ALL PRIVILEGES ON notification_microservice.* TO 'microservice_user'@'localhost';
-- FLUSH PRIVILEGES;

SHOW DATABASES;
EOF

echo "✅ База данных создана!"