CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

--gopg:split

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  email varchar(255) NOT NULL,
  email_verified boolean NOT NULL,
  twitch_id varchar(255),
  twitch_username varchar(25),
  twitch_picture text,
  mc_username varchar(16),
  mc_uuid uuid,
  last_login_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE UNIQUE INDEX twitch_id_idx ON users (twitch_id);

--gopg:split

CREATE TABLE twitch_tokens (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  access_token varchar(255) NOT NULL,
  refresh_token varchar(255) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

--gopg:split

CREATE TABLE games (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  name varchar(255) NOT NULL,
  motd varchar(255) NOT NULL,
  slots int DEFAULT 10,
  address varchar(63) NOT NULL,
  edition varchar(25) NOT NULL,
  state varchar(25) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);
