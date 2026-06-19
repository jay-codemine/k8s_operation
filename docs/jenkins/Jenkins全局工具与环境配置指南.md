# Jenkins 全局工具与环境配置指南

> 适用于 K8s Operation 平台的 Jenkins 构建服务器配置

---

## 一、概述

平台的 Java 构建模板（`java-spring-pipeline.groovy`）通过 Jenkins 全局工具自动获取 JDK 和 Maven 路径，**无需手动配置系统环境变量**，也**无需切换 jenkins 用户**执行任何操作。

### 核心原则

| 项目 | 是否需要 | 说明 |
|------|---------|------|
| Jenkins 全局工具注册（5 个路径） | ✅ 必须 | 唯一需要配置的地方 |
| 系统环境变量（JAVA_HOME 等） | ❌ 不需要 | 模板运行时自动注入 |
| `su - jenkins` 手动执行 | ❌ 不需要 | Jenkins 服务本身就是 jenkins 用户 |
| `.bashrc` / `.profile` 配置 | ❌ 不需要 | Pipeline 不加载 login shell |

---

## 二、Jenkins 全局工具配置（5 个路径）

### 操作入口

```
Jenkins 首页 → Manage Jenkins → Tools
```

### 2.1 JDK installations（添加 4 个）

> ⚠️ 取消勾选 "Install automatically"，手动填写路径

| Name（必须与模板一致） | JAVA_HOME 路径 |
|----------------------|----------------|
| `JDK-8` | `/usr/lib/jvm/java-8-openjdk-amd64` |
| `JDK-11` | `/usr/lib/jvm/java-11-openjdk-amd64` |
| `JDK-17` | `/usr/lib/jvm/java-17-openjdk-amd64` |
| `JDK-21` | `/usr/lib/jvm/java-21-openjdk-amd64` |

### 2.2 Maven installations（添加 1 个）

| Name（必须与模板一致） | MAVEN_HOME 路径 |
|----------------------|-----------------|
| `Maven-3.9` | `/opt/maven/apache-maven-3.9.16` |

### 2.3 配置截图参考

```
JDK installations:
  ┌─────────────────────────────────────────────────┐
  │ Name:      JDK-8                                │
  │ JAVA_HOME: /usr/lib/jvm/java-8-openjdk-amd64   │
  │ □ Install automatically（取消勾选）              │
  └─────────────────────────────────────────────────┘
  ┌─────────────────────────────────────────────────┐
  │ Name:      JDK-11                               │
  │ JAVA_HOME: /usr/lib/jvm/java-11-openjdk-amd64  │
  └─────────────────────────────────────────────────┘
  ┌─────────────────────────────────────────────────┐
  │ Name:      JDK-17                               │
  │ JAVA_HOME: /usr/lib/jvm/java-17-openjdk-amd64  │
  └─────────────────────────────────────────────────┘
  ┌─────────────────────────────────────────────────┐
  │ Name:      JDK-21                               │
  │ JAVA_HOME: /usr/lib/jvm/java-21-openjdk-amd64  │
  └─────────────────────────────────────────────────┘

Maven installations:
  ┌─────────────────────────────────────────────────┐
  │ Name:       Maven-3.9                           │
  │ MAVEN_HOME: /opt/maven/apache-maven-3.9.16     │
  │ □ Install automatically（取消勾选）              │
  └─────────────────────────────────────────────────┘
```

---

## 三、为什么不需要配置系统环境变量

### 3.1 模板自动处理

Pipeline 模板在 `Setup Build Tools` 阶段会根据平台传入的 `JAVA_VERSION` 参数动态设置环境：

```groovy
// 模板自动执行，不依赖系统环境变量
env.JAVA_HOME = '/usr/lib/jvm/java-17-openjdk-amd64'
env.PATH = "${jdkHome}/bin:${env.PATH}"
```

### 3.2 Jenkins Tools 机制

Jenkins `tools` 块声明的工具，在 Pipeline 启动时自动注入到构建环境中：

```groovy
tools {
    maven 'Maven-3.9'   // → 自动设置 MAVEN_HOME 和 PATH
    jdk   'JDK-17'      // → 自动设置 JAVA_HOME 和 PATH
}
```

### 3.3 结论

```
系统层面: 无需 export JAVA_HOME / MAVEN_HOME / PATH
Jenkins 层面: 只需在 Tools 里注册名称 + 路径
模板层面: 运行时自动切换，支持多版本
```

---

## 四、为什么不需要 `su - jenkins`

### 4.1 Jenkins 服务运行身份

```bash
$ id jenkins
uid=115(jenkins) gid=121(jenkins) groups=121(jenkins)
```

Jenkins 服务进程本身就以 `jenkins` 用户运行。Pipeline 中所有 `sh` 步骤自动继承此身份。

### 4.2 权限验证方法

如需验证 jenkins 用户是否能访问工具路径，使用 `sudo -u`（仅验证用，实际运行不需要）：

```bash
# 验证 Maven 可执行
$ sudo -u jenkins /opt/maven/apache-maven-3.9.16/bin/mvn -version
Apache Maven 3.9.16
Java version: 21.0.11, vendor: Ubuntu
Maven home: /opt/maven/apache-maven-3.9.16

# 验证 JDK 可读
$ sudo -u jenkins ls /usr/lib/jvm/java-17-openjdk-amd64/bin/java
/usr/lib/jvm/java-17-openjdk-amd64/bin/java
```

### 4.3 权限要求

| 路径 | 权限要求 | 默认状态 |
|------|---------|---------|
| `/usr/lib/jvm/java-*` | jenkins 用户可读+可执行 | ✅ 默认满足（755） |
| `/opt/maven/apache-maven-3.9.16` | jenkins 用户可读+可执行 | ✅ 默认满足（755） |
| `/var/lib/jenkins/.m2/repository` | jenkins 用户可读写 | ✅ jenkins 用户 home 目录 |
| `/var/lib/jenkins/.buildkit-cache` | jenkins 用户可读写 | ✅ 需自动创建 |

---

## 五、服务器当前环境确认

### 5.1 已安装的 JDK 版本

```bash
$ ls -l /usr/lib/jvm/
java-8-openjdk-amd64        # JDK 8 实际目录
java-11-openjdk-amd64       # JDK 11 实际目录
java-17-openjdk-amd64       # JDK 17 实际目录
java-21-openjdk-amd64       # JDK 21 实际目录
java-1.8.0-openjdk-amd64  → java-8-openjdk-amd64   # 软链接
java-1.11.0-openjdk-amd64 → java-11-openjdk-amd64  # 软链接
java-1.17.0-openjdk-amd64 → java-17-openjdk-amd64  # 软链接
java-1.21.0-openjdk-amd64 → java-21-openjdk-amd64  # 软链接
```

### 5.2 Maven 路径

```bash
$ which mvn
/opt/maven/apache-maven-3.9.16/bin/mvn

$ mvn -version
Apache Maven 3.9.16
Maven home: /opt/maven/apache-maven-3.9.16
```

---

## 六、模板与 Jenkins Tools 的对应关系

```
┌──────────────────────┐         ┌─────────────────────────────────────┐
│  Jenkins Global Tools │         │  java-spring-pipeline.groovy 模板   │
├──────────────────────┤         ├─────────────────────────────────────┤
│ Name: JDK-17         │◄────────│ tools { jdk 'JDK-17' }             │
│ Path: /usr/lib/jvm/  │         │                                     │
│   java-17-openjdk-   │         │ jdkMap['17'] = 同路径               │
│   amd64              │         │                                     │
├──────────────────────┤         ├─────────────────────────────────────┤
│ Name: Maven-3.9      │◄────────│ tools { maven 'Maven-3.9' }        │
│ Path: /opt/maven/    │         │                                     │
│   apache-maven-3.9.16│         │ mvn 命令自动可用                     │
└──────────────────────┘         └─────────────────────────────────────┘

运行时流程：
1. Pipeline 启动 → Jenkins 加载 tools 块 → 注入 JAVA_HOME/MAVEN_HOME
2. Setup Build Tools 阶段 → 根据 JAVA_VERSION 参数动态覆盖 JAVA_HOME
3. 编译阶段 → 使用正确版本的 JDK + Maven 执行构建
```

---

## 七、常见问题

### Q1: 只配 1 个 JDK 行不行？

**可以。** 必须至少配 `JDK-17`（因为 `tools { jdk 'JDK-17' }` 写死了这个名称）。其他版本是可选的，不存在时模板会自动回退到 JDK-17。

### Q2: Name 写错了会怎样？

Pipeline 会在启动时报错：`Tool type "jdk" does not have an install of "JDK-17" configured`。必须与模板中的名称**完全一致**（大小写敏感）。

### Q3: 路径填错了会怎样？

- `tools` 块中的 JDK-17 路径错 → Pipeline 启动报错
- `jdkMap` 中的其他版本路径错 → 打印 ⚠️ 警告，回退用 JDK-17，不会中断构建

### Q4: 需要重启 Jenkins 吗？

不需要。全局工具配置保存后立即生效，下次构建自动使用新配置。

### Q5: 多个 Jenkins Agent 节点怎么办？

每个 Agent 节点上的 JDK/Maven 路径必须一致，或者在每个节点上安装到相同路径。

---

## 八、配置验证清单

配置完成后，检查以下项目确认无误：

```bash
# 1. Jenkins Tools 已注册（在 Jenkins 页面确认）
Jenkins → Manage Jenkins → Tools
  ✅ JDK-8   → /usr/lib/jvm/java-8-openjdk-amd64
  ✅ JDK-11  → /usr/lib/jvm/java-11-openjdk-amd64
  ✅ JDK-17  → /usr/lib/jvm/java-17-openjdk-amd64
  ✅ JDK-21  → /usr/lib/jvm/java-21-openjdk-amd64
  ✅ Maven-3.9 → /opt/maven/apache-maven-3.9.16

# 2. jenkins 用户权限正常（在服务器执行）
$ sudo -u jenkins /usr/lib/jvm/java-17-openjdk-amd64/bin/java -version
$ sudo -u jenkins /opt/maven/apache-maven-3.9.16/bin/mvn -version

# 3. Maven 缓存目录可写
$ sudo -u jenkins mkdir -p /var/lib/jenkins/.m2/repository
$ sudo -u jenkins ls -ld /var/lib/jenkins/.m2/repository
```

全部验证通过后，Java 项目的 CICD 构建即可正常运行。
