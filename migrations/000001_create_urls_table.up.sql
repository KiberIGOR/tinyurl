-- migrations/000001_create_urls_table.up.sql
-- Создание таблицы ссылок
CREATE TABLE IF NOT EXISTS urls (
    short_url    VARCHAR(32)  NOT NULL PRIMARY KEY,
    original_url TEXT         NOT NULL
);
