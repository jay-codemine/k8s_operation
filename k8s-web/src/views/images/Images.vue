<template>
  <div class="images-page">
    <h2>镜像列表</h2>

    <div class="toolbar">
      <button class="btn btn-primary" @click="backToRepositories">返回仓库列表</button>
      <div class="repo-info">
        <span class="repo-name">{{ repositoryName }}</span>
        <span class="repo-url">{{ repositoryUrl }}</span>
      </div>
    </div>

    <!-- 镜像列表 -->
    <div class="images-card">
      <h3>镜像</h3>
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>镜像名称</th>
            <th>标签数量</th>
            <th>最后更新</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="image in paginatedImages" :key="image.id">
            <td>{{ image.id }}</td>
            <td>{{ image.name }}</td>
            <td>{{ image.tags.length }}</td>
            <td>{{ formatDate(image.lastUpdate) }}</td>
            <td>
              <button class="btn btn-view" @click="viewImageTags(image)">查看标签</button>
              <button class="btn btn-primary" @click="deployImage(image)">部署</button>
            </td>
          </tr>
        </tbody>
      </table>

      <Pagination
        :current-page="currentPage"
        :total-items="images.length"
        :items-per-page="pageSize"
        @update:currentPage="currentPage = $event"
      />
    </div>

    <!-- 镜像标签详情 -->
    <div class="tags-card" v-if="selectedImage">
      <h3>{{ selectedImage.name }} - 标签</h3>
      <div class="tags-container">
        <div
          v-for="tag in selectedImage.tags"
          :key="tag"
          class="tag-item"
        >
          <div class="tag-name">{{ tag }}</div>
          <div class="tag-actions">
            <button class="btn btn-small btn-primary" @click="deployTag(tag)">部署</button>
            <button class="btn btn-small btn-secondary" @click="copyImageUrl(tag)">复制镜像URL</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 部署表单 -->
    <div class="deploy-modal" v-if="showDeployForm">
      <div class="modal-content">
        <div class="modal-header">
          <h3>部署镜像</h3>
          <button class="close-btn" @click="showDeployForm = false">&times;</button>
        </div>
        <form class="deploy-form" @submit.prevent="deploy">
          <div class="form-row">
            <div class="form-group">
              <label>镜像</label>
              <input
                type="text"
                v-model="deployConfig.image"
                class="form-input"
                readonly
              />
            </div>
            <div class="form-group">
              <label>命名空间</label>
              <input
                type="text"
                v-model="deployConfig.namespace"
                class="form-input"
                required
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>部署名称</label>
              <input
                type="text"
                v-model="deployConfig.deploymentName"
                class="form-input"
                required
              />
            </div>
            <div class="form-group">
              <label>副本数量</label>
              <input
                type="number"
                v-model.number="deployConfig.replicas"
                class="form-input"
                min="1"
                required
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>部署策略</label>
              <select v-model="deployConfig.strategy" class="form-select">
                <option value="rollingUpdate">滚动更新</option>
                <option value="recreate">重新创建</option>
                <option value="blueGreen">蓝绿部署</option>
                <option value="canary">金丝雀发布</option>
              </select>
            </div>
          </div>

          <div class="form-actions">
            <button type="submit" class="btn btn-primary">部署</button>
            <button type="button" class="btn btn-secondary" @click="showDeployForm = false">取消</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getImages, deployToK8s } from '@/api/cicd.js'
import Pagination from '@/components/Pagination.vue'

export default {
  name: 'Images',
  components: {
    Pagination
  },
  setup() {
    const route = useRoute()
    const router = useRouter()
    const repoId = ref(route.params.repoId || 1)

    const images = ref([])
    const currentPage = ref(1)
    const pageSize = ref(10)
    const selectedImage = ref(null)
    const showDeployForm = ref(false)
    const repositoryName = ref('')
    const repositoryUrl = ref('')

    const deployConfig = ref({
      namespace: 'default',
      deploymentName: '',
      image: '',
      replicas: 1,
      strategy: 'rollingUpdate'
    })

    const loadImages = async () => {
      try {
        const response = await getImages(repoId.value)
        if (response.code === 0) {
          images.value = response.data
          // 模拟获取仓库信息
          repositoryName.value = 'docker-hub'
          repositoryUrl.value = 'https://registry.hub.docker.com'
        } else {
          alert(response.msg)
        }
      } catch (error) {
        alert('获取镜像列表失败')
      }
    }

    const paginatedImages = computed(() => {
      const startIndex = (currentPage.value - 1) * pageSize.value
      const endIndex = startIndex + pageSize.value
      return images.value.slice(startIndex, endIndex)
    })

    const viewImageTags = (image) => {
      selectedImage.value = image
    }

    const deployImage = (image) => {
      deployConfig.value = {
        namespace: 'default',
        deploymentName: image.name,
        image: `${image.name}:latest`,
        replicas: 1,
        strategy: 'rollingUpdate'
      }
      showDeployForm.value = true
    }

    const deployTag = (tag) => {
      if (!selectedImage.value) return

      deployConfig.value = {
        namespace: 'default',
        deploymentName: selectedImage.value.name,
        image: `${selectedImage.value.name}:${tag}`,
        replicas: 1,
        strategy: 'rollingUpdate'
      }
      showDeployForm.value = true
    }

    const deploy = async () => {
      try {
        const response = await deployToK8s(deployConfig.value)
        if (response.code === 0) {
          alert('部署成功')
          showDeployForm.value = false
        } else {
          alert(response.msg)
        }
      } catch (error) {
        alert('部署失败')
      }
    }

    const copyImageUrl = (tag) => {
      if (!selectedImage.value) return

      const imageUrl = `${repositoryUrl.value}/${selectedImage.value.name}:${tag}`
      navigator.clipboard.writeText(imageUrl).then(() => {
        alert('镜像URL已复制到剪贴板')
      }).catch(() => {
        alert('复制失败，请手动复制')
      })
    }

    const backToRepositories = () => {
      router.push('/image-repositories')
    }

    const formatDate = (dateString) => {
      if (!dateString) return ''
      const date = new Date(dateString)
      return date.toLocaleString()
    }

    onMounted(() => {
      loadImages()
    })

    return {
      images,
      currentPage,
      pageSize,
      paginatedImages,
      selectedImage,
      showDeployForm,
      deployConfig,
      repositoryName,
      repositoryUrl,
      viewImageTags,
      deployImage,
      deployTag,
      deploy,
      copyImageUrl,
      backToRepositories,
      formatDate
    }
  }
}
</script>

<style scoped>
.images-page {
  padding: 20px;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 20px;
}

.repo-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.repo-name {
  font-weight: 600;
  font-size: 16px;
  color: #2d3748;
}

.repo-url {
  font-size: 14px;
  color: #718096;
}

.images-card, .tags-card {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  padding: 20px;
  margin-bottom: 20px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 16px;
  margin-bottom: 20px;
}

.data-table th,
.data-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
}

.data-table th {
  background-color: #f7fafc;
  font-weight: 600;
  color: #4a5568;
  font-size: 14px;
}

.data-table tr:hover {
  background-color: #f7fafc;
}

.tags-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 16px;
  margin-top: 16px;
}

.tag-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #f7fafc;
  padding: 12px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.tag-name {
  font-weight: 600;
  color: #2d3748;
}

.tag-actions {
  display: flex;
  gap: 8px;
}

.deploy-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e2e8f0;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #718096;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.close-btn:hover {
  background-color: #f7fafc;
  color: #2d3748;
}

.deploy-form {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-weight: 600;
  color: #4a5568;
  font-size: 14px;
}

.form-input, .form-select {
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  font-size: 14px;
  transition: border-color 0.3s ease;
}

.form-input:focus, .form-select:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.form-input:read-only {
  background-color: #f7fafc;
  cursor: not-allowed;
}

.form-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 8px;
}

.btn {
  padding: 8px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-small {
  padding: 4px 8px;
  font-size: 12px;
}

.btn-primary {
  background-color: #326ce5;
  color: white;
  border-color: #326ce5;
}

.btn-primary:hover {
  background-color: #2554c7;
}

.btn-view {
  background-color: #6c757d;
  color: white;
  border-color: #6c757d;
}

.btn-view:hover {
  background-color: #5a6268;
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
