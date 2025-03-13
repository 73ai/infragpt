create table slack_token (
    token_id uuid primary key,
    team_id varchar(36) not null,
    token text not null,
    expired boolean not null default false,
    expired_at timestamp with time zone,
    created_at timestamp with time zone not null default now()
);