//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
//
// @project saturn
// @author DeockJin Chung (jin.freestyle@gmail.com)
// @date 2017. 3. 5. PM 4:37
//

package engine

import (
	"encoding/json"
	"path/filepath"
	"throosea.com/fatima"
	"throosea.com/fatima/builder"
	"throosea.com/fatima/lib/mbus"
	"throosea.com/log"
	"throosea.com/saturn/domain"
	"throosea.com/saturn/service"
)

const (
	propWatchDir = "stat.dir"
	propMoveDir  = "stat.move.dir"
)

func NewMbusReader(fatimaRuntime fatima.FatimaRuntime, applicationExecutor service.ApplicationExecutor) *MbusReader {
	comp := new(MbusReader)
	comp.fatimaRuntime = fatimaRuntime
	comp.executors = make(map[int]service.ApplicationExecutor)
	comp.executors[builder.ApplicationCode] = applicationExecutor

	dataDir := filepath.Join(fatimaRuntime.GetEnv().GetFolderGuide().GetFatimaHome(), builder.FatimaFolderData)
	proc := fatimaRuntime.GetEnv().GetSystemProc().GetProgramName()

	reader, err := mbus.NewMappedMBusReader(dataDir, proc, comp.consume)
	if err != nil {
		log.Error("fail to create mbus reader : %s", err.Error())
		return nil
	}

	comp.reader = reader
	return comp
}

type MbusReader struct {
	fatimaRuntime fatima.FatimaRuntime
	reader        *mbus.MappedMBusReader
	executors     map[int]service.ApplicationExecutor
}

func (sm *MbusReader) Initialize() bool {
	log.Info("ListenLog Initialize()")
	err := sm.reader.Activate()
	if err != nil {
		log.Error("mbus reader activation fail : %s", err.Error())
		return false
	}

	return true
}

func (sm *MbusReader) Bootup() {
	log.Info("MbusReader Bootup()")
}

func (sm *MbusReader) Shutdown() {
	log.Info("MbusReader Shutdown()")
	if sm.reader != nil {
		sm.reader.Close()
	}
}

func (sm *MbusReader) GetType() fatima.FatimaComponentType {
	return fatima.COMP_GENERAL
}

func (sm *MbusReader) consume(data []byte) {
	var message domain.MBusMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		log.Warn("fail to unmarshal : %s", err.Error())
		return
	}
	executor, ok := sm.executors[message.Header.ApplicationCode]
	if !ok {
		log.Warn("not found executor for application code %d", message.Header.ApplicationCode)
		return
	}

	executor.Consume(message)
}
