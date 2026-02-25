// src/api/cluster/workloads/pods.js
import http from '@/api/http'
import {K8S_BASE} from '@/api/paths'

// =========================
// Pod 模块 API（全量）
// 对应后端路由: /api/v1/k8s/pod/*
// =========================
const podsApi = {
  // =========================
  // Pod 基础 CRUD
  // =========================

  /**
   * 创建 Pod
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Pod名称
   * @param {string} data.image - 容器镜像
   * @param {Object} [data.labels] - 标签
   */
  create(data) {
    return http.post(`${K8S_BASE}/pod/create`, data)
  },

  /**
   * 从 YAML 创建 Pod
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  createFromYaml(data) {
    return http.post(`${K8S_BASE}/pod/create-from-yaml`, data)
  },

  /**
   * Pod 列表
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {number} params.page - 页码
   * @param {number} params.limit - 每页数量
   * @param {string} [params.name] - Pod名称（模糊匹配）
   */
  list(params) {
    return http.get(`${K8S_BASE}/pod/list`, {params})
  },

  /**
   * Pod 详情
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Pod名称
   */
  detail(params) {
    return http.get(`${K8S_BASE}/pod/detail`, {params})
  },

  /**
   * 更新 Pod（整体更新）
   * 注意：必须使用 POST 方法，后端不支持 PUT
   * @param {Object} data - Pod YAML / JSON 完整对象
   */
  update(data) {
    return http.post(`${K8S_BASE}/pod/update`, data)
  },

  /**
   * 删除 Pod（支持优雅删除/强制删除）
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Pod名称
   * @param {number} [params.grace_seconds] - 优雅终止秒数（默认30）
   * @param {boolean} [params.force] - 是否强制删除（默认false）
   */
  delete(params) {
    return http.delete(`${K8S_BASE}/pod/grace_delete_pod`, {params})
  },

  /**
   * 优雅删除 Pod（delete 的别名，默认优雅删除）
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Pod名称
   * @param {number} [params.grace_seconds] - 优雅终止秒数（默认30）
   */
  graceDelete(params) {
    return http.delete(`${K8S_BASE}/pod/grace_delete_pod`, {params})
  },

  /**
   * 强制删除 Pod（delete 的便捷方法，force=true）
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Pod名称
   */
  forceDelete(params) {
    return http.delete(`${K8S_BASE}/pod/grace_delete_pod`, {
      params: {...params, force: true}
    })
  },

  // =========================
  // 事件查询
  // =========================

  /**
   * 获取 Pod 相关事件
   * 使用通用事件接口，通过 kind=Pod 筛选
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Pod名称
   * @param {string} [params.type] - 事件类型（Normal | Warning）
   * @param {number} [params.limit=50] - 返回条数限制
   * @param {number} [params.since_seconds=3600] - 最近N秒的事件
   */
  events(params) {
    return http.post(`${K8S_BASE}/deployment/events`, {
      namespace: params.namespace,
      kind: 'Pod',
      name: params.name,
      type: params.type || '',
      limit: params.limit || 50,
      since_seconds: params.since_seconds || 3600,
    })
  },

  // =========================
  // 镜像相关
  // =========================

  /**
   * Patch 容器镜像
   * 基于 mergeKey=name 的 StrategicMergePatch 方式更新指定容器的镜像
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Pod名称
   * @param {string} data.container - 容器名称
   * @param {string} data.new_image - 新镜像地址
   */
  patchImage(data) {
    return http.put(`${K8S_BASE}/pod/patch_image`, data)
  },

  // =========================
  // 容器相关
  // =========================
  container: {
    /**
     * 获取 Pod 的容器名列表
     * @param {Object} params
     * @param {string} params.namespace - 命名空间
     * @param {string} params.name - Pod名称
     */
    names(params) {
      return http.get(`${K8S_BASE}/pod/container_name`, {params})
    },

    /**
     * 获取 Pod 的容器镜像列表
     * @param {Object} params
     * @param {string} params.namespace - 命名空间
     * @param {string} params.name - Pod名称
     */
    images(params) {
      return http.get(`${K8S_BASE}/pod/container_image`, {params})
    },

    /**
     * 获取容器日志（一次性返回）
     * @param {Object} params
     * @param {string} params.namespace - 命名空间
     * @param {string} params.name - Pod名称
     * @param {string} [params.container] - 容器名（多容器时建议指定）
     * @param {number} [params.tail] - 返回最后N行
     */
    logs(params) {
      return http.get(`${K8S_BASE}/pod/container_logs`, {params})
    },

    /**
     * 实时跟随 Pod 容器日志（流式）
     * @param {Object} params
     * @param {string} params.namespace - 命名空间
     * @param {string} params.name - Pod名称
     * @param {string} [params.container] - 容器名（多容器时建议指定）
     * @param {number} [params.tail] - 返回最后N行
     * @param {boolean} [params.follow] - 是否实时跟随
     */
    followLog(params) {
      return http.get(`${K8S_BASE}/pod/container_log`, {params})
    },

    // =========================
    // init 容器
    // =========================

    /**
     * 获取 Pod 的 Init 容器名列表
     * @param {Object} params
     * @param {string} params.namespace - 命名空间
     * @param {string} params.name - Pod名称
     */
    initNames(params) {
      return http.get(`${K8S_BASE}/pod/init_container_name`, {params})
    },

    /**
     * 获取 Pod 的 Init 容器镜像列表
     * @param {Object} params
     * @param {string} params.namespace - 命名空间
     * @param {string} params.name - Pod名称
     */
    initImages(params) {
      return http.get(`${K8S_BASE}/pod/init_container_image`, {params})
    },
  },

  // =========================
  // 驱逐 Pod
  // =========================

  /**
   * 驱逐指定 Pod（受 PDB 约束）
   * 相比直接删除更安全，会检查 PodDisruptionBudget
   * 注意：后端字段名必须为 podName（非 name）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.podName - Pod名称
   */
  evict(data) {
    return http.post(`${K8S_BASE}/pod/evict`, {
      namespace: data.namespace,
      podName: data.podName || data.name,  // 兼容 name 参数
    })
  },

  // =========================
  // 便捷方法（简化调用）
  // =========================

  /**
   * 获取容器日志（便捷方法，直接调用 container.logs）
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Pod名称
   * @param {string} [params.container] - 容器名
   * @param {number} [params.tail] - 返回最后N行
   */
  logs(params) {
    return http.get(`${K8S_BASE}/pod/container_logs`, {params})
  },

  /**
   * 更新 Pod 镜像（便捷方法）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Pod名称
   * @param {string} data.container - 容器名称
   * @param {string} data.image - 新镜像地址
   */
  updateImage(data) {
    return http.post(`${K8S_BASE}/pod/patch_image`, {
      namespace: data.namespace,
      name: data.name,
      container: data.container,
      new_image: data.image
    })
  },

  // =========================
  // 资源消耗（Metrics）
  // =========================

  /**
   * 获取单个 Pod 的资源使用情况（需要 metrics-server）
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Pod名称
   * @returns {Promise} - { pod_name, namespace, total_cpu, total_memory, containers: [{ name, cpu, memory }] }
   */
  metrics(params) {
    return http.get(`${K8S_BASE}/pod/metrics`, {params})
  },

  /**
   * 批量获取命名空间下所有 Pod 的资源使用情况
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @returns {Promise} - { "pod-name": { metrics... }, ... }
   */
  metricsList(params) {
    return http.get(`${K8S_BASE}/pod/metrics/list`, {params})
  },

  // =========================
  // 标签管理
  // =========================

  /**
   * 修改 Pod 标签
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Pod名称
   * @param {Object} [data.add] - 要添加/更新的标签 { key: value }
   * @param {Array} [data.remove] - 要删除的标签 key 数组
   */
  patchLabels(data) {
    return http.patch(`${K8S_BASE}/pod/labels`, data)
  },

  // =========================
  // YAML 查看/编辑
  // =========================

  /**
   * 获取 Pod 的 YAML
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Pod名称
   */
  yaml(params) {
    return http.get(`${K8S_BASE}/pod/yaml`, {params})
  },

  /**
   * 应用 Pod YAML 修改
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Pod名称
   * @param {string} data.yaml - YAML内容
   */
  applyYaml(data) {
    return http.put(`${K8S_BASE}/pod/apply_yaml`, data)
  },
}

export default podsApi
