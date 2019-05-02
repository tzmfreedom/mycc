package main

const (
	ND_EQUAL = iota + 256
	ND_NOTEQUAL
)

const (
	TYPE_INT = iota
	TYPE_CHAR
	TYPE_PTR
	TYPE_ARRAY
)

var ctypeMap = map[string]*Ctype{
	"int":  ctype_int,
	"char": ctype_char,
}

type Ctype struct {
	Value     int
	Ptrof     *Ctype
	Size      int
	ArraySize int
}

var ctype_int = &Ctype{Value: TYPE_INT, Size: 4}
var ctype_char = &Ctype{Value: TYPE_CHAR, Size: 1}

type Visitor interface {
	VisitInteger(n *Integer) (interface{}, error)
	VisitChar(n *Char) (interface{}, error)
	VisitString(n *String) (interface{}, error)
	VisitBinaryOperator(n *BinaryOperator) (interface{}, error)
	VisitCall(n *Call) (interface{}, error)
	VisitFunction(n *Function) (interface{}, error)
	VisitReturn(n *Return) (interface{}, error)
	VisitIdentifier(n *Identifier) (interface{}, error)
	VisitGlobalIdentifier(n *GlobalIdentifier) (interface{}, error)
	VisitIf(n *If) (interface{}, error)
	VisitFor(n *For) (interface{}, error)
	VisitGoto(n *Goto) (interface{}, error)
	VisitWhile(n *While) (interface{}, error)
	VisitBreak(n *Break) (interface{}, error)
	VisitContinue(n *Continue) (interface{}, error)
	VisitBlock(n *Block) (interface{}, error)
	VisitVariableDeclaration(n *VariableDeclaration) (interface{}, error)
	VisitUnaryOperator(n *UnaryOperatorNode) (interface{}, error)
	VisitGlobalVariableDeclaration(m *GlobalVariableDeclaration) (interface{}, error)
}

type Integer struct {
	Value int
}

func (n *Integer) Accept(v Visitor) (interface{}, error) {
	return v.VisitInteger(n)
}

type Char struct {
	Value int
}

func (n *Char) Accept(v Visitor) (interface{}, error) {
	return v.VisitChar(n)
}

type String struct {
	Value string
	Chars []*Char
}

func (n *String) Accept(v Visitor) (interface{}, error) {
	return v.VisitString(n)
}

type BinaryOperator struct {
	Ctype *Ctype
	Type  int
	Left  Node
	Right Node
}

func (n *BinaryOperator) Accept(v Visitor) (interface{}, error) {
	return v.VisitBinaryOperator(n)
}

type Call struct {
	Identifier string
	Args       []Node
}

func (n *Call) Accept(v Visitor) (interface{}, error) {
	return v.VisitCall(n)
}

type Function struct {
	ReturnType *Ctype
	Identifier string
	Parameters []*Parameter
	Statements []Node
	StackSize  int
}

func (n *Function) Accept(v Visitor) (interface{}, error) {
	return v.VisitFunction(n)
}

type Parameter struct {
	Identifier string
	Variable   *Variable
}

type Return struct {
	Expression Node
}

func (n *Return) Accept(v Visitor) (interface{}, error) {
	return v.VisitReturn(n)
}

type Identifier struct {
	Value    string
	Variable *Variable
}

func (n *Identifier) Accept(v Visitor) (interface{}, error) {
	return v.VisitIdentifier(n)
}

type GlobalIdentifier struct {
	Value    string
	Variable *Variable
}

func (n *GlobalIdentifier) Accept(v Visitor) (interface{}, error) {
	return v.VisitGlobalIdentifier(n)
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

type Variable struct {
	Index int
	Type  *Ctype
}

type VariableDeclaration struct {
	Variable   *Variable
	Identifier string
	Expression Node
}

func (n *VariableDeclaration) Accept(v Visitor) (interface{}, error) {
	return v.VisitVariableDeclaration(n)
}

type GlobalVariableDeclaration struct {
	Type       *Ctype
	Identifier string
	Expression Node
}

func (n *GlobalVariableDeclaration) Accept(v Visitor) (interface{}, error) {
	return v.VisitGlobalVariableDeclaration(n)
}

type Node interface {
	Accept(Visitor) (interface{}, error)
}
