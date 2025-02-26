CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


CREATE TABLE IF NOT EXISTS trees
(
    id      uuid DEFAULT uuid_generate_v4(),
    name    text UNIQUE NOT NULL,
    hash    text UNIQUE NOT NULL,
    profile varchar     NOT NULL,
    data    jsonb,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)

)