CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

--gopg:split

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  last_login timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

--gopg:split

CREATE TABLE IF NOT EXISTS twitch_accounts (
  id text PRIMARY KEY,
  email text NOT NULL,
  email_verified boolean NOT NULL,
  username varchar(25) NOT NULL,
  picture text,
  access_token varchar(255) NOT NULL,
  refresh_token varchar(255) NOT NULL,
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

--gopg:split

CREATE TABLE IF NOT EXISTS mc_accounts (
  id uuid PRIMARY KEY,
  username varchar(16) NOT NULL,
  skin text,
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

--gopg:split

CREATE TABLE IF NOT EXISTS games (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  name varchar(60) NOT NULL,
  motd varchar(59),
  slots int DEFAULT 10 NOT NULL,
  address varchar(63) NOT NULL,
  edition varchar(25) NOT NULL,
  state varchar(25) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);
