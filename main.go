package main

import (
	"fmt"
	"os"
)

func main() {
	ret := os.Args[1]
	fmt.Printf(`
.intel_syntax noprefix
.global _main
_main:
    mov rax, %s
    ret
`, ret)
}
