/*
 * Copyright 2023 github.com/fatima-go
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @project fatima-core
 * @author jin
 * @date 23. 4. 14. 오후 6:07
 */

package service

import (
	"github.com/fatima-go/fatima-core"
	"github.com/fatima-go/fatima-core/builder"
	log "github.com/fatima-go/fatima-log"
	"github.com/fatima-go/saturn/domain"
	"github.com/fatima-go/saturn/notifier/slack"
	"strings"
)

const (
	categoryMonitor        = "monitor"
	processJuno            = "juno"
	propMessageNotifyChain = "message.notify.chain"
)

var opmProcessList = []string{"jupiter", "juno", "saturn"}

type ApplicationExecutor interface {
	Consume(m domain.MBusMessage)
}

func NewFatimaApplicationExecutor(fatimaRuntime fatima.FatimaRuntime) ApplicationExecutor {
	app := FatimaApplicationExecutor{}
	app.fatimaRuntime = fatimaRuntime
	app.notifyChain = buildMessageNotifyChain(fatimaRuntime)
	return &app
}

func buildMessageNotifyChain(fatimaRuntime fatima.FatimaRuntime) []domain.MessageNotify {
	chain := make([]domain.MessageNotify, 0)
	values, ok := fatimaRuntime.GetConfig().GetValue(propMessageNotifyChain)
	if !ok {
		// add slack to default
		chain = append(chain, slack.NewSlackNotification(fatimaRuntime))
		log.Info("load notify chain : SLACK")
		return chain
	}

	for _, v := range strings.Split(strings.TrimSpace(values), ",") {
		if strings.ToLower(v) == "slack" {
			chain = append(chain, slack.NewSlackNotification(fatimaRuntime))
			log.Info("load notify chain : SLACK")
		}
		// TODO : more notifier will be added in future
		// TODO : file, db, tcp, ...
	}
	return chain
}

type FatimaApplicationExecutor struct {
	fatimaRuntime fatima.FatimaRuntime
	notifyChain   []domain.MessageNotify
}

func toLogicString(logicNo int) string {
	switch logicNo {
	case builder.LogicMeasure:
		return "measure"
	case builder.LogicNotify:
		return "notify"
	}
	return "unknown logic code"
}

func (f *FatimaApplicationExecutor) Consume(m domain.MBusMessage) {
	if log.IsDebugEnabled() {
		log.Debug("%s :: %s:%s:%s:%s:%s",
			toLogicString(m.Header.Logic),
			m.Body.PackageGroup,
			m.Body.PackageHost,
			m.Body.PackageName,
			m.Body.PackageProcess,
			m.Body.PackageProfile,
		)
	}

	if m.Header.Logic == builder.LogicMeasure {
		return
	}

	if isOpmProcess(m.Body) {
		return
	}

	log.Info("%v", m.Body)
	if isRedundant(m.Body) {
		return
	}

	for _, c := range f.notifyChain {
		c.SendNotify(m.Body)
	}
}

func isOpmProcess(msg domain.MBusMessageBody) bool {
	for _, v := range opmProcessList {
		if msg.PackageProcess == v {
			if msg.PackageProcess == processJuno && msg.GetCategory() == categoryMonitor {
				return false
			}
			return true
		}
	}

	return false
}
