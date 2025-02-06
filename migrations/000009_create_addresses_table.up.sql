CREATE TABLE IF NOT EXISTS addresses
(
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    country text NOT NULL,
    name text NOT NULL,
    organization text NOT NULL,
    street_address text[] NOT NULL,
    locality text NOT NULL,
    administrative_area text NOT NULL,
    post_code text NOT NULL,
    sorting_code text NOT NULL,
    data jsonb NOT NULL,
    lat float NOT NULL,
    lng float NOT NULL

)