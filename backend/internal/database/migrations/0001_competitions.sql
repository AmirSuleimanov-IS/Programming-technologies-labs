CREATE TABLE competitions (
    id          BIGSERIAL PRIMARY KEY,
    title       TEXT        NOT NULL,
    sport       TEXT        NOT NULL,
    location    TEXT        NOT NULL,
    starts_on   DATE        NOT NULL,
    capacity    INTEGER     NOT NULL CHECK (capacity BETWEEN 1 AND 1000),
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX competitions_starts_on_idx ON competitions (starts_on);
