# Настройки ENDECode API плагина

## Настройки в WordPress Admin → Settings → ENDECode:

### API Configuration
- **ENDECode API URL**: `http://host.docker.internal:8090`
- **API Key**: оставьте пустым (пока не требуется)
- **Default Link Expiry**: `7` дней
- **Max Downloads per Link**: `3`

Нажмите **"Save Changes"**, затем **"Test API Connection"**

✅ Должно появиться: "✓ Connection successful!"

## Проверка подключения из терминала:

```bash
# Тест API из WordPress контейнера
docker exec endecode-wordpress curl -s http://host.docker.internal:8090/health
```

## Если тест не проходит:

### Вариант 1: Использовать localhost
- **ENDECode API URL**: `http://localhost:8090`

### Вариант 2: Использовать IP адрес
```bash
# Найти IP адрес ENDECode контейнера
docker inspect endecode-compose-main-photo-processing-web-1 | grep IPAddress
```

Затем использовать: `http://[IP_ADDRESS]:8090`
