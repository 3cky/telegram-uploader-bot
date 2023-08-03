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

package uploader

import (
	"fmt"

	"github.com/golang/glog"

	"github.com/3cky/telegram-uploader-bot/bot"
	"github.com/3cky/telegram-uploader-bot/config"
	"github.com/3cky/telegram-uploader-bot/tagger"
	"github.com/3cky/telegram-uploader-bot/watcher"
)

type Uploader struct {
	tgBot *bot.Bot

	tasks []*Task

	eventCh chan watcher.Event

	stopCh chan struct{}
	doneCh chan struct{}
}

type Task struct {
	id        uint
	watcher   *watcher.Watcher
	chatId    int64
	documents bool
	taggers   []tagger.Taggable
}

func NewUploader(config *config.Config) (*Uploader, error) {
	// Check telegram bot token is set and is not empty
	if config.Telegram.Token == "" {
		return nil, fmt.Errorf("telegram bot token is not set or empty")
	}

	// Create telegram bot
	tgBot, err := bot.NewBot(config.Telegram.Token)
	if err != nil {
		return nil, fmt.Errorf("can't create telegram bot: %v", err)
	}

	// Create watch tasks
	eventCh := make(chan watcher.Event, 100) // FIXME make event buffer size configurable
	tasks := make([]*Task, 0)
	var id uint
	for _, u := range config.Uploads {
		// Create task taggables
		tags := make([]tagger.Taggable, 0)
		pt, err := tagger.NewPlainTagger(u.Tags.Plain)
		if err != nil {
			return nil, err
		}
		tags = append(tags, pt)

		rt, err := tagger.NewRegexpTagger(u.Tags.Regexp)
		if err != nil {
			return nil, fmt.Errorf("tag regexp: %v", err)
		}
		tags = append(tags, rt)

		et, err := tagger.NewExprTagger(u.Tags.Expr)
		if err != nil {
			return nil, fmt.Errorf("tag expr: %v", err)
		}
		tags = append(tags, et)

		// Create task watcher
		w, err := watcher.NewWatcher(id, eventCh, u.Directory, u.FilePatterns)
		if err != nil {
			glog.Warningf("can't watch %s: %v", u.Directory, err)
			continue
		}

		task := &Task{
			id:        id,
			watcher:   w,
			chatId:    u.ChatId,
			documents: u.Documents,
			taggers:   tags,
		}
		tasks = append(tasks, task)

		id++
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("no directories to watch for new files")
	}

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	return &Uploader{
		tgBot:   tgBot,
		tasks:   tasks,
		eventCh: eventCh,
		stopCh:  stopCh,
		doneCh:  doneCh,
	}, nil
}

func (u *Uploader) Start() {
	glog.V(1).Infoln("file uploader started")

	defer close(u.eventCh)
	defer close(u.doneCh)
	defer glog.V(1).Infoln("file uploader stopped")

	for _, t := range u.tasks {
		go t.watcher.Start()
	}

	for {
		select {
		case e := <-u.eventCh:
			fp := e.Path
			glog.V(4).Infof("new file to upload: %s", fp)
			t := u.tasks[e.Id]
			tags := make([]string, 0)
			for _, tg := range t.taggers {
				tags = append(tags, tg.Tags(fp)...)
			}
			if err := u.tgBot.UploadFile(t.chatId, fp, t.documents, tags...); err != nil {
				glog.Errorf("can't upload file %s to chat %d: %v", e.Path, t.chatId, err)
			}
			continue
		case <-u.stopCh:
			return
		}
	}
}

func (u *Uploader) Stop() {
	glog.V(1).Infoln("stopping file uploader...")
	for _, t := range u.tasks {
		t.watcher.Stop()
	}
	close(u.stopCh)
	<-u.doneCh
}
