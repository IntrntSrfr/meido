alter table guild
	add column joined_at timestamp with time zone default CURRENT_TIMESTAMP not null;
