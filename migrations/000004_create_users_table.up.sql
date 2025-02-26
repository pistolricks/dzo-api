CREATE EXTENSION IF NOT EXISTS tablefunc;


CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    version integer NOT NULL DEFAULT 1
    );


CREATE TABLE IF NOT EXISTS users_vendors
(
    user_id       bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    vendor_id bigint NOT NULL REFERENCES vendors ON DELETE CASCADE,
    PRIMARY KEY (user_id, vendor_id)
);
