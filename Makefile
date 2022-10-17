build:
	go build -o ./.bin/test_go_server cmd/app/main.go

run: build
	./.bin/test_go_server

test_run: build
	./.bin/test_go_server -p 8080  -d localhost:8081,localhost:8082