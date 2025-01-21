-- +goose Up
CREATE TABLE prices
(
    id          INT PRIMARY KEY,
    name        TEXT           NOT NULL,
    category    TEXT           NOT NULL,
    price       NUMERIC(10, 2) NOT NULL,
    create_date DATE           NOT NULL
);

-- +goose Down
DROP TABLE prices;