alter table guild
	add column auto_role_id text default '' not null;
alter table user_role
    rename to custom_role;