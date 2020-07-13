package pb

//go:generate sh -c "protoc -I../../proto --go_out=plugins=grpc:. --validate_out=\"lang=go:.\" ../../proto/*.proto"
