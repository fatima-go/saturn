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
	"fmt"
	"github.com/fatima-go/fatima-core/lib"
	"github.com/fatima-go/saturn/domain"
	"testing"
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
