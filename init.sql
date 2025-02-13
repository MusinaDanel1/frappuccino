-- Create the necessary ENUM types
CREATE TYPE order_status AS ENUM('accepted','pending', 'processing', 'completed', 'cancelled','rejected');
CREATE TYPE unit_of_measurement AS ENUM('mg', 'g', 'kg', 'oz', 'lb', 'ml', 'l', 'dl', 'fl', 'pc', 'dozen', 'cup', 'tsp', 'tbsp', 'shots'); 
CREATE TYPE type_of_transaction AS ENUM('addition', 'deduction');

-- Create Orders table
CREATE TABLE orders(
    order_id SERIAL PRIMARY KEY,
    customer_name VARCHAR(255) NOT NULL,
    special_instructions JSONB,
    total_amount NUMERIC DEFAULT 0 CHECK(total_amount >= 0),
    status order_status DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create Menu Items table
CREATE TABLE menu_items(
    menu_item_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    price NUMERIC NOT NULL CHECK(price > 0),
    categories TEXT[],
    allergens TEXT[]
);

-- Create Inventory table
CREATE TABLE inventory(
    ingredient_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    quantity NUMERIC NOT NULL CHECK(quantity >= 0),
    unit unit_of_measurement NOT NULL,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create Order Items table
CREATE TABLE order_items(
    order_item_id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(order_id) ON DELETE CASCADE,
    menu_item_id INT REFERENCES menu_items(menu_item_id) ON DELETE CASCADE,
    quantity INT NOT NULL CHECK(quantity > 0),
    price_at_order NUMERIC NOT NULL CHECK(price_at_order > 0)
);

-- Create Menu Item Ingredients table
CREATE TABLE menu_item_ingredients(
    menu_item_ingredient_id SERIAL PRIMARY KEY,
    menu_item_id INT REFERENCES menu_items(menu_item_id) ON DELETE CASCADE,
    inventory_id INT REFERENCES inventory(ingredient_id) ON DELETE CASCADE,
    quantity NUMERIC NOT NULL CHECK(quantity > 0)
);

-- Create Inventory Transactions table
CREATE TABLE inventory_transactions(
    transaction_id SERIAL PRIMARY KEY,
    ingredient_id INT REFERENCES inventory(ingredient_id) ON DELETE CASCADE,
    quantity_change NUMERIC NOT NULL,
    transaction_type type_of_transaction NOT NULL,
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create Order Status History table
CREATE TABLE order_status_history(
    status_id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(order_id) ON DELETE CASCADE,
    status order_status NOT NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create Price History table
CREATE TABLE price_history(
    price_id SERIAL PRIMARY KEY,
    menu_item_id INT REFERENCES menu_items(menu_item_id) ON DELETE CASCADE,
    old_price NUMERIC NOT NULL,
    new_price NUMERIC NOT NULL CHECK(new_price > 0),
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Insert sample data into inventory
INSERT INTO inventory(name, quantity, unit, last_updated) VALUES
    ('Coffee Beans', 5500, 'g', '2024-01-01 08:00:00'),
    ('Muffin', 3000, 'pc', '2024-01-01 09:00:00'),  
    ('Milk', 1000, 'l', '2024-01-01 09:30:00'),
    ('Sugar', 1000, 'g', '2024-01-01 10:00:00'),
    ('Flour', 10000, 'g', '2024-01-01 11:00:00'),
    ('Eggs', 500, 'pc', '2024-01-01 12:00:00'),
    ('Chocolate', 3000, 'g', '2024-01-01 13:00:00'),
    ('Vanilla Extract', 200, 'ml', '2024-01-01 14:00:00'),
    ('Butter', 2000, 'g', '2024-01-01 15:00:00'),
    ('Cinnamon', 500, 'g', '2024-01-01 16:00:00'),
    ('Salt', 1000, 'g', '2024-01-01 17:00:00'),
    ('Bananas', 150, 'pc', '2024-01-01 18:00:00'),
    ('Strawberries', 300, 'pc', '2024-01-01 19:00:00'),
    ('Blueberries', 200, 'pc', '2024-01-01 20:00:00'),
    ('Lemons', 100, 'pc', '2024-01-01 21:00:00'),
    ('Oranges', 120, 'pc', '2024-01-01 22:00:00'),
    ('Yeast', 100, 'g', '2024-01-01 23:00:00'),
    ('Baking Powder', 300, 'g', '2024-01-01 00:00:00'),
    ('Cream', 100, 'l', '2024-01-01 01:00:00'),
    ('Honey', 500, 'g', '2024-01-01 02:00:00'),
    ('Whipped Cream', 50, 'l', '2024-01-01 03:00:00');

-- Insert sample data into menu_items
INSERT INTO menu_items(name, description, price, categories, allergens) VALUES
    ('Espresso', 'Strong black coffee', 2.5, ARRAY['Beverage'], ARRAY['None']),
    ('Latte', 'Coffee with steamed milk', 3.5, ARRAY['Beverage'], ARRAY['Dairy']),
    ('Cappuccino', 'Espresso with milk foam', 3.0, ARRAY['Beverage'], ARRAY['Dairy']),
    ('Americano', 'Espresso with hot water', 2.0, ARRAY['Beverage'], ARRAY['None']),
    ('Croissant', 'Flaky pastry', 2.5, ARRAY['Pastry'], ARRAY['Gluten', 'Dairy']),
    ('Muffin', 'Sweet muffin', 3.0, ARRAY['Pastry'], ARRAY['Gluten', 'Dairy']),
    ('Cheesecake', 'Creamy cheesecake', 4.5, ARRAY['Dessert'], ARRAY['Dairy']),
    ('Brownie', 'Chocolate brownie', 3.5, ARRAY['Dessert'], ARRAY['Gluten', 'Dairy']),
    ('Bagel', 'Toasted bagel', 2.0, ARRAY['Pastry'], ARRAY['Gluten']),
    ('Pancake', 'Fluffy pancake', 3.5, ARRAY['Breakfast'], ARRAY['Gluten', 'Dairy']);

-- Insert sample data into menu_item_ingredients
INSERT INTO menu_item_ingredients(menu_item_id, inventory_id, quantity) VALUES
    (1, 1, 50),
    (2, 1, 50),
    (2, 2, 150),
    (3, 1, 50),
    (3, 2, 100),
    (4, 1, 50),
    (5, 4, 100),
    (5, 8, 50),
    (6, 4, 50),
    (6, 6, 20),
    (7, 6, 100),
    (7, 18, 50),
    (8, 6, 50),
    (8, 4, 50),
    (9, 4, 150),
    (9, 16, 5),
    (10, 4, 200),
    (10, 18, 100);

-- Insert sample data into orders
INSERT INTO orders (customer_name, special_instructions, total_amount, status, created_at, updated_at) VALUES
    ('Alice', '["No sugar", "Extra shot"]'::jsonb, 15.5, 'completed', '2024-12-01 10:00:00', '2024-12-01 10:15:00'),
    ('Bob', '["No dairy"]'::jsonb, 12.0, 'pending', '2024-12-02 12:00:00', '2024-12-02 12:05:00'),
    ('Charlie', '["Gluten free"]'::jsonb, 18.0, 'cancelled', '2024-12-03 14:00:00', '2024-12-03 14:10:00'),
    ('Diana', '["Low fat milk"]'::jsonb, 20.5, 'completed', '2024-12-04 16:00:00', '2024-12-04 16:10:00'),
    ('Eve', '["Extra vanilla", "No foam"]'::jsonb, 22.0, 'completed', '2024-12-05 08:00:00', '2024-12-05 08:10:00');

-- Insert sample data into order_items
INSERT INTO order_items(order_id, menu_item_id, quantity, price_at_order) VALUES
    (1, 1, 2, 2.5),
    (1, 2, 1, 3.5),
    (2, 3, 1, 3.0),
    (3, 5, 2, 2.5),
    (4, 10, 3, 3.5);

-- Insert sample data into order_status_history
INSERT INTO order_status_history(order_id, status, changed_at) VALUES
    (1, 'pending', '2024-12-01 10:00:00'),
    (1, 'completed', '2024-12-01 10:15:00'),
    (2, 'pending', '2024-12-02 12:00:00'),
    (2, 'completed', '2024-12-02 12:05:00'),
    (3, 'pending', '2024-12-03 14:00:00'),
    (3, 'cancelled', '2024-12-03 14:10:00');

-- Insert sampledata into price_history
INSERT INTO price_history(menu_item_id, old_price, new_price, changed_at) VALUES
    (1, 2.0, 2.5, '2024-12-01 09:00:00'),
    (2, 3.0, 3.5, '2024-12-02 11:00:00'),
    (3, 2.5, 3.0, '2024-12-03 13:00:00'),
    (4, 1.8, 2.0, '2024-12-04 15:00:00'),
    (5, 2.0, 2.5, '2024-12-05 17:00:00');

-- Insert sample data into inventory_transactions
INSERT INTO inventory_transactions (ingredient_id, quantity_change, transaction_type, transaction_date)
VALUES
(1, 10, 'deduction', '2025-02-12 10:05:00'),
(2, 500, 'addition', '2025-02-12 12:00:00'),
(3, 5, 'deduction', '2025-02-12 14:00:00'),
(1, 20, 'deduction', '2025-02-12 15:30:00'),
(4, 300, 'addition', '2025-02-12 16:00:00');

-- Index for fast inventory lookup by ID
CREATE INDEX idx_inventory_id ON inventory (ingredient_id);

-- Index for fast menu items lookup by ID
CREATE INDEX idx_menu_items_id ON menu_items (menu_item_id);

-- Index for fast orders lookup by ID
CREATE INDEX idx_orders_id ON orders (order_id);




