/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package components

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/install/upgradetool/options"
	"github.com/Tencent/bk-bcs/install/upgradetool/types"
)

// AccessTokenReq access token request
type AccessTokenReq struct {
	GrantType  string `json:"grant_type"`
	IdProvider string `json:"id_provider"`
	BkToken    string `json:"bk_token"`
}

// AccessTokenResponse access token response
type AccessTokenResponse struct {
	Code    uint32          `json:"code"`
	Message string          `json:"message"`
	Data    AccessTokenData `json:"data"`
}

// AccessTokenData access token data
type AccessTokenData struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Expired      uint32   `json:"expired_in"`
	Identity     Identity `json:"identity"`
}

// Identity user info
type Identity struct {
	Username string `json:"username"`
	UserType string `json:"user_type"`
}

// SyncClusterReq sync cluster
type SyncClusterReq struct {
	ProjectID       string             `json:"project_id"`
	ClusterID       string             `json:"cluster_id"`
	ClusterNum      int                `json:"cluster_num"`
	Name            string             `json:"name"`
	Creator         string             `json:"creator"`
	Description     string             `json:"description"`
	Type            string             `json:"type"`
	Environment     string             `json:"environment"`
	AreaID          int                `json:"area_id"`
	Status          string             `json:"status"`
	ConfigSvrCount  int                `json:"config_svr_count"`
	MasterCount     int                `json:"master_count"`
	Artifactory     string             `json:"artifactory"`
	IPResourceTotal int                `json:"ip_resource_total"`
	IPResourceUsed  int                `json:"ip_resource_used"`
	MasterIPs       []CreateMasterData `json:"master_ips"`
	NeedNAT         bool               `json:"need_nat"`
	NotNeedNAT      bool               `json:"not_need_nat"`
	ExtraClusterID  string             `json:"extra_cluster_id"`
	State           string             `json:"state"`
}

// CreateMasterData create master data
type CreateMasterData struct {
	InnerIP      string `json:"inner_ip"`
	ExtendedInfo string `json:"extended_info"`
	Backup       string `json:"backup"`
	Hostname     string `json:"hostname"`
	Status       string `json:"status"`
}

// SyncClusterResponse sync cluster resp
type SyncClusterResponse struct {
	CommonResp
	Data   types.Cluster `json:"data"`
	Result bool          `json:"result"`
}

// CommonResp common resp
type CommonResp struct {
	Code      uint   `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// ClusterParamsRequest xxx
type ClusterParamsRequest struct {
	ClusterID          string           `json:"cluster_id"`
	ClusterName        string           `json:"name"`
	ClusterDescription string           `json:"description"`
	AreaID             int              `json:"area_id"`
	VpcID              string           `json:"vpc_id"`
	Env                string           `json:"environment"`
	MasterIPs          []ManagerMasters `json:"master_ips"`
	NeedNAT            bool             `json:"need_nat"`
	Version            string           `json:"version"`
	NetworkType        string           `json:"network_type"`
	Coes               string           `json:"coes"`
	KubeProxyMode      string           `json:"kube_proxy_mode"`
	Creator            string           `json:"creator"`
	Type               string           `json:"type"`
	ExtraClusterID     string           `json:"extra_cluster_id"`
	State              string           `json:"state"`
	Status             string           `json:"status"`
}

// ManagerMasters masterIP
type ManagerMasters struct {
	InnerIP string `json:"inner_ip"`
}

// ClusterSnapShootInfo snapInfo
type ClusterSnapShootInfo struct {
	Regions                 string              `json:"regions"`
	ClusterID               string              `json:"cluster_id"`
	MasterIPList            []string            `json:"master_ip_list"`
	VpcID                   string              `json:"vpc_id"`
	SystemDataID            uint32              `json:"bcs_system_data_id"`
	ClusterCIDRSettings     ClusterCIDRInfo     `json:"ClusterCIDRSettings"`
	ClusterType             string              `json:"ClusterType"`
	ClusterBasicSettings    ClusterBasicInfo    `json:"ClusterBasicSettings"`
	ClusterAdvancedSettings ClusterAdvancedInfo `json:"ClusterAdvancedSettings"`
	NetWorkType             string              `json:"network_type"`
	EsbURL                  string              `json:"esb_url"`
	WebhookImage            string              `json:"bcs_webhook_image"`
	PrivilegeImage          string              `json:"gcs_privilege_image"`
	VersionName             string              `json:"version_name"`
	Version                 string              `json:"version"`
	ClusterVersion          string              `json:"ClusterVersion"`
	ControlIP               string              `json:"control_ip"`
	MasterIPs               []string            `json:"master_ips"`
	Env                     string              `json:"environment"`
	ProjectName             string              `json:"product_name"`
	ProjectCode             string              `json:"project_code"`
	AreaName                string              `json:"area_name"`
	ExtraClusterID          string              `json:"extra_cluster_id"`
}

// ClusterCIDRInfo cidrInfo
type ClusterCIDRInfo struct {
	ClusterCIDR          string `json:"ClusterCIDR"`
	MaxNodePodNum        uint32 `json:"MaxNodePodNum"`
	MaxClusterServiceNum uint32 `json:"MaxClusterServiceNum"`
}

// ClusterBasicInfo basicInfo
type ClusterBasicInfo struct {
	ClusterOS      string `json:"ClusterOs"`
	ClusterVersion string `json:"ClusterVersion"`
	ClusterName    string `json:"ClusterName"`
}

// ClusterAdvancedInfo advancedInfo
type ClusterAdvancedInfo struct {
	IPVS bool `json:"IPVS"`
}

// UpdateNodeParams update node params
type UpdateNodeParams struct {
	NodeUpdateDataJSON
	ClusterID string `json:"cluster_id" binding:"required"`
}

// NodeListUpdateDataJSON node list update data
type NodeListUpdateDataJSON struct {
	Updates []UpdateNodeParams `json:"updates" binding:"required,dive"`
}

// NodeUpdateDataJSON node update data
type NodeUpdateDataJSON struct {
	InnerIP     string `binding:"required" json:"inner_ip"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	InstanceID  string `json:"instance_id"`
}

// NodeListUpdateResponse node list update resp
type NodeListUpdateResponse struct {
	CommonResp
	Data   string `json:"data"`
	Result bool   `json:"result"`
}

// GetAccessToken get access token
func GetAccessToken(cc options.BCSCc, debug bool) (*AccessTokenResponse, error) {
	resp := &AccessTokenResponse{}
	req := &AccessTokenReq{
		GrantType:  "client_credentials",
		IdProvider: "client",
	}
	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		SetDebug(debug).
		TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		Post(fmt.Sprintf("%s/api/v1/auth/access-tokens", cc.SsmHost)).
		Set("Content-Type", "application/json").
		Set("X-BK-APP-CODE", cc.AppCode).
		Set("X-BK-APP-SECRET", cc.AppSecret).
		Send(req).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call bkssm api failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("call bkssm api error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return nil, errMsg
	}

	return resp, nil
}

// SyncClusterToCc sync cluster to bcs cc
func SyncClusterToCc(host, projectID, token string, debug bool, req *SyncClusterReq) (*SyncClusterResponse, error) {
	resp := &SyncClusterResponse{}
	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		SetDebug(debug).
		TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		Post(fmt.Sprintf("%s/projects/%s/clusters?access_token=%s", host, projectID, token)).
		Set("Content-Type", "application/json").
		Send(req).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call bcs cc api failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("call bcs cc api error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return nil, errMsg
	}

	return resp, nil
}

// UpdateCluster update cluster in bcs cc
func UpdateCluster(host, projectID, clusterID, token string, debug bool, req *ClusterParamsRequest) error {
	resp := &CommonResp{}
	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		SetDebug(debug).
		TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		Put(fmt.Sprintf("%s/projects/%s/clusters/%s?access_token=%s", host, projectID, clusterID, token)).
		Set("Content-Type", "application/json").
		Send(req).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call bcs cc api failed: %v", errs[0])
		return errs[0]
	}

	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("call bcs cc api error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return errMsg
	}

	return nil
}

// UpdateNodeList update node list
func UpdateNodeList(host, projectID, token string, debug bool, req *NodeListUpdateDataJSON) error {
	resp := &NodeListUpdateResponse{}
	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		SetDebug(debug).
		TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		Put(fmt.Sprintf("%s/projects/%s?access_token=%s", host, projectID, token)).
		Set("Content-Type", "application/json").
		Send(req).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call bcs cc api failed: %v", errs[0])
		return errs[0]
	}

	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("call bcs cc api error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return errMsg
	}

	return nil
}
