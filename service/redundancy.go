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
// @date 2018. 11. 14. AM 8:57
//

package service

import (
	"fmt"
	"sync"
	"throosea.com/fatima/lib"
	"throosea.com/saturn/domain"
	"time"
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
		//log.Info("clear %d old event", len(removeIdList))
		fmt.Printf("clear %d old event", len(removeIdList))
	}
}
