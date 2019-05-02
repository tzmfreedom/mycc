package main

import "strconv"

type Parser struct {
	Index   int
	Tokens  []*Token
	LVars   map[string]*Variable
	GVars   map[string]*Variable
	Strings map[string]int
}

func NewParser(tokens []*Token) *Parser {
	return &Parser{
		Index:   0,
		Tokens:  tokens,
		GVars:   map[string]*Variable{},
		Strings: map[string]int{},
	}
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
		if p.consume(TK_EOF) != nil {
			break
		}
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
	ctype := p.ctype()
	if ctype == nil {
		panic("cannot parse type")
	}
	ident := p.consume(TK_IDENT)
	if ident == nil {
		return nil
	}
	if p.current().Type == '(' {
		if f := p.function(ctype, ident.Value); f != nil {
			return f
		}
	} else {
		if t := p.consume(';'); t != nil {
			_, ok := p.GVars[ident.Value]
			if ok {
				panic("variable redeclaration: " + ident.Value)
			}
			p.GVars[ident.Value] = &Variable{Type: ctype}
			return &GlobalVariableDeclaration{
				Type:       ctype,
				Identifier: ident.Value,
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
		_, ok := p.GVars[ident.Value]
		if ok {
			panic("variable redeclaration: " + ident.Value)
		}
		p.GVars[ident.Value] = &Variable{Type: ctype}
		return &GlobalVariableDeclaration{
			Type:       ctype,
			Identifier: ident.Value,
			Expression: exp,
		}
	}
	return nil
}

func (p *Parser) ctype() *Ctype {
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
	return ctype
}

func (p *Parser) function(ctype *Ctype, ident string) Node {
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
	return &Function{
		ReturnType: ctype,
		Identifier: ident,
		Parameters: params,
		Statements: block.(*Block).Statements,
		StackSize:  p.localVariableStackSize(),
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
	v := p.createLocalVariable(ctypeMap[typeNode.Value])
	p.LVars[ident.Value] = v
	return &Parameter{
		Variable:   v,
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
	ctype := p.ctype()
	if ctype == nil {
		return nil
	}
	ident := p.consume(TK_IDENT)
	if ident == nil {
		return nil
	}
	if t := p.consume('['); t != nil {
		if num := p.consume(TK_NUMBER); num != nil {
			if t := p.consume(']'); t != nil {
				arraySize, err := strconv.Atoi(num.Value)
				if err != nil {
					panic(err)
				}
				ctype = &Ctype{
					Value:     TYPE_ARRAY,
					Ptrof:     ctype,
					Size:      arraySize * ctype.Size,
					ArraySize: arraySize,
				}
			}
		}
	}
	if colon := p.consume(';'); colon != nil {
		_, ok := p.LVars[ident.Value]
		if ok {
			panic("variable redeclaration: " + ident.Value)
		}
		v := p.createLocalVariable(ctype)
		p.LVars[ident.Value] = v
		return &VariableDeclaration{
			Variable:   v,
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
	_, ok := p.LVars[ident.Value]
	if ok {
		panic("variable redeclaration: " + ident.Value)
	}
	v := p.createLocalVariable(ctype)
	p.LVars[ident.Value] = v
	return &VariableDeclaration{
		Variable:   v,
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

func (p *Parser) arrayExpression() Node {
	left := p.callExpression()
	if left == nil {
		return nil
	}
	if t := p.consume('['); t == nil {
		return nil
	}
	right := p.callExpression()
	if right == nil {
		return nil
	}
	if t := p.consume(']'); t == nil {
		return nil
	}
	return &UnaryOperatorNode{
		Type: '*',
		Expression: &BinaryOperator{
			Type:  '+',
			Left:  left,
			Right: right,
			Ctype: p.getCtype(left, right),
		},
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
	return &Return{
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
	var left Node
	if left = p.try(p.arrayExpression); left == nil {
		if left = p.try(p.pointerExpression); left == nil {
			ident := p.consume(TK_IDENT)
			if ident == nil {
				return nil
			}
			left = p.lookup(ident.Value)
		}
	}
	token := p.consume('=')
	if token == nil {
		return nil
	}
	right := p.expression()
	if right == nil {
		return nil
	}
	return &BinaryOperator{
		Type:  token.Type,
		Left:  left,
		Right: right,
		Ctype: p.getCtype(left, right),
	}
}

func (p *Parser) add() Node {
	node := p.booleanExpression()
	if next := p.consume('+'); next != nil {
		right := p.add()
		return &BinaryOperator{
			Type:  next.Type,
			Left:  node,
			Right: right,
			Ctype: p.getCtype(node, right),
		}
	}
	if next := p.consume('-'); next != nil {
		right := p.add()
		return &BinaryOperator{
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
		return &BinaryOperator{
			Type:  ND_EQUAL,
			Left:  node,
			Right: p.booleanExpression(),
		}
	}
	if next := p.consume(TK_NOTEQUAL); next != nil {
		return &BinaryOperator{
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
		return &BinaryOperator{
			Type:  next.Type,
			Left:  node,
			Right: p.mul(),
		}
	}
	if next := p.consume('-'); next != nil {
		return &BinaryOperator{
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
			return &BinaryOperator{
				Type: '-',
				Left: &Integer{
					Value: 0,
				},
				Right: term,
			}
		}
	}
	if t := p.consume('&'); t != nil {
		ident := p.consume(TK_IDENT)
		if ident != nil {
			exp := p.lookup(ident.Value)
			if exp == nil {
				panic("no variable declaration: " + ident.Value)
			}
			return &UnaryOperatorNode{
				Type:       '&',
				Expression: exp,
			}
		}
	}
	if exp := p.try(p.pointerExpression); exp != nil {
		return exp
	}
	if token := p.consume(TK_SIZEOF); token != nil {
		if t := p.consume('('); t != nil {
			if exp := p.expression(); exp != nil {
				if t := p.consume(')'); t != nil {
					switch node := exp.(type) {
					case *Identifier:
						return &Integer{
							Value: node.Variable.Type.Size,
						}
					case *Integer:
						return &Integer{
							Value: ctype_int.Size,
						}
					case *BinaryOperator:
						return &Integer{
							Value: node.Ctype.Size,
						}
					case *Char:
						return &Integer{
							Value: ctype_char.Size,
						}
					}
				}
			}
		}
	}
	if exp := p.try(p.arrayExpression); exp != nil {
		return exp
	}
	return p.callExpression()
}

func (p *Parser) pointerExpression() Node {
	if tokens := p.repeat('*'); len(tokens) > 0 {
		if exp := p.unary(); exp != nil {
			for range tokens {
				exp = &UnaryOperatorNode{
					Type:       '*',
					Expression: exp,
				}
			}
			return exp
		}
	}
	return nil
}

func (p *Parser) callExpression() Node {
	current := p.Index
	if t := p.consume(TK_IDENT); t != nil {
		if token := p.consume('('); token != nil {
			args := p.expressionList()
			if token := p.consume(')'); token != nil {
				return &Call{
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
		return &Integer{
			Value: num,
		}
	}
	if token := p.consume(TK_CHAR); token != nil {
		return &Char{
			Value: int(rune(token.Value[0])),
		}
	}
	if token := p.consume(TK_STRING); token != nil {
		runes := []rune(token.Value)
		chars := make([]*Char, len(runes))
		for i, r := range runes {
			chars[i] = &Char{
				Value: int(r),
			}
		}
		p.Strings[token.Value] = len(p.Strings)
		return &String{
			Value: token.Value,
			Chars: chars,
		}
	}
	if ident := p.consume(TK_IDENT); ident != nil {
		if i := p.lookup(ident.Value); i != nil {
			return i
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
		if ident, ok := l.(*Identifier); ok {
			return ident.Variable.Type
		}
	}
	if r != nil {
		if ident, ok := r.(*Identifier); ok {
			return ident.Variable.Type
		}
	}
	if l != nil {
		switch node := l.(type) {
		case *BinaryOperator:
			return node.Ctype
		case *Integer:
			return ctype_int
		}
	}
	if r != nil {
		switch node := l.(type) {
		case *BinaryOperator:
			return node.Ctype
		case *Integer:
			return ctype_int
		}
	}
	return nil
}

func (p *Parser) localVariableStackSize() int {
	stackSize := 0
	for _, v := range p.LVars {
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

func (p *Parser) lookup(ident string) Node {
	if v, ok := p.LVars[ident]; ok {
		return &Identifier{
			Value:    ident,
			Variable: v,
		}
	} else if v, ok := p.GVars[ident]; ok {
		return &GlobalIdentifier{
			Value:    ident,
			Variable: v,
		}
	}
	return nil
}

func (p *Parser) createLocalVariable(ctype *Ctype) *Variable {
	return &Variable{
		Type:  ctype,
		Index: p.localVariableStackSize() / 8,
	}
}
