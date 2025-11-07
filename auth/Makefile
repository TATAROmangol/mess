generate-proto:
	protoc -I ../api --go_out=internal/grpc --go-grpc_out=internal/grpc ../api/auth.proto

tests:
	cd .. && docker compose down && docker compose up && cd auth && go test -v ./...

