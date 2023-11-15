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
	"github.com/Tencent/bk-bcs/install/upgradetool/options"
	"github.com/Tencent/bk-bcs/install/upgradetool/types"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"
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
	Code      uint32        `json:"code"`
	Message   string        `json:"message"`
	Data      types.Cluster `json:"data"`
	RequestID string        `json:"request_id"`
	Result    bool          `json:"result"`
}

// GetAccessToken get access token
func GetAccessToken(cc options.BCSCc, debug bool) (*AccessTokenResponse, error) {
	resp := &AccessTokenResponse{}
	req := &AccessTokenReq{
		GrantType:  "authorization_code",
		IdProvider: "bk_login",
		BkToken:    cc.BkToken,
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

	if result.StatusCode != 200 {
		errMsg := fmt.Errorf("call bkssm api error: code[%d], %s",
			result.StatusCode, string(body))
		return nil, errMsg
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf(resp.Message)
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

	if result.StatusCode != 200 {
		errMsg := fmt.Errorf("call bcs cc api error: code[%d], %s",
			result.StatusCode, string(body))
		return nil, errMsg
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf(resp.Message)
	}

	return resp, nil
}
