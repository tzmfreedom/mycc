package main

import (
	"fmt"
	"github.com/k0kubun/pp"
)

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
	//GlobalVariables  map[string]*Variable
	//LocalVariables   map[string]*Variable
	LabelCnt         int
	RspCounter       int
	CurrentLoopBegin string
	CurrentLoopEnd   string
	Strings          map[string]int
}

func NewGenerator(strs map[string]int) *Generator {
	return &Generator{
		//LocalVariables:  map[string]*Variable{},
		//GlobalVariables: map[string]*Variable{},
		LabelCnt:   0,
		RspCounter: 0,
		Strings:    strs,
	}
}

func (g *Generator) generate(declarations []Node) {
	for s, i := range g.Strings {
		fmt.Printf(".LC%d:\n", i)
		fmt.Printf("    .string \"%s\"\n", s)
	}
	fmt.Println(`.data
.intel_syntax noprefix
.global main`)
	for _, declaration := range declarations {
		declaration.Accept(g)
	}
}

func (g *Generator) VisitInteger(n *Integer) (interface{}, error) {
	g.generatePush(fmt.Sprintf("%d", n.Value))
	return nil, nil
}

func (g *Generator) VisitChar(n *Char) (interface{}, error) {
	g.generatePush(fmt.Sprintf("%d", n.Value))
	return nil, nil
}

func (g *Generator) VisitString(n *String) (interface{}, error) {
	g.generatePush(fmt.Sprintf("OFFSET FLAT:.LC%d", g.Strings[n.Value]))
	return nil, nil
}

func (g *Generator) VisitBinaryOperator(n *BinaryOperator) (interface{}, error) {
	switch n.Type {
	case '+', '-', '*', '/':
		n.Left.Accept(g)
		if ident, ok := n.Right.(*Identifier); ok {
			if ident.Variable.Type.Value == TYPE_PTR || ident.Variable.Type.Value == TYPE_ARRAY {
				g.generatePop("rax")
				fmt.Printf("    mov rdi, %d\n", ident.Variable.Type.Ptrof.Size)
				fmt.Printf("    mul rdi\n")
				g.generatePush("rax")
			}
		}
		n.Right.Accept(g)
		if ident, ok := n.Left.(*Identifier); ok {
			if ident.Variable.Type.Value == TYPE_PTR || ident.Variable.Type.Value == TYPE_ARRAY {
				g.generatePop("rax")
				fmt.Printf("    mov rdi, %d\n", ident.Variable.Type.Ptrof.Size)
				fmt.Printf("    mul rdi\n")
				g.generatePush("rax")
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
		case *GlobalIdentifier:
		case *Identifier:
			fmt.Printf("    mov rax, rbp\n")
			fmt.Printf("    sub rax, %d\n", (left.Variable.Index+1)*MemorySize)
			g.generatePush("rax")
		case *UnaryOperatorNode:
			left.Accept(g)
			g.generatePush("rax")
		default:
			debug(left)
			panic("no identifier: ")
		}
		n.Right.Accept(g)
		switch left := n.Left.(type) {
		case *GlobalIdentifier:
			g.generatePop("rdi")
			fmt.Printf("    mov %s[rip], rdi\n", left.Value)
		default:
			g.generatePop("rdi")
			g.generatePop("rax")
			fmt.Printf("    mov [rax], rdi\n")
		}
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

func (g *Generator) VisitReturn(n *Return) (interface{}, error) {
	n.Expression.Accept(g)
	g.generatePop("rax")
	fmt.Printf("    mov rsp, rbp\n")
	g.generatePop("rbp")
	fmt.Printf("    ret\n")
	return nil, nil
}

func (g *Generator) VisitFunction(n *Function) (interface{}, error) {
	fmt.Printf("\n")
	fmt.Printf("%s:\n", n.Identifier)
	g.generatePush("rbp")
	fmt.Printf("    mov rbp, rsp\n")

	stackSize := n.StackSize
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

func (g *Generator) VisitIdentifier(n *Identifier) (interface{}, error) {
	fmt.Printf("    mov rax, rbp\n")
	fmt.Printf("    sub rax, %d\n", (n.Variable.Index+1)*MemorySize)
	if n.Variable.Type.Value == TYPE_ARRAY {
		g.generatePush("rax")
	} else {
		g.generatePush("[rax]")
	}
	return nil, nil
}

func (g *Generator) VisitCall(n *Call) (interface{}, error) {
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
		//	v := g.LocalVariables[n.Identifier]
		//	fmt.Printf("    mov rax, rbp\n")
		//	fmt.Printf("    sub rax, %d\n", (v.Index+1)*MemorySize)
		//	g.generatePush("rax")
		//}
		fmt.Printf("    mov rax, rbp\n")
		fmt.Printf("    sub rax, %d\n", (n.Variable.Index+1)*MemorySize)
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
		if ident, ok := n.Expression.(*Identifier); ok {
			fmt.Printf("    mov rax, rbp\n")
			fmt.Printf("    sub rax, %d\n", (ident.Variable.Index+1)*MemorySize)
			g.generatePush("rax")
		} else if _, ok := n.Expression.(*GlobalIdentifier); ok {
			panic("not impl")
		}
	}
	return nil, nil
}

func (g *Generator) VisitGlobalIdentifier(n *GlobalIdentifier) (interface{}, error) {
	fmt.Printf("    mov rax, %s[rip]\n", n.Value)
	g.generatePush("rax")
	return nil, nil
}

func (g *Generator) VisitGlobalVariableDeclaration(n *GlobalVariableDeclaration) (interface{}, error) {
	fmt.Printf("%s:\n", n.Identifier)
	if n.Expression != nil {
		switch n.Type.Value {
		case TYPE_INT:
			switch node := n.Expression.(type) {
			case *Integer:
				fmt.Printf("    .long %d\n", node.Value)
				return nil, nil
			}
		}
	}
	fmt.Printf("    .text %d\n", n.Type.Size)
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
