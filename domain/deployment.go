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
 * @project fatima-go
 * @author dave_01
 * @date 23. 10. 4. 오후 5:20
 */

package domain

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

type Deployment struct {
	Valid       bool            `json:"-"`
	Process     string          `json:"process"`
	ProcessType string          `json:"process_type,omitempty"`
	Build       DeploymentBuild `json:"build,omitempty"`
}

func (d Deployment) HasBuildInfo() bool {
	if len(d.Build.BuildTime) == 0 {
		return false
	}
	return true
}

type DeploymentBuild struct {
	Git       DeploymentBuildGit `json:"git,omitempty"`
	BuildTime string             `json:"time,omitempty"`
	BuildUser string             `json:"user,omitempty"`
}

func (d DeploymentBuild) HasGit() bool {
	if len(d.Git.Branch) == 0 {
		return false
	}
	return true
}

type DeploymentBuildGit struct {
	Branch  string `json:"branch"`
	Commit  string `json:"commit"`
	Message string `json:"message,omitempty"`
}

func (d DeploymentBuildGit) String() string {
	return fmt.Sprintf("Branch=[%s], Commit=[%s]", d.Branch, d.Commit)
}

func GetTrimmedMessage(msg string) string {
	buff := bytes.Buffer{}
	lineCount := 1
	scanner := bufio.NewScanner(strings.NewReader(msg))
	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		buff.WriteString("\n")
		if lineCount > 3 { // 3줄까지만 출력해 주자..
			buff.WriteString(".....")
			break
		}
		buff.WriteString(line)
		lineCount++
	}

	return buff.String()
}
