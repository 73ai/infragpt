create table integration (
    id uuid primary key,
    provider varchar(36) not null,
    status varchar(36) not null,
    business_id uuid not null,
    provider_project_id varchar(50) not null,
    active boolean not null default true,
    created_at timestamp not null default now()
);