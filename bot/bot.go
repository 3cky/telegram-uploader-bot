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

package bot

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/time/rate"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang/glog"

	"github.com/3cky/telegram-uploader-bot/util"
)

type Chat struct {
}

type Bot struct {
	botApi *tgbotapi.BotAPI

	rateLimiter *rate.Limiter
}

func NewBot(token string) (*Bot, error) {
	// Create new telegram Bot
	botApi, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	// Check created bot
	if _, err := botApi.GetUpdates(tgbotapi.UpdateConfig{}); err != nil {
		return nil, err
	}

	// Create rate limiter (one request per second w/ burst 5 rps)
	rateLimiter := rate.NewLimiter(rate.Every(time.Second), 5)

	return &Bot{
		botApi:      botApi,
		rateLimiter: rateLimiter,
	}, nil
}

func (b *Bot) UploadFile(ctx context.Context, chatId int64, filePath string, document bool, tags ...string) error {
	glog.V(4).Infof("uploading file %s with tags %v to chat %d", filePath, tags, chatId)

	// Convert tags to hashtags
	ht := strings.Join(tags, " #")
	if len(ht) > 0 {
		ht = "#" + ht
	}

	fp := tgbotapi.FilePath(filePath)
	var m tgbotapi.Chattable
	if !document {
		m = b.getMediaMessage(chatId, fp, ht)
	}
	if m == nil {
		m = b.getDocumentMessage(chatId, fp, ht)
	}

	err := b.rateLimiter.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = b.botApi.Send(m)

	return err
}

func (b *Bot) getMediaMessage(chatId int64, fp tgbotapi.FilePath, caption string) tgbotapi.Chattable {
	var m tgbotapi.Chattable
	fn := filepath.Base(string(fp))
	if util.IsFileExtensionMatched(fn, "mp3", "m4a") {
		am := tgbotapi.NewAudio(chatId, fp)
		am.Caption = caption
		m = am
	} else if util.IsFileExtensionMatched(fn, "mp4") {
		vm := tgbotapi.NewVideo(chatId, fp)
		vm.Caption = caption
		m = vm
	} else if util.IsFileExtensionMatched(fn, "jpg", "jpeg", "png", "gif") {
		pm := tgbotapi.NewPhoto(chatId, fp)
		pm.Caption = caption
		m = pm
	}
	return m
}

func (b *Bot) getDocumentMessage(chatId int64, fp tgbotapi.FilePath, caption string) tgbotapi.Chattable {
	dm := tgbotapi.NewDocument(chatId, fp)
	dm.Caption = caption
	return dm
}
