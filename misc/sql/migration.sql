create table files
(
    name varchar not null
        primary key
        unique,
    size int,
    data bytea   not null
);

