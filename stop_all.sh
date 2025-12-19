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
