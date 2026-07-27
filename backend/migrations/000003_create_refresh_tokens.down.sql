-- Rollback: drop refresh_tokens table and its indexes.
-- CASCADE is not needed here because no other tables reference refresh_tokens.
DROP TABLE IF EXISTS refresh_tokens;
