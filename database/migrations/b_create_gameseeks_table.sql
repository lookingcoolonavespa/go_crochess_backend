CREATE SEQUENCE IF NOT EXISTS crochess.gameseek_seq
    START 1
    INCREMENT 1
    MINVALUE 1
    MAXVALUE 2147483647
    CACHE 1;

CREATE TABLE IF NOT EXISTS crochess.gameseeks (
    id SERIAL PRIMARY KEY,
    color VARCHAR(6) NOT NULL CHECK (color IN ('white', 'black', 'random')),
    time INTEGER NOT NULL ,
    increment INTEGER NOT NULL CHECK (increment >= 0 AND increment <= 60),
    seeker VARCHAR(255) NOT NULL
);


