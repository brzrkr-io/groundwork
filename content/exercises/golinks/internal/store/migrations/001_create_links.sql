-- Migration 1: the table that holds every go link.
-- Open runs each file in this folder once, in name order, and records it in
-- schema_migrations so it never runs twice.

-- TODO(2): Create the table links with two text columns: name, the primary
-- key, and url. Neither may be NULL.
