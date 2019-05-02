package main

import (
	"io/ioutil"
	"os"
)

func main() {
	content, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	parse(string(content))
}

func parse(str string) {
	l := NewLexer(str)
	tokens := l.Tokenize(str)
	p := NewParser(tokens)
	declarations := p.Parse()
	g := NewGenerator(p.Strings)
	g.generate(declarations)
}
