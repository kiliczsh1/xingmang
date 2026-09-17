// 本地插件入口代码契约
// 通过 window.__installPlugin 暴露 ChatPlugin：
//   { name, description?, enabled?, process(content, context) => string }
// 本示例：把 AI 回复末尾追加一行标记
window.__installPlugin = {
  name: 'example-suffix',
  description: '把 AI 回复末尾追加标记（示例插件）',
  enabled: true,
  process: function (content, context) {
    return content.trimEnd() + '\n\n——由示例插件生成'
  }
}
