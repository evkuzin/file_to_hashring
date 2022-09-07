create table files
(
    name varchar not null
        primary key
        unique,
    size bigint,
    data bytea   not null
);
-- alter system set log_statement to 'all';
-- select pg_reload_conf();