create table if not exists filters
(
    uid      serial        not null
        constraint filter_pkey
            primary key,
    guild_id text          not null,
    phrase   varchar(2000) not null
);

create table if not exists commandlog
(
    uid        serial        not null
        constraint commandlog_pkey
            primary key,
    command    varchar(100)  not null,
    args       varchar(2000) not null,
    user_id    text          not null,
    guild_id   text          not null,
    channel_id text          not null,
    message_id text          not null,
    tstamp     timestamp     not null
);

create table if not exists guilds
(
    uid         serial  not null
        constraint discordguild_pkey
            primary key,
    guild_id    text    not null,
    use_strikes boolean not null,
    max_strikes integer not null
);

create table if not exists userroles
(
    uid      serial not null
        constraint userrole_pkey
            primary key,
    guild_id text   not null,
    user_id  text   not null,
    role_id  text   not null
);

create table if not exists warns
(
    uid           serial               not null
        constraint strike_pkey
            primary key,
    guild_id      text                 not null,
    user_id       text                 not null,
    reason        varchar(2000)        not null,
    given_by_id   text                 not null,
    given_at      timestamp            not null,
    is_valid      boolean default true not null,
    cleared_by_id text,
    cleared_at    timestamp
);

create table if not exists autorole
(
    id      serial        not null
        constraint autorole_pkey
            primary key,
    guild_id text          not null,
    role_id text not null,
    enabled boolean not null
);