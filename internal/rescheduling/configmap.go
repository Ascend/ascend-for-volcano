/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (dealCM *DealReSchedulerConfigmap) setCMData(value map[string]string) {
	dealCM.CMData = value
}

func (dealCM *DealReSchedulerConfigmap) setCMName(value string) {
	dealCM.CMName = value
}

func (dealCM *DealReSchedulerConfigmap) setCMNameSpace(value string) {
	dealCM.CMNameSpace = value
}

func (dealCM *DealReSchedulerConfigmap) newReSchedulerCMFromEnv(env *plugin.ScheduleEnv, jobType string) error {
	reCmData, getErr := util.GetConfigMapWithRetry(env.FrameAttr.KubeClient, CmNameSpace, CmName)
	if getErr != nil {
		if !errors.IsNotFound(getErr) {
			klog.V(util.LogErrorLev).Infof("newReSchedulerCMFromEnv :%#v.", getErr)
			return getErr
		}
		klog.V(util.LogDebugLev).Infof("configmap %s not in env", RePropertyName)
		cmData := make(map[string]string, util.MapInitNum)
		cmData[CmFaultNodeKind] = ""
		cmData[jobType] = ""
		cmData[CmNodeHeartbeatKind] = ""
		cmData[CmNodeRankTimeMapKind] = ""

		var faultCM = &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      CmName,
				Namespace: CmNameSpace,
			},
			Data: cmData,
		}
		err := util.CreateOrUpdateConfigMap(env.FrameAttr.KubeClient, faultCM, CmName, CmNameSpace)
		if err != nil {
			return fmt.Errorf("create rescheduler configmap %s configmap failed: %#v", RePropertyName, err)
		}
		dealCM.setCMName(CmName)
		dealCM.setCMNameSpace(CmNameSpace)
		dealCM.setCMData(cmData)
		return fmt.Errorf("configmap %s in %s has been created", CmName, CmNameSpace)
	}

	dealCM.setCMName(reCmData.Name)
	dealCM.setCMNameSpace(reCmData.Namespace)
	dealCM.setCMData(reCmData.Data)
	return nil
}
