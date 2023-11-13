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
	"go.mongodb.org/mongo-driver/bson"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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

	kubeAgentSecretName = "bcs-client-bcs-kube-agent"
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

	if app.op.MigrateClusterData {
		err = app.migrateProjects()
		if err != nil {
			return err
		}
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

	blog.Debug("got %d projects from database", len(projects))

	projectMs := make([]types.ProjectM, 0)
	for _, p := range projects {
		dpt, _ := strconv.Atoi(p.DeployType)
		_, err := components.CreateProject(app.op.BCSApiGateway.Addr, app.op.BCSApiGateway.Token,
			&components.CreateProjectRequest{
				Creator:     p.Creator,
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
		if err != nil {
			blog.Errorf("create project %s[%s] failed, %v", p.Name, p.ProjectID, err)
			continue
		}
		blog.Infof("create project %s[%s] success", p.Name, p.ProjectID)
	}

	blog.Infof("migrated %d projects", len(projectMs))

	return nil
}

func (app *App) migrateClusters() error {
	clusterCol := app.mongoClient.Database(mongoDBNameCluster).Collection(mongoDBCollectionNameCluster)
	clusters := make([]types.Cluster, 0)
	if len(app.op.ProjectIDs) != 0 {
		app.sqlClient.Model(&types.Cluster{}).Where("project_id IN (?) AND status = ?", app.op.ProjectIDs, "normal").Find(&clusters)
	} else {
		app.sqlClient.Model(&types.Cluster{}).Where("status = ?", "normal").Find(&clusters)
	}

	dupClusters := make([]types.ClusterM, 0)
	for _, c := range clusters {
		clusterM := types.ClusterM{
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
		}

		if app.op.MigrateClusterData {
			_, err := clusterCol.InsertOne(context.Background(), clusterM)
			if err != nil {
				if strings.Contains(err.Error(), "duplicate key") {
					existCluster := types.ClusterM{}
					errFind := clusterCol.FindOne(
						context.Background(), bson.M{"clusterid": clusterM.ClusterID}).Decode(&existCluster)
					if errFind != nil {
						return errFind
					}
					if existCluster.ClusterName == clusterM.ClusterName && existCluster.BusinessID == clusterM.BusinessID {
						blog.Infof("cluster %s[%s] exists, skipping...")
						continue
					}
					dupClusters = append(dupClusters, clusterM)
					continue
				}
				return err
			}
		}

		err := deployKubeAgent(app.op, clusterM, clusterM.ProjectID, clusterM.ClusterID)
		if err != nil {
			blog.Errorf("deployKubeAgent for cluster %s failed, %v", clusterM.ClusterID, err)
		}
	}

	clusterIDMap := make(map[string]string, 0)
	for _, c := range dupClusters {
		clusterNum, err := app.generateClusterID()
		if err != nil {
			return err
		}
		newClusterID := fmt.Sprintf("BCS-K8S-%d", clusterNum)
		if app.op.MigrateClusterData {
			clusterIDMap[newClusterID] = c.ClusterID
			c.ClusterID = newClusterID
			_, err = clusterCol.InsertOne(context.Background(), c)
			if err != nil {
				return err
			}
		}

		err = deployKubeAgent(app.op, c, c.ProjectID, clusterIDMap[newClusterID])
		if err != nil {
			blog.Errorf("deployKubeAgent for cluster %s[%]s failed, %v",
				c.ClusterName, clusterIDMap[newClusterID], err)
		}
	}

	return nil
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

func deployKubeAgent(op *options.UpgradeOption, cluster types.ClusterM, projectID, clusterID string) error {
	blog.Infof("deploying new kube agent for %s[%s]", cluster.ClusterName, cluster.ClusterID)

	host := op.BCSApi.Addr
	token := op.BCSApi.Token

	id, err := components.GetClusterIdentifier(host, token, projectID, clusterID)
	if err != nil {
		blog.Errorf("get cluster %s identifier failed, %v", clusterID, err)
		return err
	}

	resp, err := components.GetClusterCredential(host, token, id.ID)
	if err != nil {
		blog.Errorf("get cluster %s credential failed, %v", clusterID, err)
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

	// create clientset from bcs-api v1.18
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

	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeAgentSecretName,
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
	imageInfo := strings.Split(op.KubeAgent.Image, ":")
	if len(imageInfo) != 2 {
		return fmt.Errorf("invalid bcs kube agent image")
	}
	name := fmt.Sprintf("bcs-kube-agent-%s", imageInfo[1])

	gAddr := strings.Split(op.BCSApiGateway.Addr, "//")
	if len(gAddr) != 2 {
		return fmt.Errorf("invalid bcs api gateway address")
	}
	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.DeploymentSpec{
			Replicas: func(a int32) *int32 { return &a }(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "bcs-kube-agent",
							Image: op.KubeAgent.Image,
							Args: []string{
								fmt.Sprintf("--bke-address=wss://%s", gAddr[1]),
								fmt.Sprintf("--cluster-id=%s", clusterID),
								fmt.Sprintf("--insecureSkipVerify=%s", "true"),
								fmt.Sprintf("--verbosity=%d", 3),
								fmt.Sprintf("--use-websocket=%s", "true"),
								fmt.Sprintf("--websocket-path=%s", "/bcsapi/v4/clustermanager/v1/websocket/connect"),
							},
							Env: []corev1.EnvVar{
								{
									Name:  "USER_TOKEN",
									Value: op.BCSApiGateway.Token,
								},
								{
									Name:  "CLIENT_CA",
									Value: "/data/bcs/cert/bcs/bcs-ca.crt",
								},
								{
									Name:  "CLIENT_CERT",
									Value: "/data/bcs/cert/bcs/bcs-client.crt",
								},
								{
									Name:  "CLIENT_KEY",
									Value: "/data/bcs/cert/bcs/bcs-client.key",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "bcs-certs",
									MountPath: "/data/bcs/cert/bcs",
								},
							},
						},
					},
					DeprecatedServiceAccount: "bcs-kube-agent",
					ServiceAccountName:       "bcs-kube-agent",
					Volumes: []corev1.Volume{
						{
							Name: "bcs-certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: kubeAgentSecretName,
									Items: []corev1.KeyToPath{
										{
											Key:  "ca.crt",
											Path: "bcs-ca.crt",
										},
										{
											Key:  "tls.crt",
											Path: "bcs-client.crt",
										},
										{
											Key:  "tls.key",
											Path: "bcs-client.key",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := clientset.AppsV1().Deployments(op.KubeAgent.Namespace).
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
