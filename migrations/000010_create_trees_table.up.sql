CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


CREATE TABLE IF NOT EXISTS trees
(
    id uuid DEFAULT uuid_generate_v4(),
    project varchar NOT NULL UNIQUE,
    data jsonb,
    PRIMARY KEY (id)

)