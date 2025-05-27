create table if not exists device (
    device_id uuid not null,
    user_id uuid not null,
    device_fingerprint text not null unique,
    name varchar(255) not null,
    os varchar(255) not null,
    brand varchar(255) not null,
    primary key (device_id)
);

create table if not exists user_session (
    user_id uuid not null,
    device_id uuid not null,
    session_id uuid not null,
    user_agent text not null,
    ip_address text not null,
    ip_country_iso varchar(2) not null,
    last_activity_at timestamptz not null default now(),
    created_at timestamptz not null default now(),
    timezone text not null,
    is_expired boolean not null default false,
    primary key (session_id)
);

create table if not exists refresh_token (
     token_id uuid not null,
     user_id uuid not null,
     session_id uuid not null,
     token_hash text not null unique,
     expiry_at timestamptz not null,
     created_at timestamptz not null default now(),
    revoked boolean not null default false,
    primary key (token_id)
    );

create index if not exists refresh_token_user_id_idx on refresh_token(user_id);