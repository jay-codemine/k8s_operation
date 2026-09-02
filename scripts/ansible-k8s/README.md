# Ansible 一键部署 Kubernetes 集群

从**裸机/虚拟机**到**可用集群**（含加节点、高可用）的全流程 playbook。

## 目录结构

```
ansible-k8s/
├── inventory/
│   ├── hosts.ini              # 主机清单（改 IP/用户名）
│   └── group_vars/all.yml     # 全局变量（版本/网段/仓库）
├── roles/
│   ├── common/                # 前置：swap/内核/防火墙/SELinux
│   ├── runtime/               # 容器运行时：containerd（systemd cgroup）
│   ├── kubernetes/            # kubeadm/kubelet/kubectl
│   ├── master-init/           # kubeadm init（第一台 master）
│   ├── network/               # CNI：Calico/Flannel
│   ├── join-worker/           # worker 节点加入
│   └── join-master/           # 额外 master 加入（高可用）
├── site.yml                   # 一键部署完整集群
├── join-node.yml              # 增量加节点
├── reset-cluster.yml          # 重置集群
└── README.md
```

## 前置条件

1. 控制机（你本地）安装 `ansible`
   ```bash
   pip install ansible
   # 或 apt install ansible / dnf install ansible
   ```
2. 所有目标节点：
   - 能 SSH 免密登录（或配置 `ansible_password`）
   - 有 root 或 sudo 权限
   - 2C2G 以上配置
3. 节点间网络互通

## 快速开始

### 1. 改主机清单

编辑 `inventory/hosts.ini`，把 IP 改成你的实际节点：

```ini
[masters]
k8s-master-1 ansible_host=192.168.1.101

[workers]
k8s-worker-1 ansible_host=192.168.1.102
k8s-worker-2 ansible_host=192.168.1.103
```

### 2. 改全局变量（关键）

编辑 `inventory/group_vars/all.yml`：

| 变量 | 必改 | 说明 |
|---|---|---|
| `control_plane_endpoint` | ⭐ 必改 | 单 master 填 IP；多 master 填 VIP/LB 地址 |
| `pod_network_cidr` | 建议 | Calico 用 `192.168.0.0/16`，Flannel 用 `10.244.0.0/16` |
| `k8s_version` | 可选 | 默认 `1.29` |
| `image_repository` | 国内必改 | 留空官方；国内填 `registry.aliyuncs.com/google_containers` |

### 3. 一键部署

```bash
# 完整部署（初始化 + 加入所有节点）
ansible-playbook -i inventory/hosts.ini site.yml

# 只初始化集群（不加节点）
# （site.yml 已含全部，一般不用单独跑）
```

### 4. 验证

```bash
# 在 master 上
kubectl get nodes
kubectl get pods -A
```

## 增量加节点

**加 worker**：在 `hosts.ini` 的 `[workers]` 组新增一行，然后：

```bash
ansible-playbook -i inventory/hosts.ini join-node.yml
```

**加 master（高可用）**：在 `[masters]` 组新增一行，然后同样执行 `join-node.yml`。

> 注意：多 master 时 `control_plane_endpoint` 必须填一个 VIP/LB（如 haproxy/keepalived），否则控制面无法高可用。

## 重置集群

```bash
ansible-playbook -i inventory/hosts.ini reset-cluster.yml
```

## 常见问题

### token 过期
worker 加入失败提示 token 过期时，在 master 手动重新生成：
```bash
kubeadm token create --print-join-command
# 生成的命令复制到目标节点手动执行
```

### 证书密钥过期（加 master）
`certificate-key` 默认 2 小时过期，加 master 时重新生成：
```bash
kubeadm init phase upload-certs --upload-certs
```

### 镜像拉取慢（国内）
把 `image_repository` 改为 `registry.aliyuncs.com/google_containers`，并在所有节点预拉取：
```bash
kubeadm config images pull --image-repository=registry.aliyuncs.com/google_containers
```

### 换容器运行时为 docker
1. `all.yml` 里 `container_runtime: docker`、`cri_socket: /var/run/cri-dockerd.sock`
2. runtime role 里把 containerd.io 换成 docker-ce + cri-dockerd

## 版本说明

- 默认：Kubernetes 1.29 + containerd + Calico
- 兼容：Ubuntu 22.04/24.04、Rocky/CentOS 9（自动检测包管理器）
- 支持：单 master 或 多 master 高可用（`upload_certs: true`）
