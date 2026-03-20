-- =========================
-- accounts
-- =========================
CREATE TABLE IF NOT EXISTS accounts (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  owner VARCHAR(255) NOT NULL,
  balance BIGINT NOT NULL,
  currency VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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

  FOREIGN KEY (account_id) REFERENCES accounts(id)
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

  FOREIGN KEY (from_account_id) REFERENCES accounts(id),
  FOREIGN KEY (to_account_id) REFERENCES accounts(id)
);

CREATE INDEX idx_transfers_from_account ON transfers(from_account_id);
CREATE INDEX idx_transfers_to_account ON transfers(to_account_id);
CREATE INDEX idx_transfers_from_to ON transfers(from_account_id, to_account_id);