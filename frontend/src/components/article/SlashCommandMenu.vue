<template>
  <div class="slash-command-menu">
    <div class="slash-command-header">
      插入内容
    </div>

    <div class="slash-command-list">
      <button
        v-for="(item, index) in items"
        :key="item.id"
        class="slash-command-item"
        :class="{
          selected: index === selectedIndex,
        }"
        @mousedown.prevent="selectItem(index)"
      >
        <span class="command-icon">
          {{ item.icon }}
        </span>

        <span class="command-content">
          <span class="command-title">
            {{ item.title }}
          </span>

          <span class="command-description">
            {{ item.description }}
          </span>
        </span>

        <span
          v-if="item.shortcut"
          class="command-shortcut"
        >
          {{ item.shortcut }}
        </span>
      </button>

      <div
        v-if="items.length === 0"
        class="slash-command-empty"
      >
        没有找到相关命令
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

export interface SlashCommandItem {
  id: string

  title: string

  description: string

  icon: string

  shortcut?: string

  keywords?: string[]

  command: (props: {
    editor: any
    range: {
      from: number
      to: number
    }
  }) => void
}

interface Props {
  items: SlashCommandItem[]

  command: (item: SlashCommandItem) => void

  query: string
}

const props = defineProps<Props>()

const selectedIndex = ref(0)

/* =========================
   搜索结果变化时重置选项
========================= */

watch(
  () => props.items,
  () => {
    selectedIndex.value = 0
  },
)

/* =========================
   上移
========================= */

function selectPrevious() {
  if (props.items.length === 0) {
    return
  }

  selectedIndex.value =
    (selectedIndex.value - 1 + props.items.length) %
    props.items.length
}

/* =========================
   下移
========================= */

function selectNext() {
  if (props.items.length === 0) {
    return
  }

  selectedIndex.value =
    (selectedIndex.value + 1) %
    props.items.length
}

/* =========================
   执行
========================= */

function selectItem(index: number) {
  const item = props.items[index]

  if (!item) {
    return
  }

  selectedIndex.value = index

  props.command(item)
}

/* =========================
   键盘事件
========================= */

function onKeyDown(event: KeyboardEvent) {
  if (event.key === 'ArrowUp') {
    selectPrevious()

    return true
  }

  if (event.key === 'ArrowDown') {
    selectNext()

    return true
  }

  if (event.key === 'Enter') {
    if (props.items.length === 0) {
      return false
    }

    selectItem(selectedIndex.value)

    return true
  }

  return false
}

defineExpose({
  onKeyDown,
})
</script>

<style scoped>
.slash-command-menu {
  width: 320px;

  max-height: 420px;

  overflow: hidden;

  border: 1px solid #e5e7eb;

  border-radius: 8px;

  background: #ffffff;

  box-shadow:
    0 8px 24px rgba(0, 0, 0, 0.12);

  color: #111827;
}

.slash-command-header {
  padding: 10px 12px 8px;

  border-bottom: 1px solid #f3f4f6;

  color: #9ca3af;

  font-size: 12px;

  font-weight: 500;
}

.slash-command-list {
  max-height: 370px;

  overflow-y: auto;

  padding: 5px;
}

.slash-command-item {
  width: 100%;

  min-height: 48px;

  display: flex;

  align-items: center;

  gap: 10px;

  padding: 7px 9px;

  border: none;

  border-radius: 6px;

  background: transparent;

  text-align: left;

  cursor: pointer;
}

.slash-command-item:hover,
.slash-command-item.selected {
  background: #f3f4f6;
}

.command-icon {
  width: 30px;

  height: 30px;

  flex: 0 0 30px;

  display: flex;

  align-items: center;

  justify-content: center;

  border: 1px solid #e5e7eb;

  border-radius: 6px;

  background: #ffffff;

  color: #374151;

  font-size: 14px;
}

.command-content {
  min-width: 0;

  flex: 1;

  display: flex;

  flex-direction: column;

  gap: 1px;
}

.command-title {
  color: #111827;

  font-size: 13px;

  line-height: 18px;

  font-weight: 500;
}

.command-description {
  color: #9ca3af;

  font-size: 11px;

  line-height: 16px;

  white-space: nowrap;

  overflow: hidden;

  text-overflow: ellipsis;
}

.command-shortcut {
  flex-shrink: 0;

  color: #9ca3af;

  font-size: 11px;
}

.slash-command-empty {
  padding: 20px;

  color: #9ca3af;

  text-align: center;

  font-size: 13px;
}
</style>