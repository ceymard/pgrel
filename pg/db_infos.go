package pg

// Introspection of the database.
type DbInfos struct {
	Types   []Type
	TypeMap map[int]*Type

	Functions []Function
	Relations []Relation

	FunctionMap map[string]*Function
	RelationMap map[string]*Relation
}

func NewDbInfos() *DbInfos {
	return &DbInfos{
		TypeMap:     make(map[int]*Type),
		FunctionMap: make(map[string]*Function),
		RelationMap: make(map[string]*Relation),
	}
}
