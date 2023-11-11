/*
 * Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
 * Copyright (C) 2017-2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	"time"
)

// Model : Base model definition
type Model struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
	Extra     string     `json:"extra" sql:"type:text"`
}

// Project Model project info in 1.18
type Project struct {
	Model
	Name           string    `json:"name" gorm:"size:64;unique"`
	EnglishName    string    `json:"english_name" gorm:"size:64;unique;index"`
	Creator        string    `json:"creator" gorm:"size:32"`
	Updator        string    `json:"updator" gorm:"size:32"`
	Description    string    `json:"desc" sql:"type:text"`
	ProjectType    uint      `json:"project_type"`
	IsOfflined     bool      `json:"is_offlined" gorm:"default:false"`
	ProjectID      string    `json:"project_id" gorm:"size:32;unique;index"`
	UseBK          bool      `json:"use_bk" gorm:"default:false"`
	CCAppID        uint      `json:"cc_app_id"`
	Kind           uint      `json:"kind"`        // 1:k8s, 2:mesos
	DeployType     string    `json:"deploy_type"` // 1: 物理机部署, 2: 容器部署
	BGID           uint      `json:"bg_id"`
	BGName         string    `json:"bg_name"`
	DeptID         uint      `json:"dept_id"`
	DeptName       string    `json:"dept_name"`
	CenterID       uint      `json:"center_id"`
	CenterName     string    `json:"center_name"`
	DataID         uint      `json:"data_id"`
	IsSecrecy      bool      `json:"is_secrecy" gorm:"default:false"`
	ApprovalStatus uint      `json:"approval_status" gorm:"default:2"` // 1.待审批 2.已审批 3.已驳回
	LogoAddr       string    `json:"logo_addr" sql:"type:text"`        // project logo address
	Approver       string    `json:"approver" gorm:"size:32"`
	Remark         string    `json:"remark" sql:"type:text"`
	ApprovalTime   time.Time `json:"approval_time"`
}

// ProjectM for project in MongoDB
type ProjectM struct {
	CreateTime  string `json:"createTime" bson:"createTime"`
	UpdateTime  string `json:"updateTime" bson:"updateTime"`
	Creator     string `json:"creator" bson:"creator"`
	Updater     string `json:"updater" bson:"updater"`
	Managers    string `json:"managers" bson:"managers"`
	ProjectID   string `json:"projectID" bson:"projectID"`
	Name        string `json:"name" bson:"name"`
	ProjectCode string `json:"projectCode" bson:"projectCode"`
	UseBKRes    bool   `json:"useBKRes" bson:"useBKRes"`
	Description string `json:"description" bson:"description"`
	IsOffline   bool   `json:"isOffline" bson:"isOffline"`
	Kind        string `json:"kind" bson:"kind"`
	BusinessID  string `json:"businessID" bson:"businessID"`
	IsSecret    bool   `json:"isSecret" bson:"isSecret"`
	ProjectType uint32 `json:"projectType" bson:"projectType"`
	DeployType  uint32 `json:"deployType" bson:"deployType"`
	BGID        string `json:"bgID" bson:"bgID"`
	BGName      string `json:"bgName" bson:"bgName"`
	DeptID      string `json:"deptID" bson:"deptID"`
	DeptName    string `json:"deptName" bson:"deptName"`
	CenterID    string `json:"centerID" bson:"centerID"`
	CenterName  string `json:"centerName" bson:"centerName"`
}

// Cluster Model : cluster info in 1.18
type Cluster struct {
	Model
	Name              string     `json:"name" gorm:"size:64;unique_index:uix_project_id_name"`
	Creator           string     `json:"creator" gorm:"size:32"`
	Description       string     `json:"description" sql:"size:128"`
	ProjectID         string     `json:"project_id" gorm:"size:32;index;unique_index:uix_project_id_name"`
	RelatedProjects   string     `json:"related_projects" sql:"type:text"`
	ClusterID         string     `json:"cluster_id" gorm:"size:64;unique_index"`
	ClusterNum        int64      `json:"cluster_num" gorm:"unique"`
	Status            string     `json:"status" gorm:"size:64"`     // 初始化的状态 uninitialized, initializing, initialized, initialize_failed
	Disabled          bool       `json:"disabled"`                  // 是否禁用掉了，default false
	Type              string     `json:"type" gorm:"size:8"`        // mesos,k8s
	Environment       string     `json:"environment" gorm:"size:8"` // stag,debug,prod
	AreaID            int        `json:"area_id"`
	ConfigSvrCount    int        `json:"config_svr_count"`
	MasterCount       int        `json:"master_count"`
	NodeCount         int        `json:"node_count"`
	IPResourceTotal   int        `json:"ip_resource_total"`
	IPResourceUsed    int        `json:"ip_resource_used"`
	Artifactory       string     `json:"artifactory" gorm:"size:256"`
	TotalMem          float64    `json:"total_mem"`
	RemainMem         float64    `json:"remain_mem"`
	TotalCPU          float64    `json:"total_cpu"`
	RemainCPU         float64    `json:"remain_cpu"`
	TotalDisk         float64    `json:"total_disk"`
	RemainDisk        float64    `json:"remain_disk"`
	CapacityUpdatedAt *time.Time `json:"capacity_updated_at"`
	NotNeedNAT        bool       `json:"not_need_nat" gorm:"default:false"`
	ExtraClusterID    string     `json:"extra_cluster_id" gorm:"size:64"`
	State             string     `json:"state" gorm:"size:16;default:'bcs_new'"`
}

// ClusterM cluster info in MongoDB
type ClusterM struct {
	ClusterID               string                    `json:"clusterID,omitempty" bson:"clusterID"`
	ClusterName             string                    `json:"clusterName,omitempty" bson:"clusterName"`
	FederationClusterID     string                    `json:"federationClusterID,omitempty" bson:"federationClusterID"`
	Provider                string                    `json:"provider,omitempty" bson:"provider"`
	Region                  string                    `json:"region,omitempty" bson:"region"`
	VpcID                   string                    `json:"vpcID,omitempty" bson:"vpcID"`
	ProjectID               string                    `json:"projectID,omitempty" bson:"projectID"`
	BusinessID              string                    `json:"businessID,omitempty" bson:"businessID"`
	Environment             string                    `json:"environment,omitempty" bson:"environment"`
	EngineType              string                    `json:"engineType,omitempty" bson:"engineType"`
	IsExclusive             bool                      `json:"isExclusive,omitempty" bson:"isExclusive"`
	ClusterType             string                    `json:"clusterType,omitempty" bson:"clusterType"`
	Labels                  map[string]string         `json:"labels,omitempty" bson:"labels"`
	Creator                 string                    `json:"creator,omitempty" bson:"creator"`
	CreateTime              string                    `json:"createTime,omitempty" bson:"createTime"`
	UpdateTime              string                    `json:"updateTime,omitempty" bson:"updateTime"`
	BcsAddons               map[string]*BKOpsPlugin   `json:"bcsAddons,omitempty" bson:"bcsAddons"`
	ExtraAddons             map[string]*BKOpsPlugin   `json:"extraAddons,omitempty" bson:"extraAddons"`
	SystemID                string                    `json:"systemID,omitempty" bson:"systemID"`
	ManageType              string                    `json:"manageType,omitempty" bson:"manageType"`
	Master                  map[string]*Node          `json:"master,omitempty" bson:"master"`
	NetworkSettings         *NetworkSetting           `json:"networkSettings,omitempty" bson:"networkSettings"`
	ClusterBasicSettings    *ClusterBasicSetting      `json:"clusterBasicSettings,omitempty" bson:"clusterBasicSettings"`
	ClusterAdvanceSettings  *ClusterAdvanceSetting    `json:"clusterAdvanceSettings,omitempty" bson:"clusterAdvanceSettings"`
	NodeSettings            *NodeSetting              `json:"nodeSettings,omitempty" bson:"nodeSettings"`
	Status                  string                    `json:"status,omitempty" bson:"status"`
	Updater                 string                    `json:"updater,omitempty" bson:"updater"`
	NetworkType             string                    `json:"networkType,omitempty" bson:"networkType"`
	AutoGenerateMasterNodes bool                      `json:"autoGenerateMasterNodes,omitempty" bson:"autoGenerateMasterNodes"`
	Template                []*InstanceTemplateConfig `json:"template,omitempty" bson:"template"`
	ExtraInfo               map[string]string         `json:"extraInfo,omitempty" bson:"extraInfo"`
	ModuleID                string                    `json:"moduleID,omitempty" bson:"moduleID"`
	ExtraClusterID          string                    `json:"extraClusterID,omitempty" bson:"extraClusterID"`
	IsCommonCluster         bool                      `json:"isCommonCluster,omitempty" bson:"isCommonCluster"`
	Description             string                    `json:"description,omitempty" bson:"description"`
	ClusterCategory         string                    `json:"clusterCategory,omitempty" bson:"clusterCategory"`
	IsShared                bool                      `json:"is_shared,omitempty" bson:"isShared"`
	KubeConfig              string                    `json:"kubeConfig,omitempty" bson:"kubeConfig"`
	ImportCategory          string                    `json:"importCategory,omitempty" bson:"importCategory"`
	CloudAccountID          string                    `json:"cloudAccountID,omitempty" bson:"cloudAccountID"`
	Area                    *CloudArea                `json:"area,omitempty" bson:"area"`
	Module                  *ModuleInfo               `json:"module,omitempty" bson:"module"`
	ClusterConnectSetting   *ClusterConnectSetting    `json:"clusterConnectSetting,omitempty" bson:"clusterConnectSetting"`
}

// ClusterConnectSetting ApiServer内外网访问信息
type ClusterConnectSetting struct {
	IsExtranet    bool                `json:"isExtranet,omitempty"`
	SubnetId      string              `json:"subnetId,omitempty"`
	Domain        string              `json:"domain,omitempty"`
	SecurityGroup string              `json:"securityGroup,omitempty"`
	Internet      *InternetAccessible `json:"internet,omitempty"`
}

// InternetAccessible 公网带宽设置
type InternetAccessible struct {
	InternetChargeType   string `json:"internetChargeType,omitempty"`
	InternetMaxBandwidth string `json:"internetMaxBandwidth,omitempty"`
	PublicIPAssigned     bool   `json:"publicIPAssigned,omitempty"`
	BandwidthPackageId   string `json:"bandwidthPackageId,omitempty"`
}

// ModuleInfo 业务模块信息,主要涉及到节点模块转移
type ModuleInfo struct {
	ScaleOutModuleID   string `json:"scaleOutModuleID,omitempty"`
	ScaleInModuleID    string `json:"scaleInModuleID,omitempty"`
	ScaleOutBizID      string `json:"scaleOutBizID,omitempty"`
	ScaleInBizID       string `json:"scaleInBizID,omitempty"`
	ScaleOutModuleName string `json:"scaleOutModuleName,omitempty"`
	ScaleInModuleName  string `json:"scaleInModuleName,omitempty"`
}

type CloudArea struct {
	BkCloudID   uint32 `json:"bkCloudID,omitempty"`
	BkCloudName string `json:"bkCloudName,omitempty"`
}

type InstanceTemplateConfig struct {
	Region             string           `json:"region,omitempty"`
	Zone               string           `json:"zone,omitempty"`
	VpcID              string           `json:"vpcID,omitempty"`
	SubnetID           string           `json:"subnetID,omitempty"`
	ApplyNum           uint32           `json:"applyNum,omitempty"`
	CPU                uint32           `json:"CPU,omitempty"`
	Mem                uint32           `json:"Mem,omitempty"`
	GPU                uint32           `json:"GPU,omitempty"`
	InstanceType       string           `json:"instanceType,omitempty"`
	InstanceChargeType string           `json:"instanceChargeType,omitempty"`
	SystemDisk         *DataDisk        `json:"systemDisk,omitempty"`
	DataDisks          []*DataDisk      `json:"dataDisks,omitempty"`
	ImageInfo          *ImageInfo       `json:"imageInfo,omitempty"`
	InitLoginPassword  string           `json:"initLoginPassword,omitempty"`
	SecurityGroupIDs   []string         `json:"securityGroupIDs,omitempty"`
	IsSecurityService  bool             `json:"isSecurityService,omitempty"`
	IsMonitorService   bool             `json:"isMonitorService,omitempty"`
	CloudDataDisks     []*CloudDataDisk `json:"cloudDataDisks,omitempty"`
	InitLoginUsername  string           `json:"initLoginUsername,omitempty"`
	KeyPair            *KeyInfo         `json:"keyPair,omitempty"`
	DockerGraphPath    string           `json:"dockerGraphPath,omitempty"`
	NodeRole           string           `json:"nodeRole,omitempty"`
}

// DataDisk 数据盘定义
type DataDisk struct {
	DiskType string `json:"diskType,omitempty"`
	DiskSize string `json:"diskSize,omitempty"`
}

// ImageInfo 创建cvm实例的镜像信息
type ImageInfo struct {
	ImageID   string `json:"imageID,omitempty"`
	ImageName string `json:"imageName,omitempty"`
	ImageType string `json:"imageType,omitempty"`
}

// CloudDataDisk 云磁盘格式化数据, 对应CVM数据盘。应用于节点模版
// 主要用于 CA 自动扩容节点并上架节点时重装系统，多块数据盘 mountTarget 不能重复
// 上架已存在节点时, 用户需指定diskPartition参数区分设备
type CloudDataDisk struct {
	DiskType           string `json:"diskType,omitempty"`
	DiskSize           string `json:"diskSize,omitempty"`
	FileSystem         string `json:"fileSystem,omitempty"`
	AutoFormatAndMount bool   `json:"autoFormatAndMount,omitempty"`
	MountTarget        string `json:"mountTarget,omitempty"`
	DiskPartition      string `json:"diskPartition,omitempty"`
}

type BKOpsPlugin struct {
	System              string            `json:"system,omitempty" bson:"system"`
	Link                string            `json:"link,omitempty" bson:"link"`
	Params              map[string]string `json:"params,omitempty" bson:"params"`
	AllowSkipWhenFailed bool              `json:"allowSkipWhenFailed,omitempty" bson:"allowSkipWhenFailed"`
}

type Node struct {
	NodeID         string `json:"nodeID,omitempty" bson:"nodeID"`
	InnerIP        string `json:"innerIP,omitempty" bson:"innerIP"`
	InstanceType   string `json:"instanceType,omitempty" bson:"instanceType"`
	CPU            uint32 `json:"CPU,omitempty" bson:"CPU"`
	Mem            uint32 `json:"mem,omitempty" bson:"mem"`
	GPU            uint32 `json:"GPU,omitempty" bson:"GPU"`
	Status         string `json:"status,omitempty" bson:"status"`
	ZoneID         string `json:"zoneID,omitempty" bson:"zoneID"`
	NodeGroupID    string `json:"nodeGroupID,omitempty" bson:"nodeGroupID"`
	ClusterID      string `json:"clusterID,omitempty" bson:"clusterID"`
	VPC            string `json:"VPC,omitempty" bson:"VPC"`
	Region         string `json:"region,omitempty" bson:"region"`
	Passwd         string `json:"passwd,omitempty" bson:"passwd"`
	Zone           uint32 `json:"zone,omitempty" bson:"zone"`
	DeviceID       string `json:"deviceID,omitempty" bson:"deviceID"`
	NodeTemplateID string `json:"nodeTemplateID,omitempty" bson:"nodeTemplateID"`
	NodeType       string `json:"nodeType,omitempty" bson:"nodeType"`
	NodeName       string `json:"nodeName,omitempty" bson:"nodeName"`
	InnerIPv6      string `json:"innerIPv6,omitempty" bson:"innerIPv6"`
	ZoneName       string `json:"zoneName,omitempty" bson:"zoneName"`
	TaskID         string `json:"taskID,omitempty" bson:"taskID"`
}

type NetworkSetting struct {
	ClusterIPv4CIDR     string   `json:"clusterIPv4CIDR,omitempty" bson:"clusterIPv4CIDR"`
	ServiceIPv4CIDR     string   `json:"serviceIPv4CIDR,omitempty" bson:"serviceIPv4CIDR"`
	MaxNodePodNum       uint32   `json:"maxNodePodNum,omitempty" bson:"maxNodePodNum"`
	MaxServiceNum       uint32   `json:"maxServiceNum,omitempty" bson:"maxServiceNum"`
	EnableVPCCni        bool     `json:"enableVPCCni,omitempty" bson:"enableVPCCni"`
	EniSubnetIDs        []string `json:"eniSubnetIDs,omitempty" bson:"eniSubnetIDs"`
	IsStaticIpMode      bool     `json:"isStaticIpMode,omitempty" bson:"isStaticIpMode"`
	ClaimExpiredSeconds uint32   `json:"claimExpiredSeconds,omitempty" bson:"claimExpiredSeconds"`
	MultiClusterCIDR    []string `json:"multiClusterCIDR,omitempty" bson:"multiClusterCIDR"`
	CidrStep            uint32   `json:"cidrStep,omitempty" bson:"cidrStep"`
	ClusterIpType       string   `json:"clusterIpType,omitempty" bson:"clusterIpType"`
	ClusterIPv6CIDR     string   `json:"clusterIPv6CIDR,omitempty" bson:"clusterIPv6CIDR"`
	ServiceIPv6CIDR     string   `json:"serviceIPv6CIDR,omitempty" bson:"serviceIPv6CIDR"`
}

type ClusterBasicSetting struct {
	OS                        string            `json:"OS,omitempty" bson:"OS"`
	Version                   string            `json:"version,omitempty" bson:"version"`
	ClusterTags               map[string]string `json:"clusterTags,omitempty" bson:"clusterTags"`
	VersionName               string            `json:"versionName,omitempty" bson:"versionName"`
	SubnetID                  string            `json:"subnetID,omitempty" bson:"subnetID"`
	ClusterLevel              string            `json:"clusterLevel,omitempty" bson:"clusterLevel"`
	IsAutoUpgradeClusterLevel bool              `json:"isAutoUpgradeClusterLevel,omitempty" bson:"isAutoUpgradeClusterLevel"`
}

type ClusterAdvanceSetting struct {
	IPVS               bool              `json:"IPVS,omitempty" bson:"IPVS"`
	ContainerRuntime   string            `json:"containerRuntime,omitempty" bson:"containerRuntime"`
	RuntimeVersion     string            `json:"runtimeVersion,omitempty" bson:"runtimeVersion"`
	ExtraArgs          map[string]string `json:"extraArgs,omitempty" bson:"extraArgs"`
	NetworkType        string            `json:"networkType,omitempty" bson:"networkType"`
	DeletionProtection bool              `json:"deletionProtection,omitempty" bson:"deletionProtection"`
	AuditEnabled       bool              `json:"auditEnabled,omitempty" bson:"auditEnabled"`
	EnableHa           bool              `json:"enableHa,omitempty" bson:"enableHa"`
}

type NodeSetting struct {
	DockerGraphPath   string            `json:"dockerGraphPath,omitempty" bson:"dockerGraphPath"`
	MountTarget       string            `json:"mountTarget,omitempty" bson:"mountTarget"`
	UnSchedulable     uint32            `json:"unSchedulable,omitempty" bson:"unSchedulable"`
	Labels            map[string]string `json:"labels,omitempty" bson:"labels"`
	ExtraArgs         map[string]string `json:"extraArgs,omitempty" bson:"extraArgs"`
	InitLoginUsername string            `json:"initLoginUsername,omitempty" bson:"initLoginUsername"`
	InitLoginPassword string            `json:"initLoginPassword,omitempty" bson:"initLoginPassword"`
	KeyPair           *KeyInfo          `json:"keyPair,omitempty" bson:"keyPair"`
	Taints            []*Taint          `json:"taints,omitempty" bson:"taints"`
}

type KeyInfo struct {
	KeyID     string `json:"keyID,omitempty" bson:"keyID"`
	KeySecret string `json:"keySecret,omitempty" bson:"keySecret"`
	KeyPublic string `json:"keyPublic,omitempty" bson:"keyPublic"`
}

// Taint for node taints
type Taint struct {
	Key    string `json:"key,omitempty" bson:"key"`
	Value  string `json:"value,omitempty" bson:"value"`
	Effect string `json:"effect,omitempty" bson:"effect"`
}
