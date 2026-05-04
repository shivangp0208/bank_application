-- =========================
-- users
-- =========================
CREATE TABLE IF NOT EXISTS users (
  username VARCHAR(255) PRIMARY KEY,
  hashed_password VARCHAR(255) NOT NULL,
  role VARCHAR(150) NOT NULL DEFAULT 'user',
  full_name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  is_verified BOOLEAN NOT NULL DEFAULT false,
  password_changed_at TIMESTAMP NULL DEFAULT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- =========================
-- verify_emails
-- =========================
CREATE TABLE IF NOT EXISTS verify_emails (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  secret_code VARCHAR(255) NOT NULL,
  is_used BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expired_at TIMESTAMP AS (created_at + INTERVAL 15 MINUTE),

  CONSTRAINT fk_verify_username FOREIGN KEY (username) REFERENCES users(username)
);

-- =========================
-- accounts
-- =========================
CREATE TABLE IF NOT EXISTS accounts (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  owner VARCHAR(255) NOT NULL,
  balance BIGINT NOT NULL,
  currency VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT fk_owner FOREIGN KEY (owner) REFERENCES users(username),
  CONSTRAINT owner_currency_key UNIQUE (owner, currency)
);
CREATE INDEX idx_accounts_owner ON accounts(owner);

-- =========================
-- entries
-- =========================
CREATE TABLE IF NOT EXISTS entries (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  account_id BIGINT UNSIGNED NOT NULL,
  amount BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT fk_account_id FOREIGN KEY (account_id) REFERENCES accounts(id)
);
CREATE INDEX idx_entries_account_id ON entries(account_id);

-- =========================
-- transfers
-- =========================
CREATE TABLE IF NOT EXISTS transfers (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  from_account_id BIGINT UNSIGNED NOT NULL,
  to_account_id BIGINT UNSIGNED NOT NULL,
  amount BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT fk_from_account FOREIGN KEY (from_account_id) REFERENCES accounts(id),
  CONSTRAINT fk_to_account FOREIGN KEY (to_account_id) REFERENCES accounts(id)
);
CREATE INDEX idx_transfers_from_account ON transfers(from_account_id);
CREATE INDEX idx_transfers_to_account ON transfers(to_account_id);
CREATE INDEX idx_transfers_from_to ON transfers(from_account_id, to_account_id);

-- =========================
-- sessions
-- =========================
CREATE TABLE IF NOT EXISTS sessions (
  id CHAR(36) PRIMARY KEY,
  username VARCHAR(255) NOT NULL,
  refresh_token TEXT NOT NULL,
  user_agent VARCHAR(255) NOT NULL,
  client_ip VARCHAR(255) NOT NULL,
  is_blocked BOOLEAN NOT NULL DEFAULT false,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT fk_session_username FOREIGN KEY (username) REFERENCES users(username)
);