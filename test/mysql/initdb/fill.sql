-- set up character set and safe modes
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

USE appdb;

-- force utc session time zone to avoid dst gaps on insert into TIMESTAMP columns
SET time_zone = '+00:00';

-- ensure event scheduler is on (may require privileges)
SET @prev_event_scheduler = @@GLOBAL.event_scheduler;
SET GLOBAL event_scheduler = ON;

-- drop objects if they exist
DROP EVENT IF EXISTS ev_cleanup_audit_log;
DROP VIEW IF EXISTS v_user_order_stats;
DROP VIEW IF EXISTS v_order_items_detailed;
DROP TRIGGER IF EXISTS trg_users_bi_normalize_email;
DROP TRIGGER IF EXISTS trg_order_items_bi_fill_price;
DROP TRIGGER IF EXISTS trg_order_items_ai_stock_and_total;
DROP TRIGGER IF EXISTS trg_order_items_au_stock_and_total;
DROP TRIGGER IF EXISTS trg_order_items_ad_stock_and_total;
DROP PROCEDURE IF EXISTS sp_generate_users;
DROP PROCEDURE IF EXISTS sp_seed_catalog;
DROP PROCEDURE IF EXISTS sp_generate_orders;

-- tables
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;

-- users table
CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  email VARCHAR(255) NOT NULL,
  full_name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY ux_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- categories table
CREATE TABLE categories (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(100) NOT NULL,
  description VARCHAR(500) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY ux_categories_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- products table
CREATE TABLE products (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(200) NOT NULL,
  category_id BIGINT UNSIGNED NOT NULL,
  price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
  stock INT NOT NULL DEFAULT 0 CHECK (stock >= 0),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY ix_products_category_id (category_id),
  CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories(id)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- orders table
CREATE TABLE orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  status ENUM('pending','paid','shipped','cancelled') NOT NULL DEFAULT 'pending',
  total_amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY ix_orders_user_id (user_id),
  CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- order_items table
CREATE TABLE order_items (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_id BIGINT UNSIGNED NOT NULL,
  product_id BIGINT UNSIGNED NOT NULL,
  quantity INT NOT NULL CHECK (quantity > 0),
  unit_price DECIMAL(10,2) NULL,
  PRIMARY KEY (id),
  KEY ix_order_items_order_id (order_id),
  KEY ix_order_items_product_id (product_id),
  CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders(id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT fk_order_items_product FOREIGN KEY (product_id) REFERENCES products(id)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- audit log table
CREATE TABLE audit_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  entity_type VARCHAR(100) NOT NULL,
  entity_id BIGINT UNSIGNED NOT NULL,
  action VARCHAR(50) NOT NULL,
  details JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY ix_audit_log_entity (entity_type, entity_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- seed reference data for categories and a few products (additional products will be added by proc)
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

-- triggers
DELIMITER $$

-- normalize user email before insert
CREATE TRIGGER trg_users_bi_normalize_email
BEFORE INSERT ON users
FOR EACH ROW
BEGIN
  SET NEW.email = LOWER(TRIM(NEW.email));
END $$

-- fill unit_price from products if not provided
CREATE TRIGGER trg_order_items_bi_fill_price
BEFORE INSERT ON order_items
FOR EACH ROW
BEGIN
  IF NEW.unit_price IS NULL THEN
    SET NEW.unit_price = (SELECT p.price FROM products p WHERE p.id = NEW.product_id);
  END IF;
END $$

-- maintain stock and order totals after insert
CREATE TRIGGER trg_order_items_ai_stock_and_total
AFTER INSERT ON order_items
FOR EACH ROW
BEGIN
  UPDATE products SET stock = stock - NEW.quantity WHERE id = NEW.product_id;
  UPDATE orders o
  SET o.total_amount = (
    SELECT IFNULL(SUM(oi.quantity * oi.unit_price), 0)
    FROM order_items oi
    WHERE oi.order_id = NEW.order_id
  )
  WHERE o.id = NEW.order_id;
  INSERT INTO audit_log(entity_type, entity_id, action, details)
  VALUES ('order', NEW.order_id, 'item_added', JSON_OBJECT('order_item_id', NEW.id, 'product_id', NEW.product_id, 'quantity', NEW.quantity, 'unit_price', NEW.unit_price));
END $$

-- maintain stock and order totals after update
CREATE TRIGGER trg_order_items_au_stock_and_total
AFTER UPDATE ON order_items
FOR EACH ROW
BEGIN
  IF NEW.product_id = OLD.product_id THEN
    UPDATE products SET stock = stock - (NEW.quantity - OLD.quantity) WHERE id = NEW.product_id;
  ELSE
    UPDATE products SET stock = stock + OLD.quantity WHERE id = OLD.product_id;
    UPDATE products SET stock = stock - NEW.quantity WHERE id = NEW.product_id;
  END IF;
  UPDATE orders o
  SET o.total_amount = (
    SELECT IFNULL(SUM(oi.quantity * oi.unit_price), 0)
    FROM order_items oi
    WHERE oi.order_id = NEW.order_id
  )
  WHERE o.id = NEW.order_id;
  INSERT INTO audit_log(entity_type, entity_id, action, details)
  VALUES ('order', NEW.order_id, 'item_updated', JSON_OBJECT('order_item_id', NEW.id, 'product_id', NEW.product_id, 'quantity', NEW.quantity, 'unit_price', NEW.unit_price));
END $$

-- maintain stock and order totals after delete
CREATE TRIGGER trg_order_items_ad_stock_and_total
AFTER DELETE ON order_items
FOR EACH ROW
BEGIN
  UPDATE products SET stock = stock + OLD.quantity WHERE id = OLD.product_id;
  UPDATE orders o
  SET o.total_amount = (
    SELECT IFNULL(SUM(oi.quantity * oi.unit_price), 0)
    FROM order_items oi
    WHERE oi.order_id = OLD.order_id
  )
  WHERE o.id = OLD.order_id;
  INSERT INTO audit_log(entity_type, entity_id, action, details)
  VALUES ('order', OLD.order_id, 'item_deleted', JSON_OBJECT('order_item_id', OLD.id, 'product_id', OLD.product_id, 'quantity', OLD.quantity, 'unit_price', OLD.unit_price));
END $$

DELIMITER ;

-- procedures
DELIMITER $$

-- generate many users to satisfy the 1000+ rows requirement
CREATE PROCEDURE sp_generate_users(IN target_count INT)
BEGIN
  DECLARE i INT DEFAULT 1;
  WHILE i <= target_count DO
    INSERT INTO users (email, full_name, created_at)
    VALUES (
      CONCAT('user', i, '@example.com'),
      CONCAT('User ', i),
      TIMESTAMP(DATE_SUB(CURDATE(), INTERVAL FLOOR(RAND() * 730) DAY), SEC_TO_TIME(FLOOR(RAND() * 86400)))
    );
    SET i = i + 1;
  END WHILE;
END $$

-- add more products to diversify catalog
CREATE PROCEDURE sp_seed_catalog(IN additional_products INT)
BEGIN
  DECLARE i INT DEFAULT 1;
  DECLARE cat_count INT DEFAULT (SELECT COUNT(*) FROM categories);
  WHILE i <= additional_products DO
    INSERT INTO products (name, category_id, price, stock)
    VALUES (
      CONCAT('generic product ', i),
      1 + FLOOR(RAND() * cat_count),
      ROUND(5 + (RAND() * 195), 2),
      100 + FLOOR(RAND() * 900)
    );
    SET i = i + 1;
  END WHILE;
END $$

-- create random orders with items
CREATE PROCEDURE sp_generate_orders(IN num_orders INT, IN max_items_per_order INT)
BEGIN
  DECLARE i INT DEFAULT 1;
  DECLARE j INT;
  DECLARE user_max BIGINT UNSIGNED;
  DECLARE product_max BIGINT UNSIGNED;
  DECLARE rnd_user BIGINT UNSIGNED;
  DECLARE rnd_product BIGINT UNSIGNED;
  DECLARE rnd_items INT;
  DECLARE rnd_qty INT;
  DECLARE new_order_id BIGINT UNSIGNED;
  SELECT MAX(id) INTO user_max FROM users;
  SELECT MAX(id) INTO product_max FROM products;

  WHILE i <= num_orders DO
    SET rnd_user = 1 + FLOOR(RAND() * user_max);
    INSERT INTO orders (user_id, status, total_amount, created_at)
    VALUES (rnd_user,
      ELT(1 + FLOOR(RAND()*4), 'pending','paid','shipped','cancelled'),
      0.00,
      TIMESTAMP(DATE_SUB(CURDATE(), INTERVAL FLOOR(RAND() * 365) DAY), SEC_TO_TIME(FLOOR(RAND() * 86400)))
    );
    SET new_order_id = LAST_INSERT_ID();

    SET rnd_items = 1 + FLOOR(RAND() * GREATEST(max_items_per_order, 1));
    SET j = 1;
    WHILE j <= rnd_items DO
      SET rnd_product = 1 + FLOOR(RAND() * product_max);
      SET rnd_qty = 1 + FLOOR(RAND() * 5);
      INSERT INTO order_items (order_id, product_id, quantity, unit_price)
      VALUES (new_order_id, rnd_product, rnd_qty, NULL);
      SET j = j + 1;
    END WHILE;

    SET i = i + 1;
  END WHILE;
END $$

DELIMITER ;

-- views
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

-- scheduled event to clean up old audit log entries
CREATE EVENT IF NOT EXISTS ev_cleanup_audit_log
ON SCHEDULE EVERY 1 DAY STARTS CURRENT_TIMESTAMP + INTERVAL 1 DAY
DO
  DELETE FROM audit_log WHERE created_at < NOW() - INTERVAL 180 DAY;

-- seed data: 1500 users to exceed 1000+ rows requirement, a larger catalog, and 500 orders
CALL sp_generate_users(1500);
CALL sp_seed_catalog(200);
CALL sp_generate_orders(500, 5);

-- restore FK checks
SET FOREIGN_KEY_CHECKS = 1;

-- optional: restore previous event scheduler value
SET GLOBAL event_scheduler = @prev_event_scheduler;


