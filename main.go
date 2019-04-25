package main

import (
	"fmt"
	"github.com/k0kubun/pp"
	"os"
)

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
	nodes := p.Parse()
	g := &Generator{}
	g.generate(nodes)
	pp.Println(nodes)
}

type Generator struct{
	Variables map[string]int
}

func (g *Generator) checkVariable(n *Node) {
	switch n.Type {
	case '+', '-', '*', '/':
		g.checkVariable(n.Left)
		g.checkVariable(n.Right)
	case '=':
		g.Variables[n.Left.Value] = len(g.Variables)
	}
}

func (g *Generator) generate(nodes []*Node) {
	for _, n := range nodes {
		g.checkVariable(n)
	}
	if len(g.Variables) > 0 {
		fmt.Printf("    push rbp\n")
		fmt.Printf("    mov rbp, rsp\n")
		fmt.Printf("    add rsp, %d\n", len(g.Variables) * 8)
	}
	for _, n := range nodes {
		g.gen(n)
	}
	fmt.Printf("    pop rax\n")
	fmt.Printf("    mov rsp, rbp\n")
	fmt.Printf("    pop rbp\n")
	fmt.Printf("    ret\n")
}

func (g *Generator) gen(n *Node) {
	switch n.Type {
	case '+', '-', '*', '/':
		g.gen(n.Left)
		g.gen(n.Right)
		fmt.Printf("    pop rdi\n")
		fmt.Printf("    pop rax\n")
		if n.Type == '/' {
			fmt.Printf("    mov rdx, 0\n")
			fmt.Printf("    div rdi\n")
		} else if n.Type == '*' {
			fmt.Printf("    mul rdi\n")
		} else {
			fmt.Printf("    %s rax, rdi\n", n.Type)
		}
		fmt.Printf("    push rax\n")
	case '=':
		fmt.Printf("    push rax\n")
	default:
		fmt.Printf("    push %s\n", n.Value)
	}
}
