CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE organizations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT organizations_name_not_empty CHECK (length(trim(name)) > 0)
);

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email text NOT NULL,
    email_normalized text NOT NULL,
    password_hash text NOT NULL,
    role text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT users_email_not_empty CHECK (length(trim(email)) > 0),
    CONSTRAINT users_email_normalized_not_empty CHECK (length(trim(email_normalized)) > 0),
    CONSTRAINT users_password_hash_not_empty CHECK (length(trim(password_hash)) > 0),
    CONSTRAINT users_role_not_empty CHECK (length(trim(role)) > 0),
    CONSTRAINT users_email_normalized_unique UNIQUE (email_normalized)
);

CREATE INDEX users_organization_id_idx ON users (organization_id);
