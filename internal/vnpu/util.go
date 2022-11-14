/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package vnpu is using for HuaWei Ascend pin fault rescheduling.

*/
package vnpu

import (
	"reflect"
	"strings"

	"volcano.sh/volcano/pkg/scheduler/api"
)

func Add(vReses []VResource) VResource {
	vResTotal := VResource{
		Aicore: 0,
		Aicpu:  0,
		Vpc:    0,
		Vdec:   0,
		Jpegd:  0,
		Pngd:   0,
		Venc:   0,
		Jpege:  0,
	}
	for _, vRes := range vReses {
		vResTotal.Aicore += vRes.Aicore
		vResTotal.Aicpu += vRes.Aicpu
		vResTotal.Vpc += vRes.Vdec
		vResTotal.Jpege += vRes.Jpege
		vResTotal.Pngd += vRes.Pngd
		vResTotal.Venc += vRes.Venc
		vResTotal.Jpege += vRes.Jpege
	}
	return vResTotal
}

func Sub(vRes1 VResource, vRes2 VResource) VResource {
	return VResource{
		Aicore: vRes1.Aicore - vRes2.Aicore,
		Aicpu:  vRes1.Aicpu - vRes2.Aicpu,
		Vpc:    vRes1.Vpc - vRes2.Vpc,
		Vdec:   vRes1.Vdec - vRes2.Vdec,
		Jpegd:  vRes1.Jpegd - vRes2.Jpegd,
		Pngd:   vRes1.Pngd - vRes2.Pngd,
		Venc:   vRes1.Venc - vRes2.Venc,
		Jpege:  vRes1.Jpege - vRes2.Jpege,
	}
}

func (vResource VResource) BeGreater(vRes VResource) bool {
	vResSub := Sub(vResource, vRes)
	v := reflect.ValueOf(vResSub)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Int() < 0 {
			return false
		}
	}
	return true
}

func GetPodUsedNPUNames(task *api.TaskInfo, resType string) []string {
	if task == nil {
		return nil
	}
	res, ok := task.Pod.Annotations[resType]
	if !ok {
		return nil
	}
	resSlice := strings.Split(res, ",")
	return resSlice
}
