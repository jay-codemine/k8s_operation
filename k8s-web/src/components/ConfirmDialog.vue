<template>
  <Teleport to="body">
    <transition name="confirm-overlay">
      <div v-if="visible" class="confirm-overlay" @click.self="handleCancel">
        <transition name="confirm-dialog">
          <div v-if="visible" class="confirm-dialog" :class="dialogState.type">
            <!-- 顶部图标区 -->
            <div class="confirm-icon-area" :class="dialogState.type">
              <div class="confirm-icon-circle">
                <!-- Warning -->
                <svg v-if="dialogState.type === 'warning'" width="28" height="28" viewBox="0 0 24 24" fill="none">
                  <path d="M12 9v4m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                <!-- Danger -->
                <svg v-else-if="dialogState.type === 'danger'" width="28" height="28" viewBox="0 0 24 24" fill="none">
                  <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 11c-.55 0-1-.45-1-1V8c0-.55.45-1 1-1s1 .45 1 1v4c0 .55-.45 1-1 1zm1 4h-2v-2h2v2z" fill="currentColor"/>
                </svg>
                <!-- Info -->
                <svg v-else-if="dialogState.type === 'info'" width="28" height="28" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/>
                  <path d="M12 16v-4m0-4h.01" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                </svg>
                <!-- Success -->
                <svg v-else width="28" height="28" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/>
                  <path d="M8 12l3 3 5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </div>
            </div>

            <!-- 标题 -->
            <h3 class="confirm-title">{{ dialogState.title }}</h3>

            <!-- 内容描述 -->
            <p v-if="dialogState.content" class="confirm-content">{{ dialogState.content }}</p>

            <!-- 详情列表 -->
            <div v-if="dialogState.details.length > 0" class="confirm-details">
              <div
                v-for="(item, i) in dialogState.details"
                :key="i"
                class="confirm-detail-row"
              >
                <span class="detail-label">{{ item.label }}</span>
                <span class="detail-value" :class="{ highlight: item.highlight, danger: item.danger, mono: item.mono !== false }">
                  {{ item.value }}
                </span>
              </div>
            </div>

            <!-- 底部提示 -->
            <p v-if="dialogState.tip" class="confirm-tip" :class="dialogState.type">
              {{ dialogState.tip }}
            </p>

            <!-- 操作按钮 -->
            <div class="confirm-actions">
              <button class="confirm-btn cancel" @click="handleCancel">
                {{ dialogState.cancelText }}
              </button>
              <button class="confirm-btn primary" :class="dialogState.type" @click="handleConfirm">
                {{ dialogState.confirmText }}
              </button>
            </div>
          </div>
        </transition>
      </div>
    </transition>
  </Teleport>
</template>

<script setup>
import { useConfirmDialog } from '@/composables/useConfirmDialog'

const { visible, dialogState, handleConfirm, handleCancel } = useConfirmDialog()
</script>

<style scoped>
/* ========== 遮罩层 ========== */
.confirm-overlay {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.55);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
}

/* ========== 弹窗主体 ========== */
.confirm-dialog {
  width: 420px;
  max-width: 90vw;
  background: #1e2030;
  border-radius: 20px;
  padding: 32px 28px 24px;
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.6),
    0 0 0 1px rgba(255, 255, 255, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.06);
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  position: relative;
  overflow: hidden;
}

/* 顶部装饰光效 */
.confirm-dialog::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 200px;
  height: 2px;
  border-radius: 2px;
}
.confirm-dialog.warning::before {
  background: linear-gradient(90deg, transparent, #f0a020, transparent);
}
.confirm-dialog.danger::before {
  background: linear-gradient(90deg, transparent, #ff4d4f, transparent);
}
.confirm-dialog.info::before {
  background: linear-gradient(90deg, transparent, #4e8ff7, transparent);
}
.confirm-dialog.success::before {
  background: linear-gradient(90deg, transparent, #52c41a, transparent);
}

/* ========== 图标区域 ========== */
.confirm-icon-area {
  margin-bottom: 16px;
}

.confirm-icon-circle {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.confirm-icon-area.warning .confirm-icon-circle {
  background: rgba(240, 160, 32, 0.12);
  color: #f0a020;
  box-shadow: 0 0 0 8px rgba(240, 160, 32, 0.06);
}
.confirm-icon-area.danger .confirm-icon-circle {
  background: rgba(255, 77, 79, 0.12);
  color: #ff4d4f;
  box-shadow: 0 0 0 8px rgba(255, 77, 79, 0.06);
}
.confirm-icon-area.info .confirm-icon-circle {
  background: rgba(78, 143, 247, 0.12);
  color: #4e8ff7;
  box-shadow: 0 0 0 8px rgba(78, 143, 247, 0.06);
}
.confirm-icon-area.success .confirm-icon-circle {
  background: rgba(82, 196, 26, 0.12);
  color: #52c41a;
  box-shadow: 0 0 0 8px rgba(82, 196, 26, 0.06);
}

/* ========== 标题 ========== */
.confirm-title {
  font-size: 17px;
  font-weight: 700;
  color: #e0e0e0;
  margin: 0 0 8px;
  line-height: 1.4;
}

/* ========== 内容 ========== */
.confirm-content {
  font-size: 13px;
  color: #8b95b0;
  margin: 0 0 16px;
  line-height: 1.6;
  max-width: 340px;
}

/* ========== 详情列表 ========== */
.confirm-details {
  width: 100%;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 12px;
  padding: 12px 16px;
  margin-bottom: 16px;
  text-align: left;
}

.confirm-detail-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 6px 0;
  gap: 12px;
}

.confirm-detail-row + .confirm-detail-row {
  border-top: 1px solid rgba(255, 255, 255, 0.04);
}

.detail-label {
  font-size: 12px;
  color: #6b7394;
  flex-shrink: 0;
  min-width: 60px;
  font-weight: 500;
}

.detail-value {
  font-size: 12px;
  color: #c0caf5;
  text-align: right;
  word-break: break-all;
  line-height: 1.5;
}
.detail-value.mono {
  font-family: 'JetBrains Mono', 'SF Mono', 'Consolas', monospace;
}
.detail-value.highlight {
  color: #4e8ff7;
  font-weight: 600;
}
.detail-value.danger {
  color: #ff4d4f;
  font-weight: 600;
}

/* ========== 底部提示 ========== */
.confirm-tip {
  font-size: 12px;
  margin: 0 0 20px;
  padding: 8px 14px;
  border-radius: 8px;
  line-height: 1.5;
  width: 100%;
  text-align: left;
}
.confirm-tip.warning {
  color: #f0a020;
  background: rgba(240, 160, 32, 0.08);
  border: 1px solid rgba(240, 160, 32, 0.15);
}
.confirm-tip.danger {
  color: #ff4d4f;
  background: rgba(255, 77, 79, 0.08);
  border: 1px solid rgba(255, 77, 79, 0.15);
}
.confirm-tip.info {
  color: #4e8ff7;
  background: rgba(78, 143, 247, 0.08);
  border: 1px solid rgba(78, 143, 247, 0.15);
}
.confirm-tip.success {
  color: #52c41a;
  background: rgba(82, 196, 26, 0.08);
  border: 1px solid rgba(82, 196, 26, 0.15);
}

/* ========== 操作按钮 ========== */
.confirm-actions {
  display: flex;
  gap: 12px;
  width: 100%;
  margin-top: 4px;
}

.confirm-btn {
  flex: 1;
  height: 40px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  border: none;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  align-items: center;
  justify-content: center;
  letter-spacing: 0.3px;
}

.confirm-btn.cancel {
  background: rgba(255, 255, 255, 0.06);
  color: #8b95b0;
  border: 1px solid rgba(255, 255, 255, 0.08);
}
.confirm-btn.cancel:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #c0caf5;
  border-color: rgba(255, 255, 255, 0.15);
}

.confirm-btn.primary {
  color: #fff;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.2);
}
.confirm-btn.primary.warning {
  background: linear-gradient(135deg, #f0a020, #e68a00);
}
.confirm-btn.primary.warning:hover {
  background: linear-gradient(135deg, #f5b030, #f0a020);
  box-shadow: 0 6px 20px rgba(240, 160, 32, 0.3);
  transform: translateY(-1px);
}
.confirm-btn.primary.danger {
  background: linear-gradient(135deg, #ff4d4f, #e8383b);
}
.confirm-btn.primary.danger:hover {
  background: linear-gradient(135deg, #ff6b6d, #ff4d4f);
  box-shadow: 0 6px 20px rgba(255, 77, 79, 0.3);
  transform: translateY(-1px);
}
.confirm-btn.primary.info {
  background: linear-gradient(135deg, #4e8ff7, #3b7de8);
}
.confirm-btn.primary.info:hover {
  background: linear-gradient(135deg, #6ba3f9, #4e8ff7);
  box-shadow: 0 6px 20px rgba(78, 143, 247, 0.3);
  transform: translateY(-1px);
}
.confirm-btn.primary.success {
  background: linear-gradient(135deg, #52c41a, #3ba00e);
}
.confirm-btn.primary.success:hover {
  background: linear-gradient(135deg, #6dd230, #52c41a);
  box-shadow: 0 6px 20px rgba(82, 196, 26, 0.3);
  transform: translateY(-1px);
}

.confirm-btn:active {
  transform: translateY(0) scale(0.98);
}

/* ========== 动画 ========== */
.confirm-overlay-enter-active {
  transition: opacity 0.25s ease;
}
.confirm-overlay-leave-active {
  transition: opacity 0.2s ease;
}
.confirm-overlay-enter-from,
.confirm-overlay-leave-to {
  opacity: 0;
}

.confirm-dialog-enter-active {
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.confirm-dialog-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 1, 1);
}
.confirm-dialog-enter-from {
  opacity: 0;
  transform: scale(0.85) translateY(20px);
}
.confirm-dialog-leave-to {
  opacity: 0;
  transform: scale(0.92) translateY(10px);
}
</style>
