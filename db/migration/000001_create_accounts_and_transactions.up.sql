-- Create accounts table
CREATE TABLE accounts (
  account_id bigint PRIMARY KEY,
  balance bigint NOT NULL CHECK (balance >= 0),
  currency varchar NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Create transactions table
CREATE TABLE transactions (
  id bigserial PRIMARY KEY,
  source_account_id bigint NOT NULL,
  destination_account_id bigint NOT NULL,
  amount bigint NOT NULL CHECK (amount > 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  
  CONSTRAINT fk_source_account
    FOREIGN KEY (source_account_id)
    REFERENCES accounts(account_id)
    ON DELETE CASCADE,

  CONSTRAINT fk_destination_account
    FOREIGN KEY (destination_account_id)
    REFERENCES accounts(account_id)
    ON DELETE CASCADE,
  
  CONSTRAINT check_source_not_eq_destination
    CHECK (source_account_id <> destination_account_id)
);

-- Indexes for faster queries
CREATE INDEX idx_source_account ON transactions(source_account_id);
CREATE INDEX idx_destination_account ON transactions(destination_account_id);
