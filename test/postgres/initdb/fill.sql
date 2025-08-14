\connect appdb

-- set up character set and timezone
SET client_encoding = 'UTF8';
SET timezone = 'UTC';

-- drop objects if they exist (PostgreSQL syntax)
DROP EVENT TRIGGER IF EXISTS ev_cleanup_audit_log CASCADE;
DROP VIEW IF EXISTS v_user_order_stats CASCADE;
DROP VIEW IF EXISTS v_order_items_detailed CASCADE;
DROP TRIGGER IF EXISTS trg_users_bi_normalize_email ON users CASCADE;
DROP TRIGGER IF EXISTS trg_order_items_bi_fill_price ON order_items CASCADE;
DROP TRIGGER IF EXISTS trg_order_items_ai_stock_and_total ON order_items CASCADE;
DROP TRIGGER IF EXISTS trg_order_items_au_stock_and_total ON order_items CASCADE;
DROP TRIGGER IF EXISTS trg_order_items_ad_stock_and_total ON order_items CASCADE;
DROP FUNCTION IF EXISTS sp_generate_users(INTEGER) CASCADE;
DROP FUNCTION IF EXISTS sp_seed_catalog(INTEGER) CASCADE;
DROP FUNCTION IF EXISTS sp_generate_orders(INTEGER, INTEGER) CASCADE;
DROP FUNCTION IF EXISTS fn_users_bi_normalize_email() CASCADE;
DROP FUNCTION IF EXISTS fn_order_items_bi_fill_price() CASCADE;
DROP FUNCTION IF EXISTS fn_order_items_ai_stock_and_total() CASCADE;
DROP FUNCTION IF EXISTS fn_order_items_au_stock_and_total() CASCADE;
DROP FUNCTION IF EXISTS fn_order_items_ad_stock_and_total() CASCADE;

-- tables
DROP TABLE IF EXISTS audit_log CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- users table
CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL,
  full_name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL DEFAULT NULL
);

-- add unique constraint for email
ALTER TABLE users ADD CONSTRAINT ux_users_email UNIQUE (email);

-- categories table
CREATE TABLE categories (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  description VARCHAR(500) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- add unique constraint for category name
ALTER TABLE categories ADD CONSTRAINT ux_categories_name UNIQUE (name);

-- products table
CREATE TABLE products (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(200) NOT NULL,
  category_id BIGINT NOT NULL,
  price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
  stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- add foreign key and index for products
ALTER TABLE products ADD CONSTRAINT fk_products_category 
  FOREIGN KEY (category_id) REFERENCES categories(id) ON UPDATE CASCADE ON DELETE RESTRICT;
CREATE INDEX ix_products_category_id ON products(category_id);

-- orders table
CREATE TABLE orders (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','paid','shipped','cancelled')),
  total_amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL DEFAULT NULL
);

-- add foreign key and index for orders
ALTER TABLE orders ADD CONSTRAINT fk_orders_user 
  FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT;
CREATE INDEX ix_orders_user_id ON orders(user_id);

-- order_items table
CREATE TABLE order_items (
  id BIGSERIAL PRIMARY KEY,
  order_id BIGINT NOT NULL,
  product_id BIGINT NOT NULL,
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  unit_price DECIMAL(10,2) NULL
);

-- add foreign keys and indexes for order_items
ALTER TABLE order_items ADD CONSTRAINT fk_order_items_order 
  FOREIGN KEY (order_id) REFERENCES orders(id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE order_items ADD CONSTRAINT fk_order_items_product 
  FOREIGN KEY (product_id) REFERENCES products(id) ON UPDATE CASCADE ON DELETE RESTRICT;
CREATE INDEX ix_order_items_order_id ON order_items(order_id);
CREATE INDEX ix_order_items_product_id ON order_items(product_id);

-- audit log table
CREATE TABLE audit_log (
  id BIGSERIAL PRIMARY KEY,
  entity_type VARCHAR(100) NOT NULL,
  entity_id BIGINT NOT NULL,
  action VARCHAR(50) NOT NULL,
  details JSONB NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- add index for audit log
CREATE INDEX ix_audit_log_entity ON audit_log(entity_type, entity_id);

-- seed reference data for categories and a few products
INSERT INTO categories (name, description) VALUES
  ('electronics', 'consumer electronics and gadgets'),
  ('books', 'fiction and non-fiction books'),
  ('home', 'home and kitchen supplies'),
  ('sports', 'sports and outdoors'),
  ('toys', 'toys and games'),
  ('beauty', 'beauty and personal care'),
  ('grocery', 'food and beverages'),
  ('clothing', 'apparel and accessories'),
  ('automotive', 'car accessories and parts'),
  ('pets', 'pet supplies');

INSERT INTO products (name, category_id, price, stock) VALUES
  ('usb-c cable', 1, 9.99, 1000),
  ('wireless mouse', 1, 24.99, 500),
  ('notebook "productivity 101"', 2, 14.50, 300),
  ('chef knife', 3, 39.95, 200),
  ('yoga mat', 4, 19.99, 400),
  ('board game classic', 5, 29.99, 250),
  ('shampoo 500ml', 6, 8.49, 600),
  ('organic almonds 1kg', 7, 12.99, 350),
  ('t-shirt cotton', 8, 11.99, 800),
  ('car phone holder', 9, 15.99, 450),
  ('dog leash', 10, 13.49, 300);

-- create trigger functions
CREATE OR REPLACE FUNCTION fn_users_bi_normalize_email()
RETURNS TRIGGER AS $$
BEGIN
  NEW.email := LOWER(TRIM(NEW.email));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_order_items_bi_fill_price()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.unit_price IS NULL THEN
    SELECT price INTO NEW.unit_price FROM products WHERE id = NEW.product_id;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_order_items_ai_stock_and_total()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE products SET stock = stock - NEW.quantity WHERE id = NEW.product_id;
  
  UPDATE orders o
  SET total_amount = (
    SELECT COALESCE(SUM(oi.quantity * oi.unit_price), 0)
    FROM order_items oi
    WHERE oi.order_id = NEW.order_id
  )
  WHERE o.id = NEW.order_id;
  
  INSERT INTO audit_log(entity_type, entity_id, action, details)
  VALUES ('order', NEW.order_id, 'item_added', 
    jsonb_build_object('order_item_id', NEW.id, 'product_id', NEW.product_id, 
                      'quantity', NEW.quantity, 'unit_price', NEW.unit_price));
  
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_order_items_au_stock_and_total()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.product_id = OLD.product_id THEN
    UPDATE products SET stock = stock - (NEW.quantity - OLD.quantity) WHERE id = NEW.product_id;
  ELSE
    UPDATE products SET stock = stock + OLD.quantity WHERE id = OLD.product_id;
    UPDATE products SET stock = stock - NEW.quantity WHERE id = NEW.product_id;
  END IF;
  
  UPDATE orders o
  SET total_amount = (
    SELECT COALESCE(SUM(oi.quantity * oi.unit_price), 0)
    FROM order_items oi
    WHERE oi.order_id = NEW.order_id
  )
  WHERE o.id = NEW.order_id;
  
  INSERT INTO audit_log(entity_type, entity_id, action, details)
  VALUES ('order', NEW.order_id, 'item_updated', 
    jsonb_build_object('order_item_id', NEW.id, 'product_id', NEW.product_id, 
                      'quantity', NEW.quantity, 'unit_price', NEW.unit_price));
  
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_order_items_ad_stock_and_total()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE products SET stock = stock + OLD.quantity WHERE id = OLD.product_id;
  
  UPDATE orders o
  SET total_amount = (
    SELECT COALESCE(SUM(oi.quantity * oi.unit_price), 0)
    FROM order_items oi
    WHERE oi.order_id = OLD.order_id
  )
  WHERE o.id = OLD.order_id;
  
  INSERT INTO audit_log(entity_type, entity_id, action, details)
  VALUES ('order', OLD.order_id, 'item_deleted', 
    jsonb_build_object('order_item_id', OLD.id, 'product_id', OLD.product_id, 
                      'quantity', OLD.quantity, 'unit_price', OLD.unit_price));
  
  RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- create triggers
CREATE TRIGGER trg_users_bi_normalize_email
  BEFORE INSERT ON users
  FOR EACH ROW
  EXECUTE FUNCTION fn_users_bi_normalize_email();

CREATE TRIGGER trg_order_items_bi_fill_price
  BEFORE INSERT ON order_items
  FOR EACH ROW
  EXECUTE FUNCTION fn_order_items_bi_fill_price();

CREATE TRIGGER trg_order_items_ai_stock_and_total
  AFTER INSERT ON order_items
  FOR EACH ROW
  EXECUTE FUNCTION fn_order_items_ai_stock_and_total();

CREATE TRIGGER trg_order_items_au_stock_and_total
  AFTER UPDATE ON order_items
  FOR EACH ROW
  EXECUTE FUNCTION fn_order_items_au_stock_and_total();

CREATE TRIGGER trg_order_items_ad_stock_and_total
  AFTER DELETE ON order_items
  FOR EACH ROW
  EXECUTE FUNCTION fn_order_items_ad_stock_and_total();

-- create stored procedures (PostgreSQL functions)
CREATE OR REPLACE FUNCTION sp_generate_users(target_count INTEGER)
RETURNS VOID AS $$
DECLARE
  i INTEGER := 1;
BEGIN
  WHILE i <= target_count LOOP
    INSERT INTO users (email, full_name, created_at)
    VALUES (
      'user' || i || '@example.com',
      'User ' || i,
      CURRENT_TIMESTAMP - (RANDOM() * INTERVAL '730 days')
    );
    i := i + 1;
  END LOOP;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION sp_seed_catalog(additional_products INTEGER)
RETURNS VOID AS $$
DECLARE
  i INTEGER := 1;
  cat_count INTEGER;
BEGIN
  SELECT COUNT(*) INTO cat_count FROM categories;
  
  WHILE i <= additional_products LOOP
    INSERT INTO products (name, category_id, price, stock)
    VALUES (
      'generic product ' || i,
      1 + FLOOR(RANDOM() * cat_count)::INTEGER,
      ROUND((5 + (RANDOM() * 195))::NUMERIC, 2),
      100 + FLOOR(RANDOM() * 900)::INTEGER
    );
    i := i + 1;
  END LOOP;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION sp_generate_orders(num_orders INTEGER, max_items_per_order INTEGER)
RETURNS VOID AS $$
DECLARE
  i INTEGER := 1;
  j INTEGER;
  user_max BIGINT;
  product_max BIGINT;
  rnd_user BIGINT;
  rnd_product BIGINT;
  rnd_items INTEGER;
  rnd_qty INTEGER;
  new_order_id BIGINT;
  statuses TEXT[] := ARRAY['pending', 'paid', 'shipped', 'cancelled'];
BEGIN
  SELECT MAX(id) INTO user_max FROM users;
  SELECT MAX(id) INTO product_max FROM products;

  WHILE i <= num_orders LOOP
    rnd_user := 1 + FLOOR(RANDOM() * user_max)::BIGINT;
    
    INSERT INTO orders (user_id, status, total_amount, created_at)
    VALUES (rnd_user,
      statuses[1 + FLOOR(RANDOM() * 4)::INTEGER],
      0.00,
      CURRENT_TIMESTAMP - (RANDOM() * INTERVAL '365 days')
    );
    
    new_order_id := LASTVAL();

    rnd_items := 1 + FLOOR(RANDOM() * GREATEST(max_items_per_order, 1))::INTEGER;
    j := 1;
    WHILE j <= rnd_items LOOP
      rnd_product := 1 + FLOOR(RANDOM() * product_max)::BIGINT;
      rnd_qty := 1 + FLOOR(RANDOM() * 5)::INTEGER;
      
      INSERT INTO order_items (order_id, product_id, quantity, unit_price)
      VALUES (new_order_id, rnd_product, rnd_qty, NULL);
      
      j := j + 1;
    END LOOP;

    i := i + 1;
  END LOOP;
END;
$$ LANGUAGE plpgsql;

-- create views
CREATE OR REPLACE VIEW v_user_order_stats AS
SELECT
  u.id AS user_id,
  u.email,
  u.full_name,
  COUNT(o.id) AS order_count,
  SUM(CASE WHEN o.status <> 'cancelled' THEN o.total_amount ELSE 0 END) AS lifetime_value,
  MIN(o.created_at) AS first_order_at,
  MAX(o.created_at) AS last_order_at
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
GROUP BY u.id, u.email, u.full_name;

CREATE OR REPLACE VIEW v_order_items_detailed AS
SELECT
  o.id AS order_id,
  o.created_at AS order_created_at,
  o.status,
  u.email AS user_email,
  p.name AS product_name,
  c.name AS category_name,
  oi.quantity,
  oi.unit_price,
  (oi.quantity * oi.unit_price) AS line_total
FROM orders o
JOIN users u ON u.id = o.user_id
JOIN order_items oi ON oi.order_id = o.id
JOIN products p ON p.id = oi.product_id
JOIN categories c ON c.id = p.category_id;

-- create scheduled job for cleanup (PostgreSQL uses pg_cron extension)
-- Note: This requires pg_cron extension to be installed
-- CREATE EXTENSION IF NOT EXISTS pg_cron;
-- SELECT cron.schedule('cleanup-audit-log', '0 2 * * *', 'DELETE FROM audit_log WHERE created_at < NOW() - INTERVAL ''180 days''');

-- seed data: 1500 users to exceed 1000+ rows requirement, a larger catalog, and 500 orders
SELECT sp_generate_users(1500);
SELECT sp_seed_catalog(200);
SELECT sp_generate_orders(500, 5);

-- create updated_at trigger function
CREATE OR REPLACE FUNCTION fn_update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at := CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- add updated_at triggers
CREATE TRIGGER trg_users_update_updated_at
  BEFORE UPDATE ON users
  FOR EACH ROW
  EXECUTE FUNCTION fn_update_updated_at();

CREATE TRIGGER trg_orders_update_updated_at
  BEFORE UPDATE ON orders
  FOR EACH ROW
  EXECUTE FUNCTION fn_update_updated_at();
