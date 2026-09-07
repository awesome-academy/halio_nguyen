-- ==============================================================================
-- SUN Booking Tours — Seed Initial Data (000002_seed_data.up.sql)
-- ==============================================================================

-- 1. Review Categories (Required by specifications: place, food, news)
INSERT INTO review_categories (id, name, slug, description) VALUES
    ('11111111-1111-1111-1111-111111111001', 'Places & Attractions', 'place', 'Reviews and guides for destinations, scenic landmarks, and viewpoints'),
    ('11111111-1111-1111-1111-111111111002', 'Food & Dining', 'food', 'Local cuisine, specialty dishes, restaurants, and food tours'),
    ('11111111-1111-1111-1111-111111111003', 'Travel News', 'news', 'Travel news, festivals, seasonal tips, and guides')
ON CONFLICT (id) DO NOTHING;

-- 2. Tour Categories
INSERT INTO categories (id, name, slug, description, sort_order) VALUES
    ('22222222-2222-2222-2222-222222222001', 'Island & Coastal Tours', 'island-coastal', 'Explore stunning islands, bays, and beaches', 1),
    ('22222222-2222-2222-2222-222222222002', 'Mountain & Trekking', 'mountain-trekking', 'Conquer mountain trails, passes, and highlands', 2),
    ('22222222-2222-2222-2222-222222222003', 'Culture & Heritage', 'culture-heritage', 'Immerse in ancient history, architecture, and cultural traditions', 3),
    ('22222222-2222-2222-2222-222222222004', 'Resort & Relaxation', 'resort-relaxation', 'Rejuvenate at luxury wellness resorts and hot springs', 4)
ON CONFLICT (id) DO NOTHING;

-- 3. Initial Users (Admin & Sample User)
-- Default password: Admin@123456 (bcrypt hash: $2a$10$w8gZ4.F3x1yN1g1B9T9aOeG0m9r2QeX1lY0b4L9h7M2s5P8t0R6uq)
INSERT INTO users (id, email, password_hash, full_name, phone, role) VALUES
    ('33333333-3333-3333-3333-333333333001', 'admin@sunbooking.com', '$2a$10$w8gZ4.F3x1yN1g1B9T9aOeG0m9r2QeX1lY0b4L9h7M2s5P8t0R6uq', 'System Administrator', '0901234567', 'admin'),
    ('33333333-3333-3333-3333-333333333002', 'tourist@sunbooking.com', '$2a$10$w8gZ4.F3x1yN1g1B9T9aOeG0m9r2QeX1lY0b4L9h7M2s5P8t0R6uq', 'John Tourist', '0987654321', 'user')
ON CONFLICT (email) DO NOTHING;

-- 4. Sample Tours
INSERT INTO tours (
    id, category_id, title, slug, description, itinerary, destination, 
    duration_days, duration_nights, price, discount_price, max_participants, 
    thumbnail_url, highlights, status
) VALUES
    (
        '44444444-4444-4444-4444-444444444001',
        '22222222-2222-2222-2222-222222222001',
        'Phu Quoc Tropical Island Discovery 3D2N',
        'phu-quoc-tropical-island-discovery-3d2n',
        'A 3-day 2-night getaway exploring Grand World, coral reef snorkeling around South Island, and fresh seafood tasting.',
        'Day 1: Airport pick-up, Grand World exploration. Day 2: 4-island speedboat tour and snorkeling. Day 3: Night market, souvenir shopping, airport drop-off.',
        'Phu Quoc, Kien Giang',
        3, 2, 4500000, 3990000, 25,
        'https://images.unsplash.com/photo-1540555700478-4be289fbecef',
        ARRAY['Grand World sleepless city', 'Snorkeling at May Rut Island', 'World longest sea-crossing cable car'],
        'published'
    ),
    (
        '44444444-4444-4444-4444-444444444002',
        '22222222-2222-2222-2222-222222222002',
        'Misty Sapa & Fansipan Peak Conquest 2D1N',
        'misty-sapa-fansipan-conquest-2d1n',
        'Experience the golden rice terraces of Northwest Vietnam, reach the 3,143m rooftop of Indochina, and discover Cat Cat village.',
        'Day 1: Hanoi to Sapa, Cat Cat village trek. Day 2: Fansipan cable car, return to Hanoi.',
        'Sapa, Lao Cai',
        2, 1, 2800000, 2490000, 20,
        'https://images.unsplash.com/photo-1528127269322-539801943592',
        ARRAY['Fansipan Peak 3,143m', 'Picturesque Cat Cat Village', 'Traditional sturgeon hotpot'],
        'published'
    )
ON CONFLICT (id) DO NOTHING;

-- 5. Sample Schedules
INSERT INTO tour_schedules (id, tour_id, departure_date, return_date, available_slots, status) VALUES
    ('55555555-5555-5555-5555-555555555001', '44444444-4444-4444-4444-444444444001', CURRENT_DATE + INTERVAL '7 days', CURRENT_DATE + INTERVAL '9 days', 25, 'open'),
    ('55555555-5555-5555-5555-555555555002', '44444444-4444-4444-4444-444444444001', CURRENT_DATE + INTERVAL '14 days', CURRENT_DATE + INTERVAL '16 days', 25, 'open'),
    ('55555555-5555-5555-5555-555555555003', '44444444-4444-4444-4444-444444444002', CURRENT_DATE + INTERVAL '5 days', CURRENT_DATE + INTERVAL '6 days', 20, 'open')
ON CONFLICT (tour_id, departure_date) DO NOTHING;

-- Refresh Materialized Views with initial seed data
CALL refresh_revenue_reports();
