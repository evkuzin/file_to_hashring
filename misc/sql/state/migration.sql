create table metadata
(
    name varchar not null
        primary key
        unique,
    size bigint,
    nodes int,
    content_type varchar
);
