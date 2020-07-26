-- needs postgresql 9.1
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS note (
    id uuid,
    taker varchar not null,
    about varchar not null,
    content text not null,
    create_date timestamptz,
    primary key (id)
);

CREATE TABLE IF NOT EXISTS warn (
    id uuid,
    moderator varchar not null,
    usr varchar not null,
    reason varchar,
    create_date timestamptz,
    primary key (id)
);

CREATE TABLE IF NOT EXISTS sysinfo (
    usr varchar,
    info jsonb,
    modify_date timestamptz,
    create_date timestamptz,
    primary key (usr)
);

CREATE TABLE IF NOT EXISTS mute (
    id uuid,
    guildid varchar not null,
    moderator varchar not null,
    usr varchar not null,
    start_time timestamptz,
    end_time timestamptz,
    reason varchar,
    active boolean not null,
    primary key (id)
);

CREATE TABLE IF NOT EXISTS profile (
    usr varchar,
    git varchar,
    dotfiles varchar,
    description varchar,
    primary key (usr)
);

CREATE TABLE IF NOT EXISTS blocked_regexes (
    pattern varchar,
    added_by varchar not null,
    primary key (pattern)
);

CREATE TABLE IF NOT EXISTS image_log_files (
	channel_id bigint not null,
	message_id bigint not null,
	attachment_id bigint not null,
	filename varchar not null,
	create_date timestamptz,
	should_delete boolean,
	object_id oid,
	primary key (attachment_id)
);

CREATE OR REPLACE PROCEDURE sysinfo_set(_usr varchar, _info jsonb, _modify_date timestamptz, _create_date timestamptz)
language plpgsql
AS $$
BEGIN
	INSERT INTO sysinfo(usr, info, modify_date, create_date) VALUES(_usr, _info, _modify_date, _create_date);
EXCEPTION WHEN unique_violation THEN
	UPDATE sysinfo SET info = _info, modify_date = _modify_date WHERE usr = _usr;
end $$;

CREATE OR REPLACE PROCEDURE profile_set(_usr varchar, _git varchar, _dotfiles varchar, _description varchar)
language plpgsql
AS $$
BEGIN
	INSERT INTO profile(usr, git, dotfiles, description) VALUES(_usr, _git, _dotfiles, _description);
EXCEPTION WHEN unique_violation THEN
	UPDATE profile SET git = _git, dotfiles = _dotfiles, description = _description WHERE usr = _usr;
END $$;
