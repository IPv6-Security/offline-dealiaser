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
// Based on zgrab2/input.go
package aliasv6

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type Command struct {
	Type       string      `json:"type"`
	Data       string      `json:"data"`
	ParsedData interface{} `json:"pdata,omitempty"`
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func duplicateIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

// InputTargets is an InputTargetsFunc that calls GetTargets with
// the an input file provided on the command line.
func InputTargets(ch chan<- Command) error {
	return GetTargets(config.inputFile, ch)
}

// GetTargets reads targets from a source, generates LookUpTargets,
// and delivers them to the provided channel.
func GetTargets(source io.Reader, ch chan<- Command) error {
	reader := bufio.NewReader(source)
	for {
		var target string
		var err error
		var command Command
		message, err := reader.ReadBytes('\n')
		if err == io.EOF {
			end := time.Now()
			log.Infof("no more input is coming: %s", end.Format(time.RFC3339))
			break
		} else if err != nil {
			return err
		}
		// if config.InputType == "ip" {
		// 	target = string(message)
		// 	if target == "quit" {
		// 		end := time.Now()
		// 		log.Infof("quit command has been received; quitting at %s", end.Format(time.RFC3339))
		// 		break
		// 	} else {
		// 		command.Data = target
		// 		command.Type = "lookup"
		// 	}
		// } else {
		err = json.Unmarshal(message, &command)
		if err != nil {
			target = string(message)
			command.Data = target
			command.Type = "lookup"
		} else {
			target = command.Data
		}
		// }
		if command.Type == "quit" {
			end := time.Now()
			log.Infof("quit command has been received; quitting at %s", end.Format(time.RFC3339))
			break
		}
		ipnet, err := ParseTarget(target)
		if err != nil {
			log.Errorf("parse error, skipping: %v", err)
			continue
		}
		command.ParsedData = ipnet
		// if command.Type == "insert" {
		// 	log.Printf("%+v\n", command)
		// }
		var ip net.IP
		if ipnet != nil {
			if ipnet.Mask != nil {
				if command.Type == "lookup" {
					// expand CIDR block into one target for each IP
					for ip = ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
						command.ParsedData = duplicateIP(ip)
						ch <- command
					}
					continue
				}
			} else {
				if command.Type == "insert" {
					return errors.New("cannot insert an IP, it should be an IP Network in CIDR notation")
				}
				command.ParsedData = ipnet.IP
			}
		}
		ch <- command
	}
	return nil
}

// ParseTarget takes a record from an input file and
// returns the specified ipnet or an error.
//
// AliasV6 input files have only one field:
//
//	IP
//
// Each line specifies a target to perform a lookup by its IP address.
// A CIDR block may be provided in the IP field, in which case the
// framework expands the record into targets for every address in the
// block.
func ParseTarget(target string) (ipnet *net.IPNet, err error) {
	target = strings.TrimSpace(target)

	if target != "" {
		if ip := net.ParseIP(target); ip != nil {
			ipnet = &net.IPNet{IP: ip}
		} else if _, cidr, er := net.ParseCIDR(target); er == nil {
			ipnet = cidr
		}
	}

	if ipnet == nil {
		err = fmt.Errorf("record doesn't specify an address or network: %s", target)
		return
	}
	return
}

// InputTargetsFunc is a function type for target input functions.
//
// A function of this type generates Labels on the provided
// channel.  It returns nil if there are no further inputs or error.
type InputTargetsFunc func(ch chan<- Command) error
