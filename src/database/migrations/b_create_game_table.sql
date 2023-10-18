CREATE TABLE IF NOT EXISTS crochess.game (
    id SERIAL PRIMARY KEY,
    white_id VARCHAR(10) NOT NULL,
    black_id VARCHAR(10) NOT NULL,
    time INTEGER NOT NULL,
    increment INTEGER NOT NULL CHECK (increment >= 0 AND increment <= 60),
    result VARCHAR(20) DEFAULT '',
    method VARCHAR(50) DEFAULT '',
    version INTEGER NOT NULL,
    time_stamp_at_turn_start BIGINT,
    white_time INT,
    black_time INT,
    moves TEXT DEFAULT '',
    black_draw_status BOOLEAN DEFAULT false,
    white_draw_status BOOLEAN DEFAULT false
);

