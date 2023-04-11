//
// Copyright (c) 2017 SK TECHX.
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
// @date 2017. 11. 3. PM 1:49
//

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
