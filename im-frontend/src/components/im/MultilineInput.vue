<template>
  <div class="multiline-input">
    <textarea
      ref="textareaRef"
      :value="modelValue"
      @input="handleInput"
      @compositionstart="handleCompositionStart"
      @compositionend="handleCompositionEnd"
      @keydown.enter="handleEnter"
      @focus="handleFocus"
      :placeholder="placeholder"
      :style="{ height: textareaHeight + 'px' }"
      rows="1"
    ></textarea>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'

const props = defineProps<{
  modelValue: string
  placeholder?: string
  maxRows?: number
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'enter'): void
  (e: 'focus'): void
}>()

const textareaRef = ref<HTMLTextAreaElement | null>(null)
const textareaHeight = ref(40)
const lineHeight = 24
const maxRows = props.maxRows || 4
/** IME 组合输入中（如中文拼音选词），回车用于上屏而非发送 */
const isComposing = ref(false)

const handleCompositionStart = () => {
  isComposing.value = true
}

const handleCompositionEnd = () => {
  isComposing.value = false
}

const isImeEnter = (e: KeyboardEvent) => {
  return isComposing.value || e.isComposing || e.keyCode === 229
}

// 调整文本框高度
const adjustHeight = () => {
  if (!textareaRef.value) return

  // 如果内容为空，重置为初始高度
  if (!props.modelValue.trim()) {
    textareaHeight.value = 40
    return
  }
  
  // 重置高度以获取实际滚动高度
  textareaRef.value.style.height = 'auto'
  
  // 计算新高度
  const scrollHeight = textareaRef.value.scrollHeight
  const maxHeight = lineHeight * maxRows
  
  // 设置新高度，不超过最大高度
  textareaHeight.value = Math.min(Math.max(scrollHeight, 40), maxHeight)
}

// 处理输入
const handleInput = (e: Event) => {
  const target = e.target as HTMLTextAreaElement
  emit('update:modelValue', target.value)
  adjustHeight()
}

const handleFocus = () => {
  emit('focus')
}

// 处理回车键：IME 组合中不发送；Shift+Enter 换行；Enter 发送
const handleEnter = (e: KeyboardEvent) => {
  if (isImeEnter(e)) {
    return
  }

  if (e.shiftKey) {
    const lines = (textareaRef.value?.value || '').split('\n')
    if (lines.length >= maxRows) {
      e.preventDefault()
    }
    return
  }

  e.preventDefault()
  emit('enter')
}

// 监听值变化
watch(() => props.modelValue, () => {
  adjustHeight()
})

onMounted(() => {
  adjustHeight()
})
</script>

<style scoped>
.multiline-input {
  width: 100%;
  position: relative;
}

.multiline-input textarea {
  width: 100%;
  resize: none;
  border: none;
  outline: none;
  background: transparent;
  font-size: 15px;
  line-height: 24px;
  padding: 8px 12px;
  box-sizing: border-box;
  overflow-y: auto;
  display: block;
  min-height: 40px;
}

/* 自定义滚动条样式 */
.multiline-input textarea::-webkit-scrollbar {
  width: 4px;
}

.multiline-input textarea::-webkit-scrollbar-track {
  background: transparent;
}

.multiline-input textarea::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 2px;
}
</style> 