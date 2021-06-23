dev: buffalo adminer migrate

migrate: db
	docker-compose run --rm buffalo whenavail db 5432 10 buffalo-pop pop migrate up
	docker-compose run --rm buffalo /bin/bash -c "grift private:seed && grift db:seed && grift minio:seed"

migratestatus: db
	docker-compose run buffalo buffalo-pop pop migrate status

migratetestdb: testdb
	docker-compose run --rm test whenavail testdb 5432 10 buffalo-pop pop migrate up

gqlgen:
	-docker-compose pause buffalo
	-docker-compose run --rm buffalo /bin/bash -c "go generate ./gqlgen"
	-docker-compose unpause buffalo

adminer:
	docker-compose up -d adminer

buffalo: db
	docker-compose up -d buffalo

swagger: swaggerspec
	docker-compose run --rm --service-ports swagger serve -p 8082 --no-open swagger.json

swaggerspec:
	docker-compose run --rm swagger generate spec -m -o swagger.json

bounce: db
	docker-compose kill buffalo
	docker-compose rm -f buffalo
	docker-compose up -d buffalo

logs:
	docker-compose logs buffalo

db:
	docker-compose up -d db

testdb:
	docker-compose up -d testdb

test:
	docker-compose run --rm test whenavail testdb 5432 10 buffalo test

testenv: migratetestdb
	@echo "\n\nIf minio hasn't been initialized, run grift minio:seed\n"
	docker-compose run --rm test bash

clean:
	docker-compose kill
	docker-compose rm -f

fresh: clean dev
