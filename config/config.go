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

package config

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DefaultConfigFile = "/usr/local/etc/telegram-uploader-bot.cfg"

	ConfigKeyBotToken = "telegram.token"
)

var ConfigFile string

type Config struct {
	Telegram Telegram
	Uploads  []Upload
}

type Telegram struct {
	Token string
}

type Upload struct {
	Directory    string
	FilePatterns []string `mapstructure:"files"`
	ChatId       int64    `mapstructure:"chat"`
	Documents    bool
	Tags         Tags
}

type Tags struct {
	Plain  []string
	Regexp []string
	Expr   []string
}

func NewConfig(cmd *cobra.Command) (*Config, error) {
	// Create empty config
	config := new(Config)

	// Read config from file
	err := readConfig(cmd)
	if err != nil {
		return nil, err
	}

	// Parse config
	err = viper.Unmarshal(&config, func(m *mapstructure.DecoderConfig) {
		m.ErrorUnused = true
	})
	if err != nil {
		return nil, err
	}

	return config, nil
}

func readConfig(cmd *cobra.Command) error {
	viper.AddConfigPath(".") // adding home directory as first search path
	viper.SetConfigFile(ConfigFile)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("can't parse config file: %w", err)
	}

	glog.V(1).Infof("using config file: %s", viper.ConfigFileUsed())

	return nil
}
