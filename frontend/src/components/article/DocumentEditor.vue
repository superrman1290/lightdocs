<template>
  <div class="document-editor">
    <!-- 顶部工具栏 -->
    <div class="editor-toolbar">
      <div class="toolbar-left">
        <!-- 撤销 / 重做 -->
        <button
          class="toolbar-button"
          :disabled="!editor?.can().undo()"
          title="撤销"
          @click="editor?.chain().focus().undo().run()"
        >
          ↶
        </button>

        <button
          class="toolbar-button"
          :disabled="!editor?.can().redo()"
          title="重做"
          @click="editor?.chain().focus().redo().run()"
        >
          ↷
        </button>

        <span class="toolbar-divider"></span>

        <!-- 标题 -->
        <select
          class="toolbar-select"
          :value="currentHeading"
          @change="changeHeading"
        >
          <option value="paragraph">正文</option>
          <option value="1">标题 1</option>
          <option value="2">标题 2</option>
          <option value="3">标题 3</option>
        </select>

        <span class="toolbar-divider"></span>

        <!-- 基础格式 -->
        <button
          class="toolbar-button"
          :class="{ active: editor?.isActive('bold') }"
          title="粗体"
          @click="editor?.chain().focus().toggleBold().run()"
        >
          <strong>B</strong>
        </button>

        <button
          class="toolbar-button"
          :class="{ active: editor?.isActive('italic') }"
          title="斜体"
          @click="editor?.chain().focus().toggleItalic().run()"
        >
          <em>I</em>
        </button>

        <button
          class="toolbar-button"
          :class="{ active: editor?.isActive('underline') }"
          title="下划线"
          @click="editor?.chain().focus().toggleUnderline().run()"
        >
          <u>U</u>
        </button>

        <button
          class="toolbar-button"
          :class="{ active: editor?.isActive('strike') }"
          title="删除线"
          @click="editor?.chain().focus().toggleStrike().run()"
        >
          <s>S</s>
        </button>

        <button
          class="toolbar-button"
          :class="{ active: editor?.isActive('highlight') }"
          title="高亮"
          @click="editor?.chain().focus().toggleHighlight().run()"
        >
          H
        </button>

        <span class="toolbar-divider"></span>

        <!-- 列表 -->
        <button
          class="toolbar-button"
          :class="{ active: editor?.isActive('bulletList') }"
          title="无序列表"
          @click="editor?.chain().focus().toggleBulletList().run()"
        >
          •☰
        </button>

        <button
          class="toolbar-button"
          :class="{ active: editor?.isActive('orderedList') }"
          title="有序列表"
          @click="editor?.chain().focus().toggleOrderedList().run()"
        >
          1☰
        </button>

        <button
          class="toolbar-button"
          :class="{ active: editor?.isActive('blockquote') }"
          title="引用"
          @click="editor?.chain().focus().toggleBlockquote().run()"
        >
          ❝
        </button>

        <span class="toolbar-divider"></span>

        <!-- 链接 -->
        <button
          class="toolbar-button"
          :class="{ active: editor?.isActive('link') }"
          title="插入链接"
          @click="setLink"
        >
          🔗
        </button>

        <!-- 图片 -->
        <button
          class="toolbar-button"
          title="插入图片"
          @click="insertImage"
        >
          🖼
        </button>

        <!-- 表格 -->
        <button
          class="toolbar-button"
          title="插入表格"
          @click="insertTable"
        >
          ▦
        </button>

        <!-- 代码块 -->
        <button
          class="toolbar-button"
          :class="{ active: editor?.isActive('codeBlock') }"
          title="代码块"
          @click="toggleCodeBlock"
        >
          &lt;/&gt;
        </button>
      </div>

      <div class="toolbar-right">
        <!-- 编辑模式 -->
        <button
          class="mode-button"
          :class="{ active: mode === 'visual' }"
          @click="switchToVisual"
        >
          可视化
        </button>

        <button
          class="mode-button"
          :class="{ active: mode === 'markdown' }"
          @click="switchToMarkdown"
        >
          Markdown
        </button>
      </div>
    </div>

    <!-- 代码块语言栏 -->
    <div
      v-if="editor?.isActive('codeBlock')"
      class="code-language-bar"
    >
      <span>代码语言：</span>

      <select v-model="currentCodeLanguage">
        <option
          v-for="language in codeLanguages"
          :key="language"
          :value="language"
        >
          {{ language }}
        </option>
      </select>
    </div>

    <!-- 主编辑区域 -->
    <div class="editor-main">

      <!-- 左侧：正文 -->
      <main class="writing-panel">

        <!-- Markdown 编辑 -->
        <textarea
          v-if="mode === 'markdown'"
          v-model="markdownContent"
          class="markdown-editor"
          placeholder="使用 Markdown 编写文章..."
          @input="handleMarkdownInput"
        ></textarea>

        <!-- 可视化编辑 -->
        <div
          v-else
          class="visual-editor"
        >
          <EditorContent :editor="editor" />

          <!-- Bubble Menu：选中文字后出现 -->
          <BubbleMenu
            v-if="editor"
            :editor="editor"
            :options="{ placement: 'top' }"
          >
            <div class="bubble-menu">
              <button
                :class="{ active: editor.isActive('bold') }"
                @click="editor.chain().focus().toggleBold().run()"
              >
                B
              </button>

              <button
                :class="{ active: editor.isActive('italic') }"
                @click="editor.chain().focus().toggleItalic().run()"
              >
                I
              </button>

              <button
                :class="{ active: editor.isActive('underline') }"
                @click="editor.chain().focus().toggleUnderline().run()"
              >
                U
              </button>

              <button
                :class="{ active: editor.isActive('strike') }"
                @click="editor.chain().focus().toggleStrike().run()"
              >
                S
              </button>

              <button
                :class="{ active: editor.isActive('highlight') }"
                @click="editor.chain().focus().toggleHighlight().run()"
              >
                H
              </button>

              <button @click="setLink">
                🔗
              </button>
            </div>
          </BubbleMenu>
        </div>
      </main>

      <!-- 右侧：目录 -->
      <aside class="toc-panel">
        <div class="toc-header">
          <span>目录</span>
        </div>

        <div
          v-if="tocItems.length === 0"
          class="toc-empty"
        >
          暂无目录
        </div>

        <div
          v-else
          class="toc-list"
        >
          <button
            v-for="item in tocItems"
            :key="item.id"
            class="toc-item"
            :class="`toc-level-${item.level}`"
            @click="scrollToTocItem(item)"
          >
            {{ item.text }}
          </button>
        </div>
      </aside>
    </div>

    <el-dialog
      v-model="imagePickerVisible"
      title="选择图片"
      width="760px"
      append-to-body
    >
      <div
        v-loading="imagePickerLoading"
        class="image-picker"
      >
        <button
          v-for="image in imageOptions"
          :key="image.id"
          class="image-picker-item"
          @click="selectImage(image.url)"
        >
          <img
            :src="image.url"
            :alt="image.name"
          />
          <span>{{ image.name }}</span>
        </button>
        <el-empty
          v-if="!imagePickerLoading && imageOptions.length === 0"
          description="暂无图片，请先在图片菜单上传"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
} from 'vue'

import { Editor, EditorContent } from '@tiptap/vue-3'
import { BubbleMenu } from '@tiptap/vue-3/menus'

import StarterKit from '@tiptap/starter-kit'
import { Markdown } from '@tiptap/markdown'
import Image from '@tiptap/extension-image'
import Link from '@tiptap/extension-link'
import { TableKit } from '@tiptap/extension-table'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import Underline from '@tiptap/extension-underline'
import Highlight from '@tiptap/extension-highlight'

import { common, createLowlight } from 'lowlight'

import SlashCommand from '../../extensions/slashCommand'
import { imageService } from '../../services/imageService'
import { resolveApiURL } from '../../services/apiClient'
import type { ImageAsset } from '../../types/image'

const lowlight = createLowlight(common)

interface Props {
  modelValue?: string
  readonly?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  readonly: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'update:outline', value: Array<{
    id: string
    level: number
    text: string
    position: number
  }>): void
}>()

/* =========================
   编辑模式
========================= */

const mode = ref<'visual' | 'markdown'>('visual')

const markdownContent = ref(normalizeMarkdownForEditor(props.modelValue || ''))

/**
 * 文章正文统一保存为 Markdown。图片可能来自旧数据中的相对路径，
 * 进入编辑器前将其解析为 Go API 的可访问地址，避免被浏览器当作
 * 前端站点相对路径。
 */
const normalizeMarkdownImageSources = (markdown: string) => {
  return markdown.replace(
    /(!\[[^\]]*\]\()([^\s)]+)(\))/g,
    (_match, prefix: string, source: string, suffix: string) => `${prefix}${resolveApiURL(source)}${suffix}`,
  )
}

/**
 * Markdown treats a tab or four leading spaces as an indented code block.
 * Content pasted from office/chat applications often contains that indentation
 * even though it is ordinary prose. Remove that accidental indentation while
 * preserving explicitly fenced code blocks.
 */
function normalizeMarkdownForEditor(markdown: string) {
  const lines = markdown.split('\n')
  let inFence = false
  return lines.map(line => {
    const trimmed = line.trimStart()
    if (/^```/.test(trimmed) || /^~~~/.test(trimmed)) {
      inFence = !inFence
      return line
    }
    if (inFence) {
      return line
    }
    return line.replace(/^\t+/, '').replace(/^ {4,}/, '')
  }).join('\n')
}

// Older saves could contain an unlabeled fence created by the accidental
// indentation conversion. Unwrap only prose/list-like blocks; language-tagged
// and code-looking fences remain untouched.
function removeAccidentalProseFences(markdown: string) {
  return markdown.replace(/^```\s*\n([\s\S]*?)\n```\s*$/gm, (_match, body: string) => {
    const prose = /(^|\n)\s*(?:\d+[.、)]|[-*]\s+|[\u4e00-\u9fff])/.test(body)
    return prose ? body : _match
  })
}

function prepareMarkdownForEditor(markdown: string) {
  return normalizeMarkdownForEditor(removeAccidentalProseFences(markdown))
}

/* =========================
   代码语言
========================= */

const codeLanguages = [
  'plaintext',
  'javascript',
  'typescript',
  'html',
  'css',
  'json',
  'bash',
  'sql',
  'python',
  'go',
  'java',
]

const currentCodeLanguage = ref('plaintext')

/* =========================
   目录
========================= */

interface TocItem {
  id: string
  text: string
  level: number
  pos: number
}

const tocItems = ref<TocItem[]>([])
const imagePickerVisible = ref(false)
const imagePickerLoading = ref(false)
const imageOptions = ref<ImageAsset[]>([])
let imageSelectionPosition: number | null = null

/* =========================
   编辑器
========================= */

const editor = new Editor({
  editable: !props.readonly,

  content: prepareMarkdownForEditor(normalizeMarkdownImageSources(props.modelValue || '')),

  // Article content is stored as Markdown. Without this option, the initial
  // edit view treats ![](url) as literal text instead of an image node.
  contentType: 'markdown',

  editorProps: {
    handlePaste: (_view, event) => {
      const files = Array.from(event.clipboardData?.files ?? [])
        .filter(file => file.type.startsWith('image/'))

      if (files.length === 0) {
        return false
      }

      void uploadPastedImages(files)
      return true
    },
  },

  extensions: [
    StarterKit.configure({
      heading: {
        levels: [1, 2, 3],
      },

      codeBlock: false,
    }),

    Markdown,

    Image.configure({
      inline: false,
      allowBase64: true,
    }),

    Link.configure({
      openOnClick: false,
      autolink: true,
      linkOnPaste: true,
      defaultProtocol: 'https',
    }),

    TableKit.configure({
      table: {
        resizable: true,
      },
    }),

    Underline,

    Highlight.configure({
      multicolor: false,
    }),

    CodeBlockLowlight.configure({
      lowlight,
      defaultLanguage: 'plaintext',
      enableTabIndentation: true,
      tabSize: 2,
    }),

    SlashCommand,
  ],

  onUpdate: ({ editor }) => {
    if (mode.value !== 'visual') {
      return
    }

    const markdown = prepareMarkdownForEditor(editor.getMarkdown())

    markdownContent.value = markdown

    emit('update:modelValue', markdown)

    updateToc()
  },
})

/* =========================
   当前标题
========================= */

const currentHeading = computed(() => {
  if (!editor) {
    return 'paragraph'
  }

  if (editor.isActive('heading', { level: 1 })) {
    return '1'
  }

  if (editor.isActive('heading', { level: 2 })) {
    return '2'
  }

  if (editor.isActive('heading', { level: 3 })) {
    return '3'
  }

  return 'paragraph'
})

function changeHeading(event: Event) {
  const value = (event.target as HTMLSelectElement).value

  if (!editor) {
    return
  }

  if (value === 'paragraph') {
    editor.chain().focus().setParagraph().run()
    return
  }

  const level = Number(value) as 1 | 2 | 3

  editor
    .chain()
    .focus()
    .toggleHeading({ level })
    .run()
}

/* =========================
   目录生成
========================= */

function updateToc() {
  if (!editor) {
    return
  }

  const items: TocItem[] = []

  editor.state.doc.descendants((node, pos) => {
    if (node.type.name !== 'heading') {
      return
    }

    const level = Number(node.attrs.level)

    if (![1, 2, 3].includes(level)) {
      return
    }

    const text = node.textContent.trim()

    if (!text) {
      return
    }

    items.push({
      id: `heading-${pos}`,
      text,
      level,
      pos,
    })
  })

  tocItems.value = items

  emit(
    'update:outline',
    items.map((item) => ({
      id: item.id,
      level: item.level,
      text: item.text,
      position: item.pos,
    })),
  )
}

/* =========================
   点击目录
========================= */

function scrollToTocItem(item: TocItem) {
  if (!editor) {
    return
  }

  const dom = editor.view.nodeDOM(item.pos)

  if (dom instanceof HTMLElement) {
    dom.scrollIntoView({
      behavior: 'smooth',
      block: 'center',
    })
  }

  editor
    .chain()
    .focus()
    .setTextSelection(item.pos + 1)
    .run()
}

function scrollToHeading(position: number) {
  const dom = editor.view.nodeDOM(position)

  if (dom instanceof HTMLElement) {
    dom.scrollIntoView({
      behavior: 'smooth',
      block: 'center',
    })
  }

  editor
    .chain()
    .focus()
    .setTextSelection(position + 1)
    .run()
}

/* =========================
   图片
========================= */

function insertImage() {
  if (!editor) {
    return
  }

  imageSelectionPosition = editor.state.selection.from
  imagePickerVisible.value = true
  imagePickerLoading.value = true

  void imageService.getImages({ page: 1, pageSize: 100 })
    .then(result => {
      imageOptions.value = result.list
    })
    .catch(() => {
      imageOptions.value = []
    })
    .finally(() => {
      imagePickerLoading.value = false
    })
}

const selectImage = (url: string) => {
  imagePickerVisible.value = false
  const chain = editor.chain().focus()

  if (imageSelectionPosition !== null) {
    chain.setTextSelection(imageSelectionPosition)
  }

  chain.setImage({ src: url }).run()
  imageSelectionPosition = null
}

const uploadPastedImages = async (files: File[]) => {
  try {
    const uploaded = await imageService.uploadImages(files)

    for (const image of uploaded) {
      editor.chain().focus().setImage({ src: image.url }).run()
    }
  } catch (error) {
    console.error('粘贴图片上传失败：', error)
  }
}

/* =========================
   链接
========================= */

function setLink() {
  if (!editor) {
    return
  }

  if (editor.isActive('link')) {
    editor
      .chain()
      .focus()
      .unsetLink()
      .run()

    return
  }

  const url = window.prompt('请输入链接地址')

  if (!url) {
    return
  }

  editor
    .chain()
    .focus()
    .setLink({
      href: url,
    })
    .run()
}

/* =========================
   表格
========================= */

function insertTable() {
  if (!editor) {
    return
  }

  editor
    .chain()
    .focus()
    .insertTable({
      rows: 3,
      cols: 3,
      withHeaderRow: true,
    })
    .run()
}

/* =========================
   代码块
========================= */

function toggleCodeBlock() {
  if (!editor) {
    return
  }

  editor
    .chain()
    .focus()
    .toggleCodeBlock()
    .run()
}

/* =========================
   Markdown
========================= */

function switchToMarkdown() {
  if (!editor) {
    return
  }

  markdownContent.value = prepareMarkdownForEditor(editor.getMarkdown())

  mode.value = 'markdown'
}

function switchToVisual() {
  if (!editor) {
    return
  }

  mode.value = 'visual'

  nextTick(() => {
    editor.commands.setContent(
      prepareMarkdownForEditor(normalizeMarkdownImageSources(markdownContent.value || '')),
      {
        contentType: 'markdown',
        emitUpdate: false,
      },
    )

    updateToc()

    editor.commands.focus()
  })
}

function handleMarkdownInput() {
  emit(
    'update:modelValue',
    markdownContent.value,
  )
}

/* =========================
   外部数据同步
========================= */

watch(
  () => props.modelValue,
  (value) => {
    if (value === markdownContent.value) {
      return
    }

    markdownContent.value = value || ''

    if (mode.value !== 'visual') {
      return
    }

    editor.commands.setContent(
      prepareMarkdownForEditor(normalizeMarkdownImageSources(markdownContent.value)),
      {
        contentType: 'markdown',
        emitUpdate: false,
      },
    )

    updateToc()
  },
)

/* =========================
   生命周期
========================= */

onMounted(() => {
  const normalized = prepareMarkdownForEditor(props.modelValue || '')
  markdownContent.value = normalized
  if (normalized !== props.modelValue) {
    emit('update:modelValue', normalized)
  }
  updateToc()
})

onBeforeUnmount(() => {
  editor.destroy()
})

/* =========================
   暴露给父组件
========================= */

defineExpose({
  editor,

  switchToVisual,

  switchToMarkdown,

  getMarkdown: () => prepareMarkdownForEditor(editor.getMarkdown()),

  updateToc,

  scrollToHeading,
})
</script>

<style scoped>
.document-editor {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  overflow: hidden;
}

.image-picker {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  min-height: 120px;
}

.image-picker-item {
  overflow: hidden;
  padding: 0;
  border: 1px solid #dbe4ef;
  border-radius: 8px;
  background: #ffffff;
  cursor: pointer;
  text-align: left;
}

.image-picker-item:hover {
  border-color: #409eff;
  box-shadow: 0 4px 14px rgb(64 158 255 / 20%);
}

.image-picker-item img {
  display: block;
  width: 100%;
  height: 110px;
  object-fit: cover;
}

.image-picker-item span {
  display: block;
  overflow: hidden;
  padding: 8px;
  color: #475569;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 640px) {
  .image-picker {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

/* =========================
   工具栏
========================= */

.editor-toolbar {
  height: 52px;
  min-height: 52px;

  display: flex;
  align-items: center;
  justify-content: space-between;

  padding: 0 16px;

  border-bottom: 1px solid #e5e7eb;

  background: #ffffff;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 4px;
}

.toolbar-button {
  width: 32px;
  height: 32px;

  display: flex;
  align-items: center;
  justify-content: center;

  border: none;
  border-radius: 6px;

  background: transparent;

  color: #374151;

  cursor: pointer;

  font-size: 14px;
}

.toolbar-button:hover {
  background: #f3f4f6;
}

.toolbar-button.active {
  background: #e5e7eb;
}

.toolbar-button:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.toolbar-select {
  height: 32px;

  padding: 0 8px;

  border: 1px solid #e5e7eb;
  border-radius: 6px;

  background: #ffffff;

  color: #374151;

  outline: none;

  cursor: pointer;
}

.toolbar-divider {
  width: 1px;
  height: 22px;

  margin: 0 6px;

  background: #e5e7eb;
}

.mode-button {
  height: 32px;

  padding: 0 12px;

  border: 1px solid #e5e7eb;
  border-radius: 6px;

  background: #ffffff;

  color: #6b7280;

  cursor: pointer;
}

.mode-button:hover {
  background: #f9fafb;
}

.mode-button.active {
  background: #111827;
  border-color: #111827;
  color: #ffffff;
}

/* =========================
   代码语言
========================= */

.code-language-bar {
  height: 38px;

  display: flex;
  align-items: center;

  gap: 8px;

  padding: 0 20px;

  border-bottom: 1px solid #e5e7eb;

  background: #f9fafb;

  color: #6b7280;

  font-size: 13px;
}

.code-language-bar select {
  height: 28px;

  padding: 0 8px;

  border: 1px solid #d1d5db;
  border-radius: 5px;

  background: #ffffff;

  font-size: 13px;
}

/* =========================
   核心布局
========================= */

.editor-main {
  flex: 1;

  min-height: 0;

  display: flex;

  flex-direction: row;

  overflow: hidden;
}

/* =========================
   左侧正文
========================= */

.writing-panel {
  flex: 1;

  min-width: 0;

  height: 100%;

  overflow: auto;

  background: #ffffff;
}

.visual-editor {
  min-height: 100%;

  padding: 40px 56px 100px;
}

:deep(.ProseMirror) {
  max-width: 900px;

  min-height: calc(100vh - 180px);

  margin: 0 auto;

  outline: none;

  color: #1f2937;

  font-size: 16px;

  line-height: 1.8;
}

:deep(.ProseMirror p) {
  margin: 0 0 16px;
}

:deep(.ProseMirror h1) {
  margin: 32px 0 20px;

  font-size: 32px;

  line-height: 1.3;

  font-weight: 700;
}

:deep(.ProseMirror h2) {
  margin: 28px 0 16px;

  font-size: 24px;

  line-height: 1.4;

  font-weight: 700;
}

:deep(.ProseMirror h3) {
  margin: 24px 0 14px;

  font-size: 20px;

  line-height: 1.5;

  font-weight: 600;
}

:deep(.ProseMirror ul),
:deep(.ProseMirror ol) {
  padding-left: 28px;

  margin-bottom: 16px;
}

:deep(.ProseMirror blockquote) {
  margin: 20px 0;

  padding: 12px 16px;

  border-left: 4px solid #d1d5db;

  background: #f9fafb;

  color: #6b7280;
}

:deep(.ProseMirror img) {
  display: block;

  max-width: 100%;

  height: auto;

  margin: 20px auto;

  border-radius: 6px;
}

:deep(.ProseMirror a) {
  color: #2563eb;

  text-decoration: underline;
}

:deep(.ProseMirror code) {
  padding: 2px 5px;

  border-radius: 4px;

  background: #f3f4f6;

  font-family:
    Consolas,
    Monaco,
    monospace;

  font-size: 14px;
}

:deep(.ProseMirror pre) {
  margin: 20px 0;

  padding: 16px;

  overflow-x: auto;

  border-radius: 8px;

  background: #1f2937;

  color: #f9fafb;
}

:deep(.ProseMirror pre code) {
  padding: 0;

  background: transparent;

  color: inherit;
}

/* =========================
   Markdown
========================= */

.markdown-editor {
  width: 100%;

  height: 100%;

  min-height: 500px;

  box-sizing: border-box;

  padding: 40px 56px;

  border: none;

  outline: none;

  resize: none;

  color: #1f2937;

  background: #ffffff;

  font-family:
    Consolas,
    Monaco,
    monospace;

  font-size: 15px;

  line-height: 1.8;
}

/* =========================
   右侧目录
========================= */

.toc-panel {
  width: 250px;

  min-width: 250px;

  height: 100%;

  overflow-y: auto;

  box-sizing: border-box;

  border-left: 1px solid #e5e7eb;

  background: #fafafa;
}

.toc-header {
  height: 52px;

  display: flex;

  align-items: center;

  padding: 0 20px;

  border-bottom: 1px solid #e5e7eb;

  color: #111827;

  font-size: 14px;

  font-weight: 600;
}

.toc-list {
  padding: 12px 10px 30px;
}

.toc-item {
  display: block;

  width: 100%;

  padding: 7px 10px;

  border: none;

  border-radius: 5px;

  background: transparent;

  color: #6b7280;

  text-align: left;

  line-height: 1.5;

  cursor: pointer;

  font-size: 13px;

  white-space: nowrap;

  overflow: hidden;

  text-overflow: ellipsis;
}

.toc-item:hover {
  background: #f3f4f6;

  color: #111827;
}

.toc-level-1 {
  padding-left: 10px;

  color: #374151;

  font-weight: 600;
}

.toc-level-2 {
  padding-left: 24px;
}

.toc-level-3 {
  padding-left: 38px;

  font-size: 12px;
}

.toc-empty {
  padding: 30px 20px;

  color: #9ca3af;

  text-align: center;

  font-size: 13px;
}

/* =========================
   Bubble Menu
========================= */

.bubble-menu {
  display: flex;

  align-items: center;

  gap: 2px;

  padding: 4px;

  border: 1px solid #e5e7eb;

  border-radius: 7px;

  background: #ffffff;

  box-shadow:
    0 4px 12px rgba(0, 0, 0, 0.12);
}

.bubble-menu button {
  width: 30px;

  height: 30px;

  border: none;

  border-radius: 5px;

  background: transparent;

  color: #374151;

  cursor: pointer;
}

.bubble-menu button:hover {
  background: #f3f4f6;
}

.bubble-menu button.active {
  background: #e5e7eb;
}
</style>
