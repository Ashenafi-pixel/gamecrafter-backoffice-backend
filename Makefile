# Migration commands using Docker
migrate-up:
	docker-compose run --rm migrate

migrate-down:
	docker-compose run --rm migrate make migrate-down

migrate-create:
	docker-compose run --rm migrate make migrate-create args="$(filter-out $@,$(MAKECMDGOALS))"

migrate-force:
	docker-compose run --rm migrate make migrate-force version="$(filter-out $@,$(MAKECMDGOALS))"

# Other commands
swagger:
	swag fmt && swag init -g cmd/main.go

run:
	docker-compose up app

sqlc:
	cd ./config && sqlc generate

# Add this to handle arguments
%:
	@: