#!/bin/bash

# Автоматическая настройка WordPress для фотомодельного агентства
echo "🚀 Начинаем автоматическую настройку WordPress..."

# Ждем пока WordPress будет готов
echo "⏳ Ожидание запуска WordPress..."
sleep 30

# Функция для выполнения WP-CLI команд
wp_cli() {
    docker exec endecode-wordpress wp --allow-root "$@"
}

# Проверка готовности WordPress
until wp_cli core is-installed 2>/dev/null; do
    echo "🔄 Установка WordPress..."
    wp_cli core install \
        --url="http://localhost:8082" \
        --title="PhotoModel Agency - Professional Photography" \
        --admin_user="admin" \
        --admin_password="admin123" \
        --admin_email="admin@photomodel.local" \
        --skip-email
    sleep 5
done

echo "✅ WordPress установлен!"

# Установка и активация тем и плагинов
echo "📦 Установка тем и плагинов..."

# Установка темы Astra (бесплатная и хорошо работает с WooCommerce)
wp_cli theme install astra --activate

# Установка WooCommerce
wp_cli plugin install woocommerce --activate

# Установка дополнительных плагинов
wp_cli plugin install contact-form-7 --activate
wp_cli plugin install wp-gallery-custom-links --activate

echo "✅ Темы и плагины установлены!"

# Настройка WooCommerce
echo "🛒 Настройка WooCommerce..."

# Базовые настройки WooCommerce
wp_cli option update woocommerce_store_address "123 Photography St"
wp_cli option update woocommerce_store_city "New York"
wp_cli option update woocommerce_default_country "US:NY"
wp_cli option update woocommerce_store_postcode "10001"
wp_cli option update woocommerce_currency "USD"
wp_cli option update woocommerce_product_type "both"
wp_cli option update woocommerce_allow_tracking "no"

# Создание базовых страниц WooCommerce
wp_cli wc --user=admin tool run install_pages

echo "✅ WooCommerce настроен!"

# Создание категорий товаров
echo "📂 Создание категорий товаров..."

wp_cli wc product_cat create \
    --user=admin \
    --name="Portrait Sessions" \
    --slug="portrait-sessions" \
    --description="Professional portrait photography sessions"

wp_cli wc product_cat create \
    --user=admin \
    --name="Wedding Photography" \
    --slug="wedding-photography" \
    --description="Wedding and event photography packages"

wp_cli wc product_cat create \
    --user=admin \
    --name="Fashion Shoots" \
    --slug="fashion-shoots" \
    --description="Fashion and model photography sessions"

wp_cli wc product_cat create \
    --user=admin \
    --name="Digital Photos" \
    --slug="digital-photos" \
    --description="Individual digital photo purchases"

echo "✅ Категории созданы!"

# Создание демо-товаров
echo "🛍️ Создание демо-товаров..."

# Portrait Session
wp_cli wc product create \
    --user=admin \
    --name="Professional Portrait Session" \
    --type="simple" \
    --regular_price="299" \
    --description="2-hour professional portrait session in our studio with professional lighting and backdrop. Includes 50 high-resolution edited photos delivered digitally." \
    --short_description="Professional 2-hour portrait session with 50 edited photos" \
    --manage_stock=true \
    --stock_quantity=10 \
    --categories='[{"id":4}]' \
    --status="publish"

# Wedding Package
wp_cli wc product create \
    --user=admin \
    --name="Complete Wedding Photography Package" \
    --type="simple" \
    --regular_price="1999" \
    --description="Full-day wedding photography coverage from preparation to reception. Includes 500+ high-resolution edited photos, online gallery, and USB drive with all images." \
    --short_description="Complete wedding day coverage with 500+ photos" \
    --manage_stock=true \
    --stock_quantity=5 \
    --categories='[{"id":5}]' \
    --status="publish"

# Fashion Shoot
wp_cli wc product create \
    --user=admin \
    --name="Fashion Model Photoshoot" \
    --type="simple" \
    --regular_price="599" \
    --description="3-hour fashion photography session with multiple outfit changes, professional makeup, and styling. Perfect for model portfolios and fashion lookbooks." \
    --short_description="3-hour fashion shoot with styling and makeup" \
    --manage_stock=true \
    --stock_quantity=8 \
    --categories='[{"id":6}]' \
    --status="publish"

echo "✅ Демо-товары созданы!"

# Создание основных страниц
echo "📄 Создание страниц сайта..."

# Главная страница
wp_cli post create --post_type=page --post_title="Home" --post_content="
<h1>Welcome to PhotoModel Agency</h1>
<p>Professional photography services for models, families, and events. Capture your perfect moments with our experienced photographers.</p>

<h2>Our Services</h2>
<div class='services-grid'>
    <div class='service'>
        <h3>Portrait Photography</h3>
        <p>Professional headshots and portrait sessions</p>
    </div>
    <div class='service'>
        <h3>Wedding Photography</h3>
        <p>Complete wedding day coverage</p>
    </div>
    <div class='service'>
        <h3>Fashion Shoots</h3>
        <p>High-end fashion and model photography</p>
    </div>
</div>

[products limit='6' columns='3']
" --post_status=publish

# О нас
wp_cli post create --post_type=page --post_title="About Us" --post_content="
<h1>About PhotoModel Agency</h1>
<p>Founded in 2020, PhotoModel Agency has been providing professional photography services to models, families, and businesses. Our team of experienced photographers specializes in capturing the perfect shot for any occasion.</p>

<h2>Our Team</h2>
<p>Our photographers have years of experience in fashion, portrait, and wedding photography. We use the latest equipment and techniques to ensure you get the best possible results.</p>

<h2>Why Choose Us?</h2>
<ul>
<li>Professional equipment and studio</li>
<li>Experienced photographers</li>
<li>Quick turnaround time</li>
<li>High-resolution digital delivery</li>
<li>Satisfaction guarantee</li>
</ul>
" --post_status=publish

# Портфолио
wp_cli post create --post_type=page --post_title="Portfolio" --post_content="
<h1>Our Portfolio</h1>
<p>Take a look at some of our recent work. Each photo tells a story, and we're here to help tell yours.</p>

<h2>Recent Sessions</h2>
<p>Browse through our gallery to see the quality and style of our photography work.</p>

[gallery ids='1,2,3,4,5,6' columns='3']
" --post_status=publish

# Установка главной страницы
HOME_ID=$(wp_cli post list --post_type=page --title="Home" --field=ID --format=csv)
wp_cli option update show_on_front page
wp_cli option update page_on_front $HOME_ID

echo "✅ Страницы созданы!"

# Настройка меню
echo "📋 Создание меню..."

# Создание главного меню
wp_cli menu create "Main Menu"
wp_cli menu item add-post main-menu $HOME_ID
wp_cli menu item add-post main-menu $(wp_cli post list --post_type=page --title="About Us" --field=ID --format=csv)
wp_cli menu item add-post main-menu $(wp_cli post list --post_type=page --title="Portfolio" --field=ID --format=csv)
wp_cli menu item add-post main-menu $(wp_cli post list --post_type=page --title="Shop" --field=ID --format=csv)

# Назначение меню
wp_cli menu location assign main-menu primary

echo "✅ Меню создано!"

# Настройка permalink structure
wp_cli rewrite structure '/%postname%/'
wp_cli rewrite flush

echo "🎉 WordPress настроен! Доступен по адресу: http://localhost:8082"
echo "👨‍💼 Админ панель: http://localhost:8082/wp-admin"
echo "🔑 Логин: admin / Пароль: admin123"
