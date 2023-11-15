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

package options

import "github.com/Tencent/bk-bcs/bcs-common/common/conf"

// UpgradeOption upgrade option
type UpgradeOption struct {
	conf.FileConfig
	conf.LogConfig

	Debug              bool        `json:"debug"`
	ProjectIDs         []string    `json:"project_ids"`
	MigrateProjectData bool        `json:"migrate_project_data"`
	MigrateClusterData bool        `json:"migrate_cluster_data"`
	DSN                string      `json:"mysql_dsn"`
	MongoDB            MongoDBConf `json:"mongoDB"`

	BCSApi        BCSConf   `json:"bcs_api"`
	BCSApiGateway BCSConf   `json:"bcs_api_gateway"`
	BCSCertName   string    `json:"bcs_cert_name"`
	BKClusterID   string    `json:"bk_cluster_id"`
	BCSCc         BCSCc     `json:"bcs_cc"`
	KubeAgent     KubeAgent `json:"kube_agent"`
	K8SWatch      K8SWatch  `json:"k8s_watch"`
}

// BCSCc bcs cc
type BCSCc struct {
	Addr      string `json:"addr"`
	SsmHost   string `json:"ssm_host"`
	BkToken   string `json:"bk_token"`
	AppCode   string `json:"app_code"`
	AppSecret string `json:"app_secret"`
}

// BCSConf bcs configuration
type BCSConf struct {
	Addr  string `json:"addr"`
	IP    string `json:"ip"`
	Token string `json:"token"`
}

// MongoDBConf MongoDB configuration
type MongoDBConf struct {
	Host     string `json:"host"`
	Port     uint32 `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// KubeAgent bcs kube agent configuration
type KubeAgent struct {
	Enable          bool   `json:"enable"`
	YamlPath        string `json:"yaml_path"`
	HelmPackagePath string `json:"helm_package_path"`
	Namespace       string `json:"namespace"`
	ServiceAccount  string `json:"service_account"`
	Image           string `json:"image"`
}

// K8SWatch bcs k8s watch configuration
type K8SWatch struct {
}
