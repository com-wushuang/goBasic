## 背景
- 由于每个 `Service` 都要有一个负载均衡服务，所以这个做法实际上既浪费成本又高。作为用户，我其实更希望看到 `Kubernetes` 为我内置一个全局的负载均衡器。然后，通过我访问的 `URL`，把请求转发给不同的后端 `Service`。
- 这种全局的、为了代理不同后端 `Service` 而设置的负载均衡服务，就是 `Kubernetes` 里的 `Ingress` 服务。
- `Ingress` 的功能其实很容易理解：所谓 `Ingress`，就是 `Service` 的 `Service`。

## 实例
- 举个例子，假如我现在有这样一个站点：`https://cafe.example.com`。
- `https://cafe.example.com/coffee`: 对应的是“咖啡点餐系统”。
- `https://cafe.example.com/tea`: 对应的是“茶水点餐系统”。
- 这两个系统，分别由名叫 `coffee` 和 `tea` 这样两个 `Deployment` 来提供服务。
- 那么现在，如何能使用 `Kubernetes` 的 `Ingress` 来创建一个统一的负载均衡器，从而实现当用户访问不同的域名时，能够访问到不同的 `Deployment` 呢？
```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: cafe-ingress
spec:
  tls:
  - hosts:
    - cafe.example.com
    secretName: cafe-secret
  rules:
  - host: cafe.example.com
    http:
      paths:
      - path: /tea
        backend:
          serviceName: tea-svc
          servicePort: 80
      - path: /coffee
        backend:
          serviceName: coffee-svc
          servicePort: 80
```
- `ingress` 的定义中最重要的是 `rules` 字段，在 `Kubernetes` 里，这个字段叫作：`IngressRule`。
- `IngressRule` 的 `Key` 是 `host`，必须是一个标准的域名格式的字符串，而不能是 `IP` 地址。
- 当访问 `cafe.example.com` 的时候，实际上访问到的是这个 `Ingress` 对象。这样，`Kubernetes` 就能使用 `IngressRule` 来对请求进行下一步转发。
- 转发的规则依赖于 `path` 字段，这里的每一个 `path` 都对应一个后端 `Service`。在本例中，定义了两个 `path`，它们分别对应 `coffee` 和 `tea` 这两个 `Deployment` 的 `Service`。

## 原理
- 通过上面的讲解，不难看到，所谓 `Ingress` 对象，其实就是 `Kubernetes` 项目对“反向代理”的一种抽象。
- 一个 `Ingress` 对象的主要内容，实际上就是一个 `反向代理` 服务（比如：`Nginx`）的配置文件的描述。而这个代理服务对应的转发规则，就是 `IngressRule`。
- 这就是为什么在每条 `IngressRule` 里，需要有一个 `host` 字段来作为这条 `IngressRule` 的入口，然后还需要有一系列 `path` 字段来声明具体的转发策略。这其实跟 `Nginx`、`HAproxy` 等项目的配置文件的写法是一致的。
- 而有了 `Ingress` 这样一个统一的抽象，`Kubernetes` 的用户就无需关心 Ingress 的具体细节了。
- 你只需要从社区里选择一个具体的 `Ingress Controller`，把它部署在 `Kubernetes` 集群里即可。
- 然后，这个 `Ingress Controller` 会根据你定义的 `Ingress` 对象，提供对应的代理能力。
- 目前，业界常用的各种反向代理项目，比如 `Nginx`、`HAProxy`、`Envoy`、`Traefik` 等，都已经为 `Kubernetes` 专门维护了对应的 `Ingress Controller`。
- 部署 `Nginx Ingress Controller` 的清单文件如下:
```yaml
kind: ConfigMap
apiVersion: v1
metadata:
  name: nginx-configuration
  namespace: ingress-nginx
  labels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: nginx-ingress-controller
  namespace: ingress-nginx
  labels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: ingress-nginx
      app.kubernetes.io/part-of: ingress-nginx
  template:
    metadata:
      labels:
        app.kubernetes.io/name: ingress-nginx
        app.kubernetes.io/part-of: ingress-nginx
      annotations:
        ...
    spec:
      serviceAccountName: nginx-ingress-serviceaccount
      containers:
        - name: nginx-ingress-controller
          image: quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.20.0
          args:
            - /nginx-ingress-controller
            - --configmap=$(POD_NAMESPACE)/nginx-configuration
            - --publish-service=$(POD_NAMESPACE)/ingress-nginx
            - --annotations-prefix=nginx.ingress.kubernetes.io
          securityContext:
            capabilities:
              drop:
                - ALL
              add:
                - NET_BIND_SERVICE
            # www-data -> 33
            runAsUser: 33
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
            - name: http
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
          ports:
            - name: http
              containerPort: 80
            - name: https
              containerPort: 443
```
- 在上述 `YAML` 文件中，我们定义了一个使用 `nginx-ingress-controller` 镜像的 `Pod`。
- 这个 `Pod` 本身，就是一个监听 `Ingress` 对象以及它所代理的后端 `Service` 变化的控制器。
- 当新的 `Ingress` 对象由用户创建后，`nginx-ingress-controller` 就会根据 `Ingress` 对象里定义的内容，生成一份对应的 `Nginx` 配置文件（`/etc/nginx/nginx.conf`），并使用这个配置文件启动一个 `Nginx` 服务。
- 而一旦 `Ingress` 对象被更新，`nginx-ingress-controller` 就会更新这个配置文件。
- 一个 `Nginx Ingress Controller` 为你提供的服务，其实是一个可以根据 `Ingress` 对象和被代理后端 `Service` 的变化，来自动进行更新的 `Nginx` 负载均衡器。
- 此外，`nginx-ingress-controller` 还允许你通过 `Kubernetes` 的 `ConfigMap` 对象来对上述 `Nginx` 配置文件进行定制。这个 `ConfigMap` 的名字，需要以参数的方式传递给 `nginx-ingress-controller`。而你在这个 `ConfigMap` 里添加的字段，将会被合并到最后生成的 `Nginx` 配置文件当中。
- 可以看到，一个 `Nginx Ingress Controller` 为你提供的服务，其实是一个可以根据 `Ingress` 对象和被代理后端 `Service` 的变化，来自动进行更新的 `Nginx` 负载均衡器。
- 为了让用户能够用到这个 `Nginx`，我们就需要创建一个 `Service` 来把 `Nginx Ingress Controller` 管理的 Nginx 服务暴露出去，如下所示：
```shell
$ kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/baremetal/service-nodeport.yaml
```
```yaml
apiVersion: v1
kind: Service
metadata:
  name: ingress-nginx
  namespace: ingress-nginx
  labels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
spec:
  type: NodePort
  ports:
    - name: http
      port: 80
      targetPort: 80
      protocol: TCP
    - name: https
      port: 443
      targetPort: 443
      protocol: TCP
  selector:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
```
- 这个 `Service` 的唯一工作，就是将所有携带 `ingress-nginx` 标签的 Pod 的 `80` 和 `433` 端口暴露出去。
- 上述操作完成后，你一定要记录下这个 `Service` 的访问入口，即：宿主机的地址和 `NodePort` 的端口，如下所示：
```shell
$ kubectl get svc -n ingress-nginx
NAME            TYPE       CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
ingress-nginx   NodePort   10.105.72.96   <none>        80:30044/TCP,443:31453/TCP   3h
```
- 在 Ingress Controller 和它所需要的 Service 部署完成后，我们就可以使用它了。

## 使用
- 首先，我们要在集群里部署我们的应用 Pod 和它们对应的 Service，如下所示：
```shell
$ kubectl create -f cafe.yaml
```
- 然后创建 `Ingress` 所需的 `SSL` 证书（`tls.crt`）和密钥（`tls.key`），这些信息都是通过 `Secret` 对象定义好的，如下所示：
```shell
$ kubectl create -f cafe-secret.yaml
```
- 这一步完成后，我们就可以创建在本篇文章一开始定义的 Ingress 对象了，如下所示：
```shell
$ kubectl create -f cafe-ingress.yaml
```
- 这时候，我们就可以查看一下这个 Ingress 对象的信息，如下所示：
```shell
$ kubectl get ingress
NAME           HOSTS              ADDRESS   PORTS     AGE
cafe-ingress   cafe.example.com             80, 443   2h

$ kubectl describe ingress cafe-ingress
Name:             cafe-ingress
Namespace:        default
Address:          
Default backend:  default-http-backend:80 (<none>)
TLS:
  cafe-secret terminates cafe.example.com
Rules:
  Host              Path  Backends
  ----              ----  --------
  cafe.example.com  
                    /tea      tea-svc:80 (<none>)
                    /coffee   coffee-svc:80 (<none>)
Annotations:
Events:
  Type    Reason  Age   From                      Message
  ----    ------  ----  ----                      -------
  Normal  CREATE  4m    nginx-ingress-controller  Ingress default/cafe-ingress
```
- 可以看到，这个 `Ingress` 对象最核心的部分，正是 `Rules` 字段。其中，我们定义的 `Host` 是 `cafe.example.com`，它有两条转发规则（`Path`），分别转发给 `tea-svc` 和 `coffee-svc`。
- 如果我的请求没有匹配到任何一条 IngressRule，那么会发生什么呢？
- 既然 `Nginx Ingress Controller` 是用 `Nginx` 实现的，那么它当然会为你返回一个 `Nginx` 的 `404` 页面。
- 不过，`Ingress Controller` 也允许你通过 `Pod` 启动命令里的 `–default-backend-service` 参数，设置一条默认规则，比如：`–default-backend-service=nginx-default-backend`。
- 这样，任何匹配失败的请求，就都会被转发到这个名叫 `nginx-default-backend` 的 `Service`。
- 所以，你就可以通过部署一个专门的 `Pod`，来为用户返回自定义的 `404` 页面了。