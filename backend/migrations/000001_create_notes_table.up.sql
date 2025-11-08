CREATE TABLE IF NOT EXISTS notes (
    id bigserial PRIMARY KEY,  
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    body text NOT NULL,
    archived boolean NOT NULL DEFAULT FALSE,
    tags text[] NOT NULL,
    version integer NOT NULL DEFAULT 1
);