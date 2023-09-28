CREATE TABLE IF NOT EXISTS crochess.gameseeks (
    id SERIAL PRIMARY KEY,
    color VARCHAR(6) NOT NULL CHECK (color IN ('white', 'black', 'random')),
    time INTEGER NOT NULL ,
    increment INTEGER NOT NULL CHECK (increment >= 0 AND increment <= 60),
    seeker INTEGER NOT NULL
);


