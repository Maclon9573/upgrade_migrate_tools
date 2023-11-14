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
	ClusterID               string                    `json:"clusterID,omitempty"`
	ClusterName             string                    `json:"clusterName,omitempty"`
	FederationClusterID     string                    `json:"federationClusterID,omitempty"`
	Provider                string                    `json:"provider,omitempty"`
	Region                  string                    `json:"region,omitempty"`
	VpcID                   string                    `json:"vpcID,omitempty"`
	ProjectID               string                    `json:"projectID,omitempty"`
	BusinessID              string                    `json:"businessID,omitempty"`
	Environment             string                    `json:"environment,omitempty"`
	EngineType              string                    `json:"engineType,omitempty"`
	IsExclusive             bool                      `json:"isExclusive,omitempty"`
	ClusterType             string                    `json:"clusterType,omitempty"`
	Labels                  map[string]string         `json:"labels,omitempty"`
	Creator                 string                    `json:"creator,omitempty"`
	CreateTime              string                    `json:"createTime,omitempty"`
	UpdateTime              string                    `json:"updateTime,omitempty"`
	BcsAddons               map[string]*BKOpsPlugin   `json:"bcsAddons,omitempty"`
	ExtraAddons             map[string]*BKOpsPlugin   `json:"extraAddons,omitempty"`
	SystemID                string                    `json:"systemID,omitempty"`
	ManageType              string                    `json:"manageType,omitempty"`
	Master                  map[string]*Node          `json:"master,omitempty"`
	NetworkSettings         *NetworkSetting           `json:"networkSettings,omitempty"`
	ClusterBasicSettings    *ClusterBasicSetting      `json:"clusterBasicSettings,omitempty"`
	ClusterAdvanceSettings  *ClusterAdvanceSetting    `json:"clusterAdvanceSettings,omitempty"`
	NodeSettings            *NodeSetting              `json:"nodeSettings,omitempty"`
	Status                  string                    `json:"status,omitempty"`
	Updater                 string                    `json:"updater,omitempty"`
	NetworkType             string                    `json:"networkType,omitempty"`
	AutoGenerateMasterNodes bool                      `json:"autoGenerateMasterNodes,omitempty"`
	Template                []*InstanceTemplateConfig `json:"template,omitempty"`
	ExtraInfo               map[string]string         `json:"extraInfo,omitempty"`
	ModuleID                string                    `json:"moduleID,omitempty"`
	ExtraClusterID          string                    `json:"extraClusterID,omitempty"`
	IsCommonCluster         bool                      `json:"isCommonCluster,omitempty"`
	Description             string                    `json:"description,omitempty"`
	ClusterCategory         string                    `json:"clusterCategory,omitempty"`
	IsShared                bool                      `json:"is_shared,omitempty"`
	KubeConfig              string                    `json:"kubeConfig,omitempty"`
	ImportCategory          string                    `json:"importCategory,omitempty"`
	CloudAccountID          string                    `json:"cloudAccountID,omitempty"`
	Area                    *CloudArea                `json:"area,omitempty"`
	Module                  *ModuleInfo               `json:"module,omitempty"`
	ClusterConnectSetting   *ClusterConnectSetting    `json:"clusterConnectSetting,omitempty"`
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

// BKOpsPlugin xxx
type BKOpsPlugin struct {
	System              string            `json:"system,omitempty"`
	Link                string            `json:"link,omitempty"`
	Params              map[string]string `json:"params,omitempty"`
	AllowSkipWhenFailed bool              `json:"allowSkipWhenFailed,omitempty"`
}

// Node xxx
type Node struct {
	NodeID         string `json:"nodeID,omitempty"`
	InnerIP        string `json:"innerIP,omitempty"`
	InstanceType   string `json:"instanceType,omitempty"`
	CPU            uint32 `json:"CPU,omitempty"`
	Mem            uint32 `json:"mem,omitempty"`
	GPU            uint32 `json:"GPU,omitempty"`
	Status         string `json:"status,omitempty"`
	ZoneID         string `json:"zoneID,omitempty"`
	NodeGroupID    string `json:"nodeGroupID,omitempty"`
	ClusterID      string `json:"clusterID,omitempty"`
	VPC            string `json:"VPC,omitempty"`
	Region         string `json:"region,omitempty"`
	Passwd         string `json:"passwd,omitempty"`
	Zone           uint32 `json:"zone,omitempty"`
	DeviceID       string `json:"deviceID,omitempty"`
	NodeTemplateID string `json:"nodeTemplateID,omitempty"`
	NodeType       string `json:"nodeType,omitempty"`
	NodeName       string `json:"nodeName,omitempty"`
	InnerIPv6      string `json:"innerIPv6,omitempty"`
	ZoneName       string `json:"zoneName,omitempty"`
	TaskID         string `json:"taskID,omitempty"`
}

// NetworkSetting xxx
type NetworkSetting struct {
	ClusterIPv4CIDR     string   `json:"clusterIPv4CIDR,omitempty"`
	ServiceIPv4CIDR     string   `json:"serviceIPv4CIDR,omitempty"`
	MaxNodePodNum       uint32   `json:"maxNodePodNum,omitempty"`
	MaxServiceNum       uint32   `json:"maxServiceNum,omitempty"`
	EnableVPCCni        bool     `json:"enableVPCCni,omitempty"`
	EniSubnetIDs        []string `json:"eniSubnetIDs,omitempty"`
	IsStaticIpMode      bool     `json:"isStaticIpMode,omitempty"`
	ClaimExpiredSeconds uint32   `json:"claimExpiredSeconds,omitempty"`
	MultiClusterCIDR    []string `json:"multiClusterCIDR,omitempty"`
	CidrStep            uint32   `json:"cidrStep,omitempty"`
	ClusterIpType       string   `json:"clusterIpType,omitempty"`
	ClusterIPv6CIDR     string   `json:"clusterIPv6CIDR,omitempty"`
	ServiceIPv6CIDR     string   `json:"serviceIPv6CIDR,omitempty"`
}

// ClusterBasicSetting xxx
type ClusterBasicSetting struct {
	OS                        string            `json:"OS,omitempty"`
	Version                   string            `json:"version,omitempty"`
	ClusterTags               map[string]string `json:"clusterTags,omitempty"`
	VersionName               string            `json:"versionName,omitempty"`
	SubnetID                  string            `json:"subnetID,omitempty"`
	ClusterLevel              string            `json:"clusterLevel,omitempty"`
	IsAutoUpgradeClusterLevel bool              `json:"isAutoUpgradeClusterLevel,omitempty"`
}

// ClusterAdvanceSetting xxx
type ClusterAdvanceSetting struct {
	IPVS               bool              `json:"IPVS,omitempty"`
	ContainerRuntime   string            `json:"containerRuntime,omitempty"`
	RuntimeVersion     string            `json:"runtimeVersion,omitempty"`
	ExtraArgs          map[string]string `json:"extraArgs,omitempty"`
	NetworkType        string            `json:"networkType,omitempty"`
	DeletionProtection bool              `json:"deletionProtection,omitempty"`
	AuditEnabled       bool              `json:"auditEnabled,omitempty"`
	EnableHa           bool              `json:"enableHa,omitempty"`
}

// NodeSetting xxx
type NodeSetting struct {
	DockerGraphPath   string            `json:"dockerGraphPath,omitempty"`
	MountTarget       string            `json:"mountTarget,omitempty"`
	UnSchedulable     uint32            `json:"unSchedulable,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	ExtraArgs         map[string]string `json:"extraArgs,omitempty"`
	InitLoginUsername string            `json:"initLoginUsername,omitempty"`
	InitLoginPassword string            `json:"initLoginPassword,omitempty"`
	KeyPair           *KeyInfo          `json:"keyPair,omitempty"`
	Taints            []*Taint          `json:"taints,omitempty"`
}

// KeyInfo xxx
type KeyInfo struct {
	KeyID     string `json:"keyID,omitempty"`
	KeySecret string `json:"keySecret,omitempty"`
	KeyPublic string `json:"keyPublic,omitempty"`
}

// Taint for node taints
type Taint struct {
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
	Effect string `json:"effect,omitempty"`
}
