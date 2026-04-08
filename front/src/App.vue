<template>
  <div id="app">
    <div v-if="!token" class="min-h-screen flex items-center justify-center p-4 bg-gray-100">
      <div class="w-full max-w-md bg-white rounded-xl shadow p-6">
        <h1 class="text-2xl text-center mb-6">账户凭证管理系统</h1>

        <el-input v-model="username" placeholder="输入用户名" class="mb-4" />
        <el-input v-model="loginCode" type="password" show-password placeholder="请输入登录口令" maxlength="6" class="mb-4" />

        <div class="flex gap-2">
          <el-button type="primary" @click="handleLogin" class="flex-1">登录</el-button>
          <el-button @click="goToRegister" class="flex-1">注册</el-button>
        </div>
      </div>
    </div>

    <div v-if="showRegisterPage" class="min-h-screen flex items-center justify-center p-4 bg-gray-100">
      <div class="w-full max-w-md bg-white rounded-xl shadow p-6">
        <h2 class="text-xl text-center mb-4">账号注册</h2>

        <el-input v-model="regUsername" placeholder="设置登录名" class="mb-4" />
        <el-input v-model="regLoginPassword" type="password" show-password placeholder="设置登录口令" class="mb-4" />

        <div class="flex gap-2 mt-4">
          <el-button type="success" @click="doRegister" class="flex-1">确认注册</el-button>
          <el-button @click="backToLogin">返回</el-button>
        </div>
      </div>
    </div>

    <div v-else-if="token" class="min-h-screen">
      <div class="bg-white shadow mb-4">
        <div class="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <h1 class="text-xl">账户凭证管理系统</h1>
          <el-button type="danger" @click="logout">退出登录</el-button>

          <div v-if="showEditRecordModal" class="bg-white p-5 rounded shadow my-6">
            <h3 class="text-lg mb-4">
              账户凭证管理：{{ currentAccountName }}
            </h3>

            <el-form :model="recordForm" label-width="80px" class="max-w-lg mb-6">
              <el-form-item label="账户简码">
                <el-input v-model="recordForm.accountCode" placeholder="输入账户简码"
                  @blur="getAccountByCode(recordForm.accountCode)" />
                <div class="text-sm mt-1">
                  对应账户：<span class="text-primary font-medium">{{ recordForm.accountName }}</span>
                </div>
              </el-form-item>
              <el-form-item label="键">
                <div class="flex gap-2 mb-2 flex-wrap">
                  <el-button size="small" type="primary" v-for="name in commonItemNames" :key="name"
                    @click="recordForm.key = name">
                    {{ name }}
                  </el-button>
                </div>
                <el-input v-model="recordForm.key" placeholder="也可手动输入" category />
              </el-form-item>
              <el-form-item label="值">
                <el-input v-model="recordForm.value"
                  :type="recordForm.isEncrypt && !recordForm.isTotp ? 'password' : ''"
                  :show-password="recordForm.isEncrypt && !recordForm.isTotp" placeholder="输入内容" />
              </el-form-item>
              <el-form-item label="时间">
                <el-date-picker v-model="recordForm.time" type="datetime" format="YYYY-MM-DD HH:mm"
                  value-format="YYYY-MM-DD HH:mm" placeholder="请选择时间" :clearable="true" :editable="true"
                  popper-append-to-body>
                </el-date-picker>
              </el-form-item>
              <el-form-item label="加密">
                <el-switch v-model="recordForm.isEncrypt" />
              </el-form-item>
              <el-form-item label="TOTP">
                <el-switch v-model="recordForm.isTotp" @change="handleTotpSwitchChange" />
              </el-form-item>
              <el-form-item label="删除">
                <el-switch v-model="recordForm.isDeleted" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="editRecord(0)">新增</el-button>
                <el-button v-if="recordForm.beforeValue" type="primary" @click="editRecord(1)">保存修改</el-button>
                <el-button v-if="recordForm.beforeValue" @click="editRecord(2)">添加查询记录</el-button>
                <el-button @click="resetrecordForm">清空</el-button>
                <el-button @click="showEditRecordModal = false">收起</el-button>
              </el-form-item>
            </el-form>
          </div>
        </div>
      </div>

      <div class="max-w-7xl mx-auto px-4">
        <el-tabs v-model="tab" class="mb-6">
          <el-tab-pane label="账号管理" name="accounts" />
          <el-tab-pane label="操作记录" name="records" />
        </el-tabs>

        <div v-show="tab === 'accounts'">
          <div class="bg-white p-5 rounded shadow mb-6">
            <h3 class="text-lg mb-4">新增账号</h3>
            <el-form v-if="showEditModal" :model="editForm" label-width="60px" class="max-w-lg">
              <el-form-item label="简码">
                <el-input v-model="editForm.code" placeholder="请输入简码" />
              </el-form-item>
              <el-form-item label="账户名称">
                <el-input v-model="editForm.accountName" placeholder="请输入账户名称" />
              </el-form-item>
              <el-form-item label="账户分类">
                <div class="flex gap-2 mb-2 flex-wrap">
                  <el-button size="small" type="primary" v-for="name in commonCategoryNames" :key="name"
                    @click="recordForm.key = name">
                    {{ name }}
                  </el-button>
                </div>
                <el-input v-model="editForm.category" placeholder="请输入账户分类" />
              </el-form-item>
              <el-form-item>
                <el-button :type="isEdit ? 'success' : 'primary'" @click="doAddEdit(0)">新增</el-button>
                <el-button v-if="editForm.id" type="primary" @click="doAddEdit(1)">保存修改</el-button>
                <el-button @click="resetForm">清空</el-button>
                <el-button @click="showAllFlag = !showAllFlag">{{ showAllFlag ? "隐藏所有" : "显示所有" }}</el-button>
                <el-button @click="showEditModal = false">收起</el-button>
              </el-form-item>
            </el-form>
          </div>

          <el-button type="primary" @click="openEditForm" class="mb-4">新增账号</el-button>

          <el-table :data="accounts" row-key="id" border class="mb-4">
            <el-table-column prop="code" label="简码" width="60" />
            <el-table-column prop="accountName" label="账户名称" width="150" />
            <el-table-column prop="category" label="分类" width="120" />
            <el-table-column label="账户信息">
              <template #default="scope">
                <div v-for="item in scope.row.authRecords" :key="item.id">
                  {{ item.key }}：
                  <span v-if="item.isTotp">
                    {{ item.totpCode }} ({{ item.totpRemaining }}s)
                  </span>
                  <span v-else :title="datetimeFormat(item.time)">
                    {{ item.showFlag || showAllFlag ? item.value : item.showValue }}
                  </span>
                  <el-button v-if="showEditRecordModal" type="text"
                    @click="openEditRecordModal(scope.row, item)">修改</el-button>
                  <el-icon class="cursor-pointer" @click="item.showFlag = !item.showFlag"></el-icon>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="时间" width="155">
              <template #default="scope">
                {{ datetimeFormat(scope.row.createdAt) }}<br />{{ datetimeFormat(scope.row.updatedAt) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="60">
              <template #default="scope">
                <el-button type="text" @click="openEditForm(scope.row)">编辑</el-button>
                <el-button type="text" @click="openEditRecordModal(scope.row)">记录</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" background
            layout="total, sizes, prev, pager, next, jumper" @size-change="loadAccounts"
            @current-change="loadAccounts" />
        </div>

        <div v-show="tab === 'records'">
          <el-table :data="records" border>
            <el-table-column prop="code" label="简码" />
            <el-table-column prop="item" label="项目" />
            <el-table-column prop="beforeValue" label="修改前" />
            <el-table-column prop="value" label="修改后" />
          </el-table>
        </div>

      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import axios from 'axios'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import CryptoJS from 'crypto-js'
import jsotp from 'jsotp'

const datetimeFormat = (dt) => dayjs(dt).format('YYYY-MM-DD HH:mm')
const datetimeFormatISO = (dt) => dayjs(dt).format()

const token = ref(localStorage.getItem('token') || '')
const username = ref('')
const loginCode = ref('')
const showAllFlag = ref(false)
const showRegisterPage = ref(false)
const showEditRecordModal = ref(false)
const regUsername = ref('')
const regLoginPassword = ref('')

const page = ref(1)
const pageSize = ref(1000)
const total = ref(0)
const tab = ref('accounts')
const accounts = ref([])
const records = ref([])
const showEditModal = ref(false)
const isEdit = ref(false)

const editForm = ref({
  id: '',
  code: '',
  accountName: '',
  category: '',
})

const recordForm = ref({
  id: 0,
  accountId: 0,
  accountCode: '',
  key: '',
  beforeValue: '',
  value: '',
  time: datetimeFormat(Date.now()),
  isEncrypt: false,
  isTotp: false,
  isDeleted: false,
})

const resetrecordForm = () => {
  recordForm.value = { id: 0, accountId: 0, accountCode: '', key: '', beforeValue: '', value: '', time: datetimeFormat(Date.now()), isEncrypt: false, isTotp: false, isDeleted: false }
  isEdit.value = false
}

const resetForm = () => {
  editForm.value = { id: '', code: '', accountName: '', category: '' }
  isEdit.value = false
}

const beforeValueProcess = (beforeValue, isEncrypt) => {
  if (!beforeValue && beforeValue !== '') return ''
  return isEncrypt ? encryptDekEcb(beforeValue) : beforeValue
}

const commonCategoryNames = ref(['',])
const commonItemNames = ref(['登录账号', '登录密码',])

const api = axios.create({ baseURL: `${window.location.protocol}//${window.location.hostname}:62823/api` })
api.interceptors.response.use(res => res.data, err => { ElMessage.error('请求失败'); return Promise.reject(err) })

const encryptKey = '123456789abcdefg'
function encryptDekEcb(dek) {
  return CryptoJS.AES.encrypt(CryptoJS.enc.Utf8.parse(dek), CryptoJS.enc.Utf8.parse(encryptKey), { mode: CryptoJS.mode.ECB, padding: CryptoJS.pad.Pkcs7 }).toString()
}
function decryptDekEcb(encryptedText) {
  if (!encryptedText) return ''
  try {
    return CryptoJS.AES.decrypt(encryptedText, CryptoJS.enc.Utf8.parse(encryptKey), { mode: CryptoJS.mode.ECB, padding: CryptoJS.pad.Pkcs7 }).toString(CryptoJS.enc.Utf8)
  } catch { return encryptedText }
}

async function handleLogin() {
  try {
    const res = await api.post('/login', { loginName: username.value, dek: encryptDekEcb(loginCode.value) })
    if (res.data.login) { token.value = 'LOGIN_SUCCESS'; localStorage.setItem('token', token.value); loadAll() }
    else ElMessage.warning('用户名或密码错误')
  } catch { ElMessage.error('登录失败') }
}

async function doRegister() {
  const res = await api.post('/register', { loginName: regUsername.value, dek: encryptDekEcb(regLoginPassword.value) })
  if (res.code === '200') { ElMessage.success('注册成功'); backToLogin() } else ElMessage.warning(res.msg)
}

function goToRegister() { showRegisterPage.value = true; regUsername.value = ''; regLoginPassword.value = '' }
function backToLogin() { showRegisterPage.value = false; username.value = ''; loginCode.value = '' }
function logout() { token.value = ''; localStorage.removeItem('token') }

async function loadAll() { await loadAccounts() }

async function loadAccounts() {
  const res = await api.get('/accounts', { params: { page: page.value, page_size: pageSize.value } })
  let list = res.data.list || []
  accounts.value = list.map(item => {
    const account = { ...item }
    if (account.authRecords) {
      account.authRecords = account.authRecords.map(record => {
        let value = record.isEncrypt ? decryptDekEcb(record.value) : record.value
        let showValue = record.isEncrypt ? '●●●●●●●●' : (/^\d{8,}$/.test(value) ? value[0] + '*'.repeat(value.length - 2) + value.slice(-1) : value)
        return { ...record, value, showValue, showFlag: false, totpCode: '------', totpRemaining: 0 }
      })
    }
    return account
  })
  total.value = res.data.total || 0
}

async function doAddEdit(mode) {
  try {
    const data = { code: editForm.value.code, accountName: editForm.value.accountName, category: editForm.value.category }
    const res = mode === 1 ? await api.put(`/accounts/${editForm.value.id}`, { id: editForm.value.id, ...data }) : await api.post('/accounts', data)
    if (res.code == '200') { ElMessage.success(mode ? '修改成功' : '新增成功'); showEditModal.value = false; loadAccounts() }
    else ElMessage.error(res.msg)
  } catch (e) { ElMessage.error('操作失败') }
}

function openEditForm(item) {
  showEditModal.value = true
  if (item) {
    editForm.value.id = item.id
    editForm.value.code = item.code
    editForm.value.accountName = item.accountName
    editForm.value.category = item.category
    isEdit.value = true
  } else {
    editForm.value = { id: '', code: '', accountName: '', category: '' }
    isEdit.value = false
  }
}

async function editRecord(mode) {
  try {
    if (!recordForm.value.accountId) return ElMessage.warning('请选择账户')
    if (!recordForm.value.key) return ElMessage.warning('请输入键')
    if (!recordForm.value.value) return ElMessage.warning('请输入值')

    const data = {
      accountId: recordForm.value.accountId,
      key: recordForm.value.key,
      beforeValue: beforeValueProcess(recordForm.value.beforeValue, recordForm.value.isEncrypt && !recordForm.value.isTotp),
      value: recordForm.value.isTotp
        ? recordForm.value.value
        : (recordForm.value.isEncrypt ? encryptDekEcb(recordForm.value.value) : recordForm.value.value),
      time: datetimeFormatISO(recordForm.value.time),
      isEncrypt: recordForm.value.isEncrypt,
      isTotp: recordForm.value.isTotp,
      isDeleted: recordForm.value.isDeleted,
    }
    let res
    if (mode === 0) res = await api.post('/accounts/records', data)
    else if (mode === 1) res = await api.post(`/accounts/records/${recordForm.value.id}`, data)
    else res = await api.post(`/accounts/records/addlog/${recordForm.value.id}`, data)

    if (res.code == '200') { ElMessage.success('操作成功'); loadAccounts() }
    else ElMessage.error(res.msg)
  } catch (e) { ElMessage.error('操作失败') }
}

function openEditRecordModal(account, r = null) {
  showEditRecordModal.value = true
  recordForm.value.accountId = account?.id || 0
  recordForm.value.accountCode = account?.code || ''
  recordForm.value.accountName = account?.accountName || ''
  if (r) {
    recordForm.value.id = r.id
    recordForm.value.key = r.key
    recordForm.value.value = r.value
    recordForm.value.beforeValue = r.value
    recordForm.value.time = r.time
    recordForm.value.isEncrypt = r.isEncrypt
    recordForm.value.isTotp = r.isTotp
    recordForm.value.isDeleted = r.isDeleted
  } else {
    recordForm.value.id = 0
    recordForm.value.key = ''
    recordForm.value.beforeValue = ''
  }
}

function getAccountByCode(code) {
  const acc = accounts.value.find(x => x.code === code)
  if (acc) { recordForm.value.accountId = acc.id; recordForm.value.accountName = acc.accountName }
  else { recordForm.value.accountId = 0; recordForm.value.accountName = '' }
}

// 生成 TOTP 验证码
function generateTOTP(secret) {
  try {
    const totp = jsotp.TOTP(secret);
    const code = totp.now();
    const remaining = 30 - (Math.floor(Date.now() / 1000) % 30);
    return { code, remaining };
  } catch (e) {
    console.error("TOTP错误", e)
    return { code: "错误", remaining: 0 };
  }
}

// 刷新所有令牌
function refreshAllTotps() {
  for (const acc of accounts.value) {
    if (!acc.authRecords) continue;
    for (const item of acc.authRecords) {
      if (!item.isTotp) continue;

      try {
        // 自动解密（你之前加密过）
        let realSecret = item.value;
        try { realSecret = decryptDekEcb(item.value); } catch { }

        const res = generateTOTP(realSecret);
        item.totpCode = res.code;
        item.totpRemaining = res.remaining;
      } catch {
        item.totpCode = "错误";
        item.totpRemaining = 0;
      }
    }
  }
}

// 自动刷新定时器
let totpTimer = null;
function startTotpAutoRefresh() {
  clearInterval(totpTimer);
  refreshAllTotps();
  totpTimer = setInterval(refreshAllTotps, 1000);
}

// TOTP 开关自动填 key=TOTP
function handleTotpSwitchChange(val) {
  if (val) {
    recordForm.value.isEncrypt = true;
    recordForm.value.key ||= "TOTP";
  }
}
onMounted(() => {
  if (token.value) {
    loadAccounts().then(() => {
      startTotpAutoRefresh()
    })
  }
})

onUnmounted(() => clearInterval(totpTimer))
</script>
