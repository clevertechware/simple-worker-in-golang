.PHONY: clean build run-basic run-waitgroup run-context-aware run-advanced

clean:
	@rm -rf bin/

build:
	go build -o bin/basic 01-basic/main.go
	go build -o bin/waitgroup 02-waitgroup/main.go
	go build -o bin/context-aware 03-context-aware/main.go
	go build -o bin/advanced 04-advanced/main.go

run-basic:
	go run 01-basic/main.go

run-waitgroup:
	go run 02-waitgroup/main.go

run-context-aware:
	go run 03-context-aware/main.go

run-advanced:
	go run 04-advanced/main.go