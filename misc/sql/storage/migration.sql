create table files
(
    name varchar not null
        primary key
        unique,
    size bigint,
    data bytea   not null
);
