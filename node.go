package main

const (
	ND_EQUAL = iota + 256
	ND_NOTEQUAL
	ND_SIZEOF
)

const (
	TYPE_INT = iota
	TYPE_CHAR
	TYPE_PTR
)

var ctypeMap = map[string]*Ctype{
	"int":  ctype_int,
	"char": ctype_char,
}

type Ctype struct {
	Value int
	Ptrof *Ctype
}

var ctype_int = &Ctype{Value: TYPE_INT}
var ctype_char = &Ctype{Value: TYPE_CHAR}

type Visitor interface {
	VisitInteger(n *IntegerNode) (interface{}, error)
	VisitString(n *StringNode) (interface{}, error)
	VisitBinaryOperator(n *BinaryOperatorNode) (interface{}, error)
	VisitCall(n *CallNode) (interface{}, error)
	VisitFunction(n *FunctionNode) (interface{}, error)
	VisitReturn(n *ReturnNode) (interface{}, error)
	VisitIdentifier(n *IdentifierNode) (interface{}, error)
	VisitIf(n *If) (interface{}, error)
	VisitFor(n *For) (interface{}, error)
	VisitGoto(n *Goto) (interface{}, error)
	VisitWhile(n *While) (interface{}, error)
	VisitBreak(n *Break) (interface{}, error)
	VisitContinue(n *Continue) (interface{}, error)
	VisitBlock(n *Block) (interface{}, error)
	VisitVariableDeclaration(n *VariableDeclaration) (interface{}, error)
	VisitUnaryOperator(n *UnaryOperatorNode) (interface{}, error)
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
	ReturnType string
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

type UnaryOperatorNode struct {
	Type       int
	Expression Node
}

func (n *UnaryOperatorNode) Accept(v Visitor) (interface{}, error) {
	return v.VisitUnaryOperator(n)
}

type If struct {
	Expression     Node
	IfStatements   Node
	ElseStatements Node
}

func (n *If) Accept(v Visitor) (interface{}, error) {
	return v.VisitIf(n)
}

type For struct {
	Init       Node
	Expression Node
	Update     Node
	Statements Node
}

func (n *For) Accept(v Visitor) (interface{}, error) {
	return v.VisitFor(n)
}

type While struct {
	Expression Node
	Statements Node
}

func (n *While) Accept(v Visitor) (interface{}, error) {
	return v.VisitWhile(n)
}

type Goto struct {
	Label string
}

func (n *Goto) Accept(v Visitor) (interface{}, error) {
	return v.VisitGoto(n)
}

type Break struct{}

func (n *Break) Accept(v Visitor) (interface{}, error) {
	return v.VisitBreak(n)
}

type Continue struct{}

func (n *Continue) Accept(v Visitor) (interface{}, error) {
	return v.VisitContinue(n)
}

type Block struct {
	Statements []Node
}

func (n *Block) Accept(v Visitor) (interface{}, error) {
	return v.VisitBlock(n)
}

type VariableDeclaration struct {
	Type       *Ctype
	Identifier string
	Expression Node
}

func (n *VariableDeclaration) Accept(v Visitor) (interface{}, error) {
	return v.VisitVariableDeclaration(n)
}

type Node interface {
	Accept(Visitor) (interface{}, error)
}
