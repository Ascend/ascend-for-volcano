/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package plugin is using for HuaWei Ascend pin affinity schedule frame.
*/

package plugin

const (
	// VolcanoNamespace the namespace of tor info configmap
	VolcanoNamespace = "volcano-system"
	// TorNodeCMName the Name of tor info configmap
	TorNodeCMName = "basic-tor-node-cm"
	// TorInfoCMKey the key of tor info in configmap
	TorInfoCMKey = "tor_info"
	// ServerListCMKey the key of serverList
	ServerListCMKey = "serverList"
	// ServerListCMPre the prefix of serverList configMap
	ServerListCMPre = "server-config-"
	// TorListVersion the version of tor info configmap
	TorListVersion = "1.0"
	// SingleJobTorId the tor id of single job
	SingleJobTorId = 0
	// ServerIPKey the key of ServerIP
	ServerIPKey = "server_ip"
	// TorAffinityKey the key of tor affinity
	TorAffinityKey = "tor-affinity"
	// LargeModelTag the value of large model
	LargeModelTag = "large-model-schema"
	// NullTag the value means not use tor affinity
	NullTag = "null"
	// Completed the status of serverList configmap
	Completed = "completed"
)
