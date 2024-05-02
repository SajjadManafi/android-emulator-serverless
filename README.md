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