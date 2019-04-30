package main

import "strconv"

type Parser struct {
	Index  int
	Tokens []*Token
	LVars  map[string]*Variable
}

func NewParser(tokens []*Token) *Parser {
	return &Parser{Index: 0, Tokens: tokens}
}

func (p *Parser) Parse() []Node {
	return p.declarations()
}

func (p *Parser) current() *Token {
	return p.Tokens[p.Index]
}

func (p *Parser) consume(t int) *Token {
	if len(p.Tokens) <= p.Index {
		return nil
	}
	current := p.current()
	if t == current.Type {
		p.Index++
		return current
	}
	return nil
}

func (p *Parser) repeat(t int) []*Token {
	tokens := []*Token{}
	for {
		token := p.consume(t)
		if token == nil {
			break
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func (p *Parser) declarations() []Node {
	declarations := []Node{}
	for {
		p.LVars = map[string]*Variable{}
		declaration := p.declaration()
		if declaration == nil {
			break
		}
		declarations = append(declarations, declaration)
	}
	return declarations
}

func (p *Parser) declaration() Node {
	returnType := p.consume(TK_IDENT)
	if returnType == nil {
		return nil
	}
	ident := p.consume(TK_IDENT)
	if ident == nil {
		return nil
	}
	if t := p.consume('('); t == nil {
		return nil
	}
	params := p.parameters()
	if t := p.consume(')'); t == nil {
		return nil
	}
	block := p.block()
	if block == nil {
		return nil
	}
	return &FunctionNode{
		ReturnType: returnType.Value,
		Identifier: ident.Value,
		Parameters: params,
		Statements: block.(*Block).Statements,
	}
}

func (p *Parser) parameters() []*Parameter {
	parameters := []*Parameter{}
	for {
		parameter := p.parameter()
		if parameter == nil {
			break
		}
		parameters = append(parameters, parameter)
		if token := p.consume(','); token == nil {
			break
		}
	}
	return parameters
}

func (p *Parser) parameter() *Parameter {
	typeNode := p.consume(TK_IDENT)
	if typeNode == nil {
		return nil
	}
	ident := p.consume(TK_IDENT)
	if ident == nil {
		return nil
	}
	return &Parameter{
		Type:       typeNode.Value,
		Identifier: ident.Value,
	}
}

func (p *Parser) statements() []Node {
	statements := []Node{}
	for {
		statement := p.statement()
		if statement == nil {
			break
		}
		statements = append(statements, statement)
	}
	return statements
}

func (p *Parser) statement() Node {
	if stmt := p.try(p.block); stmt != nil {
		return stmt
	}

	if stmt := p.try(p.expressionStatement); stmt != nil {
		return stmt
	}

	if stmt := p.try(p.returnStatement); stmt != nil {
		return stmt
	}

	if stmt := p.try(p.ifStatement); stmt != nil {
		return stmt
	}

	if stmt := p.try(p.whileStatement); stmt != nil {
		return stmt
	}

	if stmt := p.try(p.forStatement); stmt != nil {
		return stmt
	}

	if stmt := p.try(p.continueStatement); stmt != nil {
		return stmt
	}

	if stmt := p.try(p.breakStatement); stmt != nil {
		return stmt
	}

	if stmt := p.try(p.variableDeclarationStatement); stmt != nil {
		return stmt
	}
	return nil
}

func (p *Parser) variableDeclarationStatement() Node {
	typeNode := p.consume(TK_IDENT)
	if typeNode == nil {
		return nil
	}
	ptrs := p.repeat('*')
	ctype := ctypeMap[typeNode.Value]
	for i := 0; i < len(ptrs); i++ {
		ctype = &Ctype{
			Value: TYPE_PTR,
			Ptrof: ctype,
			Size:  8,
		}
	}
	ident := p.consume(TK_IDENT)
	if ident == nil {
		return nil
	}
	if colon := p.consume(';'); colon != nil {
		p.LVars[ident.Value] = &Variable{Type: ctype}
		return &VariableDeclaration{
			Type:       ctype,
			Identifier: ident.Value,
			Expression: nil,
		}
	}
	token := p.consume('=')
	if token == nil {
		return nil
	}
	exp := p.expression()
	if exp == nil {
		return nil
	}
	if colon := p.consume(';'); colon == nil {
		return nil
	}
	p.LVars[ident.Value] = &Variable{Type: ctype}
	return &VariableDeclaration{
		Type:       ctype,
		Identifier: ident.Value,
		Expression: exp,
	}
}

func (p *Parser) ifStatement() Node {
	if t := p.consume(TK_IF); t == nil {
		return nil
	}
	if t := p.consume('('); t == nil {
		return nil
	}
	expression := p.expression()
	if t := p.consume(')'); t == nil {
		return nil
	}
	stmt := p.statement()
	if stmt == nil {
		return nil
	}
	return &If{
		Expression:     expression,
		IfStatements:   stmt,
		ElseStatements: nil,
	}
}

func (p *Parser) breakStatement() Node {
	if t := p.consume(TK_BREAK); t == nil {
		return nil
	}
	if colon := p.consume(';'); colon == nil {
		return nil
	}
	return &Break{}
}

func (p *Parser) continueStatement() Node {
	if t := p.consume(TK_CONTINUE); t == nil {
		return nil
	}
	if colon := p.consume(';'); colon == nil {
		return nil
	}
	return &Continue{}
}

func (p *Parser) whileStatement() Node {
	if t := p.consume(TK_WHILE); t == nil {
		return nil
	}
	if t := p.consume('('); t == nil {
		return nil
	}
	expression := p.expression()
	if t := p.consume(')'); t == nil {
		return nil
	}
	stmt := p.statement()
	if stmt == nil {
		return nil
	}
	return &While{
		Expression: expression,
		Statements: stmt,
	}
}

func (p *Parser) forStatement() Node {
	if t := p.consume(TK_FOR); t == nil {
		return nil
	}
	if t := p.consume('('); t == nil {
		return nil
	}
	init := p.variableDeclarationStatement()
	if init == nil {
		init = p.expressionStatement()
	}
	exp := p.expression()
	if t := p.consume(';'); t == nil {
		return nil
	}
	update := p.expression()
	if t := p.consume(')'); t == nil {
		return nil
	}
	stmt := p.statement()
	if stmt == nil {
		return nil
	}
	return &For{
		Init:       init,
		Expression: exp,
		Update:     update,
		Statements: stmt,
	}
}

func (p *Parser) block() Node {
	if t := p.consume('{'); t == nil {
		return nil
	}
	statements := p.statements()
	if statements == nil {
		return nil
	}
	if t := p.consume('}'); t == nil {
		return nil
	}
	return &Block{
		Statements: statements,
	}
}

func (p *Parser) returnStatement() Node {
	if ret := p.consume(TK_RETURN); ret == nil {
		return nil
	}
	exp := p.expression()
	if exp == nil {
		return nil
	}
	if colon := p.consume(';'); colon == nil {
		return nil
	}
	return &ReturnNode{
		Expression: exp,
	}
}

func (p *Parser) expressionStatement() Node {
	exp := p.try(p.assignExpression)
	if exp == nil {
		exp = p.callExpression()
		if exp == nil {
			return nil
		}
	}
	if colon := p.consume(';'); colon == nil {
		return nil
	}
	return exp
}

func (p *Parser) expression() Node {
	if assign := p.try(p.assignExpression); assign != nil {
		return assign
	}
	if exp := p.try(p.add); exp != nil {
		return exp
	}
	return nil
}

func (p *Parser) assignExpression() Node {
	i := p.consume(TK_IDENT)
	if i == nil {
		return nil
	}
	ident := &IdentifierNode{Value: i.Value}
	token := p.consume('=')
	if token == nil {
		return nil
	}
	exp := p.expression()
	if exp == nil {
		return nil
	}
	return &BinaryOperatorNode{
		Type:  token.Type,
		Left:  ident,
		Right: exp,
		Ctype: p.getCtype(ident, exp),
	}
}

func (p *Parser) add() Node {
	node := p.booleanExpression()
	if next := p.consume('+'); next != nil {
		right := p.add()
		return &BinaryOperatorNode{
			Type:  next.Type,
			Left:  node,
			Right: right,
			Ctype: p.getCtype(node, right),
		}
	}
	if next := p.consume('-'); next != nil {
		right := p.add()
		return &BinaryOperatorNode{
			Type:  next.Type,
			Left:  node,
			Right: right,
			Ctype: p.getCtype(node, right),
		}
	}
	return node
}

func (p *Parser) booleanExpression() Node {
	node := p.mul()
	if next := p.consume(TK_EQUAL); next != nil {
		return &BinaryOperatorNode{
			Type:  ND_EQUAL,
			Left:  node,
			Right: p.booleanExpression(),
		}
	}
	if next := p.consume(TK_NOTEQUAL); next != nil {
		return &BinaryOperatorNode{
			Type:  ND_NOTEQUAL,
			Left:  node,
			Right: p.booleanExpression(),
		}
	}
	return node
}

func (p *Parser) mul() Node {
	node := p.unary()
	if next := p.consume('*'); next != nil {
		return &BinaryOperatorNode{
			Type:  next.Type,
			Left:  node,
			Right: p.mul(),
		}
	}
	if next := p.consume('-'); next != nil {
		return &BinaryOperatorNode{
			Type:  next.Type,
			Left:  node,
			Right: p.mul(),
		}
	}
	return node
}

func (p *Parser) unary() Node {
	if token := p.consume('+'); token != nil {
		return p.callExpression()
	}
	if token := p.consume('-'); token != nil {
		if term := p.callExpression(); term != nil {
			return &BinaryOperatorNode{
				Type: '-',
				Left: &IntegerNode{
					Value: 0,
				},
				Right: term,
			}
		}
	}
	if t := p.consume('&'); t != nil {
		ident := p.consume(TK_IDENT)
		if ident != nil {
			return &UnaryOperatorNode{
				Type:       '&',
				Expression: &IdentifierNode{Value: ident.Value},
			}
		}
	}
	if tokens := p.repeat('*'); len(tokens) > 0 {
		ident := p.consume(TK_IDENT)
		if ident != nil {
			return &UnaryOperatorNode{
				Type:       '*',
				Expression: &IdentifierNode{Value: ident.Value},
			}
		}
	}
	if token := p.consume(TK_SIZEOF); token != nil {
		if t := p.consume('('); t != nil {
			if exp := p.unary(); exp != nil {
				if t := p.consume(')'); t != nil {
					switch node := exp.(type) {
					case *IntegerNode:
						return &IntegerNode{
							Value: ctype_int.Size,
						}
					case *BinaryOperatorNode:
						return &IntegerNode{
							Value: node.Ctype.Size,
						}
					}
				}
			}
		}
	}
	return p.callExpression()
}

func (p *Parser) callExpression() Node {
	current := p.Index
	if t := p.consume(TK_IDENT); t != nil {
		if token := p.consume('('); token != nil {
			args := p.expressionList()
			if token := p.consume(')'); token != nil {
				return &CallNode{
					Identifier: t.Value,
					Args:       args,
				}
			}
		}
	}
	p.Index = current
	return p.term()
}

func (p *Parser) term() Node {
	if next := p.consume('('); next != nil {
		node := p.expression()
		if next := p.consume(')'); next != nil {
			return node
		}
	}
	if token := p.consume(TK_NUMBER); token != nil {
		num, _ := strconv.Atoi(token.Value)
		return &IntegerNode{
			Value: num,
		}
	}
	if ident := p.consume(TK_IDENT); ident != nil {
		return &IdentifierNode{
			Value: ident.Value,
		}
	}
	return nil
}

func (p *Parser) try(f func() Node) Node {
	current := p.Index
	ret := f()
	if ret == nil {
		p.Index = current
		return nil
	}
	return ret
}

func (p *Parser) expressionList() []Node {
	expressionList := []Node{}
	for {
		exp := p.expression()
		if exp == nil {
			break
		}
		expressionList = append(expressionList, exp)
		if token := p.consume(','); token == nil {
			break
		}
	}
	return expressionList
}

func (p *Parser) getCtype(l Node, r Node) *Ctype {
	if l != nil {
		if ident := l.(*IdentifierNode); ident != nil {
			return p.LVars[ident.Value].Type
		}
	}
	if r != nil {
		if ident := r.(*IdentifierNode); ident != nil {
			return p.LVars[ident.Value].Type
		}
	}
	if l != nil {
		switch node := l.(type) {
		case *BinaryOperatorNode:
			return node.Ctype
		case *IntegerNode:
			return ctype_int
		}
	}
	if r != nil {
		switch node := l.(type) {
		case *BinaryOperatorNode:
			return node.Ctype
		case *IntegerNode:
			return ctype_int
		}
	}
	return nil
}
