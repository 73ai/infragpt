create table google_state_token (
    token text not null,
    revoked boolean not null default false,
    revoked_at timestamp with time zone,
    expires_at timestamp with time zone not null,
    created_at timestamp with time zone not null default now()

);