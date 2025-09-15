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
	"sync"
	"time"

	"github.com/fatima-go/fatima-core/lib"
	log "github.com/fatima-go/fatima-log"
	"github.com/fatima-go/saturn/domain"
)

const (
	diffLimit = 1000 * 3 // 3 sec
)

var mutex sync.Mutex
var eventMap map[string]int
var mapClearTick *time.Ticker

func isRedundant(mbus domain.MBusMessageBody) bool {
	mutex.Lock()
	defer mutex.Unlock()

	key := mbus.GetHashsum()
	lastTime, ok := eventMap[key]
	if !ok {
		eventMap[key] = lib.CurrentTimeMillis()
		return false
	}

	if lib.CurrentTimeMillis()-lastTime > diffLimit {
		eventMap[key] = lib.CurrentTimeMillis()
		return false
	}

	eventMap[key] = lib.CurrentTimeMillis()
	return true
}

func init() {
	eventMap = make(map[string]int)
	mapClearTick = time.NewTicker(time.Minute * 1)
	go func() {
		for range mapClearTick.C {
			clearEventMap()
		}
	}()
}

func clearEventMap() {
	mutex.Lock()
	defer mutex.Unlock()

	removeIdList := make([]string, 0)
	for k, v := range eventMap {
		if lib.CurrentTimeMillis()-v > diffLimit {
			removeIdList = append(removeIdList, k)
		}
	}

	for _, k := range removeIdList {
		delete(eventMap, k)
	}

	if len(removeIdList) > 0 {
		log.Info("clear %d old event", len(removeIdList))
		//fmt.Printf("clear %d old event", len(removeIdList))
	}
}
