// CI/CD相关API
export const getPipelines = async () => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '获取流水线列表成功',
        data: [
          {
            id: 1,
            name: 'frontend-deploy',
            description: '前端应用部署流水线',
            status: 'running',
            lastRunTime: '2024-05-20T14:30:00Z',
            lastRunStatus: 'success',
            gitRepo: 'https://github.com/example/frontend.git',
            branch: 'main'
          },
          {
            id: 2,
            name: 'backend-deploy',
            description: '后端服务部署流水线',
            status: 'idle',
            lastRunTime: '2024-05-20T13:15:00Z',
            lastRunStatus: 'failed',
            gitRepo: 'https://github.com/example/backend.git',
            branch: 'develop'
          },
          {
            id: 3,
            name: 'database-migration',
            description: '数据库迁移流水线',
            status: 'idle',
            lastRunTime: '2024-05-19T09:45:00Z',
            lastRunStatus: 'success',
            gitRepo: 'https://github.com/example/database.git',
            branch: 'main'
          },
          {
            id: 1001,
            name: 'hello-app-pipeline',
            description: 'Hello应用演示流水线',
            status: 'running',
            lastRunTime: '2024-05-20T14:30:00Z',
            lastRunStatus: 'success',
            gitRepo: 'https://github.com/example/hello-app.git',
            branch: 'main'
          },
          {
            id: 2001,
            name: 'java-app-pipeline',
            description: 'Java/Spring Boot应用部署流水线',
            status: 'running',
            lastRunTime: '2024-05-20T22:30:00Z',
            lastRunStatus: 'success',
            gitRepo: 'https://github.com/example/java-app.git',
            branch: 'main'
          }
        ]
      })
    }, 1000)
  })
}

export const getPipelineDetail = async (id) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 演示Hello应用流水线
      if (id === 1001) {
        resolve({
          code: 0,
          msg: '获取流水线详情成功',
          data: {
            id: id,
            name: 'hello-app-pipeline',
            description: 'Hello应用演示流水线',
            status: 'running',
            gitRepo: 'https://github.com/example/hello-app.git',
            branch: 'main',
            stages: [
              { name: 'checkout', status: 'success', startTime: '2024-05-20T14:30:00Z', endTime: '2024-05-20T14:30:30Z' },
              { name: 'install', status: 'success', startTime: '2024-05-20T14:30:30Z', endTime: '2024-05-20T14:32:00Z' },
              { name: 'test', status: 'success', startTime: '2024-05-20T14:32:00Z', endTime: '2024-05-20T14:33:00Z' },
              { name: 'build', status: 'success', startTime: '2024-05-20T14:33:00Z', endTime: '2024-05-20T14:35:00Z' },
              { name: 'build-image', status: 'running', startTime: '2024-05-20T14:35:00Z', endTime: null },
              { name: 'deploy', status: 'pending', startTime: null, endTime: null }
            ],
            envVars: [
              { name: 'APP_NAME', value: 'hello-app' },
              { name: 'APP_PORT', value: '8080' },
              { name: 'BUILD_VERSION', value: '1.0.0' }
            ],
            deploymentConfig: {
              namespace: 'default',
              deploymentName: 'hello-app-deployment',
              image: 'example/hello-app:v1.0.0',
              replicas: 2,
              strategy: 'rollingUpdate'
            }
          }
        })
      }
      // 演示Java项目流水线
      else if (id === 2001) {
        resolve({
          code: 0,
          msg: '获取流水线详情成功',
          data: {
            id: id,
            name: 'java-app-pipeline',
            description: 'Java/Spring Boot应用部署流水线',
            status: 'running',
            gitRepo: 'https://github.com/example/java-app.git',
            branch: 'main',
            stages: [
              { name: 'checkout', status: 'success', startTime: '2024-05-20T22:30:00Z', endTime: '2024-05-20T22:30:30Z' },
              { name: 'build', status: 'success', startTime: '2024-05-20T22:30:30Z', endTime: '2024-05-20T22:35:00Z' },
              { name: 'test', status: 'success', startTime: '2024-05-20T22:35:00Z', endTime: '2024-05-20T22:40:00Z' },
              { name: 'code-quality', status: 'success', startTime: '2024-05-20T22:40:00Z', endTime: '2024-05-20T22:42:00Z' },
              { name: 'package', status: 'success', startTime: '2024-05-20T22:42:00Z', endTime: '2024-05-20T22:45:00Z' },
              { name: 'build-image', status: 'running', startTime: '2024-05-20T22:45:00Z', endTime: null },
              { name: 'deploy', status: 'pending', startTime: null, endTime: null }
            ],
            envVars: [
              { name: 'APP_NAME', value: 'java-app' },
              { name: 'APP_PORT', value: '8080' },
              { name: 'BUILD_VERSION', value: '1.0.0' },
              { name: 'MAVEN_OPTS', value: '-Xmx1024m -Xms512m' },
              { name: 'SPRING_PROFILES_ACTIVE', value: 'production' },
              { name: 'SONAR_HOST_URL', value: 'http://sonarqube.example.com' }
            ],
            deploymentConfig: {
              namespace: 'default',
              deploymentName: 'java-app-deployment',
              image: 'example/java-app:v1.0.0',
              replicas: 3,
              strategy: 'rollingUpdate'
            }
          }
        })
      }
      // 原始流水线详情
      else {
        resolve({
          code: 0,
          msg: '获取流水线详情成功',
          data: {
            id: id,
            name: 'frontend-deploy',
            description: '前端应用部署流水线',
            status: 'running',
            gitRepo: 'https://github.com/example/frontend.git',
            branch: 'main',
            stages: [
              { name: 'checkout', status: 'success', startTime: '2024-05-20T14:30:00Z', endTime: '2024-05-20T14:30:30Z' },
              { name: 'build', status: 'success', startTime: '2024-05-20T14:30:30Z', endTime: '2024-05-20T14:35:00Z' },
              { name: 'test', status: 'success', startTime: '2024-05-20T14:35:00Z', endTime: '2024-05-20T14:40:00Z' },
              { name: 'build-image', status: 'running', startTime: '2024-05-20T14:40:00Z', endTime: null },
              { name: 'deploy', status: 'pending', startTime: null, endTime: null }
            ],
            envVars: [
              { name: 'NODE_ENV', value: 'production' },
              { name: 'API_URL', value: 'https://api.example.com' }
            ],
            deploymentConfig: {
              namespace: 'default',
              deploymentName: 'frontend-deployment',
              image: 'example/frontend:latest',
              replicas: 3,
              strategy: 'rollingUpdate'
            }
          }
        })
      }
    }, 1000)
  })
}

export const createPipeline = async (pipelineData) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '创建流水线成功',
        data: {
          id: Date.now(),
          name: pipelineData.name,
          description: pipelineData.description,
          status: 'idle',
          gitRepo: pipelineData.gitRepo,
          branch: pipelineData.branch,
          createdAt: new Date().toISOString()
        }
      })
    }, 1500)
  })
}

export const updatePipeline = async (id, pipelineData) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '更新流水线成功',
        data: {
          id: id,
          name: pipelineData.name,
          description: pipelineData.description,
          gitRepo: pipelineData.gitRepo,
          branch: pipelineData.branch,
          updatedAt: new Date().toISOString()
        }
      })
    }, 1500)
  })
}

export const deletePipeline = async (id) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '删除流水线成功',
        data: {
          id: id
        }
      })
    }, 1000)
  })
}

export const runPipeline = async (id) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '流水线启动成功',
        data: {
          pipelineId: id,
          runId: Date.now(),
          startTime: new Date().toISOString()
        }
      })
    }, 1000)
  })
}

export const stopPipeline = async (id) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '流水线停止成功'
      })
    }, 1000)
  })
}

export const getPipelineLogs = async (id, runId) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 演示Hello应用流水线日志
      if (id === 1001) {
        resolve({
          code: 0,
          msg: '获取流水线日志成功',
          data: {
            logs: [
              '[2024-05-20T14:30:00Z] 开始Hello应用流水线执行',
              '[2024-05-20T14:30:05Z] 开始拉取代码: git clone https://github.com/example/hello-app.git',
              '[2024-05-20T14:30:20Z] 代码拉取成功，检出分支: main',
              '[2024-05-20T14:30:30Z] 开始安装依赖: npm install',
              '[2024-05-20T14:31:45Z] 依赖安装成功',
              '[2024-05-20T14:32:00Z] 开始运行单元测试: npm test',
              '[2024-05-20T14:32:30Z] 测试用例: testGetHelloMessage 执行成功',
              '[2024-05-20T14:32:45Z] 测试用例: testServerStart 执行成功',
              '[2024-05-20T14:33:00Z] 所有2个测试用例通过',
              '[2024-05-20T14:33:05Z] 开始构建应用: npm run build',
              '[2024-05-20T14:34:30Z] 应用构建成功，生成dist目录',
              '[2024-05-20T14:35:00Z] 开始构建Docker镜像: docker build -t example/hello-app:v1.0.0 .',
              '[2024-05-20T14:35:30Z] 镜像构建中...',
              '[2024-05-20T14:36:00Z] 镜像构建完成，大小: 120MB',
              '[2024-05-20T14:36:15Z] 开始推送镜像到仓库: docker push example/hello-app:v1.0.0',
              '[2024-05-20T14:37:30Z] 镜像推送成功',
              '[2024-05-20T14:38:00Z] 开始部署到K8s集群',
              '[2024-05-20T14:38:15Z] 使用kubectl创建部署: kubectl apply -f deployment.yaml',
              '[2024-05-20T14:38:30Z] 部署创建成功，等待Pod启动...',
              '[2024-05-20T14:39:00Z] Pod 1/2 运行中',
              '[2024-05-20T14:39:30Z] Pod 2/2 运行中',
              '[2024-05-20T14:40:00Z] 部署完成，应用正在运行',
              '[2024-05-20T14:40:15Z] Hello应用演示流水线执行成功！'
            ]
          }
        })
      }
      // 演示Java项目流水线日志
      else if (id === 2001) {
        resolve({
          code: 0,
          msg: '获取流水线日志成功',
          data: {
            logs: [
              '[2024-05-20T22:30:00Z] 开始Java应用流水线执行',
              '[2024-05-20T22:30:05Z] 开始拉取代码: git clone https://github.com/example/java-app.git',
              '[2024-05-20T22:30:20Z] 代码拉取成功，检出分支: main',
              '[2024-05-20T22:30:30Z] 开始构建项目: mvn clean compile -DskipTests',
              '[2024-05-20T22:31:00Z] 下载依赖中...',
              '[2024-05-20T22:33:45Z] 依赖下载完成',
              '[2024-05-20T22:34:30Z] 编译成功，生成class文件',
              '[2024-05-20T22:35:00Z] 构建完成',
              '[2024-05-20T22:35:05Z] 开始运行测试: mvn test',
              '[2024-05-20T22:36:30Z] 运行单元测试...',
              '[2024-05-20T22:38:15Z] 单元测试通过: 25/25',
              '[2024-05-20T22:38:30Z] 运行集成测试...',
              '[2024-05-20T22:39:45Z] 集成测试通过: 10/10',
              '[2024-05-20T22:40:00Z] 所有测试通过',
              '[2024-05-20T22:40:05Z] 开始代码质量检查: mvn sonar:sonar -Dsonar.host.url=http://sonarqube.example.com',
              '[2024-05-20T22:41:15Z] 代码质量检查完成',
              '[2024-05-20T22:41:30Z] 代码质量评分: A (95/100)',
              '[2024-05-20T22:42:00Z] 开始打包应用: mvn package -DskipTests',
              '[2024-05-20T22:43:30Z] 生成JAR文件: target/java-app-1.0.0.jar',
              '[2024-05-20T22:45:00Z] 打包完成',
              '[2024-05-20T22:45:05Z] 开始构建Docker镜像: docker build -t example/java-app:v1.0.0 .',
              '[2024-05-20T22:46:30Z] 镜像构建中...',
              '[2024-05-20T22:48:00Z] 镜像构建完成，大小: 250MB',
              '[2024-05-20T22:48:15Z] 开始推送镜像到仓库: docker push example/java-app:v1.0.0',
              '[2024-05-20T22:50:30Z] 镜像推送成功',
              '[2024-05-20T22:50:35Z] 开始部署到K8s集群',
              '[2024-05-20T22:50:50Z] 使用kubectl创建部署: kubectl apply -f k8s/deployment.yaml',
              '[2024-05-20T22:51:05Z] 部署创建成功，等待Pod启动...',
              '[2024-05-20T22:52:00Z] Pod 1/3 运行中',
              '[2024-05-20T22:52:30Z] Pod 2/3 运行中',
              '[2024-05-20T22:53:00Z] Pod 3/3 运行中',
              '[2024-05-20T22:53:15Z] 服务已就绪，可正常访问',
              '[2024-05-20T22:53:30Z] 运行健康检查: curl -s http://java-app.default.svc.cluster.local:8080/actuator/health',
              '[2024-05-20T22:53:45Z] 健康检查通过: {"status":"UP"}',
              '[2024-05-20T22:54:00Z] 部署完成，应用正在运行',
              '[2024-05-20T22:54:15Z] Java应用流水线执行成功！'
            ]
          }
        })
      }
      // 原始流水线日志
      else {
        resolve({
          code: 0,
          msg: '获取流水线日志成功',
          data: {
            logs: [
              '[2024-05-20T14:30:00Z] 开始流水线执行',
              '[2024-05-20T14:30:05Z] 拉取代码成功',
              '[2024-05-20T14:30:30Z] 代码检查通过',
              '[2024-05-20T14:31:00Z] 开始构建应用',
              '[2024-05-20T14:34:30Z] 应用构建成功',
              '[2024-05-20T14:35:00Z] 开始运行测试',
              '[2024-05-20T14:39:45Z] 所有测试通过',
              '[2024-05-20T14:40:00Z] 开始构建镜像',
              '[2024-05-20T14:42:30Z] 镜像构建中...'
            ]
          }
        })
      }
    }, 1000)
  })
}

// 镜像仓库相关API
export const getImageRepositories = async () => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '获取镜像仓库列表成功',
        data: [
          {
            id: 1,
            name: 'docker-hub',
            type: 'docker',
            url: 'https://registry.hub.docker.com',
            status: 'connected'
          },
          {
            id: 2,
            name: 'private-registry',
            type: 'docker',
            url: 'https://registry.example.com',
            status: 'connected'
          },
          {
            id: 3,
            name: 'harbor',
            type: 'harbor',
            url: 'https://harbor.example.com',
            status: 'disconnected'
          }
        ]
      })
    }, 1000)
  })
}

export const createImageRepository = async (repoData) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '创建镜像仓库成功',
        data: {
          id: Date.now(),
          name: repoData.name,
          type: repoData.type,
          url: repoData.url,
          status: 'disconnected',
          createdAt: new Date().toISOString()
        }
      })
    }, 1500)
  })
}

export const updateImageRepository = async (id, repoData) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '更新镜像仓库成功',
        data: {
          id: id,
          name: repoData.name,
          type: repoData.type,
          url: repoData.url,
          status: repoData.status || 'disconnected',
          updatedAt: new Date().toISOString()
        }
      })
    }, 1500)
  })
}

export const deleteImageRepository = async (id) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '删除镜像仓库成功',
        data: {
          id: id
        }
      })
    }, 1000)
  })
}

export const getImages = async (repoId) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '获取镜像列表成功',
        data: [
          {
            id: 1,
            name: 'frontend',
            tags: ['latest', 'v1.0.0', 'v1.0.1', 'v1.1.0-beta'],
            repositoryId: repoId,
            lastUpdate: '2024-05-20T14:42:30Z'
          },
          {
            id: 2,
            name: 'backend',
            tags: ['latest', 'v2.0.0', 'v2.0.1'],
            repositoryId: repoId,
            lastUpdate: '2024-05-20T13:15:00Z'
          },
          {
            id: 3,
            name: 'database',
            tags: ['latest', 'v1.0.0'],
            repositoryId: repoId,
            lastUpdate: '2024-05-19T09:45:00Z'
          }
        ]
      })
    }, 1000)
  })
}

export const getImageTags = async (repoId, imageName) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '获取镜像标签列表成功',
        data: {
          tags: ['latest', 'v1.0.0', 'v1.0.1', 'v1.1.0-beta'],
          imageName: imageName,
          repositoryId: repoId
        }
      })
    }, 1000)
  })
}

export const deleteImageTag = async (repoId, imageName, tag) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '删除镜像标签成功',
        data: {
          repoId: repoId,
          imageName: imageName,
          tag: tag
        }
      })
    }, 1500)
  })
}

// K8s环境管理API
export const getK8sEnvironments = async () => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '获取K8s环境列表成功',
        data: [
          {
            id: 1,
            name: 'production',
            description: '生产环境',
            type: 'production',
            apiUrl: 'https://k8s-prod.example.com:6443',
            status: 'connected',
            clusterName: 'production-cluster',
            namespace: 'default'
          },
          {
            id: 2,
            name: 'staging',
            description: '预发布环境',
            type: 'staging',
            apiUrl: 'https://k8s-staging.example.com:6443',
            status: 'connected',
            clusterName: 'staging-cluster',
            namespace: 'default'
          },
          {
            id: 3,
            name: 'testing',
            description: '测试环境',
            type: 'testing',
            apiUrl: 'https://k8s-testing.example.com:6443',
            status: 'connected',
            clusterName: 'testing-cluster',
            namespace: 'default'
          },
          {
            id: 4,
            name: 'development',
            description: '开发环境',
            type: 'development',
            apiUrl: 'https://k8s-dev.example.com:6443',
            status: 'connected',
            clusterName: 'development-cluster',
            namespace: 'development'
          }
        ]
      })
    }, 1000)
  })
}

export const getK8sEnvironmentDetail = async (id) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 根据id返回不同环境的详情
      const environments = {
        1: {
          name: 'production',
          description: '生产环境',
          type: 'production',
          apiUrl: 'https://k8s-prod.example.com:6443',
          status: 'connected',
          clusterName: 'production-cluster',
          namespace: 'default'
        },
        2: {
          name: 'staging',
          description: '预发布环境',
          type: 'staging',
          apiUrl: 'https://k8s-staging.example.com:6443',
          status: 'connected',
          clusterName: 'staging-cluster',
          namespace: 'default'
        },
        3: {
          name: 'testing',
          description: '测试环境',
          type: 'testing',
          apiUrl: 'https://k8s-testing.example.com:6443',
          status: 'connected',
          clusterName: 'testing-cluster',
          namespace: 'default'
        },
        4: {
          name: 'development',
          description: '开发环境',
          type: 'development',
          apiUrl: 'https://k8s-dev.example.com:6443',
          status: 'connected',
          clusterName: 'development-cluster',
          namespace: 'development'
        }
      };

      const env = environments[id] || environments[1];

      resolve({
        code: 0,
        msg: '获取K8s环境详情成功',
        data: {
          id: id,
          ...env,
          certificateAuthority: '-----BEGIN CERTIFICATE-----...-----END CERTIFICATE-----',
          clientCertificate: '-----BEGIN CERTIFICATE-----...-----END CERTIFICATE-----',
          clientKey: '-----BEGIN PRIVATE KEY-----...-----END PRIVATE KEY-----',
          createdAt: '2024-01-15T10:00:00Z',
          updatedAt: '2024-05-20T14:30:00Z'
        }
      })
    }, 1000)
  })
}

export const createK8sEnvironment = async (envData) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '创建K8s环境成功',
        data: {
          id: Date.now(),
          name: envData.name,
          description: envData.description,
          type: envData.type || 'development',
          apiUrl: envData.apiUrl,
          clusterName: envData.clusterName,
          namespace: envData.namespace || 'default',
          status: 'disconnected',
          createdAt: new Date().toISOString()
        }
      })
    }, 1500)
  })
}

export const updateK8sEnvironment = async (id, envData) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '更新K8s环境成功',
        data: {
          id: id,
          name: envData.name,
          description: envData.description,
          type: envData.type || 'development',
          apiUrl: envData.apiUrl,
          clusterName: envData.clusterName,
          namespace: envData.namespace,
          updatedAt: new Date().toISOString()
        }
      })
    }, 1500)
  })
}

export const deleteK8sEnvironment = async (id) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '删除K8s环境成功',
        data: {
          id: id
        }
      })
    }, 1000)
  })
}

// K8s集群管理API
export const getK8sClusters = async () => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '获取K8s集群列表成功',
        data: [
          {
            id: 1,
            clusterName: 'production-cluster',
            clusterVersion: 'v1.28.3',
            status: 'active',
            apiServer: 'https://k8s-prod.example.com:6443',
            nodeCount: 5,
            createdAt: '2024-01-15T10:00:00Z',
            isDefault: true
          },
          {
            id: 2,
            clusterName: 'staging-cluster',
            clusterVersion: 'v1.27.7',
            status: 'active',
            apiServer: 'https://k8s-staging.example.com:6443',
            nodeCount: 3,
            createdAt: '2024-02-20T14:30:00Z',
            isDefault: false
          },
          {
            id: 3,
            clusterName: 'testing-cluster',
            clusterVersion: 'v1.28.2',
            status: 'inactive',
            apiServer: 'https://k8s-testing.example.com:6443',
            nodeCount: 2,
            createdAt: '2024-03-10T09:15:00Z',
            isDefault: false
          }
        ]
      })
    }, 1000)
  })
}

export const createK8sCluster = async (clusterData) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '创建K8s集群成功',
        data: {
          id: Date.now(),
          clusterName: clusterData.cluster_name,
          clusterVersion: clusterData.cluster_version,
          status: 'active',
          apiServer: clusterData.api_server || '',
          nodeCount: 0,
          createdAt: new Date().toISOString(),
          isDefault: clusterData.set_as_default || false
        }
      })
    }, 1500)
  })
}

export const updateK8sCluster = async (id, clusterData) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '更新K8s集群成功',
        data: {
          id: id,
          clusterName: clusterData.cluster_name,
          clusterVersion: clusterData.cluster_version,
          status: clusterData.status,
          apiServer: clusterData.api_server || '',
          nodeCount: clusterData.node_count || 0,
          updatedAt: new Date().toISOString(),
          isDefault: clusterData.set_as_default || false
        }
      })
    }, 1500)
  })
}

export const deleteK8sCluster = async (id) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '删除K8s集群成功',
        data: {
          id: id
        }
      })
    }, 1000)
  })
}

export const getK8sClusterDetail = async (id) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '获取K8s集群详情成功',
        data: {
          id: id,
          clusterName: 'production-cluster',
          clusterVersion: 'v1.28.3',
          status: 'active',
          apiServer: 'https://k8s-prod.example.com:6443',
          nodeCount: 5,
          nodes: [
            { name: 'node-1', status: 'ready', cpu: '80%', memory: '75%' },
            { name: 'node-2', status: 'ready', cpu: '65%', memory: '60%' },
            { name: 'node-3', status: 'ready', cpu: '70%', memory: '68%' },
            { name: 'node-4', status: 'ready', cpu: '55%', memory: '50%' },
            { name: 'node-5', status: 'ready', cpu: '45%', memory: '40%' }
          ],
          namespaces: ['default', 'kube-system', 'kube-public'],
          createdAt: '2024-01-15T10:00:00Z',
          isDefault: true
        }
      })
    }, 1000)
  })
}

export const checkClusterConnectivity = async (id) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '检查集群连通性成功',
        data: {
          id: id,
          connected: true,
          latency: 45,
          lastChecked: new Date().toISOString()
        }
      })
    }, 1000)
  })
}

// 部署相关API
export const deployToK8s = async (deploymentConfig) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '部署成功',
        data: {
          deploymentId: Date.now(),
          namespace: deploymentConfig.namespace,
          deploymentName: deploymentConfig.deploymentName,
          image: deploymentConfig.image,
          replicas: deploymentConfig.replicas,
          strategy: deploymentConfig.strategy,
          startTime: new Date().toISOString(),
          environmentId: deploymentConfig.environmentId,
          environmentName: deploymentConfig.environmentName
        }
      })
    }, 1500)
  })
}

export const getDeploymentHistory = async (namespace, deploymentName) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '获取部署历史成功',
        data: [
          {
            revision: 3,
            image: 'example/frontend:v1.0.1',
            replicas: 3,
            strategy: 'rollingUpdate',
            deploymentTime: '2024-05-20T14:30:00Z',
            status: 'success',
            environmentName: 'production'
          },
          {
            revision: 2,
            image: 'example/frontend:v1.0.0',
            replicas: 2,
            strategy: 'rollingUpdate',
            deploymentTime: '2024-05-19T10:20:00Z',
            status: 'success',
            environmentName: 'production'
          },
          {
            revision: 1,
            image: 'example/frontend:v0.9.0',
            replicas: 1,
            strategy: 'recreate',
            deploymentTime: '2024-05-18T16:45:00Z',
            status: 'success',
            environmentName: 'production'
          }
        ]
      })
    }, 1000)
  })
}

// 流水线模板相关API
export const getPipelineTemplates = async () => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '获取流水线模板列表成功',
        data: [
          {
            id: 1,
            name: '前端应用部署模板',
            description: '适用于Vue/React等前端应用的部署流水线模板',
            type: 'frontend',
            stages: [
              { name: 'checkout', description: '拉取代码' },
              { name: 'install', description: '安装依赖' },
              { name: 'build', description: '构建应用' },
              { name: 'test', description: '运行测试' },
              { name: 'build-image', description: '构建镜像' },
              { name: 'deploy', description: '部署到K8s' }
            ],
            defaultEnvVars: [
              { name: 'NODE_ENV', value: 'production' },
              { name: 'BUILD_VERSION', value: '${BUILD_NUMBER}' }
            ],
            defaultDeploymentConfig: {
              replicas: 3,
              strategy: 'rollingUpdate',
              resources: {
                limits: {
                  cpu: '500m',
                  memory: '512Mi'
                },
                requests: {
                  cpu: '200m',
                  memory: '256Mi'
                }
              }
            },
            createdAt: '2024-01-15T10:00:00Z',
            updatedAt: '2024-05-20T14:30:00Z'
          },
          {
            id: 2,
            name: '后端服务部署模板',
            description: '适用于Spring Boot/Node.js等后端服务的部署流水线模板',
            type: 'backend',
            stages: [
              { name: 'checkout', description: '拉取代码' },
              { name: 'install', description: '安装依赖' },
              { name: 'build', description: '构建应用' },
              { name: 'test', description: '运行单元测试' },
              { name: 'code-quality', description: '代码质量检查' },
              { name: 'build-image', description: '构建镜像' },
              { name: 'deploy', description: '部署到K8s' }
            ],
            defaultEnvVars: [
              { name: 'SPRING_PROFILES_ACTIVE', value: 'production' },
              { name: 'DB_URL', value: 'jdbc:mysql://db:3306/app' }
            ],
            defaultDeploymentConfig: {
              replicas: 3,
              strategy: 'rollingUpdate',
              resources: {
                limits: {
                  cpu: '1000m',
                  memory: '1Gi'
                },
                requests: {
                  cpu: '500m',
                  memory: '512Mi'
                }
              }
            },
            createdAt: '2024-01-20T14:20:00Z',
            updatedAt: '2024-05-15T09:15:00Z'
          },
          {
            id: 3,
            name: '数据库迁移模板',
            description: '适用于数据库迁移的流水线模板',
            type: 'database',
            stages: [
              { name: 'checkout', description: '拉取迁移脚本' },
              { name: 'validate', description: '验证迁移脚本' },
              { name: 'backup', description: '备份数据库' },
              { name: 'migrate', description: '执行迁移' },
              { name: 'verify', description: '验证迁移结果' }
            ],
            defaultEnvVars: [
              { name: 'DB_HOST', value: 'db.example.com' },
              { name: 'DB_PORT', value: '3306' }
            ],
            defaultDeploymentConfig: {},
            createdAt: '2024-02-05T16:45:00Z',
            updatedAt: '2024-05-10T11:30:00Z'
          },
          {
            id: 4,
            name: 'Hello应用演示模板',
            description: '用于演示完整CI/CD流程的Hello应用模板',
            type: 'demo',
            stages: [
              { name: 'checkout', description: '拉取Hello应用代码' },
              { name: 'install', description: '安装依赖' },
              { name: 'test', description: '运行单元测试' },
              { name: 'build', description: '构建应用' },
              { name: 'build-image', description: '构建Docker镜像' },
              { name: 'deploy', description: '部署到K8s集群' }
            ],
            defaultEnvVars: [
              { name: 'APP_NAME', value: 'hello-app' },
              { name: 'APP_PORT', value: '8080' },
              { name: 'BUILD_VERSION', value: '${BUILD_NUMBER}' }
            ],
            defaultDeploymentConfig: {
              replicas: 2,
              strategy: 'rollingUpdate',
              resources: {
                limits: {
                  cpu: '200m',
                  memory: '256Mi'
                },
                requests: {
                  cpu: '100m',
                  memory: '128Mi'
                }
              }
            },
            createdAt: '2024-05-20T10:00:00Z',
            updatedAt: '2024-05-20T14:30:00Z'
          },
          {
            id: 5,
            name: 'Java项目部署模板',
            description: '适用于Java/Spring Boot项目的完整CI/CD流水线模板',
            type: 'java',
            stages: [
              { name: 'checkout', description: '拉取Java项目代码' },
              { name: 'build', description: '构建Java项目 (Maven/Gradle)' },
              { name: 'test', description: '运行单元测试和集成测试' },
              { name: 'code-quality', description: '代码质量检查 (SonarQube)' },
              { name: 'package', description: '打包成JAR/WAR文件' },
              { name: 'build-image', description: '构建Docker镜像' },
              { name: 'deploy', description: '部署到K8s集群' }
            ],
            defaultEnvVars: [
              { name: 'APP_NAME', value: 'java-app' },
              { name: 'APP_PORT', value: '8080' },
              { name: 'BUILD_VERSION', value: '${BUILD_NUMBER}' },
              { name: 'MAVEN_OPTS', value: '-Xmx1024m -Xms512m' },
              { name: 'SPRING_PROFILES_ACTIVE', value: 'production' },
              { name: 'SONAR_HOST_URL', value: 'http://sonarqube.example.com' }
            ],
            defaultDeploymentConfig: {
              replicas: 3,
              strategy: 'rollingUpdate',
              resources: {
                limits: {
                  cpu: '1000m',
                  memory: '1Gi'
                },
                requests: {
                  cpu: '500m',
                  memory: '512Mi'
                }
              }
            },
            createdAt: '2024-05-20T10:00:00Z',
            updatedAt: '2024-05-20T14:30:00Z'
          }
        ]
      })
    }, 1000)
  })
}

export const getPipelineTemplateDetail = async (id) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '获取流水线模板详情成功',
        data: {
          id: id,
          name: '前端应用部署模板',
          description: '适用于Vue/React等前端应用的部署流水线模板',
          type: 'frontend',
          stages: [
            { name: 'checkout', description: '拉取代码' },
            { name: 'install', description: '安装依赖' },
            { name: 'build', description: '构建应用' },
            { name: 'test', description: '运行测试' },
            { name: 'build-image', description: '构建镜像' },
            { name: 'deploy', description: '部署到K8s' }
          ],
          defaultEnvVars: [
            { name: 'NODE_ENV', value: 'production' },
            { name: 'BUILD_VERSION', value: '${BUILD_NUMBER}' }
          ],
          defaultDeploymentConfig: {
            replicas: 3,
            strategy: 'rollingUpdate',
            resources: {
              limits: {
                cpu: '500m',
                memory: '512Mi'
              },
              requests: {
                cpu: '200m',
                memory: '256Mi'
              }
            }
          },
          createdAt: '2024-01-15T10:00:00Z',
          updatedAt: '2024-05-20T14:30:00Z'
        }
      })
    }, 1000)
  })
}

export const createPipelineTemplate = async (templateData) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '创建流水线模板成功',
        data: {
          id: Date.now(),
          name: templateData.name,
          description: templateData.description,
          type: templateData.type,
          stages: templateData.stages,
          defaultEnvVars: templateData.defaultEnvVars || [],
          defaultDeploymentConfig: templateData.defaultDeploymentConfig || {},
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        }
      })
    }, 1500)
  })
}

export const updatePipelineTemplate = async (id, templateData) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '更新流水线模板成功',
        data: {
          id: id,
          name: templateData.name,
          description: templateData.description,
          type: templateData.type,
          stages: templateData.stages,
          defaultEnvVars: templateData.defaultEnvVars || [],
          defaultDeploymentConfig: templateData.defaultDeploymentConfig || {},
          updatedAt: new Date().toISOString()
        }
      })
    }, 1500)
  })
}

export const deletePipelineTemplate = async (id) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        code: 0,
        msg: '删除流水线模板成功',
        data: {
          id: id
        }
      })
    }, 1000)
  })
}
