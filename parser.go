package main

type Parser struct {
	Index  int
	Tokens []*Token
}

func NewParser(tokens []*Token) *Parser {
	return &Parser{Index: 0, Tokens: tokens}
}

func (p *Parser) Parse() *Node {
	return p.add()
}

func (p *Parser) current() *Token {
	return p.Tokens[p.Index]
}

func (p *Parser) consume(t string) *Token {
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

func (p *Parser) add() *Node {
	node := p.mul()
	if next := p.consume("add"); next != nil {
		return &Node{
			Type:  next.Type,
			Left:  node,
			Right: p.add(),
		}
	}
	if next := p.consume("sub"); next != nil {
		return &Node{
			Type:  next.Type,
			Left:  node,
			Right: p.add(),
		}
	}
	return node
}

func (p *Parser) mul() *Node {
	node := p.term()
	if next := p.consume("mul"); next != nil {
		return &Node{
			Type:  next.Type,
			Left:  node,
			Right: p.mul(),
		}
	}
	if next := p.consume("div"); next != nil {
		return &Node{
			Type:  next.Type,
			Left:  node,
			Right: p.mul(),
		}
	}
	return node
}

func (p *Parser) term() *Node {
	token := p.consume("number")
	if token != nil {
		return &Node{
			Type:  "number",
			Value: token.Value,
		}
	}
	panic("not number: " + p.peek().Type)
}

func (p *Parser) peek() *Token {
	return p.Tokens[p.Index+1]
}

type Node struct {
	Type  string
	Value string
	Left  *Node
	Right *Node
}
