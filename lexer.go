package main

type Lexer struct {
	Original string
	Runes    []rune
	Index    int
	Line     int
	Column   int
}

func NewLexer(str string) *Lexer {
	runes := []rune(str)
	return &Lexer{
		Original: str,
		Runes:    runes,
		Index:    0,
		Line:     1,
		Column:   0,
	}
}

func (l *Lexer) Tokenize(str string) []*Token {
	tokens := []*Token{}
	for {
		if l.Index >= len(l.Runes) {
			break
		}
		r := l.Runes[l.Index]
		var token *Token
		switch r {
		case '+':
			token = l.createToken("add", string(r))
		case '-':
			token = l.createToken("sub", string(r))
		case '*':
			token = l.createToken("mul", string(r))
		case '/':
			token = l.createToken("div", string(r))
		case '\n':
			l.Column = 0
			l.Line++
		case ' ', '　':
			l.Index++
			l.Column++
			continue
		default:
			if r >= '0' && r <= '9' {
				token = l.parseNumber()
			}
		}
		tokens = append(tokens, token)
		l.Index++
		l.Column++
	}
	return tokens
}

func (l *Lexer) createToken(t string, v string) *Token {
	return &Token{
		Type:   t,
		Value:  v,
		Line:   l.Line,
		Column: l.Column,
	}
}

func (l *Lexer) parseNumber() *Token {
	runes := []rune{l.Runes[l.Index]}
	for {
		l.Index++
		l.Column++
		if len(l.Runes) <= l.Index {
			break
		}
		r := l.Runes[l.Index]
		if r >= '0' && r <= '9' {
			runes = append(runes, r)
		} else {
			break
		}
	}
	return l.createToken("number", string(runes))
}
