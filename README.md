<div align="center">

[![StarsL.cn](https://img.shields.io/badge/website-StarsL.cn-orange)](https://starsl.cn)
[![Commits](https://img.shields.io/github/commit-activity/m/CassInfra/KubeDoor?color=ffff00)](https://github.com/CassInfra/KubeDoor/commits/main)
[![open issues](http://isitmaintained.com/badge/open/CassInfra/KubeDoor.svg)](https://github.com/CassInfra/KubeDoor/issues)
[![Python](https://img.shields.io/badge/python-v3.11-3776ab)](https://nodejs.org)
[![Node.js](https://img.shields.io/badge/node.js-v22-229954)](https://nodejs.org)
[![GitHub license](https://img.shields.io/badge/license-Apache-blueviolet)](https://github.com/CassInfra/KubeDoor/blob/main/LICENSE)
[![OSCS Status](https://www.oscs1024.com/platform/badge/CassInfra/KubeDoor.svg?size=small)](https://www.murphysec.com/dr/Zoyt5g0huRavAtItj2)

<img src="https://github.com/user-attachments/assets/3dc6a022-cacf-4b89-9e26-24909102552c" width="80;" alt="kubedoor"/>

# 花折 - KubeDoor

花开堪折直须折🌻莫待无花空折枝

### 开思开源第一弹：**基于AI推荐+专家经验的K8S负载感知调度与容量管控系统**

</div>

---

## 🏷目录
* [🌈概述](#概述)
* [💠架构图](#架构图)
* [💎功能描述](#功能描述)
* [🚀部署说明](#部署说明)
* [⛔注意事项](#注意事项)
* [🌰管控例子](#管控例子)
* [🚩管控原则](#管控原则)
* [🎯线路图](#线路图)
* [🔔KubeDoor交流群](#KubeDoor交流群与🧧赞赏)
* [🙇贡献者](#贡献者)
* [⭐STAR](#STAR)
* [🥰鸣谢](#鸣谢)
* 
---

### 🌈概述

🌼**花折 - KubeDoor** 是一个使用Python + Vue开发，基于K8S准入控制机制的微服务资源管控平台。专注微服务每日高峰时段的资源视角，实现了微服务的资源分析统计与强管控，确保微服务资源的资源申请率和真实使用率一致。

### 💠架构图
![图片](https://github.com/user-attachments/assets/59b7d650-1bf0-4019-bc0b-5613e449b525)

### 💎功能描述

#### 📊采集K8S微服务每日业务高峰时段P95的CPU内存消耗，以及需求、限制值与Pod数。基于采集的数据实现了一个Grafana看板并集成到了WEB UI。
  - 🎨**基于日维度采集每日高峰时段P95的资源数据**,可以很好的观察各微服务长期的资源变化情况，即使查看1年的数据也很流畅。
  - 🏅高峰时段全局资源统计与各**资源TOP10**
  - 🔎命名空间级别高峰时段P95资源使用量与**资源消耗占整体资源的比例**
  - 🧿**微服务级别**高峰期整体资源与使用率分析
  - 📈微服务与**Pod级别**的资源曲线图(需求值,限制值,使用值)
<div align="center">
<img src="https://github.com/user-attachments/assets/8b5f0a7e-42a4-4fc1-9020-4be619f845fd" width="400;" alt="kubedoor1"/>
<img src="https://github.com/user-attachments/assets/cb822804-1b9e-415f-867f-9835375bf439" width="400;" alt="kubedoor2"/>
<img src="https://github.com/user-attachments/assets/ff207f11-a885-4fb3-9662-d795044ee9d4" width="400;" alt="kubedoor3"/>
<img src="https://github.com/user-attachments/assets/0188c261-ec47-49db-b6ba-7ca90235b9b9" width="400;" alt="kubedoor4"/>
</div>

#### 🎡每日从采集的数据中，获取最近10天各微服务的资源信息，获取资源消耗最大日的P95资源，作为微服务的需求值写入数据库。
  - ✨**基于准入控制机制**实现K8S微服务资源的**真实使用率和资源申请需求值保持一致**，具有非常重要的意义。
  - 🌊**K8S调度器**通过真实的资源需求值就能够更精确地将Pod调度到合适的节点上，**避免资源碎片，实现节点的资源均衡**。
  - ♻**K8S自动扩缩容**也依赖资源需求值来判断，**真实的需求值可以更精准的触发扩缩容操作**。
  - 🛡**K8S的保障服务质量**（QoS机制）与需求值结合，真实需求值的Pod会被优先保留，**保证关键服务的正常运行**。

#### 🌐实现了一个K8S管控与展示的WEB UI。

  - ⚙️对微服务的最新、每日高峰期的**P95资源展示**，以及对**Pod数、资源限制值**的维护管理。
  - ⏱️支持**即时、定时、周期性**任务执行微服务的**扩缩容和重启**操作。 
  - 🔒基于NGINX basic**认证**，支持LDAP，支持所有**操作审计**日志与通知。 
<div align="center">
<img src="https://github.com/user-attachments/assets/03f91c23-43e6-4df1-b874-ea3e49d7d04e"  width="900;" alt="kubedoor5"/>
</div>

#### 🚧当微服务更新部署时，基于K8S准入控制机制对资源进行管控【默认不开启】：
  - 🧮**控制每个微服务的Pod数、需求值、限制值**必须与数据库一致，以确保微服务的真实使用率和资源申请需求值相等，从而实现微服务的统一管控与Pod的负载感知调度均衡能力。
  - 🚫**对未管控的微服务，会部署失败并通知**，必须在WEB UI新增微服务后才能部署。（作为新增微服务的唯一管控入口，杜绝未经允许的新服务部署。）
  - 🌟通过本项目基于**K8S准入机制的扩展**思路，大家可以自行简单定制需求，即可对K8S实现各种高灵活性与扩展性附加能力，诸如统一或者个性化的**拦截、管理、策略、标记微服务**等功能。


<div align="center">
  
**K8S准入控制逻辑**
![图片](https://github.com/user-attachments/assets/2052e559-113a-4c32-8abc-b1d1508f70a8)

#### 如果觉得项目不错，麻烦动动小手点个⭐️Star⭐️ 如果你还有其他想法或者需求，欢迎在 issue 中交流

</div>

---

### 🚀部署说明
#### 0. 需要已有 Prometheus监控K8S
需要有`cadvisor`和`kube-state-metrics`这2个JOB，才能采集到K8S的以下指标
- `container_cpu_usage_seconds_total`
- `container_memory_working_set_bytes`
- `container_spec_cpu_quota`
- `kube_pod_container_info`
- `kube_pod_container_resource_limits`
- `kube_pod_container_resource_requests`

#### 1. 部署 Cert-manager

用于K8S Mutating Webhook的强制https认证
```
kubectl apply -f https://StarsL.cn/kubedoor/00.cert-manager_v1.16.2_cn.yaml
```

#### 2. 部署 ClickHouse 并初始化

用于存储采集的指标数据与微服务的资源信息

```bash
curl -s https://StarsL.cn/kubedoor/install-clickhouse.sh|sudo bash
# 完成启动（启动后会自动初始化表结构）
cd /opt/clickhouse && docker compose up -d
```

如果已有ClickHouse，请逐行执行以下SQL，完成初始化表结构

```bash
https://StarsL.cn/kubedoor/kubedoor-init.sql
```

#### 3. 部署KubeDoor

```bash
wget https://StarsL.cn/kubedoor/kubedoor.tgz
tar -zxvf kubedoor.tgz
# 编辑values.yaml文件，根据描述修改内容。
vi kubedoor/values.yaml
# 使用helm安装
helm install kubedoor ./kubedoor
```

#### 4. 访问WebUI 并初始化数据

- 使用节点IP + kubedoor-web的NodePort访问，默认账号密码都是`kubedoor`

- 点击`配置中心`，输入需要采集的历史数据时长，点击更新，即可采集历史数据并更新高峰时段数据到管控表。

- 点击`管控状态`的开关，显示`管控已启用`，表示已开启。

---

### ⛔注意事项

- 部署完成后，默认不会开启管控机制，你可以按上述操作通过WebUI 来开关管控能力。特殊情况下，你也可以使用`kubectl`来开关管控功能：

    ```bash
    # 开启管控
    kubectl apply -f https://StarsL.cn/kubedoor/99.kubedoor-Mutating.yaml
    
    # 关闭管控
    kubectl delete mutatingwebhookconfigurations kubedoor-webhook-configuration
    ```

- 开启开启管控机制后，目前只会拦截deployment的创建，更新，扩缩容操作；管控pod数，需求值，限制值。不会控制其它操作和属性。

- 通过任何方式对Deployment执行扩缩容或者更新操作都会受到管控，管控的目标为Pod数，资源需求值，和资源限制值。

### 🌰管控例子

- 你通过Kubectl对一个Deployment执行了扩容10个Pod后，**会触发拦截机制**，到数据库中去查询该微服务的Pod，然后使用该值来进行实际的扩缩容。（正确的做法应该是在KubeDoor-Web来执行扩缩容操作。）

- 你通过某发布系统修改了Deployment的镜像版本，执行发布操作，**会触发拦截机制**，到数据库中去查询该微服务的Pod数，需求值，限制值，然后使用这些值值以及新的镜像来进行实际的更新操作。

### 🚩管控原则

- **你对deployment的操作不会触发deployment重启的，也没有修改Pod数的：** 触发管控拦截后，只会按照你的操作来更新deployment（不会重启Deployment）

- **你对deployment的操作不会触发deployment重启的，并且修改Pod数的：** 触发管控拦截后，Pod数会根据数据库的值以及你修改的其它信息来更新Deployment。（不会重启Deployment）

- **你对deployment的操作会触发deployment重启的：** 触发管控拦截后，会到数据库中去查询该微服务的Pod数，需求值，限制值，然后使用这些值以及你修改的其它信息来更新Deployment。（会重启Deployment）

### 🎯线路图

- 🥇多K8S支持：在统一的WebUI对多K8S做管控和资源分析展示。
- 🥈英文版发布
- 🥉集成K8S实时监控能力，实现一键部署，整合K8S实时资源看板，接入K8S异常AI分析能力。
- 🏅微服务AI评分：根据资源使用情况，发现资源浪费的问题，做AI综合评分。
- 🏅微服务AI缩容：基于每个微服务高峰期的资源信息，对接AI分析与专家经验，计算每个微服务的pod数是否合理，生成缩容信息统计与缩容指令。
- 🏅根据K8S节点资源使用率做节点管控与调度分析
- 🚩采集更多的微服务资源信息: QPS/JVM/GC
- 🚩针对微服务Pod做精细化操作：隔离、dump、jstack、jfr、jvm、删除

### 🔔KubeDoor交流群与🧧赞赏

### 🙇贡献者
<div align="center">
<a href="https://github.com/CassInfra/KubeDoor/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=CassInfra/KubeDoor" />
</a>
</div>

### ⭐STAR

<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=CassInfra/KubeDoor&type=Date)](https://github.com/CassInfra/KubeDoor)

</div>

### 🥰鸣谢

感谢如下优秀的项目，没有这些项目，不可能会有**KubeDoor**：

- 后端技术栈
  - [Flask](https://flask.palletsprojects.com)
  - [Grafana](https://grafana.com/)
  - [Nginx](https://nginx.org/)

- 前端技术栈
  - [Vue](https://vuejs.org/)
  - [Element Plus](https://element-plus.org)
  - [pure-admin](https://pure-admin.cn/)

- **另外感谢**
  - [CassTime](https://www.casstime.com)：**KubeDoor**的诞生离不开🦄**开思**的支持。
