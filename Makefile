.PHONY: build buildexample mockgen proto

build:
	go build -o build/qim .

buildexample:
	go build -o build/examples examples/main.go

mockgen:
	mockgen --source server.go -package qim -destination server_mock.go
	# mockgen --source storage.go -package qim -destination storage_mock.go
	# mockgen --source dispatcher.go -package qim -destination dispatcher_mock.go

proto:
	protoc -I wire/proto/ --go_out=./wire/ wire/proto/*.proto
