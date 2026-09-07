-- ==============================================================================
-- SUN Booking Tours — Database Schema Migration Rollback (000001_init_schema.down.sql)
-- ==============================================================================

-- Drop Materialized Views and Procedures
DROP PROCEDURE IF EXISTS refresh_revenue_reports();
DROP MATERIALIZED VIEW IF EXISTS mv_monthly_revenue_report;
DROP MATERIALIZED VIEW IF EXISTS mv_daily_revenue_report;

-- Drop Triggers & Functions
DROP TRIGGER IF EXISTS trg_comments_stat ON comments;
DROP FUNCTION IF EXISTS update_review_comment_stats();

DROP TRIGGER IF EXISTS trg_review_likes_stat ON review_likes;
DROP FUNCTION IF EXISTS update_review_like_stats();

DROP TRIGGER IF EXISTS trg_tour_ratings_stat ON tour_ratings;
DROP FUNCTION IF EXISTS update_tour_rating_stats();

DROP TRIGGER IF EXISTS set_timestamp_tour_ratings ON tour_ratings;
DROP TRIGGER IF EXISTS set_timestamp_comments ON comments;
DROP TRIGGER IF EXISTS set_timestamp_reviews ON reviews;
DROP TRIGGER IF EXISTS set_timestamp_review_categories ON review_categories;
DROP TRIGGER IF EXISTS set_timestamp_payments ON payments;
DROP TRIGGER IF EXISTS set_timestamp_bookings ON bookings;
DROP TRIGGER IF EXISTS set_timestamp_tour_schedules ON tour_schedules;
DROP TRIGGER IF EXISTS set_timestamp_tours ON tours;
DROP TRIGGER IF EXISTS set_timestamp_categories ON categories;
DROP TRIGGER IF EXISTS set_timestamp_bank ON user_bank_accounts;
DROP TRIGGER IF EXISTS set_timestamp_oauth ON user_oauth_accounts;
DROP TRIGGER IF EXISTS set_timestamp_users ON users;
DROP FUNCTION IF EXISTS trigger_set_timestamp();
DROP FUNCTION IF EXISTS f_unaccent(text);

-- Drop Tables in Reverse Dependency Order
DROP TABLE IF EXISTS activity_logs;
DROP TABLE IF EXISTS tour_ratings;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS review_likes;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS review_categories;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS tour_schedules;
DROP TABLE IF EXISTS tour_images;
DROP TABLE IF EXISTS tours;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS user_bank_accounts;
DROP TABLE IF EXISTS user_oauth_accounts;
DROP TABLE IF EXISTS users;
