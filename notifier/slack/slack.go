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

package slack

import (
	"bytes"
	"encoding/json"
	"github.com/fatima-go/fatima-core"
	"github.com/fatima-go/fatima-log"
	"github.com/fatima-go/saturn/domain"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	fileWebhookSlack      = "webhook.slack"
	attachmentsColorGreen = "#00FF00"
	attachmentsColorRed   = "#FF0000"
	attachmentColorOrange = "#FFA500"
	// ##439fe0
	attachmentColorBlue      = "#439FE0"
	attachmentColorYellow    = "#FFFF00"
	userName                 = "FATIMA"
	footerIcon               = "https://platform.slack-edge.com/img/default_application_icon.png"
	applicationJsonUtf8Value = "application/json;charset=UTF-8"
	PropertyFmonUrl          = "fmon.url"

	deployCategory = "deploy"
)

func NewSlackNotification(fatimaRuntime fatima.FatimaRuntime) *SlackNotification {
	return NewSlackNotificationWithKey(fatimaRuntime, "default")
}

func NewSlackNotificationWithKey(fatimaRuntime fatima.FatimaRuntime, key string) *SlackNotification {
	slack := SlackNotification{}
	slack.fatimaRuntime = fatimaRuntime
	slack.mutex = &sync.Mutex{}
	slack.alarm.Active = false
	slack.event.Active = false
	slack.alarmCategory = make(map[string]SlackConfig)

	// load fmon property
	fmonUrl, err := fatimaRuntime.GetConfig().GetString(PropertyFmonUrl)
	if err == nil {
		slack.fmonUrl = fmonUrl
	}

	log.Info("slack.fmonUrl=[%s]", slack.fmonUrl)
	return &slack
}

type SlackNotification struct {
	fatimaRuntime   fatima.FatimaRuntime
	lastLoadingTime time.Time
	alarm           SlackConfig
	event           SlackConfig
	alarmCategory   map[string]SlackConfig
	mutex           *sync.Mutex
	fmonUrl         string
}

type SlackConfig struct {
	Active  bool
	Url     string
	Channel string
}

func (s *SlackNotification) GetFmonUrl() string {
	return s.fmonUrl
}

func (s *SlackNotification) SendNotify(mbus domain.MBusMessageBody) {
	message := s.buildSlackMessage(mbus)
	cate := mbus.GetCategory()

	if mbus.IsAlarm() {
		if mbus.IsProcessStartupOrShutdown() && len(cate) == 0 {
			cate = deployCategory
		}
		s.sendAlarm(message, cate)
	} else {
		s.sendEvent(message)
	}
}

func (s *SlackNotification) loading() {
	s.lastLoadingTime = time.Now()
	if s.fatimaRuntime == nil {
		log.Warn("fatimaRuntime is nil")
		return
	}

	webhookConfigFile := filepath.Join(s.fatimaRuntime.GetEnv().GetFolderGuide().GetDataFolder(), fileWebhookSlack)
	dataBytes, err := os.ReadFile(webhookConfigFile)
	if err != nil {
		return
	}

	var data map[string]SlackConfig
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		return
	}

	c, ok := data["alarm"]
	if !ok {
		log.Warn("alarm config is not found")
		return
	}
	s.alarm.Active = c.Active
	s.alarm.Url = c.Url

	c, ok = data["event"]
	if !ok {
		log.Warn("event config is not found")
		return
	}
	s.event.Active = c.Active
	s.event.Url = c.Url

	for k, v := range data {
		if k == "alarm" || k == "event" {
			continue
		}
		s.alarmCategory[k] = v
	}

	log.Debug("slack config loaded : alarm[%v], event[%v], alarmCategory[%d]", s.alarm, s.event, len(s.alarmCategory))
}

func (s *SlackNotification) isEventWritable() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	deadline := time.Now().Add(-time.Minute)
	if s.lastLoadingTime.Before(deadline) {
		s.loading()
	}
	if !s.event.Active || len(s.event.Url) < 6 {
		return false
	}
	return true
}

func (s *SlackNotification) isAlarmWritable() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	deadline := time.Now().Add(-time.Second * 10)
	if s.lastLoadingTime.Before(deadline) {
		s.loading()
	}
	if !s.alarm.Active || len(s.alarm.Url) < 6 {
		return false
	}
	return true
}

func (s *SlackNotification) isAlarmCategoryWritable(cate string) bool {
	if len(cate) == 0 {
		return false
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	deadline := time.Now().Add(-time.Second * 10)
	if s.lastLoadingTime.Before(deadline) {
		s.loading()
	}
	config, ok := s.alarmCategory[cate]
	if !ok {
		return false
	}

	if !config.Active || len(config.Url) < 6 {
		return false
	}
	return true
}

func (s *SlackNotification) getAlarmCategoryUrlAndChannel(cate string) (url string, channel string) {
	if len(cate) == 0 {
		return
	}

	config, ok := s.alarmCategory[cate]
	if !ok {
		return
	}

	return config.Url, config.Channel
}

func (s *SlackNotification) sendEvent(m map[string]interface{}) {
	if !s.isEventWritable() {
		return
	}

	b, err := json.Marshal(m)
	if err != nil {
		log.Warn("fail to build json : %s", err.Error())
		return
	}

	go func() {
		sendMessageToSlack(s.event.Url, b)
	}()
}

func (s *SlackNotification) sendAlarm(m map[string]interface{}, cate string) {
	if len(cate) == 0 {
		s.sendAlarmCase(m)
		return
	}

	// send with category
	if !s.isAlarmCategoryWritable(cate) {
		return
	}

	url, channel := s.getAlarmCategoryUrlAndChannel(cate)

	if len(url) > 0 {
		if len(channel) > 0 {
			m["channel"] = channel
		}

		b, err := json.Marshal(m)
		if err != nil {
			log.Warn("fail to build json : %s", err.Error())
			return
		}

		go func() {
			log.Info("sending with cate %s, url=%s", cate, url)
			sendMessageToSlack(url, b)
		}()
	}
}

func (s *SlackNotification) sendAlarmCase(m map[string]interface{}) {
	if !s.isAlarmWritable() {
		return
	}

	b, err := json.Marshal(m)
	if err != nil {
		log.Warn("fail to build json : %s", err.Error())
		return
	}

	go func() {
		sendMessageToSlack(s.alarm.Url, b)
	}()
}

func sendMessageToSlack(url string, b []byte) {
	resp, err := http.Post(url, applicationJsonUtf8Value, bytes.NewBuffer(b))
	if err != nil {
		log.Warn("fail to send slack notification : %s", err.Error())
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Debug("successfully send to slack : %d", len(b))
	} else {
		log.Info("slack response : %s", resp.Status)
	}
}

func (s *SlackNotification) buildSlackMessage(mbus domain.MBusMessageBody) map[string]interface{} {
	m := make(map[string]interface{})
	m["username"] = userName
	list := make([]interface{}, 0)
	list = append(list, s.buildAttachment(mbus))
	m["attachments"] = list

	return m
}

func (s *SlackNotification) buildAttachment(mbus domain.MBusMessageBody) map[string]interface{} {
	m := make(map[string]interface{})
	m["pretext"] = buildPretext(mbus)
	m["color"] = attachmentsColorGreen
	switch mbus.Message[domain.MessageKeyType] {
	case "ALARM":
		alevel, ok := mbus.Message[domain.MessageKeyAlarmLevel]
		if ok {
			switch alevel {
			case domain.AlarmLevelWarn:
				m["color"] = attachmentColorYellow
			case domain.AlarmLevelMinor:
				m["color"] = attachmentColorBlue
			case domain.AlarmLevelMajor:
				m["color"] = attachmentsColorRed
			}
		}
	}
	m["text"] = mbus.GetMessageText(s.GetFmonUrl())
	m["footer"] = mbus.PackageProcess
	m["footer_icon"] = footerIcon
	m["ts"] = mbus.EventTime / 1000
	return m
}

func buildPretext(mbus domain.MBusMessageBody) string {
	var buff bytes.Buffer
	if len(mbus.PackageProfile) > 0 {
		buff.WriteByte('[')
		buff.WriteString(mbus.PackageProfile)
		buff.WriteByte(']')
		buff.WriteByte(' ')
	}
	buff.WriteString(mbus.PackageGroup)
	buff.WriteByte(':')
	buff.WriteString(mbus.PackageHost)
	if mbus.PackageName != "default" {
		buff.WriteByte(':')
		buff.WriteString(mbus.PackageName)
	}
	return buff.String()
}
