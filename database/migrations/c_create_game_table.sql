CREATE TABLE IF NOT EXISTS crochess.game (
    id SERIAL PRIMARY KEY,
    white_id VARCHAR(255) NOT NULL,
    black_id VARCHAR(255) NOT NULL,
    time INTEGER NOT NULL,
    increment INTEGER NOT NULL CHECK (increment >= 0 AND increment <= 60),
    result VARCHAR(20),
    method VARCHAR(50),
    version INTEGER NOT NULL,
    time_stamp_at_turn_start BIGINT,
    white_time BIGINT,
    black_time BIGINT,
    history TEXT,
    moves TEXT
);

