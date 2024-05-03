# android-emulator-serverless
android emulator serverless project


### dependencies

- serverless
- serverless offline
- AWS Lambda for Go


### install dependencies

```
npm install -g serverless
npm install
go mod tidy
docker pull public.ecr.aws/lambda/go
```

### build

```
make build
```

### run service
```
make start
```

### simple ping request

```
 curl http://0.0.0.0.0:3000/ping
```


### simple register request
```
curl -X POST http://0.0.0.0:3000/register \
     -H "Content-Type: application/json" \
     -d '{"name":"sajjad","username":"sajjadma","password":"testpass"}'
```

### simple login request
```
curl -i -X POST http://0.0.0.0:3000/login \
     -H "Content-Type: application/json" \
     -d '{"username":"sajjadma","password":"testpass"}'
```