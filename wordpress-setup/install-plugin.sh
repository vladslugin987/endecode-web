#!/bin/bash

echo "🔌 Установка ENDECode Integration плагина..."

# Создаем директорию плагина в WordPress
docker exec endecode-wordpress mkdir -p /var/www/html/wp-content/plugins/endecode-integration

# Копируем плагин
docker cp ./plugins/endecode-integration/endecode-integration.php endecode-wordpress:/var/www/html/wp-content/plugins/endecode-integration/

# Активируем плагин
docker exec endecode-wordpress wp --allow-root plugin activate endecode-integration

echo "✅ Плагин установлен и активирован!"

# Настройка базовых опций плагина
docker exec endecode-wordpress wp --allow-root option update endecode_api_url "http://host.docker.internal:8090"
docker exec endecode-wordpress wp --allow-root option update endecode_default_expiry_days "7"
docker exec endecode-wordpress wp --allow-root option update endecode_max_downloads "3"

echo "⚙️ Базовые настройки применены!"
echo "🌐 Настройки плагина: http://localhost:8082/wp-admin/options-general.php?page=endecode-settings"
