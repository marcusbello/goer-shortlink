
gen_grpc:
	protoc -I ./proto --go_out=. --go-grpc_out=. ./proto/urlshortener.proto

run:
	go run server/server.go
