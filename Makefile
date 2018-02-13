build:
	dep ensure
	env GOOS=linux go build -ldflags="-s -w" -o bin/hello hello/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/upload upload/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/presignurl presignurl/main.go
