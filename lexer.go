package main

const (
	TK_NUMBER = iota + 256
	TK_IDENT
	TK_EOF
	TK_RETURN
)

var reservationTypes = map[string]int{
	"return": TK_RETURN,
}

type Token struct {
	Type   int
	Value  string
	Line   int
	Column int
}

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
		Column:   1,
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
		case '+', '-', '*', '/', '(', ')', '=', ';':
			token = l.createToken(int(r), string(r))
			l.Index++
			l.Column++
		case '\n':
			l.Column = 1
			l.Line++
			continue
		case ' ', '　':
			l.Index++
			l.Column++
			continue
		default:
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				token = l.parseIdentifier()
			} else if r >= '0' && r <= '9' {
				token = l.parseNumber()
			}
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func (l *Lexer) createToken(t int, v string) *Token {
	return &Token{
		Type:   t,
		Value:  v,
		Line:   l.Line,
		Column: l.Column,
	}
}

func (l *Lexer) parseNumber() *Token {
	runes := []rune{}
	for {
		r := l.Runes[l.Index]
		if r >= '0' && r <= '9' {
			runes = append(runes, r)
		} else {
			break
		}
		l.Index++
		l.Column++
		if len(l.Runes) <= l.Index {
			break
		}
	}
	return l.createToken(TK_NUMBER, string(runes))
}

func (l *Lexer) parseIdentifier() *Token {
	runes := []rune{}
	for {
		r := l.Runes[l.Index]
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9' && len(runes) > 0) {
			runes = append(runes, r)
		} else {
			break
		}
		l.Index++
		l.Column++
		if len(l.Runes) <= l.Index {
			break
		}
	}
	ident := string(runes)
	if v, ok := reservationTypes[ident]; ok {
		return l.createToken(v, ident)
	}
	return l.createToken(TK_IDENT, ident)
}
