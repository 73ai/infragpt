create table if not exists users (
    user_id uuid not null,
    email varchar(255) not null,
    password_hash text not null,
    is_email_verified boolean not null default false,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (user_id)
);

create table if not exists email_verification (
    verification_id uuid not null,
    user_id uuid not null,
    email varchar(255) not null,
    expiry_at timestamptz not null,
    created_at timestamptz not null default now(),
    primary key (verification_id)
);

create table if not exists reset_password (
    reset_id uuid not null,
    user_id uuid not null,
    expiry_at timestamptz not null,
    created_at timestamptz not null default now(),
    primary key (reset_id)
);
