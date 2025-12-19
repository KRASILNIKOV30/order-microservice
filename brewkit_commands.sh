#!/bin/bash

# Скрипт для запуска различных команд brewkit во всех микросервисах
# Использование: ./brewkit_commands.sh [command]
# Примеры: ./brewkit_commands.sh test
#          ./brewkit_commands.sh check

COMMAND=${1:-build}

if [ "$COMMAND" != "build" ] && [ "$COMMAND" != "test" ] && [ "$COMMAND" != "check" ] && [ "$COMMAND" != "generate" ]; then
    echo "❌ Неизвестная команда: $COMMAND"
    echo "Доступные команды: build, test, check, generate"
    exit 1
fi

echo "🚀 Запуск brewkit $COMMAND для всех микросервисов..."

# Функция для запуска команды brewkit в сервисе
run_brewkit_command() {
    local service_name=$1
    local service_path=$2
    local command=$3
    
    echo ""
    echo "🔧 $command для $service_name..."
    
    # Переходим в директорию сервиса
    cd "$service_path" || exit 1
    
    # Проверяем наличие brewkit.jsonnet
    if [ ! -f "brewkit.jsonnet" ]; then
        echo "⚠️  Файл brewkit.jsonnet не найден в $service_path"
        return 1
    fi
    
    # Запускаем команду brewkit
    echo "   Выполняю brewkit $command для $service_name"
    brewkit $command
    
    if [ $? -eq 0 ]; then
        echo "   ✅ $command для $service_name завершено успешно"
    else
        echo "   ❌ Ошибка при выполнении $command для $service_name"
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

# Запускаем команду для каждого сервиса
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
    if ! run_brewkit_command "$service_name" "$service_path" "$COMMAND"; then
        failed_services+=("$service_name")
    fi
done

echo ""
echo "📊 Результат выполнения $COMMAND:"

if [ ${#failed_services[@]} -eq 0 ]; then
    echo "🎉 Команда $COMMAND выполнена успешно для всех сервисов!"
else
    echo "❌ Ошибки при выполнении $COMMAND для сервисов:"
    for failed_service in "${failed_services[@]}"; do
        echo "   • $failed_service"
    done
    exit 1
fi