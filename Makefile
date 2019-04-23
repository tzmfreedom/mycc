.PHONY: test
test:
	./test.sh

run:
	./main

main: main.s
	gcc -o main main.s

mycc: main.go
	go build -o mycc main.go
