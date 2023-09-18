CREATE TABLE IF NOT EXISTS crochess.drawrecord (
    white BOOLEAN,
    black BOOLEAN,
    game_id INTEGER NOT NULL,
    CONSTRAINT drawrecord_pkey PRIMARY KEY (game_id),
    CONSTRAINT drawrecord_game_id_fkey FOREIGN KEY (game_id) REFERENCES crochess.game (id)
);
