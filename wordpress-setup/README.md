# WordPress + WooCommerce Demo Setup

Этот Docker Compose setup создает локальную среду для тестирования интеграции ENDECode с WordPress/WooCommerce.

## Запуск

```bash
cd wordpress-setup
docker compose up -d
```

## Доступ к сервисам

- **WordPress**: http://localhost:8082
- **phpMyAdmin**: http://localhost:8081 (пользователь: `root`, пароль: `rootpassword`)
- **MySQL**: localhost:3306 (база: `wordpress`, пользователь: `wordpress`, пароль: `wordpress`)

## Первоначальная настройка WordPress

1. Откройте http://localhost:8082
2. Выберите язык и нажмите "Продолжить"
3. Заполните форму установки:
   - **Название сайта**: "PhotoModel Agency Demo"
   - **Имя пользователя**: admin
   - **Пароль**: admin123
   - **Email**: admin@photomodel.local
4. Нажмите "Установить WordPress"

## Установка WooCommerce

1. Войдите в админ панель WordPress: http://localhost:8082/wp-admin
2. Перейдите в "Плагины" → "Добавить новый"
3. Найдите "WooCommerce" и установите его
4. Активируйте плагин
5. Пройдите мастер настройки WooCommerce:
   - **Тип магазина**: Физические товары
   - **Отрасль**: Фотография/Медиа
   - **Где продаете**: Только онлайн
   - **Валюта**: USD ($)
   - **Платежи**: Пропустить (настроим позже)
   - **Доставка**: Пропустить
   - **Налоги**: Не настраивать

## Создание демо-сайта фотомодельного агентства

### Темы
1. Перейдите в "Внешний вид" → "Темы"
2. Установите тему "Astra" или "Storefront" (оптимизированы для WooCommerce)
3. Активируйте выбранную тему

### Структура сайта
Создайте следующие страницы:

1. **Главная страница**
   - Название: "PhotoModel Agency - Professional Photography Services"
   - Слоган: "Capture your beauty with our professional photographers"

2. **О нас** 
   - Расскажите о фотомодельном агентстве
   - Добавьте информацию о услугах

3. **Портфолио**
   - Галерея работ (можно использовать stock фото)

4. **Услуги**
   - Фотосессии разных типов
   - Обработка фотографий
   - Цифровые услуги

### Создание товаров WooCommerce

Создайте следующие категории товаров:
1. **Photo Sessions** - Фотосессии
2. **Digital Processing** - Цифровая обработка
3. **Photo Packages** - Пакеты фотографий

Примеры товаров:
1. **Professional Portrait Session**
   - Цена: $299
   - Категория: Photo Sessions
   - Описание: 2-hour professional portrait session with 50 edited photos

2. **Wedding Photography Package**
   - Цена: $1,500
   - Категория: Photo Sessions
   - Описание: Full day wedding coverage with 300+ edited photos

3. **Photo Processing Service**
   - Цена: $5 per photo
   - Категория: Digital Processing
   - Описание: Professional photo editing and enhancement

### Настройка метаполей для интеграции

Для интеграции с ENDECode нужно добавить кастомные поля к товарам:

```php
// Добавить в functions.php темы или плагин
add_action('woocommerce_product_options_general_product_data', 'add_endecode_fields');
function add_endecode_fields() {
    woocommerce_wp_text_input(array(
        'id' => '_photo_num_copies',
        'label' => 'Number of Copies',
        'placeholder' => '1',
        'description' => 'Number of photo copies to create',
        'type' => 'number'
    ));
    
    woocommerce_wp_text_input(array(
        'id' => '_photo_base_text',
        'label' => 'Base Text for Watermark',
        'placeholder' => 'Client Name',
        'description' => 'Base text for watermarking'
    ));
    
    woocommerce_wp_checkbox(array(
        'id' => '_photo_add_swap',
        'label' => 'Add File Swap Protection'
    ));
    
    woocommerce_wp_checkbox(array(
        'id' => '_photo_add_watermark',
        'label' => 'Add Watermarks',
        'value' => 'yes'
    ));
    
    woocommerce_wp_checkbox(array(
        'id' => '_photo_create_zip',
        'label' => 'Create ZIP Archive',
        'value' => 'yes'
    ));
}

add_action('woocommerce_process_product_meta', 'save_endecode_fields');
function save_endecode_fields($post_id) {
    $fields = ['_photo_num_copies', '_photo_base_text', '_photo_add_swap', '_photo_add_watermark', '_photo_create_zip'];
    foreach ($fields as $field) {
        if (isset($_POST[$field])) {
            update_post_meta($post_id, $field, sanitize_text_field($_POST[$field]));
        }
    }
}
```

## Webhook для интеграции с ENDECode

1. В WooCommerce перейдите в "WooCommerce" → "Настройки" → "Дополнительно" → "Webhooks"
2. Создайте новый webhook:
   - **Name**: ENDECode Processing
   - **Status**: Active
   - **Topic**: Order updated
   - **Delivery URL**: http://host.docker.internal:8090/api/woocommerce/webhook
   - **Secret**: endecode_webhook_secret_2024

## Тестирование

1. Создайте тестовый заказ на сайте
2. Измените статус заказа на "Completed"
3. Проверьте логи ENDECode сервера для обработки webhook

## Остановка

```bash
docker compose down
```

Для полной очистки (включая данные):
```bash
docker compose down -v
```
