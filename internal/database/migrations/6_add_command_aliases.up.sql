CREATE TABLE IF NOT EXISTS command_alias (
    guild_id TEXT NOT NULL,
    alias TEXT NOT NULL,
    command TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (guild_id, alias)
);

CREATE OR REPLACE FUNCTION command_alias_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_command_alias_set_updated_at ON command_alias;
CREATE TRIGGER trg_command_alias_set_updated_at
BEFORE UPDATE ON command_alias
FOR EACH ROW EXECUTE FUNCTION command_alias_set_updated_at();
