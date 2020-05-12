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

create or replace procedure sysinfo_set(_usr varchar, _info jsonb, _modify_date timestamptz, _create_date timestamptz)
language plpgsql
as $$
BEGIN
	insert into sysinfo(usr, info, modify_date, create_date) values(_usr, _info, _modify_date, _create_date);
exception when unique_violation then
	update sysinfo set info = _info, modify_date = _modify_date where usr = _usr;
END $$;