<script setup>
import { ref } from 'vue'
import { GenerateRegCode } from "../wailsjs/go/main/App.js"
const machine = ref('')
const code = ref('')
const error = ref('')
const loading = ref(false)

async function generate() {
  error.value = ''
  code.value = ''
  const m = machine.value.trim()
  if (!m) { error.value = '请输入机器码'; return }
  loading.value = true
  try {
    // 调用 wails 后端 GenerateRegCode
    const resp = await GenerateRegCode(m)
    // resp 可能是字符串，也可能是 { ok, code, error }
    if (typeof resp === 'string') {
      code.value = resp || ''
    } else if (resp && resp.ok) {
      code.value = resp.code || ''
    } else {
      error.value = resp && resp.error ? resp.error : '生成失败'
    }
  } catch (e) {
    error.value = e && e.message ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function copyCode() {
  if (!code.value) return
  try {
    await navigator.clipboard.writeText(code.value)
  } catch (_) {
    // fallback
    const ta = document.createElement('textarea')
    ta.value = code.value
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
  }
}
</script>

<template>
  <div class="wrap">
    <div class="card">
      <h1>注册码生成器</h1>
      <div class="desc">输入机器码，生成注册码</div>

      <div class="form">
        <input v-model="machine" placeholder="请输入机器码，例如：ABCDEF-123456" @keyup.enter="generate" />
        <button @click="generate" :disabled="loading">{{ loading ? '生成中...' : '生成注册码' }}</button>
      </div>

      <div class="result" v-if="code">
        <div class="label">生成的注册码：</div>
        <textarea class="value" readonly :value="code" rows="8"></textarea>
        <div class="actions">
          <button @click="copyCode">复制</button>
        </div>
      </div>

      <div class="error" v-if="error">{{ error }}</div>

    </div>
  </div>
</template>

<style scoped>
.wrap { min-height:100vh; display:flex; align-items:center; justify-content:center; background:#1b2636; color:#e6eef8; padding:20px }
.card { width:540px; background:#0f1720; border-radius:8px; padding:22px; box-shadow:0 8px 30px rgba(0,0,0,0.5) }
h1 { margin:0 0 6px 0; font-size:20px }
.desc { color:#9fb0c8; font-size:13px; margin-bottom:12px }
.form { display:flex; gap:8px }
input { flex:1; padding:10px 12px; border-radius:6px; border:1px solid #223042; background:#071026; color:#e6eef8 }
button { padding:8px 12px; border-radius:6px; border:0; background:#2563eb; color:#fff; cursor:pointer }
.result { margin-top:14px; padding:12px; background:#071623; border-radius:6px; border:1px solid #233040 }
.label { color:#9fb0c8; font-size:13px }
.value {
  width:100%;
  margin-top:8px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, 'Roboto Mono', 'Segoe UI Mono', monospace;
  font-size:13px;
  line-height:1.4;
  color:#e6eef8;
  background:#061018;
  border:1px solid #233040;
  padding:10px;
  border-radius:6px;
  white-space:pre-wrap;
  word-break:break-word;
  overflow:auto;
  resize:vertical;
}
.actions { margin-top:8px }
.error { margin-top:10px; color:#ffb4b4 }
.note { margin-top:12px; color:#95a7bf; font-size:12px }
</style>
