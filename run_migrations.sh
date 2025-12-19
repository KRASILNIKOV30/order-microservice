#!/bin/bash

# Скрипт для запуска миграций во всех микросервисах
# Выполните этот скрипт после создания базы данных

set -e

echo "🚀 Запуск миграций для всех микросервисов..."

# Функция для запуска миграций в сервисе
run_migrations() {
    local service_name=$1
    local service_path=$2
    local db_name=$3
    
    echo ""
    echo "🔧 Миграции для $service_name..."
    
    # Переходим в директорию сервиса
    cd "$service_path" || exit 1
    
    # Проверяем наличие .env файла
    if [ ! -f ".env" ]; then
        echo "⚠️  Файл .env не найден в $service_path"
        echo "   Скопируйте .env.example в .env и настройте подключение к БД"
        return 1
    fi
    
    # Запускаем миграции
    echo "   Выполняю миграции для базы: $db_name"
    ./bin/$service_name migrate
    
    if [ $? -eq 0 ]; then
        echo "   ✅ Миграции для $service_name выполнены успешно"
    else
        echo "   ❌ Ошибка при выполнении миграций для $service_name"
        exit 1
    fi
    
    cd - > /dev/null
}

# Запускаем миграции для каждого сервиса
services=(
    "order:order:order_microservice"
    "user:user:user_microservice"
    "payment:payment:payment_microservice"
    "product:product:product_microservice"
    "notification:notification:notification_microservice"
)

for service_info in "${services[@]}"; do
    IFS=':' read -r service_name service_path db_name <<< "$service_info"
    run_migrations "$service_name" "$service_path" "$db_name"
done

echo ""
echo "🎉 Все миграции выполнены успешно!"
echo ""
echo "📊 Список таблиц в базах данных:"
echo "   • order_microservice: orders, order_items"
echo "   • user_microservice: users"
echo "   • payment_microservice: payments, wallets"
echo "   • product_microservice: products"
echo "   • notification_microservice: notifications, recipients"