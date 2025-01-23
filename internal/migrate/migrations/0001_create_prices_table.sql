-- +goose Up
CREATE TABLE prices
(
    id          SERIAL PRIMARY KEY,
    name        TEXT           NOT NULL,
    category    TEXT           NOT NULL,
    price       NUMERIC(10, 2) NOT NULL,
    create_date TIMESTAMP      NOT NULL
);

-- +goose Down
DROP TABLE prices;