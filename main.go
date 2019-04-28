package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(`
.intel_syntax noprefix
.global _main
_main:`)
	parse(os.Args[1])
	//fmt.Println("    pop rax")
	//fmt.Println("    ret")
}

func parse(str string) {
	l := NewLexer(str)
	tokens := l.Tokenize(str)
	p := NewParser(tokens)
	nodes := p.Parse()
	g := NewGenerator()
	g.generate(nodes)
	// pp.Println(nodes)
}

const MemorySize = 8

type Generator struct {
	Variables map[string]int
}

func NewGenerator() *Generator {
	return &Generator{
		Variables: map[string]int{},
	}
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
	fmt.Printf("    push rbp\n")
	fmt.Printf("    mov rbp, rsp\n")
	fmt.Printf("    sub rsp, %d\n", len(g.Variables)*MemorySize)
	for _, n := range nodes {
		g.gen(n)
		fmt.Printf("    pop rax\n")
	}
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
		} else if n.Type == '+' {
			fmt.Printf("    add rax, rdi\n")
		} else if n.Type == '-' {
			fmt.Printf("    sub rax, rdi\n")
		} else {
			fmt.Printf("    %s rax, rdi\n", string(rune(n.Type)))
		}
		fmt.Printf("    push rax\n")
	case '=':
		index := g.Variables[n.Left.Value]
		fmt.Printf("    mov rax, rbp\n")
		fmt.Printf("    sub rax, %d\n", index*MemorySize)
		fmt.Printf("    push rax\n")
		g.gen(n.Right)
		fmt.Printf("    pop rdi\n")
		fmt.Printf("    pop rax\n")
		fmt.Printf("    mov [rax], rdi\n")
		fmt.Printf("    push rdi\n")
	case ND_NUMBER:
		fmt.Printf("    push %s\n", n.Value)
	case ND_IDENT:
		index := g.Variables[n.Value]
		fmt.Printf("    mov rax, rbp\n")
		fmt.Printf("    sub rax, %d\n", index*MemorySize)
		fmt.Printf("    push [rax]\n")
	case ND_RETURN:
		g.gen(n.Left)
		fmt.Printf("    pop rax\n")
		fmt.Printf("    mov rsp, rbp\n")
		fmt.Printf("    pop rbp\n")
		fmt.Printf("    ret\n")
	default:
		fmt.Printf("    push %s\n", n.Value)
	}
}
