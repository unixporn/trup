{ pkgs ? import <nixpkgs> {} }:
pkgs.stdenv.mkDerivation {
	name = "trup-devenv";
	buildInputs = with pkgs; [
		postgresql
		go
	];
	shellHook = ''
		export GOPATH=$HOME/go
		export PGDATA=$PWD/postgres_data
		export PGHOST=$PWD/postgres
		export LOG_PATH=$PWD/postgres/LOG
		export PGDATABASE=trup
		export DATABASE_URL="postgresql:///postgres/?host=$PGHOST&database=$PGDATABASE"
		. .env 2>/dev/null >/dev/null
		if [ ! -d $PGHOST ]; then
			mkdir -p $PGHOST
		fi
		if [ ! -d $PGDATA ]; then
			echo 'Initializing postgresql database...'
			initdb $PGDATA --auth=trust >/dev/null
		fi
		pg_ctl stop -D $PWD/postgres_data 2>/dev/null >/dev/null
		setsid pg_ctl start -l $LOG_PATH -o "-c listen_addresses= -c unix_socket_directories=$PGHOST"
		createdb trup
		psql trup <$PWD/db/structure.sql
	'';

  # needed by initdb on non-NixOS systems
	LOCALE_ARCHIVE = if pkgs.stdenv.isLinux then "${pkgs.glibcLocales}/lib/locale/locale-archive" else "";
}
