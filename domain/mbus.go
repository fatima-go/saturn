//
// Copyright (c) 2018 SK TECHX.
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
// @date 2018. 8. 14. PM 4:34
//

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
