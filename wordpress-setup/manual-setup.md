# Ручная настройка WordPress + ENDECode Integration

## 1. Настройка WordPress

Откройте http://localhost:8082 и выполните следующие шаги:

### Установка WordPress
1. **Выберите язык**: English (US)
2. **Заполните форму установки**:
   - Site Title: `PhotoModel Agency`
   - Username: `admin`
   - Password: `admin123` (надежный пароль)
   - Your Email: `admin@photomodel.local`
   - Search engine visibility: не отмечать
3. Нажмите **Install WordPress**

### Установка WooCommerce
1. Войдите в админ панель: http://localhost:8082/wp-admin
2. Перейдите в **Plugins → Add New**
3. Найдите `WooCommerce` и нажмите **Install Now**
4. Нажмите **Activate**
5. Пройдите мастер настройки:
   - **Store Details**: 
     - Address: `123 Photography Street`
     - City: `New York`
     - Country: `United States`
     - State: `New York`
     - Postcode: `10001`
   - **Industry**: `Photography`
   - **Product Types**: `Physical products`
   - **Business Details**: `I'm just starting my business`
   - **Free Features**: оставьте по умолчанию
   - **Theme**: выберите Storefront или оставьте текущую
6. Завершите настройку

## 2. Установка ENDECode Integration плагина

### Способ 1: Копирование файлов
```bash
cd wordpress-setup
docker cp ./plugins/endecode-integration endecode-wordpress:/var/www/html/wp-content/plugins/
```

### Способ 2: Через админ панель WordPress
1. Заархивируйте папку `plugins/endecode-integration` в ZIP файл
2. В админ панели перейдите в **Plugins → Add New → Upload Plugin**
3. Загрузите ZIP файл и активируйте плагин

## 3. Настройка ENDECode плагина

1. Перейдите в **Settings → ENDECode**
2. Заполните настройки:
   - **ENDECode API URL**: `http://host.docker.internal:8090`
   - **API Key**: оставьте пустым (пока не требуется)
   - **Default Link Expiry**: `7` дней
   - **Max Downloads per Link**: `3`
3. Нажмите **Save Changes**
4. Нажмите **Test API Connection** для проверки связи

## 4. Создание демо-товаров

### Категории товаров
Создайте категории в **Products → Categories**:
1. **Portrait Sessions** - Professional portrait photography
2. **Wedding Photography** - Wedding and event photography packages  
3. **Fashion Shoots** - Fashion and model photography sessions
4. **Digital Photos** - Individual digital photo purchases

### Демо-товары
Создайте товары в **Products → Add New**:

#### 1. Professional Portrait Session
- **Product Name**: `Professional Portrait Session`
- **Price**: `$299`
- **Description**: 
  ```
  2-hour professional portrait session in our studio with professional lighting and backdrop. 
  Includes 50 high-resolution edited photos delivered digitally within 24-48 hours.
  
  What's included:
  • 2-hour studio session
  • Professional lighting setup
  • Multiple backdrop options
  • 50 high-resolution edited photos
  • Online gallery for easy sharing
  • Commercial usage rights
  ```
- **Short Description**: `Professional 2-hour portrait session with 50 edited photos`
- **Categories**: Portrait Sessions
- **Photo Processing Settings** (прокрутите вниз):
  - ✅ Enable Photo Processing
  - Number of Copies: `1`
  - Base Text for Watermark: `Portrait Session`
  - ✅ Add Invisible Watermarks
  - ✅ Create ZIP Archive
  - Source Photo Folder: `/photos/portrait-demo`
  - Download Link Expiry: `7 days`

#### 2. Wedding Photography Package
- **Product Name**: `Complete Wedding Photography Package`
- **Price**: `$1999`
- **Description**:
  ```
  Full-day wedding photography coverage from preparation to reception. 
  Professional photographer will capture every precious moment of your special day.
  
  Package includes:
  • 8-10 hours of coverage
  • 500+ high-resolution edited photos
  • Online gallery with download access
  • USB drive with all images
  • Print release for personal use
  • Sneak peek photos within 48 hours
  ```
- **Categories**: Wedding Photography
- **Photo Processing Settings**:
  - ✅ Enable Photo Processing
  - Number of Copies: `2`
  - Base Text for Watermark: `Wedding`
  - ✅ Add File Swap Protection
  - ✅ Add Invisible Watermarks
  - ✅ Create ZIP Archive
  - Source Photo Folder: `/photos/wedding-demo`
  - Download Link Expiry: `14 days`

#### 3. Fashion Model Photoshoot
- **Product Name**: `Fashion Model Photoshoot`
- **Price**: `$599`
- **Description**:
  ```
  3-hour fashion photography session with multiple outfit changes, 
  professional makeup, and styling. Perfect for model portfolios and fashion lookbooks.
  
  Session includes:
  • 3-hour shoot with wardrobe changes
  • Professional makeup artist
  • Multiple background setups
  • 100+ edited high-resolution photos
  • Portfolio-ready images
  • Online gallery access
  ```
- **Categories**: Fashion Shoots
- **Photo Processing Settings**:
  - ✅ Enable Photo Processing
  - Number of Copies: `3`
  - Base Text for Watermark: `Fashion Portfolio`
  - Visible Watermark Text: `PhotoModel Agency`
  - ✅ Add Invisible Watermarks
  - ✅ Create ZIP Archive
  - Source Photo Folder: `/photos/fashion-demo`
  - Download Link Expiry: `10 days`

## 5. Настройка страниц

### Главная страница
1. Создайте новую страницу **Pages → Add New**
2. Название: `Home`
3. Содержимое:
   ```html
   <h1>PhotoModel Agency</h1>
   <p>Professional photography services for models, families, and events. Capture your perfect moments with our experienced photographers.</p>
   
   <h2>Our Services</h2>
   <div class="services-grid">
       <h3>Portrait Photography</h3>
       <p>Professional headshots and portrait sessions for individuals and families.</p>
       
       <h3>Wedding Photography</h3>
       <p>Complete wedding day coverage with hundreds of professional photos.</p>
       
       <h3>Fashion Shoots</h3>
       <p>High-end fashion and model photography for portfolios and commercial use.</p>
   </div>
   
   [products limit="6" columns="3"]
   ```
4. Установите как главную страницу: **Settings → Reading → Homepage displays → A static page → Homepage: Home**

### Меню навигации
1. **Appearance → Menus → Create a new menu**
2. Название меню: `Main Navigation`
3. Добавьте страницы: Home, Shop, About, Contact
4. Назначьте меню: **Theme locations → Primary Menu**

## 6. Тестирование интеграции

### Проверка API связи
1. В админ панели перейдите в **Settings → ENDECode**
2. Нажмите **Test API Connection**
3. Должно появиться: "✓ Connection successful!"

### Тестовый заказ
1. Откройте магазин: http://localhost:8082/shop
2. Добавьте товар в корзину
3. Оформите заказ с тестовыми данными
4. В админ панели измените статус заказа на **Completed**
5. Проверьте логи ENDECode: http://localhost:8090

### Генерация ссылки для скачивания
1. В админ панели откройте заказ: **WooCommerce → Orders**
2. Нажмите кнопку **Generate Download Link**
3. Скопируйте ссылку и отправьте клиенту

## 7. Доступные URL

- **WordPress сайт**: http://localhost:8082
- **WordPress админ**: http://localhost:8082/wp-admin (admin / admin123)
- **ENDECode API**: http://localhost:8090
- **ENDECode админ**: http://localhost:8090 (нажмите Admin)
- **phpMyAdmin**: http://localhost:8081 (root / rootpassword)

## API Endpoints для тестирования

### Тестирование WooCommerce интеграции
```bash
# Симуляция заказа от WooCommerce
curl -X POST http://localhost:8090/api/woocommerce/process-order \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "12345",
    "customer_email": "test@example.com",
    "customer_name": "John Doe",
    "product_id": "1",
    "product_name": "Portrait Session",
    "quantity": 1,
    "settings": {
      "source_folder": "/photos/demo-session",
      "num_copies": 1,
      "base_text": "Portrait Session - John Doe",
      "add_swap": false,
      "add_watermark": true,
      "create_zip": true,
      "watermark_text": "PhotoModel Agency",
      "expiry_days": 7
    }
  }'

# Создание временной ссылки для скачивания
curl -X POST http://localhost:8090/api/downloads/create \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "12345",
    "file_path": "/processed/demo-session.zip",
    "customer_email": "test@example.com",
    "expiry_hours": 168,
    "max_downloads": 3
  }'

# Проверка статуса ссылки
curl http://localhost:8090/api/downloads/status/{token}
```

## Расширенные возможности

### Webhook уведомления
Для автоматической обработки заказов настройте webhook в WooCommerce:
1. **WooCommerce → Settings → Advanced → Webhooks**
2. **Add webhook**:
   - Name: `ENDECode Processing`
   - Status: `Active`
   - Topic: `Order updated`
   - Delivery URL: `http://host.docker.internal:8090/api/woocommerce/webhook`
   - Secret: `endecode_secret_2024`

### Кастомные поля для клиентов
Добавьте дополнительные поля в checkout для персонализации:
- Название для водяного знака
- Особые пожелания
- Количество копий
- Срок действия ссылки
