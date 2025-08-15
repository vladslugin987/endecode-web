# 📧 Настройка уведомлений ENDECode

## 🎯 Обзор

ENDECode поддерживает автоматические уведомления для администраторов и клиентов:

- **📧 Email уведомления** (SMTP)
- **📱 Telegram уведомления** для админов
- **🔄 Статус страницы** для клиентов

## 📧 Настройка Email уведомлений

### 1. Gmail (рекомендуется)

1. **Включите двухфакторную аутентификацию** в Google аккаунте
2. **Создайте пароль приложения**:
   - Перейдите в Google Account → Security → App passwords
   - Выберите "Mail" и устройство
   - Скопируйте сгенерированный пароль

3. **Добавьте переменные в .env**:
```bash
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=generated-app-password
ADMIN_EMAIL=admin@photocompany.com
```

### 2. Другие провайдеры

```bash
# Outlook/Hotmail
SMTP_HOST=smtp-mail.outlook.com
SMTP_PORT=587

# Yahoo
SMTP_HOST=smtp.mail.yahoo.com
SMTP_PORT=587

# Custom SMTP
SMTP_HOST=mail.your-domain.com
SMTP_PORT=587
```

## 📱 Настройка Telegram уведомлений

### 1. Создайте Telegram бота

1. **Найдите @BotFather** в Telegram
2. **Отправьте** `/newbot`
3. **Введите название** бота (например: "PhotoCompany Notifications")
4. **Введите username** бота (например: "photocompany_bot")
5. **Сохраните токен** (формат: `1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZ123456789`)

### 2. Получите Chat ID

**Способ 1 - Личные сообщения:**
1. Напишите боту `/start`
2. Перейдите по ссылке: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
3. Найдите `"chat":{"id":12345}` - это ваш Chat ID

**Способ 2 - Групповой чат:**
1. Добавьте бота в группу
2. Дайте боту права администратора
3. Напишите в группу `/start@your_bot_username`
4. Используйте ссылку выше для получения Chat ID (будет отрицательным)

### 3. Добавьте в .env

```bash
TELEGRAM_BOT_TOKEN=1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZ123456789
TELEGRAM_CHAT_ID=12345  # или -1001234567890 для группы
```

## ⚙️ Полная настройка

### 1. Создайте .env файл

```bash
cd /home/vsdev/Documents/endecode-compose/endecode-compose-main
cp .env.example .env
```

### 2. Заполните настройки

```bash
# Email notifications (SMTP)
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
ADMIN_EMAIL=admin@photocompany.com

# Telegram notifications
TELEGRAM_BOT_TOKEN=1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZ123456789
TELEGRAM_CHAT_ID=-1001234567890

# Optional
LOG_LEVEL=info
```

### 3. Перезапустите сервисы

```bash
docker-compose down
docker-compose up -d
```

## 📋 Типы уведомлений

### Для администратора:

1. **📧 Email**: "New Order #12345 Ready for Review"
2. **📱 Telegram**: 
```
🚨 *New Order Processing Complete*

Order ID: 12345
Job ID: woo_12345_1234567890
Status: Ready for Review
Admin Panel: http://localhost:8090

Please review and approve for delivery.
```

### Для клиента:

1. **📧 Обработка началась**: "Your Order is Being Processed"
2. **📧 Ожидание одобрения**: "Your Order is Pending Approval"  
3. **📧 Готово к скачиванию**: "Your Photos Are Ready for Download"

## 🔍 Тестирование

### 1. Проверка настроек

```bash
# Проверка логов
docker logs endecode-compose-main-photo-processing-web-1

# Поиск ошибок уведомлений
docker logs endecode-compose-main-photo-processing-web-1 | grep -i "notification\|telegram\|smtp"
```

### 2. Тестовый заказ

1. Создайте тестовый товар с включенной обработкой фото
2. Оформите заказ
3. Измените статус на "Completed"
4. Проверьте уведомления

## 🚨 Troubleshooting

### Email не отправляются

```bash
# Проверьте настройки SMTP
docker exec endecode-compose-main-photo-processing-web-1 env | grep SMTP

# Возможные проблемы:
# - Неверный пароль приложения Gmail
# - Блокировка провайдером
# - Неправильные SMTP настройки
```

### Telegram не работает

```bash
# Проверьте токен бота
curl "https://api.telegram.org/bot<YOUR_TOKEN>/getMe"

# Проверьте Chat ID
curl "https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates"

# Возможные проблемы:
# - Неверный токен
# - Неправильный Chat ID
# - Бот не добавлен в группу
```

### Ошибки загрузки файлов

**Проблема**: "POST Content-Length exceeds the limit"

```bash
# Проверьте PHP лимиты
docker exec endecode-wordpress php -i | grep -E "upload_max_filesize|post_max_size|memory_limit"

# Ожидаемые значения:
# upload_max_filesize = 1024M
# post_max_size = 1024M  
# memory_limit = 1024M
```

**Если лимиты неправильные**, отредактируйте `wordpress-setup/uploads.ini`:
```ini
file_uploads = On
memory_limit = 1024M
upload_max_filesize = 1024M
post_max_size = 1024M
max_execution_time = 1200
max_input_time = 1200
max_input_vars = 10000
```

Затем перезапустите WordPress:
```bash
cd wordpress-setup
docker-compose restart wordpress
```

**Проблема**: "✗ Upload failed: undefined"

1. **Проверьте логи WordPress**:
```bash
docker logs endecode-wordpress --tail 20
```

2. **Проверьте логи ENDECode**:
```bash
docker logs endecode-compose-main-photo-processing-web-1 --tail 20
```

3. **Проверьте сетевое подключение**:
```bash
# Из WordPress к ENDECode
docker exec endecode-wordpress curl -I http://endecode-compose-main-photo-processing-web-1:8080/api/upload
```

## 📚 Дополнительные возможности

### Кастомизация сообщений

Отредактируйте файл `notification.go`:
- Измените шаблоны сообщений
- Добавьте дополнительные поля
- Настройте форматирование

### Добавление новых каналов

- **Slack**: Добавьте webhook интеграцию
- **Discord**: Используйте Discord webhook API
- **SMS**: Интегрируйте с Twilio или другими провайдерами 