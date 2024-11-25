-- +goose Up
CREATE TABLE dummy_b_table (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    fieldA text NOT NULL
);

-- +goose Down
DROP TABLE dummy_b_table;
