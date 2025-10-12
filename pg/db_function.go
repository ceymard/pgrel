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

type FunctionArgument struct {
	Index      int    `json:"index"`
	Name       string `json:"name"`
	Type       *Type  `json:"type"`
	Mode       string `json:"mode"`
	IsVariadic bool   `json:"is_variadic"`
}

type Function struct {
	Identifier SqlIdentifier `json:"identifier"`

	Arguments []*FunctionArgument `json:"args"`

	// Whether this function returns a table() or a setof ReturnType
	ReturnsSet bool  `json:"returns_set"`
	ReturnType *Type `json:"return_type"`

	// Other function attributes that are not relevant as of now
	IsStrict            bool `json:"is_strict"`
	IsVolatile          bool `json:"is_volatile"`
	IsLeakproof         bool `json:"is_leakproof"`
	IsCalledOnNullInput bool `json:"is_called_on_null_input"`
	IsImmutable         bool `json:"is_immutable"`
	IsStable            bool `json:"is_stable"`
}
