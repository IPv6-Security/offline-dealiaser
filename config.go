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
// Based on zgrab2/config.go

package aliasv6

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// Config is the high level framework options that will be parsed
// from the command line
type Config struct {
	OutputFileName      string  `short:"o" long:"output-file" default:"-" description:"Output filename, use - for stdout"`
	InputFileName       string  `short:"f" long:"input-file" default:"-" description:"Input filename, use - for stdin"`
	MetaFileName        string  `short:"m" long:"metadata-file" default:"-" description:"Metadata filename, use - for stderr"`
	LogFileName         string  `short:"l" long:"log-file" default:"-" description:"Log filename, use - for stderr"`
	ConstructInputFile  string  `short:"c" long:"construct-input-file" description:"List of ips/prefixes input file to construct Tree/Trie for tests"`
	CheckpointBaseName  string  `long:"checkpoint-base-name" default:"checkpoint" description:"Base name for the Tree/Trie checkpoints if there is a change. It will be followed by the timestamp of the checkpoint"`
	CheckpointFrequency float32 `long:"checkpoint-frequency" default:"30.0" description:"Frequency in seconds to export Tree/Trie checkpoints"`
	TestType            string  `long:"test-type" default:"radix" choice:"radix" choice:"stats" choice:"stress" description:"Testing mode"`
	TestInputFile       string  `long:"test-input-file" description:"List of ips input file for Tree/Trie for tests (e.x. lookup tests for radix)"`
	TestOutputFile      string  `long:"test-output-file" description:"File to export results of stats test mode"`
	Test                bool    `short:"t" long:"test" description:"Set program to testing mode"`
	Flush               bool    `long:"flush" description:"Flush after each line of output."`
	Expanded            bool    `long:"expanded" description:"Print IPs in an expanded format"`
	TestStepSize        int     `long:"test-step-size" default:"1000000" description:"Checkpoint or logging step size for tests"`
	NumLookUpWorkers    int     `long:"num-lookup-workers" default:"1000" description:"Number of workers to perform concurrent lookup operations"`
	// InputType           string  `long:"input-type" default:"command" choice:"command" choice:"ip" description:"Input feed type. Command has to be in JSON format, and ip is a IPv6 address as a string."`
	inputFile     *os.File
	outputFile    *os.File
	metaFile      *os.File
	logFile       *os.File
	InputTargets  InputTargetsFunc
	OutputResults OutputResultsFunc
}

var config Config

// SetInputFunc sets the target input function to the provided function.
func SetInputFunc(f InputTargetsFunc) {
	config.InputTargets = f
}

// SetOutputFunc sets the result output function to the provided function.
func SetOutputFunc(f OutputResultsFunc) {
	config.OutputResults = f
}

func validateFrameworkConfiguration() {
	// validate files
	if config.LogFileName == "-" {
		config.logFile = os.Stderr
	} else {
		var err error
		if config.logFile, err = os.Create(config.LogFileName); err != nil {
			log.Fatal(err)
		}
		log.Infof("log file is being set to %s", config.LogFileName)
		log.SetOutput(config.logFile)
	}

	if config.InputFileName == "-" {
		config.inputFile = os.Stdin
	} else {
		var err error
		if config.inputFile, err = os.Open(config.InputFileName); err != nil {
			log.Fatal(err)
		}
	}
	SetInputFunc(InputTargets)

	if config.OutputFileName == "-" {
		config.outputFile = os.Stdout
	} else {
		var err error
		if config.outputFile, err = os.Create(config.OutputFileName); err != nil {
			log.Fatal(err)
		}
	}
	outputFunc := OutputResultsWriterFunc(config.outputFile)
	SetOutputFunc(outputFunc)

	if config.MetaFileName == "-" {
		config.metaFile = os.Stderr
	} else {
		var err error
		if config.metaFile, err = os.Create(config.MetaFileName); err != nil {
			log.Fatal(err)
		}
	}

	if config.Test && config.TestInputFile == "" {
		log.Fatal("test input file should be provided in any test mode")
	} else {
		if config.TestType == "stats" && config.TestOutputFile == "" {
			log.Fatal("test output file should be provided in stats test mode")
		}
	}

	//validate senders
	if config.NumLookUpWorkers <= 0 {
		log.Fatalf("need at least one lookup worker, given %d", config.NumLookUpWorkers)
	}
}

// GetInputFile returns the file to which the program receives input from
func GetInputFile() *os.File {
	return config.inputFile
}

// GetOutputFile returns the file to which program writes the output
func GetOutputFile() *os.File {
	return config.outputFile
}

// GetMetaFile returns the file to which metadata should be output
func GetMetaFile() *os.File {
	return config.metaFile
}

// GetLogFile returns the file to which log should be output
func GetLogFile() *os.File {
	return config.logFile
}
