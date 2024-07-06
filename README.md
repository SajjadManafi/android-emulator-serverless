# android-emulator-serverless
android emulator serverless project


### dependencies

- [serverless](https://www.npmjs.com/package/serverless)
- [serverless offline](https://www.npmjs.com/package/serverless-offline)
- [AWS Lambda for Go](https://github.com/aws/aws-lambda-go)


### install dependencies

```
npm install -g serverless
npm install
go mod tidy
docker pull public.ecr.aws/lambda/go
make redis-up
```

### build

```
make build
```

### run service
```
make start-docker-handler
make start
```

### simple ping request

```
 curl http://0.0.0.0:3000/ping
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

### simple get user request
```
curl -X GET http://0.0.0.0:3000/getUser \
     -H "Authorization:YOUR_ACCESS_TOKEN" 
```



### simple register device request
```
curl -X POST http://0.0.0.0:3000/registerDevice \
     -H "Content-Type: application/json" \
     -H "Authorization:YOUR_ACCESS_TOKEN" \
     -d '{
           "android_api": "API_LEVEL",
           "device_name": "DEVICE_NAME"
         }'
```


### simple get device request
```
curl -X GET http://0.0.0.0:3000/getDevice \
     -H "Authorization:YOUR_ACCESS_TOKEN" 
```

### simple delete device request
```
curl -X DELETE http://0.0.0.0:3000/deleteDevice \
     -H "Authorization:YOUR_ACCESS_TOKEN"
```


### TODO:


- [ ] Web Interface
- [ ] TLS
- [X] Emulator Auth (Reverse Proxy)
- [ ] Edit Emulator docker images
- [ ] Write Documentation
- [ ] Write Tests
- [ ] Write comments for code
- [ ] Deploy 
