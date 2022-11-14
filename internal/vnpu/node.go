/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package vnpu is using for HuaWei Ascend pin fault rescheduling.

*/
package vnpu

// IsNodeStable todo: function to be completed
func (vNode *VNode) IsNodeStable() bool {
	return true
}

// ExcludeResFromVNode todo: function to be completed
func (vNode *VNode) ExcludeResFromVNode(resource VResource) {

}

// AddResToVNode todo: function to be completed
func (vNode *VNode) AddResToVNode(resource VResource) {

}

// IsChipStable todo: function to be completed
func (vChip *VChip) IsChipStable() bool {
	return true
}

// ExcludeResFromVChip todo: function to be completed
func (vChip *VChip) ExcludeResFromVChip(resource VResource) {

}

// AddResToVChip todo: function to be completed
func (vChip *VChip) AddResToVChip(resource VResource) {

}
