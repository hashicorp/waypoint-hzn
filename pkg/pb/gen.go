package pb

//go:generate sh -c "protoc -I../../proto --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --validate_out=\"lang=go:.\" --validate_opt=paths=source_relative ../../proto/server.proto"
