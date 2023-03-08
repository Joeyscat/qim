.PHONY: build mockgen example

build:
	go build -o build/qim .

mockgen:
	mockgen --source server.go -package qim -destination server_mock.go
	# mockgen --source storage.go -package qim -destination storage_mock.go
	# mockgen --source dispatcher.go -package qim -destination dispatcher_mock.go

buildexample:
	go build -o build/examples examples/main.go
