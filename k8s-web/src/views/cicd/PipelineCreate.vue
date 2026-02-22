<template>
  <div class="pipeline-create">
    <h2>{{ isEdit ? '编辑流水线' : '创建流水线' }}</h2>

    <form class="pipeline-form" @submit.prevent="submit">
      <div class="form-row">
        <div class="form-group">
          <label>流水线名称</label>
          <input
            type="text"
            v-model="pipelineData.name"
            class="form-input"
            placeholder="请输入流水线名称"
            required
          />
        </div>

        <div class="form-group">
          <label>模板选择</label>
          <select
            v-model="selectedTemplateId"
            class="form-input"
            @change="handleTemplateChange"
          >
            <option value="">不使用模板</option>
            <option
              v-for="template in templates"
              :key="template.id"
              :value="template.id"
            >
              {{ template.name }} - {{ template.description }}
            </option>
          </select>
        </div>
      </div>

      <div class="form-group">
        <label>描述</label>
        <textarea
          v-model="pipelineData.description"
          class="form-textarea"
          placeholder="请输入流水线描述"
          rows="3"
        ></textarea>
      </div>

      <div class="form-row">
        <div class="form-group">
          <label>Git仓库URL</label>
          <input
            type="url"
            v-model="pipelineData.gitRepo"
            class="form-input"
            placeholder="请输入Git仓库URL"
            required
          />
        </div>

        <div class="form-group">
          <label>分支</label>
          <input
            type="text"
            v-model="pipelineData.branch"
            class="form-input"
            placeholder="请输入分支名称"
            required
          />
        </div>
      </div>

      <!-- 环境变量配置 -->
      <div class="form-section">
        <h3>环境变量</h3>
        <div class="env-vars">
          <div
            v-for="(envVar, index) in pipelineData.envVars"
            :key="index"
            class="env-var-item"
          >
            <input
              type="text"
              v-model="envVar.name"
              class="form-input"
              placeholder="变量名称"
              style="margin-right: 10px; width: 200px;"
            />
            <input
              type="text"
              v-model="envVar.value"
              class="form-input"
              placeholder="变量值"
              style="flex: 1;"
            />
            <button
              type="button"
              class="btn btn-danger"
              @click="removeEnvVar(index)"
              style="margin-left: 10px;"
            >
              删除
            </button>
          </div>
          <button
            type="button"
            class="btn btn-outline"
            @click="addEnvVar"
            style="margin-top: 10px;"
          >
            添加环境变量
          </button>
        </div>
      </div>

      <!-- 部署配置 -->
      <div class="form-section">
        <h3>部署配置</h3>
        <div class="form-row">
          <div class="form-group">
            <label>副本数</label>
            <input
              type="number"
              v-model.number="pipelineData.deploymentConfig.replicas"
              class="form-input"
              min="1"
            />
          </div>

          <div class="form-group">
            <label>部署策略</label>
            <select v-model="pipelineData.deploymentConfig.strategy" class="form-input">
              <option value="rollingUpdate">滚动更新</option>
              <option value="recreate">重新创建</option>
              <option value="blueGreen">蓝绿部署</option>
              <option value="canary">金丝雀部署</option>
            </select>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>CPU限制</label>
            <input
              type="text"
              v-model="pipelineData.deploymentConfig.resources.limits.cpu"
              class="form-input"
              placeholder="例如: 500m"
            />
          </div>

          <div class="form-group">
            <label>内存限制</label>
            <input
              type="text"
              v-model="pipelineData.deploymentConfig.resources.limits.memory"
              class="form-input"
              placeholder="例如: 512Mi"
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>CPU请求</label>
            <input
              type="text"
              v-model="pipelineData.deploymentConfig.resources.requests.cpu"
              class="form-input"
              placeholder="例如: 200m"
            />
          </div>

          <div class="form-group">
            <label>内存请求</label>
            <input
              type="text"
              v-model="pipelineData.deploymentConfig.resources.requests.memory"
              class="form-input"
              placeholder="例如: 256Mi"
            />
          </div>
        </div>
      </div>

      <div class="form-actions">
        <button type="button" class="btn btn-secondary" @click="cancel">取消</button>
        <button type="submit" class="btn btn-primary" :disabled="submitting">{{ submitting ? '提交中...' : '提交' }}</button>
      </div>
    </form>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  createPipeline,
  updatePipeline,
  getPipelineDetail,
  getPipelineTemplates
} from '@/api/cicd.js'

export default {
  name: 'PipelineCreate',
  setup() {
    const router = useRouter()
    const route = useRoute()
    const pipelineId = route.params.id
    const isEdit = !!pipelineId

    const templates = ref([])
    const selectedTemplateId = ref('')

    const pipelineData = ref({
      name: '',
      description: '',
      gitRepo: '',
      branch: 'main',
      envVars: [],
      deploymentConfig: {
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
      }
    })

    const submitting = ref(false)

    // 加载流水线模板
    const loadTemplates = async () => {
      try {
        const response = await getPipelineTemplates()
        if (response.code === 0) {
          templates.value = response.data
        }
      } catch (error) {
        console.error('获取流水线模板失败:', error)
      }
    }

    const loadPipelineData = async () => {
      if (isEdit) {
        try {
          const response = await getPipelineDetail(pipelineId)
          if (response.code === 0) {
            pipelineData.value = {
              name: response.data.name,
              description: response.data.description,
              gitRepo: response.data.gitRepo,
              branch: response.data.branch,
              envVars: response.data.envVars || [],
              deploymentConfig: response.data.deploymentConfig || {
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
              }
            }
          }
        } catch (error) {
          alert('获取流水线详情失败')
        }
      }
    }

    // 处理模板选择变化
    const handleTemplateChange = () => {
      if (selectedTemplateId.value) {
        const template = templates.value.find(t => t.id === parseInt(selectedTemplateId.value))
        if (template) {
          // 应用模板的环境变量和部署配置
          pipelineData.value.envVars = JSON.parse(JSON.stringify(template.defaultEnvVars))
          pipelineData.value.deploymentConfig = JSON.parse(JSON.stringify(template.defaultDeploymentConfig))
        }
      } else {
        // 重置为默认值
        pipelineData.value.envVars = []
        pipelineData.value.deploymentConfig = {
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
        }
      }
    }

    // 添加环境变量
    const addEnvVar = () => {
      pipelineData.value.envVars.push({ name: '', value: '' })
    }

    // 删除环境变量
    const removeEnvVar = (index) => {
      pipelineData.value.envVars.splice(index, 1)
    }

    const submit = async () => {
      try {
        submitting.value = true
        let response

        if (isEdit) {
          response = await updatePipeline(pipelineId, pipelineData.value)
        } else {
          response = await createPipeline(pipelineData.value)
        }

        if (response.code === 0) {
          alert(isEdit ? '更新流水线成功' : '创建流水线成功')
          router.push('/pipelines')
        } else {
          alert(response.msg)
        }
      } catch (error) {
        alert(isEdit ? '更新流水线失败' : '创建流水线失败')
      } finally {
        submitting.value = false
      }
    }

    const cancel = () => {
      router.push('/pipelines')
    }

    onMounted(() => {
      loadTemplates()
      loadPipelineData()
    })

    return {
      isEdit,
      pipelineData,
      templates,
      selectedTemplateId,
      submitting,
      handleTemplateChange,
      addEnvVar,
      removeEnvVar,
      submit,
      cancel
    }
  }
}
</script>

<style scoped>
.pipeline-create {
  padding: 20px;
}

.pipeline-form {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  padding: 24px;
  max-width: 600px;
}

.form-group {
  margin-bottom: 20px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #4a5568;
  font-size: 14px;
}

.form-input,
.form-textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  font-size: 14px;
  transition: border-color 0.3s ease;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.form-textarea {
  resize: vertical;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #e2e8f0;
}

.btn {
  padding: 10px 20px;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-primary {
  background-color: #326ce5;
  color: white;
  border-color: #326ce5;
}

.btn-primary:hover:not(:disabled) {
  background-color: #2554c7;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background-color: #6c757d;
  color: white;
  border-color: #6c757d;
}

.btn-secondary:hover {
  background-color: #5a6268;
}
</style>
