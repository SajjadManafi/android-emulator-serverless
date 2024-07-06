clean:
	@go clean
	@rm -rf ./bin

build: clean
	GOOS=linux GOARCH=amd64 go build -o bin/ping functions/ping/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/register functions/register/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/login functions/login/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/registerDevice functions/registerDevice/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/getDevice functions/getDevice/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/deleteDevice functions/deleteDevice/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/getUser functions/getUser/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/updateUser functions/updateUser/main.go

start:
	sudo sls offline --useDocker start --host 0.0.0.0

start-docker-handler:
	go run cmd/main.go

redis-up:
	docker run --name redis-db -d -p 6379:6379 redis 