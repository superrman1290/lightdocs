import { Extension } from '@tiptap/core'
import Suggestion from '@tiptap/suggestion'
import { VueRenderer } from '@tiptap/vue-3'

import SlashCommandMenu, {
  type SlashCommandItem,
} from '../components/article/SlashCommandMenu.vue'

/* =====================================================
   Slash Command Extension
===================================================== */

const SlashCommand = Extension.create({
  name: 'slashCommand',

  addProseMirrorPlugins() {
    return [
      Suggestion({
        editor: this.editor,

        /*
         * 触发字符
         *
         * 注意：
         * Suggestion 本身会在输入 "/" 时启动。
         *
         * 我们在 allow 中进一步限制：
         * 必须已经输入 "/ " 才允许菜单真正出现。
         */
        char: '/',

        /*
         * 允许命令关键字中包含空格。
         *
         * 如果保持默认值 false，输入第二个字符空格时，
         * Suggestion 的匹配范围会立即结束，allow 将不会
         * 有机会让 "/ " 命令菜单保持激活。
         */
        allowSpaces: true,

        /*
         * 只允许在段落开头触发
         */
        startOfLine: true,

        /*
         * 菜单位置
         */
        placement: 'bottom-start',

        offset: {
          mainAxis: 6,
          crossAxis: 0,
        },

        /*
         * 将菜单选中的命令转发给对应的菜单项。
         *
         * Suggestion 默认的 command 是空函数；如果不在这里
         * 调用 item.command，菜单虽然可以显示，但点击后不会
         * 修改编辑器内容。
         */
        command: ({ editor, range, props }) => {
          props.command({
            editor,
            range,
          })
        },

        /*
         * 只允许普通段落
         */
        allow: ({ state, range }) => {
          const $from = state.doc.resolve(range.from)

          /*
           * 必须处于 paragraph
           */
          if ($from.parent.type.name !== 'paragraph') {
            return false
          }

          /*
           * 读取当前段落从行首到命令范围末尾的内容。
           * range.to 会包含已经输入的空格和命令关键字。
           */
          const lineText = state.doc.textBetween(
            $from.start(),
            range.to,
            '\n',
            '\n',
          )

          /*
           * 必须是：
           *
           * "/ "
           *
           * 或者：
           *
           * "/ 标题"
           *
           * "/ 代码"
           *
           * 等
           */
          return /^\/\s/.test(lineText)
        },

        /*
         * 获取命令
         */
        items: ({ query }) => {
          const items = createSlashCommands()

          const normalizedQuery =
            query.trim().toLowerCase()

          /*
           * 没有搜索内容
           *
           * 显示全部命令
           */
          if (!normalizedQuery) {
            return items
          }

          /*
           * 有搜索内容
           */
          return items.filter((item) => {
            const searchText = [
              item.title,
              item.description,
              ...(item.keywords || []),
            ]
              .join(' ')
              .toLowerCase()

            return searchText.includes(
              normalizedQuery,
            )
          })
        },

        /*
         * 菜单渲染
         */
        render: () => {
          let component:
            | VueRenderer
            | null = null

          let unmount:
            | (() => void)
            | undefined

          return {
            /*
             * 菜单开始
             */
            onStart: (props: any) => {
              component = new VueRenderer(
                SlashCommandMenu,
                {
                  props,
                  editor: props.editor,
                },
              )

              if (!component.element) {
                return
              }

              unmount = props.mount(
                component.element,
              )
            },

            /*
             * 内容更新
             */
            onUpdate: (props: any) => {
              component?.updateProps(props)
            },

            /*
             * 键盘事件
             */
            onKeyDown: (props: any) => {
              /*
               * ESC 关闭菜单
               */
              if (
                props.event.key === 'Escape'
              ) {
                return false
              }

              return (
                component?.ref?.onKeyDown(
                  props.event,
                ) || false
              )
            },

            /*
             * 菜单退出
             */
            onExit: () => {
              unmount?.()

              component?.destroy()

              component = null

              unmount = undefined
            },
          }
        },
      }),
    ]
  },
})

/* =====================================================
   创建命令
===================================================== */

function createSlashCommands(): SlashCommandItem[] {
  return [
    /* ---------------------------------------------
       正文
    --------------------------------------------- */

    {
      id: 'paragraph',

      title: '正文',

      description: '插入普通正文段落',

      icon: '¶',

      keywords: [
        '正文',
        '段落',
        'paragraph',
        'text',
      ],

      command: ({ editor, range }) => {
        editor
          .chain()
          .focus()
          .deleteRange(range)
          .setParagraph()
          .run()
      },
    },

    /* ---------------------------------------------
       一级标题
    --------------------------------------------- */

    {
      id: 'heading1',

      title: '一级标题',

      description: '插入一级标题',

      icon: 'H1',

      shortcut: '#',

      keywords: [
        '标题',
        '一级标题',
        'h1',
        'heading',
      ],

      command: ({ editor, range }) => {
        editor
          .chain()
          .focus()
          .deleteRange(range)
          .setHeading({
            level: 1,
          })
          .run()
      },
    },

    /* ---------------------------------------------
       二级标题
    --------------------------------------------- */

    {
      id: 'heading2',

      title: '二级标题',

      description: '插入二级标题',

      icon: 'H2',

      shortcut: '##',

      keywords: [
        '标题',
        '二级标题',
        'h2',
        'heading',
      ],

      command: ({ editor, range }) => {
        editor
          .chain()
          .focus()
          .deleteRange(range)
          .setHeading({
            level: 2,
          })
          .run()
      },
    },

    /* ---------------------------------------------
       三级标题
    --------------------------------------------- */

    {
      id: 'heading3',

      title: '三级标题',

      description: '插入三级标题',

      icon: 'H3',

      shortcut: '###',

      keywords: [
        '标题',
        '三级标题',
        'h3',
        'heading',
      ],

      command: ({ editor, range }) => {
        editor
          .chain()
          .focus()
          .deleteRange(range)
          .setHeading({
            level: 3,
          })
          .run()
      },
    },

    /* ---------------------------------------------
       无序列表
    --------------------------------------------- */

    {
      id: 'bullet-list',

      title: '无序列表',

      description: '创建项目符号列表',

      icon: '•',

      keywords: [
        '列表',
        '无序',
        'bullet',
        'list',
      ],

      command: ({ editor, range }) => {
        editor
          .chain()
          .focus()
          .deleteRange(range)
          .toggleBulletList()
          .run()
      },
    },

    /* ---------------------------------------------
       有序列表
    --------------------------------------------- */

    {
      id: 'ordered-list',

      title: '有序列表',

      description: '创建数字编号列表',

      icon: '1.',

      keywords: [
        '列表',
        '有序',
        'ordered',
        'list',
      ],

      command: ({ editor, range }) => {
        editor
          .chain()
          .focus()
          .deleteRange(range)
          .toggleOrderedList()
          .run()
      },
    },

    /* ---------------------------------------------
       引用
    --------------------------------------------- */

    {
      id: 'blockquote',

      title: '引用',

      description: '插入引用内容',

      icon: '❝',

      keywords: [
        '引用',
        'quote',
        'blockquote',
      ],

      command: ({ editor, range }) => {
        editor
          .chain()
          .focus()
          .deleteRange(range)
          .toggleBlockquote()
          .run()
      },
    },

    /* ---------------------------------------------
       代码块
    --------------------------------------------- */

    {
      id: 'code-block',

      title: '代码块',

      description: '插入代码块',

      icon: '</>',

      keywords: [
        '代码',
        '代码块',
        'code',
        'codeblock',
      ],

      command: ({ editor, range }) => {
        editor
          .chain()
          .focus()
          .deleteRange(range)
          .toggleCodeBlock()
          .run()
      },
    },

    /* ---------------------------------------------
       表格
    --------------------------------------------- */

    {
      id: 'table',

      title: '表格',

      description: '插入 3 × 3 表格',

      icon: '▦',

      keywords: [
        '表格',
        'table',
      ],

      command: ({ editor, range }) => {
        editor
          .chain()
          .focus()
          .deleteRange(range)
          .insertTable({
            rows: 3,
            cols: 3,
            withHeaderRow: true,
          })
          .run()
      },
    },

    /* ---------------------------------------------
       图片
    --------------------------------------------- */

    {
      id: 'image',

      title: '图片',

      description: '通过 URL 插入图片',

      icon: '🖼',

      keywords: [
        '图片',
        'image',
        'img',
      ],

      command: ({ editor, range }) => {
        const url = window.prompt(
          '请输入图片 URL',
        )

        if (!url) {
          return
        }

        editor
          .chain()
          .focus()
          .deleteRange(range)
          .setImage({
            src: url,
          })
          .run()
      },
    },
  ]
}

export default SlashCommand
