CREATE TABLE IF NOT EXISTS products (
    id bigint PRIMARY KEY,
    type text,
    attrs JSONB
);