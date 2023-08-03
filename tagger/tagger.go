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

package tagger

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/golang/glog"
)

type Taggable interface {
	Tags(filePath string) []string
}

type PlainTagger struct {
	Taggable

	tags []string
}

func NewPlainTagger(tags []string) (*PlainTagger, error) {
	return &PlainTagger{
		tags: tags,
	}, nil
}

func (pt *PlainTagger) Tags(filePath string) []string {
	return pt.tags
}

type RegexpTagger struct {
	Taggable

	tagRegexps []*regexp.Regexp
}

func NewRegexpTagger(tagRegexps []string) (*RegexpTagger, error) {
	trs := make([]*regexp.Regexp, 0)
	for _, tr := range tagRegexps {
		regexp, err := regexp.Compile(tr)
		if err != nil {
			return nil, err
		}
		if len(regexp.SubexpNames()) < 2 { // first subexp always is regexp itself
			return nil, fmt.Errorf("tag regexp must have groups: %s", tr)
		}
		trs = append(trs, regexp)
	}
	return &RegexpTagger{
		tagRegexps: trs,
	}, nil
}

func (rt *RegexpTagger) Tags(filePath string) []string {
	tags := make([]string, 0)
	for _, tr := range rt.tagRegexps {
		sms := tr.FindStringSubmatch(filePath)
		if sms == nil {
			continue
		}
		smns := tr.SubexpNames()
		for i, smn := range smns {
			if i == 0 {
				continue
			}
			if i >= len(sms) {
				break
			}
			tag := smn + sms[i] // concatenate submatch/group name (if defined) and submatch itself
			tag = strings.TrimSpace(tag)
			if len(tag) > 0 {
				tags = append(tags, tag)
			}
		}
	}
	return tags
}

type ExprTagger struct {
	Taggable

	tagExprs []*vm.Program
}

func NewExprTagger(tagExprs []string) (*ExprTagger, error) {
	tes := make([]*vm.Program, 0)
	for _, te := range tagExprs {
		expr, err := expr.Compile(te)
		if err != nil {
			return nil, fmt.Errorf("can't compile tag expr [%s]: %v", te, err)
		}
		tes = append(tes, expr)
	}
	return &ExprTagger{
		tagExprs: tes,
	}, nil
}

func (et *ExprTagger) Tags(filePath string) []string {
	tags := make([]string, 0)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		glog.Errorf("can't stat file %s: %v", filePath, err)
		return tags
	}
	env := map[string]interface{}{
		"path": filePath,
		"file": fileInfo,

		"sprintf": fmt.Sprintf,
	}
	for _, te := range et.tagExprs {
		itag, err := expr.Run(te, env)
		if err != nil {
			glog.Errorf("can't tag expr file %s: %v", filePath, err)
			continue
		}
		tag := strings.TrimSpace(fmt.Sprintf("%v", itag))
		if len(tag) > 0 {
			tags = append(tags, tag)
		}
	}
	return tags
}
