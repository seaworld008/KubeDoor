# 基于Mutating Webhook的针对微服务Pod数、需求值、限制值强管控的准入控制能力

#### 🚧当微服务更新部署时，基于K8S准入控制机制对资源进行管控【默认不开启】：
  - 🧮**控制每个微服务的Pod数、需求值、限制值**必须与数据库一致，以确保微服务的真实使用率和资源申请需求值相等，从而实现微服务的统一管控与Pod的负载感知调度均衡能力。
  - 🚫**对未管控的微服务，会部署失败并通知**，必须在WEB UI新增微服务后才能部署。（作为新增微服务的唯一管控入口，杜绝未经允许的新服务部署。）
  - 🌟通过本项目基于**K8S准入机制的扩展**思路，大家可以自行简单定制需求，即可对K8S实现各种高灵活性与扩展性附加能力，诸如统一或者个性化的**拦截、管理、策略、标记微服务**等功能。

<div align="center">

**K8S准入控制逻辑**

![kd-k8s](https://raw.githubusercontent.com/CassInfra/KubeDoor/refs/heads/main/screenshot/kd-k8s.png)

</div>

## ⛔注意事项

- 部署完成后，**默认不会开启管控机制**，你可以按上述操作通过WebUI 来开关管控能力。特殊情况下，你也可以使用`kubectl`来开关管控功能：

    ```bash
    # 开启管控
    kubectl apply -f https://StarsL.cn/kubedoor/99.kubedoor-Mutating.yaml
    
    # 关闭管控
    kubectl delete mutatingwebhookconfigurations kubedoor-webhook-configuration
    ```

- **开启管控机制后**，目前只会拦截**deployment的创建，更新，扩缩容**操作；管控**pod数，需求值，限制值**。不会控制其它操作和属性。

- **开启管控机制后**，通过任何方式对Deployment执行扩缩容或者更新操作都会受到管控。

- **开启管控机制后**，扩缩容或者重启Deployment时，Pod数优先取`指定Pod`字段，若该字段为-1，则取`当日Pod`字段。

## 🌰管控例子

- 你通过Kubectl对一个Deployment执行了扩容10个Pod后，**会触发拦截机制**，到数据库中去查询该微服务的Pod，然后使用该值来进行实际的扩缩容。（正确的做法应该是在KubeDoor-Web来执行扩缩容操作。）

- 你通过某发布系统修改了Deployment的镜像版本，执行发布操作，**会触发拦截机制**，到数据库中去查询该微服务的Pod数，需求值，限制值，然后使用这些值值以及新的镜像来进行实际的更新操作。

## 🚩管控原则

- **你对deployment的操作不会触发deployment重启的，也没有修改Pod数的：** 触发管控拦截后，只会按照你的操作来更新deployment（不会重启Deployment）

- **你对deployment的操作不会触发deployment重启的，并且修改Pod数的：** 触发管控拦截后，Pod数会根据数据库的值以及你修改的其它信息来更新Deployment。（不会重启Deployment）

- **你对deployment的操作会触发deployment重启的：** 触发管控拦截后，会到数据库中去查询该微服务的Pod数，需求值，限制值，然后使用这些值以及你修改的其它信息来更新Deployment。（会重启Deployment）

---

