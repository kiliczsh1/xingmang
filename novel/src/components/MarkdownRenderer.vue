<template>
  <div class="markdown-body" v-html="renderedContent"></div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'

const props = defineProps<{
  content: string
  renderMode?: 'markdown' | 'html'
}>()

const createMarkdownIt = (allowHtml: boolean) => new MarkdownIt({
 html: allowHtml,
 breaks: true,
 linkify: true,
 typographer: true,
 highlight: function (str, lang) {
 if (lang && hljs.getLanguage(lang)) {
 try {
 return hljs.highlight(str, { language: lang }).value
 } catch (__) {}
 }
 return ''
 }
 })

 const markdownRenderer = createMarkdownIt(true)

 const renderedContent = computed(() => {
 if (props.renderMode === 'html') {
 return props.content || ''
 }
 return markdownRenderer.render(props.content || '')
 })
</script>

<style>
.markdown-body {
  min-width: 0;
  max-width: 100%;
  font-size: 15px;
  line-height: 1.7;
  word-wrap: break-word;
  overflow-wrap: anywhere;
  box-sizing: border-box;
}

.markdown-body h1,
.markdown-body h2,
.markdown-body h3,
.markdown-body h4,
.markdown-body h5,
.markdown-body h6 {
  margin-top: 24px;
  margin-bottom: 16px;
  font-weight: 600;
  line-height: 1.25;
}

.markdown-body h1 {
  font-size: 2em;
  border-bottom: 1px solid #eaecef;
  padding-bottom: 0.3em;
}

.markdown-body h2 {
  font-size: 1.5em;
  border-bottom: 1px solid #eaecef;
  padding-bottom: 0.3em;
}

.markdown-body h3 {
  font-size: 1.25em;
}

.markdown-body p {
  margin-top: 0;
  margin-bottom: 16px;
}

.markdown-body img {
  display: block;
  max-width: 100%;
  width: auto;
  height: auto;
  margin: 18px 0;
  border-radius: 14px;
  object-fit: contain;
  box-sizing: border-box;
}

.markdown-body code {
  padding: 0.2em 0.4em;
  margin: 0;
  font-size: 85%;
  background-color: rgba(27, 31, 35, 0.05);
  border-radius: 6px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.markdown-body pre {
  padding: 16px;
  overflow: auto;
  font-size: 85%;
  line-height: 1.45;
  background-color: #f6f8fa;
  border-radius: 6px;
  margin-bottom: 16px;
}

.markdown-body pre code {
  display: inline;
  padding: 0;
  margin: 0;
  overflow: visible;
  line-height: inherit;
  word-wrap: normal;
  background-color: transparent;
  border: 0;
}

.markdown-body ul,
.markdown-body ol {
  padding-left: 2em;
  margin-bottom: 16px;
}

.markdown-body li {
  margin-top: 0.25em;
}

.markdown-body blockquote {
  margin: 0;
  padding: 0 1em;
  color: #6a737d;
  border-left: 0.25em solid #dfe2e5;
  margin-bottom: 16px;
}

.markdown-body table {
  display: block;
  max-width: 100%;
  overflow-x: auto;
  border-spacing: 0;
  border-collapse: collapse;
  margin-bottom: 16px;
}

.markdown-body table th,
.markdown-body table td {
  padding: 6px 13px;
  border: 1px solid #dfe2e5;
}

.markdown-body table th {
  font-weight: 600;
  background-color: #f6f8fa;
}

.markdown-body a {
  color: #0366d6;
  text-decoration: none;
}

.markdown-body a:hover {
  text-decoration: underline;
}

/* 暗色主题适配 */
:root[data-theme='dark'] .markdown-body {
  color: #e5e7eb !important;
  background-color: #1e293b !important;
}

:root[data-theme='dark'] .markdown-body .hljs {
  background: #0f172a !important;
  color: #e5e7eb !important;
}

:root[data-theme='dark'] .markdown-body h1,
:root[data-theme='dark'] .markdown-body h2,
:root[data-theme='dark'] .markdown-body h3,
:root[data-theme='dark'] .markdown-body h4,
:root[data-theme='dark'] .markdown-body h5,
:root[data-theme='dark'] .markdown-body h6 {
  color: #f3f4f6 !important;
  border-bottom-color: rgba(71, 85, 105, 0.4) !important;
}

:root[data-theme='dark'] .markdown-body img {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3) !important;
}

:root[data-theme='dark'] .markdown-body code {
  background-color: rgba(51, 65, 85, 0.6) !important;
  color: #f3f4f6 !important;
}

:root[data-theme='dark'] .markdown-body pre {
  background-color: rgba(30, 41, 59, 0.95) !important;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3) !important;
  border: 1px solid rgba(71, 85, 105, 0.3) !important;
}

:root[data-theme='dark'] .markdown-body pre code {
  color: #e5e7eb !important;
}

:root[data-theme='dark'] .markdown-body blockquote {
  color: #9ca3af !important;
  border-left-color: rgba(94, 234, 212, 0.3) !important;
}

:root[data-theme='dark'] .markdown-body table th,
:root[data-theme='dark'] .markdown-body table td {
  border-color: rgba(71, 85, 105, 0.4) !important;
}

:root[data-theme='dark'] .markdown-body table th {
  background-color: rgba(51, 65, 85, 0.6) !important;
  color: #f3f4f6 !important;
}

:root[data-theme='dark'] .markdown-body table td {
  color: #e5e7eb !important;
}

:root[data-theme='dark'] .markdown-body a {
  color: #5eead4 !important;
}

:root[data-theme='dark'] .markdown-body a:hover {
  color: #2dd4bf !important;
}

:root[data-theme='dark'] .markdown-body ul,
:root[data-theme='dark'] .markdown-body ol {
  color: #e5e7eb !important;
}

:root[data-theme='dark'] .markdown-body li {
  color: #e5e7eb !important;
}

:root[data-theme='dark'] .markdown-body p {
  color: #e5e7eb !important;
}

/* 代码高亮暗色主题 */
:root[data-theme='dark'] .markdown-body .hljs-comment,
:root[data-theme='dark'] .markdown-body .hljs-quote {
  color: #8b949e !important;
}

:root[data-theme='dark'] .markdown-body .hljs-keyword,
:root[data-theme='dark'] .markdown-body .hljs-selector-tag,
:root[data-theme='dark'] .markdown-body .hljs-addition {
  color: #ff7b72 !important;
}

:root[data-theme='dark'] .markdown-body .hljs-string,
:root[data-theme='dark'] .markdown-body .hljs-meta .hljs-string,
:root[data-theme='dark'] .markdown-body .hljs-regexp,
:root[data-theme='dark'] .markdown-body .hljs-attribute {
  color: #a5d6ff !important;
}

:root[data-theme='dark'] .markdown-body .hljs-number,
:root[data-theme='dark'] .markdown-body .hljs-literal,
:root[data-theme='dark'] .markdown-body .hljs-variable,
:root[data-theme='dark'] .markdown-body .hljs-template-variable,
:root[data-theme='dark'] .markdown-body .hljs-tag .hljs-attr {
  color: #79c0ff !important;
}

:root[data-theme='dark'] .markdown-body .hljs-title,
:root[data-theme='dark'] .markdown-body .hljs-section,
:root[data-theme='dark'] .markdown-body .hljs-selector-id {
  color: #d2a8ff !important;
}

:root[data-theme='dark'] .markdown-body .hljs-type,
:root[data-theme='dark'] .markdown-body .hljs-class .hljs-title,
:root[data-theme='dark'] .markdown-body .hljs-built_in {
  color: #ffa657 !important;
}

:root[data-theme='dark'] .markdown-body .hljs-symbol,
:root[data-theme='dark'] .markdown-body .hljs-bullet,
:root[data-theme='dark'] .markdown-body .hljs-subst,
:root[data-theme='dark'] .markdown-body .hljs-meta,
:root[data-theme='dark'] .markdown-body .hljs-meta .hljs-keyword {
  color: #8b949e !important;
}

:root[data-theme='dark'] .markdown-body .hljs-deletion {
  color: #ffa198 !important;
}

:root[data-theme='dark'] .markdown-body .hljs-addition {
  color: #7ee787 !important;
}

:root[data-theme='dark'] .markdown-body .hljs-emphasis {
  font-style: italic !important;
}

:root[data-theme='dark'] .markdown-body .hljs-strong {
  font-weight: bold !important;
}
</style>

