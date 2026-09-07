-- ==============================================================================
-- SUN Booking Tours — Database Schema Migration (000001_init_schema.up.sql)
-- Target: PostgreSQL 16+
-- ==============================================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "unaccent";

-- Immutable wrapper for unaccent to allow use in indexed expressions
CREATE OR REPLACE FUNCTION f_unaccent(text)
RETURNS text AS $$
SELECT public.unaccent('public.unaccent', $1);
$$ LANGUAGE sql IMMUTABLE PARALLEL SAFE;

-- ------------------------------------------------------------------------------
-- 1. USERS
-- ------------------------------------------------------------------------------
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NULL,
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NULL,
    avatar_url TEXT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users(role) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- ------------------------------------------------------------------------------
-- 2. USER OAUTH ACCOUNTS (Facebook, Twitter, Google)
-- ------------------------------------------------------------------------------
CREATE TABLE user_oauth_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(20) NOT NULL CHECK (provider IN ('google', 'facebook', 'twitter')),
    provider_user_id VARCHAR(255) NOT NULL,
    email VARCHAR(255) NULL,
    access_token TEXT NULL,
    refresh_token TEXT NULL,
    token_expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_oauth_provider_user UNIQUE (provider, provider_user_id)
);

CREATE INDEX idx_oauth_user_id ON user_oauth_accounts(user_id);

-- ------------------------------------------------------------------------------
-- 3. USER BANK ACCOUNTS
-- ------------------------------------------------------------------------------
CREATE TABLE user_bank_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bank_name VARCHAR(100) NOT NULL,
    bank_code VARCHAR(30) NULL,
    account_number VARCHAR(100) NOT NULL,
    account_holder_name VARCHAR(255) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_bank_accounts_user_id ON user_bank_accounts(user_id) WHERE deleted_at IS NULL;

-- ------------------------------------------------------------------------------
-- 4. TOUR CATEGORIES
-- ------------------------------------------------------------------------------
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(120) NOT NULL UNIQUE,
    description TEXT NULL,
    image_url TEXT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_categories_slug ON categories(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_categories_is_active ON categories(is_active) WHERE deleted_at IS NULL;

-- ------------------------------------------------------------------------------
-- 5. TOURS
-- ------------------------------------------------------------------------------
CREATE TABLE tours (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(280) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    itinerary TEXT NULL,
    destination VARCHAR(255) NOT NULL,
    duration_days INT NOT NULL CHECK (duration_days > 0),
    duration_nights INT NOT NULL CHECK (duration_nights >= 0),
    price NUMERIC(15, 2) NOT NULL CHECK (price >= 0),
    discount_price NUMERIC(15, 2) NULL CHECK (discount_price IS NULL OR (discount_price >= 0 AND discount_price <= price)),
    max_participants INT NOT NULL CHECK (max_participants > 0),
    thumbnail_url TEXT NULL,
    highlights TEXT[] NULL,
    inclusions TEXT NULL,
    exclusions TEXT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    avg_rating NUMERIC(3, 2) NOT NULL DEFAULT 0.00,
    total_ratings INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_tours_category_id ON tours(category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tours_status ON tours(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_tours_destination ON tours(destination) WHERE deleted_at IS NULL;
CREATE INDEX idx_tours_price ON tours(price) WHERE deleted_at IS NULL;
CREATE INDEX idx_tours_created_at ON tours(created_at DESC) WHERE deleted_at IS NULL;

-- Full-text & Trigram search indexes
CREATE INDEX idx_tours_trgm_title ON tours USING gin (title gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_tours_trgm_dest ON tours USING gin (destination gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_tours_fts ON tours USING gin (
    to_tsvector('simple', f_unaccent(title || ' ' || destination || ' ' || COALESCE(description, '')))
) WHERE deleted_at IS NULL;

-- ------------------------------------------------------------------------------
-- 6. TOUR IMAGES
-- ------------------------------------------------------------------------------
CREATE TABLE tour_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tour_id UUID NOT NULL REFERENCES tours(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    caption VARCHAR(255) NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_tour_images_tour_id ON tour_images(tour_id) WHERE deleted_at IS NULL;

-- ------------------------------------------------------------------------------
-- 7. TOUR SCHEDULES (Departure Schedules)
-- ------------------------------------------------------------------------------
CREATE TABLE tour_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tour_id UUID NOT NULL REFERENCES tours(id) ON DELETE CASCADE,
    departure_date DATE NOT NULL,
    return_date DATE NOT NULL,
    available_slots INT NOT NULL CHECK (available_slots >= 0),
    price_override NUMERIC(15, 2) NULL CHECK (price_override IS NULL OR price_override >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT check_schedule_dates CHECK (return_date >= departure_date),
    CONSTRAINT uq_tour_schedule_date UNIQUE (tour_id, departure_date)
);

CREATE INDEX idx_tour_schedules_tour_id ON tour_schedules(tour_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tour_schedules_departure ON tour_schedules(departure_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_tour_schedules_status ON tour_schedules(status) WHERE deleted_at IS NULL;

-- ------------------------------------------------------------------------------
-- 8. BOOKINGS
-- ------------------------------------------------------------------------------
CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_code VARCHAR(32) NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    tour_id UUID NOT NULL REFERENCES tours(id) ON DELETE RESTRICT,
    schedule_id UUID NOT NULL REFERENCES tour_schedules(id) ON DELETE RESTRICT,
    num_participants INT NOT NULL CHECK (num_participants > 0),
    unit_price NUMERIC(15, 2) NOT NULL CHECK (unit_price >= 0),
    total_price NUMERIC(15, 2) NOT NULL CHECK (total_price >= 0),
    contact_name VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(20) NOT NULL,
    contact_email VARCHAR(255) NOT NULL,
    special_requests TEXT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'completed', 'cancelled')),
    cancelled_at TIMESTAMPTZ NULL,
    cancellation_reason TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_bookings_user_id ON bookings(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_bookings_tour_id ON bookings(tour_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_bookings_schedule_id ON bookings(schedule_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_bookings_status ON bookings(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_bookings_created_at ON bookings(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_bookings_code ON bookings(booking_code);

-- ------------------------------------------------------------------------------
-- 9. PAYMENTS (Internet Banking)
-- ------------------------------------------------------------------------------
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE RESTRICT,
    user_bank_account_id UUID NULL REFERENCES user_bank_accounts(id) ON DELETE SET NULL,
    amount NUMERIC(15, 2) NOT NULL CHECK (amount >= 0),
    payment_method VARCHAR(30) NOT NULL DEFAULT 'internet_banking' 
        CHECK (payment_method IN ('internet_banking', 'bank_transfer', 'vnpay', 'momo')),
    transaction_ref VARCHAR(100) NULL UNIQUE,
    bank_name VARCHAR(100) NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' 
        CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'refunded')),
    payment_gateway_response JSONB NULL,
    paid_at TIMESTAMPTZ NULL,
    failed_reason TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_booking_id ON payments(booking_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_transaction_ref ON payments(transaction_ref);
CREATE INDEX idx_payments_paid_at ON payments(paid_at) WHERE status = 'completed';

-- ------------------------------------------------------------------------------
-- 10. REVIEW CATEGORIES (place, food, news)
-- ------------------------------------------------------------------------------
CREATE TABLE review_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    slug VARCHAR(50) NOT NULL UNIQUE,
    description TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_review_categories_slug ON review_categories(slug) WHERE deleted_at IS NULL;

-- ------------------------------------------------------------------------------
-- 11. REVIEWS (Independent place, food, news articles)
-- ------------------------------------------------------------------------------
CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    review_category_id UUID NOT NULL REFERENCES review_categories(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(280) NOT NULL UNIQUE,
    content TEXT NOT NULL,
    thumbnail_url TEXT NULL,
    like_count INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'published' CHECK (status IN ('draft', 'published', 'hidden')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_reviews_user_id ON reviews(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_reviews_category_id ON reviews(review_category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_reviews_status ON reviews(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_reviews_created_at ON reviews(created_at DESC) WHERE deleted_at IS NULL;

-- Trigram search for reviews
CREATE INDEX idx_reviews_trgm_title ON reviews USING gin (title gin_trgm_ops) WHERE deleted_at IS NULL;

-- ------------------------------------------------------------------------------
-- 12. REVIEW LIKES
-- ------------------------------------------------------------------------------
CREATE TABLE review_likes (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    review_id UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, review_id)
);

CREATE INDEX idx_review_likes_review_id ON review_likes(review_id);

-- ------------------------------------------------------------------------------
-- 13. COMMENTS (Hierarchical: review comments & nested comments)
-- ------------------------------------------------------------------------------
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id UUID NULL REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_comments_review_id ON comments(review_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_comments_parent_id ON comments(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_comments_user_id ON comments(user_id) WHERE deleted_at IS NULL;

-- ------------------------------------------------------------------------------
-- 14. TOUR RATINGS (Only for users with a confirmed/completed booking)
-- ------------------------------------------------------------------------------
CREATE TABLE tour_ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tour_id UUID NOT NULL REFERENCES tours(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE RESTRICT,
    score SMALLINT NOT NULL CHECK (score >= 1 AND score <= 5),
    comment TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT uq_tour_booking_rating UNIQUE (booking_id)
);

CREATE INDEX idx_tour_ratings_tour_id ON tour_ratings(tour_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tour_ratings_user_id ON tour_ratings(user_id) WHERE deleted_at IS NULL;

-- ------------------------------------------------------------------------------
-- 15. ACTIVITY LOGS (User & Admin Activities)
-- ------------------------------------------------------------------------------
CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL CHECK (action IN (
        'booking_tour', 
        'cancel_tour', 
        'create_review', 
        'pay_tour', 
        'rate_tour', 
        'comment_review', 
        'like_review', 
        'login', 
        'logout'
    )),
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    metadata JSONB NULL,
    ip_address INET NULL,
    user_agent TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_action ON activity_logs(action);
CREATE INDEX idx_activity_logs_created_at ON activity_logs(created_at DESC);
CREATE INDEX idx_activity_logs_entity ON activity_logs(entity_type, entity_id);

-- ------------------------------------------------------------------------------
-- 16. BATCH REVENUE REPORTING (Materialized Views)
-- ------------------------------------------------------------------------------

-- Daily Revenue Materialized View
CREATE MATERIALIZED VIEW mv_daily_revenue_report AS
SELECT
    DATE_TRUNC('day', p.paid_at)::DATE AS report_date,
    t.id AS tour_id,
    t.title AS tour_title,
    c.id AS category_id,
    c.name AS category_name,
    COUNT(DISTINCT b.id) AS total_bookings,
    SUM(b.num_participants) AS total_participants,
    SUM(p.amount) AS total_revenue
FROM payments p
JOIN bookings b ON p.booking_id = b.id
JOIN tours t ON b.tour_id = t.id
JOIN categories c ON t.category_id = c.id
WHERE p.status = 'completed'
  AND p.paid_at IS NOT NULL
  AND b.deleted_at IS NULL
GROUP BY DATE_TRUNC('day', p.paid_at)::DATE, t.id, t.title, c.id, c.name;

CREATE UNIQUE INDEX idx_mv_daily_revenue ON mv_daily_revenue_report(report_date, tour_id);

-- Monthly Revenue Materialized View
CREATE MATERIALIZED VIEW mv_monthly_revenue_report AS
SELECT
    TO_CHAR(DATE_TRUNC('month', report_date), 'YYYY-MM') AS report_month,
    category_id,
    category_name,
    SUM(total_bookings) AS total_bookings,
    SUM(total_participants) AS total_participants,
    SUM(total_revenue) AS total_revenue
FROM mv_daily_revenue_report
GROUP BY TO_CHAR(DATE_TRUNC('month', report_date), 'YYYY-MM'), category_id, category_name;

CREATE UNIQUE INDEX idx_mv_monthly_revenue ON mv_monthly_revenue_report(report_month, category_id);

-- Stored Procedure to Refresh Revenue Views in Batch
CREATE OR REPLACE PROCEDURE refresh_revenue_reports()
LANGUAGE plpgsql
AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_daily_revenue_report;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_monthly_revenue_report;
END;
$$;

-- ------------------------------------------------------------------------------
-- 17. DATABASE TRIGGERS
-- ------------------------------------------------------------------------------

-- Function to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_timestamp_users BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp_oauth BEFORE UPDATE ON user_oauth_accounts FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp_bank BEFORE UPDATE ON user_bank_accounts FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp_categories BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp_tours BEFORE UPDATE ON tours FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp_tour_schedules BEFORE UPDATE ON tour_schedules FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp_bookings BEFORE UPDATE ON bookings FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp_payments BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp_review_categories BEFORE UPDATE ON review_categories FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp_reviews BEFORE UPDATE ON reviews FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp_comments BEFORE UPDATE ON comments FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp_tour_ratings BEFORE UPDATE ON tour_ratings FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();

-- Trigger to update tours.avg_rating and tours.total_ratings when rating changes
CREATE OR REPLACE FUNCTION update_tour_rating_stats()
RETURNS TRIGGER AS $$
DECLARE
    target_tour_id UUID;
BEGIN
    IF TG_OP = 'DELETE' THEN
        target_tour_id := OLD.tour_id;
    ELSE
        target_tour_id := NEW.tour_id;
    END IF;

    UPDATE tours
    SET avg_rating = COALESCE((
            SELECT ROUND(AVG(score)::numeric, 2)
            FROM tour_ratings
            WHERE tour_id = target_tour_id AND deleted_at IS NULL
        ), 0.00),
        total_ratings = COALESCE((
            SELECT COUNT(*)
            FROM tour_ratings
            WHERE tour_id = target_tour_id AND deleted_at IS NULL
        ), 0)
    WHERE id = target_tour_id;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_tour_ratings_stat
AFTER INSERT OR UPDATE OR DELETE ON tour_ratings
FOR EACH ROW EXECUTE FUNCTION update_tour_rating_stats();

-- Trigger to increment/decrement reviews.like_count
CREATE OR REPLACE FUNCTION update_review_like_stats()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE reviews SET like_count = like_count + 1 WHERE id = NEW.review_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE reviews SET like_count = GREATEST(0, like_count - 1) WHERE id = OLD.review_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_review_likes_stat
AFTER INSERT OR DELETE ON review_likes
FOR EACH ROW EXECUTE FUNCTION update_review_like_stats();

-- Trigger to update reviews.comment_count
CREATE OR REPLACE FUNCTION update_review_comment_stats()
RETURNS TRIGGER AS $$
DECLARE
    target_review_id UUID;
BEGIN
    IF TG_OP = 'DELETE' THEN
        target_review_id := OLD.review_id;
    ELSE
        target_review_id := NEW.review_id;
    END IF;

    UPDATE reviews
    SET comment_count = (
        SELECT COUNT(*)
        FROM comments
        WHERE review_id = target_review_id AND deleted_at IS NULL AND is_hidden = FALSE
    )
    WHERE id = target_review_id;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_comments_stat
AFTER INSERT OR UPDATE OR DELETE ON comments
FOR EACH ROW EXECUTE FUNCTION update_review_comment_stats();
