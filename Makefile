clean:
	@go clean
	@rm -rf ./bin

build: clean
	GOOS=linux GOARCH=amd64 go build -o bin/ping functions/ping/main.go

start:
	sudo sls offline --useDocker start --host 0.0.0.0