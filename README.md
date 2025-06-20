<div align="center">

[![StarsL.cn](https://img.shields.io/badge/website-StarsL.cn-orange)](https://starsl.cn)
[![Commits](https://img.shields.io/github/commit-activity/m/CassInfra/KubeDoor?color=ffff00)](https://github.com/CassInfra/KubeDoor/commits/main)
[![open issues](http://isitmaintained.com/badge/open/CassInfra/KubeDoor.svg)](https://github.com/CassInfra/KubeDoor/issues)
[![Python](https://img.shields.io/badge/python-v3.11-3776ab)](https://nodejs.org)
[![Node.js](https://img.shields.io/badge/node.js-v22-229954)](https://nodejs.org)
[![GitHub license](https://img.shields.io/badge/license-MIT-blueviolet)](https://github.com/CassInfra/KubeDoor/blob/main/LICENSE)
[![OSCS Status](https://www.oscs1024.com/platform/badge/CassInfra/KubeDoor.svg?size=small)](https://www.murphysec.com/dr/Zoyt5g0huRavAtItj2)

<img src="https://github.com/user-attachments/assets/3dc6a022-cacf-4b89-9e26-24909102552c" width="80;" alt="kubedoor"/>

# 花折 - KubeDoor

花开堪折直须折🌻莫待无花空折枝

</div>

---
**国内用户如果访问图片异常，可以访问Gitee同步站：https://gitee.com/starsl/KubeDoor**


## 🏷目录
* [🌈概述](#概述)
* [💠全新架构](#全新架构)
* [💎功能描述](#功能描述)
* [🚀部署说明](#kubedoor-全新架构全新部署)
* [📀使用介绍: KubeDoor MCP来啦](#kubedoor-使用介绍-mcp来啦)
* [🔔KubeDoor交流群](#KubeDoor交流群与赞赏)
* [🙇贡献者](#贡献者)
* [🎯2025 KubeDoor RoadMap](#2025-kubedoor-roadmap)
* [⭐STAR History](#STAR-History)
* [🥰鸣谢](#鸣谢)

---

## 🌈概述

🌼**花折 - KubeDoor** 是一个使用Python + Vue开发，基于K8S准入控制机制的微服务资源管控平台，以及支持多K8S集群统一远程存储、监控、告警、通知、展示的一站式K8S监控平台，并且专注微服务每日高峰时段的资源视角，实现了微服务的资源分析统计与强管控，确保微服务资源的资源申请率和真实使用率一致。
## 💠全新架构
<div align="center">

### 🎉花折 - KubeDoor 1.3发布！多K8S支持的MCP来啦！全新架构，多K8S管控支持，多K8S统一监控、告警、展示最佳实践🎉
![KubeDoor1 0 0 drawio](screenshot/1.0/kubedoor1.0-arch.png)
</div>

## 💎功能描述
### 💥花折 KubeDoor MCP
**💠KubeDoor MCP 体验版来啦！接入任意MCP客户端，即可与LLM对话方式：对您的所有K8S集群操作、资源异常故障分析。**
- 使用任意MCP客户端, 新增MCP服务器, 选择sse类型,地址输入：`http://{nodeIP}:{kubedoor-mcp-NodePort}/sse`即可接入KubeDoor MCP.
- 基于KubeDoor的多K8S管理监控架构与已有的API接口，可以快速生成大量MCP工具来操作K8S与Grafana数据分析，敬请期待！
- 已接入的工具列表
![图片](https://github.com/user-attachments/assets/19f50de7-248d-429d-9c19-c3a6a2282716)
- 演示
![kubedoor-mcp](https://github.com/user-attachments/assets/26e03c8e-4038-4094-affe-1d4de85d4675)
>由于大语言模型的不确定性，请尽量使用带工具确认的的MCP客户端（cline、Cursor）。
>
>目前为体验版，暂未做MCP WEB客户端以及工具调用鉴权，请勿将MCP服务端地址暴露到公网。
### 1. 📡资源监控分析：多K8S集群统一监控最佳实践
<div align="center">
<img src="./screenshot/1.0/vm-arch.png" width="650;" />
</div>

  - 🌊支持**多K8S集群**统一远程存储、监控、告警、通知、展示的一站式K8S监控方案。
  - 📀Helm一键部署完成监控、采集、展示、告警、通知（多K8S集群监控从未如此简单✨）。
  - 🚀基于VictoriaMetrics全套方案实现多K8S统一监控，统一告警规则管理，实现免配置完整指标采集。
  - 🎨WEBUI集成了K8S节点监控看板与K8S资源监控看板，均支持在单一看板中查看各个K8S集群的资源情况。
  - 📐集成了大量K8S资源，JVM资源与K8S节点的告警规则，并支持统一维护管理，支持对接企微，钉钉，飞书告警通知及灵活的@机制。
  - 🎭实时监控管理页面，对K8S资源，节点资源统一监控展示的Grafana看板。
  - ⏱️支持**即时、定时、周期性**任务执行微服务的**扩缩容和重启**操作。 
<div align="center">
   
| <img src="./screenshot/1.0/1.png"/> | <img src="./screenshot/1.0/2.jpg"/> |<img src="./screenshot/1.0/3.png"/> | 
| ------------------------------------| ----------------------------------- | ---------------------------------- |
| <img src="./screenshot/alert1.png"/> | <img src="./screenshot/alert2.png"/> |<img src="./screenshot/1.0/14.png"/>| 

</div>

### 2. 🧬异常告警聚合

- 🦄K8S微服务统一告警分析与处理页面，告警按天智能聚合，相同告警按日累计计数，每日告警清晰明了。
- 🕹️支持对POD进行隔离，删除，Java dump，jstack，jfr，JVM数据采集分析等操作，并通知到群。

<div align="center">

| <img src="./screenshot/1.0/4.jpg"/> | <img src="./screenshot/1.0/5.png"/> | <img src="./screenshot/1.0/15.jpg"/> |
| ------------------------------------| ----------------------------------- | ----------------------------------- |
| <img src="./screenshot/1.0/6.png"/> | <img src="./screenshot/1.0/7.jpg"/> | <img src="./screenshot/1.0/8.png"/> |
</div>

### 3. 💠高峰期P95资源采集分析

#### 📊采集K8S微服务每日业务高峰时段P95的CPU内存消耗，以及需求、限制值与Pod数。基于采集的数据实现了一个Grafana看板并集成到了WEB UI。
<div align="center">
   
| <img src="./screenshot/1.0/9.png" width="700;" />| 
| ------------------------------------|
</div>

  - ⚙️对微服务每日高峰期的**P95资源展示**，以及对**Pod数、资源限制值**的维护管理。
  - 🎨**基于日维度采集每日高峰时段P95的资源数据**,可以很好的观察各微服务长期的资源变化情况，即使查看1年的数据也很流畅。
  - 🏅高峰时段全局资源统计与各**资源TOP10**
  - 🔎命名空间级别高峰时段P95资源使用量与**资源消耗占整体资源的比例**
  - 🧿**微服务级别**高峰期整体资源与使用率分析
  - 📈微服务与**Pod级别**的资源曲线图(需求值,限制值,使用值)
<div align="center">
  
|<img src="./screenshot/kd1.jpg"/>|<img src="./screenshot/kd2.jpg"/>|
|-|-|
|<img src="./screenshot/kd3.jpg"/>|<img src="./screenshot/kd4.jpg"/>|
</div>

#### 🎡每日从采集的数据中，获取最近10天各微服务的资源信息，获取资源消耗最大日的P95资源，作为微服务的需求值写入数据库。
  - ✨**基于准入控制机制**实现K8S微服务资源的**真实使用率和资源申请需求值保持一致**，具有非常重要的意义。
  - 🌊**K8S调度器**通过真实的资源需求值就能够更精确地将Pod调度到合适的节点上，**避免资源碎片，实现节点的资源均衡**。
  - ♻**K8S自动扩缩容**也依赖资源需求值来判断，**真实的需求值可以更精准的触发扩缩容操作**。
  - 🛡**K8S的保障服务质量**（QoS机制）与需求值结合，真实需求值的Pod会被优先保留，**保证关键服务的正常运行**。

### 4. 🛡基于K8S准入控制针对微服务Pod数、需求值、限制值与高峰资源强一致的管控能力

- #### <a target="_blank" href="docs/K8S资源管控功能说明.md">👑基于准入控制的管控能力详解</a>

### 5. ✨其它
  - ❤️Agent管理页面：更新，维护Agent状态，配置采集与管控。
  - 🔒基于NGINX basic**认证**，支持LDAP，支持接口级权限控制，支持所有**操作审计**日志与通知。
  - 📊所有看板基于Grafana创建，并整合到前端UI内，使得数据分析可以快速实现更优雅的展示。
<div align="center">
   
| <img src="./screenshot/1.0/11.jpg" width="800;" />| 
| ------------------------------------|
</div>

## 📀KubeDoor 全新架构，全新部署

#### 👉🏻<a target="_blank" href="docs/灵活部署方案.md">点击查看灵活部署方案</a>

#### 🛠快速安装
```
### 【下载helm包】
wget https://StarsL.cn/kubedoor/kubedoor-1.3.2.tgz
tar -zxvf kubedoor-1.3.2.tgz
cd kubedoor
### 【master端安装】
# 编辑values-master.yaml文件，请仔细阅读注释，根据描述修改配置内容。
# try
helm upgrade -i kubedoor . --namespace kubedoor --create-namespace --values values-master.yaml --dry-run --debug
# install
helm upgrade -i kubedoor . --namespace kubedoor --create-namespace --values values-master.yaml
### 【agent端安装】
# 编辑values-agent.yaml文件，请仔细阅读注释，根据描述修改配置内容。
helm upgrade -i kubedoor-agent . --namespace kubedoor --create-namespace --values values-agent.yaml --set tsdb.external_labels_value=xxxxxxxx
```
### ♻更新KubeDoor 1.3.2
**♻[更新说明](https://github.com/CassInfra/KubeDoor/releases/tag/1.3.2)**
```
# 下载安装包
wget https://StarsL.cn/kubedoor/kubedoor-1.3.2.tgz
tar -zxvf kubedoor-1.3.2.tgz
```
```
# 旧版本更新到1.3.0及以上版本需要新增2个数据库字段
ALTER TABLE kubedoor.k8s_agent_status ADD COLUMN nms_not_confirm Bool DEFAULT false AFTER admission_namespace;
ALTER TABLE kubedoor.k8s_agent_status ADD COLUMN scheduler Bool DEFAULT false AFTER nms_not_confirm;
```
##### **请参考已经已经部署好的configmap: `kubedoor-info`中的`VictoriaMetrics`, `ClickHouse`等配置项修改`values-master.yaml`, `values-agent.yaml`的对应配置，确保使用的配置与旧版本一致。(因yaml配置有调整，不可直接替换文件。)**
```
#【master端更新】
helm upgrade -i kubedoor . --namespace kubedoor --create-namespace --values values-master.yaml
#【agent端更新】
helm upgrade -i kubedoor-agent . --namespace kubedoor --create-namespace --values values-agent.yaml --set tsdb.external_labels_value=xxxxxxxx
```

## 📀KubeDoor 使用介绍 MCP来啦

#### 🌐访问WebUI 并初始化数据

1. 使用K8S节点IP + kubedoor-web的NodePort访问，默认账号密码都是 **`kubedoor`**
2. 点击`agent管理`，先开启`自动采集`，设置好`高峰期时段`，再执行采集：输入需要采集的`历史数据时长`，点击`采集`，即可采集历史数据并更新高峰时段数据到管控表。
>开启自动采集后，每天凌晨1点会采集前一天的高峰期数据，并将10天内最大资源消耗日的数据写入到更新管控表。

>重复执行`采集`不会导致重复写入数据，请放心使用；每次采集后都会自动将10天内最大资源消耗日的数据写入到管控表。如果耗时较长，请等待采集完成或缩短采集时长。

>如果您是新安装的监控系统，并且已过了当天的高峰期时段，将会无法采集到数据；需要等第二天高峰期时段之后才能采集到数据。

#### 💠使用KubeDoor MCP
##### 花折 KubeDoor MCP 体验版上线啦
- 使用任意MCP客户端, 新增MCP服务器, 选择sse类型,地址输入：`http://{nodeIP}:{kubedoor-mcp-NodePort}/sse`即可接入KubeDoor MCP.
- 由于大语言模型的不确定性，请尽量使用带工具确认的的MCP客户端（cline、Cursor）。
- 目前为体验版，暂未做MCP WEB客户端以及工具调用鉴权，请勿将MCP服务端地址暴露到公网。
- 基于KubeDoor的多K8S管理监控架构与已有的API接口，可以快速生成大量MCP工具来操作K8S与Grafana数据分析，敬请期待！
- 已接入的工具列表
![图片](https://github.com/user-attachments/assets/19f50de7-248d-429d-9c19-c3a6a2282716)
- 演示
![kubedoor-mcp](https://github.com/user-attachments/assets/26e03c8e-4038-4094-affe-1d4de85d4675)

---

## 🔔KubeDoor交流群与🧧赞赏

<div align="center">

#### 如果觉得项目不错，麻烦动动小手点个⭐️Star⭐️ 如果你还有其他想法或者需求，欢迎在 issue 中交流
<img src="https://github.com/user-attachments/assets/eb324f3d-ea4e-4d30-a80c-36c5dfb7c090" width="600;" alt="kubedoor"/>

无法加群可关注公众号入群

</div>

## 🙇贡献者
<div align="center">
<table>
<tr>
    <td align="center">
        <a href="https://github.com/starsliao">
            <img src="https://avatars.githubusercontent.com/u/3349611?v=4" width="100;" alt="StarsL.cn"/>
            <br />
            <sub><b>StarsL.cn</b></sub>
        </a>
    </td>
    <td align="center">
        <a href="https://github.com/xiaofennie">
            <img src="https://avatars.githubusercontent.com/u/47970207?v=4" width="100;" alt="xiaofennie"/>
            <br />
            <sub><b>xiaofennie</b></sub>
        </a>
    </td>
    <td align="center">
        <a href="https://github.com/shidousanxia">
            <img src="https://avatars.githubusercontent.com/u/61586033?v=4" width="100;" alt="shidousanxia"/>
            <br />
            <sub><b>shidousanxia</b></sub>
        </a>
    </td>
    <td align="center">
        <a href="https://github.com/comqx">
            <img src="https://avatars.githubusercontent.com/u/30148386?v=4" width="100;" alt="comqx"/>
            <br />
            <sub><b>comqx</b></sub>
        </a>
    </td>
  </tr>
</table>
</div>

## 🎯2025 KubeDoor RoadMap
- **[📅KubeDoor 项目进度](https://github.com/orgs/CassInfra/projects/1/views/1)**
- 🥈英文版发布
- 🏅微服务AI评分：根据资源使用情况，发现资源浪费的问题，结合AI缩容，降本增效，做AI综合评分，接入K8S异常AI分析能力。
- 🏅微服务AI缩容：基于微服务高峰期的资源信息，对接AI分析与专家经验，计算微服务Pod数是否合理，生成缩容指令与统计。
- 🏅根据K8S节点资源使用率做节点管控与调度分析
- ✅采集更多的微服务资源信息: QPS/JVM/GC
- ✅针对微服务Pod做精细化操作：隔离、删除、dump、jstack、jfr、jvm
- ✅K8S资源告警管理，按日智能聚合。
- ✅多K8S支持：在统一的WebUI对多K8S做管控和资源分析展示。
- ✅集成K8S实时监控能力，实现一键部署，整合K8S实时资源看板。

## ⭐STAR History

<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=CassInfra/KubeDoor&type=Date)](https://github.com/CassInfra/KubeDoor)

</div>

## 🥰鸣谢

感谢如下优秀的项目，没有这些项目，不可能会有**KubeDoor**：

- 前后端技术栈
  - [Vue](https://vuejs.org/)
  - [Element Plus](https://element-plus.org)
  - [pure-admin](https://pure-admin.cn/)
  - [Flask](https://flask.palletsprojects.com)

- 基础服务
  - [VictoriaMetrics](https://victoriametrics.com/)
  - [ClickHouse](https://clickhouse.com/)
  - [Grafana](https://grafana.com/)
  - [Nginx](https://nginx.org/)

- **特别鸣谢**
  - [**CassTime**](https://www.casstime.com)：**KubeDoor**的诞生离不开🦄**开思**的支持。
