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
	"crypto/sha256"
	"fmt"
)

const (
	MessageKeyType       = "type"
	MessageKeyAlarmLevel = "alarm_level"
	MessageKeyCategory   = "category"
	MessageKeyMessage    = "message"
	AlarmLevelWarn       = "WARN"
	AlarmLevelMinor      = "MINOR"
	AlarmLevelMajor      = "MAJOR"
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

func (mbus MBusMessageBody) IsAlarm() bool {
	switch mbus.Message[MessageKeyType] {
	case "ALARM":
		return true
	}
	return false
}

func (mbus MBusMessageBody) GetCategory() string {
	category, ok := mbus.Message[MessageKeyCategory]
	if !ok {
		return ""
	}
	if s, ok := category.(string); ok {
		return s
	}
	return ""
}

func (mbus MBusMessageBody) getFootprint() string {
	msg, ok := mbus.Message[MessageKeyMessage]
	if !ok {
		return ""
	}

	return fmt.Sprintf("%s.%s.%s.%s.%s.%s",
		mbus.PackageGroup,
		mbus.PackageHost,
		mbus.PackageName,
		mbus.PackageProcess,
		mbus.PackageProfile,
		msg)
}

func (mbus MBusMessageBody) GetHashsum() string {
	footprint := mbus.getFootprint()

	hashing := sha256.New()
	hashing.Write([]byte(footprint))
	return fmt.Sprintf("%x", hashing.Sum(nil))
}
