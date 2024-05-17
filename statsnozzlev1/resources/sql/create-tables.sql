drop table if exists statsnozzle;
create table statsnozzle
(
    id     integer   not null primary key,
    time   timestamp not null,
    ip char(64) not null,
    peer_type char(8) not null,
    method  char(8) not null,
    status_code integer not null,
    content_length integer not null,
    uri char(512) not null,
    remote char(64) not null,
    remote_port char(8) not null,
    forwarded char(128) not null,
    useragent char(256) not null
);
