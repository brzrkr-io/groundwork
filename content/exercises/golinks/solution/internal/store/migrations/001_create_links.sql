-- Migration 1: the table that holds every go link.
-- Open runs each file in this folder once, in name order, and records it in
-- schema_migrations so it never runs twice.

CREATE TABLE links (
    name TEXT NOT NULL PRIMARY KEY,
    url  TEXT NOT NULL
);
