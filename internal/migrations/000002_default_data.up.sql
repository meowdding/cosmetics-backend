begin;

insert into players(id, data) values ('e90ea9ec-080a-401b-8d10-6a53c407ac53', '{"default": "data"}') on conflict do nothing;
insert into cosmetics(id, version, data) values ('default', 1, '{}') on conflict do nothing;
insert into player_cosmetics(player_id, cosmetic_id) values ('e90ea9ec-080a-401b-8d10-6a53c407ac53', 'default') on conflict do nothing;

commit;