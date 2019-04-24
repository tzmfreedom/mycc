package main

import (
	"fmt"
	"os"
)

type Token struct {
	Type   string
	Value  string
	Line   int
	Column int
}

func main() {
	fmt.Println(`
.intel_syntax noprefix
.global _main
_main:`)
	parse(os.Args[1])
	fmt.Println("    pop rax")
	fmt.Println("    ret")
}

func parse(str string) {
	l := NewLexer(str)
	tokens := l.Tokenize(str)
	p := NewParser(tokens)
	tree := p.Parse()
	generate(tree)
}

func generate(n *Node) {
	switch n.Type {
	case "add", "sub", "mul", "div":
		generate(n.Left)
		generate(n.Right)
		fmt.Printf("    pop rdi\n")
		fmt.Printf("    pop rax\n")
		if n.Type == "div" {
			fmt.Printf("    mov rdx, 0\n")
			fmt.Printf("    div rdi\n")
		} else if n.Type == "mul" {
			fmt.Printf("    mul rdi\n")
		} else {
			fmt.Printf("    %s rax, rdi\n", n.Type)
		}
		fmt.Printf("    push rax\n")
	default:
		fmt.Printf("    push %s\n", n.Value)
	}
}
