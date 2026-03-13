DB_URL=postgresql://pintour:pintour@localhost:5432/pintour_db?sslmode=disable

postgres:
	docker run --name pintour_postgres -p 5432:5432 -e POSTGRES_USER=pintour -e POSTGRES_PASSWORD=pintour -e POSTGRES_DB=pintour_db -d postgres:16-alpine

createdb:
	docker exec -it pintour_postgres createdb --username=pintour --owner=pintour pintour_db

dropdb:
	docker exec -it pintour_postgres dropdb pintour_db

migrateup:
	migrate -path internal/db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path internal/db/migration -database "$(DB_URL)" -verbose down

sqlc:
	~/go/bin/sqlc generate

proto:
	rm -f pb/pintour/v1/*.go
	rm -f docs/swagger/*.json
	protoc \
		--proto_path=proto \
		--proto_path=third_party/googleapis \
		--proto_path=third_party/grpc-gateway \
		--proto_path=/usr/include \
		--go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=docs/swagger --openapiv2_opt=allow_merge=true,merge_file_name=pintour \
		pintour/v1/common.proto \
		pintour/v1/auth.proto \
		pintour/v1/tour.proto \
		pintour/v1/booking.proto \
		pintour/v1/payment.proto \
		pintour/v1/review.proto
	export PATH=$$PATH:~/go/bin

server:
	go run ./cmd/server/

build:
	go build -o bin/server ./cmd/server/

test:
	go test -v -cover ./...

.PHONY: postgres createdb dropdb migrateup migratedown sqlc proto server build test
