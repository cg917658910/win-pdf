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
            <button @click="setExpire">设置文档时效</button>
          </div>
  
          <div class="file-list">
            <div class="file-item" v-for="(file, index) in files" :key="index">
              <span>{{ file.name }}</span>
              <span class="path">{{ file.path }}</span>
              <button @click="removeFile(index)" style="margin-left:8px">删除</button>
            </div>
          </div>
  
          <div class="tips">
            <p>说明：</p>
            <p>1、添加进来的PDF文件以文件名加路径的形式展示。</p>
            <p>2、在文件列表上右键选中某个文件可以删除。</p>
          </div>
        </div>
  
        <!-- 右侧 -->
        <div class="right-panel">
          <div class="card">
            <h3>加密选项</h3>
  
            <div class="checkbox-group">
              <label><input type="checkbox" v-model="options.copy"> 允许复制</label>
              <label><input type="checkbox" v-model="options.edit"> 允许编辑</label>
              <label><input type="checkbox" v-model="options.convert"> 允许转换</label>
              <label><input type="checkbox" v-model="options.print"> 允许打印</label>
            </div>
  
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
  
          <div class="logo">
            <img src="./assets/images/logo.png" alt="logo" />
            <!-- <div>
              <div class="cn">远信软件</div>
              <div class="en">Yuanxin Software</div>
            </div> -->
          </div>
        </div>
      </div>
    </div>
  </template>
  
  <script setup>
  import { ref } from "vue"
  import { SetExpiry } from "../wailsjs/go/main/App.js"
  import { engine } from "../wailsjs/go/models"
  import { CanResolveFilePaths, ResolveFilePaths, LogPrint } from "../wailsjs/runtime/runtime.js"
  import { OpenDirectoryDialog, MessageDialog, OpenMultipleFilesDialog } from "../wailsjs/go/main/App.js"
  
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
    ;(async () => {
      try {
        const paths = await OpenMultipleFilesDialog()
        if (Array.isArray(paths) && paths.length > 0) {
          files.value = paths.map(p => ({ name: p.split(/[/\\]/).pop(), path: p }))
        }
      } catch (err) {
        console.error('OpenMultipleFilesDialog error', err)
        await MessageDialog('提示', '选择文件已取消或出错')
      }
    })()
  }
  
  function removeFile(index) {
    if (index >= 0 && index < files.value.length) {
      files.value.splice(index, 1)
    }
  }
  
  async function setExpire() {
    // 直接提交给后端，不再选择保存路径
    if (sending.value) return
    if (files.value.length === 0) {
      await MessageDialog('提示', '请先通过“添加文件”选择要处理的文件')
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
    // 构造输出文件路径
    //const ts = new Date().toISOString().replace(/[:.]/g, '-')
   /*  if (files.value.length === 1) {
      const base = files.value[0].name.replace(/\.[^/.]+$/, '')
      opts.OutputDir = folderPath + (folderPath.endsWith('\\') || folderPath.endsWith('/') ? '' : '\\') + base + "_expiry.pdf"
    } else {
      opts.OutputDir = folderPath + (folderPath.endsWith('\\') || folderPath.endsWith('/') ? '' : '\\') + "batch_expiry_" + ts + ".pdf"
    } */
    opts.OutputDir = folderPath
    opts.StartTime = startTime.value ? new Date(startTime.value).toISOString() : null
    opts.EndTime = endTime.value ? new Date(endTime.value).toISOString() : null
    opts.Watermark = ""
    opts.ExperiredText = expiredText.value
    opts.UnsupportedText = unsupportedText.value
    opts.DisablePrint = !options.value.print
    opts.DisableCopy = !options.value.copy

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
  </script>
  
  <style scoped>
  .app-container {
    font-family: "Microsoft YaHei", sans-serif;
    padding: 10px;
    color: #222;
    background: linear-gradient(180deg, #f7f9fb 0%, #eef3f8 100%);
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
    height: 260px;
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
  
  .checkbox-group {
    display: flex;
    gap: 20px;
    margin-bottom: 10px;
  }
  
  .block {
    display: block;
    margin-top: 8px;
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
    display: flex;
    align-items: center;
    gap: 10px;
    margin-top: 20px;
  }
  
  .logo img {
    width: 120px;
    max-width: 100%;
    height: auto;
  }
  
  .cn {
    font-size: 18px;
    font-weight: bold;
  }
  
  .en {
    font-size: 14px;
  }
  </style>
