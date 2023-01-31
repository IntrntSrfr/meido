create table aquarium
(
    user_id text primary key,
    common int not null default 0,
    uncommon int not null default 0,
    rare int not null default 0,
    super_rare int not null default 0,
    legendary int not null default 0
);

create table command_log
(
    id         serial primary key,
    command    text                     not null,
    args       text                     not null,
    user_id    text                     not null,
    guild_id   text                     not null default '',
    channel_id text                     not null,
    message_id text                     not null,
    sent_at    timestamp with time zone not null
);

create table user_role
(
    id       serial primary key,
    guild_id text not null,
    user_id  text not null,
    role_id  text not null
);

create table filter
(
    id       serial primary key,
    guild_id text not null,
    phrase   text not null
);

create table auto_role
(
    id       serial primary key,
    guild_id text not null,
    role_id  text not null,
    enabled  boolean default true
);

create table warn
(
    id            serial primary key,
    guild_id      text                     not null,
    user_id       text                     not null,
    reason        text                     not null,
    given_by_id   text                     not null,
    given_at      timestamp with time zone not null,
    is_valid      boolean default true,
    cleared_by_id text,
    cleared_at    timestamp with time zone
);

create table guild
(
    guild_id            text primary key,
    use_warns           boolean not null default false,
    max_warns           integer not null default 3,
    warn_duration       integer not null default 30,
    automod_log_channel text    not null default '',
    fishing_channel     text    not null default ''
);

