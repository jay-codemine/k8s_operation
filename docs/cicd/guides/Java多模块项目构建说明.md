# Java 多模块 Maven 项目构建说明

> **适用模板**: `configs/jenkins-templates/java-spring-pipeline.groovy`  
> **更新日期**: 2026-07-03  
> **适用版本**: v15.6+

---

## 一、多模块概念

Maven 多模块项目通过一个**父 POM** 聚合多个**子模块**，实现依赖统一管理、构建顺序编排：

```
myapp/                          ← Git 仓库根目录
├── pom.xml                     ← 父 POM（聚合模块）
├── common/                     ← 子模块：公共库
│   └── pom.xml
├── services/
│   ├── user-service/           ← 子模块：用户服务（可运行）
│   │   └── pom.xml
│   └── order-service/          ← 子模块：订单服务（可运行）
│       └── pom.xml
└── gateway/                    ← 子模块：网关（可运行）
    └── pom.xml
```

### Maven 父子关系

```
父 POM                          子 POM
┌──────────────────────┐        ┌──────────────────────┐
│ <packaging>pom</>    │  路径   │ <parent>              │
│ <modules>            │──────→│   <artifactId>myapp</>│
│   <module>common</>  │        │ </parent>             │
│   <module>services/  │←──────│                       │
│    user-service</>   │ GAV坐标 │ <artifactId>         │
│ </modules>           │        │   user-service       │
└──────────────────────┘        │ </artifactId>         │
                                └──────────────────────┘
```

- **父 → 子**：通过 `<modules>` 声明子模块目录路径
- **子 → 父**：通过 `<parent>` 声明父 POM 的 Maven 坐标（groupId + artifactId + version）

---

## 二、模板自动检测逻辑

模板在 **Setup & Compile 阶段** 执行以下检测流程：

```
BUILD_DIR 指定了？
  ├─ 是 → 校验 BUILD_DIR/pom.xml 存在 → 用指定路径
  └─ 否 → 自动搜索:
        ├─ 根目录有 pom.xml → buildDir = '.'
        └─ 根目录无 pom.xml → 递归 find pom.xml:
              ├─ 只有 1 个 → 用该目录
              ├─ 多个 pom.xml → 优先含 spring-boot-maven-plugin 的
              └─ 兜底 → 取最浅层级
                      ↓
              确定 buildDir 后，检查根 pom.xml 是否含 <modules>
                ├─ 有 <modules> → 多模块 → mvn package -pl <buildDir> -am
                └─ 无 <modules> → 单/独立模块 → mvn package -f <buildDir>/pom.xml
```

### 关键参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `BUILD_DIR` | 空（自动检测） | pom.xml 所在路径，如 `services/user-service/` |
| `MAVEN_GOALS` | `clean package -DskipTests -B` | Maven 构建命令（模板自动拼接 `-pl`/`-am`/`-f` 等平台参数，不要在其中添加这些参数） |
| `MAVEN_THREADS` | `1C` | Maven 并行构建线程数（`1C`=每核1线程, `2C`=每核2线程, `4`=固定4线程） |

---

## 三、BUILD_DIR 留空 vs 指定

### 场景 1：留空（自动检测）

| 项目结构 | 自动检测结果 | Maven 命令 |
|----------|-------------|-----------|
| 根目录有 pom.xml | buildDir = `.` | `mvn package` |
| 根目录无 pom，只有一个子 pom | 用该子目录 | `mvn package -f <dir>/pom.xml` |
| 根目录无 pom，多个子 pom | 优先 spring-boot 模块 | `mvn package -f <dir>/pom.xml` |

**注意**：留空时如果根目录就有 pom.xml（含 `<modules>`），模板会在根目录直接 `mvn package`，由 Maven Reactor 自己按 `<modules>` 顺序逐个构建所有子模块——构建全部模块，不是只构建一个。

### 场景 2：指定 BUILD_DIR（推荐多模块项目）

```
BUILD_DIR = services/user-service/
```

| 根 pom.xml 含 `<modules>`? | Maven 命令 |
|---|---|
| 是 | `mvn package -pl services/user-service -am` |
| 否 | `mvn package -f services/user-service/pom.xml` |

**`-pl -am` 的含义**：

| 参数 | 全称 | 作用 |
|------|------|------|
| `-pl` | `--projects` | 只构建指定的子模块 |
| `-am` | `--also-make` | 同时构建该模块依赖的所有兄弟模块 |

```
mvn package -pl services/user-service -am

Maven Reactor 构建顺序：
  common (被 user-service 依赖)  →  services/user-service (目标模块)
                                  ↑ -am 自动拉入
```

### 场景对照

项目结构：`myapp`（父 pom + common + user-service + order-service）

| 配置 | 构建哪些模块 | 适用场景 |
|------|------------|---------|
| 留空 | 全部模块（父 pom Reactor 构建） | 全量构建 |
| `BUILD_DIR=services/user-service/` | common + user-service | **只部署一个服务（推荐）** |
| `BUILD_DIR=services/order-service/` | common + order-service | 只部署订单服务 |

---

## 四、必须遵从的 Maven 规范

### 1. 父 POM

```xml
<!-- 根目录 pom.xml -->
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>com.example</groupId>
    <artifactId>myapp-parent</artifactId>
    <version>1.0.0</version>
    <packaging>pom</packaging>        <!-- ① 必须是 pom，不能是 jar -->

    <modules>                         <!-- ② 模板靠这个识别多模块 -->
        <module>common</module>
        <module>services/user-service</module>
        <module>services/order-service</module>
    </modules>

    <!-- ③ 公共依赖版本管理（可选但推荐） -->
    <dependencyManagement>
        <dependencies>
            <dependency>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-dependencies</artifactId>
                <version>3.2.0</version>
                <type>pom</type>
                <scope>import</scope>
            </dependency>
        </dependencies>
    </dependencyManagement>
</project>
```

### 2. 子模块 POM（可运行模块）

```xml
<!-- services/user-service/pom.xml -->
<project>
    <modelVersion>4.0.0</modelVersion>

    <parent>                           <!-- ④ 通过 GAV 指向父 POM -->
        <groupId>com.example</groupId>
        <artifactId>myapp-parent</artifactId>
        <version>1.0.0</version>
        <relativePath>../../pom.xml</relativePath>
    </parent>

    <artifactId>user-service</artifactId>
    <!-- 继承父 POM 的 groupId 和 version，不需要重复声明 -->

    <dependencies>
        <dependency>
            <groupId>com.example</groupId>
            <artifactId>common</artifactId>  <!-- 兄弟模块依赖 -->
        </dependency>
    </dependencies>

    <build>
        <plugins>
            <plugin>                     <!-- ⑤ 可运行模块必须有这个插件 -->
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-maven-plugin</artifactId>
            </plugin>
        </plugins>
    </build>
</project>
```

### 3. 子模块 POM（库模块）

```xml
<!-- common/pom.xml -->
<project>
    <modelVersion>4.0.0</modelVersion>

    <parent>
        <groupId>com.example</groupId>
        <artifactId>myapp-parent</artifactId>
        <version>1.0.0</version>
    </parent>

    <artifactId>common</artifactId>
    <!-- 不需要 spring-boot-maven-plugin，因为这不是可运行模块 -->
</project>
```

---

## 五、常见问题

### Q1: 我的根目录没有 pom.xml，只有子目录有，能构建吗？

**可以**。模板会 `find . -name pom.xml` 递归搜索。同样支持多模块，前提是有某个 pom.xml 声明了 `<modules>`。

### Q2: 我有多个可运行模块（user-service、order-service），怎么分别部署？

为每个可运行模块创建**独立流水线**，分别设置 `BUILD_DIR`：

| 流水线 | BUILD_DIR |
|--------|-----------|
| 用户服务 | `services/user-service/` |
| 订单服务 | `services/order-service/` |

模板会用 `-pl` 指定目标模块，只构建它和它的依赖。

### Q3: 模板怎么区分父模块和子模块？

模板只做一件事：`grep '<modules>'` 根目录的 `pom.xml`。有 `<modules>` 就是父（多模块项目），没有就是单模块。**不需要逐个分析每个 pom.xml 的内容**。

### Q4: Gradle 多模块支持吗？

当前模板只支持 Maven。Gradle 项目请使用 `language_type=custom` + 自定义 Jenkinsfile。

### Q5: 多模块项目构建失败，提示找不到兄弟模块依赖？

检查两点：
1. 父 pom.xml 的 `<modules>` 是否包含所有子模块（路径正确）
2. 子模块的 `<parent>` 指向是否正确（`<relativePath>` 路径）

---

## 六、构建流程完整示例

以 `services/user-service` 为例，假定根 pom.xml 含 `<modules>`：

```
Jenkins 构建启动
  │
  ├─ Checkout: git clone + 切换分支
  │
  ├─ Setup & Compile:
  │   ① 检测根 pom.xml 有 <modules> → isMultiModule = true
  │   ② mvn package -pl services/user-service -am -DskipTests -B -T 1C
  │   ③ Maven Reactor 构建顺序: common → user-service
  │   ④ 查找 JAR: services/user-service/target/*.jar
  │
  ├─ Test（可选）:
  │   mvn test -pl services/user-service -am -B -T 1C
  │
  ├─ Build & Push Image:
  │   自动生成 Dockerfile → Kaniko 构建 → 推送镜像仓库
  │
  └─ Callback: 通知平台构建结果
```
