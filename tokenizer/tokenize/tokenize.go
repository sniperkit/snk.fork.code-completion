/*
Sniperkit-Bot
- Status: analyzed
*/

package tokenize

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"strings"
)

func TokenizeScope(content []byte, pos int, full bool) TokenList {
	if len(content) >= pos {
		if pos < 0 {
			pos += len(content)
		}
		contentSoFar := string(content[:pos])
		idx := strings.LastIndex(contentSoFar, "\nfunc ")
		if idx >= 0 {
			return TokenizeRange(content, idx, pos, full)
		}
	}

	f, err := parser.ParseFile(token.NewFileSet(), "file.go", content, 0)
	if err != nil {
		return TokenizeRange(content, 0, pos, full)
	}

	tokens := tokenizeCurrentBlock(content, f, pos, full)
	if len(tokens) == 0 {
		tokens = TokenizeRange(content, 0, pos, full)
	}

	return tokens
}

func tokenizeCurrentBlock(content []byte, f *ast.File, pos int, full bool) TokenList {
	for _, obj := range f.Scope.Objects {
		node, ok := obj.Decl.(ast.Node)
		if !ok {
			continue
		}

		if node.Pos() < token.Pos(pos) && node.End() > token.Pos(pos) {
			return TokenizeRange(content, int(node.Pos()-1), pos, full)
		}
	}

	return nil
}

func TokenizeRange(content []byte, start, end int, full bool) TokenList {
	if len(content)-1 < end {
		end = len(content) - 1
	}

	return Tokenize(string(content[start:end]), full)
}

func Identifiers(content []byte) []string {
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(content))
	s.Init(file, []byte(content), nil, 0)

	var tokens []string
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			return tokens
		}

		if tok == token.IDENT {
			tokens = append(tokens, lit)
		}
	}
}

func Tokenize(content string, full bool) TokenList {
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(content))
	s.Init(file, []byte(content), nil, 0)

	var tokens TokenList
	var lastTok token.Token
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			if lastTok == token.SEMICOLON {
				tokens = tokens[:len(tokens)-1]
			}
			break
		}

		tokens = append(tokens, NewToken(tok, lit))
		if full &&
			tok == token.IDENT &&
			lit != "false" &&
			lit != "nil" &&
			lit != "true" {
			tokens = append(tokens, &fixToken{lit})
		}
		lastTok = tok
	}

	return tokens
}

type TokenList []Token

func (l TokenList) String() string {
	var parts = make([]string, len(l))
	for i := 0; i < len(l); i++ {
		parts[i] = l[i].String()
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}

type Token interface {
	isToken()
	fmt.Stringer
}

func NewToken(tok token.Token, lit string) Token {
	switch tok {
	case token.IMAG:
		return varToken{ImagLit}
	case token.INT:
		return varToken{IntLit}
	case token.STRING:
		return varToken{StrLit}
	case token.FLOAT:
		return varToken{FloatLit}
	case token.CHAR:
		return varToken{CharLit}
	case token.IDENT:
		switch lit {
		case "true", "false":
			return varToken{BoolLit}
		case "nil", "err":
			return fixToken{lit}
		}

		return varToken{Ident}
	}

	return fixToken{tok.String()}
}

type varToken struct {
	kind VarTokenKind
}

func (v varToken) String() string {
	switch v.kind {
	case StrLit:
		return "ID_LIT_STR"
	case IntLit:
		return "ID_LIT_INT"
	case FloatLit:
		return "ID_LIT_FLOAT"
	case BoolLit:
		return "ID_LIT_BOOL"
	case CharLit:
		return "ID_LIT_CHAR"
	case ImagLit:
		return "ID_LIT_IMAG"
	case Ident:
		return "ID_S"
	default:
		panic("unreachable")
	}
}
func (varToken) isToken() {}

type VarTokenKind byte

const (
	Ident VarTokenKind = iota
	StrLit
	IntLit
	FloatLit
	BoolLit
	CharLit
	ImagLit
)

type fixToken struct {
	val string
}

func (f fixToken) String() string {
	return fmt.Sprintf("%q", f.val)
}
func (fixToken) isToken() {}
