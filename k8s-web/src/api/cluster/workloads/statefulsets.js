// src/api/cluster/workloads/statefulsets.js
import http from '@/api/http'
import {K8S_BASE} from '@/api/paths'

// =========================
// StatefulSet 模块 API（全量）
// 对应后端路由: /api/v1/k8s/statefulset/*
// =========================
const statefulsetsApi = {
  // =========================
  // StatefulSet 基础 CRUD
  // =========================

  /**
   * 创建 StatefulSet（可选创建 Service）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - StatefulSet 名称
   * @param {string} data.container_image - 容器镜像
   * @param {number} [data.replicas=1] - 副本数
   * @param {Object} [data.labels] - 标签
   * @param {boolean} [data.is_create_service] - 是否同时创建 Service
   * @param {string} [data.service_name] - Headless Service 名称
   */
  create(data) {
    return http.post(`${K8S_BASE}/statefulset/create`, data)
  },

  /**
   * StatefulSet 列表
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {number} params.page - 页码
   * @param {number} params.limit - 每页数量
   * @param {string} [params.name] - StatefulSet 名称（模糊匹配）
   */
  list(params) {
    return http.get(`${K8S_BASE}/statefulset/list`, {params})
  },

  /**
   * StatefulSet 详情
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - StatefulSet 名称
   */
  detail(params) {
    return http.get(`${K8S_BASE}/statefulset/detail`, {params})
  },

  /**
   * 删除 StatefulSet
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - StatefulSet 名称
   */
  delete(params) {
    return http.delete(`${K8S_BASE}/statefulset/delete`, {params})
  },

  /**
   * 删除 StatefulSet 对应的 Service
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - StatefulSet 名称（用于查找关联的 Service）
   */
  deleteService(params) {
    return http.delete(`${K8S_BASE}/statefulset/delete_svc`, {params})
  },

  // =========================
  // 扩缩容
  // =========================

  /**
   * 扩缩容（修改副本数）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - StatefulSet 名称
   * @param {number} data.scale_num - 目标副本数
   */
  scale(data) {
    return http.put(`${K8S_BASE}/statefulset/scale`, data)
  },

  // =========================
  // 镜像更新
  // =========================

  /**
   * 更新镜像（触发滚动升级）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - StatefulSet 名称
   * @param {string} data.container - 容器名称
   * @param {string} data.image - 新镜像地址
   */
  updateImage(data) {
    return http.put(`${K8S_BASE}/statefulset/update_image`, data)
  },

  // =========================
  // 重启
  // =========================

  /**
   * 滚动重启 StatefulSet
   * 通过更新 annotation 触发滚动更新
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - StatefulSet 名称
   */
  restart(data) {
    return http.post(`${K8S_BASE}/statefulset/restart`, data)
  },

  // =========================
  // 关联资源
  // =========================

  /**
   * 获取 StatefulSet 对应的 Pod 列表
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - StatefulSet 名称
   */
  pods(params) {
    return http.get(`${K8S_BASE}/statefulset/sts_pods`, {params})
  },

  /**
   * 获取 StatefulSet 相关事件
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - StatefulSet 名称
   * @param {string} [params.type] - 事件类型（Normal | Warning）
   * @param {number} [params.limit=50] - 返回条数限制
   * @param {number} [params.since_seconds=3600] - 最近N秒的事件
   */
  events(params) {
    return http.post(`${K8S_BASE}/statefulset/events`, {
      namespace: params.namespace,
      kind: 'StatefulSet',
      name: params.name,
      type: params.type || '',
      limit: params.limit || 50,
      since_seconds: params.since_seconds || 3600,
    })
  },

  // =========================
  // 历史版本
  // =========================

  /**
   * 获取 StatefulSet 的历史版本（ControllerRevision 列表）
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - StatefulSet 名称
   */
  history(params) {
    return http.get(`${K8S_BASE}/statefulset/history`, {params})
  },

  /**
   * 回滚到指定 ControllerRevision
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - StatefulSet 名称
   * @param {string} data.revision_name - 目标版本名称（ControllerRevision 名称）
   */
  rollback(data) {
    return http.post(`${K8S_BASE}/statefulset/rollback`, data)
  },

  // =========================
  // YAML 查看/编辑
  // =========================

  /**
   * 获取 StatefulSet 的 YAML 配置
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - StatefulSet 名称
   */
  yaml(params) {
    return http.get(`${K8S_BASE}/statefulset/yaml`, {params})
  },

  /**
   * 应用 YAML 更改
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - StatefulSet 名称
   * @param {string} data.yaml - YAML 内容
   */
  applyYaml(data) {
    return http.put(`${K8S_BASE}/statefulset/apply_yaml`, data)
  },

  /**
   * 从 YAML 创建 StatefulSet
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  createFromYaml(data) {
    return http.post(`${K8S_BASE}/statefulset/create-from-yaml`, data)
  },
}

export default statefulsetsApi
