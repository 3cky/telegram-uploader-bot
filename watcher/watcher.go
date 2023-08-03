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

package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"

	"github.com/rjeczalik/notify"
)

type Watcher struct {
	id           uint
	dir          string
	filePatterns []string
	notifyCh     chan notify.EventInfo
	eventCh      chan Event
	stopCh       chan struct{}
	doneCh       chan struct{}
}

type Event struct {
	Id   uint
	Path string
}

func NewWatcher(id uint, eventCh chan Event, d string, filePatterns []string) (*Watcher, error) {
	// Check d exist and is a directory
	fileinfo, err := os.Stat(d)
	if err != nil {
		return nil, err
	}
	if !fileinfo.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", d)
	}

	// Validate file name patterns
	for _, fp := range filePatterns {
		if _, err := filepath.Match(fp, d); err != nil {
			return nil, fmt.Errorf("invaild file name pattern: %s", fp)
		}
	}

	// Create filesystem watcher
	notifyCh := make(chan notify.EventInfo, 3)
	err = notify.Watch(d, notifyCh, notify.InCloseWrite, notify.InMovedTo)
	if err != nil {
		return nil, err
	}

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	return &Watcher{
		id:           id,
		dir:          d,
		filePatterns: filePatterns,
		notifyCh:     notifyCh,
		eventCh:      eventCh,
		stopCh:       stopCh,
		doneCh:       doneCh,
	}, nil
}

func (w *Watcher) Start() {
	glog.V(3).Infof("starting watcher [%d] (%s)", w.id, w.dir)

	defer notify.Stop(w.notifyCh)
	defer close(w.doneCh)
	defer glog.V(3).Infof("watcher [%d] stopped", w.id)

	nameMatcher := func(name string) bool {
		if len(w.filePatterns) == 0 {
			return true
		}
		for _, fp := range w.filePatterns {
			// Case insensitive file name pattern match
			if matched, _ := filepath.Match(strings.ToLower(fp), strings.ToLower(name)); matched {
				return true
			}
		}
		return false
	}

	for {
		select {
		case e := <-w.notifyCh:
			glog.V(4).Infof("watcher [%d] event: %v", w.id, e)
			path := e.Path()
			name := filepath.Base(path)
			if matched := nameMatcher(name); matched {
				w.eventCh <- Event{
					Id:   w.id,
					Path: path,
				}
			}
			continue
		case <-w.stopCh:
			return
		}
	}
}

func (w *Watcher) Stop() {
	glog.V(3).Infof("stopping watcher [%d] (%s)", w.id, w.dir)
	close(w.stopCh)
	<-w.doneCh
}
