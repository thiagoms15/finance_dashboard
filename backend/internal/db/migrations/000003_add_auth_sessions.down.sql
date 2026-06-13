DROP TABLE IF EXISTS refresh_tokens;

ALTER TABLE users
DROP COLUMN IF EXISTS locked_until,
DROP COLUMN IF EXISTS first_failed_login_at,
DROP COLUMN IF EXISTS failed_login_attempts;
