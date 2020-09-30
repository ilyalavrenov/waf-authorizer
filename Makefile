build:
	dep ensure -v
	env GOOS=linux go build -v -ldflags="-d -s -w" -a -tags netgo -installsuffix netgo -o bin/redirect redirect/main.go
	env GOOS=linux go build -v -ldflags="-d -s -w" -a -tags netgo -installsuffix netgo -o bin/generate generate/main.go
	env GOOS=linux go build -v -ldflags="-d -s -w" -a -tags netgo -installsuffix netgo -o bin/allowlist allowlist/main.go
	env GOOS=linux go build -v -ldflags="-d -s -w" -a -tags netgo -installsuffix netgo -o bin/cleanup cleanup/main.go