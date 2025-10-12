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

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/tozd/go/errors"
)

// Introspection of the database.
type Db struct {
	Pool *pgxpool.Pool

	Types   []Type
	TypeMap map[string]*Type

	Functions []Function
	Relations []InfoRelation

	FunctionMap map[string]*Function
	RelationMap map[string]*InfoRelation
}

// Create a database connection and fill the informations
func NewDb(uri string) (*Db, error) {
	pool, err := pgxpool.New(context.Background(), uri)
	if err != nil {
		return nil, errors.Errorf("failed to create pool: %w", err)
	}

	var db = &Db{
		Pool:        pool,
		TypeMap:     make(map[string]*Type),
		FunctionMap: make(map[string]*Function),
		RelationMap: make(map[string]*InfoRelation),
	}

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		panic(err)
	}
	defer conn.Release()

	db.Fill(conn.Conn())

	return db, nil
}

// Fill informations from the database
func (db *Db) Fill(conn *pgx.Conn) error {
	if err := FillTypeInformations(db, conn); err != nil {
		return err
	}

	for _, t := range db.Types {
		db.TypeMap[t.Oid] = &t
	}

	if err := FillFunctionInformations(db, conn); err != nil {
		return err
	}

	for _, f := range db.Functions {
		db.FunctionMap[f.Identifier.String()] = &f
	}

	// if err := FillRelationInformations(infos, conn); err != nil {
	// 	return err
	// }

	return nil
}
