package main

const (
	ND_EQUAL = iota + 256
	ND_NOTEQUAL
)

type Visitor interface {
	VisitInteger(n *IntegerNode) (interface{}, error)
	VisitString(n *StringNode) (interface{}, error)
	VisitBinaryOperator(n *BinaryOperatorNode) (interface{}, error)
	VisitCall(n *CallNode) (interface{}, error)
	VisitFunction(n *FunctionNode) (interface{}, error)
	VisitReturn(n *ReturnNode) (interface{}, error)
	VisitIdentifier(n *IdentifierNode) (interface{}, error)
}

type IntegerNode struct {
	Value int
}

func (n *IntegerNode) Accept(v Visitor) (interface{}, error) {
	return v.VisitInteger(n)
}

type StringNode struct {
	Value string
}

func (n *StringNode) Accept(v Visitor) (interface{}, error) {
	return v.VisitString(n)
}

type BinaryOperatorNode struct {
	Type  int
	Left  Node
	Right Node
}

func (n *BinaryOperatorNode) Accept(v Visitor) (interface{}, error) {
	return v.VisitBinaryOperator(n)
}

type CallNode struct {
	Identifier string
	Args       []Node
}

func (n *CallNode) Accept(v Visitor) (interface{}, error) {
	return v.VisitCall(n)
}

type FunctionNode struct {
	Identifier string
	Parameters []*Parameter
	Statements []Node
}

func (n *FunctionNode) Accept(v Visitor) (interface{}, error) {
	return v.VisitFunction(n)
}

type Parameter struct {
	Type       string
	Identifier string
}

type ReturnNode struct {
	Expression Node
}

func (n *ReturnNode) Accept(v Visitor) (interface{}, error) {
	return v.VisitReturn(n)
}

type IdentifierNode struct {
	Value string
}

func (n *IdentifierNode) Accept(v Visitor) (interface{}, error) {
	return v.VisitIdentifier(n)
}

type Node interface {
	Accept(Visitor) (interface{}, error)
}
