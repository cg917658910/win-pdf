<template>
    <div class="app-container">
      <!-- 使用 Go 的原生多文件选择对话框，不需要浏览器 file input -->
      <!-- 顶部菜单 -->
      <!-- <div class="top-tabs">
        <button :class="{ active: activeTab === 'file' }" @click="activeTab='file'">文件</button>
        <button :class="{ active: activeTab === 'register' }" @click="activeTab='register'">注册</button>
        <button :class="{ active: activeTab === 'menu' }" @click="activeTab='menu'">菜单栏</button>
      </div> -->
  
      <!-- 主体 -->
      <div class="main">
        <!-- 左侧 -->
        <div class="left-panel">
          <div class="action-buttons">
            <button @click="addFile">添加文件</button>
            <button @click="addFolder">添加文件夹</button>
           <!--  <button @click="setExpire">设置文档时效</button> -->
          </div>
  
          <div class="file-list">
            <div class="file-item" v-for="(file, index) in files" :key="index">
              <span>{{ file.name }}</span>
              <span class="path">{{ file.path }}</span>
              <button @click="removeFile(index)" style="margin-left:8px">删除</button>
            </div>
          </div>
  
          <!-- <div class="tips">
            <p>说明：</p>
            <p>1、添加进来的PDF文件以文件名加路径的形式展示。</p>
            <p>2、在文件列表上右键选中某个文件可以删除。</p>
          </div> -->
        </div>
  
        <!-- 右侧 -->
        <div class="right-panel">
          <div class="card">
            <h3>加密选项</h3>
  
            <div class="checkbox-group" style="margin-bottom:15px;">
              <label><input type="checkbox" v-model="options.copy"> 允许复制</label>
              <label><input type="checkbox" v-model="options.edit"> 允许编辑</label>
              <label><input type="checkbox" v-model="options.convert"> 允许转换</label>
              <label><input type="checkbox" v-model="options.print"> 允许打印</label>
            </div>
          </div>
          <div class="card">
            <label class="block">
              <input type="checkbox" v-model="options.unsupportedTip">
              文件过期或在非指定的PDF阅读器中打开时的提示
            </label>
  
            <textarea v-model="unsupportedText" rows="2"></textarea>
  
            <label class="block">
              <input type="checkbox" v-model="options.expiredTip">
              文档过期后显示以下提示
            </label>
  
            <textarea v-model="expiredText" rows="2"></textarea>
          </div>
  
          <div class="card">
            <h3>设置时效</h3>
            <div class="time-row">
              <span>开始时间</span>
              <input type="datetime-local" v-model="startTime" />
              <span>结束时间</span>
              <input type="datetime-local" v-model="endTime" />
            </div>
          </div>
          <div class="card-b pwd-card" id="user-password-card">
            <!-- <h3>用户密码</h3> -->
            <div class="pwd-row">
              <button class="btn primary" @click="openPwdModal">设置文档密码</button>
              <button class="btn primary" @click="setExpire">点击完成设置</button>
            </div>
          </div>
         
        </div>
      </div>
      
      <!-- 底部版权信息 -->
      <div class="app-footer">
        <div class="logo">
            <img src="./assets/images/logo.png" alt="logo" />
          </div>
        <div>
            ©远信软件技术服务有限公司&nbsp;&nbsp;2026.02&nbsp;&nbsp;未经许可，禁止售卖、传播，违者必究！
        </div>
        <div class="machine-code">激活码：{{ machineCode }}</div>
      </div>
      <!-- 注册模态 -->
      <div v-if="showRegisterModal" class="modal-overlay">
        <div class="modal">
          <h3>在线注册</h3>
          <div class="register-machine">
            <div>机器码：</div>
            <div class="machine-row">
              <span class="machine-value">{{ machineCode }}</span>
              <button class="copy-btn" @click="copyMachineCode">复制</button>
            </div>
          </div>
          <input v-model="activationCode" placeholder="输入注册码" />
          <div class="modal-actions">
            <button @click="onRegister" :disabled="registerLoading">注册</button>
            <button @click="onCancel" :disabled="registerLoading">取消</button>
          </div>
        </div>
      </div>
      <div v-if="showPwdModal" class="modal-overlay">
        <div class="modal">
          <h3>设置文档密码</h3>
          <input v-model="pwd" type="password" placeholder="请输入至少6位的密码" autofocus @keyup.enter="confirmPwdModal" />
          <div class="modal-actions">
            <button @click="confirmPwdModal">确定</button>
            <button @click="cancelPwdModal">取消</button>
          </div>
        </div>
      </div>
    </div>
  </template>
  
  <script setup>
  import { ref, onMounted } from "vue"
  import { SetExpiry } from "../wailsjs/go/main/App.js"
  import { engine } from "../wailsjs/go/models"
  import { CanResolveFilePaths, ResolveFilePaths, LogPrint, EventsOn } from "../wailsjs/runtime/runtime.js"
  import { OpenDirectoryDialog, MessageDialog, OpenMultipleFilesDialog, IsRegistered, GetMachineCode, Register } from "../wailsjs/go/main/App.js"
  
  const activeTab = ref("file")
  
  const files = ref([
   
  ])
  
  const options = ref({
    copy: false,
    edit: false,
    convert: false,
    print: false,
    unsupportedTip: true,
    expiredTip: true,
  })
  
  const unsupportedText = ref(
    "文件显示错误！请使用Adobe Reader、PDF-Xchange或福昕PDF阅读器打开当前文档！"
  )
  
  const expiredText = ref("您查看的文档已过期！")
  
  const pwdEnabled = ref(false)
  const pwd = ref("")
  
  // 注册相关
  const showRegisterModal = ref(false)
  const activationCode = ref("")
  const machineCode = ref("")
  const registerLoading = ref(false)
  
  // 默认显示当前本地时间，格式符合 input[type=datetime-local]
  function formatLocalDatetime(d) {
    const pad = (n) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
  }
  const nowStr = formatLocalDatetime(new Date())
  const startTime = ref(nowStr)
  const endTime = ref(nowStr)
  const sending = ref(false)
  const fileInput = ref(null)
  
  function addFile() {
    // 调用 Go 原生多文件选择对话框
    (async () => {
      try {
        const paths = await OpenMultipleFilesDialog()
        if (Array.isArray(paths) && paths.length > 0) {
            // 追加去重
            const existingPaths = new Set(files.value.map(f => f.path))
            const newPaths = paths.filter(p => !existingPaths.has(p))
            newPaths.forEach(p => files.value.push({ name: p.split(/[/\\]/
).pop(), path: p }))             
          //files.value = paths.map(p => ({ name: p.split(/[/\\]/).pop(), path: p }))
        }
      } catch (err) {
        console.error('OpenMultipleFilesDialog error', err)
        await MessageDialog('提示', '选择文件已取消或出错')
      }
    })()
  }

  function addFolder() {
    // 调用 Go 原生选择文件夹对话框
    (async () => {
      try {
        const folderPath = await OpenDirectoryDialog()
        if (folderPath) {
          // 解析该文件夹下的所有PDF文件路径
          const canResolve = await CanResolveFilePaths(folderPath)
          if (!canResolve) {
            await MessageDialog('提示', '无法解析所选文件夹的文件路径')
            return
          }
          const resolvedPaths = await ResolveFilePaths(folderPath)
          // 过滤出PDF文件
          const pdfPaths = resolvedPaths.filter(p => p.toLowerCase().endsWith('.pdf'))
          if (pdfPaths.length === 0) {
            await MessageDialog('提示', '所选文件夹下没有找到PDF文件')
            return
          }
          // 追加去重
          const existingPaths = new Set(files.value.map(f => f.path))
          const newPaths = pdfPaths.filter(p => !existingPaths.has(p))
          newPaths.forEach(p => files.value.push({ name: p.split(/[/\\]/).pop(), path: p }))             
        }
      } catch (err) {
        LogPrint('OpenDirectoryDialog error', err)
        await MessageDialog('提示', '选择文件夹已取消或出错')
      }
    })()
  }
  
  function removeFile(index) {
    if (index >= 0 && index < files.value.length) {
      files.value.splice(index, 1)
    }
  }
  
    const showPwdModal = ref(false)
    function openPwdModal() {
      showPwdModal.value = true
    }

    async function confirmPwdModal() {
      if (!pwd.value || pwd.value.length < 6) {
        await MessageDialog('提示', '密码至少6位')
        return
      }
      showPwdModal.value = false
    }
    function cancelPwdModal() {
      pwd.value = ""
      showPwdModal.value = false
    }
  async function setExpire() {
    // 直接提交给后端，不再选择保存路径
    if (sending.value) return
    if (files.value.length === 0) {
      await MessageDialog('提示', '请先通过“添加文件”选择要处理的文件')
      return
    }
    // 验证密码长度
    if ((pwd.value && pwd.value.length < 6)) {
      await MessageDialog('提示', '密码至少6位')
      return
    }
    // 弹出原生选择目录对话框以获取保存目录（OpenDirectoryDialog）
    let folderPath = ''
    try {
      folderPath = await OpenDirectoryDialog()
      // 打印返回的保存目录到 Wails 日志（使用 runtime 的 LogPrint）
      try { LogPrint(folderPath) } catch (e) { console.debug('LogPrint error', e) }
    } catch (err) {
      console.error('OpenDirectoryDialog error', err)
      await MessageDialog('提示', '选择目录已取消或出错')
      return
    }
    if (!folderPath) {
      await MessageDialog('提示', '未选择保存目录')
      return
    }
  
    const opts = new engine.Options()
    opts.Files = files.value.map(f => f.path).join(';')
    opts.OutputDir = folderPath
    opts.StartTime = startTime.value ? new Date(startTime.value).toISOString() : null
    opts.EndTime = endTime.value ? new Date(endTime.value).toISOString() : null
    opts.Watermark = ""
    opts.ExperiredText = expiredText.value
    opts.UnsupportedText = unsupportedText.value
    opts.AllowedPrint = !options.value.print
    opts.AllowedCopy = !options.value.copy
    // 编辑
    opts.AllowedEdit = !options.value.edit
    opts.AllowedConvert = !options.value.convert
    // 用户密码绑定
    opts.PwdEnabled = pwdEnabled.value
    opts.UserPassword = pwd.value

    try {
      sending.value = true
      const res =await SetExpiry(opts)
      await MessageDialog('提示', res)
      LogPrint(res)
    } catch (err) {
      console.error('SetExpiry error', err)
      await MessageDialog('错误', '设置文档时效请求失败：' + (err && err.message ? err.message : err))
    } finally {
      sending.value = false
    }
  }

  async function onRegister() {
    if (!activationCode.value) {
      await MessageDialog('提示', '请输入注册码')
      return
    }
    try {
      registerLoading.value = true
      const res = await Register(activationCode.value)
      await MessageDialog('提示', res || '注册成功')
      showRegisterModal.value = false
      // 更新已注册状态和机器码显示
      machineCode.value = await GetMachineCode()
    } catch (err) {
      console.error('Register error', err)
      await MessageDialog('错误', '注册失败：' + (err && err.message ? err.message : err))
    } finally {
      registerLoading.value = false
    }
  }

  function onCancel() {
    showRegisterModal.value = false
  }

  async function copyMachineCode() {
    if (!machineCode.value) return
    try {
      if (navigator && navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(machineCode.value)
        await MessageDialog('提示', '机器码已复制到剪贴板')
      } else {
        // fallback: create temporary textarea
        const ta = document.createElement('textarea')
        ta.value = machineCode.value
        document.body.appendChild(ta)
        ta.select()
        try { document.execCommand('copy') } catch (e) { console.debug('execCommand copy failed', e) }
        ta.remove()
        await MessageDialog('提示', '机器码已复制到剪贴板')
      }
    } catch (e) {
      console.error('copyMachineCode error', e)
      await MessageDialog('错误', '复制失败：' + (e && e.message ? e.message : e))
    }
  }

  onMounted(async () => {
    // 获取并显示机器码
    try {
      machineCode.value = await GetMachineCode()
    } catch (e) {
      console.debug('GetMachineCode error', e)
    }
    // 监听菜单触发注册事件
    try {
      EventsOn('menu:register', async () => {
        const reg = await IsRegistered()
        if (reg) {
          await MessageDialog('提示', '软件已注册')
          return
        }
        showRegisterModal.value = true
      })
    } catch (e) {
      console.debug('EventsOn error', e)
    }
  })
  </script>
  
  <style scoped>
  .app-container {
    font-family: "Microsoft YaHei", sans-serif;
    padding: 10px;
    color: #222;
    background: linear-gradient(180deg, #f7f9fb 0%, #eef3f8 100%);
    padding-bottom: 44px;
  }
  
  .top-tabs {
    display: flex;
    gap: 10px;
    margin-bottom: 10px;
  }
  
  .top-tabs button {
    padding: 4px 12px;
    border: 1px solid #999;
    background: #f5f5f5;
    cursor: pointer;
  }
  
  .top-tabs button.active {
    background: #fff;
    border-bottom: 2px solid #333;
  }
  
  .main {
    display: flex;
  }
  
  .left-panel {
    width: 45%;
    border: 1px solid #ddd;
    padding: 10px;
    background: #fff;
    box-shadow: 0 1px 3px rgba(0,0,0,0.04);
  }
  
  .action-buttons {
    display: flex;
    gap: 10px;
    margin-bottom: 10px;
  }
  
  .action-buttons button{
    padding:6px 10px;
    border-radius:4px;
    border:1px solid #cfcfcf;
    background:#f3f6f9;
    cursor:pointer;
  }
  
  .file-list {
    border: 1px solid #eee;
    overflow: auto;
    padding: 5px;
    background:#fafafa;
  }
  
  .file-item {
    display: flex;
    justify-content: space-between;
    font-size: 14px;
    padding:6px 8px;
    border-radius:4px;
  }
  
  .file-item + .file-item{margin-top:6px}
  .file-item:hover{background:#f0f6ff}
  
  .file-item .path {
    color: #888;
    font-size:12px;
  }
  
  .tips {
    color: #a00;
    font-size: 13px;
    margin-top: 10px;
  }
  
  .right-panel {
    width: 55%;
    padding-left: 15px;
  }
  
  .card {
    border: 1px solid #eee;
    padding: 10px;
    margin-bottom: 15px;
    background:#fff;
  }
  .card-b {
    border: 1px solid #eee;
    padding: 10px;
    background:#fff;
  }
  
  .checkbox-group {
    display: flex;
    gap: 20px;
    margin-bottom: 10px;
  }
  
  .block {
    display: block;
    margin-top: 10px;
    margin-bottom: 5px;
  }
  
  textarea {
    width: 100%;
    resize: none;
    margin-top: 5px;
    min-height:44px;
    padding:6px;
    box-sizing:border-box;
    border:1px solid #e6e6e6;
    border-radius:4px;
    background:#fff;
  }
  
  .time-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  
  .logo {
    align-items: center;
    gap: 10px;
    margin-top: 20px;
  }
  
  .logo img {
    width: 120px;
    max-width: 100%;
    height: auto;
    margin-bottom: 10px;
  }
  
  .cn {
    font-size: 18px;
    font-weight: bold;
  }
  
  .en {
    font-size: 14px;
  }

  .app-footer {
    position: fixed;
    left: 0;
    right: 0;
    bottom: 0;
    text-align: center;
    padding: 8px 12px;
    font-size: 12px;
    color: #888;
    background: rgba(255,255,255,0.85);
    border-top: 1px solid rgba(0,0,0,0.04);
    z-index: 9999;
  }

  .modal-overlay {
    position: fixed;
    left: 0; right: 0; top: 0; bottom: 0;
    background: rgba(0,0,0,0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10000;
  }
  .modal {
    background: #fff;
    padding: 18px;
    border-radius: 6px;
    width: 360px;
    box-shadow: 0 6px 24px rgba(0,0,0,0.2);
  }
  .modal input { width: 100%; padding:8px; margin-top:10px; box-sizing:border-box }
  .modal-actions { margin-top:12px; display:flex; gap:10px; justify-content:flex-end }
  .machine-code { margin-top:6px; font-size:12px; color:#666 }
  .register-machine { margin-bottom: 8px; }
  .machine-row { display:flex; align-items:center; gap:8px; margin-top:4px }
  .machine-value { font-family: monospace; background:#f4f4f4; padding:4px 8px; border-radius:4px; color:#333 }
  .copy-btn { padding:4px 8px; border:1px solid #cfcfcf; background:#fff; cursor:pointer }

  /* 用户密码卡居中显示并圆角 */
  .pwd-card {
    margin: 12px auto 0; /* 水平居中并与下方间距 */
    box-shadow: 0 4px 12px rgba(0,0,0,0.06);
    text-align: center;
    padding: 14px;
    border: 1px solid rgba(0,0,0,0.06);
    background: #fff;
    min-height: 100px;
    display: flex;
    flex-direction: column;
    justify-content: center; /* 垂直居中 */
  }

  .pwd-card .pwd-row {
    display: flex;
    gap: 16px;
    justify-content: center;
    align-items: center;
  }

  .pwd-card .btn{
    padding: 12px 18px;
    border-radius: 15px;
    border: 1px solid #d0d0d0;
    background: #fff;
    cursor: pointer;
    min-width: 140px;
    margin-right: 15px;
  }
  .pwd-card .btn.primary{ background:#f3f6f9 }
  .pwd-card .btn.success{ background:#4caf50; color:#fff; border-color:#4caf50 }
  </style>
