// Copyright 2023 Victor Antonovich <v.antonovich@gmail.com>
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

package log

import (
	"flag"
	"log"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	"github.com/3cky/telegram-uploader-bot/util"
)

var logFlushFreq = pflag.Duration("log-flush-frequency", 5*time.Second, "Maximum number of seconds between log flushes")

func init() {
	// Don't steal command line arguments from `go test`
	if flag.Lookup("test.v") != nil {
		return
	}
	// Trick to avoid 'logging before flag.Parse' warning
	if err := flag.CommandLine.Parse([]string{}); err != nil {
		return
	}
	_ = flag.Set("logtostderr", "true")
}

// GlogWriter serves as a bridge between the standard log package and the glog package.
type GlogWriter struct{}

// Write implements the io.Writer interface.
func (writer GlogWriter) Write(data []byte) (n int, err error) {
	glog.Info(string(data))
	return len(data), nil
}

// Initialize logging
func InitLog() {
	log.SetOutput(GlogWriter{})
	log.SetFlags(0)
	// The default glog flush interval is 30 seconds, which is frighteningly long.
	go util.Until(FlushLog, *logFlushFreq, util.NeverStop)
}

// Flushes log immediately
func FlushLog() {
	glog.Flush()
}
