<template>
  <component :is="iconMap[name]" :style="{ fontSize: size + 'px' }" v-if="iconMap[name]" />
  <span v-else :style="{ fontSize: size + 'px' }">{{ fallback || '?' }}</span>
</template>

<script setup>
import { shallowRef } from 'vue'

import IconHome from '@arco-design/web-vue/es/icon/icon-home'
import IconApps from '@arco-design/web-vue/es/icon/icon-apps'
import IconDesktop from '@arco-design/web-vue/es/icon/icon-desktop'
import IconStorage from '@arco-design/web-vue/es/icon/icon-storage'
import IconThunderbolt from '@arco-design/web-vue/es/icon/icon-thunderbolt'
import IconSettings from '@arco-design/web-vue/es/icon/icon-settings'
import IconSync from '@arco-design/web-vue/es/icon/icon-sync'
import IconFile from '@arco-design/web-vue/es/icon/icon-file'
import IconGift from '@arco-design/web-vue/es/icon/icon-gift'
import IconSend from '@arco-design/web-vue/es/icon/icon-send'
import IconCalendar from '@arco-design/web-vue/es/icon/icon-calendar'
import IconCheckCircle from '@arco-design/web-vue/es/icon/icon-check-circle'
import IconLanguage from '@arco-design/web-vue/es/icon/icon-language'
import IconEdit from '@arco-design/web-vue/es/icon/icon-edit'
import IconDriveFile from '@arco-design/web-vue/es/icon/icon-drive-file'
import IconTool from '@arco-design/web-vue/es/icon/icon-tool'
import IconSafe from '@arco-design/web-vue/es/icon/icon-safe'
import IconSearch from '@arco-design/web-vue/es/icon/icon-search'
import IconUser from '@arco-design/web-vue/es/icon/icon-user'
import IconUserGroup from '@arco-design/web-vue/es/icon/icon-user-group'
import IconLock from '@arco-design/web-vue/es/icon/icon-lock'
import IconUnlock from '@arco-design/web-vue/es/icon/icon-unlock'
import IconList from '@arco-design/web-vue/es/icon/icon-list'
import IconLayers from '@arco-design/web-vue/es/icon/icon-layers'
import IconTag from '@arco-design/web-vue/es/icon/icon-tag'
import IconCommand from '@arco-design/web-vue/es/icon/icon-command'
import IconCode from '@arco-design/web-vue/es/icon/icon-code'
import IconPlayArrow from '@arco-design/web-vue/es/icon/icon-play-arrow'
import IconPause from '@arco-design/web-vue/es/icon/icon-pause'
import IconStop from '@arco-design/web-vue/es/icon/icon-stop'
import IconDelete from '@arco-design/web-vue/es/icon/icon-delete'
import IconPlus from '@arco-design/web-vue/es/icon/icon-plus'
import IconMinus from '@arco-design/web-vue/es/icon/icon-minus'
import IconCopy from '@arco-design/web-vue/es/icon/icon-copy'
import IconExport from '@arco-design/web-vue/es/icon/icon-export'
import IconImport from '@arco-design/web-vue/es/icon/icon-import'
import IconRefresh from '@arco-design/web-vue/es/icon/icon-refresh'
import IconEye from '@arco-design/web-vue/es/icon/icon-eye'
import IconEyeInvisible from '@arco-design/web-vue/es/icon/icon-eye-invisible'
import IconInfoCircle from '@arco-design/web-vue/es/icon/icon-info-circle'
import IconExclamationCircle from '@arco-design/web-vue/es/icon/icon-exclamation-circle'
import IconCloseCircle from '@arco-design/web-vue/es/icon/icon-close-circle'
import IconClockCircle from '@arco-design/web-vue/es/icon/icon-clock-circle'
import IconCloud from '@arco-design/web-vue/es/icon/icon-cloud'
import IconWifi from '@arco-design/web-vue/es/icon/icon-wifi'
import IconBulb from '@arco-design/web-vue/es/icon/icon-bulb'
import IconFire from '@arco-design/web-vue/es/icon/icon-fire'
import IconStar from '@arco-design/web-vue/es/icon/icon-star'
import IconHeart from '@arco-design/web-vue/es/icon/icon-heart'
import IconLink from '@arco-design/web-vue/es/icon/icon-link'
import IconBranch from '@arco-design/web-vue/es/icon/icon-branch'
import IconExperiment from '@arco-design/web-vue/es/icon/icon-experiment'
import IconDashboard from '@arco-design/web-vue/es/icon/icon-dashboard'
import IconHistory from '@arco-design/web-vue/es/icon/icon-history'
import IconBook from '@arco-design/web-vue/es/icon/icon-book'
import IconEmpty from '@arco-design/web-vue/es/icon/icon-empty'
import IconNotification from '@arco-design/web-vue/es/icon/icon-notification'
import IconMessage from '@arco-design/web-vue/es/icon/icon-message'
import IconRobot from '@arco-design/web-vue/es/icon/icon-robot'
import IconImage from '@arco-design/web-vue/es/icon/icon-image'
import IconMenu from '@arco-design/web-vue/es/icon/icon-menu'
import IconMenuFold from '@arco-design/web-vue/es/icon/icon-menu-fold'
import IconMenuUnfold from '@arco-design/web-vue/es/icon/icon-menu-unfold'
import IconDown from '@arco-design/web-vue/es/icon/icon-down'
import IconUp from '@arco-design/web-vue/es/icon/icon-up'
import IconRight from '@arco-design/web-vue/es/icon/icon-right'
import IconLeft from '@arco-design/web-vue/es/icon/icon-left'
import IconSwap from '@arco-design/web-vue/es/icon/icon-swap'
import IconFilter from '@arco-design/web-vue/es/icon/icon-filter'
import IconOrderedList from '@arco-design/web-vue/es/icon/icon-ordered-list'
import IconSubscribe from '@arco-design/web-vue/es/icon/icon-subscribe'
import IconShareInternal from '@arco-design/web-vue/es/icon/icon-share-internal'
import IconPublic from '@arco-design/web-vue/es/icon/icon-public'
import IconShake from '@arco-design/web-vue/es/icon/icon-shake'
import IconBug from '@arco-design/web-vue/es/icon/icon-bug'
import IconThumbUp from '@arco-design/web-vue/es/icon/icon-thumb-up'
import IconLoading from '@arco-design/web-vue/es/icon/icon-loading'

const props = defineProps({
  name: { type: String, required: true },
  size: { type: Number, default: 18 }
})

const iconMap = shallowRef({
  home: IconHome,
  apps: IconApps,
  desktop: IconDesktop,
  storage: IconStorage,
  thunderbolt: IconThunderbolt,
  settings: IconSettings,
  sync: IconSync,
  file: IconFile,
  gift: IconGift,
  send: IconSend,
  calendar: IconCalendar,
  'check-circle': IconCheckCircle,
  language: IconLanguage,
  edit: IconEdit,
  'drive-file': IconDriveFile,
  tool: IconTool,
  safe: IconSafe,
  search: IconSearch,
  user: IconUser,
  'user-group': IconUserGroup,
  lock: IconLock,
  unlock: IconUnlock,
  list: IconList,
  layers: IconLayers,
  tag: IconTag,
  command: IconCommand,
  code: IconCode,
  'play-arrow': IconPlayArrow,
  pause: IconPause,
  stop: IconStop,
  delete: IconDelete,
  plus: IconPlus,
  minus: IconMinus,
  copy: IconCopy,
  export: IconExport,
  import: IconImport,
  refresh: IconRefresh,
  eye: IconEye,
  'eye-invisible': IconEyeInvisible,
  'info-circle': IconInfoCircle,
  'exclamation-circle': IconExclamationCircle,
  'close-circle': IconCloseCircle,
  'clock-circle': IconClockCircle,
  cloud: IconCloud,
  wifi: IconWifi,
  bulb: IconBulb,
  fire: IconFire,
  star: IconStar,
  heart: IconHeart,
  link: IconLink,
  branch: IconBranch,
  experiment: IconExperiment,
  dashboard: IconDashboard,
  history: IconHistory,
  book: IconBook,
  empty: IconEmpty,
  notification: IconNotification,
  message: IconMessage,
  robot: IconRobot,
  image: IconImage,
  menu: IconMenu,
  'menu-fold': IconMenuFold,
  'menu-unfold': IconMenuUnfold,
  down: IconDown,
  up: IconUp,
  right: IconRight,
  left: IconLeft,
  swap: IconSwap,
  filter: IconFilter,
  'ordered-list': IconOrderedList,
  subscribe: IconSubscribe,
  'share-internal': IconShareInternal,
  public: IconPublic,
  shake: IconShake,
  bug: IconBug,
  'thumb-up': IconThumbUp,
  loading: IconLoading
})
</script>
