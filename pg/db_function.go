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
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"gitlab.com/tozd/go/errors"
)

var sqlGetFunctions = `WITH fulltypes AS (` + typeQuery + /* sql */ `)
SELECT json_agg(S) FROM	(SELECT
  json_build_object(
    'Schema', n.nspname,
    'Name', p.proname
  ) AS "Identifier",
	t."Type" AS "ReturnType",
  l.lanname AS "Language",
  p.proretset AS "ReturnsSet",
  p.proisstrict AS "IsStrict",
  p.prosecdef AS "IsSetuid",
  t.typtype = 'c' AS "ReturnsComposite",
  p.provolatile = 'i' AS "IsImmutable",
  p.provolatile = 's' AS "IsStable",
  p.provolatile = 'v' AS "IsVolatile",
  (
    SELECT json_agg(S) FROM (
			SELECT
				argnb AS "Index",
				p.proargnames[argnb] AS "Name",
				p.proargmodes[argnb] AS "Mode",
				t."Type" AS "Type"
			FROM generate_series(1, array_length(proallargtypes, 1)) argnb
			INNER JOIN fulltypes t ON t.oid = p.proallargtypes[argnb-1]
		) S) AS "Arguments"
  FROM pg_proc p
  LEFT JOIN pg_namespace n ON p.pronamespace = n.oid
  LEFT JOIN pg_language l ON p.prolang = l.oid
  LEFT JOIN fulltypes t ON t.oid = p.prorettype
  ORDER BY n.nspname, p.proname) S;
`

type FunctionArgument struct {
	Index      int
	Name       string
	Type       Type
	Mode       string // 'i'IN, 'o' OUT, 'b' INOUT, 'v' VARIADIC
	IsVariadic bool
}

type Function struct {
	Identifier SqlIdentifier

	Arguments []FunctionArgument

	ReturnsSet       bool // Whether this function returns a table() or a setof ReturnType
	ReturnsComposite bool // Always true if ReturnsSet ?
	ReturnType       Type

	// Other function attributes that are not relevant as of now
	IsStrict            bool // proisstrict
	IsSetUid            bool
	IsVolatile          bool
	IsLeakproof         bool
	IsCalledOnNullInput bool
	IsImmutable         bool
	IsStable            bool
}

// Query the database and fill the infos
func FillFunctionInformations(infos *DbInfos, conn *pgx.Conn) error {
	rows, err := conn.Query(context.Background(), sqlGetFunctions)
	if err != nil {
		return errors.Errorf("failed to query functions: %w", err)
	}
	defer rows.Close()

	var funcs []Function
	if !rows.Next() {
		return errors.Errorf("no functions found")
	}

	var jsonstr string
	err = rows.Scan(&jsonstr)
	if err != nil {
		return errors.Errorf("failed to scan functions: %w", err)
	}

	err = json.Unmarshal([]byte(jsonstr), &funcs)
	if err != nil {
		return errors.Errorf("failed to unmarshal functions: %w", err)
	}

	return nil
}

// IsExportable returns true if the function is exportable to the web.
// It can only be exported if all its arguments are IN, the return type is OUT and all arguments are named.
func (f *Function) IsExportable() bool {
	// return f.Arguments[0].Mode == "IN" && f.Arguments[len(f.Arguments)-1].Mode == "OUT"
	return false
}
