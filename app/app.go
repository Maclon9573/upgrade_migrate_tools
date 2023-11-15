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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"k8s.io/client-go/kubernetes/scheme"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/install/upgradetool/components"
	"github.com/Tencent/bk-bcs/install/upgradetool/options"
	"github.com/Tencent/bk-bcs/install/upgradetool/types"
)

const (
	mongoDBNameCluster           = "clustermanager"
	mongoDBCollectionNameCluster = "bcsclustermanagerv2_cluster"
)

// App for app
type App struct {
	op          *options.UpgradeOption
	sqlClient   *gorm.DB
	mongoClient *mongo.Client
}

// NewApp create App
func NewApp(op *options.UpgradeOption) *App {
	return &App{
		op: op,
	}
}

// DoMigrate migrate data and deploy bcs components
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

	if app.op.MigrateProjectData {
		err = app.migrateProjects()
		if err != nil {
			return err
		}
	}

	// migrate clusters in normal status
	successClusters, changedClusters, err := app.migrateClusters()
	if err != nil {
		return err
	}

	blog.Debug("got %d changed clusters: %v", len(changedClusters), changedClusters)

	// deploy bcs kube agent
	if app.op.KubeAgent.Enable {
		blog.Infof("deploy new bcs kube agent enabled")

		blog.Infof("will deploy new bcs kube agent on %d clusters", len(successClusters))
		for _, c := range successClusters {
			err := deployKubeAgent(app.op, c, changedClusters)
			if err != nil {
				blog.Errorf("deploy kube agent for cluster %s failed, %v", c.ClusterID, err)
			}
		}
	}

	return nil
}

func (app *App) migrateProjects() error {
	projects := make([]types.Project, 0)
	successProjects, failedProjects := make(map[string]string, 0), make(map[string]string, 0)
	if len(app.op.ProjectIDs) != 0 {
		app.sqlClient.Model(&types.Project{}).Where("project_id IN (?)", app.op.ProjectIDs).Find(&projects)
	} else {
		app.sqlClient.Model(&types.Project{}).Find(&projects)
	}

	blog.Infof("got %d projects from database", len(projects))

	for _, p := range projects {
		dpt, _ := strconv.Atoi(p.DeployType)
		_, err := components.CreateProject(app.op.BCSApiGateway.Addr, app.op.BCSApiGateway.Token, app.op.Debug,
			&components.CreateProjectRequest{
				Creator:     p.Creator,
				ProjectID:   p.ProjectID,
				Name:        p.Name,
				ProjectCode: p.EnglishName,
				UseBKRes:    p.UseBK,
				Description: p.Description,
				IsOffline:   p.IsOfflined,
				Kind:        "k8s",
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
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				blog.Infof(err.Error())
				successProjects[p.ProjectID] = p.Name
				continue
			}
			blog.Errorf("create project %s[%s] failed, %v", p.Name, p.ProjectID, err)
			failedProjects[p.ProjectID] = p.Name
			continue
		}
		successProjects[p.ProjectID] = p.Name
		blog.Infof("create project %s[%s] success", p.Name, p.ProjectID)
	}

	blog.Infof("migrated %d projects", len(successProjects))
	blog.Infof("%d projects failed: %v", len(failedProjects), failedProjects)

	return nil
}

func (app *App) migrateClusters() ([]types.ClusterM, map[string]string, error) {
	clusterCol := app.mongoClient.Database(mongoDBNameCluster).Collection(mongoDBCollectionNameCluster)
	clusters := make([]types.Cluster, 0)
	successClusters := make([]types.ClusterM, 0)
	failedClusters := make([]types.ClusterM, 0)
	changedClusters := make(map[string]string, 0)

	if len(app.op.ProjectIDs) != 0 {
		app.sqlClient.Model(&types.Cluster{}).Where("project_id IN (?) AND status = ?", app.op.ProjectIDs, "normal").Find(&clusters)
	} else {
		app.sqlClient.Model(&types.Cluster{}).Where("status = ?", "normal").Find(&clusters)
	}

	blog.Infof("got %d clusters from database", len(clusters))

	dupClusters := make([]types.ClusterM, 0)
	for _, c := range clusters {
		now := time.Now()
		clusterM := types.ClusterM{
			CreateTime:  now.Format("2006-01-02T15:04:05Z"),
			UpdateTime:  now.Format("2006-01-02T15:04:05Z"),
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
		}

		existClustersMongo := make([]types.ClusterM, 0)
		cursor, err := clusterCol.Find(context.Background(), bson.M{})
		if err != nil {
			return successClusters, nil, err
		}

		if err = cursor.All(context.Background(), &existClustersMongo); err != nil {
			return successClusters, nil, err
		}
		cursor.Close(context.Background())

		exist := false
		// 重复执行可能存在clusterID变更而且已经迁移的情况
		for _, cm := range existClustersMongo {
			if cm.ProjectID == c.ProjectID && cm.ClusterName == c.Name &&
				cm.Description == c.Description && c.ClusterID != cm.ClusterID {
				changedClusters[cm.ClusterID] = c.ClusterID
				blog.Infof("clusterID of cluster[%s] changed from %s to %s", c.Name, c.ClusterID, cm.ClusterID)
				clusterM.ClusterID = cm.ClusterID
				successClusters = append(successClusters, clusterM)
				exist = true
				break
			}
		}

		if app.op.MigrateClusterData {
			if exist {
				blog.Infof("cluster %s[%s] imported already, skipping...",
					clusterM.ClusterName, clusterM.ClusterID)
				continue
			}
			_, err = clusterCol.InsertOne(context.Background(), clusterM)
			if err != nil {
				if strings.Contains(err.Error(), "duplicate key") {
					dupClusters = append(dupClusters, clusterM)
					continue
				}

				failedClusters = append(failedClusters, clusterM)
				blog.Errorf("migrate cluster %s[%s] failed, %v", clusterM.ClusterID, clusterM.ClusterName, err)
				continue
			}

			err = createClusterInCc(app.op, clusterM)
			if err != nil {
				failedClusters = append(failedClusters, clusterM)
				continue
			}
			successClusters = append(successClusters, clusterM)
		}
	}

	successClusters, failedClusters = app.processDupClusters(dupClusters, successClusters, failedClusters, changedClusters)
	blog.Infof("migrated %d clusters", len(successClusters))
	blog.Infof("%d clusters failed: %v", len(failedClusters), failedClusters)

	return successClusters, changedClusters, nil
}

func createClusterInCc(op *options.UpgradeOption, cluster types.ClusterM) error {
	masters, err := getMasterNodes(op, cluster)
	if err != nil {
		blog.Errorf("get master nodes for cluster %s[%s] failed, %v", cluster.ClusterName, cluster.ClusterID, err)
		return err
	}

	resp, err := components.GetAccessToken(op.BCSCc, op.Debug)
	if err != nil {
		blog.Errorf("get access token failed")
		return err
	}

	masterIPs := make([]components.CreateMasterData, 0)
	for _, m := range masters {
		for _, ip := range m.Status.Addresses {
			if ip.Type == corev1.NodeInternalIP {
				masterIPs = append(masterIPs, components.CreateMasterData{InnerIP: ip.Address})
				break
			}
		}
	}

	clusterNum, _ := strconv.Atoi(strings.TrimPrefix(cluster.ClusterID, "BCS-K8S-"))
	_, err = components.SyncClusterToCc(op.BCSCc.Addr, cluster.ProjectID, resp.Data.AccessToken, op.Debug,
		&components.SyncClusterReq{
			ProjectID:   cluster.ProjectID,
			ClusterID:   cluster.ClusterID,
			ClusterNum:  clusterNum,
			Name:        cluster.ClusterName,
			Creator:     cluster.Creator,
			Description: cluster.Description,
			Type:        "k8s",
			Environment: "prod",
			AreaID:      1,
			Status:      cluster.Status,
			MasterIPs:   masterIPs,
		})
	if err != nil {
		blog.Errorf("sync cluster %s[%s] to bcs cc failed, %v", cluster.ClusterName, cluster.ClusterID, err)
		return err
	}

	return nil
}

func (app *App) processDupClusters(dupClusters, success, failed []types.ClusterM, changedClusters map[string]string) (
	[]types.ClusterM, []types.ClusterM) {
	clusterCol := app.mongoClient.Database(mongoDBNameCluster).Collection(mongoDBCollectionNameCluster)
	for _, c := range dupClusters {
		clusterNum, err := app.generateClusterID()
		if err != nil {
			blog.Errorf("processDupClusters generateClusterID failed, %v", err)
			failed = append(failed, c)
			continue
		}
		newClusterID := fmt.Sprintf("BCS-K8S-%d", clusterNum)
		blog.Infof("clusterID of cluster[%s] changed from %s to %s", c.ClusterName, c.ClusterID, newClusterID)
		changedClusters[newClusterID] = c.ClusterID
		c.ClusterID = newClusterID

		if app.op.MigrateClusterData {
			_, err = clusterCol.InsertOne(context.Background(), c)
			if err != nil {
				blog.Errorf("processDupClusters %s[%s] failed, %v", c.ClusterName, c.ClusterID, err)
				failed = append(failed, c)
				continue
			}
			err = createClusterInCc(app.op, c)
			if err != nil {
				failed = append(failed, c)
				continue
			}
			success = append(success, c)
		}
	}
	return success, failed
}

func (app *App) generateClusterID() (int, error) {
	clusterCol := app.mongoClient.Database(mongoDBNameCluster).Collection(mongoDBCollectionNameCluster)
	cursor, err := clusterCol.Find(context.Background(), bson.M{})
	if err != nil {
		return 0, err
	}
	defer cursor.Close(context.Background())

	var clusters []types.ClusterM
	if err = cursor.All(context.Background(), &clusters); err != nil {
		return 0, err
	}

	clusterNumIDs := make([]int, 0)
	for i := range clusters {
		clusterStrs := strings.Split(clusters[i].ClusterID, "-")
		if len(clusterStrs) != 3 {
			continue
		}

		id, _ := strconv.Atoi(clusterStrs[2])
		clusterNumIDs = append(clusterNumIDs, id)
	}
	sort.Ints(clusterNumIDs)

	if len(clusterNumIDs) == 0 {
		return 1, nil
	}

	return clusterNumIDs[len(clusterNumIDs)-1] + 1, nil
}

func getMasterNodes(op *options.UpgradeOption, cluster types.ClusterM) ([]*corev1.Node, error) {
	blog.Infof("deploying new kube agent for %s[%s]", cluster.ClusterName, cluster.ClusterID)

	host := op.BCSApi.Addr
	token := op.BCSApi.Token
	id, err := components.GetClusterIdentifier(host, token, cluster.ProjectID, cluster.ClusterID, op.Debug)
	if err != nil {
		blog.Errorf("get cluster %s identifier failed, %v", cluster.ClusterID, err)
		return nil, err
	}

	resp, err := components.GetClusterCredential(host, token, id.ID, op.Debug)
	if err != nil {
		blog.Errorf("get cluster %s credential failed, %v", cluster.ClusterID, err)
		return nil, err
	}

	config := &rest.Config{
		Host: fmt.Sprintf("%s/tunnels/clusters/%s", host, id.Identifier),
		TLSClientConfig: rest.TLSClientConfig{
			CertData: []byte(base64.StdEncoding.EncodeToString([]byte(resp.CaCert))),
			Insecure: true,
		},
		BearerToken: resp.UserToken,
	}

	// create clientset from bcs-api
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	masters := make([]*corev1.Node, 0)
	nodeList, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, n := range nodeList.Items {
		if _, ok := n.Labels["node-role.kubernetes.io/master"]; ok {
			masters = append(masters, &n)
		} else if _, ok := n.Labels["node-role.kubernetes.io/control-plane"]; ok {
			masters = append(masters, &n)
		}
	}

	return masters, nil
}

func deployKubeAgent(op *options.UpgradeOption, cluster types.ClusterM, changeClusters map[string]string) error {
	blog.Infof("deploying new kube agent for %s[%s]", cluster.ClusterName, cluster.ClusterID)

	host := op.BCSApi.Addr
	token := op.BCSApi.Token
	orgClusterID := cluster.ClusterID
	if value, ok := changeClusters[cluster.ClusterID]; ok {
		orgClusterID = value
	}

	id, err := components.GetClusterIdentifier(host, token, cluster.ProjectID, orgClusterID, op.Debug)
	if err != nil {
		blog.Errorf("get cluster %s identifier failed, %v", orgClusterID, err)
		return err
	}

	resp, err := components.GetClusterCredential(host, token, id.ID, op.Debug)
	if err != nil {
		blog.Errorf("get cluster %s credential failed, %v", orgClusterID, err)
		return err
	}

	config := &rest.Config{
		Host: fmt.Sprintf("%s/tunnels/clusters/%s", host, id.Identifier),
		TLSClientConfig: rest.TLSClientConfig{
			CertData: []byte(base64.StdEncoding.EncodeToString([]byte(resp.CaCert))),
			Insecure: true,
		},
		BearerToken: resp.UserToken,
	}

	// create clientset from bcs-api
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	err = createKubeAgentSecret(op, clientset)
	if err != nil {
		return err
	}

	err = createKubeAgent(op, clientset, cluster.ClusterID)
	if err != nil {
		return err
	}

	blog.Infof("deploy new kube agent for %s[%s] success", cluster.ClusterName, cluster.ClusterID)

	return nil
}

func createKubeAgentSecret(op *options.UpgradeOption, clientset *kubernetes.Clientset) error {
	// construct k8s client config by bcs api gateway in new version
	restConfig := &rest.Config{
		Host:        op.BCSApiGateway.Addr + "/clusters/" + op.BKClusterID,
		BearerToken: op.BCSApiGateway.Token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
		QPS:   100,
		Burst: 100,
	}
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	// get secret from blueking cluster
	secret, err := client.CoreV1().Secrets("bcs-system").
		Get(context.Background(), op.BCSCertName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// create secret by old version bcs api
	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      op.BCSCertName,
			Namespace: op.KubeAgent.Namespace,
		},
		Immutable:  secret.Immutable,
		Data:       secret.Data,
		StringData: secret.StringData,
		Type:       secret.Type,
	}
	_, err = clientset.CoreV1().Secrets(op.KubeAgent.Namespace).
		Create(context.Background(), newSecret, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func createKubeAgent(op *options.UpgradeOption, clientset *kubernetes.Clientset, clusterID string) error {
	gAddr := strings.Split(op.BCSApiGateway.Addr, "//")
	if len(gAddr) != 2 {
		return fmt.Errorf("invalid bcs api gateway address")
	}
	hostAliaas := []corev1.HostAlias{}
	if op.BCSApiGateway.IP != "" {
		hostAliaas = append(hostAliaas, corev1.HostAlias{
			IP:        op.BCSApiGateway.IP,
			Hostnames: []string{gAddr[1]},
		})
	}

	// 从文件中读取 YAML 内容
	yamlFile, err := ioutil.ReadFile(op.KubeAgent.YamlPath)
	if err != nil {
		return err
	}

	// 将 YAML 转换为 Deployment 对象
	deployment := &v1.Deployment{}
	_, _, err = scheme.Codecs.UniversalDeserializer().Decode(yamlFile, nil, deployment)
	if err != nil {
		return err
	}
	deployment.Namespace = op.KubeAgent.Namespace
	deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args,
		fmt.Sprintf("--bke-address=wss://%s", gAddr[1]),
		fmt.Sprintf("--cluster-id=%s", clusterID))
	deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env,
		corev1.EnvVar{
			Name:  "USER_TOKEN",
			Value: op.BCSApiGateway.Token,
		})
	deployment.Spec.Template.Spec.Containers[0].Image = op.KubeAgent.Image
	deployment.Spec.Template.Spec.ServiceAccountName = op.KubeAgent.ServiceAccount
	deployment.Spec.Template.Spec.DeprecatedServiceAccount = op.KubeAgent.ServiceAccount
	deployment.Spec.Template.Spec.HostAliases = hostAliaas
	for _, v := range deployment.Spec.Template.Spec.Volumes {
		if v.Name == "bcs-certs" {
			v.Secret.SecretName = op.BCSCertName
		}
	}

	_, err = clientset.AppsV1().Deployments(op.KubeAgent.Namespace).
		Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

//func deployKubeAgentByHelm(op *options.UpgradeOption, projectID, clusterID string) error {
//	host := op.BCSApi.Addr
//	token := op.BCSApi.Token
//
//	id, err := components.GetClusterIdentifier(host, token, projectID, clusterID)
//	if err != nil {
//		blog.Errorf("get cluster %s identifier failed, %v", clusterID, err)
//		return err
//	}
//
//	resp, err := components.GetClusterCredential(host, token, id.ID)
//	if err != nil {
//		blog.Errorf("get cluster %s credential failed, %v", clusterID, err)
//		return err
//	}
//
//	config := &rest.Config{
//		Host: fmt.Sprintf("%s/tunnels/clusters/%s", host, id.Identifier),
//		TLSClientConfig: rest.TLSClientConfig{
//			CertData: []byte(base64.StdEncoding.EncodeToString([]byte(resp.CaCert))),
//		},
//		BearerToken: resp.UserToken,
//	}
//
//	chart, err := loader.Load(op.KubeAgent.HelmPackagePath)
//	if err != nil {
//		blog.Errorf("load helm package failed, %v", err)
//		return err
//	}
//
//	imageInfo := strings.Split(op.KubeAgent.Image, ":")
//	if len(imageInfo) != 2 {
//		return fmt.Errorf("invalid bcs kube agent image")
//	}
//
//	releaseName := fmt.Sprintf("bcs-kube-agent-%s", imageInfo[1])
//	namespace := op.KubeAgent.Namespace
//	actionConfig := new(action.Configuration)
//	if err := actionConfig.Init(GetHelmConfig(config.Host, config.BearerToken, "", namespace),
//		namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
//		blog.Errorf("init helm config failed, %v", err)
//		return err
//	}
//
//	iCli := action.NewInstall(actionConfig)
//	iCli.Namespace = namespace
//	iCli.ReleaseName = releaseName
//
//	gAddr := strings.Split(op.BCSApiGateway.Addr, "//")
//	if len(gAddr) != 2 {
//		return fmt.Errorf("invalid bcs api gateway address")
//	}
//
//	values := make(map[string]interface{}, 0)
//	values["image.registry"] = ""
//	values["image.repository"] = imageInfo[0]
//	values["image.tag"] = imageInfo[1]
//	values["args.BK_BCS_API"] = fmt.Sprintf("wss://%s", gAddr[1])
//	values["args.BK_BCS_clusterId"] = clusterID
//	values["args.BK_BCS_insecureSkipVerify"] = "true"
//	values["args.BK_BCS_kubeAgentWSTunnel"] = "true"
//	values["args.BK_BCS_websocketPath"] = "/bcsapi/v4/clustermanager/v1/websocket/connect"
//	values["args.BK_BCS_APIToken"] = op.BCSApiGateway.Token
//
//	_, err = iCli.Run(chart, values)
//	if err != nil {
//		blog.Errorf("install helm package failed, %v", err)
//		return err
//	}
//
//	blog.Infof("install release %s/%s success", namespace, releaseName)
//
//	return nil
//}

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
	blog.Infof("initializing mysql database")

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

	blog.Infof("init mysql database done")

	return nil
}

func (app *App) initMongoClient() error {
	blog.Infof("initializing mongoDB database")

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

	blog.Infof("init mongoDB database done")

	return nil
}
