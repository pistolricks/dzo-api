CREATE TABLE IF NOT EXISTS contents
(
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    hash text UNIQUE NOT NULL,
    name text NOT NULL,
    original text NOT NULL,
    src text NOT NULL,
    type text NOT NULL,
    size integer NOT NULL,
    user_id text NOT NULL
)