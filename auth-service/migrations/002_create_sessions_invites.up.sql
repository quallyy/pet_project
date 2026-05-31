-- 002_create_sessions_invites.up.sql

CREATE TABLE sessions (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(255) NOT NULL UNIQUE,  -- SHA-256 hashed
    device_info   TEXT,
    ip_address    INET,
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id       ON sessions (user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions (refresh_token);

-- dealer invite tokens — admin generates, dealer consumes once
CREATE TABLE invite_tokens (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash     VARCHAR(255)  NOT NULL UNIQUE,
    email          VARCHAR(255)  NOT NULL,
    dealer_country dealer_country NOT NULL,
    dealer_name    VARCHAR(255)  NOT NULL,
    created_by     UUID          NOT NULL REFERENCES users(id),
    used           BOOLEAN       NOT NULL DEFAULT false,
    expires_at     TIMESTAMPTZ   NOT NULL,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- audit log — every auth event, never deleted
CREATE TABLE audit_log (
    id          BIGSERIAL   PRIMARY KEY,
    user_id     UUID        REFERENCES users(id),   -- NULL on failed attempts
    event       VARCHAR(50) NOT NULL,               -- 'login.success', 'login.failed', etc.
    ip_address  INET,
    device_info TEXT,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_user_id ON audit_log (user_id, created_at DESC);
CREATE INDEX idx_audit_event   ON audit_log (event,   created_at DESC);