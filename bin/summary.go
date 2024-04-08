/*
Copyright 2024 Georgia Institute of Technology

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Based on zgrab2/bin/summary.go
package bin

import (
	"aliasv6"
)

// Summary holds the results of a run of a Aliasv6 binary.
type Summary struct {
	Status    *aliasv6.State `json:"status"`
	StartTime string         `json:"start"`
	EndTime   string         `json:"end"`
	Duration  string         `json:"duration"`
}
