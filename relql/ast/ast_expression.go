package ast

type IAstExpression interface {
}

type AstBinaryExpression struct {
	Left     IAstExpression
	Right    IAstExpression
	Operator string
}
