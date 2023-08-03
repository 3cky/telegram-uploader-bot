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

package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"

	"github.com/3cky/telegram-uploader-bot/build"
	"github.com/3cky/telegram-uploader-bot/config"
	"github.com/3cky/telegram-uploader-bot/log"
	uploaderpkg "github.com/3cky/telegram-uploader-bot/uploader"
)

const (
	FlagVersion = "version"
	FlagHelpMd  = "help-md"
	FlagConfig  = "config"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "telegram-uploader-bot",
		Long: "Watches for files and uploads them to Telegram.",
		Run:  runCmd,
	}

	initCmd(cmd)

	return cmd
}

func initCmd(cmd *cobra.Command) {
	// Command-related flags set
	f := cmd.Flags()

	f.Bool(FlagVersion, false, "display the version number and build timestamp")
	f.Bool(FlagHelpMd, false, "get help in Markdown format")
	f.StringVarP(&config.ConfigFile, FlagConfig, "c", config.DefaultConfigFile, "config file")

	pflag.CommandLine.AddFlagSet(f)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// Init logging
	log.InitLog()
	defer log.FlushLog()
}

func runCmd(cmd *cobra.Command, _ []string) {
	if f, _ := cmd.Flags().GetBool(FlagVersion); f {
		fmt.Printf("Build version: %s\n", build.Version)
		fmt.Printf("Build timestamp: %s\n", build.Timestamp)
		return
	}

	if f, _ := cmd.Flags().GetBool(FlagHelpMd); f {
		out := new(bytes.Buffer)
		if err := doc.GenMarkdown(cmd, out); err != nil {
			fmt.Printf("can't generate help in markdown format: %v", err)
			return
		}
		fmt.Println(out)
		return
	}

	glog.Infof("Starting %s...", cmd.Name())

	getConfig := func() (*config.Config, error) {
		config, err := config.NewConfig(cmd)
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	config, err := getConfig()
	if err != nil {
		glog.Errorf("config read error: %v, exiting...", err)
		return
	}

	uploader, err := uploaderpkg.NewUploader(config)
	if err != nil {
		glog.Errorf("config couldn't be used: %v", err)
		return
	}

	// Start files uploading
	go uploader.Start()

	// Listen for signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	// Control loop
	for {
		sig := <-signalCh
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			glog.V(2).Infof("received %v signal", sig)
			// Stop files uploading and exit
			uploader.Stop()
			return
		case syscall.SIGHUP:
			glog.V(2).Infof("received %v signal, reloading config", sig)
			config, err := getConfig()
			if err != nil {
				glog.Errorf("config reloading error: %v", err)
				continue
			}
			newUploader, err := uploaderpkg.NewUploader(config)
			if err != nil {
				glog.Errorf("reloaded config can't be used: %v", err)
				continue
			}
			// Stop files uploading using current config
			uploader.Stop()
			// Start files uploading using new config
			uploader = newUploader
			go uploader.Start()
		}
	}
}
