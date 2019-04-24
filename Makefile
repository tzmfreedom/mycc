SOURCES  = $(wildcard *.go)

.PHONY: test
test: mycc
	./test.sh

run:
	./main

main: main.s
	gcc -o main main.s

mycc: $(SOURCES)
	gofmt -w .
	go build -o mycc .
