-- One-off maintenance script (NOT a migration — migrations 000001/000002 are
-- frozen, decisions.md A2). Run this on any database where
-- 000002_seed_data.up.sql already ran before this hash fix landed: the
-- seed file's INSERT is ON CONFLICT (email) DO NOTHING, so re-running it on
-- an already-seeded database silently changes nothing.
--
-- Usage: Get-Content apps/api/scripts/fix-admin-seed-hash.sql -Raw |
--        docker exec -i sun_booking_postgres psql -U sun_booking -d sun_booking_tours
--
-- Sets both seeded accounts to the password "Admin@123456" (bcrypt cost 10,
-- verified independently against bcrypt.CompareHashAndPassword —
-- decisions.md D5).

UPDATE users
SET password_hash = '$2a$10$ZUWCjBrvW9sUo5dxwmSXGOro0PHNQXXzKTx/PfbtDD.W0czJfETGi'
WHERE email IN ('admin@sunbooking.com', 'tourist@sunbooking.com');

-- Verify: SELECT email, password_hash FROM users WHERE email = 'admin@sunbooking.com';
