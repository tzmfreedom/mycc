package main

const (
	ND_NUMBER = iota + 256
	ND_IDENT
	ND_RETURN
)

type Node struct {
	Type  int
	Value string
	Left  *Node
	Right *Node
}

type Parser struct {
	Index  int
	Tokens []*Token
}

func NewParser(tokens []*Token) *Parser {
	return &Parser{Index: 0, Tokens: tokens}
}

func (p *Parser) Parse() []*Node {
	return p.statements()
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

func (p *Parser) statements() []*Node {
	statements := []*Node{}
	for {
		statement := p.statement()
		if statement == nil {
			break
		}
		statements = append(statements, statement)
	}
	return statements
}

func (p *Parser) statement() *Node {
	if stmt := p.expressionStatement(); stmt != nil {
		return stmt
	}

	if stmt := p.returnStatement(); stmt != nil {
		return stmt
	}
	return nil
}

func (p *Parser) returnStatement() *Node {
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
	return &Node{
		Type: ND_RETURN,
		Left: exp,
	}
}

func (p *Parser) expressionStatement() *Node {
	exp := p.assignExpression()
	if colon := p.consume(';'); colon == nil {
		return nil
	}
	return exp
}

func (p *Parser) expression() *Node {
	if exp := p.add(); exp != nil {
		return exp
	}
	if assign := p.assignExpression(); assign != nil {
		return assign
	}
	return nil
}

func (p *Parser) assignExpression() *Node {
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
	return &Node{
		Type:  token.Type,
		Left:  ident,
		Right: exp,
	}
}

func (p *Parser) identifier() *Node {
	token := p.consume(TK_IDENT)
	if token == nil {
		return nil
	}
	return &Node{
		Type:  ND_IDENT,
		Value: token.Value,
	}
}

func (p *Parser) add() *Node {
	node := p.mul()
	if next := p.consume('+'); next != nil {
		return &Node{
			Type:  next.Type,
			Left:  node,
			Right: p.add(),
		}
	}
	if next := p.consume('-'); next != nil {
		return &Node{
			Type:  next.Type,
			Left:  node,
			Right: p.add(),
		}
	}
	return node
}

func (p *Parser) mul() *Node {
	node := p.unary()
	if next := p.consume('*'); next != nil {
		return &Node{
			Type:  next.Type,
			Left:  node,
			Right: p.mul(),
		}
	}
	if next := p.consume('-'); next != nil {
		return &Node{
			Type:  next.Type,
			Left:  node,
			Right: p.mul(),
		}
	}
	return node
}

func (p *Parser) unary() *Node {
	if token := p.consume('+'); token != nil {
		return p.term()
	}
	if token := p.consume('-'); token != nil {
		if term := p.term(); term != nil {
			return &Node{
				Type: '-',
				Left: &Node{
					Type:  TK_NUMBER,
					Value: "0",
				},
				Right: term,
			}
		}
	}
	return p.term()
}

func (p *Parser) term() *Node {
	if next := p.consume('('); next != nil {
		node := p.expression()
		if next := p.consume(')'); next != nil {
			return node
		}
	}

	if token := p.consume(TK_NUMBER); token != nil {
		return &Node{
			Type:  ND_NUMBER,
			Value: token.Value,
		}
	}
	if ident := p.consume(TK_IDENT); ident != nil {
		return &Node{
			Type:  ND_IDENT,
			Value: ident.Value,
		}
	}
	return nil
}
