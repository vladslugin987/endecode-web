<?php
/**
 * Plugin Name: ENDECode Integration
 * Description: Integrates WooCommerce with ENDECode photo processing system
 * Version: 1.0.0
 * Author: ENDECode Team
 * Text Domain: endecode-integration
 */

// Prevent direct access
if (!defined('ABSPATH')) {
    exit;
}

// Define plugin constants
define('ENDECODE_PLUGIN_URL', plugin_dir_url(__FILE__));
define('ENDECODE_PLUGIN_PATH', plugin_dir_path(__FILE__));
define('ENDECODE_VERSION', '1.0.0');

// Main plugin class
class ENDECodeIntegration {
    
    private $api_url;
    private $api_key;
    
    public function __construct() {
        add_action('init', array($this, 'init'));
        
        // Admin hooks
        add_action('admin_menu', array($this, 'add_admin_menu'));
        add_action('admin_init', array($this, 'admin_init'));
        
        // Product hooks
        add_action('woocommerce_product_options_general_product_data', array($this, 'add_photo_processing_fields'));
        add_action('woocommerce_process_product_meta', array($this, 'save_photo_processing_fields'));
        
        // Order hooks
        add_action('woocommerce_order_status_completed', array($this, 'process_completed_order'));
        add_action('woocommerce_thankyou', array($this, 'add_order_status_message'));
        
        // Admin order page hooks
        add_action('add_meta_boxes', array($this, 'add_order_photo_processing_metabox'));
        add_action('wp_ajax_approve_photo_processing', array($this, 'approve_photo_processing'));
        
        // AJAX hooks
        add_action('wp_ajax_test_endecode_connection', array($this, 'test_api_connection'));
        add_action('wp_ajax_generate_download_link', array($this, 'generate_download_link'));
        add_action('wp_ajax_upload_photos_to_endecode', array($this, 'handle_photo_upload'));
        
        // Add download link to order emails
        add_action('woocommerce_email_order_details', array($this, 'add_download_link_to_email'), 10, 4);
        
        // Add shortcode for order status
        add_shortcode('endecode_order_status', array($this, 'order_status_shortcode'));
        
        // Create order status page on activation
        register_activation_hook(__FILE__, array($this, 'create_order_status_page'));
    }
    
    public function create_order_status_page() {
        // Check if page already exists
        $page = get_page_by_path('order-status');
        if (!$page) {
            $page_data = array(
                'post_title'    => 'Order Status',
                'post_content'  => '[endecode_order_status]',
                'post_status'   => 'publish',
                'post_type'     => 'page',
                'post_slug'     => 'order-status'
            );
            wp_insert_post($page_data);
        }
    }
    
    public function add_order_status_message($order_id) {
        $order = wc_get_order($order_id);
        if (!$order) return;
        
        // Check if any products have photo processing enabled
        $has_photo_products = false;
        foreach ($order->get_items() as $item) {
            $product = $item->get_product();
            if (get_post_meta($product->get_id(), '_enable_photo_processing', true) === 'yes') {
                $has_photo_products = true;
                break;
            }
        }
        
        if ($has_photo_products) {
            echo '<div class="woocommerce-message" style="margin: 20px 0; padding: 15px; background: #f7f7f7; border-left: 4px solid #0073aa;">';
            echo '<h3 style="margin-top: 0;">📸 Photo Processing Order</h3>';
            echo '<p><strong>Your order is being processed!</strong></p>';
            echo '<p>Your photos are currently being processed and will be ready soon. You will receive an email notification when your photos are ready for download.</p>';
            echo '<p><strong>What happens next:</strong></p>';
            echo '<ul>';
            echo '<li>✅ Order received and confirmed</li>';
            echo '<li>🔄 Photos are being processed (watermarking, encoding)</li>';
            echo '<li>👨‍💼 Admin review and approval</li>';
            echo '<li>📧 Download link sent to your email</li>';
            echo '</ul>';
            echo '<p><em>Expected processing time: 12-24 hours</em></p>';
            echo '<p>Order ID: <strong>' . $order_id . '</strong></p>';
            echo '</div>';
        }
    }
    
    public function order_status_shortcode($atts) {
        $atts = shortcode_atts(array(
            'order_id' => ''
        ), $atts);
        
        // If no order_id provided, try to get from URL
        if (empty($atts['order_id'])) {
            $atts['order_id'] = isset($_GET['order_id']) ? sanitize_text_field($_GET['order_id']) : '';
        }
        
        if (empty($atts['order_id'])) {
            return '<div class="woocommerce-error">Please provide an order ID.</div>';
        }
        
        $order = wc_get_order($atts['order_id']);
        if (!$order) {
            return '<div class="woocommerce-error">Order not found.</div>';
        }
        
        // Check if user can view this order
        if (!current_user_can('manage_options') && $order->get_customer_id() !== get_current_user_id()) {
            return '<div class="woocommerce-error">You do not have permission to view this order.</div>';
        }
        
        $job_id = get_post_meta($order->get_id(), '_endecode_job_id', true);
        $download_link = get_post_meta($order->get_id(), '_endecode_download_link', true);
        
        ob_start();
        ?>
        <div class="endecode-order-status">
            <h2>Order Status: #<?php echo $order->get_id(); ?></h2>
            
            <div class="order-info" style="background: #f9f9f9; padding: 20px; margin: 20px 0; border-radius: 5px;">
                <h3>Order Information</h3>
                <p><strong>Order Date:</strong> <?php echo $order->get_date_created()->format('F j, Y g:i A'); ?></p>
                <p><strong>Customer:</strong> <?php echo $order->get_billing_first_name() . ' ' . $order->get_billing_last_name(); ?></p>
                <p><strong>Email:</strong> <?php echo $order->get_billing_email(); ?></p>
                <p><strong>Status:</strong> <?php echo wc_get_order_status_name($order->get_status()); ?></p>
            </div>
            
            <?php if ($job_id): ?>
            <div class="processing-status" style="background: #e7f3ff; padding: 20px; margin: 20px 0; border-radius: 5px; border-left: 4px solid #0073aa;">
                <h3>📸 Photo Processing Status</h3>
                <p><strong>Job ID:</strong> <?php echo esc_html($job_id); ?></p>
                
                <?php if ($download_link): ?>
                    <div style="background: #d4edda; padding: 15px; border-radius: 5px; border-left: 4px solid #28a745;">
                        <h4>✅ Your Photos Are Ready!</h4>
                        <p>Your photos have been processed and approved for download.</p>
                        <p><a href="<?php echo esc_url($download_link); ?>" class="button button-primary" style="background: #28a745; border-color: #28a745;">Download Photos</a></p>
                    </div>
                <?php else: ?>
                    <div style="background: #fff3cd; padding: 15px; border-radius: 5px; border-left: 4px solid #ffc107;">
                        <h4>🔄 Processing In Progress</h4>
                        <p>Your photos are currently being processed. This includes:</p>
                        <ul>
                            <li>✅ Photo encoding and watermarking</li>
                            <li>🔄 Quality review and approval</li>
                            <li>⏳ Download link generation</li>
                        </ul>
                        <p><strong>Expected completion:</strong> Within 12-24 hours</p>
                        <p>You will receive an email notification when your photos are ready for download.</p>
                    </div>
                <?php endif; ?>
            </div>
            <?php endif; ?>
            
            <div class="order-items">
                <h3>Order Items</h3>
                <table class="woocommerce-table" style="width: 100%; border-collapse: collapse;">
                    <thead>
                        <tr style="background: #f7f7f7;">
                            <th style="padding: 10px; border: 1px solid #ddd;">Product</th>
                            <th style="padding: 10px; border: 1px solid #ddd;">Quantity</th>
                            <th style="padding: 10px; border: 1px solid #ddd;">Photo Processing</th>
                        </tr>
                    </thead>
                    <tbody>
                        <?php foreach ($order->get_items() as $item): ?>
                            <?php $product = $item->get_product(); ?>
                            <tr>
                                <td style="padding: 10px; border: 1px solid #ddd;"><?php echo $item->get_name(); ?></td>
                                <td style="padding: 10px; border: 1px solid #ddd;"><?php echo $item->get_quantity(); ?></td>
                                <td style="padding: 10px; border: 1px solid #ddd;">
                                    <?php if (get_post_meta($product->get_id(), '_enable_photo_processing', true) === 'yes'): ?>
                                        <span style="color: #28a745;">✅ Enabled</span>
                                    <?php else: ?>
                                        <span style="color: #6c757d;">❌ Not applicable</span>
                                    <?php endif; ?>
                                </td>
                            </tr>
                        <?php endforeach; ?>
                    </tbody>
                </table>
            </div>
        </div>
        <?php
        return ob_get_clean();
    }
    
    public function init() {
        $this->api_url = get_option('endecode_api_url', 'http://host.docker.internal:8090');
        $this->api_key = get_option('endecode_api_key', '');
    }
    
    public function add_admin_menu() {
        add_options_page(
            'ENDECode Settings',
            'ENDECode',
            'manage_options',
            'endecode-settings',
            array($this, 'admin_page')
        );
    }
    
    public function admin_init() {
        register_setting('endecode_settings', 'endecode_api_url');
        register_setting('endecode_settings', 'endecode_api_key');
        register_setting('endecode_settings', 'endecode_default_expiry_days');
        register_setting('endecode_settings', 'endecode_max_downloads');
        
        add_settings_section(
            'endecode_api_section',
            'API Configuration',
            null,
            'endecode_settings'
        );
        
        add_settings_field(
            'endecode_api_url',
            'ENDECode API URL',
            array($this, 'api_url_field'),
            'endecode_settings',
            'endecode_api_section'
        );
        
        add_settings_field(
            'endecode_api_key',
            'API Key',
            array($this, 'api_key_field'),
            'endecode_settings',
            'endecode_api_section'
        );
        
        add_settings_field(
            'endecode_default_expiry_days',
            'Default Link Expiry (days)',
            array($this, 'expiry_days_field'),
            'endecode_settings',
            'endecode_api_section'
        );
        
        add_settings_field(
            'endecode_max_downloads',
            'Max Downloads per Link',
            array($this, 'max_downloads_field'),
            'endecode_settings',
            'endecode_api_section'
        );
    }
    
    public function admin_page() {
        ?>
        <div class="wrap">
            <h1>ENDECode Integration Settings</h1>
            <form method="post" action="options.php">
                <?php
                settings_fields('endecode_settings');
                do_settings_sections('endecode_settings');
                submit_button();
                ?>
            </form>
            
            <h2>Test Connection</h2>
            <p>
                <button type="button" id="test-connection" class="button">Test API Connection</button>
                <span id="connection-result"></span>
            </p>
            
            <script>
            document.getElementById('test-connection').addEventListener('click', function() {
                var button = this;
                var result = document.getElementById('connection-result');
                
                button.disabled = true;
                result.innerHTML = 'Testing...';
                
                fetch(ajaxurl, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/x-www-form-urlencoded',
                    },
                    body: 'action=test_endecode_connection&_wpnonce=<?php echo wp_create_nonce('test_connection'); ?>'
                })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        result.innerHTML = '<span style="color: green;">✓ Connection successful!</span>';
                    } else {
                        result.innerHTML = '<span style="color: red;">✗ Connection failed: ' + data.data + '</span>';
                    }
                })
                .catch(error => {
                    result.innerHTML = '<span style="color: red;">✗ Error: ' + error.message + '</span>';
                })
                .finally(() => {
                    button.disabled = false;
                });
            });
            </script>
        </div>
        <?php
    }
    
    public function api_url_field() {
        $value = get_option('endecode_api_url', 'http://host.docker.internal:8090');
        echo '<input type="url" name="endecode_api_url" value="' . esc_attr($value) . '" class="regular-text" placeholder="http://localhost:8090" />';
        echo '<p class="description">URL of your ENDECode API server</p>';
    }
    
    public function api_key_field() {
        $value = get_option('endecode_api_key', '');
        echo '<input type="password" name="endecode_api_key" value="' . esc_attr($value) . '" class="regular-text" />';
        echo '<p class="description">API key for authentication (if required)</p>';
    }
    
    public function expiry_days_field() {
        $value = get_option('endecode_default_expiry_days', '7');
        echo '<input type="number" name="endecode_default_expiry_days" value="' . esc_attr($value) . '" min="1" max="365" />';
        echo '<p class="description">Default expiration time for download links in days</p>';
    }
    
    public function max_downloads_field() {
        $value = get_option('endecode_max_downloads', '3');
        echo '<input type="number" name="endecode_max_downloads" value="' . esc_attr($value) . '" min="1" max="100" />';
        echo '<p class="description">Maximum number of downloads allowed per link</p>';
    }
    
    public function add_photo_processing_fields() {
        global $post;
        
        echo '<div class="options_group">';
        echo '<h3>Photo Processing Settings</h3>';
        
        woocommerce_wp_checkbox(array(
            'id' => '_enable_photo_processing',
            'label' => 'Enable Photo Processing',
            'description' => 'Enable ENDECode photo processing for this product'
        ));
        
        woocommerce_wp_text_input(array(
            'id' => '_photo_num_copies',
            'label' => 'Number of Copies',
            'placeholder' => '1',
            'description' => 'Number of photo copies to create',
            'type' => 'number',
            'custom_attributes' => array('min' => '1', 'max' => '100')
        ));
        
        woocommerce_wp_text_input(array(
            'id' => '_photo_base_text',
            'label' => 'Base Text for Watermark',
            'placeholder' => 'Customer Name',
            'description' => 'Base text for watermarking (customer name will be appended)'
        ));
        
        woocommerce_wp_checkbox(array(
            'id' => '_photo_use_order_number',
            'label' => 'Use Order Number',
            'description' => 'Use order number instead of customer name for watermarking (format: ORDER XXX)'
        ));
        
        // Advanced settings section
        echo '<h4>Advanced Processing Settings</h4>';
        
        woocommerce_wp_checkbox(array(
            'id' => '_photo_add_swap',
            'label' => 'Add File Swap Protection',
            'description' => 'Enable file swap protection for additional security'
        ));
        
        woocommerce_wp_textarea_input(array(
            'id' => '_photo_swap_pairs',
            'label' => 'File Swap Pairs',
            'placeholder' => 'Auto: Order 001 swaps photo 1 ↔ photo 11' . "\n" . 'Order 002 swaps photo 2 ↔ photo 12' . "\n" . 'Custom: photo_001.jpg <-> photo_005.jpg',
            'description' => 'Leave empty for automatic swap (order number + 10), or specify custom pairs (one per line, format: file1.jpg <-> file2.jpg)'
        ));
        
        woocommerce_wp_checkbox(array(
            'id' => '_photo_auto_swap',
            'label' => 'Automatic Swap by Order',
            'description' => 'Automatically swap photos based on order number (e.g., order 001 swaps photo 1 ↔ photo 11)',
            'value' => 'yes' // Default enabled
        ));
        
        woocommerce_wp_checkbox(array(
            'id' => '_photo_add_watermark',
            'label' => 'Add Invisible Watermarks',
            'description' => 'Add invisible watermarks to photos',
            'value' => 'yes' // Default to enabled
        ));
        
        woocommerce_wp_textarea_input(array(
            'id' => '_photo_watermark_positions',
            'label' => 'Watermark Photo Numbers',
            'placeholder' => '1, 3, 5, 7-10, 15',
            'description' => 'Which photo numbers to watermark (e.g., 1,3,5,7-10,15). Leave empty for all photos.'
        ));
        
        woocommerce_wp_checkbox(array(
            'id' => '_photo_add_visible_watermark',
            'label' => 'Add Visible Watermarks',
            'description' => 'Add visible text watermarks to photos (optional)'
        ));
        
        woocommerce_wp_text_input(array(
            'id' => '_photo_watermark_text',
            'label' => 'Visible Watermark Text',
            'placeholder' => 'Optional visible watermark',
            'description' => 'Optional visible watermark text (only if visible watermarks enabled)'
        ));
        
        woocommerce_wp_checkbox(array(
            'id' => '_photo_create_zip',
            'label' => 'Create ZIP Archive',
            'description' => 'Package processed photos in ZIP archive'
        ));
        
        woocommerce_wp_text_input(array(
            'id' => '_photo_source_folder',
            'label' => 'Source Photo Folder',
            'placeholder' => 'e.g., Collection1 or /photos/Collection1',
            'description' => 'Collection name or path to source photos (e.g., Collection1, Collection2, etc.)'
        ));
        
        woocommerce_wp_textarea_input(array(
            'id' => '_photo_collection_info',
            'label' => 'Collection Contents',
            'placeholder' => 'Example:' . "\n" . 'Photos: 50 high-res images' . "\n" . 'Videos: 2 clips (vid1.mp4, vid2.mp4)',
            'description' => 'Describe what this collection contains (photos, videos, etc.)'
        ));
        
        // File upload section
        echo '<h4>Photo Upload</h4>';
        echo '<div style="margin: 10px 0;">';
        echo '<label><strong>Upload Photos from Computer:</strong></label><br>';
        echo '<input type="file" id="photo_upload" multiple webkitdirectory accept="image/*,video/*" style="margin: 5px 0;">';
        echo '<p class="description">Select a folder from your computer to upload photos. All files will be uploaded to ENDECode server.</p>';
        echo '<button type="button" id="upload_photos" class="button" disabled>Upload Selected Folder</button>';
        echo '<div id="upload_progress" style="margin-top: 10px; display: none;">';
        echo '<progress id="upload_progress_bar" value="0" max="100" style="width: 100%;"></progress>';
        echo '<div id="upload_status"></div>';
        echo '</div>';
        echo '</div>';
        
        woocommerce_wp_select(array(
            'id' => '_download_expiry_days',
            'label' => 'Download Link Expiry',
            'options' => array(
                '1' => '1 day',
                '3' => '3 days',
                '7' => '1 week',
                '14' => '2 weeks',
                '30' => '1 month',
                '90' => '3 months'
            ),
            'description' => 'How long the download link should remain valid'
        ));
        
        // JavaScript for file upload
        echo '<script>
        document.addEventListener("DOMContentLoaded", function() {
            const fileInput = document.getElementById("photo_upload");
            const uploadButton = document.getElementById("upload_photos");
            const progressDiv = document.getElementById("upload_progress");
            const progressBar = document.getElementById("upload_progress_bar");
            const statusDiv = document.getElementById("upload_status");
            const sourceFolder = document.getElementById("_photo_source_folder");
            
            fileInput.addEventListener("change", function() {
                uploadButton.disabled = fileInput.files.length === 0;
            });
            
            uploadButton.addEventListener("click", function() {
                if (fileInput.files.length === 0) return;
                
                const formData = new FormData();
                const folderName = "product_" + ' . $post->ID . ' + "_" + Date.now();
                
                formData.append("action", "upload_photos_to_endecode");
                formData.append("folderName", folderName);
                formData.append("_wpnonce", "' . wp_create_nonce('upload_photos') . '");
                
                Array.from(fileInput.files).forEach((file, index) => {
                    formData.append("photos[]", file);
                });
                
                progressDiv.style.display = "block";
                uploadButton.disabled = true;
                statusDiv.textContent = "Uploading...";
                
                fetch(ajaxurl, {
                    method: "POST",
                    body: formData
                })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        statusDiv.innerHTML = "<span style=\"color: green;\">✓ Upload successful! " + data.data.files_uploaded + " files uploaded.</span>";
                        sourceFolder.value = data.data.folder_path;
                        progressBar.value = 100;
                    } else {
                        statusDiv.innerHTML = "<span style=\"color: red;\">✗ Upload failed: " + data.data + "</span>";
                    }
                })
                .catch(error => {
                    statusDiv.innerHTML = "<span style=\"color: red;\">✗ Error: " + error.message + "</span>";
                })
                .finally(() => {
                    uploadButton.disabled = false;
                });
            });
        });
        </script>';
        
        echo '</div>';
    }
    
    public function save_photo_processing_fields($post_id) {
        $fields = [
            '_enable_photo_processing',
            '_photo_num_copies',
            '_photo_base_text',
            '_photo_use_order_number',
            '_photo_add_swap',
            '_photo_swap_pairs',
            '_photo_auto_swap',
            '_photo_add_watermark',
            '_photo_watermark_positions',
            '_photo_add_visible_watermark',
            '_photo_create_zip',
            '_photo_watermark_text',
            '_photo_source_folder',
            '_photo_collection_info',
            '_download_expiry_days'
        ];
        
        foreach ($fields as $field) {
            if (isset($_POST[$field])) {
                update_post_meta($post_id, $field, sanitize_textarea_field($_POST[$field]));
            } else {
                // For checkboxes, delete meta if not checked
                if (strpos($field, '_photo_add_') !== false || $field === '_enable_photo_processing') {
                    delete_post_meta($post_id, $field);
                }
            }
        }
    }
    
    public function handle_photo_upload() {
        // Verify nonce
        if (!wp_verify_nonce($_POST['_wpnonce'], 'upload_photos')) {
            wp_send_json_error('Invalid nonce');
            return;
        }
        
        if (empty($_FILES['photos'])) {
            wp_send_json_error('No files uploaded');
            return;
        }
        
        $folder_name = sanitize_text_field($_POST['folderName']);
        $api_url = get_option('endecode_api_url');
        
        if (empty($api_url)) {
            wp_send_json_error('ENDECode API URL not configured');
            return;
        }
        
        // Prepare files for upload to ENDECode
        $files = $_FILES['photos'];
        $upload_data = array();
        
        // Handle both single and multiple file uploads
        if (is_array($files['name'])) {
            // Multiple files
            for ($i = 0; $i < count($files['name']); $i++) {
                if ($files['error'][$i] === UPLOAD_ERR_OK) {
                    $upload_data[] = array(
                        'name' => $files['name'][$i],
                        'tmp_name' => $files['tmp_name'][$i],
                        'type' => $files['type'][$i],
                        'size' => $files['size'][$i]
                    );
                }
            }
        } else {
            // Single file
            if ($files['error'] === UPLOAD_ERR_OK) {
                $upload_data[] = array(
                    'name' => $files['name'],
                    'tmp_name' => $files['tmp_name'],
                    'type' => $files['type'],
                    'size' => $files['size']
                );
            }
        }
        
        if (empty($upload_data)) {
            wp_send_json_error('No valid files to upload');
            return;
        }
        
        // Create proper multipart/form-data request
        $boundary = wp_generate_password(24, false);
        $delimiter = '-------------' . $boundary;
        $eol = "\r\n";
        
        $body = '';
        
        // Add folder name field
        $body .= '--' . $delimiter . $eol;
        $body .= 'Content-Disposition: form-data; name="folderName"' . $eol . $eol;
        $body .= $folder_name . $eol;
        
        // Add files
        foreach ($upload_data as $file) {
            $body .= '--' . $delimiter . $eol;
            $body .= 'Content-Disposition: form-data; name="photos[]"; filename="' . $file['name'] . '"' . $eol;
            $body .= 'Content-Type: ' . $file['type'] . $eol . $eol;
            $body .= file_get_contents($file['tmp_name']) . $eol;
        }
        
        $body .= '--' . $delimiter . '--' . $eol;
        
        $response = wp_remote_post($api_url . '/api/upload', array(
            'timeout' => 300, // 5 minutes for large uploads
            'headers' => array(
                'Content-Type' => 'multipart/form-data; boundary=' . $delimiter,
                'Content-Length' => strlen($body)
            ),
            'body' => $body
        ));
        
        if (is_wp_error($response)) {
            wp_send_json_error('Upload failed: ' . $response->get_error_message());
            return;
        }
        
        $response_code = wp_remote_retrieve_response_code($response);
        $response_body = wp_remote_retrieve_body($response);
        
        // Log for debugging
        error_log('ENDECode Upload Response Code: ' . $response_code);
        error_log('ENDECode Upload Response Body: ' . $response_body);
        
        if ($response_code !== 200) {
            wp_send_json_error('Upload failed with status: ' . $response_code . '. Response: ' . $response_body);
            return;
        }
        
        $response_data = json_decode($response_body, true);
        
        if (isset($response_data['success']) && $response_data['success']) {
            wp_send_json_success(array(
                'files_uploaded' => count($upload_data),
                'folder_path' => '/uploads/' . $folder_name,
                'message' => 'Photos uploaded successfully'
            ));
        } else {
            $error = isset($response_data['error']) ? $response_data['error'] : 'Unknown error';
            wp_send_json_error('Upload failed: ' . $error);
        }
    }
    
    public function process_completed_order($order_id) {
        $order = wc_get_order($order_id);
        if (!$order) return;
        
        foreach ($order->get_items() as $item) {
            $product = $item->get_product();
            
            // Check if photo processing is enabled for this product
            if (get_post_meta($product->get_id(), '_enable_photo_processing', true) !== 'yes') {
                continue;
            }
            
            $this->create_processing_job($order, $item, $product);
        }
    }
    
    private function create_processing_job($order, $item, $product) {
        $customer_name = $order->get_billing_first_name() . ' ' . $order->get_billing_last_name();
        $base_text = get_post_meta($product->get_id(), '_photo_base_text', true) ?: 'Customer';
        
        // Determine watermark text
        if (get_post_meta($product->get_id(), '_photo_use_order_number', true) === 'yes') {
            $watermark_text = 'ORDER ' . $order->get_order_number();
        } else {
            $watermark_text = $base_text . ' - ' . $customer_name;
        }
        
        // Parse swap pairs
        $swap_pairs_text = get_post_meta($product->get_id(), '_photo_swap_pairs', true);
        $auto_swap = get_post_meta($product->get_id(), '_photo_auto_swap', true) === 'yes';
        $swap_pairs = array();
        
        if ($auto_swap || empty($swap_pairs_text)) {
            // Automatic swap mode: order number + 10
            $swap_pairs[] = array(
                'mode' => 'auto',
                'description' => 'Automatic swap based on order number (baseNumber + 10)'
            );
        } else {
            // Manual mode: parse custom pairs
            $lines = explode("\n", $swap_pairs_text);
            foreach ($lines as $line) {
                $line = trim($line);
                if (preg_match('/(.+?)\s*<->\s*(.+)/', $line, $matches)) {
                    $swap_pairs[] = array(
                        'file1' => trim($matches[1]),
                        'file2' => trim($matches[2])
                    );
                }
            }
        }
        
        // Parse watermark positions
        $watermark_positions_text = get_post_meta($product->get_id(), '_photo_watermark_positions', true);
        $watermark_positions = array();
        if ($watermark_positions_text) {
            $positions = explode(',', $watermark_positions_text);
            foreach ($positions as $pos) {
                $pos = trim($pos);
                if (strpos($pos, '-') !== false) {
                    // Range like "7-10"
                    list($start, $end) = explode('-', $pos);
                    for ($i = intval($start); $i <= intval($end); $i++) {
                        $watermark_positions[] = $i;
                    }
                } else {
                    // Single number
                    $watermark_positions[] = intval($pos);
                }
            }
        }
        
        $processing_data = array(
            'order_id' => $order->get_id(),
            'customer_email' => $order->get_billing_email(),
            'customer_name' => $customer_name,
            'product_id' => $product->get_id(),
            'product_name' => $product->get_name(),
            'quantity' => $item->get_quantity(),
            'settings' => array(
                'source_folder' => get_post_meta($product->get_id(), '_photo_source_folder', true),
                'num_copies' => (int) get_post_meta($product->get_id(), '_photo_num_copies', true) ?: 1,
                'base_text' => $watermark_text,
                'add_swap' => get_post_meta($product->get_id(), '_photo_add_swap', true) === 'yes',
                'swap_pairs' => $swap_pairs,
                'add_watermark' => get_post_meta($product->get_id(), '_photo_add_watermark', true) === 'yes',
                'watermark_positions' => $watermark_positions,
                'add_visible_watermark' => get_post_meta($product->get_id(), '_photo_add_visible_watermark', true) === 'yes',
                'visible_watermark_text' => get_post_meta($product->get_id(), '_photo_watermark_text', true) ?: $watermark_text,
                'create_zip' => get_post_meta($product->get_id(), '_photo_create_zip', true) === 'yes',
                'zip_name' => $this->generate_zip_name($product, $order),
                'watermark_text' => $watermark_text,
                'expiry_days' => (int) get_post_meta($product->get_id(), '_download_expiry_days', true) ?: 7
            )
        );
        
        $response = $this->send_to_endecode('/api/woocommerce/process-order', $processing_data);
        
        if ($response && isset($response['success']) && $response['success']) {
            // Add order note
            $order->add_order_note(
                sprintf('Photo processing job created: %s', $response['job_id']),
                false,
                true
            );
            
            // Store job ID and initial status
            update_post_meta($order->get_id(), '_endecode_job_id', $response['job_id']);
            update_post_meta($order->get_id(), '_endecode_processing_status', 'processing');
        } else {
            $error_message = isset($response['error']) ? $response['error'] : 'Unknown error';
            $order->add_order_note(
                sprintf('Photo processing failed: %s', $error_message),
                false,
                true
            );
            
            update_post_meta($order->get_id(), '_endecode_processing_status', 'failed');
        }
    }
    
    public function test_api_connection() {
        check_ajax_referer('test_connection');
        
        if (!current_user_can('manage_options')) {
            wp_die('Unauthorized');
        }
        
        $response = $this->send_to_endecode('/api/info');
        
        if ($response && isset($response['name'])) {
            wp_send_json_success('Connected to ' . $response['name']);
        } else {
            wp_send_json_error('Could not connect to ENDECode API');
        }
    }
    
    public function generate_download_link() {
        check_ajax_referer('generate_download_link');
        
        $order_id = intval($_POST['order_id']);
        $expiry_days = intval($_POST['expiry_days']) ?: 7;
        
        $job_id = get_post_meta($order_id, '_endecode_job_id', true);
        if (!$job_id) {
            wp_send_json_error('No processing job found for this order');
        }
        
        $data = array(
            'job_id' => $job_id,
            'expiry_days' => $expiry_days
        );
        
        $response = $this->send_to_endecode('/api/admin/jobs/' . $job_id . '/approve', $data, 'POST');
        
        if ($response && isset($response['token'])) {
            $download_url = $this->api_url . '/api/download/' . $response['token'];
            
            // Store the download link
            update_post_meta($order_id, '_endecode_download_link', $download_url);
            update_post_meta($order_id, '_endecode_download_expiry', date('Y-m-d H:i:s', strtotime('+' . $expiry_days . ' days')));
            
            wp_send_json_success(array('download_url' => $download_url));
        } else {
            wp_send_json_error('Failed to generate download link');
        }
    }
    
    public function add_download_link_to_email($order, $sent_to_admin, $plain_text, $email) {
        if ($sent_to_admin || $email->id !== 'customer_completed_order') {
            return;
        }
        
        $download_link = get_post_meta($order->get_id(), '_endecode_download_link', true);
        $expiry = get_post_meta($order->get_id(), '_endecode_download_expiry', true);
        
        if ($download_link) {
            if ($plain_text) {
                echo "\n\nYour processed photos are ready for download!\n";
                echo "Download Link: " . $download_link . "\n";
                echo "Link expires: " . date('F j, Y g:i A', strtotime($expiry)) . "\n";
            } else {
                echo '<h2>Your Photos Are Ready!</h2>';
                echo '<p>Your processed photos are ready for download.</p>';
                echo '<p><strong><a href="' . esc_url($download_link) . '" style="background-color: #0073aa; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Download Photos</a></strong></p>';
                echo '<p><small>This link expires on ' . date('F j, Y g:i A', strtotime($expiry)) . '</small></p>';
            }
        }
    }
    
    private function send_to_endecode($endpoint, $data = null, $method = 'GET') {
        $url = $this->api_url . $endpoint;
        
        $args = array(
            'method' => $method,
            'timeout' => 30,
            'headers' => array(
                'Content-Type' => 'application/json'
            )
        );
        
        if ($this->api_key) {
            $args['headers']['Authorization'] = 'Bearer ' . $this->api_key;
        }
        
        if ($data && $method !== 'GET') {
            $args['body'] = json_encode($data);
        }
        
        $response = wp_remote_request($url, $args);
        
        if (is_wp_error($response)) {
            return false;
        }
        
        $body = wp_remote_retrieve_body($response);
        return json_decode($body, true);
    }

    private function generate_zip_name($product, $order) {
        $product_name = $product->get_name();
        $order_number = $order->get_order_number();
        $collection_name = get_post_meta($product->get_id(), '_photo_source_folder', true);

        // Clean up product name for filename
        $clean_product_name = sanitize_file_name($product_name);
        $clean_collection_name = sanitize_file_name($collection_name);

        // Construct the zip filename
        $zip_name = $clean_collection_name . '-' . $order_number . '.zip';

        // Ensure uniqueness if multiple products in the same order
        $existing_zips = get_posts(array(
            'post_type' => 'attachment',
            'post_status' => 'any',
            'meta_key' => '_wp_attached_file',
            'meta_value' => $zip_name,
            'post_parent' => $order->get_id(),
            'fields' => 'ids'
        ));

        if (!empty($existing_zips)) {
            $counter = 1;
            do {
                $new_zip_name = $clean_collection_name . '-' . $order_number . '-' . $counter . '.zip';
                $counter++;
            } while (get_posts(array(
                'post_type' => 'attachment',
                'post_status' => 'any',
                'meta_key' => '_wp_attached_file',
                'meta_value' => $new_zip_name,
                'post_parent' => $order->get_id(),
                'fields' => 'ids'
            )));
            $zip_name = $new_zip_name;
        }

        return $zip_name;
    }
    
    public function add_order_photo_processing_metabox() {
        global $post;
        
        if (!$post || $post->post_type !== 'shop_order') {
            return;
        }
        
        $order = wc_get_order($post->ID);
        if (!$order) {
            return;
        }
        
        // Check if order has photo processing products
        $has_photo_processing = false;
        foreach ($order->get_items() as $item) {
            $product = $item->get_product();
            if (get_post_meta($product->get_id(), '_enable_photo_processing', true) === 'yes') {
                $has_photo_processing = true;
                break;
            }
        }
        
        if ($has_photo_processing) {
            add_meta_box(
                'endecode_photo_processing',
                '📸 Photo Processing Status',
                array($this, 'render_photo_processing_metabox'),
                'shop_order',
                'side',
                'high'
            );
        }
    }
    
    public function render_photo_processing_metabox($post) {
        $order = wc_get_order($post->ID);
        $job_id = get_post_meta($order->get_id(), '_endecode_job_id', true);
        $download_link = get_post_meta($order->get_id(), '_endecode_download_link', true);
        $processing_status = get_post_meta($order->get_id(), '_endecode_processing_status', true);
        
        ?>
        <div id="endecode-processing-status">
            <?php if ($job_id): ?>
                <p><strong>Job ID:</strong> <?php echo esc_html($job_id); ?></p>
                <p><strong>Status:</strong> 
                    <span class="status-<?php echo esc_attr($processing_status); ?>">
                        <?php echo ucfirst($processing_status ?: 'pending'); ?>
                    </span>
                </p>
                
                <?php if ($processing_status === 'pending_approval'): ?>
                    <div style="background: #fff3cd; padding: 10px; margin: 10px 0; border-left: 4px solid #ffc107;">
                        <p><strong>⏳ Ready for Admin Approval</strong></p>
                        <p>Photos have been processed and are ready for review.</p>
                        <button type="button" class="button button-primary" id="approve-processing" 
                                data-order-id="<?php echo $order->get_id(); ?>">
                            ✅ Approve & Generate Download Link
                        </button>
                    </div>
                <?php elseif ($download_link): ?>
                    <div style="background: #d4edda; padding: 10px; margin: 10px 0; border-left: 4px solid #28a745;">
                        <p><strong>✅ Approved & Ready</strong></p>
                        <p><a href="<?php echo esc_url($download_link); ?>" target="_blank" class="button">📥 Download Link</a></p>
                    </div>
                <?php elseif ($processing_status === 'processing'): ?>
                    <div style="background: #cce5ff; padding: 10px; margin: 10px 0; border-left: 4px solid #0073aa;">
                        <p><strong>🔄 Processing Photos</strong></p>
                        <p>Photos are being processed. Please wait...</p>
                    </div>
                <?php endif; ?>
            <?php else: ?>
                <p>No photo processing job found for this order.</p>
                <?php if ($order->get_status() === 'completed'): ?>
                    <button type="button" class="button" onclick="location.reload();">🔄 Refresh Status</button>
                <?php endif; ?>
            <?php endif; ?>
        </div>
        
        <script>
        document.addEventListener('DOMContentLoaded', function() {
            const approveBtn = document.getElementById('approve-processing');
            if (approveBtn) {
                approveBtn.addEventListener('click', function() {
                    const orderId = this.dataset.orderId;
                    const button = this;
                    
                    button.disabled = true;
                    button.textContent = 'Processing...';
                    
                    fetch(ajaxurl, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/x-www-form-urlencoded',
                        },
                        body: 'action=approve_photo_processing&order_id=' + orderId + '&_wpnonce=<?php echo wp_create_nonce('approve_processing'); ?>'
                    })
                    .then(response => response.json())
                    .then(data => {
                        if (data.success) {
                            location.reload();
                        } else {
                            alert('Error: ' + data.data);
                            button.disabled = false;
                            button.textContent = '✅ Approve & Generate Download Link';
                        }
                    })
                    .catch(error => {
                        alert('Error: ' + error.message);
                        button.disabled = false;
                        button.textContent = '✅ Approve & Generate Download Link';
                    });
                });
            }
        });
        </script>
        
        <style>
        .status-pending { color: #f57c00; }
        .status-processing { color: #1976d2; }
        .status-pending_approval { color: #ff9800; }
        .status-approved { color: #388e3c; }
        .status-failed { color: #d32f2f; }
        </style>
        <?php
    }
    
    public function approve_photo_processing() {
        check_ajax_referer('approve_processing');
        
        if (!current_user_can('manage_woocommerce')) {
            wp_send_json_error('Insufficient permissions');
            return;
        }
        
        $order_id = intval($_POST['order_id']);
        $order = wc_get_order($order_id);
        
        if (!$order) {
            wp_send_json_error('Order not found');
            return;
        }
        
        $job_id = get_post_meta($order_id, '_endecode_job_id', true);
        if (!$job_id) {
            wp_send_json_error('No processing job found');
            return;
        }
        
        // Call ENDECode API to approve and generate download link
        $response = $this->send_to_endecode('/api/admin/jobs/' . $job_id . '/approve', array(
            'order_id' => $order_id,
            'expiry_days' => 7
        ), 'POST');
        
        if ($response && isset($response['success']) && $response['success']) {
            // Update order status
            update_post_meta($order_id, '_endecode_processing_status', 'approved');
            
            if (isset($response['download_link'])) {
                update_post_meta($order_id, '_endecode_download_link', $response['download_link']);
            }
            
            // Add order note
            $order->add_order_note('Photo processing approved by admin. Download link generated.', false, true);
            
            wp_send_json_success('Processing approved successfully');
        } else {
            $error = isset($response['error']) ? $response['error'] : 'Failed to approve processing';
            wp_send_json_error($error);
        }
    }
}

// Initialize the plugin
new ENDECodeIntegration();

// Add admin order actions
add_action('woocommerce_order_item_add_action_buttons', function($order) {
    $job_id = get_post_meta($order->get_id(), '_endecode_job_id', true);
    if ($job_id) {
        echo '<button type="button" class="button generate-download-link" data-order-id="' . $order->get_id() . '">Generate Download Link</button>';
    }
});

// Add admin scripts
add_action('admin_footer', function() {
    if (get_current_screen()->id === 'shop_order') {
        ?>
        <script>
        jQuery(document).ready(function($) {
            $('.generate-download-link').click(function() {
                var orderId = $(this).data('order-id');
                var button = $(this);
                
                button.prop('disabled', true).text('Generating...');
                
                $.post(ajaxurl, {
                    action: 'generate_download_link',
                    order_id: orderId,
                    expiry_days: 7,
                    _wpnonce: '<?php echo wp_create_nonce('generate_download_link'); ?>'
                }, function(response) {
                    if (response.success) {
                        alert('Download link generated: ' + response.data.download_url);
                        location.reload();
                    } else {
                        alert('Error: ' + response.data);
                    }
                }).always(function() {
                    button.prop('disabled', false).text('Generate Download Link');
                });
            });
        });
        </script>
        <?php
    }
});
?>
