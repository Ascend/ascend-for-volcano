/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package ascend310p is using for HuaWei Ascend pin affinity schedule.

*/
package ascend310p

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// TestNew
func TestNew(t *testing.T) {
	t.Run("test New", func(t *testing.T) {
		npu := New(PluginName)
		if npu.GetPluginName() != PluginName {
			t.Errorf("New() npu Name: %s, wantName: %s.", npu.GetPluginName(), PluginName)
		}
		if npu.GetAnnoName() != util.NPU310PCardName {
			t.Errorf("New() npu annoName: %s, wantAnnoName: %s.", npu.GetPluginName(), util.NPU310PCardName)
		}
		if npu.GetAnnoPreVal() != util.NPU310PCardNamePre {
			t.Errorf("New() npu annoNamePre: %s, wantAnnoNamePre: %s.",
				npu.GetPluginName(), util.NPU310PCardNamePre)
		}
	})
}
