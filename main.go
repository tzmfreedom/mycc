package main

import (
	"fmt"
	"github.com/k0kubun/pp"
	"io/ioutil"
	"os"
)

func main() {
	content, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	parse(string(content))
}

func parse(str string) {
	l := NewLexer(str)
	tokens := l.Tokenize(str)
	p := NewParser(tokens)
	declarations := p.Parse()
	g := NewGenerator()
	g.generate(declarations)
}

const MemorySize = 8

var registerIndex = []string{
	"rdi",
	"rsi",
	"rdx",
	"rcx",
	"r8",
	"r9",
}

type Generator struct {
	Variables map[string]int
}

func NewGenerator() *Generator {
	return &Generator{
		Variables: map[string]int{},
	}
}

func (g *Generator) checkVariable(n Node) {
	switch node := n.(type) {
	case *BinaryOperatorNode:
		if node.Type == '=' {
			ident := node.Left.(*IdentifierNode).Value
			g.Variables[ident] = len(g.Variables)
		}
		g.checkVariable(node.Left)
		g.checkVariable(node.Right)
	}
}

func (g *Generator) VisitInteger(n *IntegerNode) (interface{}, error) {
	fmt.Printf("    push %d\n", n.Value)
	return nil, nil
}

func (g *Generator) VisitString(n *StringNode) (interface{}, error) {
	return nil, nil
}

func (g *Generator) VisitBinaryOperator(n *BinaryOperatorNode) (interface{}, error) {
	switch n.Type {
	case '+', '-', '*', '/':
		n.Left.Accept(g)
		n.Right.Accept(g)
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
		ident, ok := n.Left.(*IdentifierNode)
		if !ok {
			panic("no identifier")
		}
		index := g.Variables[ident.Value]
		fmt.Printf("    mov rax, rbp\n")
		fmt.Printf("    sub rax, %d\n", (index+1)*MemorySize)
		fmt.Printf("    push rax\n")
		n.Right.Accept(g)
		fmt.Printf("    pop rdi\n")
		fmt.Printf("    pop rax\n")
		fmt.Printf("    mov [rax], rdi\n")
		fmt.Printf("    push rdi\n")
	case ND_EQUAL:
		n.Left.Accept(g)
		n.Right.Accept(g)
		fmt.Printf("    pop rax\n")
		fmt.Printf("    pop rdi\n")
		fmt.Printf("    cmp rdi, rax\n")
		fmt.Printf("    sete al\n")
		fmt.Printf("    movzx rax, al\n")
		fmt.Printf("    push rax\n")
	case ND_NOTEQUAL:
		n.Left.Accept(g)
		n.Right.Accept(g)
		fmt.Printf("    pop rax\n")
		fmt.Printf("    pop rdi\n")
		fmt.Printf("    cmp rdi, rax\n")
		fmt.Printf("    setne al\n")
		fmt.Printf("    movzx rax, al\n")
		fmt.Printf("    push rax\n")
	}
	return nil, nil
}

func (g *Generator) VisitReturn(n *ReturnNode) (interface{}, error) {
	n.Expression.Accept(g)
	fmt.Printf("    pop rax\n")
	fmt.Printf("    mov rsp, rbp\n")
	fmt.Printf("    pop rbp\n")
	fmt.Printf("    ret\n")
	return nil, nil
}

func (g *Generator) VisitFunction(n *FunctionNode) (interface{}, error) {
	return nil, nil
}

func (g *Generator) generate(declarations []Node) {
	fmt.Println(`
.intel_syntax noprefix
.global _main`)
	for _, declaration := range declarations {
		if f, ok := declaration.(*FunctionNode); ok {
			fmt.Printf("\n")
			fmt.Printf("_%s:\n", f.Identifier)
			fmt.Printf("    push rbp\n")
			fmt.Printf("    mov rbp, rsp\n")

			for i, p := range f.Parameters {
				g.Variables[p.Identifier] = i
			}
			for _, stmt := range f.Statements {
				g.checkVariable(stmt)
			}
			fmt.Printf("    sub rsp, %d\n", len(g.Variables)*MemorySize)

			for i, _ := range f.Parameters {
				fmt.Printf("    mov rax, rbp\n")
				fmt.Printf("    sub rax, %d\n", (i+1)*MemorySize)
				fmt.Printf("    mov [rax], %s\n", registerIndex[i])
			}

			for _, stmt := range f.Statements {
				stmt.Accept(g)
				fmt.Printf("    pop rax\n")
			}
			//fmt.Printf("    mov rsp, rbp\n")
			//fmt.Printf("    pop rbp\n")
			//fmt.Printf("    ret\n")
		}
	}
}

func (g *Generator) VisitIdentifier(n *IdentifierNode) (interface{}, error) {
	index := g.Variables[n.Value]
	fmt.Printf("    mov rax, rbp\n")
	fmt.Printf("    sub rax, %d\n", (index+1)*MemorySize)
	fmt.Printf("    push [rax]\n")
	return nil, nil
}

func (g *Generator) VisitCall(n *CallNode) (interface{}, error) {
	if len(n.Args) > 0 {
		n.Args[0].Accept(g)
		fmt.Printf("    pop rdi\n")
	}
	if len(n.Args) > 1 {
		n.Args[1].Accept(g)
		fmt.Printf("    pop rsi\n")
	}
	if len(n.Args) > 2 {
		n.Args[2].Accept(g)
		fmt.Printf("    pop rdx\n")
	}
	if len(n.Args) > 3 {
		n.Args[3].Accept(g)
		fmt.Printf("    pop rcx\n")
	}
	if len(n.Args) > 4 {
		n.Args[4].Accept(g)
		fmt.Printf("    pop r8\n")
	}
	if len(n.Args) > 5 {
		n.Args[5].Accept(g)
		fmt.Printf("    pop r9\n")
	}
	fmt.Printf("    call _%s\n", n.Identifier)
	fmt.Printf("    push rax\n")
	return nil, nil
}

func debug(args ...interface{}) {
	pp.Println(args...)
}
