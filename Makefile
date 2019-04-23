.PHONY: test
test: mycc
	./test.sh

run:
	./main

main: main.s
	gcc -o main main.s

mycc: main.go
	gofmt -w .
	go build -o mycc main.go
