-- internal/database/migrations/001_initial_schema.up.sql

-- Database Schema untuk Sistem Pemesanan Restoran

-- 1. Tabel untuk meja restoran
CREATE TABLE IF NOT EXISTS tables (
    id SERIAL PRIMARY KEY,
    table_number VARCHAR(10) UNIQUE NOT NULL,
    qr_code TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'available' CHECK (status IN ('available', 'occupied', 'reserved')),
    capacity INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Tabel kategori menu
CREATE TABLE IF NOT EXISTS menu_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Tabel menu makanan/minuman
CREATE TABLE IF NOT EXISTS menu_items (
    id SERIAL PRIMARY KEY,
    category_id INTEGER REFERENCES menu_categories(id) ON DELETE SET NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    image_url TEXT,
    image_360_url TEXT,
    video_url TEXT,
    is_available BOOLEAN DEFAULT TRUE,
    preparation_time INTEGER, -- dalam menit
    stock_quantity INTEGER DEFAULT NULL, -- NULL = unlimited
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Tabel untuk menyimpan sesi customer
CREATE TABLE IF NOT EXISTS customer_sessions (
    id SERIAL PRIMARY KEY,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    table_id INTEGER REFERENCES tables(id),
    customer_name VARCHAR(100),
    customer_phone VARCHAR(20),
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP
);

-- 5. Tabel order utama
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    session_id INTEGER REFERENCES customer_sessions(id),
    table_id INTEGER REFERENCES tables(id),
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'preparing', 'ready', 'served', 'completed', 'cancelled')),
    total_amount DECIMAL(10,2) NOT NULL,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    service_charge DECIMAL(10,2) DEFAULT 0,
    grand_total DECIMAL(10,2) NOT NULL,
    payment_status VARCHAR(50) DEFAULT 'unpaid' CHECK (payment_status IN ('unpaid', 'pending', 'paid', 'failed', 'refunded')),
    payment_method VARCHAR(50),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 6. Tabel detail order
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id INTEGER REFERENCES menu_items(id),
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    subtotal DECIMAL(10,2) NOT NULL,
    notes TEXT,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'preparing', 'ready', 'served', 'cancelled')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 7. Tabel untuk payment gateway (Midtrans)
CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id),
    midtrans_order_id VARCHAR(255) UNIQUE,
    midtrans_transaction_id VARCHAR(255),
    payment_type VARCHAR(50),
    transaction_status VARCHAR(50),
    transaction_time TIMESTAMP,
    gross_amount DECIMAL(10,2),
    currency VARCHAR(10) DEFAULT 'IDR',
    va_number VARCHAR(100),
    bank VARCHAR(50),
    fraud_status VARCHAR(50),
    signature_key TEXT,
    status_message TEXT,
    raw_response JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 8. Tabel untuk invoice
CREATE TABLE IF NOT EXISTS invoices (
    id SERIAL PRIMARY KEY,
    invoice_number VARCHAR(100) UNIQUE NOT NULL,
    order_id INTEGER REFERENCES orders(id),
    issued_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    due_date TIMESTAMP,
    pdf_url TEXT,
    sent_to_email VARCHAR(255),
    sent_to_whatsapp VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 9. Tabel untuk staff/admin
CREATE TABLE IF NOT EXISTS staff (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    full_name VARCHAR(200) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'cashier', 'waiter', 'kitchen')),
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 10. Tabel untuk notifikasi ke staff
CREATE TABLE IF NOT EXISTS staff_notifications (
    id SERIAL PRIMARY KEY,
    staff_id INTEGER REFERENCES staff(id) ON DELETE CASCADE,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('new_order', 'payment_received', 'order_ready', 'assistance_request')),
    message TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 11. Tabel untuk inventory tracking
CREATE TABLE IF NOT EXISTS inventory_logs (
    id SERIAL PRIMARY KEY,
    menu_item_id INTEGER REFERENCES menu_items(id),
    quantity_change INTEGER NOT NULL, -- negative for deductions
    reason VARCHAR(100),
    order_item_id INTEGER REFERENCES order_items(id),
    staff_id INTEGER REFERENCES staff(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 12. Tabel untuk media files
CREATE TABLE IF NOT EXISTS media_files (
    id SERIAL PRIMARY KEY,
    file_type VARCHAR(50) CHECK (file_type IN ('image', 'video', 'image_360')),
    file_url TEXT NOT NULL,
    thumbnail_url TEXT,
    file_size BIGINT,
    mime_type VARCHAR(100),
    menu_item_id INTEGER REFERENCES menu_items(id) ON DELETE CASCADE,
    uploaded_by INTEGER REFERENCES staff(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes untuk performa
CREATE INDEX idx_orders_table_id ON orders(table_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_payment_status ON orders(payment_status);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_status ON order_items(status);
CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_midtrans_order_id ON payments(midtrans_order_id);
CREATE INDEX idx_menu_items_category_id ON menu_items(category_id);
CREATE INDEX idx_menu_items_is_available ON menu_items(is_available);
CREATE INDEX idx_customer_sessions_table_id ON customer_sessions(table_id);
CREATE INDEX idx_customer_sessions_token ON customer_sessions(session_token);
CREATE INDEX idx_staff_notifications_staff_id ON staff_notifications(staff_id);
CREATE INDEX idx_staff_notifications_is_read ON staff_notifications(is_read);
CREATE INDEX idx_inventory_logs_menu_item_id ON inventory_logs(menu_item_id);
CREATE INDEX idx_media_files_menu_item_id ON media_files(menu_item_id);

-- Trigger untuk update timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$ language 'plpgsql';

CREATE TRIGGER update_tables_updated_at BEFORE UPDATE ON tables
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_menu_categories_updated_at BEFORE UPDATE ON menu_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_menu_items_updated_at BEFORE UPDATE ON menu_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_order_items_updated_at BEFORE UPDATE ON order_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_staff_updated_at BEFORE UPDATE ON staff
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Initial data seed
-- Insert default admin user (password: admin123)
INSERT INTO staff (username, email, password_hash, full_name, role, is_active)
VALUES ('admin', 'admin@restaurant.com', '$2a$10$YourHashedPasswordHere', 'System Admin', 'admin', true);

-- Insert sample tables
INSERT INTO tables (table_number, qr_code, status, capacity) VALUES
('T01', 'QR_T01_GENERATED', 'available', 4),
('T02', 'QR_T02_GENERATED', 'available', 4),
('T03', 'QR_T03_GENERATED', 'available', 6),
('T04', 'QR_T04_GENERATED', 'available', 2),
('T05', 'QR_T05_GENERATED', 'available', 8);

-- Insert sample menu categories
INSERT INTO menu_categories (name, description, display_order, is_active) VALUES
('Appetizers', 'Start your meal with our delicious appetizers', 1, true),
('Main Course', 'Our signature main dishes', 2, true),
('Beverages', 'Refreshing drinks and beverages', 3, true),
('Desserts', 'Sweet treats to end your meal', 4, true);

-- Insert sample menu items
INSERT INTO menu_items (category_id, name, description, price, is_available, preparation_time) VALUES
(1, 'Spring Rolls', 'Crispy vegetable spring rolls with sweet chili sauce', 45000, true, 10),
(1, 'Chicken Satay', 'Grilled chicken skewers with peanut sauce', 55000, true, 15),
(2, 'Nasi Goreng Special', 'Indonesian fried rice with chicken and shrimp', 75000, true, 20),
(2, 'Grilled Salmon', 'Fresh salmon with lemon butter sauce', 120000, true, 25),
(3, 'Fresh Orange Juice', 'Freshly squeezed orange juice', 25000, true, 5),
(3, 'Iced Coffee', 'Indonesian style iced coffee', 30000, true, 5),
(4, 'Chocolate Lava Cake', 'Warm chocolate cake with vanilla ice cream', 45000, true, 15);