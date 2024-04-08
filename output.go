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
// Based on zgrab2/output.go
package aliasv6

import (
	"bufio"
	"io"
)

// OutputResultsWriterFunc returns an OutputResultsFunc that wraps an io.Writer
// in a buffered writer, and uses OutputResults.
func OutputResultsWriterFunc(w io.Writer) OutputResultsFunc {
	buf := bufio.NewWriter(w)
	return func(result <-chan []byte) error {
		defer buf.Flush()
		return OutputResults(buf, result)
	}
}

// OutputResults writes results to a buffered Writer from a channel.
func OutputResults(w *bufio.Writer, results <-chan []byte) error {
	for result := range results {
		if _, err := w.Write(result); err != nil {
			return err
		}
		if err := w.WriteByte('\n'); err != nil {
			return err
		}
		if config.Flush {
			w.Flush()
		}
	}
	return nil
}

// OutputResultsFunc is a function type for result output functions.
//
// A function of this type receives results on the provided channel
// and outputs them somehow.  It returns nil if there are no further
// results or error.
type OutputResultsFunc func(results <-chan []byte) error
