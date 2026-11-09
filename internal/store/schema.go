package store

const schema = `
CREATE TABLE IF NOT EXISTS loans (
  id                INTEGER PRIMARY KEY AUTOINCREMENT,
  qarz_uid          TEXT    NOT NULL UNIQUE,
  borrower          TEXT    NOT NULL,
  principal_minor   INTEGER NOT NULL CHECK (principal_minor > 0),
  lent_on           TEXT    NOT NULL,
  due_on            TEXT    NOT NULL,
  note              TEXT,
  metadata          TEXT,
  voided_at         TEXT,
  calendar_event_id TEXT,
  sync_state        TEXT    NOT NULL CHECK (sync_state IN
                      ('synced','pending_create','pending_update','pending_delete')),
  created_at        TEXT    NOT NULL,
  updated_at        TEXT    NOT NULL,
  CHECK (due_on >= lent_on)
);

CREATE TABLE IF NOT EXISTS repayments (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  loan_id      INTEGER NOT NULL REFERENCES loans(id),
  amount_minor INTEGER NOT NULL CHECK (amount_minor > 0),
  paid_on      TEXT    NOT NULL,
  note         TEXT
);

CREATE INDEX IF NOT EXISTS idx_repayments_loan_id ON repayments(loan_id);
`
