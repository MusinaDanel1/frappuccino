-- Create the necessary ENUM types
CREATE TYPE order_status AS ENUM('pending', 'processing', 'completed', 'cancelled');
CREATE TYPE unit_of_measurement AS ENUM('mg', 'g', 'kg', 'oz', 'lb', 'ml', 'l', 'dl', 'fl', 'pc', 'dozen', 'cup', 'tsp', 'tbsp', 'shots');  -- Allow Pastry for non-standard units
CREATE TYPE type_of_transaction AS ENUM('addition', 'deduction');

-- Create Orders table
CREATE TABLE orders(
    order_id SERIAL PRIMARY KEY,
    customer_name VARCHAR(255) NOT NULL,
    special_instructions TEXT[],
    total_amount NUMERIC DEFAULT 0 CHECK(total_amount >= 0),
    status order_status DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create Menu Items table
CREATE TABLE menu_items(
    menu_item_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price NUMERIC NOT NULL CHECK(price > 0),
    categories TEXT[],
    allergens TEXT[],
    metadata JSONB
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
    price_at_order NUMERIC NOT NULL CHECK(price_at_order > 0),
    customization_options JSONB
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
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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

-- Create Indexes
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_inventory_name ON inventory(name);
CREATE INDEX idx_menu_item_ingredients ON menu_item_ingredients(menu_item_id, inventory_id);
CREATE INDEX idx_menu_items_fulltext ON menu_items USING gin (to_tsvector('english', name));

-- Insert sample data into inventory
INSERT INTO inventory(name, quantity, unit, last_updated) VALUES
    ('Coffee Beans', 5500, 'g', '2024-01-01'),
    ('Muffin', 300, 'pc', '2024-01-01'),  
    ('Milk', 1000, 'l', '2024-01-01'),
    ('Sugar', 1000, 'g', '2024-01-01'),
    ('Flour', 10000, 'g', '2024-01-01'),
    ('Eggs', 500, 'pc', '2024-01-01'),
    ('Chocolate', 3000, 'g', '2024-01-01'),
    ('Vanilla Extract', 200, 'ml', '2024-01-01'),
    ('Butter', 2000, 'g', '2024-01-01'),
    ('Cinnamon', 500, 'g', '2024-01-01'),
    ('Salt', 1000, 'g', '2024-01-01'),
    ('Bananas', 150, 'pc', '2024-01-01'),
    ('Strawberries', 300, 'pc', '2024-01-01'),
    ('Blueberries', 200, 'pc', '2024-01-01'),
    ('Lemons', 100, 'pc', '2024-01-01'),
    ('Oranges', 120, 'pc', '2024-01-01'),
    ('Yeast', 100, 'g', '2024-01-01'),
    ('Baking Powder', 300, 'g', '2024-01-01'),
    ('Cream', 100, 'l', '2024-01-01'),
    ('Honey', 500, 'g', '2024-01-01'),
    ('Whipped Cream', 50, 'l', '2024-01-01');

-- Insert sample data into menu_items
INSERT INTO menu_items(name, price, categories) VALUES 
    ('Espresso', 2.5, ARRAY['Beverage']),
    ('Latte', 3.5, ARRAY['Beverage']),
    ('Cappuccino', 3.0, ARRAY['Beverage']),
    ('Americano', 2.0, ARRAY['Beverage']),
    ('Croissant', 2.5, ARRAY['Pastry']),
    ('Muffin', 3.0, ARRAY['Pastry']),
    ('Cheesecake', 4.5, ARRAY['Dessert']),
    ('Brownie', 3.5, ARRAY['Dessert']),
    ('Bagel', 2.0, ARRAY['Pastry']),
    ('Pancake', 3.5, ARRAY['Breakfast']);

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
INSERT INTO orders(customer_name, total_amount, status, created_at) VALUES
    ('Alice', 15.5, 'completed', '2024-12-01'),
    ('Bob', 12.0, 'pending', '2024-12-02'),
    ('Charlie', 18.0, 'cancelled', '2024-12-03'),
    ('Diana', 20.5, 'completed', '2024-12-04'),
    ('Eve', 22.0, 'completed', '2024-12-05');

-- Insert sample data into order_items
INSERT INTO order_items(order_id, menu_item_id, quantity, price_at_order) VALUES
    (1, 1, 2, 2.5),
    (1, 2, 1, 3.5),
    (2, 3, 1, 3.0),
    (3, 5, 2, 2.5),
    (4, 10, 3, 3.5);

-- Insert sample data into order_status_history
INSERT INTO order_status_history(order_id, status, changed_at) VALUES
    (1, 'pending', '2024-12-01'),
    (1, 'completed', '2024-12-01'),
    (2, 'pending', '2024-12-02'),
    (2, 'completed', '2024-12-02'),
    (3, 'pending', '2024-12-03'),
    (3, 'cancelled', '2024-12-03');

-- Create a trigger to ensure valid status transitions (pending -> processing -> completed/cancelled)
CREATE OR REPLACE FUNCTION validate_order_status_change() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'processing' AND OLD.status = 'completed' THEN
        RAISE EXCEPTION 'Invalid status transition from completed to processing';
    ELSIF NEW.status = 'completed' AND OLD.status != 'processing' THEN
        RAISE EXCEPTION 'Order must be in processing before it can be completed';
    ELSIF NEW.status = 'cancelled' AND OLD.status != 'pending' THEN
        RAISE EXCEPTION 'Order can only be cancelled if it is still pending';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger on orders table to call this function
CREATE TRIGGER validate_order_status
BEFORE UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION validate_order_status_change();
