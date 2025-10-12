// Copyright 2025 Christophe Eymard
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pg

import "github.com/jackc/pgx/v5"

// Introspection of the database.
type DbInfos struct {
	Types   []Type
	TypeMap map[string]*Type

	Functions []Function
	Relations []InfoRelation

	FunctionMap map[string]*Function
	RelationMap map[string]*InfoRelation
}

func NewDbInfos() *DbInfos {
	return &DbInfos{
		TypeMap:     make(map[string]*Type),
		FunctionMap: make(map[string]*Function),
		RelationMap: make(map[string]*InfoRelation),
	}
}

func (infos *DbInfos) Fill(conn *pgx.Conn) error {
	if err := FillTypeInformations(infos, conn); err != nil {
		return err
	}

	for _, t := range infos.Types {
		infos.TypeMap[t.Oid] = &t
	}

	if err := FillFunctionInformations(infos, conn); err != nil {
		return err
	}

	for _, f := range infos.Functions {
		infos.FunctionMap[f.Identifier.String()] = &f
	}

	// if err := FillRelationInformations(infos, conn); err != nil {
	// 	return err
	// }

	return nil
}
