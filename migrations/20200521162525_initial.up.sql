CREATE TABLE IF NOT EXISTS registrations (
  id SERIAL PRIMARY KEY,
  account_id bytea NOT NULL UNIQUE,
  email text NOT NULL,
  name text,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS hostnames (
  id SERIAL PRIMARY KEY,
  registration_id int NOT NULL,
  hostname text NOT NULL UNIQUE,
  labels text[] NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);
