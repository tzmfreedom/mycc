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
	Variables        map[string]*Variable
	LabelCnt         int
	RspCounter       int
	CurrentLoopBegin string
	CurrentLoopEnd   string
}

type Variable struct {
	Index int
	Type  *Ctype
}

func NewGenerator() *Generator {
	return &Generator{
		Variables:  map[string]*Variable{},
		LabelCnt:   0,
		RspCounter: 0,
	}
}

func (g *Generator) variableStackSize() int {
	stackSize := 0
	for _, v := range g.Variables {
		var size int
		if v.Type.Size%8 == 0 {
			size = v.Type.Size
		} else {
			size = v.Type.Size - v.Type.Size%8 + 8
		}
		stackSize += size
	}
	return stackSize
}

func (g *Generator) generate(declarations []Node) {
	fmt.Println(`
.intel_syntax noprefix
.global _main`)
	for _, declaration := range declarations {
		declaration.Accept(g)
	}
}

func (g *Generator) checkVariable(n Node) {
	switch node := n.(type) {
	case *Block:
		for _, stmt := range node.Statements {
			g.checkVariable(stmt)
		}
	case *While:
		g.checkVariable(node.Expression)
		g.checkVariable(node.Statements)
	case *For:
		g.checkVariable(node.Init)
		g.checkVariable(node.Update)
		g.checkVariable(node.Expression)
		g.checkVariable(node.Statements)
	case *If:
		g.checkVariable(node.IfStatements)
		g.checkVariable(node.Expression)
	case *BinaryOperatorNode:
		if node.Type == '=' {
			switch left := node.Left.(type) {
			case *IdentifierNode:
				if _, ok := g.Variables[left.Value]; !ok {
					panic("no variable declaration: " + left.Value)
				}
			}
		}
		g.checkVariable(node.Right)
	case *VariableDeclaration:
		_, ok := g.Variables[node.Identifier]
		if ok {
			panic("variable redeclaration: " + node.Identifier)
		} else {
			g.Variables[node.Identifier] = &Variable{
				Index: g.variableStackSize() / 8,
				Type:  node.Type,
			}
		}
		g.checkVariable(node.Expression)
	}
}

func (g *Generator) VisitInteger(n *IntegerNode) (interface{}, error) {
	g.generatePush(fmt.Sprintf("%d", n.Value))
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
		if ident, ok := n.Left.(*IdentifierNode); ok {
			if v := g.Variables[ident.Value]; v.Type.Value == TYPE_PTR || v.Type.Value == TYPE_ARRAY {
				g.generatePop("rax")
				fmt.Printf("    mov rdi, %d\n", v.Type.Ptrof.Size)
				fmt.Printf("    mul rdi\n")
				g.generatePush("rax")
				g.generatePop("rdi")
				g.generatePop("rax")
				if n.Type == '+' {
					fmt.Printf("    add rax, rdi\n")
				} else if n.Type == '-' {
					fmt.Printf("    sub rax, rdi\n")
				}
				g.generatePush("rax")
				return nil, nil
			}
		}
		g.generatePop("rdi")
		g.generatePop("rax")
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
		g.generatePush("rax")
	case '=':
		switch left := n.Left.(type) {
		case *IdentifierNode:
			v := g.Variables[left.Value]
			fmt.Printf("    mov rax, rbp\n")
			fmt.Printf("    sub rax, %d\n", (v.Index+1)*MemorySize)
		case *UnaryOperatorNode:
			left.Accept(g)
		default:
			debug(left)
			panic("no identifier: ")
		}
		g.generatePush("rax")
		n.Right.Accept(g)
		g.generatePop("rdi")
		g.generatePop("rax")
		fmt.Printf("    mov [rax], rdi\n")
		g.generatePush("rdi")
	case ND_EQUAL:
		n.Left.Accept(g)
		n.Right.Accept(g)
		g.generatePop("rax")
		g.generatePop("rdi")
		fmt.Printf("    cmp rdi, rax\n")
		fmt.Printf("    sete al\n")
		fmt.Printf("    movzx rax, al\n")
		g.generatePush("rax")
	case ND_NOTEQUAL:
		n.Left.Accept(g)
		n.Right.Accept(g)
		g.generatePop("rax")
		g.generatePop("rdi")
		fmt.Printf("    cmp rdi, rax\n")
		fmt.Printf("    setne al\n")
		fmt.Printf("    movzx rax, al\n")
		g.generatePush("rax")
	}
	return nil, nil
}

func (g *Generator) VisitReturn(n *ReturnNode) (interface{}, error) {
	n.Expression.Accept(g)
	g.generatePop("rax")
	fmt.Printf("    mov rsp, rbp\n")
	g.generatePop("rbp")
	fmt.Printf("    ret\n")
	return nil, nil
}

func (g *Generator) VisitFunction(n *FunctionNode) (interface{}, error) {
	fmt.Printf("\n")
	fmt.Printf("_%s:\n", n.Identifier)
	g.generatePush("rbp")
	fmt.Printf("    mov rbp, rsp\n")

	for i, p := range n.Parameters {
		g.Variables[p.Identifier] = &Variable{
			Index: i,
			Type:  ctypeMap[p.Type],
		}
	}
	for _, stmt := range n.Statements {
		g.checkVariable(stmt)
	}
	stackSize := g.variableStackSize()
	fmt.Printf("    sub rsp, %d\n", stackSize)
	g.RspCounter -= stackSize

	for i, _ := range n.Parameters {
		fmt.Printf("    mov rax, rbp\n")
		fmt.Printf("    sub rax, %d\n", (i+1)*MemorySize)
		fmt.Printf("    mov [rax], %s\n", registerIndex[i])
	}

	for _, stmt := range n.Statements {
		stmt.Accept(g)
		g.generatePop("rax")
	}
	return nil, nil
}

func (g *Generator) VisitIdentifier(n *IdentifierNode) (interface{}, error) {
	v := g.Variables[n.Value]
	fmt.Printf("    mov rax, rbp\n")
	fmt.Printf("    sub rax, %d\n", (v.Index+1)*MemorySize)
	if v.Type.Value == TYPE_ARRAY {
		g.generatePush("rax")
	} else {
		g.generatePush("[rax]")
	}
	return nil, nil
}

func (g *Generator) VisitCall(n *CallNode) (interface{}, error) {
	if len(n.Args) > 0 {
		n.Args[0].Accept(g)
		g.generatePop("rdi")
	}
	if len(n.Args) > 1 {
		n.Args[1].Accept(g)
		g.generatePop("rsi")
	}
	if len(n.Args) > 2 {
		n.Args[2].Accept(g)
		g.generatePop("rdx")
	}
	if len(n.Args) > 3 {
		n.Args[3].Accept(g)
		g.generatePop("rcx")
	}
	if len(n.Args) > 4 {
		n.Args[4].Accept(g)
		g.generatePop("r8")
	}
	if len(n.Args) > 5 {
		n.Args[5].Accept(g)
		g.generatePop("r9")
	}
	if g.RspCounter%16 == 0 {
		fmt.Printf("    sub rsp, 8\n")
	}
	fmt.Printf("    call _%s\n", n.Identifier)
	if g.RspCounter%16 == 0 {
		fmt.Printf("    add rsp, 8\n")
	}
	g.generatePush("rax")
	return nil, nil
}

func (g *Generator) VisitIf(n *If) (interface{}, error) {
	n.Expression.Accept(g)
	label := fmt.Sprintf(".Lend%04d", g.LabelCnt)
	g.generatePop("rax")
	fmt.Printf("    cmp rax, 0\n")
	fmt.Printf("    je %s\n", label)
	g.LabelCnt++
	n.IfStatements.Accept(g)
	fmt.Printf("%s:\n", label)
	return nil, nil
}

func (g *Generator) VisitFor(n *For) (interface{}, error) {
	beginLabel := fmt.Sprintf(".Lbegin%04d", g.LabelCnt)
	endLabel := fmt.Sprintf(".Lend%04d", g.LabelCnt)
	oldBegin := g.CurrentLoopBegin
	oldEnd := g.CurrentLoopEnd
	g.CurrentLoopBegin = beginLabel
	g.CurrentLoopEnd = endLabel
	g.LabelCnt++

	n.Init.Accept(g)
	g.generatePop("rax")
	fmt.Printf("%s:\n", beginLabel)
	n.Expression.Accept(g)
	g.generatePop("rax")
	fmt.Printf("    cmp rax, 0\n")
	fmt.Printf("    je %s\n", endLabel)
	n.Statements.Accept(g)
	n.Update.Accept(g)
	g.generatePop("rax")
	fmt.Printf("jmp %s\n", beginLabel)
	fmt.Printf("%s:\n", endLabel)

	g.CurrentLoopBegin = oldBegin
	g.CurrentLoopEnd = oldEnd
	return nil, nil
}

func (g *Generator) VisitWhile(n *While) (interface{}, error) {
	beginLabel := fmt.Sprintf(".Lbegin%04d", g.LabelCnt)
	endLabel := fmt.Sprintf(".Lend%04d", g.LabelCnt)
	oldBegin := g.CurrentLoopBegin
	oldEnd := g.CurrentLoopEnd
	g.CurrentLoopBegin = beginLabel
	g.CurrentLoopEnd = endLabel
	g.LabelCnt++

	fmt.Printf("%s:\n", beginLabel)
	n.Expression.Accept(g)
	g.generatePop("rax")
	fmt.Printf("    cmp rax, 0\n")
	fmt.Printf("    je %s\n", endLabel)
	n.Statements.Accept(g)
	fmt.Printf("jmp %s\n", beginLabel)
	fmt.Printf("%s:\n", endLabel)
	g.CurrentLoopBegin = oldBegin
	g.CurrentLoopEnd = oldEnd
	g.generatePush("rax")
	return nil, nil
}

func (g *Generator) VisitGoto(n *Goto) (interface{}, error) {
	return nil, nil
}

func (g *Generator) VisitContinue(n *Continue) (interface{}, error) {
	fmt.Printf("jmp %s\n", g.CurrentLoopBegin)
	return nil, nil
}

func (g *Generator) VisitBreak(n *Break) (interface{}, error) {
	fmt.Printf("jmp %s\n", g.CurrentLoopEnd)
	return nil, nil
}

func (g *Generator) VisitBlock(n *Block) (interface{}, error) {
	for _, stmt := range n.Statements {
		stmt.Accept(g)
		g.generatePop("rax")
	}
	return nil, nil
}

func (g *Generator) VisitVariableDeclaration(n *VariableDeclaration) (interface{}, error) {
	g.generatePush("rax") // TODO: delete
	if n.Expression != nil {
		//if n.Type.Value == TYPE_ARRAY {
		//	v := g.Variables[n.Identifier]
		//	fmt.Printf("    mov rax, rbp\n")
		//	fmt.Printf("    sub rax, %d\n", (v.Index+1)*MemorySize)
		//	g.generatePush("rax")
		//}
		v := g.Variables[n.Identifier]
		fmt.Printf("    mov rax, rbp\n")
		fmt.Printf("    sub rax, %d\n", (v.Index+1)*MemorySize)
		g.generatePush("rax")
		n.Expression.Accept(g)
		g.generatePop("rdi")
		g.generatePop("rax")
		fmt.Printf("    mov [rax], rdi\n")
		g.generatePush("rdi")
	}
	return nil, nil
}

func (g *Generator) VisitUnaryOperator(n *UnaryOperatorNode) (interface{}, error) {
	switch n.Type {
	case '*':
		n.Expression.Accept(g)
		g.generatePop("rax")
		g.generatePush("[rax]")
	case '&':
		ident := n.Expression.(*IdentifierNode)
		v := g.Variables[ident.Value]
		fmt.Printf("    mov rax, rbp\n")
		fmt.Printf("    sub rax, %d\n", (v.Index+1)*MemorySize)
		g.generatePush("rax")
	}
	return nil, nil
}

func (g *Generator) generatePush(register string) {
	g.RspCounter += 8
	fmt.Printf("    push %s\n", register)
}

func (g *Generator) generatePop(register string) {
	g.RspCounter -= 8
	fmt.Printf("    pop %s\n", register)
}

func debug(args ...interface{}) {
	pp.Println(args...)
}
