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

// A query we will use to fetch complex type information
var typeQuery = /* sql */ `
SELECT
  t.oid AS oid,
  t_elem.oid AS elem_oid,
  t_elem.typname AS elem_name,
  n_elem.nspname AS elem_schema,
  t.typname AS type_name,
  n.nspname AS type_schema,
	coalesce(t_elem.typtype = 'c', t.typtype = 'c') AS "IsComposite",
  t.typelem <> 0 AND t_elem.typname IS NOT NULL AS "IsArray",
  json_build_object(
		'Name', COALESCE(t_elem.typname, t.typname),
		'Schema', COALESCE(n_elem.nspname, n.nspname),
    'IsArray', t.typelem <> 0 AND t_elem.typname IS NOT NULL,
		'IsComposite', coalesce(t_elem.typtype = 'c', t.typtype = 'c'),
    'Oid', COALESCE(t_elem.oid, t.oid)::text,
    'OidArray', CASE WHEN t_elem.oid IS NOT NULL THEN t.oid::text ELSE NULL END,
		'RelId', coalesce(t_elem.typrelid, t.typrelid)::text
  ) AS "Type"
FROM
  pg_type t
  INNER JOIN pg_namespace n ON n.oid = t.typnamespace
  LEFT JOIN pg_type t_elem ON t.typelem = t_elem.oid AND t_elem.typarray = t.oid
  LEFT JOIN pg_namespace n_elem ON n_elem.oid = t_elem.typnamespace
`

type Type struct {
	SqlIdentifier
	IsArray     bool
	IsComposite bool

	Oid      string
	OidArray string // If IsArray, the oid of the array type
	RelId    string // When this type is a composite type
}

// Query the database and fill the infos
func FillTypeInformations(infos *DbInfos, conn *pgx.Conn) error {
	rows, err := conn.Query(context.Background(), "SELECT json_agg(t.\"Type\") FROM ("+typeQuery+") t")
	if err != nil {
		return errors.Errorf("failed to query types: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return errors.Errorf("no types found")
	}

	var jsonstr string

	err = rows.Scan(&jsonstr)
	if err != nil {
		return errors.Errorf("failed to scan types: %w", err)
	}

	err = json.Unmarshal([]byte(jsonstr), &infos.Types)
	if err != nil {
		return errors.Errorf("failed to unmarshal types: %w", err)
	}

	return nil
}
