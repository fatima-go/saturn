//
// Copyright (c) 2018 SK Planet.
// All right reserved.
//
// This software is the confidential and proprietary information of K Planet.
// You shall not disclose such Confidential Information and
// shall use it only in accordance with the terms of the license agreement
// you entered into with SK Planet.
//
//
// @project saturn
// @author 1100282
// @date 2018. 11. 14. AM 8:59
//

package service

import (
	"fmt"
	"testing"
	"throosea.com/fatima/lib"
	"throosea.com/saturn/domain"
	"time"
)

func TestRedundancy(t *testing.T) {
	m1 := buildSampleMBusBody("sample process shutdowned")
	m2 := buildSampleMBusBody("sample process started")
	m3 := buildSampleMBusBody("sample process shutdowned.")

	fmt.Printf("[%s]\n", m1.GetHashsum())
	fmt.Printf("[%s]\n", m2.GetHashsum())
	fmt.Printf("[%s]\n", m3.GetHashsum())

	if isRedundant(m1) {
		t.Fatalf("m1 should not be redundant")
		return
	}

	if isRedundant(m2) {
		t.Fatalf("m2 should not be redundant")
		return
	}

	if isRedundant(m3) {
		t.Fatalf("m3 should not be redundant")
		return
	}

	if !isRedundant(m1) {
		t.Fatalf("m1 should BE redundant")
		return
	}

	time.Sleep(time.Minute * 2)

	if isRedundant(m2) {
		t.Fatalf("m2 should not be redundant")
		return
	}
}

func buildSampleMBusBody(msg string) domain.MBusMessageBody {
	m := domain.MBusMessageBody{}
	m.EventTime = lib.CurrentTimeMillis()
	m.Message = make(map[string]interface{})
	m.PackageGroup = "test_group"
	m.PackageHost = "test_host"
	m.PackageName = "default"
	m.PackageProcess = "test"
	m.PackageProfile = "local"

	//message:helloworld process shutdowned timestamp:2018-11-14 09:04:48 type:ALARM alarm_level:MAJOR from:go-fatima initiator:go-fatima
	m.Message["message"] = msg
	m.Message["timestamp"] = "2018-11-14 09:04:48"
	m.Message["type"] = "ALARM"
	m.Message["from"] = "go-fatima"
	m.Message["initiator"] = "go-fatima"
	return m
}
