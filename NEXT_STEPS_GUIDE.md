# 🚀 ENDECode + WordPress Integration - Следующие шаги

## 📊 Текущий статус проекта

### ✅ Что уже сделано:
- ✅ **Система подписок** с криптоплатежами (BTC, ETH, USDT)
- ✅ **4 тарифных плана**: Free, Basic ($29.99), Pro ($89.99), Enterprise ($299.99)  
- ✅ **Супер админ панель** для управления подписками пользователей
- ✅ **WordPress + WooCommerce** локально настроены и работают
- ✅ **ENDECode Integration плагин** создан и скопирован в WordPress
- ✅ **Система временных ссылок** с настраиваемым сроком действия
- ✅ **API интеграция** полностью реализована и протестирована
- ✅ **Профессиональный UI** без эмодзи, современный дизайн
- ✅ **Уведомления Telegram и Email** для администраторов и клиентов
- ✅ **Загрузка фото-папок** прямо из WordPress администрации
- ✅ **Невидимые водяные знаки** по умолчанию на всех фотографиях
- ✅ **Кодирование по номеру заказа** (формат "ORDER XXX")
- ✅ **Структура коллекций** как описал шеф (Collection1, Collection2, etc.)
- ✅ **Статус "ожидание подтверждения"** с админ-панелью для одобрения
- ✅ **ZIP файлы с именем заказа** (Collection1-Order123.zip)

### 🔧 Доступные сервисы:
- **ENDECode API**: http://localhost:8090
- **ENDECode Admin**: http://localhost:8090 → кнопка "Admin" (admin/admin123)
- **WordPress**: http://localhost:8082  
- **WordPress Admin**: http://localhost:8082/wp-admin
- **phpMyAdmin**: http://localhost:8081 (root/rootpassword)

## 🎯 Новые возможности

### 📸 Расширенные настройки товаров в WooCommerce:
- **Невидимые водяные знаки**: включены по умолчанию
- **Видимые водяные знаки**: опционально
- **Кодирование по номеру заказа**: формат "ORDER XXX"
- **Файловые пары для свапа**: точечная настройка какие файлы менять местами
- **Водяные знаки на конкретных фото**: указание номеров фотографий (1,3,5,7-10,15)
- **Загрузка папок с компьютера**: прямо в администрации WordPress
- **Структура коллекций**: Collection1, Collection2, Collection3, Collection4

### 🔄 Полный рабочий процесс:
1. ✅ **Клиент покупает товар** на сайте
2. ✅ **Система получает сигнал** о новом заказе
3. ✅ **Фотографии обрабатываются** автоматически
4. ✅ **Администратор получает уведомление** (Telegram + Email)
5. ✅ **Администратор видит в админ-панели** статус заказа
6. ✅ **Администратор может загрузить фото** со своего компьютера
7. ✅ **Администратор видит обработанные фото** в админ-панели
8. ✅ **Администратор нажимает "Одобрить"** для генерации ссылки
9. ✅ **Ссылка отправляется клиенту** на почту автоматически
10. ✅ **Клиент видит статус** "заказ оформлен, ждите одобрения"
11. ✅ **ZIP файл называется** Collection1-Order123.zip

### 📧 Система уведомлений:
- **Email уведомления** через SMTP (Gmail, Outlook, и др.)
- **Telegram уведомления** для админов  
- **Автоматические статусы** для клиентов
- **Интеграция с WordPress** через webhooks

## 📁 Структура коллекций

### Как настроена файловая система:
```
/app/uploads/OriginalFiles/
├── Collection1/
│   ├── vid1.mp4
│   ├── vid2.mp4  
│   └── PhotoFolder/
├── Collection2/
│   ├── vid1.mp4
│   ├── vid2.mp4
│   └── PhotoFolder/
├── Collection3/
│   ├── vid1.mp4
│   ├── vid2.mp4
│   └── PhotoFolder/
└── Collection4/
    └── PhotoFolderOnly/
```

### В WooCommerce товарах:
- **Source Photo Folder**: `Collection1` (или `Collection2`, etc.)
- **Collection Contents**: описание содержимого коллекции

## 🎛️ Админ-панель в WordPress

### 📋 Мета-бокс "Photo Processing Status":
- **Статус обработки**: pending → processing → pending_approval → approved  
- **Job ID**: уникальный идентификатор задачи
- **Кнопка "Одобрить"**: генерирует ссылку для клиента
- **Ссылка на скачивание**: активна после одобрения

## 🛠️ Быстрый старт: Настройка и тестирование

### Шаг 1: Активация плагинов
1. Откройте **WordPress Admin**: http://localhost:8082/wp-admin
2. **Plugins → Installed Plugins**
3. **Активируйте "ENDECode Integration"** и **"WooCommerce"**

### Шаг 2: Настройка ENDECode API
1. **Settings → ENDECode**
2. **API URL**: `http://endecode-compose-main-photo-processing-web-1:8080`
3. **Нажмите "Test API Connection"** - должно показать ✓ Connection successful!

### Шаг 3: Настройка WooCommerce
1. **WooCommerce → Settings → General**
2. **Store Address**: любой адрес
3. **Currency**: USD ($)

### Шаг 4: Создание тестового товара
1. **Products → Add New**
2. **Product name**: "Portrait Session Collection1"
3. **Regular price**: `299`
4. **Photo Processing Settings**:
   - ✅ **Enable Photo Processing**
   - **Source Photo Folder**: `Collection1`
   - ✅ **Add Invisible Watermarks** (по умолчанию включено)
   - ✅ **Use Order Number** (для кодирования ORDER XXX)
   - ✅ **Automatic Swap by Order** (фото 1↔11, 2↔12, etc.)
   - ✅ **Add Visible Watermarks** (опционально)
   - **Visible Watermark Text**: оставить пустым (будет использовать номер заказа)
   - ✅ **Create ZIP Archive**
5. **Publish**

### Шаг 5: Тестирование покупки
1. **Перейти на сайт**: http://localhost:8082
2. **Shop** или **Products**
3. **Add to Cart** → **View Cart** → **Proceed to Checkout**
4. **Заполнить данные** (любые тестовые)
5. **Place Order**

### Шаг 6: Обработка заказа (админ)
1. **WooCommerce → Orders**
2. **Найти новый заказ** → открыть
3. **Изменить статус на "Completed"** → **Update**
4. **В мета-боксе "Photo Processing Status"** должно появиться:
   - Job ID
   - Статус: processing → pending_approval
5. **Нажать "✅ Approve & Generate Download Link"**
6. **Ссылка для скачивания** готова!

## 🔧 Решение проблем

### Проблема: "Buy Product" не работает
1. **Проверить статус WooCommerce**:
   - WooCommerce → Status → должно быть зеленое
2. **Проверить страницы WooCommerce**:
   - WooCommerce → Settings → Advanced → Page setup
   - Все страницы должны быть назначены
3. **Если нет товаров на главной**:
   - Appearance → Customize → Homepage Settings
   - "Your homepage displays" → "A static page"
   - Homepage: "Shop"

### Проблема: API connection failed
- **Попробовать URL**: `http://endecode-compose-main-photo-processing-web-1:8080`
- **Проверить сеть**: `docker network ls`
- **Перезапустить контейнеры**: `docker-compose restart`

### Проблема: Upload failed
- **Проверить лимиты PHP**: `docker exec endecode-wordpress php -i | grep upload_max_filesize`
- **Должно быть**: 1024M
- **Если нет**: перезапустить WordPress `docker-compose restart wordpress`

### 📁 Загрузка ваших фотографий через терминал:
```bash
# Загрузить фотографии в Collection1
docker cp /путь/к/вашим/фото/* endecode-compose-main-photo-processing-web-1:/app/uploads/OriginalFiles/Collection1/

# Например:
docker cp ~/Pictures/MyPhotos/* endecode-compose-main-photo-processing-web-1:/app/uploads/OriginalFiles/Collection1/

# Или создать новую коллекцию
docker exec -u root endecode-compose-main-photo-processing-web-1 mkdir -p /app/uploads/OriginalFiles/Collection5
docker cp ~/Pictures/Wedding/* endecode-compose-main-photo-processing-web-1:/app/uploads/OriginalFiles/Collection5/
```

## ✨ Новые возможности в действии

### 🔄 Автоматический File Swap:
- **Order 001**: фото 1 ↔ фото 11
- **Order 002**: фото 2 ↔ фото 12  
- **Order 003**: фото 3 ↔ фото 13
- Логика из десктопного приложения: `baseNumber + 10`

### 🏷️ Автоматический Visible Watermark:
- **Если включено "Add Visible Watermarks"** но **текст не указан**
- **Используется номер заказа**: "ORDER 001", "ORDER 002", etc.
- **Если текст указан**: используется указанный текст

### 📧 Уведомления в Telegram и Email:
- **Админу**: "🚨 New Order Processing Complete"
- **Клиенту**: "Your order is being processed" → "Ready for download"
