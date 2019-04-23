package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println(`
.intel_syntax noprefix
.global _main
_main:`)
	parse(os.Args[1])
	fmt.Println("    ret")
}

func parse(str string) {
	prev := ""
	tokens := strings.Split(str, " ")
	for _, token := range tokens {
		switch token {
		case "+":
			prev = "+"
		case "-":
			prev = "-"
		default:
			if prev == "+" {
				fmt.Printf("    add rax, %s\n", token)
			} else if prev == "-" {
				fmt.Printf("    sub rax, %s\n", token)
			} else {
				fmt.Printf("    mov rax, %s\n", token)
			}
		}
	}
}
