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
// Based on zgrab2/scanner.go
package aliasv6

import (
	"aliasv6/radix"
	"errors"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"inet.af/netaddr"
)

// LookUpResponse is the result of a lookup on a single ip
type LookUpResponse struct {
	// IP and Status are required for all lookups.
	IP        string       `json:"ip"`
	Status    LookUpStatus `json:"status"`
	Result    interface{}  `json:"result,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
	Error     string       `json:"error,omitempty"`
}

// RunLookUp runs a single lookup on a target and returns the resulting data
func RunLookUp(l *radix.Radix, mon *Monitor, target net.IP, expanded bool) LookUpResponse {
	t := time.Now()
	label := l.LookUp(target)
	var status LookUpStatus
	var err string
	if label.Aliased {
		mon.statusesChan <- statusSuccess
		status = LOOKUP_SUCCESS
		err = ""
	} else {
		mon.statusesChan <- statusFailure
		status = LOOKUP_NO_MATCH
		err = NewLookUpError(LOOKUP_NO_MATCH, errors.New(label.Metadata)).Err.Error()
	}
	var srcIPStr string
	if expanded {
		if srcNetAddrIP, ok := netaddr.FromStdIPRaw(target); ok {
			srcIPStr = srcNetAddrIP.StringExpanded()
		} else {
			log.Warnf("cannot expand IP address %s", target)
			srcIPStr = target.String()
		}
	} else {
		srcIPStr = target.String()
	}
	resp := LookUpResponse{IP: srcIPStr, Result: label, Error: err, Timestamp: t.Format(time.RFC3339), Status: status}
	return resp
}
