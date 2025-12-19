#!/bin/bash

# Скрипт для локального запуска всех микросервисов
# Используется для тестирования и разработки

set -e

echo "🚀 Запуск всех микросервисов локально..."

# Проверяем, что бинарники собраны
echo "🔍 Проверка собранных бинарников..."

services=(
    "order:order"
    "user:user"
    "payment:payment"
    "product:product"
    "notification:notification"
)

missing_binaries=()

for service_info in "${services[@]}"; do
    IFS=':' read -r service_name service_path <<< "$service_info"
    if [ ! -f "$service_path/bin/$service_name" ]; then
        missing_binaries+=("$service_name")
        echo "❌ $service_path/bin/$service_name не найден"
    else
        echo "✅ $service_path/bin/$service_name найден"
    fi
done

if [ ${#missing_binaries[@]} -gt 0 ]; then
    echo ""
    echo "⚠️  Обнаружены отсутствующие бинарники:"
    for missing in "${missing_binaries[@]}"; do
        echo "   • $missing"
    done
    echo ""
    echo "🔧 Запустите сборку: ./build_all.sh"
    exit 1
fi

echo ""
echo "🎯 Запуск сервисов..."

# Функция для запуска сервиса в фоне
start_service() {
    local service_name=$1
    local service_path=$2
    local port=$3
    
    echo "📦 Запуск $service_name на порту $port..."
    
    cd "$service_path"
    
    # Запускаем сервис в фоне и сохраняем PID
    ./bin/$service_name service &
    local pid=$!
    
    echo "   PID: $pid, Порт: $port"
    
    # Сохраняем PID для последующего завершения
    echo "$pid" > "/tmp/$service_name.pid"
    
    cd - > /dev/null
    
    # Небольшая задержка между запусками
    sleep 2
}

# Останавливаем все запущенные сервисы перед стартом
echo "🛑 Остановка предыдущих запусков..."
for service_info in "${services[@]}"; do
    IFS=':' read -r service_name service_path <<< "$service_info"
    if [ -f "/tmp/$service_name.pid" ]; then
        local pid=$(cat "/tmp/$service_name.pid")
        if kill -0 "$pid" 2>/dev/null; then
            echo "   Останавливаю $service_name (PID: $pid)..."
            kill "$pid"
            rm -f "/tmp/$service_name.pid"
        fi
    fi
done

sleep 1

# Запускаем сервисы на разных портах
start_service "order" "order" "8081"
start_service "user" "user" "8082" 
start_service "payment" "payment" "8083"
start_service "product" "product" "8084"
start_service "notification" "notification" "8085"

echo ""
echo "🎉 Все сервисы запущены!"
echo ""
echo "📊 Статус сервисов:"
echo "   • Order Service:      http://localhost:8081"
echo "   • User Service:       http://localhost:8082"
echo "   • Payment Service:   http://localhost:8083"
echo "   • Product Service:    http://localhost:8084"
echo "   • Notification Service: http://localhost:8085"
echo ""
echo "🔍 Для проверки статуса процессов:"
echo "   ps aux | grep $service_name"
echo ""
echo "⏹️  Для остановки всех сервисов:"
echo "   ./stop_all.sh"

# Создаем скрипт для остановки
cat > /home/bogdan.krasilnikov/projects/order-microservice/stop_all.sh << 'EOF'
#!/bin/bash

echo "🛑 Остановка всех микросервисов..."

services=("order" "user" "payment" "product" "notification")

for service_name in "${services[@]}"; do
    if [ -f "/tmp/$service_name.pid" ]; then
        local pid=$(cat "/tmp/$service_name.pid")
        if kill -0 "$pid" 2>/dev/null; then
            echo "   Останавливаю $service_name (PID: $pid)..."
            kill "$pid"
            rm -f "/tmp/$service_name.pid"
        else
            echo "   $service_name уже остановлен"
            rm -f "/tmp/$service_name.pid"
        fi
    else
        echo "   $service_name не запущен"
    fi
done

echo "✅ Все сервисы остановлены"
EOF

chmod +x /home/bogdan.krasilnikov/projects/order-microservice/stop_all.sh

echo "💡 Используйте ./stop_all.sh для остановки всех сервисов"