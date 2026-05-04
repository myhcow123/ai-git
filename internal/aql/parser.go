package aql

import (
	"fmt"
	"strings"
)

type TokenType int

const (
	TokenSymbol TokenType = iota
	TokenDependencies
	TokenSearch
	TokenModify
	TokenCreate
	TokenWhere
	TokenWith
	TokenLimit
	TokenDepth
	TokenTypeName
	TokenDirection
	TokenString
	TokenNumber
	TokenIdentifier
	TokenLParen
	TokenRParen
	TokenComma
	TokenEqual
	TokenEOF
)

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	input  string
	pos    int
	tokens []Token
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		tokens: []Token{},
	}
}

func (l *Lexer) Tokenize() []Token {
	for l.pos < len(l.input) {
		l.skipWhitespace()

		if l.pos >= len(l.input) {
			break
		}

		ch := l.input[l.pos]

		if isLetter(ch) {
			l.readIdentifier()
		} else if isDigit(ch) {
			l.readNumber()
		} else if ch == '"' || ch == '\'' {
			l.readString(ch)
		} else if ch == '(' {
			l.tokens = append(l.tokens, Token{Type: TokenLParen, Value: "("})
			l.pos++
		} else if ch == ')' {
			l.tokens = append(l.tokens, Token{Type: TokenRParen, Value: ")"})
			l.pos++
		} else if ch == ',' {
			l.tokens = append(l.tokens, Token{Type: TokenComma, Value: ","})
			l.pos++
		} else if ch == '=' {
			l.tokens = append(l.tokens, Token{Type: TokenEqual, Value: "="})
			l.pos++
		} else {
			l.pos++
		}
	}

	l.tokens = append(l.tokens, Token{Type: TokenEOF, Value: ""})
	return l.tokens
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && isWhitespace(l.input[l.pos]) {
		l.pos++
	}
}

func (l *Lexer) readIdentifier() {
	start := l.pos
	for l.pos < len(l.input) && isLetter(l.input[l.pos]) {
		l.pos++
	}

	value := strings.ToUpper(l.input[start:l.pos])

	switch value {
	case "SYMBOL":
		l.tokens = append(l.tokens, Token{Type: TokenSymbol, Value: value})
	case "DEPENDENCIES":
		l.tokens = append(l.tokens, Token{Type: TokenDependencies, Value: value})
	case "SEARCH":
		l.tokens = append(l.tokens, Token{Type: TokenSearch, Value: value})
	case "MODIFY":
		l.tokens = append(l.tokens, Token{Type: TokenModify, Value: value})
	case "CREATE":
		l.tokens = append(l.tokens, Token{Type: TokenCreate, Value: value})
	case "WHERE":
		l.tokens = append(l.tokens, Token{Type: TokenWhere, Value: value})
	case "WITH":
		l.tokens = append(l.tokens, Token{Type: TokenWith, Value: value})
	case "LIMIT":
		l.tokens = append(l.tokens, Token{Type: TokenLimit, Value: value})
	case "DEPTH":
		l.tokens = append(l.tokens, Token{Type: TokenDepth, Value: value})
	case "TYPE":
		l.tokens = append(l.tokens, Token{Type: TokenTypeName, Value: value})
	case "DIRECTION":
		l.tokens = append(l.tokens, Token{Type: TokenDirection, Value: value})
	default:
		l.tokens = append(l.tokens, Token{Type: TokenIdentifier, Value: l.input[start:l.pos]})
	}
}

func (l *Lexer) readNumber() {
	start := l.pos
	for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
		l.pos++
	}
	l.tokens = append(l.tokens, Token{Type: TokenNumber, Value: l.input[start:l.pos]})
}

func (l *Lexer) readString(quote byte) {
	l.pos++
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != quote {
		l.pos++
	}
	l.tokens = append(l.tokens, Token{Type: TokenString, Value: l.input[start:l.pos]})
	l.pos++
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

type QueryType int

const (
	QuerySymbol QueryType = iota
	QueryDependencies
	QuerySearch
	QueryModify
	QueryCreate
)

type Query struct {
	Type          QueryType
	Target        string
	Condition     *Condition
	Options       *Options
	Modifications []Modification
}

type Condition struct {
	Field    string
	Operator string
	Value    interface{}
}

type Options struct {
	With      []string
	Limit     int
	Depth     int
	Type      string
	Direction string
}

type Modification struct {
	Type  string
	Value interface{}
}

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

func (p *Parser) Parse() (*Query, error) {
	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("empty query")
	}

	token := p.tokens[p.pos]

	switch token.Type {
	case TokenSymbol:
		return p.parseSymbolQuery()
	case TokenDependencies:
		return p.parseDependenciesQuery()
	case TokenSearch:
		return p.parseSearchQuery()
	case TokenModify:
		return p.parseModifyQuery()
	case TokenCreate:
		return p.parseCreateQuery()
	default:
		return nil, fmt.Errorf("unknown query type: %s", token.Value)
	}
}

func (p *Parser) parseSymbolQuery() (*Query, error) {
	p.pos++

	if p.pos >= len(p.tokens) || p.tokens[p.pos].Type != TokenIdentifier {
		return nil, fmt.Errorf("expected symbol name")
	}

	query := &Query{
		Type:    QuerySymbol,
		Target:  p.tokens[p.pos].Value,
		Options: &Options{},
	}
	p.pos++

	for p.pos < len(p.tokens) && p.tokens[p.pos].Type != TokenEOF {
		token := p.tokens[p.pos]

		switch token.Type {
		case TokenWith:
			p.pos++
			for p.pos < len(p.tokens) && p.tokens[p.pos].Type == TokenIdentifier {
				query.Options.With = append(query.Options.With, p.tokens[p.pos].Value)
				p.pos++
			}
		case TokenWhere:
			p.pos++
			cond, err := p.parseCondition()
			if err != nil {
				return nil, err
			}
			query.Condition = cond
		default:
			p.pos++
		}
	}

	return query, nil
}

func (p *Parser) parseDependenciesQuery() (*Query, error) {
	p.pos++

	if p.pos >= len(p.tokens) || p.tokens[p.pos].Type != TokenIdentifier {
		return nil, fmt.Errorf("expected symbol name")
	}

	query := &Query{
		Type:    QueryDependencies,
		Target:  p.tokens[p.pos].Value,
		Options: &Options{Depth: 1},
	}
	p.pos++

	for p.pos < len(p.tokens) && p.tokens[p.pos].Type != TokenEOF {
		token := p.tokens[p.pos]

		switch token.Type {
		case TokenDepth:
			p.pos++
			if p.pos < len(p.tokens) && p.tokens[p.pos].Type == TokenNumber {
				fmt.Sscanf(p.tokens[p.pos].Value, "%d", &query.Options.Depth)
				p.pos++
			}
		case TokenDirection:
			p.pos++
			if p.pos < len(p.tokens) && p.tokens[p.pos].Type == TokenIdentifier {
				query.Options.Direction = p.tokens[p.pos].Value
				p.pos++
			}
		default:
			p.pos++
		}
	}

	return query, nil
}

func (p *Parser) parseSearchQuery() (*Query, error) {
	p.pos++

	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("expected search query")
	}

	query := &Query{
		Type:    QuerySearch,
		Target:  p.tokens[p.pos].Value,
		Options: &Options{Limit: 10},
	}
	p.pos++

	for p.pos < len(p.tokens) && p.tokens[p.pos].Type != TokenEOF {
		token := p.tokens[p.pos]

		switch token.Type {
		case TokenLimit:
			p.pos++
			if p.pos < len(p.tokens) && p.tokens[p.pos].Type == TokenNumber {
				fmt.Sscanf(p.tokens[p.pos].Value, "%d", &query.Options.Limit)
				p.pos++
			}
		case TokenTypeName:
			p.pos++
			if p.pos < len(p.tokens) && p.tokens[p.pos].Type == TokenIdentifier {
				query.Options.Type = p.tokens[p.pos].Value
				p.pos++
			}
		default:
			p.pos++
		}
	}

	return query, nil
}

func (p *Parser) parseModifyQuery() (*Query, error) {
	p.pos++

	if p.pos >= len(p.tokens) || p.tokens[p.pos].Type != TokenIdentifier {
		return nil, fmt.Errorf("expected symbol name")
	}

	query := &Query{
		Type:    QueryModify,
		Target:  p.tokens[p.pos].Value,
		Options: &Options{},
	}
	p.pos++

	return query, nil
}

func (p *Parser) parseCreateQuery() (*Query, error) {
	p.pos++

	if p.pos >= len(p.tokens) || p.tokens[p.pos].Type != TokenIdentifier {
		return nil, fmt.Errorf("expected type name")
	}

	query := &Query{
		Type:    QueryCreate,
		Target:  p.tokens[p.pos].Value,
		Options: &Options{},
	}
	p.pos++

	return query, nil
}

func (p *Parser) parseCondition() (*Condition, error) {
	if p.pos >= len(p.tokens) || p.tokens[p.pos].Type != TokenIdentifier {
		return nil, fmt.Errorf("expected field name")
	}

	cond := &Condition{
		Field: p.tokens[p.pos].Value,
	}
	p.pos++

	if p.pos >= len(p.tokens) || p.tokens[p.pos].Type != TokenEqual {
		return nil, fmt.Errorf("expected operator")
	}
	cond.Operator = "="
	p.pos++

	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("expected value")
	}

	switch p.tokens[p.pos].Type {
	case TokenString, TokenNumber, TokenIdentifier:
		cond.Value = p.tokens[p.pos].Value
		p.pos++
	default:
		return nil, fmt.Errorf("expected value")
	}

	return cond, nil
}

func ParseAQL(query string) (*Query, error) {
	lexer := NewLexer(query)
	tokens := lexer.Tokenize()
	parser := NewParser(tokens)
	return parser.Parse()
}
