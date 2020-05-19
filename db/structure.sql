-- needs postgresql 9.1
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

create table if not exists note (
    id uuid,
    taker varchar not null,
    about varchar not null,
    content text not null,
    create_date timestamptz,
    primary key (id)
);

create table if not exists warn (
    id uuid,
    moderator varchar not null,
    usr varchar not null,
    reason varchar,
    create_date timestamptz,
    primary key (id)
);

create table if not exists sysinfo (
    usr varchar,
    info jsonb,
    modify_date timestamptz,
    create_date timestamptz,
    primary key (usr)
);

create table if not exists profile (
    usr varchar,
    git varchar,
    dotfiles varchar,
    description varchar,
    primary key (usr)
);

create or replace procedure sysinfo_set(_usr varchar, _info jsonb, _modify_date timestamptz, _create_date timestamptz)
language plpgsql
as $$
BEGIN
	insert into sysinfo(usr, info, modify_date, create_date) values(_usr, _info, _modify_date, _create_date);
exception when unique_violation then
	update sysinfo set info = _info, modify_date = _modify_date where usr = _usr;
END $$;

create or replace procedure profile_set(_usr varchar, _git varchar, _dotfiles varchar, _description varchar)
language plpgsql
as $$
BEGIN
	insert into profile(usr, git, dotfiles, description) values(_usr, _git, _dotfiles, _description);
exception when unique_violation then
	update profile set git = _git, dotfiles = _dotfiles, description = _description where usr = _usr;
END $$;
