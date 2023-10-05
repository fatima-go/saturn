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

package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	log "github.com/fatima-go/fatima-log"
)

const (
	MessageKeyType       = "type"   // type : notify level
	MessageKeyAction     = "action" // action : kind of action
	MessageKeyAlarmLevel = "alarm_level"
	MessageKeyCategory   = "category"
	MessageKeyMessage    = "message"
	MessageKeyDeployment = "deployment"
	AlarmLevelWarn       = "WARN"
	AlarmLevelMinor      = "MINOR"
	AlarmLevelMajor      = "MAJOR"
)

const (
	NotifyAlarm          = "ALARM"
	ActionProcessStartup = "PROCESS_STARTUP"
)

type MessageNotify interface {
	SendNotify(mbus MBusMessageBody)
}

type MBusMessage struct {
	Header MBusMessageHeader `json:"header"`
	Body   MBusMessageBody   `json:"body"`
}

type MBusMessageHeader struct {
	ApplicationCode int `json:"application_code"`
	Logic           int `json:"logic"`
}

type MBusMessageBody struct {
	EventTime      int                    `json:"event_time"`
	Message        map[string]interface{} `json:"message"`
	PackageGroup   string                 `json:"package_group"`
	PackageHost    string                 `json:"package_host"`
	PackageName    string                 `json:"package_name"`
	PackageProcess string                 `json:"package_process"`
	PackageProfile string                 `json:"package_profile,omitempty"`
}

func (m MBusMessageBody) GetMessageText(fmonUrl string) interface{} {
	buff := bytes.Buffer{}
	if txt, ok := m.Message[MessageKeyMessage].(string); ok {
		buff.WriteString(txt)
	}

	if !m.IsAlarm() || !m.IsProcessStartup() {
		return buff.String()
	}

	dep := m.GetDeployment()
	if !dep.Valid || !dep.HasBuildInfo() {
		return buff.String()
	}

	// xxx process started
	// deployment : djin.chung, master, hashxxxx
	// build message : something has fixed....
	// fmon : https://fmon.music-flo.io/xxx
	buff.WriteString(fmt.Sprintf("\ndeploy user : %s", dep.Build.BuildUser))
	buff.WriteString(fmt.Sprintf("\nbuild time : %s", dep.Build.BuildTime))
	if !dep.Build.HasGit() {
		return buff.String()
	}
	buff.WriteString(fmt.Sprintf("\ngit commit : %s (%s)", dep.Build.Git.Commit, dep.Build.Git.Branch))
	buff.WriteString(fmt.Sprintf("\ngit message : %s", GetTrimmedMessage(dep.Build.Git.Message)))

	if len(fmonUrl) > 10 {
		link := fmt.Sprintf(fmonUrl, m.PackageHost, m.PackageProcess)
		buff.WriteString(fmt.Sprintf("\n<%s|배포 히스토리 보기>\n", link))
	}

	return buff.String()
}

func (m MBusMessageBody) IsAlarm() bool {
	switch m.Message[MessageKeyType] {
	case NotifyAlarm:
		return true
	}
	return false
}

func (m MBusMessageBody) IsProcessStartup() bool {
	switch m.Message[MessageKeyAction] {
	case ActionProcessStartup:
		return true
	}
	return false
}

func (m MBusMessageBody) GetCategory() string {
	category, ok := m.Message[MessageKeyCategory]
	if !ok {
		return ""
	}
	if s, ok := category.(string); ok {
		return s
	}
	return ""
}

func (m MBusMessageBody) getFootprint() string {
	msg, ok := m.Message[MessageKeyMessage]
	if !ok {
		return ""
	}

	return fmt.Sprintf("%s.%s.%s.%s.%s.%s",
		m.PackageGroup,
		m.PackageHost,
		m.PackageName,
		m.PackageProcess,
		m.PackageProfile,
		msg)
}

func (m MBusMessageBody) GetHashsum() string {
	footprint := m.getFootprint()

	hashing := sha256.New()
	hashing.Write([]byte(footprint))
	return fmt.Sprintf("%x", hashing.Sum(nil))
}

func (m MBusMessageBody) GetDeployment() Deployment {
	deployment := Deployment{Valid: false}

	msg, ok := m.Message[MessageKeyDeployment]
	if !ok {
		return deployment
	}

	b, err := json.Marshal(msg)
	if err != nil {
		log.Warn("fail to make deployment json : %s", err.Error())
		return deployment
	}

	err = json.Unmarshal(b, &deployment)
	if err != nil {
		log.Warn("fail to unmarshal deployment : %s", err.Error())
		return deployment
	}

	deployment.Valid = true
	return deployment
}
