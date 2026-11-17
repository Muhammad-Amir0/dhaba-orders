CREATE TABLE menu_items (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    price_pkr INTEGER NOT NULL CHECK (price_pkr > 0),
    category TEXT NOT NULL
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    items JSONB NOT NULL,
    total_pkr INTEGER NOT NULL CHECK (total_pkr >= 0),
    status TEXT NOT NULL DEFAULT 'new'
        CHECK (status IN ('new', 'cooking', 'ready')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO menu_items (name, price_pkr, category) VALUES
    ('Chicken Karahi', 1450, 'Karahi'),
    ('Mutton Karahi', 2200, 'Karahi'),
    ('Chicken Biryani', 550, 'Rice'),
    ('Beef Biryani', 650, 'Rice'),
    ('Plain Naan', 70, 'Bread'),
    ('Garlic Naan', 130, 'Bread'),
    ('Seekh Kebab', 480, 'BBQ'),
    ('Chapli Kebab', 420, 'BBQ'),
    ('Chicken Tikka', 600, 'BBQ'),
    ('Daal Makhani', 500, 'Vegetarian'),
    ('Mango Lassi', 220, 'Drinks'),
    ('Sweet Lassi', 180, 'Drinks');
