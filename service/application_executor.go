//
// Copyright (c) 2017 SK TECHX.
// All right reserved.
//
// This software is the confidential and proprietary information of SK TECHX.
// You shall not disclose such Confidential Information and
// shall use it only in accordance with the terms of the license agreement
// you entered into with SK TECHX.
//
//
// @project saturn
// @author 1100282
// @date 2017. 11. 3. PM 1:39
//

package service

import (
	"strings"
	"throosea.com/fatima"
	"throosea.com/fatima/builder"
	"throosea.com/log"
	"throosea.com/saturn/domain"
	"throosea.com/saturn/notifier/slack"
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
