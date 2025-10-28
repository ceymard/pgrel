package ast

type IAstField interface {
}

type AstField struct {
	Id    *AstSqlIdentifier
	Alias string
}
