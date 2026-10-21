CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT        NOT NULL UNIQUE,
    full_name     TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL CHECK (role IN ('ADMIN', 'USER')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
