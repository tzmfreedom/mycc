package main

import "strconv"

type Parser struct {
	Index  int
	Tokens []*Token
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

func (p *Parser) declarations() []Node {
	declarations := []Node{}
	for {
		declaration := p.declaration()
		if declaration == nil {
			break
		}
		declarations = append(declarations, declaration)
	}
	return declarations
}

func (p *Parser) declaration() Node {
	ident := p.identifier()
	if t := p.consume('('); t == nil {
		return nil
	}
	params := p.parameters()
	if t := p.consume(')'); t == nil {
		return nil
	}
	if t := p.consume('{'); t == nil {
		return nil
	}
	statements := p.statements()
	if t := p.consume('}'); t == nil {
		return nil
	}
	return &FunctionNode{
		Identifier: ident.Value,
		Parameters: params,
		Statements: statements,
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
	typeNode := p.identifier()
	if typeNode == nil {
		return nil
	}
	ident := p.identifier()
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
	if stmt := p.expressionStatement(); stmt != nil {
		return stmt
	}

	if stmt := p.returnStatement(); stmt != nil {
		return stmt
	}
	return nil
}

func (p *Parser) returnStatement() *ReturnNode {
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
	exp := p.assignExpression()
	if colon := p.consume(';'); colon == nil {
		return nil
	}
	return exp
}

func (p *Parser) expression() Node {
	if exp := p.try(p.add); exp != nil {
		return exp
	}
	if assign := p.try(p.assignExpression); assign != nil {
		return assign
	}
	return nil
}

func (p *Parser) assignExpression() Node {
	ident := p.identifier()
	if ident == nil {
		return nil
	}
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
	}
}

func (p *Parser) identifier() *IdentifierNode {
	token := p.consume(TK_IDENT)
	if token == nil {
		return nil
	}
	return &IdentifierNode{
		Value: token.Value,
	}
}

func (p *Parser) add() Node {
	node := p.booleanExpression()
	if next := p.consume('+'); next != nil {
		return &BinaryOperatorNode{
			Type:  next.Type,
			Left:  node,
			Right: p.add(),
		}
	}
	if next := p.consume('-'); next != nil {
		return &BinaryOperatorNode{
			Type:  next.Type,
			Left:  node,
			Right: p.add(),
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
		return p.call()
	}
	if token := p.consume('-'); token != nil {
		if term := p.call(); term != nil {
			return &BinaryOperatorNode{
				Type: '-',
				Left: &IntegerNode{
					Value: 0,
				},
				Right: term,
			}
		}
	}
	return p.call()
}

func (p *Parser) call() Node {
	current := p.Index
	if ident := p.identifier(); ident != nil {
		if token := p.consume('('); token != nil {
			args := p.expressionList()
			if token := p.consume(')'); token != nil {
				return &CallNode{
					Identifier: ident.Value,
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
