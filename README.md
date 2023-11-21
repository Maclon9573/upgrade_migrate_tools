# BCS集群迁移工具

### 背景

BCS现采用容器化部署，与老版本之间版本跨度较大，无法平滑升级。

**注意：**

- 本方案只迁移BCS的数据，容器管理平台依赖的CMDB业务数据，新旧环境的业务ID需要一致。

迁移工具网络要求：

| 目标                                                         | 端口         |
| ------------------------------------------------------------ | ------------ |
| 旧环境bcs api的8443端口                                      | 8443         |
| 旧环境bcs cc所用的mysql                                      | 3306（默认） |
| 新环境bcs api（bcs-api.xxx.com）、bcs cc（bcs-cc.xxx.com）、bkssm（bkssm.xxx.com） | 80/443       |
| 新环境bcs使用的MongoDB                                       | 27017        |

### 环境信息

旧版本：v1.18

新版本：v1.28

### 方案设计

1. 迁移旧环境中mysql数据库的projects和clusters信息到新环境的MongoDB中
2. 在k8s集群中部署第二套bcs-kube-agent，上报证书、token信息到新环境中
3. 在各个集群中部署第二套bcs-k8s-watch，上报集群资源（开发中）
4. 删除旧版本的bcs-kube-agent、bcs-k8s-watch

### 迁移工具使用方法

编译

```
go build -o cluster-migrate-tool main.go
```

执行命令：

```
./cluster-migrate-tool -f conf.json
```

conf.json配置说明：

```
{
  "alsologtostderr": true,  // 打印日志到标准输出
  "log_dir": "./logs",    // 日志存放目录，需要提前创建
  "v": "3",   // 日志等级
  "debug": true,    // 是否开启http请求的debug模式
  "project_ids": [],    // 需要迁移项目id列表，如果为空，默认迁移所有项目
  "migrate_project_data": true,    // 是否迁移项目数据，如果项目数据已经使用本工具迁移完成，则设置为false
  "migrate_cluster_data": true,    // 是否迁移集群数据，如果集群数据已经使用本工具迁移完成，则设置为false
  "bcs_api": {   // 二进制版本的bcs api配置
    "addr": "https://192.168.xxx.xxx:8443",
    "token": ""    //  bcs api的认证token，推荐使用admin token（获取方式见下文）
  },
  "mysql_dsn": "user:password@tcp(192.168.xxx.xxx:3306)/bk_bcs_cc?charset=utf8mb4&parseTime=True&loc=Local",  // 二进制版本的bcs cc数据库dsn
  "bcs_api_gateway": {    // 容器化版本的bcs api gateway配置
    "addr": "http://bcs-api.xxx.com",   // 一般为bcs-api+域名
    "token": ""  //  bcs api gateway的认证token，推荐使用admin token（获取方式见下文）
  },
  "mongoDB": {    // 容器化版本的bcs MongoDB配置
    "host": "",
    "port": "",
    "username": "",
    "password": ""
  },
  "bcs_cc": {   // 容器化版本的bcs cc配置
    "addr": "http://bcs-cc.xxx.com",
    "ssm_host": "http://bkssm.xxx.com",
    "app_code": "bk_bcs",
    "app_secret": ""    // 容器管理平台的secret，可在开发者中心查看
  },
  "bk_cluster_id": "BCS-K8S-00000",   // 容器化版本的蓝鲸集群id
  "bcs_cert_name": "bcs-client-bcs-services-stack",   // 容器化版本的bcs cert所在secret名称（获取方式见下文）
  "kube_agent": {   // 容器化版本bcs kube agent配置
    "enable": true,   // 是否安装bcs kube agent
    "yaml_path": "/root/cluster-migrate-tool/kube-agent-deployment.yaml",   // bcs kube agent Deployment路径
    "namespace": "bcs-nodes",   // bcs kube agent命名空间，需要与老版本一致
    "image": ""
  }
}
```

说明：

- 可以根据需要更改kube agent的Deployment，如修改nodeAffinity、resources等

#### **二进制版本的bcs api的认证token**

方法一：从bcs api的数据库bke_core的user_tokens表中获取，value字段即为token信息

方法二：在bcs-cc的配置文件(一般为/data/bkee/etc/bcs/cc.yml)中，最后一行Bearer后面即为admin token

#### 容器化版本的bcs api gateway token

使用webconsole登录蓝鲸集群，执行

```
kubectl get secrets -n bcs-system bcs-password -oyaml |awk '/  gateway_token:/ {print $2}'| base64 -d
```

#### bcs_cert_name获取

使用webconsole登录蓝鲸集群，执行

```
kubectl get deployments.apps -n bcs-system bcs-kube-agent -oyaml
```

volume：bcs-certs使用的secret name即为bcs_cert_name。

      volumes:
	  - name: bcs-certs
	    projected:
	      defaultMode: 420
	      sources:
	      - secret:
	          items:
	          - key: ca.crt
	            path: bcs-ca.crt
	          - key: tls.crt
	            path: bcs-client.crt
	          - key: tls.key
	            path: bcs-client.key
	          name: bcs-client-bcs-services-stack # 该值即为bcs_cert_name