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

package utility

import "fmt"

func GetIntFromMap(m map[string]interface{}, key string) (int, error) {
	f, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("not found map key : %s", key)
	}

	if v, ok := f.(float64); ok {
		return int(v), nil
	}
	return 0, fmt.Errorf("not numeric value in key %s", key)
}
