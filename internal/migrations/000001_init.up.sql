begin;

create table if not exists cosmetics
(
    id      varchar primary key
        constraint resource_location_like check ( id ~ '^[a-z\_\-0-9.]+$' ),
    version int
        constraint positive_version check ( version > 0 ),
    data    json not null
);

create table if not exists players
(
    id   uuid primary key,
    data json not null default '{}'
);

create table if not exists player_cosmetics
(
    player_id   uuid    not null references players (id) on delete cascade,
    cosmetic_id varchar not null references cosmetics (id) on delete cascade,
    constraint player_cosmetic_pair unique (player_id, cosmetic_id)
);

create view players_with_cosmetics as
select players.id                                         as player_id,
       players.data                                       as player_data,
       coalesce(array_agg(cosmetics.id), '{}'::varchar[]) as cosmetics
from players
         left join player_cosmetics on players.id = player_cosmetics.player_id
         left join cosmetics on cosmetics.id = player_cosmetics.cosmetic_id
group by players.id;

commit;