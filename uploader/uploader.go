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
	"context"
	"fmt"
	"os"

	"github.com/golang/glog"

	"github.com/3cky/telegram-uploader-bot/bot"
	"github.com/3cky/telegram-uploader-bot/config"
	"github.com/3cky/telegram-uploader-bot/tagger"
	"github.com/3cky/telegram-uploader-bot/watcher"
)

const MAX_UPLOAD_SIZE = 50 * 1024 * 1024 // 50 MB is default Telegram API file size limit

type Uploader struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	tgBot *bot.Bot

	tasks []*Task

	eventCh chan watcher.Event
	doneCh  chan struct{}
}

type Task struct {
	id       uint
	watcher  *watcher.Watcher
	minSize  uint64
	maxSize  uint64
	chatId   int64
	document bool
	taggers  []tagger.Taggable
}

func NewUploader(ctx context.Context, config *config.Config) (*Uploader, error) {
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
		// Check file size limits
		minSize := u.MinSize.Bytes()
		maxSize := u.MaxSize.Bytes()
		if maxSize == 0 {
			maxSize = MAX_UPLOAD_SIZE
		}
		if minSize > maxSize {
			return nil, fmt.Errorf("max upload size (%d) must not be less than min size (%d)", maxSize, minSize)
		}

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
			id:       id,
			watcher:  w,
			minSize:  minSize,
			maxSize:  maxSize,
			chatId:   u.ChatId,
			document: u.Document,
			taggers:  tags,
		}
		tasks = append(tasks, task)

		id++
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("no directories to watch for new files")
	}

	// Create cancelable context
	ctxWithCancel, ctxCancel := context.WithCancel(ctx)

	doneCh := make(chan struct{})

	return &Uploader{
		ctx:       ctxWithCancel,
		ctxCancel: ctxCancel,
		tgBot:     tgBot,
		tasks:     tasks,
		eventCh:   eventCh,
		doneCh:    doneCh,
	}, nil
}

func (u *Uploader) Start() {
	glog.V(1).Infoln("file uploader started")

	defer close(u.eventCh)
	defer close(u.doneCh)
	defer u.ctxCancel()
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
			// Check file size
			fi, err := os.Stat(fp)
			if err != nil {
				glog.Errorf("can't stat file to upload %s: %v", fp, err)
				continue
			}
			if fi.Size() < int64(t.minSize) {
				glog.V(3).Infof("skipping uploading of too small file (%d byte(s)): %s", fi.Size(), fp)
				continue
			}
			if fi.Size() > int64(t.maxSize) {
				glog.Warningf("skipping uploading of too big file (%d byte(s)): %s", fi.Size(), fp)
				continue
			}
			// Get file tags
			tags := make([]string, 0)
			for _, tg := range t.taggers {
				tags = append(tags, tg.Tags(fp)...)
			}
			// Upload file to Telegram
			if err := u.tgBot.UploadFile(u.ctx, t.chatId, fp, t.document, tags...); err != nil {
				glog.Errorf("can't upload file %s to chat %d: %v", e.Path, t.chatId, err)
			}
			continue
		case <-u.ctx.Done():
			return
		}
	}
}

func (u *Uploader) Stop() {
	glog.V(1).Infoln("stopping file uploader...")
	for _, t := range u.tasks {
		t.watcher.Stop()
	}
	u.ctxCancel()
	<-u.doneCh
}
