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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/parnurzeal/gorequest"
)

// CreateProjectRequest create project request
type CreateProjectRequest struct {
	CreateTime  string `json:"createTime,omitempty"`
	Creator     string `json:"creator,omitempty"`
	ProjectID   string `json:"projectID,omitempty"`
	Name        string `json:"name,omitempty"`
	ProjectCode string `json:"projectCode,omitempty"`
	UseBKRes    bool   `json:"useBKRes,omitempty"`
	Description string `json:"description,omitempty"`
	IsOffline   bool   `json:"isOffline,omitempty"`
	Kind        string `json:"kind,omitempty"`
	BusinessID  string `json:"businessID,omitempty"`
	IsSecret    bool   `json:"isSecret,omitempty"`
	ProjectType uint32 `json:"projectType,omitempty"`
	DeployType  uint32 `json:"deployType,omitempty"`
	BGID        string `json:"BGID,omitempty"`
	BGName      string `json:"BGName,omitempty"`
	DeptID      string `json:"deptID,omitempty"`
	DeptName    string `json:"deptName,omitempty"`
	CenterID    string `json:"centerID,omitempty"`
	CenterName  string `json:"centerName,omitempty"`
}

// ProjectResponse create project response
type ProjectResponse struct {
	Code           uint32   `json:"code,omitempty"`
	Message        string   `json:"message,omitempty"`
	Data           *Project `json:"data,omitempty"`
	RequestID      string   `json:"requestID,omitempty"`
	WebAnnotations *Perms   `json:"web_annotations,omitempty"`
}

// Project for project info
type Project struct {
	CreateTime   string `json:"createTime,omitempty"`
	UpdateTime   string `json:"updateTime,omitempty"`
	Creator      string `json:"creator,omitempty"`
	Updater      string `json:"updater,omitempty"`
	Managers     string `json:"managers,omitempty"`
	ProjectID    string `json:"projectID,omitempty"`
	Name         string `json:"name,omitempty"`
	ProjectCode  string `json:"projectCode,omitempty"`
	UseBKRes     bool   `json:"useBKRes,omitempty"`
	Description  string `json:"description,omitempty"`
	IsOffline    bool   `json:"isOffline,omitempty"`
	Kind         string `json:"kind,omitempty"`
	BusinessID   string `json:"businessID,omitempty"`
	IsSecret     bool   `json:"isSecret,omitempty"`
	ProjectType  uint32 `json:"projectType,omitempty"`
	DeployType   uint32 `json:"deployType,omitempty"`
	BGID         string `json:"BGID,omitempty"`
	BGName       string `json:"BGName,omitempty"`
	DeptID       string `json:"deptID,omitempty"`
	DeptName     string `json:"deptName,omitempty"`
	CenterID     string `json:"centerID,omitempty"`
	CenterName   string `json:"centerName,omitempty"`
	BusinessName string `json:"businessName,omitempty"`
}

// Perms xxx
type Perms struct {
	Perms *_struct.Struct `json:"perms,omitempty"`
}

// CreateProject create project
func CreateProject(host, token string, debug bool, req *CreateProjectRequest) (*ProjectResponse, error) {
	resp := &ProjectResponse{}
	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		SetDebug(debug).
		TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		Post(fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects", host)).
		Set("Authorization", fmt.Sprintf("Bearer %s", token)).
		Send(req).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call bcs project manager api failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != 200 {
		errMsg := fmt.Errorf("call bcs project manager api error: code[%d], %s",
			result.StatusCode, string(body))
		return nil, errMsg
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf(resp.Message)
	}

	return resp, nil
}
