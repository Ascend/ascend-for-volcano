/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package ascend310p is using for HuaWei 310P Ascend pin affinity schedule.
*/
package ascend310p

import (
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

func (tp *ascend310P) InitVNPU() {
}

func (tp *ascend310P) checkDyVJobReq() error {
	return nil
}

func (tp *ascend310P) validDyVNPUJob() *api.ValidateResult {
	return nil
}

func (tp *ascend310P) preStartVNPU(ssn *framework.Session) error {
	return nil
}
