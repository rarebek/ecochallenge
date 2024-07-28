CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    avatar TEXT,
    birth_date TIMESTAMP,
    location TEXT,
    phone_number VARCHAR(20),
    xp BIGINT NOT NULL DEFAULT 0,
    region TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE events (
    id UUID PRIMARY KEY,
    image TEXT,
    name TEXT NOT NULL,
    description TEXT,
    total_xp BIGINT NOT NULL DEFAULT 0,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    resp_officer VARCHAR(255) NOT NULL,
    resp_officer_image TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE history (
    id UUID PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    event_id UUID NOT NULL REFERENCES events(id),
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    xp_earned BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS market(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    count BIGINT NOT NULL,
    xp BIGINT NOT NULL,
    category_name TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO market (name, description, count, xp, category_name) VALUES
('MacBook Pro', 'Apple MacBook Pro 16-inch', 50, 1000, 'Electronics'),
('iPhone 13', 'Apple iPhone 13 128GB', 200, 800, 'Electronics'),
('Samsung Galaxy S21', 'Samsung Galaxy S21 256GB', 150, 750, 'Electronics'),
('Dell XPS 13', 'Dell XPS 13 Laptop', 100, 900, 'Electronics'),
('Sony WH-1000XM4', 'Sony Noise Cancelling Headphones', 300, 400, 'Accessories'),
('Apple Watch Series 7', 'Apple Watch Series 7 45mm', 250, 500, 'Wearables'),
('iPad Pro', 'Apple iPad Pro 12.9-inch', 80, 950, 'Tablets'),
('Google Pixel 6', 'Google Pixel 6 128GB', 180, 700, 'Electronics'),
('Amazon Echo Dot', 'Amazon Echo Dot 4th Gen', 400, 200, 'Smart Home'),
('Nintendo Switch', 'Nintendo Switch Console', 120, 600, 'Gaming');



INSERT INTO users (id, first_name, last_name, avatar, birth_date, location, phone_number, xp)
VALUES
    (1, 'John', 'Doe', 'avatar1.png', '1990-01-01', 'New York', '1234567890', 100),
    (2, 'Jane', 'Smith', 'avatar2.png', '1992-02-02', 'Los Angeles', '0987654321', 200),
    (3, 'Alice', 'Johnson', 'avatar3.png', '1985-03-03', 'Chicago', '1122334455', 150),
    (4, 'Bob', 'Brown', 'avatar4.png', '1988-04-04', 'Houston', '5566778899', 120),
    (5, 'Charlie', 'Davis', 'avatar5.png', '1995-05-05', 'Phoenix', '6677889900', 80),
    (6, 'Daisy', 'Wilson', 'avatar6.png', '1991-06-06', 'Philadelphia', '3344556677', 90),
    (7, 'Ethan', 'Martinez', 'avatar7.png', '1993-07-07', 'San Antonio', '4455667788', 110),
    (8, 'Fiona', 'Garcia', 'avatar8.png', '1987-08-08', 'San Diego', '5566778899', 130),
    (9, 'George', 'Lee', 'avatar9.png', '1994-09-09', 'Dallas', '6677889900', 140),
    (10, 'Hannah', 'Walker', 'avatar10.png', '1996-10-10', 'San Jose', '7788990011', 70);

INSERT INTO events (id, name, description, total_xp, start_date, end_date, resp_officer, resp_officer_image)
VALUES
    ('e1a0f4a6-dc77-4e5a-bb56-f03339d8d4f5', 'Tree Planting', 'Community tree planting event', 50, '2024-08-01', '2024-08-01', 'Alice Green', 'alice_green.png'),
    ('f2a5e2b8-4134-4f7b-b44e-a0e8e01e90c6', 'Beach Clean-up', 'Clean-up event at the local beach', 70, '2024-08-15', '2024-08-15', 'Bob White', 'bob_white.png'),
    ('c3b2c5d9-527e-4e0b-9c3e-e28e98a5b849', 'Recycling Workshop', 'Workshop on recycling techniques', 30, '2024-09-01', '2024-09-01', 'Carol Black', 'carol_black.png'),
    ('d4c3d6ea-6389-4f8b-ac4f-14a2b1b7d90a', 'Wildlife Conservation', 'Awareness event on wildlife conservation', 60, '2024-09-15', '2024-09-15', 'David Gray', 'david_gray.png'),
    ('e5d4e7fb-748e-4f9c-8d5f-25b3c2c8e1a1', 'Energy Saving', 'Event on energy-saving methods', 40, '2024-10-01', '2024-10-01', 'Eva Blue', 'eva_blue.png'),
    ('f6e7f8a9-789b-4d9c-a2b3-d4e5f6a7b8c9', 'Water Conservation', 'Workshop on water-saving techniques', 45, '2024-10-15', '2024-10-15', 'Frank Orange', 'frank_orange.png'),
    ('a1b2c3d4-5678-9e0f-1a2b-3c4d5e6f7a8b', 'Sustainable Living', 'Seminar on sustainable living practices', 55, '2024-11-01', '2024-11-01', 'Grace Pink', 'grace_pink.png'),
    ('b2c3d4e5-6789-0f1a-2b3c-4d5e6f7a8b9c', 'Eco-Friendly Products', 'Expo of eco-friendly products', 65, '2024-11-15', '2024-11-15', 'Henry Cyan', 'henry_cyan.png'),
    ('c3d4e5f6-7890-1a2b-3c4d-5e6f7a8b9c0d', 'Renewable Energy', 'Conference on renewable energy sources', 75, '2024-12-01', '2024-12-01', 'Ivy Lavender', 'ivy_lavender.png'),
    ('d4e5f6a7-8901-2b3c-4d5e-6f7a8b9c0d1e', 'Climate Change Awareness', 'Campaign on climate change awareness', 80, '2024-12-15', '2024-12-15', 'Jack Indigo', 'jack_indigo.png');

INSERT INTO history (id, user_id, event_id, start_date, end_date, xp_earned)
VALUES
    ('c3d4e5f6-7890-1a2b-3c4d-5e6f7a8b9c0d', 1, 'e1a0f4a6-dc77-4e5a-bb56-f03339d8d4f5', '2024-08-01', '2024-08-01', 50),
    ('c3d4e5f6-7890-1a2b-3c4d-5e6f7a8b9c1d', 2, 'f2a5e2b8-4134-4f7b-b44e-a0e8e01e90c6', '2024-08-15', '2024-08-15', 70),
    ('c3d4e5f6-7890-1a2b-3c4d-5e6f7a8b9c2d', 3, 'c3b2c5d9-527e-4e0b-9c3e-e28e98a5b849', '2024-09-01', '2024-09-01', 30),
    ('c3d4e5f6-7890-1a2b-3c4d-5e6f7a8b9c4d', 4, 'd4c3d6ea-6389-4f8b-ac4f-14a2b1b7d90a', '2024-09-15', '2024-09-15', 60),
    ('c3d4e5f6-7890-1a2b-3c4d-5e6f7a8b9c5d', 5, 'e5d4e7fb-748e-4f9c-8d5f-25b3c2c8e1a1', '2024-10-01', '2024-10-01', 40),
    ('c3d4e5f6-7890-1a2b-3c4d-5e6f7a8b9c6d', 6, 'f6e7f8a9-789b-4d9c-a2b3-d4e5f6a7b8c9', '2024-10-15', '2024-10-15', 45),
    ('c3d4e5f6-7890-1a2b-3c4d-5e6f7a8b9c7d', 7, 'a1b2c3d4-5678-9e0f-1a2b-3c4d5e6f7a8b', '2024-11-01', '2024-11-01', 55),
    ('c3d4e5f6-7890-1a2b-3c4d-5e6f7a8b9c8d', 8, 'b2c3d4e5-6789-0f1a-2b3c-4d5e6f7a8b9c', '2024-11-15', '2024-11-15', 65),
    ('c3d4e5f6-7890-1a2b-3c4d-5e6f7a8b9c9d', 9, 'c3d4e5f6-7890-1a2b-3c4d-5e6f7a8b9c0d', '2024-12-01', '2024-12-01', 75),
    ('c3d4e5f6-7890-1a2b-3c4d-5e6f7a869c0d', 10, 'd4e5f6a7-8901-2b3c-4d5e-6f7a8b9c0d1e', '2024-12-15', '2024-12-15', 80);



