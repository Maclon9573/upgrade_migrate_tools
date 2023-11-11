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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"
)

const defaultTimeOut = time.Second * 60

type IdentifierResp struct {
	ID         string `json:"id"`
	Provider   uint32 `json:"provider"`
	CreatorId  uint32 `json:"creator_id"`
	Identifier string `json:"identifier"`
	CreatedAt  string `json:"created_at"`
}

type ClusterCredentialResp struct {
	ClusterID  string `json:"cluster_id"`
	ServerPath string `json:"server_address_path"`
	UserToken  string `json:"user_token"`
	CaCert     string `json:"cacert_data"`
}

func GetClusterIdentifier(host, token, projectID, clusterID string) (string, error) {
	resp := &IdentifierResp{}
	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		Get(fmt.Sprintf("%s/rest/clusters/bcs/query_by_id?project_id=%s&cluster_id=%s", host, projectID, clusterID)).
		SetDebug(true).
		Set("Authorization", fmt.Sprintf("Bearer %s", token)).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call BCS API failed: %v", errs[0])
		return "", errs[0]
	}

	if result.StatusCode != 200 {
		errMsg := fmt.Errorf("call BCS API error: code[%d], %s",
			result.StatusCode, string(body))
		return "", errMsg
	}

	return resp.Identifier, nil
}

func GetClusterCredential(host, token, id string) (*ClusterCredentialResp, error) {
	resp := &ClusterCredentialResp{}
	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		Get(fmt.Sprintf("%s/rest/clusters/%s/client_credentials", host, id)).
		Set("Authorization", fmt.Sprintf("Bearer %s", token)).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call BCS API failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != 200 {
		errMsg := fmt.Errorf("call BCS API error: code[%d], %s",
			result.StatusCode, string(body))
		return nil, errMsg
	}

	return resp, nil
}
