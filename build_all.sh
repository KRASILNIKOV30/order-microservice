#!/bin/bash

# Скрипт для запуска brewkit build во всех микросервисах
# Выполните этот скрипт для сборки всех сервисов

set -e

echo "🔨 Запуск brewkit build для всех микросервисов..."

# Функция для запуска brewkit build в сервисе
build_service() {
    local service_name=$1
    local service_path=$2
    
    echo ""
    echo "🔧 Сборка $service_name..."
    
    # Переходим в директорию сервиса
    cd "$service_path" || exit 1
    
    # Проверяем наличие brewkit.jsonnet
    if [ ! -f "brewkit.jsonnet" ]; then
        echo "⚠️  Файл brewkit.jsonnet не найден в $service_path"
        return 1
    fi
    
    # Запускаем brewkit build
    echo "   Выполняю brewkit build для $service_name"
    brewkit build
    
    if [ $? -eq 0 ]; then
        echo "   ✅ Сборка $service_name завершена успешно"
    else
        echo "   ❌ Ошибка при сборке $service_name"
        return 1
    fi
    
    cd - > /dev/null
}

# Проверяем доступность brewkit
if ! command -v brewkit &> /dev/null; then
    echo "❌ Ошибка: brewkit не найден в PATH"
    echo "   Убедитесь, что brewkit установлен и добавлен в PATH"
    exit 1
fi

echo "✅ brewkit найден: $(which brewkit)"

# Запускаем сборку для каждого сервиса
services=(
    "order:order"
    "user:user"
    "payment:payment"
    "product:product"
    "notification:notification"
)

failed_services=()

for service_info in "${services[@]}"; do
    IFS=':' read -r service_name service_path <<< "$service_info"
    if ! build_service "$service_name" "$service_path"; then
        failed_services+=("$service_name")
    fi
done

echo ""
echo "📊 Результат сборки:"

if [ ${#failed_services[@]} -eq 0 ]; then
    echo "🎉 Все сервисы собраны успешно!"
    echo ""
    echo "📦 Собранные бинарники:"
    for service_info in "${services[@]}"; do
        IFS=':' read -r service_name service_path <<< "$service_info"
        if [ -f "$service_path/bin/$service_name" ]; then
            echo "   ✅ $service_path/bin/$service_name"
        fi
    done
else
    echo "❌ Ошибки при сборке сервисов:"
    for failed_service in "${failed_services[@]}"; do
        echo "   • $failed_service"
    done
    exit 1
fi