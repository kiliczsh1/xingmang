// useChatPanel.ts —— 核心组合函数
// 封装 @ai-sdk/vue 的 useChat，加上对话管理（创建/选择/删除对话）和插件链
import { ref, computed, watch, nextTick, shallowRef, type Ref } from 'vue'
import { useChat } from '@ai-sdk/vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { conversationAPI, configAPI } from '@/api'
import type { ChatMessage, ApiModel, RelatedContent } from '@/types'
import { CustomChatTransport } from '../bridge/ChatApiBridge'
import { useChatPlugins } from './useChatPlugins'
import { useRegexFilter } from './useRegexFilter'
import { useHistoryRecords } from './useHistoryRecords'
import type { PluginContext } from '../plugins/type'

export function useChatPanel(options: {
  bookId: number
}) {
  const { bookId } = options

  // ========= 模型配置 =========
  const selectedConfigId = ref<number | undefined>(undefined)
  const currentModelName = computed(() => {
    const config = apiConfigs.value.find(c => c.id === selectedConfigId.value)
    return config?.name || ''
  })
  const modelTemperature = ref(0.7)
  const modelTopP = ref(0.9)
  const apiConfigs = ref<ApiModel[]>([])

  // ========= 对话管理 =========
  const conversations = ref<any[]>([])
  const currentConversation = ref<any>(null)

  // ========= 插件链 =========
  const pluginChain = useChatPlugins()

  // ========= 正则过滤 =========
  const regexFilter = useRegexFilter('write-regex-settings-v1')

  // ========= 历史记录 =========
  const historyRecords = useHistoryRecords('write-history-records-v1')

  // ========= 关联状态 =========
  const worldBookLinked = ref(false)
  const selectedPrompts = ref<number[]>([])
  const relatedContent = ref<RelatedContent[]>([])
  const worldBookEntries = ref<any[]>([])
  const prompts = ref<any[]>([])

  // ========= 动态变量 =========
  /** 用户为变量输入的值 */
  const promptVariables = ref<Record<string, string>>({})
  /** 从当前选中 prompt 内容中解析出的变量名列表 */
  const parsedVariableNames = computed(() => {
    const selected = prompts.value.filter((p: any) => selectedPrompts.value.some(id => id == p.id))
    const vars = new Set<string>()
    const regex = /\{([^}]+)\}/g
    for (const p of selected) {
      let match
      while ((match = regex.exec(p.content || '')) !== null) {
        vars.add(match[1])
      }
    }
    return Array.from(vars)
  })

  /** 将 {变量名} 替换为用户输入的值 */
  function applyVariables(text: string, values: Record<string, string>): string {
    return text.replace(/\{([^}]+)\}/g, (_, name) => {
      const val = values[name]
      return val !== undefined && val !== '' ? val : `{${name}}`
    })
  }

  // ========= 输入 =========
  const userInput = ref('')

  // ========= 自定义 Transport =========
  // 注意：transport 在模块初始化时就创建并传递给 useChat（不用 getter！）
  // 之后 fetchConfigs 加载完后通过 updateConfig 更新内部配置
  const transport = new CustomChatTransport({
    configId: undefined,
    temperature: 0.7,
    topP: 0.9,
  })

  // ========= @ai-sdk/vue useChat =========
  // 直接传 transport 实例，不用 getter，避免 getter 求值时序问题
  const chat = useChat({
    transport,
    onFinish: async (message) => {
      // 保存消息到后端
      if (currentConversation.value) {
        const actualContent = regexFilter.applyActualPipeline(
          message.message.parts?.filter((p: any) => p.type === 'text').map((p: any) => p.text).join('') || ''
        )
        const displayContent = regexFilter.computeAssistantDisplay(actualContent)
        await conversationAPI.saveMessage(currentConversation.value.id, {
          role: 'assistant',
          content: actualContent,
          displayContent,
        })
        await fetchConversations()
      }

      // 保存历史记录
      const promptNames = prompts.value
        .filter((p: any) => selectedPrompts.value.some(id => id == p.id))
        .map((p: any) => p.name).join('、')
      const systemPrompts = prompts.value
        .filter((p: any) => selectedPrompts.value.some(id => id == p.id))
        .map((p: any) => p.content)
      const relatedContentSummary = relatedContent.value.length > 0
        ? relatedContent.value.map(item => `${item.type === 'chapter' ? '章节' : '备忘录'}：${item.title}`).join('、')
        : ''

      historyRecords.saveRecord({
        title: `正文对话 - ${currentConversation.value?.title || '新对话'}`,
        source: 'chat',
        sourceLabel: '正文对话',
        promptName: promptNames,
        promptCount: systemPrompts.length,
        status: 'completed',
        previewContent: (message.message.parts?.filter((p: any) => p.type === 'text').map((p: any) => p.text).join('') || '').substring(0, 140),
        messages: historyRecords.buildMessages({
          systemContents: relatedContentSummary ? [relatedContentSummary] : [],
          promptContents: systemPrompts,
          userContents: chat.messages.value
            .filter((m: any) => m.role === 'user')
            .map((m: any) => m.parts?.filter((p: any) => p.type === 'text').map((p: any) => p.text).join('') || ''),
          assistantContents: [message.message.parts?.filter((p: any) => p.type === 'text').map((p: any) => p.text).join('') || ''],
        }),
      })
    },
    onError: (error) => {
      const msg = error?.message || error?.toString() || '未知错误'
      console.error('[useChatPanel] AI 响应失败:', error)
      ElMessage.error('AI 响应失败: ' + msg)
    },
  })

  // ========= 加载状态 =========
  // v7 状态值: submitted（已发送等待响应） → streaming（流式中） → ready（完成） / error（错误）
  const isLoading = computed(() =>
    chat.status.value === 'submitted'
    || chat.status.value === 'streaming'
  )

  // ========= 发送消息（包装插件链） =========
  const sendMessage = async () => {
    if (!currentConversation.value) {
      ElMessage.warning('请先选择或创建对话')
      return
    }

    if (!transport.configId) {
      ElMessage.warning('AI 模型尚未加载，请先在API配置中添加模型')
      return
    }

    const userContent = userInput.value.trim()
    if (!userContent && selectedPrompts.value.length === 0 && relatedContent.value.length === 0) {
      ElMessage.warning('请输入消息或选择提示词/关联内容')
      return
    }

    // 构建插件上下文（含变量替换）
    const selectedPromptRecords = prompts.value
      .filter((p: any) => selectedPrompts.value.some(id => id == p.id))
      .map((p: any) => ({
        ...p,
        // 将 prompt 内容中的 {变量名} 替换为用户输入的值
        content: applyVariables(p.content || '', promptVariables.value),
      }))
    const context: PluginContext = {
      input: userContent,
      relatedContent: relatedContent.value,
      selectedPrompts: selectedPromptRecords,
      worldBookLinked: worldBookLinked.value,
      worldBookEntries: worldBookEntries.value,
    }

    // 插件链处理
    const processedContent = pluginChain.process(context)

    if (!processedContent.trim()) {
      ElMessage.warning('请输入消息或选择提示词/关联内容')
      return
    }

    // 保存用户消息到后端
    const userActualInput = userContent || '[已注入提示词/关联内容]'
    await conversationAPI.saveMessage(currentConversation.value.id, {
      role: 'user',
      content: processedContent,
      displayContent: userActualInput,
    })

    // 清空输入
    userInput.value = ''

    // 手动将用户消息添加到本地 chatMessages
    chatMessages.value.push({
      role: 'user',
      content: processedContent,
      displayContent: userActualInput,
    })

    // 使用 @ai-sdk/vue 的 sendMessage
    // 注意：v7 必须用 text 字段（而非 content），text 会自动构建正确的 parts 结构
    await chat.sendMessage({
      text: processedContent,
    })
  }

  // ========= 停止生成 =========
  const stopGeneration = () => {
    chat.stop()
    ElMessage.info('已停止生成')
  }

  // ========= 重新生成 =========
  const regenerateMessage = async (index: number) => {
    // 获取要重新生成的消息
    const msg = chatMessages.value[index]
    if (!msg) return

    // 如果有后端 ID，删除后端记录
    if (msg.id && currentConversation.value) {
      try {
        await conversationAPI.deleteMessage(currentConversation.value.id, msg.id)
      } catch {
        // 忽略
      }
    }

    // 从本地列表中移除
    chatMessages.value.splice(index, 1)

    // 利用 useChat 重新生成（需要传前一条用户消息的内容）
    const prevUserMsg = [...chatMessages.value].reverse().find(m => m.role === 'user')
    if (prevUserMsg) {
      await chat.sendMessage({ text: prevUserMsg.content })
    }
  }

  // ========= 对话 CRUD =========
  const fetchConversations = async () => {
    try {
      const res = await conversationAPI.getByBook(bookId)
      if (res.success && res.data) {
        conversations.value = res.data
        if (res.data.length > 0 && !currentConversation.value) {
          await selectConversation(res.data[0])
        } else if (res.data.length === 0) {
          await createConversation()
        }
      }
    } catch (error) {
      console.error('加载对话列表失败:', error)
    }
  }

  const createConversation = async () => {
    try {
      const res = await conversationAPI.create({
        book_id: bookId,
        title: `新对话 ${new Date().toLocaleTimeString()}`,
      })
      if (res.success && res.data) {
        await fetchConversations()
        await selectConversation(res.data)
      }
    } catch (error) {
      console.error('创建对话失败:', error)
    }
  }

  const selectConversation = async (conv: any) => {
    currentConversation.value = conv
    try {
      const res = await conversationAPI.getMessages(conv.id)
      if (res.success && res.data) {
        // 将后端消息转换为 useChat 的 UIMessage 格式
        // 但由于 useChat 管理消息列表，我们这里只做数据加载
        // 实际显示时需要映射
        chatMessages.value = (res.data as ChatMessage[]).map(msg => {
          if (!msg.displayContent && (msg as any).display_content) {
            msg.displayContent = (msg as any).display_content
          }
          if (msg.role === 'user' && !msg.displayContent) {
            msg.displayContent = extractUserDisplayContent(msg.content)
          }
          if (msg.role === 'assistant' && !msg.displayContent) {
            msg.displayContent = regexFilter.computeAssistantDisplay(msg.content)
          }
          return msg
        })
      }
    } catch (error) {
      console.error('加载消息失败:', error)
    }
  }

  const deleteConversation = async (conv: any) => {
    try {
      await ElMessageBox.confirm(`确定删除对话"${conv.title}"吗？`, '提示', {
        type: 'warning',
      })
      const res = await conversationAPI.delete(conv.id)
      if (res.success) {
        if (currentConversation.value?.id === conv.id) {
          currentConversation.value = null
          chatMessages.value = []
        }
        await fetchConversations()
        ElMessage.success('删除成功')
      }
    } catch {
      // 用户取消
    }
  }

  const renameConversation = async (conv: any) => {
    try {
      const { value } = await ElMessageBox.prompt('请输入新的对话名称', '重命名对话', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputValue: conv.title || '新对话',
        inputPattern: /.+/,
        inputErrorMessage: '对话名称不能为空',
      })
      if (value) {
        const res = await conversationAPI.update(conv.id, { title: value })
        if (res.success) {
          ElMessage.success('重命名成功')
          const target = conversations.value.find((c: any) => c.id === conv.id)
          if (target) target.title = value
          if (currentConversation.value?.id === conv.id) {
            currentConversation.value.title = value
          }
        }
      }
    } catch {
      // 用户取消
    }
  }

  // ========= 删除消息 =========
  const deleteMessage = async (index: number) => {
    try {
      await ElMessageBox.confirm('确定删除这条消息吗？', '提示', { type: 'warning' })
      const msg = chatMessages.value[index]
      if (!msg) return
      if (msg.id && currentConversation.value) {
        const res = await conversationAPI.deleteMessage(currentConversation.value.id, msg.id)
        if (!res.success) {
          ElMessage.error('删除失败')
          return
        }
      }
      chatMessages.value.splice(index, 1)
      ElMessage.success('删除成功')
    } catch {
      // 用户取消
    }
  }

  // ========= 消息列表（本地维护，兼容旧格式） =========
  // 由于 useChat 的 messages 格式 (UIMessage) 与后端 ChatMessage 不同，
  // 我们仍使用本地 chatMessages 来维护消息，useChat 仅负责流式通信
  const chatMessages = ref<ChatMessage[]>([])
  const chatMessagesRef = ref<HTMLElement>()

  const scrollToBottom = () => {
    if (chatMessagesRef.value) {
      nextTick(() => {
        chatMessagesRef.value!.scrollTop = chatMessagesRef.value!.scrollHeight
      })
    }
  }

  // 监听 useChat 的消息变化，同步到本地 chatMessages
  // 注意：v7 流式响应以 text-delta 逐步到达，assistant 消息内容会逐步增长。
  // 不能用找「空内容」的方式判断（因首次 delta 后就有了内容），而应看最后一条是否为 assistant。
  watch(() => chat.messages.value, (newMessages) => {
    if (newMessages.length === 0) return

    const lastMessage = newMessages[newMessages.length - 1]
    if (lastMessage.role === 'assistant') {
      const text = lastMessage.parts?.filter((p: any) => p.type === 'text').map((p: any) => p.text).join('') || ''
      const lastIsAssistant = chatMessages.value.length > 0
        && chatMessages.value[chatMessages.value.length - 1].role === 'assistant'

      if (lastIsAssistant) {
        // 正在流式中的同一条 assistant 消息，原地更新内容
        const idx = chatMessages.value.length - 1
        chatMessages.value[idx].content = text
        chatMessages.value[idx].displayContent = regexFilter.computeAssistantDisplay(text)
      } else if (text) {
        // 上一条不是 assistant（或列表为空），新增一条
        chatMessages.value.push({
          role: 'assistant',
          content: text,
          displayContent: regexFilter.computeAssistantDisplay(text),
        })
      }
      scrollToBottom()
    }
  }, { deep: true })

  // ========= 提取用户显示内容 =========
  const extractUserDisplayContent = (fullContent: string): string => {
    const labelRE = /^(提示词：|世界书：|关联内容：)[\s\S]*?\n\n(?=提示词：|世界书：|关联内容：|\S|$)/
    let result = fullContent
    let prev = ''
    while (result !== prev) {
      prev = result
      result = result.replace(labelRE, '')
    }
    return result.trim()
  }

  // ========= 获取显示内容 =========
  const getAssistantDisplayContent = (message: ChatMessage) => {
    return message.displayContent || regexFilter.computeAssistantDisplay(message.content || '')
  }

  const getMessageCopyContent = (message: ChatMessage) => {
    if (message.role === 'assistant' && regexFilter.copyUsesFiltered.value) {
      return getAssistantDisplayContent(message)
    }
    return message.displayContent || extractUserDisplayContent(message.content) || message.content
  }

  // ========= 应用到编辑器 =========
  const applyToCursor = ref<((content: string) => void) | null>(null)

  // ========= 初始化 =========
  const fetchConfigs = async () => {
    try {
      const res = await configAPI.getAll()
      if (res.success && res.data) {
        apiConfigs.value = res.data.filter((m: ApiModel) => m.enabled !== 0)
        if (res.data.length > 0 && !selectedConfigId.value) {
          selectedConfigId.value = res.data[0].id
        }
      }
    } catch (error) {
      console.error('加载API配置失败:', error)
    }
  }

  // 监听 selectedConfigId 变化，更新 transport 配置
  watch(selectedConfigId, (val) => {
    if (val) {
      transport.updateConfig({ configId: val })
      console.log('[useChatPanel] Transport 已更新 configId:', val)
    }
  })

  // 监听配置参数变化，更新 transport
  watch([selectedConfigId, modelTemperature, modelTopP], () => {
    if (selectedConfigId.value) {
      transport.updateConfig({
        configId: selectedConfigId.value,
        temperature: modelTemperature.value,
        topP: modelTopP.value,
      })
    }
  })

  // 所选 prompt 变化时，清除不再存在的变量值
  watch(parsedVariableNames, (newVars) => {
    const oldKeys = Object.keys(promptVariables.value)
    for (const key of oldKeys) {
      if (!newVars.includes(key)) {
        delete promptVariables.value[key]
      }
    }
  })

  // 初始化加载
  const init = async () => {
    await fetchConfigs()
    await fetchConversations()
  }

  return {
    // useChat 暴露
    chatStatus: chat.status,
    chatError: chat.error,
    isLoading,

    // 对话管理
    conversations,
    currentConversation,
    fetchConversations,
    createConversation,
    selectConversation,
    deleteConversation,
    renameConversation,

    // 消息
    chatMessages,
    chatMessagesRef,
    sendMessage,
    stopGeneration,
    regenerateMessage,
    deleteMessage,
    scrollToBottom,
    getAssistantDisplayContent,
    getMessageCopyContent,

    // 输入
    userInput,

    // 模型配置
    selectedConfigId,
    currentModelName,
    modelTemperature,
    modelTopP,
    apiConfigs,
    fetchConfigs,

    // 插件
    pluginChain,

    // 正则过滤
    regexFilter,

    // 历史记录
    historyRecords,

    // 关联状态
    worldBookLinked,
    selectedPrompts,
    relatedContent,
    worldBookEntries,
    prompts,

    // 动态变量
    promptVariables,
    parsedVariableNames,
    applyVariables,

    // 应用到编辑器
    applyToCursor,

    // 初始化
    init,

    // 工具
    extractUserDisplayContent,
  }
}