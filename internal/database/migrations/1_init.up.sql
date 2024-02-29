create table if not exists aquarium
(
    user_id    text not null primary key,
    common     integer default 0 not null,
    uncommon   integer default 0 not null,
    rare       integer default 0 not null,
    super_rare integer default 0 not null,
    legendary  integer default 0 not null
);

create table if not exists guild
(
    guild_id               text not null primary key,
    use_warns              boolean default false not null,
    max_warns              integer default 3     not null,
    warn_duration          integer default 30    not null,
    automod_log_channel_id text    default ''    not null,
    fishing_channel_id     text    default ''    not null
);

create table if not exists command_log
(
    uid        serial primary key,
    command    text not null,
    args       text not null,
    user_id    text not null,
    guild_id   text
        references guild,
    channel_id text not null,
    message_id text not null,
    sent_at    timestamp with time zone not null
);

create table if not exists filter
(
    uid      serial primary key,
    guild_id text not null
        references guild,
    phrase   text not null
);

create table if not exists warn
(
    uid           serial primary key,
    guild_id      text not null
        constraint warn_guild_id
            references guild,
    user_id       text not null,
    reason        text not null,
    given_by_id   text not null,
    given_at      timestamp with time zone not null,
    is_valid      boolean default true not null,
    cleared_by_id text,
    cleared_at    timestamp with time zone
);

create table if not exists user_role
(
    uid      serial primary key,
    guild_id text not null
        references guild,
    user_id  text not null,
    role_id  text not null
);
