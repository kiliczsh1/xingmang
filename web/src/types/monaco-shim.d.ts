/**
 * monaco-editor 0.57 的 exports 映射把 "./*" 指向 esm/vs/*.js，
 * 因此运行时要走 `monaco-editor/editor/editor.api.js` 这类短路径，但包里的
 * 类型声明只挂在包根上，这里补一个环境模块声明把类型接回去。
 *
 * 注意：我们**不用** editor.main.js（它会把 css/html/typescript 语言服务与
 * 81 个基础语言高亮一并拉进产物，worker 因此多出约 8.4 MB）。
 * 精简入口见 src/views/workflow/monaco-lite.ts。
 */
declare module 'monaco-editor/editor/editor.api.js' {
  import * as monaco from 'monaco-editor'
  export = monaco
}

/**
 * monaco-lite.ts 里按副作用导入的 contrib / 基础模块在包内没有 .d.ts，
 * 而 tsconfig 开了 noUncheckedSideEffectImports，必须显式声明这些通配模块，
 * 否则每一行 import 都会报「找不到模块」。
 * 这些声明只解除类型报错，不产生任何运行时代码。
 */
declare module 'monaco-editor/contrib/*'
declare module 'monaco-editor/browser/*'
declare module 'monaco-editor/features/*'
declare module 'monaco-editor/base/*'
declare module 'monaco-editor/common/*'
declare module 'monaco-editor/languages/*'

/** Monaco 的 worker 工厂挂在全局 MonacoEnvironment 上 */
declare interface Window {
  MonacoEnvironment?: {
    getWorker?(workerId: string, label: string): Worker
    getWorkerUrl?(workerId: string, label: string): string
  }
}
