-- +goose Up
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL CHECK (LENGTH(TRIM(name)) >= 1 AND LENGTH(name) <= 200),
    parent_id INTEGER REFERENCES departments(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (name, parent_id)
);

CREATE INDEX idx_departments_parent_id ON departments(parent_id);

-- +goose Down
DROP TABLE IF EXISTS departments;
