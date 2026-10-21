CREATE TABLE applications (
    id             BIGSERIAL PRIMARY KEY,
    competition_id BIGINT      NOT NULL REFERENCES competitions (id) ON DELETE CASCADE,
    user_id        BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status         TEXT        NOT NULL CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Не больше одной активной заявки пользователя на соревнование
CREATE UNIQUE INDEX applications_one_active_idx ON applications (user_id, competition_id)
    WHERE status IN ('pending', 'approved');

CREATE INDEX applications_competition_idx ON applications (competition_id, status);
