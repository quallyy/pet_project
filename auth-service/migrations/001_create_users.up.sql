-- 001_create_users.up.sql

CREATE TYPE user_role AS ENUM ('customer', 'dealer', 'admin');
CREATE TYPE dealer_country AS ENUM ('CN', 'TR', 'RU');

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role          user_role    NOT NULL,
    phone         VARCHAR(20)  UNIQUE,           -- customers only
    email         VARCHAR(255) UNIQUE,           -- dealers + admins
    password_hash VARCHAR(255),                  -- NULL for customers
    pin_hash      VARCHAR(255),                  -- optional, customers
    dealer_country dealer_country,               -- NULL unless role = dealer
    dealer_name   VARCHAR(255),
    is_active     BOOLEAN      NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_phone ON users (phone) WHERE phone IS NOT NULL;
CREATE INDEX idx_users_email ON users (email) WHERE email IS NOT NULL;