<template>
  <div class="sa-view">
    <div class="sa-header">
      <div class="sa-header-left">
        <div class="sa-icon">🔐</div>
        <div class="sa-title-group">
          <h1>ServiceAccount 管理</h1>
          <p>管理 Kubernetes 服务账户，用于 Pod 身份认证与权限控制</p>
        </div>
      </div>
      <div class="sa-header-right">
        <ClusterSelector v-model="selectedClusterId" :clusters="clusters" :show-all-option="false" label="集群" @change="onClusterChange" />
      </div>
    </div>

    <div class="sa-stats">
      <div class="stat-card primary"><div class="stat-icon">👤</div><div class="stat-value">{{ serviceAccounts.length }}</div><div class="stat-label">SA 总数</div></div>
      <div class="stat-card success"><div class="stat-icon">🔑</div><div class="stat-value">{{ totalSecretsCount }}</div><div class="stat-label">关联 Secrets</div></div>
      <div class="stat-card info"><div class="stat-icon">📁</div><div class="stat-value">{{ uniqueNamespaces }}</div><div class="stat-label">命名空间数</div></div>
      <div class="stat-card warning"><div class="stat-icon">🔄</div><div class="stat-value">{{ autoMountCount }}</div><div class="stat-label">自动挂载 Token</div></div>
    </div>

    <div class="sa-toolbar">
      <div class="toolbar-left">
        <div class="sa-search"><span>🔍</span><input v-model="searchQuery" placeholder="搜索名称..." /></div>
        <select class="ns-select" v-model="namespaceFilter" @change="loadServiceAccounts">
          <option value="">所有命名空间</option>
          <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
        </select>
        <div class="view-toggle">
          <button :class="['vbtn', viewMode==='table'?'active':'']" @click="viewMode='table'">☰</button>
          <button :class="['vbtn', viewMode==='card'?'active':'']"  @click="viewMode='card'">⊞</button>
        </div>
        <label class="auto-ref"><input type="checkbox" v-model="autoRefresh" /><span>自动刷新</span><span v-if="autoRefresh" class="ref-dot">●</span></label>
      </div>
      <div class="toolbar-right">
        <button class="sa-btn sec" @click="loadServiceAccounts" :disabled="loading">{{ loading?'⏳':'sync' }} 刷新</button>
        <button class="sa-btn pri" @click="openCreateModal">＋ 创建 ServiceAccount</button>
      </div>
    </div>

    <div v-if="loading && serviceAccounts.length===0" class="sa-loading"><div class="spin"></div><span>加载中...</span></div>

    <!-- 表格视图 -->
    <div v-else-if="viewMode==='table'" class="sa-table-wrap">
      <table class="sa-table">
        <thead><tr><th>名称</th><th>命名空间</th><th>Secrets</th><th>自动挂载</th><th>标签数</th><th>创建时间</th><th>操作</th></tr></thead>
        <tbody>
          <tr v-for="sa in pagedItems" :key="`${sa.namespace}/${sa.name}`" class="sa-row">
            <td>
              <div class="name-cell">
                <span class="sa-av">👤</span>
                <div><div class="nm">{{ sa.name }}</div><div class="meta">{{ sa.secrets?.length||0 }} Secrets</div></div>
              </div>
            </td>
            <td><span class="ns-tag">{{ sa.namespace }}</span></td>
            <td><span class="bdg info">{{ sa.secrets?.length||0 }}</span></td>
            <td><span class="bdg" :class="sa.automount_token?'suc':'def'">{{ sa.automount_token?'启用':'禁用' }}</span></td>
            <td><span class="bdg info">{{ Object.keys(sa.labels||{}).length }}</span></td>
            <td class="tc">{{ formatDate(sa.created_at) }}</td>
            <td>
              <div class="act-grp">
                <button class="ab view"   @click="openDetail(sa)"       title="详情">👁️</button>
                <button class="ab edit"   @click="openEditModal(sa)"    title="编辑">✏️</button>
                <button class="ab yaml"   @click="openYamlModal(sa)"    title="YAML">📄</button>
                <button class="ab ev"     @click="openEventModal(sa)"   title="事件">📋</button>
                <button class="ab del"    @click="confirmDelete(sa)"    title="删除">🗑️</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="filteredItems.length===0&&!loading" class="sa-empty">
        <div class="ei">👤</div><div class="et">暂无 ServiceAccount</div>
        <div class="ed">{{ searchQuery?'没有匹配结果':'点击上方按钮创建' }}</div>
        <button v-if="!searchQuery" class="sa-btn pri" @click="openCreateModal">＋ 创建</button>
      </div>
      <div v-if="filteredItems.length>pageSize" class="sa-page">
        <button class="pb" :disabled="currentPage<=1" @click="currentPage--">‹ 上一页</button>
        <span class="pi">第 {{ currentPage }}/{{ totalPages }} 页（共 {{ filteredItems.length }} 条）</span>
        <button class="pb" :disabled="currentPage>=totalPages" @click="currentPage++">下一页 ›</button>
      </div>
    </div>

    <!-- 卡片视图 -->
    <div v-else class="sa-cards">
      <div v-for="sa in pagedItems" :key="`${sa.namespace}/${sa.name}`" class="sa-card">
        <div class="ch"><div class="ci">👤</div><div><div class="cn">{{ sa.name }}</div><span class="ns-tag">{{ sa.namespace }}</span></div></div>
        <div class="cb">
          <div class="cr"><span class="cl">自动挂载</span><span class="bdg" :class="sa.automount_token?'suc':'def'">{{ sa.automount_token?'启用':'禁用' }}</span></div>
          <div class="cr"><span class="cl">Secrets</span><span class="bdg info">{{ sa.secrets?.length||0 }}</span></div>
          <div class="cr"><span class="cl">Labels</span><span class="bdg info">{{ Object.keys(sa.labels||{}).length }}</span></div>
          <div class="cr"><span class="cl">创建时间</span><span class="cv">{{ formatDate(sa.created_at) }}</span></div>
        </div>
        <div class="cf">
          <button class="ab view" @click="openDetail(sa)" title="详情">👁️</button>
          <button class="ab edit" @click="openEditModal(sa)" title="编辑">✏️</button>
          <button class="ab yaml" @click="openYamlModal(sa)" title="YAML">📄</button>
          <button class="ab ev"   @click="openEventModal(sa)" title="事件">📋</button>
          <button class="ab del"  @click="confirmDelete(sa)" title="删除">🗑️</button>
        </div>
      </div>
      <div v-if="filteredItems.length===0&&!loading" class="sa-empty fw">
        <div class="ei">👤</div><div class="et">暂无 ServiceAccount</div>
      </div>
      <div v-if="filteredItems.length>pageSize" class="sa-page fw">
        <button class="pb" :disabled="currentPage<=1" @click="currentPage--">‹ 上一页</button>
        <span class="pi">第 {{ currentPage }}/{{ totalPages }} 页（共 {{ filteredItems.length }} 条）</span>
        <button class="pb" :disabled="currentPage>=totalPages" @click="currentPage++">下一页 ›</button>
      </div>
    </div>

    <!-- ===== 详情抽屉 ===== -->
    <div v-if="showDetail" class="overlay" @click.self="showDetail=false">
      <div class="drawer">
        <div class="dh">
          <div class="dtrow"><span class="di">👤</span><div><div class="dn">{{ currentSA.name }}</div><span class="ns-tag">{{ currentSA.namespace }}</span></div></div>
          <div class="dtabs">
            <button v-for="t in dtabs" :key="t.k" :class="['dtab',activeTab===t.k?'active':'']" @click="switchTab(t.k)">{{ t.l }}</button>
          </div>
          <button class="dc" @click="showDetail=false">✕</button>
        </div>
        <div class="db">
          <div v-if="activeTab==='info'">
            <div class="dgrid">
              <div class="ditem fw"><span class="dl">名称</span><span class="dv mono">{{ currentSA.name }}</span></div>
              <div class="ditem"><span class="dl">命名空间</span><span class="ns-tag">{{ currentSA.namespace }}</span></div>
              <div class="ditem"><span class="dl">自动挂载</span><span class="bdg" :class="currentSA.automount_token?'suc':'def'">{{ currentSA.automount_token?'启用':'禁用' }}</span></div>
              <div class="ditem"><span class="dl">创建时间</span><span class="dv">{{ formatDate(currentSA.created_at) }}</span></div>
              <div class="ditem"><span class="dl">Secrets</span><span class="bdg info">{{ currentSA.secrets?.length||0 }}</span></div>
            </div>
            <div class="dsec"><div class="dst">Labels（{{ Object.keys(currentSA.labels||{}).length }}）</div>
              <div v-if="Object.keys(currentSA.labels||{}).length" class="lbs">
                <span v-for="(v,k) in currentSA.labels" :key="k" class="lc">{{ k }}={{ v }}</span>
              </div>
              <div v-else class="eh">无 Labels</div>
            </div>
          </div>
          <div v-if="activeTab==='secrets'">
            <div class="dsec"><div class="dst">关联 Secrets（{{ currentSA.secrets?.length||0 }}）</div>
              <div v-if="currentSA.secrets?.length" class="sl">
                <div v-for="s in currentSA.secrets" :key="s.name" class="sr">
                  <span>🔑</span><span class="mono">{{ s.name }}</span><span v-if="s.type" class="bdg info ml">{{ s.type }}</span>
                </div>
              </div>
              <div v-else class="eh">暂无关联 Secret</div>
            </div>
          </div>
          <div v-if="activeTab==='events'">
            <div v-if="evLoading" class="lrow"><div class="spin sm"></div> 加载事件中...</div>
            <div v-else-if="saEvents.length===0" class="eh">最近 1 小时内无事件</div>
            <div v-else class="etbl-wrap">
              <table class="etbl"><thead><tr><th>类型</th><th>原因</th><th>消息</th><th>次数</th><th>时间</th></tr></thead>
                <tbody>
                  <tr v-for="(ev,i) in saEvents" :key="i">
                    <td><span class="bdg" :class="ev.type==='Warning'?'warn':'suc'">{{ ev.type }}</span></td>
                    <td class="mono">{{ ev.reason }}</td><td class="em">{{ ev.message }}</td>
                    <td>{{ ev.count }}</td><td class="tc">{{ formatDate(ev.event_time) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
        <div class="df">
          <button class="sa-btn sec" @click="openEditModal(currentSA)">✏️ 编辑</button>
          <button class="sa-btn sec" @click="openYamlModal(currentSA)">📄 YAML</button>
          <button class="sa-btn danger" @click="confirmDelete(currentSA)">🗑️ 删除</button>
        </div>
      </div>
    </div>

    <!-- ===== 创建 Modal ===== -->
    <div v-if="showCreate" class="mbg" @click.self="showCreate=false">
      <div class="modal">
        <div class="mh"><h2>创建 ServiceAccount</h2><button class="mc" @click="showCreate=false">✕</button></div>
        <div class="mb">
          <div class="fg"><label>名称 <span class="req">*</span></label><input v-model="cf.name" placeholder="例如：my-service-account" /></div>
          <div class="fg"><label>命名空间 <span class="req">*</span></label>
            <select v-model="cf.namespace"><option value="">请选择</option><option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option></select>
          </div>
          <div class="fg"><label>自动挂载 Token</label>
            <label class="tw"><input type="checkbox" v-model="cf.autoMount" /><span>{{ cf.autoMount?'启用':'禁用' }}</span></label>
            <p class="ht">启用后，使用此 SA 的 Pod 自动挂载 Token Secret</p>
          </div>
          <div class="fg"><label>Labels（可选）</label>
            <div v-for="(l,i) in cf.labels" :key="i" class="lr">
              <input v-model="l.k" placeholder="key" class="li" /><span class="sep">=</span><input v-model="l.v" placeholder="value" class="li" />
              <button class="rb" @click="cf.labels.splice(i,1)">✕</button>
            </div>
            <button class="sa-btn sec sm" @click="cf.labels.push({k:'',v:''})">＋ 添加 Label</button>
          </div>
        </div>
        <div class="mf">
          <button class="sa-btn sec" @click="showCreate=false">取消</button>
          <button class="sa-btn pri" @click="submitCreate" :disabled="submitting">{{ submitting?'创建中...':'创建' }}</button>
        </div>
      </div>
    </div>

    <!-- ===== 编辑 Modal ===== -->
    <div v-if="showEdit" class="mbg" @click.self="showEdit=false">
      <div class="modal">
        <div class="mh"><h2>编辑 ServiceAccount</h2><button class="mc" @click="showEdit=false">✕</button></div>
        <div class="mb">
          <div class="fg ro"><label>名称</label><span class="rv mono">{{ ef.name }}</span></div>
          <div class="fg ro"><label>命名空间</label><span class="ns-tag">{{ ef.namespace }}</span></div>
          <div class="fg"><label>自动挂载 Token</label>
            <label class="tw"><input type="checkbox" v-model="ef.autoMount" /><span>{{ ef.autoMount?'启用':'禁用' }}</span></label>
          </div>
          <div class="fg"><label>Labels</label>
            <div v-for="(l,i) in ef.labels" :key="i" class="lr">
              <input v-model="l.k" placeholder="key" class="li" /><span class="sep">=</span><input v-model="l.v" placeholder="value" class="li" />
              <button class="rb" @click="ef.labels.splice(i,1)">✕</button>
            </div>
            <button class="sa-btn sec sm" @click="ef.labels.push({k:'',v:''})">＋ 添加</button>
          </div>
          <div class="fg"><label>Annotations</label>
            <div v-for="(a,i) in ef.annotations" :key="i" class="lr">
              <input v-model="a.k" placeholder="key" class="li" /><span class="sep">=</span><input v-model="a.v" placeholder="value" class="li" />
              <button class="rb" @click="ef.annotations.splice(i,1)">✕</button>
            </div>
            <button class="sa-btn sec sm" @click="ef.annotations.push({k:'',v:''})">＋ 添加</button>
          </div>
        </div>
        <div class="mf">
          <button class="sa-btn sec" @click="showEdit=false">取消</button>
          <button class="sa-btn pri" @click="submitEdit" :disabled="submitting">{{ submitting?'保存中...':'保存' }}</button>
        </div>
      </div>
    </div>

    <!-- ===== YAML Modal ===== -->
    <div v-if="showYaml" class="mbg" @click.self="closeYaml">
      <div class="modal wide">
        <div class="mh">
          <h2>YAML — {{ yamlTarget?.name }}</h2>
          <div class="ymtabs"><button :class="['ymtab',yamlMode==='view'?'active':'']" @click="yamlMode='view'">查看</button><button :class="['ymtab',yamlMode==='edit'?'active':'']" @click="yamlMode='edit'">编辑</button></div>
          <button class="mc" @click="closeYaml">✕</button>
        </div>
        <div class="mb yaml-mb">
          <div v-if="yamlLoading" class="lrow"><div class="spin"></div> 加载 YAML...</div>
          <textarea v-else v-model="yamlContent" class="ye" :readonly="yamlMode==='view'" spellcheck="false"></textarea>
        </div>
        <div class="mf">
          <button class="sa-btn sec" @click="copyYaml">📋 复制</button>
          <button v-if="yamlMode==='edit'" class="sa-btn pri" @click="submitYaml" :disabled="submitting">{{ submitting?'应用中...':'✅ 应用 YAML' }}</button>
          <button class="sa-btn sec" @click="closeYaml">关闭</button>
        </div>
      </div>
    </div>

    <!-- ===== 事件 Modal ===== -->
    <div v-if="showEvent" class="mbg" @click.self="showEvent=false">
      <div class="modal wide">
        <div class="mh"><h2>事件 — {{ evTarget?.name }}</h2><button class="mc" @click="showEvent=false">✕</button></div>
        <div class="mb">
          <div v-if="evLoading" class="lrow"><div class="spin"></div> 加载事件...</div>
          <div v-else-if="saEvents.length===0" class="eh">最近 1 小时无事件记录</div>
          <div v-else class="etbl-wrap">
            <table class="etbl">
              <thead><tr><th>类型</th><th>原因</th><th>消息</th><th>来源</th><th>次数</th><th>时间</th></tr></thead>
              <tbody>
                <tr v-for="(ev,i) in saEvents" :key="i">
                  <td><span class="bdg" :class="ev.type==='Warning'?'warn':'suc'">{{ ev.type }}</span></td>
                  <td class="mono">{{ ev.reason }}</td><td class="em">{{ ev.message }}</td>
                  <td class="mono">{{ ev.source_component||'-' }}</td><td>{{ ev.count }}</td>
                  <td class="tc">{{ formatDate(ev.event_time) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
        <div class="mf">
          <button class="sa-btn sec" @click="loadEvents(evTarget)">🔄 刷新</button>
          <button class="sa-btn sec" @click="showEvent=false">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { Message } from '@arco-design/web-vue'
import ClusterSelector from '@/components/cluster/ClusterSelector.vue'
import {
  listServiceAccounts,
  createServiceAccount    as createSAApi,
  updateServiceAccount    as updateSAApi,
  deleteServiceAccount    as deleteSAApi,
  getServiceAccountYaml   as getSAYamlApi,
  applyServiceAccountYaml as applySAYamlApi,
  getServiceAccountEvents as getSAEventsApi,
} from '@/api/k8sRbac'
import { getClusterList } from '@/api/cluster'
import { getNamespaces }  from '@/api/namespace'
import permissionStore    from '@/stores/permission'
import { useConfirmDialog } from '@/composables/useConfirmDialog'

const { confirm: showConfirm } = useConfirmDialog()

const loading = ref(true), submitting = ref(false)
const searchQuery = ref(''), namespaceFilter = ref('')
const serviceAccounts = ref([]), namespaces = ref([])
const viewMode = ref('table'), autoRefresh = ref(false)
let refreshTimer = null
const clusters = ref([]), selectedClusterId = ref('')
const currentPage = ref(1), pageSize = 20

const totalSecretsCount = computed(() => serviceAccounts.value.reduce((s,sa) => s+(sa.secrets?.length||0),0))
const uniqueNamespaces  = computed(() => new Set(serviceAccounts.value.map(sa=>sa.namespace)).size)
const autoMountCount    = computed(() => serviceAccounts.value.filter(sa=>sa.automount_token).length)

const filteredItems = computed(() => {
  let l = serviceAccounts.value
  if (namespaceFilter.value) l = l.filter(sa=>sa.namespace===namespaceFilter.value)
  if (searchQuery.value) { const q=searchQuery.value.toLowerCase(); l=l.filter(sa=>sa.name.toLowerCase().includes(q)||sa.namespace.toLowerCase().includes(q)) }
  return l
})
const totalPages = computed(() => Math.max(1, Math.ceil(filteredItems.value.length/pageSize)))
watch(filteredItems, () => { currentPage.value=1 })
const pagedItems = computed(() => { const s=(currentPage.value-1)*pageSize; return filteredItems.value.slice(s,s+pageSize) })

// 详情
const showDetail = ref(false), currentSA = ref(null), activeTab = ref('info')
const dtabs = [{k:'info',l:'基本信息'},{k:'secrets',l:'Secrets'},{k:'events',l:'事件'}]
function openDetail(sa) { currentSA.value=sa; activeTab.value='info'; showDetail.value=true }
function switchTab(k) { activeTab.value=k; if(k==='events'&&currentSA.value) loadEvents(currentSA.value) }

// 创建
const showCreate = ref(false)
const cf = ref({name:'',namespace:'',autoMount:true,labels:[]})
function openCreateModal() { cf.value={name:'',namespace:namespaceFilter.value||'',autoMount:true,labels:[]}; showCreate.value=true }
async function submitCreate() {
  if(!cf.value.name.trim()||!cf.value.namespace){ Message.warning({content:'名称和命名空间为必填项'}); return }
  submitting.value=true
  try {
    const labels={}; cf.value.labels.filter(l=>l.k.trim()).forEach(l=>{ labels[l.k]=l.v })
    await createSAApi(selectedClusterId.value, { name:cf.value.name, namespace:cf.value.namespace, auto_mount_token:cf.value.autoMount, labels:Object.keys(labels).length?labels:undefined })
    Message.success({content:'创建成功'}); showCreate.value=false; await loadServiceAccounts()
  } catch(e) { Message.error({content:'创建失败: '+(e?.msg||e?.message||String(e))}) }
  finally { submitting.value=false }
}

// 编辑
const showEdit = ref(false)
const ef = ref({name:'',namespace:'',autoMount:true,labels:[],annotations:[]})
function openEditModal(sa) {
  ef.value={ name:sa.name, namespace:sa.namespace, autoMount:sa.automount_token!==undefined?sa.automount_token:true,
    labels:Object.entries(sa.labels||{}).map(([k,v])=>({k,v})), annotations:Object.entries(sa.annotations||{}).map(([k,v])=>({k,v})) }
  showEdit.value=true
}
async function submitEdit() {
  submitting.value=true
  try {
    const labels={},annotations={}
    ef.value.labels.filter(l=>l.k.trim()).forEach(l=>{ labels[l.k]=l.v })
    ef.value.annotations.filter(a=>a.k.trim()).forEach(a=>{ annotations[a.k]=a.v })
    await updateSAApi(selectedClusterId.value, { name:ef.value.name, namespace:ef.value.namespace, auto_mount_token:ef.value.autoMount, labels, annotations })
    Message.success({content:'更新成功'}); showEdit.value=false
    const idx=serviceAccounts.value.findIndex(sa=>sa.name===ef.value.name&&sa.namespace===ef.value.namespace)
    if(idx!==-1){ serviceAccounts.value[idx].labels=labels; serviceAccounts.value[idx].automount_token=ef.value.autoMount }
    if(showDetail.value&&currentSA.value?.name===ef.value.name){ currentSA.value.labels=labels; currentSA.value.automount_token=ef.value.autoMount }
  } catch(e) { Message.error({content:'更新失败: '+(e?.msg||e?.message||String(e))}) }
  finally { submitting.value=false }
}

// YAML
const showYaml=ref(false),yamlTarget=ref(null),yamlContent=ref(''),yamlMode=ref('view'),yamlLoading=ref(false)
async function openYamlModal(sa) {
  yamlTarget.value=sa; yamlMode.value='view'; yamlContent.value=''; showYaml.value=true; yamlLoading.value=true
  try { const r=await getSAYamlApi(selectedClusterId.value,sa.namespace,sa.name); if(r.code===0) yamlContent.value=r.data?.yaml||'' }
  catch(e){ Message.error({content:'获取 YAML 失败'}) } finally{ yamlLoading.value=false }
}
function closeYaml(){ showYaml.value=false; yamlTarget.value=null; yamlContent.value='' }
async function submitYaml() {
  if(!yamlContent.value.trim()) return; submitting.value=true
  try { await applySAYamlApi(selectedClusterId.value,yamlContent.value); Message.success({content:'YAML 应用成功'}); closeYaml(); await loadServiceAccounts() }
  catch(e){ Message.error({content:'应用失败: '+(e?.msg||e?.message||String(e))}) } finally{ submitting.value=false }
}
async function copyYaml() {
  try{ await navigator.clipboard.writeText(yamlContent.value); Message.success({content:'已复制到剪贴板'}) }
  catch{ Message.warning({content:'请手动复制'}) }
}

// 事件
const showEvent=ref(false),evTarget=ref(null),saEvents=ref([]),evLoading=ref(false)
async function openEventModal(sa){ evTarget.value=sa; showEvent.value=true; await loadEvents(sa) }
async function loadEvents(sa) {
  evLoading.value=true; saEvents.value=[]
  try{ const r=await getSAEventsApi(selectedClusterId.value,sa.namespace,sa.name); if(r.code===0) saEvents.value=r.data?.list||[] }
  catch(e){ console.error(e) } finally{ evLoading.value=false }
}

// 删除
async function confirmDelete(sa) {
  const ok=await showConfirm({ title:'确认删除 ServiceAccount', type:'danger',
    details:[{label:'名称',value:sa.name,mono:true},{label:'命名空间',value:sa.namespace,mono:true}],
    confirmText:'确认删除', cancelText:'取消' })
  if(!ok) return; loading.value=true
  try {
    await deleteSAApi(selectedClusterId.value,sa.namespace,sa.name)
    Message.success({content:'删除成功'})
    if(showDetail.value&&currentSA.value?.name===sa.name) showDetail.value=false
    await loadServiceAccounts()
  } catch(e){ Message.error({content:'删除失败: '+(e?.msg||e?.message||String(e))}) }
  finally{ loading.value=false }
}

// 数据加载
async function loadClusters() {
  try {
    const r=await getClusterList({page:1,limit:100})
    if(r.code===0&&r.data?.list) {
      clusters.value=r.data.list.filter(c=>permissionStore.state.isSuperAdmin||permissionStore.state.accessibleClusterIds.includes(c.id)).map(c=>({...c,name:c.cluster_name||c.name}))
      if(clusters.value.length>0&&!selectedClusterId.value) selectedClusterId.value=clusters.value[0].id
    }
  } catch(e){ console.error(e) }
}
async function loadNamespaces() {
  if(!selectedClusterId.value) return
  try{ const r=await getNamespaces(selectedClusterId.value); if(r.code===0&&r.data?.list) namespaces.value=r.data.list.map(ns=>ns.name||ns) }
  catch{ namespaces.value=['default','kube-system','kube-public'] }
}
async function loadServiceAccounts() {
  if(!selectedClusterId.value) return; loading.value=true
  try{ const r=await listServiceAccounts(selectedClusterId.value,namespaceFilter.value); if(r.code===0) serviceAccounts.value=r.data?.list||[] }
  catch(e){ Message.error({content:'加载失败'}) } finally{ loading.value=false }
}
function onClusterChange(id){ if(id){ loadNamespaces(); loadServiceAccounts() } }
watch(selectedClusterId, id=>{ if(id){ loadNamespaces(); loadServiceAccounts() } })
watch(autoRefresh, val=>{ if(val){ refreshTimer=setInterval(loadServiceAccounts,30000) } else{ clearInterval(refreshTimer);refreshTimer=null } })

function formatDate(s) {
  if(!s) return '-'
  try{ const d=new Date(s); return isNaN(d)?s:d.toLocaleString('zh-CN',{hour12:false}) } catch{ return s }
}

onMounted(async()=>{ await loadClusters(); if(selectedClusterId.value){ loadNamespaces(); loadServiceAccounts() } })
onUnmounted(()=>{ if(refreshTimer) clearInterval(refreshTimer) })
</script>


<style scoped>
.sa-view{padding:24px;background:#f0f2f5;min-height:100vh;}
.sa-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:20px;}
.sa-header-left{display:flex;align-items:center;gap:14px;}
.sa-icon{font-size:36px;}
.sa-title-group h1{margin:0;font-size:22px;font-weight:700;color:#1d2129;}
.sa-title-group p{margin:4px 0 0;font-size:13px;color:#86909c;}
.sa-stats{display:grid;grid-template-columns:repeat(4,1fr);gap:16px;margin-bottom:20px;}
.stat-card{background:#fff;border-radius:10px;padding:18px 20px;display:flex;flex-direction:column;gap:6px;border-left:4px solid transparent;}
.stat-card.primary{border-color:#165dff;}.stat-card.success{border-color:#00b42a;}.stat-card.info{border-color:#0fc6c2;}.stat-card.warning{border-color:#ff7d00;}
.stat-icon{font-size:22px;}.stat-value{font-size:28px;font-weight:700;color:#1d2129;}.stat-label{font-size:12px;color:#86909c;}
.sa-toolbar{display:flex;justify-content:space-between;align-items:center;background:#fff;border-radius:10px;padding:12px 16px;margin-bottom:16px;gap:12px;flex-wrap:wrap;}
.toolbar-left,.toolbar-right{display:flex;align-items:center;gap:10px;flex-wrap:wrap;}
.sa-search{display:flex;align-items:center;gap:6px;background:#f2f3f5;border-radius:6px;padding:6px 10px;}
.sa-search input{border:none;background:transparent;outline:none;font-size:13px;width:180px;}
.ns-select{border:1px solid #e5e6eb;border-radius:6px;padding:6px 10px;font-size:13px;background:#fff;cursor:pointer;}
.view-toggle{display:flex;border:1px solid #e5e6eb;border-radius:6px;overflow:hidden;}
.vbtn{border:none;background:#fff;padding:6px 10px;cursor:pointer;font-size:14px;color:#86909c;}
.vbtn.active{background:#e8f3ff;color:#165dff;}
.auto-ref{display:flex;align-items:center;gap:6px;font-size:13px;color:#4e5969;cursor:pointer;}
.ref-dot{color:#00b42a;animation:pulse 1.5s infinite;}
@keyframes pulse{0%,100%{opacity:1}50%{opacity:.3}}
.sa-btn{padding:7px 16px;border-radius:6px;border:none;cursor:pointer;font-size:13px;font-weight:500;transition:.15s;white-space:nowrap;}
.sa-btn.pri{background:#165dff;color:#fff;}.sa-btn.pri:hover{background:#0e42d2;}
.sa-btn.sec{background:#fff;color:#4e5969;border:1px solid #e5e6eb;}.sa-btn.sec:hover{background:#f2f3f5;}
.sa-btn.danger{background:#fff1f0;color:#f53f3f;border:1px solid #ffa39e;}
.sa-btn.sm{padding:4px 10px;font-size:12px;}
.sa-btn:disabled{opacity:.5;cursor:not-allowed;}
.sa-loading{display:flex;align-items:center;justify-content:center;gap:10px;padding:80px;color:#86909c;}
.spin{width:20px;height:20px;border:2px solid #e5e6eb;border-top-color:#165dff;border-radius:50%;animation:spin .7s linear infinite;}
.spin.sm{width:14px;height:14px;}
@keyframes spin{to{transform:rotate(360deg)}}
.lrow{display:flex;align-items:center;gap:8px;color:#86909c;padding:20px 0;}
.sa-table-wrap{background:#fff;border-radius:10px;overflow:hidden;}
.sa-table{width:100%;border-collapse:collapse;}
.sa-table thead{background:#f7f8fa;}
.sa-table th{padding:12px 16px;text-align:left;font-size:12px;font-weight:600;color:#86909c;text-transform:uppercase;border-bottom:1px solid #e5e6eb;white-space:nowrap;}
.sa-table td{padding:12px 16px;font-size:13px;color:#4e5969;border-bottom:1px solid #f2f3f5;vertical-align:middle;}
.sa-row:hover{background:#f7f9ff;}
.name-cell{display:flex;align-items:center;gap:10px;}
.sa-av{font-size:18px;}
.nm{font-weight:600;color:#1d2129;font-size:13px;}
.meta{font-size:11px;color:#86909c;margin-top:2px;}
.tc{font-size:12px;color:#86909c;white-space:nowrap;}
.act-grp{display:flex;gap:6px;}
.ab{width:28px;height:28px;border:none;border-radius:6px;cursor:pointer;font-size:13px;display:flex;align-items:center;justify-content:center;transition:.15s;}
.ab.view{background:#e8f3ff;color:#165dff;}.ab.edit{background:#e8ffea;color:#00b42a;}
.ab.yaml{background:#f5f0ff;color:#7b4de9;}.ab.ev{background:#fff7e0;color:#ff7d00;}
.ab.del{background:#fff1f0;color:#f53f3f;}
.ab:hover{filter:brightness(.9);transform:scale(1.06);}
.bdg{display:inline-flex;align-items:center;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:500;}
.bdg.suc{background:#e8ffea;color:#00b42a;}.bdg.warn{background:#fff7e0;color:#ff7d00;}
.bdg.info{background:#e8f3ff;color:#165dff;}.bdg.def{background:#f2f3f5;color:#86909c;}
.ns-tag{background:#f0f4ff;color:#3370ff;border-radius:4px;padding:2px 8px;font-size:11px;font-weight:500;}
.mono{font-family:'JetBrains Mono','Fira Code',monospace;font-size:12px;}
.ml{margin-left:6px;}
.sa-empty{text-align:center;padding:60px 20px;}
.ei{font-size:48px;margin-bottom:12px;}.et{font-size:16px;font-weight:600;color:#4e5969;margin-bottom:6px;}.ed{font-size:13px;color:#86909c;margin-bottom:16px;}
.sa-page{display:flex;align-items:center;justify-content:center;gap:12px;padding:16px;}
.pb{padding:6px 14px;border:1px solid #e5e6eb;border-radius:6px;background:#fff;cursor:pointer;font-size:13px;}
.pb:hover:not(:disabled){background:#f0f4ff;border-color:#165dff;color:#165dff;}
.pb:disabled{opacity:.4;cursor:not-allowed;}
.pi{font-size:13px;color:#86909c;}
.sa-cards{display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:16px;}
.fw{grid-column:1/-1;}
.sa-card{background:#fff;border-radius:10px;border:1px solid #e5e6eb;transition:.2s;}
.sa-card:hover{box-shadow:0 4px 12px rgba(0,0,0,.1);transform:translateY(-2px);}
.ch{padding:16px;display:flex;align-items:center;gap:10px;background:#f7f9ff;border-bottom:1px solid #e5e6eb;}
.ci{font-size:24px;}.cn{font-weight:700;color:#1d2129;font-size:14px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;}
.cb{padding:14px 16px;display:flex;flex-direction:column;gap:8px;}
.cr{display:flex;justify-content:space-between;align-items:center;}
.cl{font-size:12px;color:#86909c;}.cv{font-size:11px;color:#4e5969;}
.cf{display:flex;gap:8px;padding:10px 16px;border-top:1px solid #f2f3f5;justify-content:flex-end;}
/* 抽屉 */
.overlay{position:fixed;inset:0;background:rgba(0,0,0,.35);z-index:1000;display:flex;justify-content:flex-end;}
.drawer{width:520px;max-width:95vw;background:#fff;height:100vh;display:flex;flex-direction:column;animation:slin .25s ease;}
@keyframes slin{from{transform:translateX(100%)}to{transform:translateX(0)}}
.dh{padding:16px 20px;border-bottom:1px solid #e5e6eb;position:relative;}
.dtrow{display:flex;align-items:center;gap:10px;margin-bottom:12px;}
.di{font-size:28px;}.dn{font-size:16px;font-weight:700;color:#1d2129;}
.dtabs{display:flex;border-bottom:2px solid #e5e6eb;}
.dtab{padding:8px 16px;border:none;background:transparent;cursor:pointer;font-size:13px;color:#86909c;border-bottom:2px solid transparent;margin-bottom:-2px;}
.dtab.active{color:#165dff;border-bottom-color:#165dff;font-weight:600;}
.dc{position:absolute;top:16px;right:16px;border:none;background:none;font-size:18px;cursor:pointer;color:#86909c;width:28px;height:28px;display:flex;align-items:center;justify-content:center;border-radius:4px;}
.dc:hover{background:#f2f3f5;}
.db{flex:1;overflow-y:auto;padding:20px;}
.df{display:flex;gap:10px;padding:16px 20px;border-top:1px solid #e5e6eb;}
.dgrid{display:grid;grid-template-columns:repeat(2,1fr);gap:14px;margin-bottom:20px;}
.ditem{display:flex;flex-direction:column;gap:4px;}.ditem.fw{grid-column:1/-1;}
.dl{font-size:11px;color:#86909c;text-transform:uppercase;letter-spacing:.3px;}
.dv{font-size:13px;color:#1d2129;font-weight:500;}
.dsec{margin-bottom:20px;}.dst{font-size:13px;font-weight:600;color:#1d2129;margin-bottom:10px;padding-bottom:6px;border-bottom:1px solid #f2f3f5;}
.lbs{display:flex;flex-wrap:wrap;gap:6px;}.lc{background:#f2f3f5;border-radius:4px;padding:3px 8px;font-size:12px;font-family:monospace;color:#4e5969;}
.eh{color:#c9cdd4;font-size:13px;padding:10px 0;}
.sl{display:flex;flex-direction:column;gap:6px;}
.sr{display:flex;align-items:center;gap:8px;padding:8px 10px;background:#f7f8fa;border-radius:6px;}
.etbl-wrap{overflow-x:auto;}
.etbl{width:100%;border-collapse:collapse;font-size:12px;}
.etbl th{padding:8px 12px;text-align:left;background:#f7f8fa;color:#86909c;border-bottom:1px solid #e5e6eb;white-space:nowrap;}
.etbl td{padding:8px 12px;border-bottom:1px solid #f2f3f5;vertical-align:middle;}
.em{max-width:260px;word-break:break-word;}
/* Modal */
.mbg{position:fixed;inset:0;background:rgba(0,0,0,.45);z-index:1100;display:flex;align-items:center;justify-content:center;}
.modal{background:#fff;border-radius:12px;width:520px;max-width:95vw;max-height:90vh;display:flex;flex-direction:column;box-shadow:0 8px 32px rgba(0,0,0,.18);overflow:hidden;}
.modal.wide{width:760px;}
.mh{display:flex;align-items:center;justify-content:space-between;padding:16px 20px;border-bottom:1px solid #e5e6eb;gap:12px;}
.mh h2{margin:0;font-size:16px;font-weight:700;color:#1d2129;}
.mc{border:none;background:none;font-size:18px;cursor:pointer;color:#86909c;width:28px;height:28px;display:flex;align-items:center;justify-content:center;border-radius:4px;}
.mc:hover{background:#f2f3f5;}
.mb{padding:20px;overflow-y:auto;flex:1;}
.mf{display:flex;justify-content:flex-end;gap:10px;padding:14px 20px;border-top:1px solid #e5e6eb;}
.fg{margin-bottom:16px;}
.fg label{display:block;font-size:13px;color:#4e5969;margin-bottom:6px;font-weight:500;}
.fg input:not([type='checkbox']),.fg select{width:100%;padding:8px 12px;border:1px solid #e5e6eb;border-radius:6px;font-size:13px;outline:none;box-sizing:border-box;}
.fg input:focus,.fg select:focus{border-color:#165dff;box-shadow:0 0 0 2px rgba(22,93,255,.12);}
.req{color:#f53f3f;}.tw{display:inline-flex;align-items:center;gap:8px;cursor:pointer;font-size:13px;}
.ht{font-size:12px;color:#86909c;margin-top:4px;}
.ro{display:flex;align-items:center;gap:12px;}.ro label{margin-bottom:0;min-width:80px;}.rv{font-size:13px;color:#1d2129;}
.lr{display:flex;align-items:center;gap:6px;margin-bottom:8px;}
.li{flex:1;padding:6px 10px;border:1px solid #e5e6eb;border-radius:6px;font-size:13px;outline:none;}
.li:focus{border-color:#165dff;}
.sep{color:#86909c;font-weight:700;}.rb{border:none;background:#fff1f0;color:#f53f3f;width:24px;height:24px;border-radius:4px;cursor:pointer;font-size:14px;}
.ymtabs{display:flex;gap:4px;}
.ymtab{padding:5px 14px;border:1px solid #e5e6eb;border-radius:6px;background:#fff;cursor:pointer;font-size:12px;color:#86909c;}
.ymtab.active{background:#e8f3ff;color:#165dff;border-color:#165dff;}
.yaml-mb{padding:0;}
.ye{width:100%;height:480px;border:none;padding:16px;font-family:'JetBrains Mono','Fira Code',monospace;font-size:12px;line-height:1.6;color:#1d2129;resize:none;outline:none;box-sizing:border-box;}
</style>
