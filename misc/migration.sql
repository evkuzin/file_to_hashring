create table files
(
    name varchar not null
        primary key
        unique,
    filesize int,
    data bytea   not null
);

