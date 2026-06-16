<template>
  <div class="pagination-wrapper" v-if="totalItems > 0">
    <a-pagination
      :current="currentPage"
      :total="totalItems"
      :page-size="itemsPerPage"
      show-total
      show-jumper
      show-page-size
      :page-size-options="pageSizeOptions"
      @change="handlePageChange"
      @page-size-change="handlePageSizeChange"
    />
  </div>
</template>

<script setup>
import { computed, watch } from 'vue'

const props = defineProps({
  currentPage: { type: Number, default: 1 },
  totalItems: { type: Number, default: 0 },
  itemsPerPage: { type: Number, default: 10 },
  pageSizeOptions: { 
    type: Array, 
    default: () => [10, 20, 50, 100] 
  },
})

const emit = defineEmits(['update:currentPage', 'update:itemsPerPage'])

// 计算总页数
const totalPages = computed(() => {
  const per = Math.max(1, Number(props.itemsPerPage || 10))
  const total = Math.max(0, Number(props.totalItems || 0))
  return Math.max(1, Math.ceil(total / per))
})

// 修正后的当前页
const currentPage = computed(() => {
  const p = Number(props.currentPage || 1)
  return Math.min(Math.max(1, p), totalPages.value)
})

// 越界自动修正
watch(
  () => [props.totalItems, props.itemsPerPage, props.currentPage],
  () => {
    if (props.currentPage !== currentPage.value) {
      emit('update:currentPage', currentPage.value)
    }
  }
)

// 页码变化
const handlePageChange = (page) => {
  emit('update:currentPage', page)
}

// 每页条数变化
const handlePageSizeChange = (pageSize) => {
  emit('update:itemsPerPage', pageSize)
  // 切换每页条数时回到第一页
  emit('update:currentPage', 1)
}
</script>

<style scoped>
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 0.75rem 0.875rem;
  border-top: 1px solid #eef2f7;
  background: #fff;
}

/* 自定义 Arco Design 分页样式，与项目主题色一致 */
:deep(.arco-pagination-item-active) {
  background-color: #326ce5 !important;
  color: #fff !important;
}

:deep(.arco-pagination-item:hover:not(.arco-pagination-item-active)) {
  color: #326ce5;
}

:deep(.arco-pagination-jumper-input:focus) {
  border-color: #326ce5;
}

/* 响应式：小屏幕隐藏部分元素 */
@media (max-width: 768px) {
  :deep(.arco-pagination-jumper) {
    display: none;
  }
  
  :deep(.arco-pagination-size-selector) {
    display: none;
  }
}
</style>
