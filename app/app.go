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

package app

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/install/upgradetool/components"
	"github.com/Tencent/bk-bcs/install/upgradetool/options"
	"github.com/Tencent/bk-bcs/install/upgradetool/types"
)

const (
	mongoDBNameProject           = "bcsproject_project"
	mongoDBNameCluster           = "clustermanager"
	mongoDBCollectionNameProject = "bcsproject_project"
	mongoDBCollectionNameCluster = "bcsclustermanagerv2_cluster"
)

type App struct {
	op          *options.UpgradeOption
	sqlClient   *gorm.DB
	mongoClient *mongo.Client
	client      *http.Client
}

func NewApp(op *options.UpgradeOption) *App {

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 60 * time.Second,
	}
	return &App{
		op:     op,
		client: client,
	}
}

func (app *App) DoMigrate() error {
	err := app.initMysqlClient()
	if err != nil {
		return err
	}
	defer func() {
		err = app.sqlClient.Close()
		if err != nil {
			blog.Errorf("disconnect mysql failed, %v", err)
		}
	}()
	app.sqlClient.AutoMigrate(&types.Project{}, &types.Cluster{})

	err = app.initMongoClient()
	if err != nil {
		return err
	}
	defer func() {
		err = app.mongoClient.Disconnect(context.Background())
		if err != nil {
			blog.Errorf("disconnect mongoDB failed, %v", err)
		}
	}()

	err = app.migrateProjects()
	if err != nil {
		return err
	}

	// migrate clusters in normal status
	err = app.migrateClusters()
	if err != nil {
		return err
	}

	return nil
}

func (app *App) migrateProjects() error {
	projects := make([]types.Project, 0)
	if len(app.op.ProjectIDs) != 0 {
		app.sqlClient.Model(&types.Project{}).Where("project_id IN (?)", app.op.ProjectIDs).Find(&projects)
	} else {
		app.sqlClient.Model(&types.Project{}).Find(&projects)
	}

	projectMs := make([]types.ProjectM, 0)
	interfaces := make([]interface{}, 0)
	for _, p := range projects {
		dpt, _ := strconv.Atoi(p.DeployType)
		projectMs = append(projectMs, types.ProjectM{
			CreateTime:  p.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdateTime:  p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Creator:     p.Creator,
			Updater:     p.Updator,
			Managers:    p.Creator,
			ProjectID:   p.ProjectID,
			Name:        p.Name,
			ProjectCode: p.EnglishName,
			UseBKRes:    p.UseBK,
			Description: p.Description,
			IsOffline:   p.IsOfflined,
			Kind:        strconv.Itoa(int(p.Kind)),
			BusinessID:  strconv.Itoa(int(p.CCAppID)),
			IsSecret:    p.IsSecrecy,
			ProjectType: uint32(p.ProjectType),
			DeployType:  uint32(dpt),
			BGID:        strconv.Itoa(int(p.BGID)),
			BGName:      p.BGName,
			DeptID:      strconv.Itoa(int(p.DeptID)),
			DeptName:    p.DeptName,
			CenterID:    strconv.Itoa(int(p.CenterID)),
			CenterName:  p.CenterName,
		})
	}

	for _, v := range projectMs {
		interfaces = append(interfaces, v)
	}

	_, err := app.mongoClient.Database(mongoDBNameProject).Collection(mongoDBCollectionNameProject).
		InsertMany(context.Background(), interfaces)
	if err != nil {
		blog.Errorf("insert projects to mongoDB failed, %v", err)
		return err
	}

	blog.Infof("migrated %d projects", len(interfaces))

	return nil
}

func (app *App) migrateClusters() error {
	clusters := make([]types.Cluster, 0)
	if len(app.op.ProjectIDs) != 0 {
		app.sqlClient.Model(&types.Cluster{}).Where("project_id IN (?) AND status = ?", app.op.ProjectIDs, "normal").Find(&clusters)
	} else {
		app.sqlClient.Model(&types.Cluster{}).Where("status = ?", "normal").Find(&clusters)
	}

	clusterMs := make([]types.ClusterM, 0)
	interfaces := make([]interface{}, 0)
	for _, c := range clusters {
		clusterMs = append(clusterMs, types.ClusterM{
			CreateTime:  c.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdateTime:  c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			ClusterID:   c.ClusterID,
			ClusterName: c.Name,
			Provider:    "bluekingCloud",
			Region:      "default",
			ProjectID:   c.ProjectID,
			BusinessID:  app.getClusterBusinessID(c.ProjectID),
			Environment: c.Environment,
			EngineType:  c.Type,
			ClusterType: "single",
			Creator:     c.Creator,
			ManageType:  "INDEPENDENT_CLUSTER",
			Status:      "RUNNING",
			NetworkType: "overlay",
			Description: c.Description,
		})
	}

	for _, v := range clusterMs {
		interfaces = append(interfaces, v)
	}

	_, err := app.mongoClient.Database(mongoDBNameCluster).Collection(mongoDBCollectionNameCluster).
		InsertMany(context.Background(), interfaces)
	if err != nil {
		blog.Errorf("insert projects to mongoDB failed, %v", err)
		return err
	}

	blog.Infof("migrated %d clusters", len(interfaces))

	for _, v := range clusterMs {
		err = deployKubeAgent(app.op, v.ProjectID, v.ClusterID)
		if err != nil {
			blog.Errorf("deployKubeAgent for cluster %s failed, %v", v.ClusterID, err)
		}
	}

	return nil
}

//func deployKubeAgent(op options.UpgradeOption, projectID, clusterID string) error {
//	clientSet, err := generateClientSet(op, projectID, clusterID)
//	if err != nil {
//		return err
//	}
//	clientSet.AppsV1().Deployments().Create()
//	return nil
//}

func deployKubeAgent(op *options.UpgradeOption, projectID, clusterID string) error {
	host := op.BCSApi.Addr
	token := op.BCSApi.Token

	id, err := components.GetClusterIdentifier(host, token, projectID, clusterID)
	if err != nil {
		blog.Errorf("get cluster %s identifier failed, %v", clusterID, err)
		return err
	}

	resp, err := components.GetClusterCredential(host, token, id)
	if err != nil {
		blog.Errorf("get cluster %s credential failed, %v", clusterID, err)
		return err
	}

	config := &rest.Config{
		Host: fmt.Sprintf("%s/tunnels/clusters/%s", host, id),
		TLSClientConfig: rest.TLSClientConfig{
			CertData: []byte(base64.StdEncoding.EncodeToString([]byte(resp.CaCert))),
		},
		BearerToken: resp.UserToken,
	}

	chart, err := loader.Load(op.KubeAgent.HelmPackagePath)
	if err != nil {
		blog.Errorf("load helm package failed, %v", err)
		return err
	}

	imageInfo := strings.Split(op.KubeAgent.Image, ":")
	if len(imageInfo) != 2 {
		return fmt.Errorf("invalid bcs kube agent image")
	}

	releaseName := fmt.Sprintf("bcs-kube-agent-%s", imageInfo[1])
	namespace := op.KubeAgent.Namespace
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(GetHelmConfig(config.Host, config.BearerToken, "", namespace),
		namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		blog.Errorf("init helm config failed, %v", err)
		return err
	}

	iCli := action.NewInstall(actionConfig)
	iCli.Namespace = namespace
	iCli.ReleaseName = releaseName

	gAddr := strings.Split(op.BCSApiGateway.Addr, "//")
	if len(gAddr) != 2 {
		return fmt.Errorf("invalid bcs api gateway address")
	}

	values := make(map[string]interface{}, 0)
	values["image.registry"] = ""
	values["image.repository"] = imageInfo[0]
	values["image.tag"] = imageInfo[1]
	values["args.BK_BCS_API"] = fmt.Sprintf("wss://%s", gAddr[1])
	values["args.BK_BCS_clusterId"] = clusterID
	values["args.BK_BCS_insecureSkipVerify"] = "true"
	values["args.BK_BCS_kubeAgentWSTunnel"] = "true"
	values["args.BK_BCS_websocketPath"] = "/bcsapi/v4/clustermanager/v1/websocket/connect"
	values["args.BK_BCS_APIToken"] = op.BCSApiGateway.Token

	_, err = iCli.Run(chart, values)
	if err != nil {
		blog.Errorf("install helm package failed, %v", err)
		return err
	}

	blog.Infof("install release %s/%s success", namespace, releaseName)

	return nil
}

func GetHelmConfig(host, userToken, context, namespace string) *genericclioptions.ConfigFlags {
	cf := genericclioptions.NewConfigFlags(true)
	insecure := true
	cf.Namespace = &namespace
	cf.APIServer = &host
	cf.BearerToken = &userToken
	cf.Context = &context
	cf.Insecure = &insecure
	return cf
}

func (app *App) getClusterBusinessID(projectID string) string {
	var project types.Project
	app.sqlClient.Model(&types.Project{}).Where("project_id = (?)", projectID).Find(&project)

	return strconv.Itoa(int(project.CCAppID))
}

func (app *App) initMysqlClient() error {
	if app.op.DSN == "" {
		return fmt.Errorf("empty mysql dsn")
	}

	db, err := gorm.Open("mysql", app.op.DSN)
	if err != nil {
		blog.Errorf("connect to mysql failed, %v", err)
		return err
	}
	db.DB().SetConnMaxLifetime(60 * time.Second)
	db.DB().SetMaxIdleConns(20)
	db.DB().SetMaxOpenConns(20)

	app.sqlClient = db
	return nil
}

func (app *App) initMongoClient() error {
	if app.op.MongoDB.Username == "" || app.op.MongoDB.Password == "" ||
		app.op.MongoDB.Host == "" || app.op.MongoDB.Port == 0 {
		return fmt.Errorf("lost mongo info in config")
	}

	clientOptions := mongooptions.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%d",
		app.op.MongoDB.Username,
		app.op.MongoDB.Password,
		app.op.MongoDB.Host,
		app.op.MongoDB.Port,
	))

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		blog.Errorf("connect to mongoDB failed, %v", err)
		return err
	}

	app.mongoClient = client

	return nil
}
