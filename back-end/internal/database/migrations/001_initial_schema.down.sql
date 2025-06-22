-- internal/database/migrations/001_initial_schema.down.sql

-- Drop triggers
DROP TRIGGER IF EXISTS update_tables_updated_at ON tables;
DROP TRIGGER IF EXISTS update_menu_categories_updated_at ON menu_categories;
DROP TRIGGER IF EXISTS update_menu_items_updated_at ON menu_items;
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS update_order_items_updated_at ON order_items;
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP TRIGGER IF EXISTS update_staff_updated_at ON staff;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_orders_table_id;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_payment_status;
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP INDEX IF EXISTS idx_order_items_status;
DROP INDEX IF EXISTS idx_payments_order_id;
DROP INDEX IF EXISTS idx_payments_midtrans_order_id;
DROP INDEX IF EXISTS idx_menu_items_category_id;
DROP INDEX IF EXISTS idx_menu_items_is_available;
DROP INDEX IF EXISTS idx_customer_sessions_table_id;
DROP INDEX IF EXISTS idx_customer_sessions_token;
DROP INDEX IF EXISTS idx_staff_notifications_staff_id;
DROP INDEX IF EXISTS idx_staff_notifications_is_read;
DROP INDEX IF EXISTS idx_inventory_logs_menu_item_id;
DROP INDEX IF EXISTS idx_media_files_menu_item_id;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS media_files;
DROP TABLE IF EXISTS inventory_logs;
DROP TABLE IF EXISTS staff_notifications;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS customer_sessions;
DROP TABLE IF EXISTS menu_items;
DROP TABLE IF EXISTS menu_categories;
DROP TABLE IF EXISTS staff;
DROP TABLE IF EXISTS tables;