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

type Column struct {
	Name      string
	PgTypeOid int

	Type *Type

	DefaultExpression string

	IsPrimaryKey bool
	IsIdentity   bool
	IsGenerated  bool
	IsUnique     bool
	IsNotNull    bool
	IsNullable   bool
}

// A table or view
type Relation struct {
	Identifier SqlIdentifier

	IsView             bool
	IsMaterializedView bool

	Columns    []*Column
	ColumnsMap map[string]*Column

	PrimaryKey     []string
	UniqueTogether [][]string
	IndexedColumns [][]string // On top of PrimaryKey and UniqueTogether, columns that are susceptible to be used for joining on this table as an incoming multiple-row foreign key.

	Type *Type // The related type

	// Unique list of foreign keys for this relation
	OutgoingForeignKeys []*OutgoingForeignKey
	IncomingForeignKeys []*IncomingForeignKey

	outgoingForeignKeysMap map[string]*OutgoingForeignKey
	incomingForeignKeysMap map[string]*IncomingForeignKey

	PgRelId   int
	PgTypeOid int
}

func (r *Relation) GetOutgoingFkByName(name string) *OutgoingForeignKey {
	return r.outgoingForeignKeysMap[name]
}
func (r *Relation) GetIncomingFkByName(name string) *IncomingForeignKey {
	return r.incomingForeignKeysMap[name]
}

func FillRelationInformations(infos *DbInfos, conn *pgx.Conn) error {
	if err := scanIntoThroughJsonAgg(conn, INFO_QUERY_RELATIONS, &infos.Relations); err != nil {
		return err
	}

	return nil
}

var INFO_QUERY_RELATIONS = /* sql */ `
SELECT json_agg(R) FROM (SELECT

	pg_class.oid::integer AS "PgRelId",

	json_build_object(
		'Schema', pg_class.relnamespace::regnamespace,
		'Name', pg_class.relname
	) AS "Identifier",

	json_agg(json_build_object(
		'Name', column_name,
		'Index', ordinal_position,
		'DefaultExpression', CASE WHEN is_identity = 'YES' AND pg_get_serial_sequence(col.table_schema || '.' || col.table_name, col.column_name) IS NOT NULL
			THEN 'nextval(''' || pg_get_serial_sequence(col.table_schema || '.' || col.table_name, col.column_name) || ''')'
			ELSE column_default
		END,
		'IsNullable', is_nullable = 'YES',
		'IsSelfReferencing', is_self_referencing = 'YES',
		'IsIdentity', is_identity = 'YES',
		'PgTypeOid', (SELECT t.oid::INT FROM pg_type t WHERE t.typname = udt_name AND t.typnamespace = udt_schema::regnamespace),
		'DomainIdentifier', CASE WHEN domain_schema IS NULL THEN NULL ELSE json_build_object(
			'Schema', domain_schema,
			'Name', domain_name
		) END
	) ORDER BY ordinal_position
	) AS "Columns"
FROM information_schema.columns col
INNER JOIN pg_class ON pg_class.relname = col.table_name AND pg_class.relnamespace = col.table_schema::regnamespace
GROUP BY
pg_class.oid, pg_class.relnamespace, pg_class.relname
) R;`
