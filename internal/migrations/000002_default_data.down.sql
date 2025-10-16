begin;

delete from players where id = 'e90ea9ec-080a-401b-8d10-6a53c407ac53' and data = '{"default": "data"}';
delete from cosmetics where id = 'default' and version = 1 and data = '{}';

commit;
