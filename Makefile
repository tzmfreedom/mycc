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

test/run: mycc
	./mycc "$(CODE)" > tmp/tmp.s
	gcc -o tmp/tmp tmp/tmp.s
	./tmp/tmp
