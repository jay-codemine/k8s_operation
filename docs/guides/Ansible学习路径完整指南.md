# Ansible 学习路径完整指南

> 从零基础到生产级自动化，配套 `scripts/ansible-k8s/` 中的 Kubernetes 集群部署实战案例。

## 目录

- [第一部分：初阶 —— 入门基础](#第一部分初阶--入门基础)
- [第二部分：中阶 —— 进阶技巧](#第二部分中阶--进阶技巧)
- [第三部分：高阶 —— 生产实践](#第三部分高阶--生产实践)
- [第四部分：实战 —— Kubernetes 集群部署](#第四部分实战--kubernetes-集群部署)
- [第五部分：学习路线图](#第五部分学习路线图)

---

## 第一部分：初阶 —— 入门基础

### 1.1 Ansible 是什么

Ansible 是一个**无 Agent 的自动化运维工具**，通过 SSH 连接目标机器执行任务。核心特点：

| 特性 | 说明 |
|---|---|
| 无 Agent | 目标机器无需安装客户端，只需 SSH + Python |
| 声明式 | 描述"期望状态"，而非"执行步骤" |
| 幂等 | 重复执行不会产生副作用（大部分模块天然幂等） |
| 易读 | 用 YAML 编写，接近自然语言 |

**对比其他工具**：

| 工具 | Agent | 语言 | 适用 |
|---|---|---|---|
| Ansible | 无 | YAML | 配置管理、应用部署、编排 |
| Puppet/Chef | 有 | DSL/Ruby | 大型长期配置管理 |
| SaltStack | 有 | YAML/Python | 大规模实时执行 |
| Terraform | 无 | HCL | 基础设施（云资源）编排 |

### 1.2 架构原理

```
┌─────────────────┐         SSH          ┌─────────────────┐
│   控制机         │  ───────────────▶   │   目标节点 1     │
│  (Control Node) │         SSH          │   (Managed Node)│
│  ansible 安装处  │  ───────────────▶   ├─────────────────┤
│  playbook/inv   │         SSH          │   目标节点 2     │
└─────────────────┘  ───────────────▶   └─────────────────┘
```

- **控制机**：安装 Ansible 的机器（你的笔记本/跳板机）
- **目标节点**：被管理的服务器，只需 SSH 可达
- **执行方式**：控制机把模块（Python 代码）通过 SSH 推到目标节点执行，拿回结果

### 1.3 安装 Ansible

```bash
# 推荐用 pip（版本最新）
pip install ansible

# 或系统包管理器
apt install ansible          # Ubuntu
dnf install ansible          # Rocky/CentOS
brew install ansible         # macOS

# 验证
ansible --version
```

### 1.4 第一个 ad-hoc 命令

ad-hoc 是"一次性命令"，不用写文件，适合快速执行：

```bash
# 语法：ansible <目标> -m <模块> -a "<参数>"

# ping 所有主机（测试连通性）
ansible all -i inventory/hosts.ini -m ping

# 看所有主机的内核版本
ansible all -i inventory/hosts.ini -m command -a "uname -r"

# 复制文件到目标机
ansible webservers -m copy -a "src=/tmp/a.txt dest=/tmp/a.txt"

# 安装软件包
ansible all -m apt -a "name=nginx state=present" --become
```

> `-i` 指定 inventory（主机清单），`--become`（或 `-b`）表示提权到 root。

### 1.5 核心概念

#### 1.5.1 Inventory（主机清单）

定义"管理哪些机器、如何分组"。支持 INI 和 YAML 两种格式：

```ini
# inventory/hosts.ini
[all:vars]                        # 所有主机共享变量
ansible_user=root

[masters]                         # master 分组
k8s-master-1 ansible_host=192.168.1.101

[workers]                         # worker 分组
k8s-worker-1 ansible_host=192.168.1.102

[k8s_cluster:children]            # 组合分组
masters
workers
```

关键点：
- `[组名]` 定义分组，`[组名:children]` 定义组嵌套
- `ansible_host` 指定实际连接 IP，主机名可自定义别名
- `[all:vars]` 是全局变量段

#### 1.5.2 Module（模块）

模块是 Ansible 的最小执行单元。常用模块：

| 模块 | 用途 | 示例 |
|---|---|---|
| `ping` | 连通性测试 | `-m ping` |
| `command`/`shell` | 执行命令 | `-a "uptime"` |
| `copy` | 复制文件 | `src= dest=` |
| `file` | 文件/目录操作 | `path= state=directory` |
| `apt`/`dnf`/`yum` | 包管理 | `name= state=` |
| `systemd`/`service` | 服务管理 | `name= state=started` |
| `lineinfile`/`replace` | 行级/正则文件编辑 | |
| `template` | 渲染 Jinja2 模板 | |
| `get_url` | 下载文件 | `url= dest=` |
| `sysctl` | 内核参数 | |

**模块 vs 命令**：优先用模块（幂等、可读、有返回值），`command`/`shell` 是"最后手段"。

#### 1.5.3 Task / Play / Playbook

```
Playbook (YAML 文件)
└── Play (一个 hosts 段，针对一组主机)
    └── Task (一个具体动作)
        └── Module (执行单元)
```

```yaml
# 一个最小的 playbook
- name: 安装 nginx          # Play 名称
  hosts: webservers          # 针对哪些主机
  become: true               # 是否提权
  tasks:                     # 任务列表
    - name: 安装 nginx       # Task 名称
      apt:
        name: nginx
        state: present
```

#### 1.5.4 变量

```yaml
# 定义在 playbook 里
vars:
  package_name: nginx

# 引用
- name: 安装 {{ package_name }}
  apt:
    name: "{{ package_name }}"
    state: present
```

变量来源（优先级从低到高）：inventory → play vars → role defaults → role vars → host_vars/group_vars → 命令行 `-e`。

#### 1.5.5 Facts（系统信息）

Ansible 自动收集目标机的系统信息（IP、OS、CPU、内存等），存为变量：

```yaml
- name: 打印操作系统
  debug:
    msg: "OS = {{ ansible_os_family }}, 版本 = {{ ansible_distribution_version }}"
```

常用的 fact 变量：
- `ansible_os_family`：`Debian` / `RedHat`（用于写跨系统分支）
- `ansible_distribution_release`：Ubuntu 的代号（如 `jammy`）
- `ansible_hostname`、`ansible_default_ipv4.address`

#### 1.5.6 Handler（触发器）

Handler 是"被通知才执行"的特殊任务，常用于"配置变了才重启服务"：

```yaml
tasks:
  - name: 修改配置
    copy:
      src: nginx.conf
      dest: /etc/nginx/nginx.conf
    notify: restart nginx          # 通知 handler

handlers:
  - name: restart nginx            # 收到通知才执行
    systemd:
      name: nginx
      state: restarted
```

> 核心价值：多个 task 通知同一个 handler，handler **只执行一次**（在 play 结束时）。

### 1.6 第一个 Playbook

```yaml
# my-first.yml
---
- name: 配置 Web 服务器
  hosts: webservers
  become: true
  vars:
    web_port: 8080

  tasks:
    - name: 安装 nginx
      apt:
        name: nginx
        state: present
        update_cache: true

    - name: 确保监听端口正确
      lineinfile:
        path: /etc/nginx/sites-available/default
        regexp: '^listen '
        line: 'listen {{ web_port }};'
      notify: restart nginx

    - name: 启动服务
      systemd:
        name: nginx
        state: started
        enabled: true

  handlers:
    - name: restart nginx
      systemd:
        name: nginx
        state: restarted
```

执行：

```bash
ansible-playbook -i inventory/hosts.ini my-first.yml

# 检查语法（不执行）
ansible-playbook my-first.yml --syntax-check

# 模拟执行（dry-run，看会改什么）
ansible-playbook my-first.yml --check
```

---

## 第二部分：中阶 —— 进阶技巧

### 2.1 变量体系与优先级

完整优先级（高→低）：

```
1. 命令行 -e 参数            (ansible-playbook -e "key=val")
2. play 的 vars 和 vars_files
3. role 的 vars
4. host_vars / group_vars（文件或目录）
5. inventory 里的 [all:vars]
6. role 的 defaults
```

**最佳实践**：
- 可变、需用户修改的 → `role/defaults`
- 固定、内部使用的 → `role/vars`
- 环境相关 → `group_vars` / `host_vars`

```yaml
# group_vars/all.yml（我们的 K8s 项目就是这么做的）
k8s_version: "1.29"
pod_network_cidr: "192.168.0.0/16"
```

### 2.2 条件与循环

#### 条件 when

```yaml
- name: 用 apt 装（Debian 系）
  apt:
    name: containerd.io
  when: ansible_os_family == "Debian"

- name: 用 dnf 装（RedHat 系）
  dnf:
    name: containerd.io
  when: ansible_os_family == "RedHat"
```

复合条件：

```yaml
when:
  - cni_plugin == "calico"
  - pod_network_cidr != "192.168.0.0/16"
```

#### 循环 loop

```yaml
- name: 加载多个内核模块
  modprobe:
    name: "{{ item }}"
  loop:
    - overlay
    - br_netfilter
```

字典循环（我们的 K8s sysctl 就是这么用的）：

```yaml
- name: 设置内核参数
  sysctl:
    name: "{{ item.name }}"
    value: "{{ item.value }}"
  loop:
    - { name: 'net.ipv4.ip_forward', value: '1' }
    - { name: 'net.bridge.bridge-nf-call-iptables', value: '1' }
```

### 2.3 Jinja2 模板

`template` 模块用 Jinja2 渲染，可以动态生成配置文件：

```yaml
- name: 渲染 kubelet 配置
  template:
    src: kubelet.conf.j2
    dest: /etc/sysconfig/kubelet
```

模板文件 `kubelet.conf.j2`：

```jinja2
KUBELET_EXTRA_ARGS="{{ kubelet_extra_args }}"
{% if node_ip %}NODE_IP={{ node_ip }}{% endif %}
```

Jinja2 常用语法：
- `{{ 变量 }}` 输出
- `{% if %}...{% endif %}` 条件
- `{% for %}...{% endfor %}` 循环
- 过滤器：`{{ var | default('fallback') }}`、`{{ list | join(',') }}`

### 2.4 Role 组织

当 playbook 变复杂，用 role 把相关任务、变量、handler、模板打包：

```
roles/
└── runtime/                 # 角色名
    ├── tasks/main.yml       # 入口任务
    ├── handlers/main.yml    # 触发器
    ├── defaults/main.yml    # 默认变量
    ├── vars/main.yml        # 固定变量
    ├── templates/           # 模板
    └── files/               # 静态文件
```

引用 role：

```yaml
- hosts: all
  roles:
    - common
    - runtime
    - kubernetes
```

### 2.5 幂等性设计

**幂等** = 重复执行结果不变。这是 Ansible 的灵魂。

- 用 `copy` 而非 `command: echo ... > file`（copy 会比对内容）
- 用 `file state=directory` 而非 `mkdir`
- 用 `package state=present` 而非 `apt-get install`
- 命令类模块用 `creates` 参数标记"已存在则跳过"

```yaml
# 好的做法：creates 标记，已存在则跳过
- name: 下载并转换 GPG key
  shell: curl ... | gpg --dearmor -o /etc/apt/keyrings/k8s.gpg
  args:
    creates: /etc/apt/keyrings/k8s.gpg   # 文件已存在则跳过
```

### 2.6 Tags（标签）

给任务打标签，可选择性执行：

```yaml
tasks:
  - name: 安装运行时
    apt: name=containerd.io
    tags: [runtime, install]

  - name: 配置内核
    sysctl: ...
    tags: [config]
```

```bash
# 只执行带 runtime 标签的任务
ansible-playbook site.yml --tags runtime

# 跳过带 config 标签的任务
ansible-playbook site.yml --skip-tags config
```

### 2.7 Vault 加密

敏感信息（密码、token）用 vault 加密：

```bash
# 创建加密文件
ansible-vault create secrets.yml
ansible-vault encrypt group_vars/all/passwords.yml

# 运行时提示输入密码
ansible-playbook site.yml --ask-vault-pass

# 或从文件读密码
ansible-playbook site.yml --vault-password-file .vault_pass
```

---

## 第三部分：高阶 —— 生产实践

### 3.1 多主机编排

Play 按 `hosts` 分阶段执行，不同阶段可以针对不同主机组：

```yaml
- name: 先初始化所有节点
  hosts: k8s_cluster
  roles: [common, runtime, kubernetes]

- name: 再初始化第一个 master
  hosts: masters
  roles:
    - { role: master-init, when: inventory_hostname == groups['masters'][0] }

- name: 最后加入 worker
  hosts: workers
  roles: [join-worker]
```

> 关键：play 之间是**串行**的，前面的 play 完成才进入下一个。这正是我们 site.yml 的顺序逻辑。

### 3.2 delegate_to / run_once / hostvars

这三个是跨主机协作的核心：

```yaml
# delegate_to：任务在别的机器上执行（默认在当前主机执行）
- name: 在 master 上生成 token
  command: kubeadm token create --print-join-command
  delegate_to: "{{ groups['masters'][0] }}"
  register: kubeadm_join

# run_once：整个 play 只执行一次（即使有 10 台主机）
- name: 只初始化一次
  command: kubeadm init ...
  run_once: true

# hostvars：访问其他主机的变量（跨主机取 token）
- name: 用 master 上的 token 加入集群
  command: "{{ hostvars[groups['masters'][0]]['kubeadm_join'].stdout }}"
```

**这是加节点的核心机制**（我们的 join-worker 就是这么做的）：
1. `delegate_to` master 执行 `kubeadm token create`，结果注册到变量
2. worker 通过 `hostvars[master]` 拿到 token
3. worker 执行 join 命令

### 3.3 动态 Inventory

静态 inventory 适合固定集群，云环境用动态 inventory 脚本自动发现主机：

```bash
# 使用动态 inventory 脚本
ansible-playbook -i aws_ec2.yml site.yml

# 或 inventory 目录里放多个源
inventory/
├── hosts.ini          # 静态
└── dynamic/           # 动态脚本
```

常见动态 inventory 插件：`aws_ec2`、`gcp_compute`、`azure_rm`、`vmware_vm_inventory`。

### 3.4 自定义 Module

当现有 3000+ 模块不够用，用 Python 写自定义模块：

```python
# library/k8s_join.py
#!/usr/bin/python
from ansible.module_utils.basic import AnsibleModule

def main():
    module = AnsibleModule(argument_spec=dict(
        token=dict(required=True, no_log=True),
    ))
    # 业务逻辑...
    module.exit_json(changed=True, result="joined")

if __name__ == '__main__':
    main()
```

放在 `library/` 目录即可被 playbook 使用。

### 3.5 性能优化

| 优化项 | 配置 | 说明 |
|---|---|---|
| 并发数 | `forks = 50` | 默认 5，调大并行执行 |
| 关闭 facts | `gather_facts: false` | 不需要 facts 时关掉提速 |
| SSH 长连接 | `pipelining = true` | 减少 SSH 连接数 |
| 异步任务 | `async: 3600` + `poll: 0` | 长任务不阻塞 |
| 策略插件 | `strategy: free` | 主机间不等待（默认 linear） |

`ansible.cfg` 示例：

```ini
[defaults]
forks = 50
pipelining = true
host_key_checking = false
```

---

## 第四部分：实战 —— Kubernetes 集群部署

> 完整代码在 `scripts/ansible-k8s/`，本节讲解它如何落地前面所有知识点。

### 4.1 需求分析

目标：**从裸机到可用 K8s 集群**，支持加 worker / 加 master（高可用）。

拆解需求 → 映射 Ansible 概念：

| 需求 | Ansible 实现 |
|---|---|
| 管理多台机器 | Inventory（masters/workers 分组） |
| 前置准备（swap/内核/防火墙） | `common` role |
| 装容器运行时 | `runtime` role（containerd） |
| 装 kubeadm 三件套 | `kubernetes` role |
| 初始化第一台 master | `master-init` role + `run_once` 语义 |
| 装网络插件 | `network` role |
| 加 worker | `join-worker` role + `delegate_to` 取 token |
| 加 master（HA） | `join-master` role + certificate-key |

### 4.2 目录结构解析

```
ansible-k8s/
├── inventory/
│   ├── hosts.ini              # 主机分组（改这里加节点）
│   └── group_vars/all.yml     # 全局变量（改这里调参数）
├── roles/                     # 6 个角色，各司其职
│   ├── common/                # 所有节点通用前置
│   ├── runtime/               # 容器运行时
│   ├── kubernetes/            # kubeadm/kubelet/kubectl
│   ├── master-init/           # 初始化控制面
│   ├── network/               # CNI
│   ├── join-worker/           # worker 加入
│   └── join-master/           # master 加入
├── site.yml                   # 一键部署
├── join-node.yml              # 增量加节点
└── reset-cluster.yml          # 重置
```

### 4.3 六个 Role 逐一详解

#### common（前置准备）

```yaml
- name: 关闭 swap（当前生效）
  command: swapoff -a
  changed_when: false
  ignore_errors: true

- name: 永久关闭 swap（Ubuntu）
  replace:
    path: /etc/fstab
    regexp: '^([^#].*?\sswap\s+sw\s+.*)$'
    replace: '# \1'
  when: ansible_os_family == "Debian"
```

**知识点**：
- `command` + `changed_when: false`：命令本身非幂等，声明"不算变更"
- `replace` 正则注释 swap 行：幂等（已注释则不再匹配）
- `when: ansible_os_family == "Debian"`：跨系统分支

#### runtime（containerd）

```yaml
- name: 生成 containerd 默认配置
  command: containerd config default
  register: containerd_default_config
  changed_when: false

- name: 写入配置（启用 systemd cgroup）
  copy:
    content: "{{ containerd_default_config.stdout | regex_replace('SystemdCgroup = false', 'SystemdCgroup = true') }}"
    dest: /etc/containerd/config.toml
  notify: restart containerd
```

**知识点**：
- `register` 捕获命令输出到变量
- `regex_replace` Jinja2 过滤器做字符串替换
- `notify` 触发 handler 重启 containerd

#### kubernetes（装 kubeadm）

```yaml
- name: 添加 k8s APT 仓库（Ubuntu）
  apt_repository:
    repo: "deb [signed-by=...] https://pkgs.k8s.io/core:/stable:/{{ k8s_version }}/deb/ /"
    filename: kubernetes
    state: present
  when: ansible_os_family == "Debian"
```

**知识点**：`{{ k8s_version }}` 变量引用让仓库地址可配，改版本只需改变量。

#### master-init（初始化控制面）

```yaml
- name: 执行 kubeadm init
  command: >
    kubeadm init
    --control-plane-endpoint={{ control_plane_endpoint }}
    --pod-network-cidr={{ pod_network_cidr }}
    --cri-socket={{ cri_socket }}
    {{ '--upload-certs' if upload_certs else '' }}
```

**知识点**：YAML 折行符 `>` 拼长命令；Jinja2 三元表达式拼可选参数。

#### join-worker（加 worker —— 核心机制）

```yaml
- name: 从 master 生成 worker 加入命令
  command: kubeadm token create --print-join-command
  delegate_to: "{{ groups['masters'][0] }}"
  register: kubeadm_join
  run_once: true
  changed_when: false

- name: 加入 worker 节点
  command: "{{ hostvars[groups['masters'][0]]['kubeadm_join'].stdout }} --cri-socket={{ cri_socket }}"
```

**知识点（加节点最关键的三步）**：
1. `delegate_to` 把 token 生成任务派到第一台 master 执行
2. `register` 保存 token 命令到变量
3. `hostvars[master]` 跨主机读取 token，worker 本机执行 join

#### join-master（加 master —— 高可用）

```yaml
- name: 获取证书密钥
  command: kubeadm init phase upload-certs --upload-certs
  delegate_to: "{{ groups['masters'][0] }}"
  register: cert_key_result

- name: 加入控制平面节点
  command: >
    {{ join_cmd.stdout }}
    --control-plane
    --certificate-key {{ cert_key_result.stdout_lines[-1] }}
```

**知识点**：`stdout_lines[-1]` 取输出最后一行（certificate key）。

### 4.4 执行流程（site.yml 的顺序）

```
Play 1: 所有节点  → common + runtime + kubernetes   （并行）
Play 2: 第一台 master → master-init（kubeadm init）
Play 3: 第一台 master → network（装 Calico）
Play 4: 所有 worker → join-worker（并行加入）
Play 5: 其余 master → join-master（高可用）
```

**为什么这个顺序**：
1. 先装好所有节点的环境（运行时 + kubeadm）
2. 再 init 第一台 master（此时才有控制面）
3. 装 CNI 让 master 变 Ready
4. worker 才能加入（依赖 master 的 token）
5. 最后加额外 master

### 4.5 完整操作手册

```bash
# 1. 改 inventory/hosts.ini 的 IP
vim inventory/hosts.ini

# 2. 改 group_vars/all.yml 的关键变量
#    - control_plane_endpoint（必改）
#    - image_repository（国内必改）
vim inventory/group_vars/all.yml

# 3. 检查语法
ansible-playbook -i inventory/hosts.ini site.yml --syntax-check

# 4. 一键部署（或先 --check 模拟）
ansible-playbook -i inventory/hosts.ini site.yml

# 5. 验证
ansible masters -i inventory/hosts.ini -m shell -a "kubectl get nodes"
```

### 4.6 加 worker 节点

```bash
# 1. 在 hosts.ini 的 [workers] 组加一行
#    k8s-worker-3 ansible_host=192.168.1.104
vim inventory/hosts.ini

# 2. 增量加节点（幂等，老节点跳过）
ansible-playbook -i inventory/hosts.ini join-node.yml
```

### 4.7 加 master 节点（高可用）

```bash
# 1. 在 [masters] 组加一行
vim inventory/hosts.ini

# 2. 确保 control_plane_endpoint 是 VIP/LB 地址
# 3. 执行
ansible-playbook -i inventory/hosts.ini join-node.yml
```

> 高可用前提：`control_plane_endpoint` 指向 VIP/负载均衡（如 haproxy + keepalived），否则只是"多个 master"，无法故障切换。

### 4.8 常见排错

| 现象 | 原因 | 解决 |
|---|---|---|
| token 过期 | token 默认 24h | master 上 `kubeadm token create --print-join-command` |
| cert-key 过期 | 默认 2h | `kubeadm init phase upload-certs --upload-certs` |
| 镜像拉取慢 | 访问不了 gcr/k8s.gcr | `image_repository=registry.aliyuncs.com/google_containers` |
| cgroup 报错 | containerd 未用 systemd cgroup | 检查 config.toml 的 `SystemdCgroup = true` |
| worker NotReady | CNI 未装或网段不符 | 检查 Calico 与 `pod_network_cidr` 一致 |
| SSH 连接失败 | 没配免密 | 先 `ssh-copy-id root@节点` |

---

## 第五部分：学习路线图

### 5.1 阶段规划

| 阶段 | 周期 | 目标 | 关键内容 |
|---|---|---|---|
| **初阶** | 1-2 周 | 会用 ad-hoc + 简单 playbook | inventory、module、task、变量、facts |
| **中阶** | 2-4 周 | 写规范的 role + 幂等 | when/loop、template、handler、role、vault |
| **高阶** | 1-2 月 | 生产级自动化 | delegate_to、hostvars、动态 inventory、自定义 module、性能 |
| **实战** | 持续 | 落地项目 | 用本项目 K8s 部署案例反复练习 |

### 5.2 学习建议

1. **先跑起来**：把 `scripts/ansible-k8s/` 的 playbook 部署成功一次，再逐行理解
2. **看 module 文档**：`ansible-doc <模块名>` 是最好用的参考
3. **改参数验证**：改 `group_vars/all.yml` 的变量，观察行为变化
4. **拆解重构**：把 site.yml 拆成自己的 role，加深理解
5. **读官方最佳实践**：Ansible 官方 docs 的 "Playbook Best Practices"

### 5.3 常用命令速查

```bash
# 检查语法
ansible-playbook site.yml --syntax-check

# 列出所有主机
ansible-playbook site.yml --list-hosts

# 列出所有任务
ansible-playbook site.yml --list-tasks

# 模拟执行
ansible-playbook site.yml --check

# 只跑某台主机
ansible-playbook site.yml --limit k8s-worker-1

# 查看模块用法
ansible-doc systemd
ansible-doc -l | grep k8s
```

### 5.4 推荐资源

- **官方文档**：docs.ansible.com（最权威）
- **Ansible Galaxy**：galaxy.ansible.com（现成的 role 参考）
- **最佳实践**：官方 "Playbook Best Practices" 章节
- **本项目案例**：`scripts/ansible-k8s/`（从 0 到 1 的完整落地）

---

## 附录：与项目其他脚本的配合

| 脚本 | 作用 | 与本 playbook 关系 |
|---|---|---|
| `scripts/ansible-k8s/` | 部署 K8s 集群 | 本文主角 |
| `scripts/deploy-k8s.sh` | 部署平台应用 | 集群就绪后部署平台 |
| `scripts/quick-start.sh` | 快速启动平台 | 集群已就绪时用 |
| `deploy/` | 平台各组件 YAML | 集群就绪后 kubectl apply |

**完整链路**：Ansible 建集群 → `deploy/` 部署平台 → 平台管理多集群。
