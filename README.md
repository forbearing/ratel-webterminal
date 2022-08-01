## ratel-webterminal



该项目大量参考了以下两个项目

- https://github.com/maoqide/kubeutil
- https://github.com/kubernetes/dashboard

感谢 kubeutil 和 dashboard 作者!



### 1. 开启运行

```bash
NAMESPACE="ratel" go run ratel-webterminal.go --kubeconfig ~/.kube/config
```



`go run ratel-webterminal.go --kubeconfig ~/.kube/config`

### 2. 在 default namespace 下创建一个 pod

`kubectl -n default run nginx --image nginx`

### 3. 打开 webterminal

http://localhost:8080/terminal?namespace=default&pod=nginx&container=nginx

### 4. 查看 pod 日志

http://localhost:8080/logs?namespace=default&pod=nginx&container=nginx



## TODO

- [x] 通过 pod informer 来监控所有 pod, 通过 pod lister 来获取 pod 资源, 而不是每次通过 RESTClient 来直接访问 kube-apiserver, 减少访问 kube-apiserver 的次数, 减轻 kube-apiserver 的压力.
- [ ] leader election