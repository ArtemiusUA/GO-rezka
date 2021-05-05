create table genres
(
    id serial not null
        constraint genres_pkey
            primary key,
    name varchar(256)
        constraint genres_name_key
            unique
);

alter table genres owner to postgres;

create table videos
(
    id serial not null
        constraint videos_pkey
            primary key,
    name varchar(256)
        constraint videos_name_key
            unique,
    name_orig varchar(256),
    url text,
    image_url text,
    description text,
    rating double precision,
    video_urls json
);

alter table videos owner to postgres;

create table videos_genres
(
    video_id integer not null
        constraint videos_genres_video_id_fkey
            references videos,
    genre_id integer not null
        constraint videos_genres_genre_id_fkey
            references genres,
    unique(video_id, genre_id)
);

alter table videos_genres owner to postgres;

