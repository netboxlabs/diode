-- +goose Up
CREATE TABLE dummy_table (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    fieldA text NOT NULL
);

-- +goose Down
DROP TABLE dummy_table;
