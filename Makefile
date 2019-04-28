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
	echo "$(CODE)" > tmp/tmp.c
	./mycc tmp/tmp.c > tmp/tmp.s
	gcc -o tmp/tmp tmp/tmp.s
	./tmp/tmp

tmp/hoge.o: tmp/hoge.c
	gcc -c tmp/hoge.c -o tmp/hoge.o
