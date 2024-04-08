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
// Based on zgrab2/monitor.go
package aliasv6

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Monitor is a collection of states per lookup and a channel to communicate
// those lookups to the monitor
type Monitor struct {
	state        *State
	statusesChan chan status
	// Callback is invoked after each lookup.
	Callback func(string)
}

// State contains the respective number of successes and failures
// for a given lookup
type State struct {
	Successes uint `json:"successes"`
	Failures  uint `json:"failures"`
}

type status uint

const (
	statusSuccess status = iota
	statusFailure status = iota
)

// GetStatuses returns the current number
// of successes and failures for the lookup operations
func (m *Monitor) GetStatus() *State {
	return m.state
}

func (m *Monitor) GetStatusChan() chan status {
	return m.statusesChan
}

// Stop indicates the monitor is done and the internal channel should be closed.
// This function does not block, but will allow a call to Wait() on the
// WaitGroup passed to MakeMonitor to return.
func (m *Monitor) Stop() {
	close(m.statusesChan)
}

// MakeMonitor returns a Monitor object that can be used to collect and send
// the status of a running lookup
func MakeMonitor(statusChanSize int, wg *sync.WaitGroup) *Monitor {
	m := new(Monitor)
	m.statusesChan = make(chan status, statusChanSize)
	m.state = &State{}
	wg.Add(1)
	timerReady := new(sync.WaitGroup)
	timerReady.Add(1)
	var timerWorkerDone sync.WaitGroup
	timerWorkerDone.Add(1)
	ticker := time.NewTicker(time.Duration(time.Second))
	quitTimerChannel := make(chan struct{})

	go func() {
		tickerCount := uint(0)
		lastTotal := uint(0)
		timerReady.Done()
		for {
			select {
			case <-ticker.C:
				tickerCount++
				ticker.Stop()
				success := m.state.Successes
				failure := m.state.Failures
				log.Infof("Total Processed: %d (%.2f IPs/sec; +m: %d) -> Aliased: %d; No-match: %d", success+failure, float64(success+failure)/float64(tickerCount), (success+failure)-lastTotal, success, failure)
				ticker.Reset(time.Duration(time.Second))
				lastTotal = success + failure
			case <-quitTimerChannel:
				ticker.Stop()
				timerWorkerDone.Done()
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		timerReady.Wait()
		for s := range m.statusesChan {
			switch s {
			case statusSuccess:
				m.state.Successes++
			case statusFailure:
				m.state.Failures++
			default:
				continue
			}
		}
		quitTimerChannel <- struct{}{}
		timerWorkerDone.Wait()
	}()
	return m
}
