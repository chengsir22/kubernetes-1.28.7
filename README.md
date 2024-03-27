# go-DeBug-Kubernetes

> 前言，叒🐶想学云原生，但是因为没有实习经验，加上研究生有不是这个课题，找实习过程中被各大厂商 “人嫌狗厌”。（认识到自己的不足，遂去找实习，但是你想实习，得有实习经验啊，所以我找实习啊，但是你得有实习经验 🤡。。。。。。）
> 面试时各个面试官都建议我看kubernetes源码，但是k8s代码太多了，根本不知道怎么看。
> 发现 b站up有源码解读的视频，感谢 [up海龙张](https://space.bilibili.com/1748865719) 带我入门 [Kubernetes源码开发之旅](https://www.bilibili.com/video/BV11U4y1y7V7/?vd_source=143ed9e5b9a8342f01a329d8e2cbaed2) , 师父引进门修行看个人，至少找到门了🥰
## 环境配置

已经有同学记了[配置笔记](https://blog.csdn.net/a1369760658/article/details/135118415)，跟着配置就行，高版本坑还是挺多的，配了一晚上🥲 我的是1.28.7，参考[大佬1.29配置过程](https://www.jianshu.com/p/e3a2bb22667f)
注意在一个干净的没有安装k8s的linux中进行配置。为了方便后续的操作，请以[root身份登陆桌面](https://zhuanlan.zhihu.com/p/355555221?utm_id=0)，[root登陆ssh](https://blog.csdn.net/zhangzeyuaaa/article/details/132764683)

补充设置：[ubuntu 设置静态 ip](https://blog.csdn.net/coward__123/article/details/134285762)

```console
sudo apt update
sudo apt install openssh-server  # 配置 ssh
ssh-copy-id -p 10200 root@192.168.3.87

sudo systemctl disable --now ufw # 关闭防火墙
```

### 环境变量

```console
export GOPATH="/root/go"  
export GOBIN="/root/go/bin"  
export PATH="/usr/local/go/bin:$GOPATH/bin:${PATH}"  

export PATH="/root/go/src/k8s.io/kubernetes-1.28.7/third_party/etcd:${PATH}"

export CONTAINER_RUNTIME_ENDPOINT="unix:///run/containerd/containerd.sock"
export KUBECONFIG=/var/run/kubernetes/admin.kubeconfig
```
### 编译k8s

关闭 swap 分区

```console
sudo vim /etc/fstab  # 注释带swap的一行
sudo swapoff -a
free -h
```

```bash
# 编译并启动本地单节点集群，别忘了以debug模式，golang.sh部分命令需要修改
./hack/local-up-cluster.sh 

# 想要使用集群，可以开启另一个端口，执行下面的两条命令，是在指定kubeconfig文件
export KUBECONFIG=/var/run/kubernetes/admin.kubeconfig
# 就是kubectl命令的位置，我们使用它就是在使用kubectl
cluster/kubectl.sh get nodes
```

巨坑，coredns 状态 CrashLoopBackOff
```bash
vim /etc/resolv.conf
nameserver 192.168.40.2
nameserver 8.8.8.8
nameserver 114.114.114.114
```
### git

```bash
sudo apt install git
git config --global user.name "LuCheng"
git config --global user.email "2437383481@qq.com"

// 配置github
ssh-keygen -t ed25519 -C "2437383481@qq.com"
cat ~/.ssh/id_ed25519.pub
ssh -T git@github.com
```

## Go-Delve

[delve 使用入门](https://blog.csdn.net/RA681t58CJxsgCkJ31/article/details/129964636)

```shell
dlv debug <包名> -- <子命令> <参数>
```

### 开始调试

```console
# 打断点
break <文件路径>:<行号>

continue # 向下运行到断点
next # 单步调试
step # 单步调试某个函数
stepout # 从当前函数跳出
```

### 远程调试

```console
dlv --headless debug <包名> -- <子命令> <参数>
dlv connect <IP>:<port>
```

## Debug 模式启动集群

```bash
hack/local-up-cluster.sh

# 查看机器启动了哪些k8s组件
ps -a | grep kube

# 以API Server举例，如何使用delve重新启动kubernetes的一个组件，进而可以进行远程调试
ps -ef | grep kube-apiserver
# 记录kube-apiserver进程启动的命令
root        2795    1789 18 15:55 pts/0    00:00:21 /root/go/src/k8s.io/kubernetes-1.28.7/_output/local/bin/linux/amd64/kube-apiserver --authorization-mode=Node,RBAC  --cloud-provider= --cloud-config=   --v=3 --vmodule= --audit-policy-file=/tmp/local-up-cluster.sh.PwhXqJ/kube-audit-policy-file --audit-log-path=/tmp/kube-apiserver-audit.log --authorization-webhook-config-file= --authentication-token-webhook-config-file= --cert-dir=/var/run/kubernetes --egress-selector-config-file=/tmp/local-up-cluster.sh.PwhXqJ/kube_egress_selector_configuration.yaml --client-ca-file=/var/run/kubernetes/client-ca.crt --kubelet-client-certificate=/var/run/kubernetes/client-kube-apiserver.crt --kubelet-client-key=/var/run/kubernetes/client-kube-apiserver.key --service-account-key-file=/tmp/local-up-cluster.sh.PwhXqJ/kube-serviceaccount.key --service-account-lookup=true --service-account-issuer=https://kubernetes.default.svc --service-account-jwks-uri=https://kubernetes.default.svc/openid/v1/jwks --service-account-signing-key-file=/tmp/local-up-cluster.sh.PwhXqJ/kube-serviceaccount.key --enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,Priority,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota,NodeRestriction --disable-admission-plugins= --admission-control-config-file= --bind-address=0.0.0.0 --secure-port=6443 --tls-cert-file=/var/run/kubernetes/serving-kube-apiserver.crt --tls-private-key-file=/var/run/kubernetes/serving-kube-apiserver.key --storage-backend=etcd3 --storage-media-type=application/vnd.kubernetes.protobuf --etcd-servers=http://127.0.0.1:2379 --service-cluster-ip-range=10.0.0.0/24 --feature-gates=AllAlpha=false --external-hostname=localhost --requestheader-username-headers=X-Remote-User --requestheader-group-headers=X-Remote-Group --requestheader-extra-headers-prefix=X-Remote-Extra- --requestheader-client-ca-file=/var/run/kubernetes/request-header-ca.crt --requestheader-allowed-names=system:auth-proxy --proxy-client-cert-file=/var/run/kubernetes/client-auth-proxy.crt --proxy-client-key-file=/var/run/kubernetes/client-auth-proxy.key --cors-allowed-origins=/127.0.0.1(:[0-9]+)?$,/localhost(:[0-9]+)?$
root        5109    3583  0 15:57 pts/3    00:00:00 grep --color=auto kube-apiserver

# 杀掉api-server
kill -9 2795
# 使用delve重新启动
dlv --headless exec /root/go/src/k8s.io/kubernetes-1.28.7/_output/local/bin/linux/amd64/kube-apiserver --listen=:12345 --api-version=2 --log --log-output=debugger,gdbwire,lldbout,debuglineerr,rpc,dap,fncall,minidump --log-dest=/root/delve-log/log -- --authorization-mode=Node,RBAC  --cloud-provider= --cloud-config=   --v=3 --vmodule= --audit-policy-file=/tmp/local-up-cluster.sh.PwhXqJ/kube-audit-policy-file --audit-log-path=/tmp/kube-apiserver-audit.log --authorization-webhook-config-file= --authentication-token-webhook-config-file= --cert-dir=/var/run/kubernetes --egress-selector-config-file=/tmp/local-up-cluster.sh.PwhXqJ/kube_egress_selector_configuration.yaml --client-ca-file=/var/run/kubernetes/client-ca.crt --kubelet-client-certificate=/var/run/kubernetes/client-kube-apiserver.crt --kubelet-client-key=/var/run/kubernetes/client-kube-apiserver.key --service-account-key-file=/tmp/local-up-cluster.sh.PwhXqJ/kube-serviceaccount.key --service-account-lookup=true --service-account-issuer=https://kubernetes.default.svc --service-account-jwks-uri=https://kubernetes.default.svc/openid/v1/jwks --service-account-signing-key-file=/tmp/local-up-cluster.sh.PwhXqJ/kube-serviceaccount.key --enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,Priority,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota,NodeRestriction --disable-admission-plugins= --admission-control-config-file= --bind-address=0.0.0.0 --secure-port=6443 --tls-cert-file=/var/run/kubernetes/serving-kube-apiserver.crt --tls-private-key-file=/var/run/kubernetes/serving-kube-apiserver.key --storage-backend=etcd3 --storage-media-type=application/vnd.kubernetes.protobuf --etcd-servers=http://127.0.0.1:2379 --service-cluster-ip-range=10.0.0.0/24 --feature-gates=AllAlpha=false --external-hostname=localhost --requestheader-username-headers=X-Remote-User --requestheader-group-headers=X-Remote-Group --requestheader-extra-headers-prefix=X-Remote-Extra- --requestheader-client-ca-file=/var/run/kubernetes/request-header-ca.crt --requestheader-allowed-names=system:auth-proxy --proxy-client-cert-file=/var/run/kubernetes/client-auth-proxy.crt --proxy-client-key-file=/var/run/kubernetes/client-auth-proxy.key --cors-allowed-origins="/127.0.0.1(:[0-9]+)?$,/localhost(:[0-9]+)?$"
```

### 命令行模式

```console
dlv connect 192.168.40.52:12345
break cmd/kube-apiserver/apiserver.go:33 # 开始调试
```
### vscode
vscode有点小坑，需要添加配置
```json
"go.delveConfig": {
    "debugAdapter": "legacy"
}
```
launch.json
```json
{
    // 使用 IntelliSense 了解相关属性。 
    // 悬停以查看现有属性的描述。
    // 欲了解更多信息，请访问: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        
        {
            "name": "Connect to server",
            "type": "go",
            "request": "attach",
            "mode": "remote",
            // "remotePath": "${workspaceFolder}",
            "port": 12345,
            "host": "192.168.3.87"
        }
    ]
}
```

也可以使用 goland 配置更加简单，但是我就是喜欢vscode🥰
### postman

[大佬已经写得很详细了](https://blog.csdn.net/a1369760658/article/details/135147441) ，照着做不会有问题。