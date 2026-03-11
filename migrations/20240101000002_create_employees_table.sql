-- +goose Up
CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    department_id INTEGER NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    full_name VARCHAR(200) NOT NULL CHECK (LENGTH(TRIM(full_name)) >= 1 AND LENGTH(full_name) <= 200),
    position VARCHAR(200) NOT NULL CHECK (LENGTH(TRIM(position)) >= 1 AND LENGTH(position) <= 200),
    hired_at DATE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_employees_department_id ON employees(department_id);
CREATE INDEX idx_employees_created_at ON employees(created_at);
CREATE INDEX idx_employees_full_name ON employees(full_name);

-- +goose Down
DROP TABLE IF EXISTS employees;
