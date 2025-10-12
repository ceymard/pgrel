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

package main

import (
	"fmt"
	"os"

	"github.com/ceymard/pgrel/pg"
)

func main() {
	var srv string
	if len(os.Args) > 1 {
		srv = os.Args[1]
	}

	if infos, err := pg.NewDb(srv); err != nil {
		fmt.Println("Failed to create db:", err)
		printStackTrace(err)
		return
	} else {
		fmt.Println(infos)
	}

}
