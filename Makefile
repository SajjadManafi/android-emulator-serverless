clean:
	@go clean
	@rm -rf ./bin

build: clean
	GOOS=linux GOARCH=amd64 go build -o bin/ping functions/ping/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/register functions/register/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/login functions/login/main.go

start:
	sudo sls offline --useDocker start --host 0.0.0.0

redis:
	docker run --name redis-db -d -p 6379:6379 redis 