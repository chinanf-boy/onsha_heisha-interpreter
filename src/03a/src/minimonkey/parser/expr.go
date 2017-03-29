package parser

import (
	"fmt"
	"minimonkey/ast"
	"minimonkey/token"
	"strconv"
)

func (p *Parser) parseExpression(precedence int) (ast.Expression, error) {
	var err error
	var leftExp ast.Expression

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		return leftExp, fmt.Errorf("no prefix parse function for %s found", p.curToken.Type)
	}
	leftExp, err = prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp, nil
		}

		p.nextToken()

		leftExp, err = infix(leftExp)
	}

	return leftExp, err
}

func (p *Parser) parseIdentifier() (ast.Expression, error) {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}, nil
}

func (p *Parser) parseIntegerLiteral() (ast.Expression, error) {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse integer %q", p.curToken.Literal)
	}

	lit.Value = value

	return lit, nil
}

func (p *Parser) parsePrefixExpression() (ast.Expression, error) {
	exp := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	right, err := p.parseExpression(PREFIX)
	if err != nil {
		return nil, err
	}
	exp.Right = right

	return exp, nil
}

func (p *Parser) parseInfixExpression(left ast.Expression) (ast.Expression, error) {
	exp := &ast.InfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	precedence := p.curPrecedence()

	p.nextToken()

	right, err := p.parseExpression(precedence)
	if err != nil {
		return nil, err
	}

	exp.Right = right

	return exp, nil
}

func (p *Parser) parseGroupedExpression() (ast.Expression, error) {
	p.nextToken()

	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	if !p.expectPeek(token.RPAREN) {
		return nil, p.peekError(token.RPAREN)
	}

	return exp, nil
}

// fn(<identifier>...) { <statement>... }
func (p *Parser) parseFunctionLiteral() (ast.Expression, error) {
	var err error
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil, p.peekError(token.LPAREN)
	}

	lit.Parameters, err = p.parseFunctionParameters()
	if err != nil {
		return nil, err
	}

	if !p.expectPeek(token.LBRACE) {
		return nil, p.peekError(token.LBRACE)
	}

	lit.Body, err = p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	return lit, nil
}

func (p *Parser) parseFunctionParameters() ([]*ast.Identifier, error) {
	identifiers := []*ast.Identifier{}

	// fn() { ... }
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken() // curToken == RPAREN
		return identifiers, nil
	}

	if !p.expectPeek(token.IDENT) {
		return nil, p.peekError(token.IDENT)
	}

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // curToken == COMMA
		p.nextToken() // curToken == IDENT
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil, p.peekError(token.RPAREN)
	}

	return identifiers, nil
}

// <identifier or function_literal or integer_literal>(expression...)
func (p *Parser) parseCallExpression(function ast.Expression) (ast.Expression, error) {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	args, err := p.parseCallArguments()
	if err != nil {
		return nil, err
	}
	exp.Arguments = args
	return exp, nil
}

func (p *Parser) parseCallArguments() ([]ast.Expression, error) {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken() // curToken == RPAREN
		return args, nil
	}

	p.nextToken()
	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	args = append(args, exp)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // curToken == COMMA
		p.nextToken()
		exp, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		args = append(args, exp)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil, p.peekError(token.RPAREN)
	}

	return args, nil
}
