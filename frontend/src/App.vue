<template>
    <div class="app-container">
      <!-- 使用 Go 的原生多文件选择对话框，不需要浏览器 file input -->
      <!-- 顶部菜单 -->
      <!-- <div class="top-tabs">
        <button :class="{ active: activeTab === 'file' }" @click="activeTab='file'">文件</button>
        <button :class="{ active: activeTab === 'register' }" @click="activeTab='register'">注册</button>
        <button :class="{ active: activeTab === 'menu' }" @click="activeTab='menu'">菜单</button>
      </div> -->
  
      <!-- 主体 -->
      <div class="main">
        <!-- 左侧 -->
        <div class="left-panel">
          <div class="action-buttons">
            <button @click="addFile">添加文件</button>
            <button @click="addFolder">添加文件夹</button>
            <button @click="removeFileAll">清空列表</button>
          </div>
  
          <div class="file-list">
            <div class="file-item" v-for="(file, index) in files" :key="index">
              <span>{{ file.name }}</span>
              <span class="path">{{ file.path }}</span>
             <button @click="removeFile(index)" style="margin-left:8px">删除</button>
            </div>
          </div>
  
          <!-- <div class="tips">
            <p>说明</p>
            <p>1、添加进来的PDF文件以文件名加路径的形式展示。</p>
            <p>2、在文件列表上右键选中某个文件可以删除。</p>
          </div> -->
        </div>
  
        <!-- 右侧 -->
        <div class="right-panel">
          <div class="card">
            <h3>加密选项</h3>
  
            <div class="checkbox-group" style="margin-bottom:15px;">
              <label><input type="checkbox" v-model="options.copy"> 允许复制转换</label>
              <label><input type="checkbox" v-model="options.edit"> 允许编辑</label>
              <!-- <label><input type="checkbox" v-model="options.convert"> 允许转换</label> -->
              <label><input type="checkbox" v-model="options.print"> 允许打印</label>
              <button class="watermark-btn" @click="openWatermarkModal">添加水印...</button>
            </div>
          </div>
          <div class="card">
            <label class="block">
              <input type="checkbox" v-model="options.unsupportedTip">
              在非指定的PDF阅读器中打开时的提示
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
          <div class="pwd-card" id="user-password-card">
            <!-- <h3>用户密码</h3> -->
            <div class="pwd-row">
              <button class="btn primary" @click="openPwdModal" :disabled="sending">设置文档密码</button>
              <button class="btn primary" @click="setExpire" :disabled="sending">
                {{ sending ? "设置中..." : "点击完成设置" }}
              </button>
            </div>
            <p v-if="runStatus" class="run-status">{{ runStatus }}</p>
          </div>
         
        </div>
      </div>
      
      <!-- 底部版权信息 -->
      <div class="app-footer">
        <div class="logo">
            <img src="./assets/images/logo.png" alt="logo" />
          </div>
        <div>
            ©远信软件技术服务有限公司&nbsp;&nbsp;2026&nbsp;&nbsp;未经许可，禁止售卖、传播，违者必究！
        </div>
      </div>
      <!-- 注册模块-->
      <div v-if="showRegisterModal" class="modal-overlay">
        <div class="modal register-modal">
          <h3 class="modal-title">在线注册</h3>
          <div class="register-form">
            <div class="register-label">机器码：</div>
            <div class="register-machine-row">
              <span class="register-machine-value">{{ machineCode }}</span>
              <button class="register-copy-btn" @click="copyMachineCode">复制</button>
            </div>
            <div class="register-label">注册码：</div>
            <input class="register-input" v-model="activationCode" placeholder="请输入注册码" />
          </div>
          <div class="register-actions">
            <button class="register-btn register-btn-primary" @click="onRegister" :disabled="registerLoading">注册</button>
            <button class="register-btn" @click="onCancel" :disabled="registerLoading">取消</button>
          </div>
        </div>
      </div>
      <div v-if="showPwdModal" class="modal-overlay">
        <div class="modal">
          <h3>批量设置文档密码</h3>
          <input v-model="pwd" type="password" placeholder="请输入至少6位的密码" autofocus @keyup.enter="confirmPwdModal" />
          <div class="modal-actions">
            <button @click="confirmPwdModal">确定</button>
            <button @click="cancelPwdModal">取消</button>
          </div>
        </div>
      </div>
      <div v-if="showWatermarkModal" class="modal-overlay">
        <div class="modal watermark-modal">
          <h3 class="modal-title">水印设置</h3>
          <div class="watermark-form">
            <label class="watermark-label">字体名称：</label>
            <select v-model="watermarkFontName" class="watermark-select">
              <option v-for="font in watermarkFontOptions" :key="font" :value="font">{{ font }}</option>
            </select>
            <label class="watermark-label">字体颜色：</label>
            <input v-model="watermarkColor" type="color" class="watermark-color" />
            <label class="watermark-label">字体大小：</label>
            <input v-model="watermarkFontSize" type="number" class="watermark-number" min="6" max="200" />
            <label class="watermark-label">位置：</label>
            <select v-model="watermarkPos" class="watermark-select">
              <option value="tl">左上</option>
              <option value="tc">上中</option>
              <option value="tr">右上</option>
              <option value="l">左中</option>
              <option value="c">居中</option>
              <option value="r">右中</option>
              <option value="bl">左下</option>
              <option value="bc">下中</option>
              <option value="br">右下</option>
            </select>
            <label class="watermark-label">旋转角度：</label>
            <input v-model="watermarkRotation" type="number" class="watermark-number" min="-180" max="180" />
            <label class="watermark-label">透明度：</label>
            <input v-model="watermarkOpacity" type="number" class="watermark-number" step="0.05" min="0" max="1" />
            <label class="watermark-label">设置间距：</label>
            <input v-model="watermarkSpacing" type="number" class="watermark-number" min="50" max="1000" />
            <label class="watermark-label">铺满页面：</label>
            <select v-model="watermarkTiled" class="watermark-select">
              <option :value="false">否</option>
              <option :value="true">是</option>
            </select>
           <!--  <label class="watermark-label">不嵌入字体：</label>
            <label class="watermark-toggle-inline">
              <input v-model="watermarkNoEmbed" type="checkbox" />
              减小体积
            </label> -->
            <label class="watermark-label">水印内容：</label>
            <textarea v-model="watermarkText" rows="2"></textarea>
          </div>
          <div class="modal-actions">
            <button @click="confirmWatermarkModal">确定</button>
            <button @click="cancelWatermarkModal">取消</button>
          </div>
        </div>
      </div>
    </div>
  </template>
  
  <script setup>
  import { computed, onMounted, ref, watch } from "vue"
import { GetMachineCode, GetTitleWithRegStatus, IsRegistered, MessageDialog, OpenDirectoryAndListFiles, OpenDirectoryDialog, OpenMultipleFilesDialog, Register, SetExpiry } from "../wailsjs/go/main/App.js"
import { engine } from "../wailsjs/go/models"
import { EventsOn, LogPrint, WindowSetTitle } from "../wailsjs/runtime/runtime.js"
  
  const activeTab = ref("file")
  
  const files = ref([
   
  ])
  
  const options = ref({
    copy: false,
    edit: false,
    convert: false,
    print: false,
    unsupportedTip: false,
    expiredTip: false,
  })
  const watermarkEnabled = ref(false)
  const watermarkText = ref("")
  const watermarkDesc = ref("")
  const showWatermarkModal = ref(false)
  const watermarkFontName = ref("Helvetica")
  const watermarkFontSize = ref(25)
  const watermarkRotation = ref(15)
  const watermarkOpacity = ref(0.3)
  const watermarkColor = ref("#808080")
  const watermarkPos = ref("c")
  const watermarkSpacing = ref(200)
  const watermarkTiled = ref(false)
  const watermarkNoEmbed = ref(true)
  const latinWatermarkFonts = [
    "Helvetica",
    "Helvetica-Bold",
    "Helvetica-Oblique",
    "Times-Roman",
    "Times-Bold",
    "Times-Italic",
    "Courier",
    "Courier-Bold",
    "Courier-Oblique",
  ]
  const cjkWatermarkFonts = [
    "MicrosoftYaHei",
    "SimSun",
    "SimHei",
    "FangSong",
    "KaiTi",
    "DengXian",
  ]
  const watermarkFontOptions = computed(() => {
    if (/[^\x00-\x7F]/.test(watermarkText.value || "")) return cjkWatermarkFonts
    return [...latinWatermarkFonts, ...cjkWatermarkFonts]
  })

  watch(watermarkText, (val) => {
    if (/[^\x00-\x7F]/.test(val || "")) {
      const picked = (watermarkFontName.value || "").trim()
      if (!cjkWatermarkFonts.includes(picked)) {
        watermarkFontName.value = "MicrosoftYaHei"
      }
    }
  })
  
  const unsupportedText = ref(
    "文档显示错误！请使用Adobe Reader、PDF-Xchange或福昕PDF阅读器打开当前文档！"
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
  const runStatus = ref("")
  const fileInput = ref(null)
  
  function addFile() {
    // 调用 Go 原生多文件选择对话框
    (async () => {
      try {
        const paths = await OpenMultipleFilesDialog()
        addFilesAndUnique(paths)
      } catch (err) {
        console.error('OpenMultipleFilesDialog error', err)
        await MessageDialog('提示', err, 'error')
      }
    })()
  }
  // 新增文件并且去重
  function addFilesAndUnique(paths){
    if (Array.isArray(paths) && paths.length > 0) {
            // 追加去重
            const existingPaths = new Set(files.value.map(f => f.path))
            const newPaths = paths.filter(p => !existingPaths.has(p))
            newPaths.forEach(p => files.value.push({ name: p.split(/[/\\]/
).pop(), path: p }))             
        }
  }
  async function addFolder() {  
      try {
      const pdfPaths = await OpenDirectoryAndListFiles()
     addFilesAndUnique(pdfPaths)
    } catch (err) {
     LogPrint("选择文件夹已取消或出错"+err)  
      await MessageDialog('提示', "选择文件夹已取消或出错", 'warning')
    }
  }
  
  function removeFileAll() {
    files.value = []
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
        await MessageDialog('提示', '密码至少6位', 'warning')
        return
      }
      showPwdModal.value = false
    }
    function cancelPwdModal() {
      pwd.value = ""
      showPwdModal.value = false
    }
    async function MessageDialogWithType(title, message) {
      return MessageDialog(title, message, dialogTypeForTitle(title))
    }
    function dialogTypeForTitle(title) {
      const t = String(title || "")
      if (t.includes("错误")) return "error"
      if (t.includes("警告")) return "warning"
      if (t.includes("询问")) return "question"
      return "info"
    }
    function openWatermarkModal() {
      showWatermarkModal.value = true
    }
    function confirmWatermarkModal() {
      watermarkText.value = (watermarkText.value || "").trim()
      watermarkDesc.value = buildWatermarkDesc()
      watermarkEnabled.value = watermarkText.value.length > 0
      showWatermarkModal.value = false
    }
    function cancelWatermarkModal() {
      showWatermarkModal.value = false
    }
    function buildWatermarkDesc() {
      const pickedFont = (watermarkFontName.value || "Helvetica").trim()
      const fontSize = Number(watermarkFontSize.value) || 36
      const rotation = Number(watermarkRotation.value) || 0
      const opacity = Math.min(1, Math.max(0, Number(watermarkOpacity.value) || 0.3))
      const color = (watermarkColor.value || "#808080").trim()
      const pos = (watermarkPos.value || "c").trim()
      const hasCJK = /[^\x00-\x7F]/.test(watermarkText.value || "")
      const fontName = hasCJK
        ? (cjkWatermarkFonts.includes(pickedFont) ? pickedFont : "MicrosoftYaHei")
        : pickedFont
      const noEmbed = watermarkNoEmbed.value && /[^\x00-\x7F]/.test(watermarkText.value || "")
      const scriptName = noEmbed ? ", scriptname:HANS" : ""
      return `fontname:${fontName}, points:${fontSize}, scale:1 abs, fillcolor:${color}, opacity:${opacity}, rot:${rotation}, pos:${pos}${scriptName}`
    }
  async function setExpire() {
    // 直接提交给后端，不再选择保存路径
    if (sending.value) return
    runStatus.value = ""
    if (files.value.length === 0) {
      await MessageDialog('提示', '请先通过“添加文件”选择要处理的文件', 'warning')
      return
    }
    // 验证密码长度
    if ((pwd.value && pwd.value.length < 6)) {
      await MessageDialog('提示', '密码至少6位', 'warning')
      return
    }
    // 弹出原生选择目录对话框以获取保存目录（OpenDirectoryDialog）
    //runStatus.value = "请选择输出目录..."
    let folderPath = ''
    try {
      folderPath = await OpenDirectoryDialog()
      // 打印返回的保存目录到 Wails 日志（使用 runtime 的 LogPrint）
      try { LogPrint(folderPath) } catch (e) { console.debug('LogPrint error', e) }
    } catch (err) {
      console.error('OpenDirectoryDialog error', err)
      await MessageDialog('提示', '选择目录已取消或出错', 'info')
      return
    }
    if (!folderPath) {
      await MessageDialog('提示', '未选择保存目录', 'warning')
      runStatus.value = ""
      return
    }
  
    const opts = new engine.Options()
    opts.Files = files.value.map(f => f.path).join(';')
    opts.OutputDir = folderPath
    opts.StartTime = startTime.value ? new Date(startTime.value).toISOString() : null
    opts.EndTime = endTime.value ? new Date(endTime.value).toISOString() : null
    opts.WatermarkEnabled = watermarkEnabled.value
    opts.WatermarkText = watermarkText.value
    opts.WatermarkDesc = watermarkDesc.value
    opts.WatermarkTiled = watermarkTiled.value
    opts.WatermarkSpacing = Number(watermarkSpacing.value) || 0
    if(options.value.expiredTip){
      opts.ExperiredText = expiredText.value
    }
    if(options.value.unsupportedTip){
      opts.UnsupportedText = unsupportedText.value
    }
    opts.AllowedPrint = options.value.print
    opts.AllowedCopy = options.value.copy
    // 编辑
    opts.AllowedEdit = options.value.edit
    opts.AllowedConvert = options.value.convert
    // 用户密码绑定
    opts.PwdEnabled = pwdEnabled.value
    opts.UserPassword = pwd.value

    try {
      sending.value = true
      runStatus.value = "正在批量设置，请稍候..."
      const res =await SetExpiry(opts)
      await MessageDialog('提示', res,'')
      LogPrint(res)
      //runStatus.value = "设置完成"
    } catch (err) {
      console.error('SetExpiry error', err)
      await MessageDialog('错误', '设置文档时效请求失败：' + (err && err.message ? err.message : err), 'error')
      //runStatus.value = "设置失败"
    } finally {
      sending.value = false
    }
  }

  async function onRegister() {
    if (!activationCode.value) {
      await MessageDialog('提示', '请输入注册码', 'warning')
      return
    }
    try {
      registerLoading.value = true
      const res = await Register(activationCode.value)
      await MessageDialog('提示', res || '注册成功', 'info')
      showRegisterModal.value = false
      // 更新已注册状态和机器码显示
      machineCode.value = await GetMachineCode()
    } catch (err) {
      console.error('Register error', err)
      await MessageDialog('错误',  (err && err.message ? err.message : err), 'error')
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
        await MessageDialog('提示', '机器码已复制到剪贴板', 'info')
      } else {
        // fallback: create temporary textarea
        const ta = document.createElement('textarea')
        ta.value = machineCode.value
        document.body.appendChild(ta)
        ta.select()
        try { document.execCommand('copy') } catch (e) { console.debug('execCommand copy failed', e) }
        ta.remove()
        await MessageDialog('提示', '机器码已复制到剪贴板', 'info')
      }
    } catch (e) {
      console.error('copyMachineCode error', e)
      await MessageDialog('错误', '复制失败：' + (e && e.message ? e.message : e), 'error')
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
          await MessageDialog('提示', '软件已注册', 'info')
          return
        }
        showRegisterModal.value = true
      })
      // 监听用户注册成功事件，获取最新标题并更新标题
      EventsOn('user:registered', async () => {
        const appTitle = await GetTitleWithRegStatus()
        await WindowSetTitle(appTitle)
      })
      // 监听用户注销事件，获取最新标题并更新标题
      EventsOn('user:unregistered', async () => {
        const appTitle = await GetTitleWithRegStatus()
        await WindowSetTitle(appTitle)
      })
      // 监听文件添加事件
      EventsOn('user:filesSelected', async (newPaths) => {
        addFilesAndUnique(newPaths)
      })
      // 提示试用
      await MessageDialog('易诚无忧提示', '当前处于试用阶段！', 'info')
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
  height: 100%;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  --footer-h: 88px;
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
  /* 让左右区域在窗口高度内布局，避免被文件列表撑高 */
  flex: 1 1 auto;
  min-height: 0;
  box-sizing: border-box;
  padding-bottom: var(--footer-h); /* 预留底部固定版权高度 */
  gap: 10px;
}
  
.left-panel {
  flex: 0 0 45%;
  max-width: 45%;
  border: 1px solid #ddd;
  padding: 10px;
  background: #fff;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
  /* 让内部使用列布局，文件列表占据剩余高度可滚动 */
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}
  
  .action-buttons {
    display: flex;
    gap: 10px;
    margin-bottom: 10px;
    flex-shrink: 0;
  }
  
  .file-list {
    border: 1px solid #eee;
    overflow-y: auto; /* 纵向滚动 */
    padding: 5px;
    padding-bottom: 60px; /* 为底部预留更多空间，避免最后一条被 footer 遮挡 */
    padding-right: 8px; /* 预留滚动条空间，避免文本被挡住*/
    background: #fafafa;
    /* 占据除按钮外的剩余高度*/
    flex: 1 1 auto;
  }
  
  .file-item {
    display: grid;
    grid-template-columns: minmax(0, 2fr) minmax(0, 3fr) 72px; /* 名称、路径、删除按钮固定宽度*/
    align-items: center;
    column-gap: 8px;
    font-size: 14px;
    padding: 6px 8px;
    border-radius: 4px;
  }
  
  .file-item .name,
  .file-item .path {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis; /* 太长用省略号，避免挤压删除按钮*/
  }
  
  .file-item button {
    justify-self: end;
    width: 64px;           /* 删除按钮固定宽度，保证完全显示*/
    padding: 6px 0;
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
  flex: 1 1 0;
  min-width: 0;
  /* 与左侧同高，内部自然滚动页面 */
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 12px;
  overflow: hidden;
}
  
.card {
  border: 1px solid #eee;
  padding: 10px;
  margin-bottom: 0;
  background:#fff;
}
  .card-b {
    border: 1px solid #eee;
    padding: 10px;
    background:#fff;
  }
  
  .checkbox-group {
    display: flex;
  justify-content: center;   /* 水平居中整体 */
  align-items: center;
  gap: 32px;                 /* 选项之间的间距，可按需要调 */
  margin: 12px 0 6px;
  }
  .checkbox-group label {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
}
  .watermark-btn {
    padding: 6px 14px;
    border-radius: 6px;
    border: 1px solid #9fb7d6;
    background: #eaf2ff;
    color: #1a4c86;
    cursor: pointer;
  }
  .watermark-btn:hover {
    background: #dbe9ff;
  }
  .watermark-modal {
    width: 520px;
  }
  .watermark-form {
    display: grid;
    grid-template-columns: 90px 1fr 90px 1fr;
    gap: 10px 12px;
    align-items: center;
  }
  .watermark-label {
    font-size: 13px;
    color: #333;
  }
  .watermark-select,
  .watermark-number {
    width: 100%;
    padding: 6px 8px;
    border: 1px solid #d0d0d0;
    border-radius: 4px;
    box-sizing: border-box;
  }
  .watermark-color {
    width: 100%;
    height: 32px;
    border: 1px solid #d0d0d0;
    border-radius: 4px;
    box-sizing: border-box;
    padding: 2px 4px;
    background: #fff;
  }
  .watermark-toggle-inline {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
  }
  .watermark-form textarea {
    grid-column: 2 / 5;
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
    justify-content: center;   /* 水平整体居中 */
    align-items: center;
    gap: 24px;
    margin-top: 10px;
    margin-bottom: 18px;
  }
  .time-row span {
    font-size: 14px;
  }

  .time-row input[type="datetime-local"] {
    min-width: 200px;          /* 让两个时间输入宽度一致、更好看 */
  }
.logo {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  margin-top: 0;
}

.logo img {
  height: 32px;
  width: auto;
  max-width: 100%;
  margin-bottom: 4px;
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
  height: var(--footer-h);
  box-sizing: border-box;
  text-align: center;
  padding: 8px 12px;
  font-size: 12px;
  color: #888;
  background: rgba(255,255,255,0.85);
  border-top: 1px solid rgba(0,0,0,0.04);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

:global(html, body, #app) {
  height: 100%;
  margin: 0;
  overflow: hidden;
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
  width: 100%;
  box-sizing: border-box;
  margin: 0;
  box-shadow: 0 4px 12px rgba(0,0,0,0.06);
  text-align: center;
  padding: 18px;
  border: 1px solid rgba(0,0,0,0.06);
  background: #fff;
  min-height: 130px;
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
  .pwd-card .btn:disabled {
    opacity: 0.65;
    cursor: not-allowed;
  }
  .run-status {
    margin: 12px 0 0;
    font-size: 13px;
    color: #1a4c86;
  }

  .modal-title {
    margin: 0 0 18px;
    text-align: center;
    font-size: 22px;
    font-weight: 700;
  }

  .register-modal {
    padding: 26px 32px 22px;
  }

  .register-form {
    display: grid;
    grid-template-columns: 70px 1fr;
    gap: 14px 12px;
    align-items: center;
    margin-bottom: 10px;
  }

  .register-label {
    font-size: 14px;
    text-align: right;
    color: #222;
  }

  .register-machine-row {
    display:flex;
    align-items:center;
    gap:10px;
  }

  .register-machine-value {
    flex: 1 1 auto;
    font-family: monospace;
    background:#f4f4f4;
    padding:6px 10px;
    border-radius:4px;
    color:#333;
    font-size: 13px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .register-input {
    width: 100%;
    padding: 10px 10px;
    box-sizing: border-box;
    border-radius: 4px;
    border: 1px solid #d0d0d0;
    font-size: 14px;
  }

  .register-actions {
    margin-top: 18px;
    display:flex;
    justify-content: flex-end;
    gap: 10px;
  }

  .register-copy-btn {
    padding:6px 12px;
    border:1px solid #d9d9d9;
    background:#fff;
    cursor:pointer;
    border-radius:4px;
    font-size: 13px;
  }

  .register-btn {
    padding: 7px 18px;
    border-radius: 4px;
    border: 1px solid #d0d0d0;
    background: #fff;
    cursor: pointer;
    min-width: 86px;
    font-size: 14px;
  }

  .register-btn-primary {
    background: #1677ff;
    border-color: #1677ff;
    color: #fff;
  }
  </style>









