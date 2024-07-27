CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    avatar TEXT,
    birth_date TIMESTAMP,
    location TEXT,
    phone_number VARCHAR(20),
    xp BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE events (
    id UUID PRIMARY KEY,
    name TEXT,
    description TEXT,
    total_xp BIGINT,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE history (
    id UUID PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    event_id UUID REFERENCES events(id),
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    xp_given BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users (id, first_name, last_name, avatar, birth_date, location, phone_number, xp)
VALUES
    (1, 'John', 'Doe', 'avatar1.png', '1990-01-01', 'New York', '1234567890', 100),
    (2, 'Jane', 'Smith', 'avatar2.png', '1992-02-02', 'Los Angeles', '0987654321', 200),
    (3, 'Alice', 'Johnson', 'avatar3.png', '1985-03-03', 'Chicago', '1122334455', 150),
    (4, 'Bob', 'Brown', 'avatar4.png', '1988-04-04', 'Houston', '5566778899', 120),
    (5, 'Charlie', 'Davis', 'avatar5.png', '1995-05-05', 'Phoenix', '6677889900', 80);

INSERT INTO events (id, name, description, total_xp, start_date, end_date)
VALUES
    ('e1a0f4a6-dc77-4e5a-bb56-f03339d8d4f5', 'Tree Planting', 'Community tree planting event', 50, '2024-08-01', '2024-08-01'),
    ('f2a5e2b8-4134-4f7b-b44e-a0e8e01e90c6', 'Beach Clean-up', 'Clean-up event at the local beach', 70, '2024-08-15', '2024-08-15'),
    ('c3b2c5d9-527e-4e0b-9c3e-e28e98a5b849', 'Recycling Workshop', 'Workshop on recycling techniques', 30, '2024-09-01', '2024-09-01'),
    ('d4c3d6ea-6389-4f8b-ac4f-14a2b1b7d90a', 'Wildlife Conservation', 'Awareness event on wildlife conservation', 60, '2024-09-15', '2024-09-15'),
    ('e5d4e7fb-748e-4f9c-8d5f-25b3c2c8e1a1', 'Energy Saving', 'Event on energy-saving methods', 40, '2024-10-01', '2024-10-01');

INSERT INTO history (id, user_id, event_id, start_date, end_date, xp_given)
VALUES
    ('e5d4e7fb-748e-4f9c-8d5f-25b3c2c8e1a2', 1, 'e1a0f4a6-dc77-4e5a-bb56-f03339d8d4f5', '2024-08-01', '2024-08-01', 50),
    ('e5d4e7fb-748e-4f9c-8d5f-25b3c2c8e1a3', 2, 'f2a5e2b8-4134-4f7b-b44e-a0e8e01e90c6', '2024-08-15', '2024-08-15', 70),
    ('e5d4e7fb-748e-4f9c-8d5f-25b3c2c8e1a4', 3, 'c3b2c5d9-527e-4e0b-9c3e-e28e98a5b849', '2024-09-01', '2024-09-01', 30),
    ('e5d4e7fb-748e-4f9c-8d5f-25b3c2c8e1a5', 4, 'd4c3d6ea-6389-4f8b-ac4f-14a2b1b7d90a', '2024-09-15', '2024-09-15', 60),
    ('e5d4e7fb-748e-4f9c-8d5f-25b3c2c8e1a6', 5, 'e5d4e7fb-748e-4f9c-8d5f-25b3c2c8e1a1', '2024-10-01', '2024-10-01', 40);
