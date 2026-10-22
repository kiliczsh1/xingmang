<template>
  <transition name="slide-fade">
    <div v-if="visible" class="wf-drawer" @keydown.stop>
      <div class="wf-drawer-header">
        <div class="wf-drawer-title"><i class="fas fa-sliders-h" /><span>节点属性</span></div>
        <el-button text @click="$emit('close')"><i class="fas fa-times" /></el-button>
      </div>
      <div class="wf-drawer-body">
        <div v-if="!node" class="empty-state">
          <i class="fas fa-hand-pointer" />
          <p>在画布上点一个节点</p>
        </div>

        <el-form v-else :label-position="'top'" :model="form" class="property-form" @submit.prevent>
          <el-form-item label="节点名称">
            <el-input v-model="form.label" placeholder="请输入节点名称" @change="commit" />
          </el-form-item>

          <!-- ================= AI 处理 ================= -->
          <template v-if="node.type === 'aiProcess'">
            <el-alert type="info" :closable="false" class="io-alert">
              <template #title>
                <div class="io-line"><span>上下文输入：</span><code>{{ ctxNames.join(', ') || '（无）' }}</code></div>
                <div class="io-line"><span>输出：</span><code>aiResult / lastAIOutput{{ extraOutVar ? ' / ' + extraOutVar : '' }}</code></div>
              </template>
            </el-alert>

            <el-divider content-position="left">AI 配置</el-divider>
            <el-form-item label=" AI 模型 ">
              <ModelSelect :model-value="cfg.configId" @picked="onModelPicked" />
            </el-form-item>
            <el-form-item label=" 用户提示词 ">
              <el-input
                v-model="cfg.userPrompt"
                type="textarea"
                :rows="5"
                placeholder="请输入提示词，使用 {{var}} 引用上下文变量"
                @change="commit"
              />
            </el-form-item>
            <el-form-item label="上下文变量">
              <div class="var-rows">
                <div v-for="(name, i) in ctxNames" :key="i" class="variable-row">
                  <el-select
                    :model-value="name"
                    filterable
                    allow-create
                    default-first-option
                    size="small"
                    placeholder="选择节点和变量"
                    class="var-select"
                    @update:model-value="(v: any) => setCtxVar(i, v)"
                  >
                    <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
                  </el-select>
                  <el-button text size="small" @click="removeCtxVar(i)"><i class="fas fa-times" /></el-button>
                </div>
                <el-button size="small" class="add-btn" @click="addCtxVar">
                  <i class="fas fa-plus" /> 添加上下文变量
                </el-button>
              </div>
              <div class="form-tip">从上游节点选择输出变量作为本节点的上下文输入</div>
            </el-form-item>
            <el-form-item label="每个数组变量取最近 N 条">
              <el-input-number v-model="cfg.contextMaxItems" :min="1" :max="50" @change="commit" />
            </el-form-item>
            <el-form-item label="Temperature">
              <el-input-number v-model="cfg.temperature" :min="0" :max="2" :step="0.1" @change="commit" />
            </el-form-item>
            <el-form-item label="Max Tokens">
              <el-input-number v-model="cfg.maxTokens" :min="0" :max="200000" :step="500" @change="commit" />
              <div class="form-tip">设置为 0 表示不限制，使用模型的最大输出长度</div>
            </el-form-item>
            <el-form-item label="输出变量">
              <el-input v-model="cfg.outputVariable" placeholder="留空则输出 aiResult / lastAIOutput" @change="commit" />
            </el-form-item>
            <el-form-item label="系统提示词（可直接填写，或用提示词库 id）">
              <el-input
                v-model="cfg.systemPrompt"
                type="textarea"
                :rows="3"
                placeholder="可选；也可在下方填提示词库 id"
                @change="commit"
              />
            </el-form-item>
            <el-form-item label="提示词库 id">
              <el-input-number v-model="cfg.systemPromptId" :min="0" :controls="false" placeholder="从提示词库选择" @change="commit" />
            </el-form-item>
          </template>

          <!-- ================= 开始 ================= -->
          <template v-else-if="node.type === 'start'">
            <el-divider content-position="left">初始变量</el-divider>
            <el-alert type="info" :closable="false" class="io-alert">
              <template #title>输出下方配置的所有初始变量，变量名和值由用户自定义</template>
            </el-alert>
            <div class="var-rows">
              <div v-for="(row, i) in varRows" :key="i" class="variable-row two">
                <el-input v-model="row.k" size="small" placeholder="变量名" @change="commitVars" />
                <el-input v-model="row.v" size="small" placeholder="变量值" @change="commitVars" />
                <el-button text size="small" @click="removeVar(i)"><i class="fas fa-times" /></el-button>
              </div>
              <el-button size="small" class="add-btn" @click="addVar"><i class="fas fa-plus" /> 添加变量</el-button>
            </div>
          </template>

          <!-- ================= 文本 ================= -->
          <template v-else-if="node.type === 'text'">
            <el-divider content-position="left">文本内容</el-divider>
            <el-form-item label="文本内容">
              <el-input
                v-model="cfg.text"
                type="textarea"
                :rows="6"
                placeholder="输入文本内容，使用 {{var}} 引用上下文变量"
                @change="commit"
              />
            </el-form-item>
            <el-form-item label="输出变量名">
              <el-input v-model="cfg.outputVariable" placeholder="默认：textResult（可选）" @change="commit" />
              <div class="form-tip">留空则使用默认变量名 textResult</div>
            </el-form-item>
          </template>

          <!-- ================= 合并 ================= -->
          <template v-else-if="node.type === 'merge'">
            <el-divider content-position="left">合并设置</el-divider>
            <el-form-item label="输入变量">
              <el-select
                :model-value="mergeNames"
                multiple
                filterable
                allow-create
                default-first-option
                placeholder="选择要合并的变量（必填）"
                style="width: 100%"
                @update:model-value="(v: any) => (cfg.inputVariables = v)"
                @change="commit"
              >
                <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
              </el-select>
            </el-form-item>
            <el-form-item label="合并类型">
              <el-select v-model="cfg.mergeType" placeholder="选择合并类型" style="width: 100%" @change="commit">
                <el-option label="拼接合并" value="concat" />
                <el-option label="数组合并" value="array" />
                <el-option label="对象合并" value="object" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="(cfg.mergeType || 'concat') === 'concat'" label="分隔符">
              <el-input v-model="cfg.separator" placeholder="例如：\n\n" @change="commit" />
            </el-form-item>
          </template>

          <!-- ================= 循环 ================= -->
          <template v-else-if="node.type === 'loop'">
            <el-divider content-position="left">循环设置</el-divider>
            <el-form-item label="循环类型" required>
              <el-select v-model="cfg.loopType" placeholder="选择循环类型" style="width: 100%" @change="commit">
                <el-option label="📋 数组循环" value="array" />
                <el-option label="🔢 固定循环" value="fixed" />
                <el-option label="❓ 条件循环" value="condition" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="cfg.loopType === 'array'" label="数组源" required>
              <el-select
                :model-value="cfg.arraySource?.variableName"
                filterable
                allow-create
                default-first-option
                clearable
                placeholder="选择数组变量"
                style="width: 100%"
                @update:model-value="(v: any) => (cfg.arraySource = v ? { variableName: v } : null)"
                @change="commit"
              >
                <el-option v-for="n in arrayVariables" :key="n" :label="n" :value="n" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="cfg.loopType === 'fixed'" label="固定循环次数">
              <el-input-number v-model="cfg.fixedCount" :min="1" :max="9999" @change="commit" />
            </el-form-item>
            <el-form-item v-if="cfg.loopType === 'condition'" label="最大循环次数">
              <el-input-number v-model="cfg.maxIterations" :min="1" :max="9999" @change="commit" />
            </el-form-item>
            <el-form-item label="保持上下文">
              <el-switch v-model="cfg.keepContext" @change="commit" />
            </el-form-item>
            <el-form-item label="历史窗口（注入最近 N 章）">
              <el-input-number v-model="cfg.historyWindow" :min="0" :max="50" @change="commit" />
            </el-form-item>
          </template>

          <!-- ================= 上下文 ================= -->
          <template v-else-if="node.type === 'context'">
            <el-form-item label="操作类型">
              <el-select v-model="cfg.action" style="width: 100%" @change="commit">
                <el-option label="保存变量（save）" value="save" />
                <el-option label="追加历史（append）" value="append" />
                <el-option label="解析 JSON（parseJson）" value="parseJson" />
                <el-option label="选择数组元素（pick）" value="pick" />
                <el-option label="读取变量（load）" value="load" />
                <el-option label="清除变量（clear）" value="clear" />
                <el-option label="保存章节文件（saveChapter）" value="saveChapter" />
              </el-select>
            </el-form-item>
            <el-form-item label="变量名" required>
              <el-input v-model="cfg.variableName" placeholder="写入 / 读取的变量名" @change="commit" />
            </el-form-item>
            <template v-if="['save', 'append', 'parseJson'].includes(cfg.action || 'save')">
              <el-form-item label="源变量">
                <el-select
                  :model-value="ctxInputNames"
                  multiple
                  filterable
                  allow-create
                  default-first-option
                  placeholder="留空则取 lastOutput"
                  style="width: 100%"
                  @update:model-value="(v: any) => (cfg.inputVariables = toRefs(v))"
                  @change="commit"
                >
                  <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
                </el-select>
              </el-form-item>
            </template>
            <el-form-item v-if="cfg.action === 'append'" label="最多保留条数">
              <el-input-number v-model="cfg.maxHistory" :min="1" :max="500" @change="commit" />
            </el-form-item>
            <el-form-item v-if="cfg.action === 'append'" label="自动加章节标题">
              <el-switch v-model="cfg.includeTitle" @change="commit" />
            </el-form-item>
            <template v-if="cfg.action === 'pick'">
              <el-form-item label="数组变量">
                <el-input v-model="cfg.arrayVariable" placeholder="chapterPlans" @change="commit" />
              </el-form-item>
              <el-form-item label="索引变量">
                <el-input v-model="cfg.indexVariable" placeholder="currentIteration" @change="commit" />
              </el-form-item>
            </template>
            <template v-if="cfg.action === 'saveChapter'">
              <el-form-item label="章节序号变量">
                <el-input v-model="cfg.chapterIndex" placeholder="currentIteration" @change="commit" />
              </el-form-item>
              <el-form-item label="章节计划变量">
                <el-input v-model="cfg.chapterPlans" placeholder="chapterPlans" @change="commit" />
              </el-form-item>
              <el-form-item label="正文变量">
                <el-input v-model="cfg.content" placeholder="aiResult" @change="commit" />
              </el-form-item>
            </template>
          </template>

          <!-- ================= 结束 ================= -->
          <template v-else-if="node.type === 'end'">
            <el-divider content-position="left">输出设置</el-divider>
            <el-form-item label="输出目录">
              <el-input v-model="cfg.outputDir" placeholder="留空则使用项目根目录" @change="commit" />
              <div class="form-tip">星芒固定落盘在数据目录的 workflow_outputs/ 内，此处仅作记录</div>
            </el-form-item>
            <el-form-item label="输出文件名">
              <el-input v-model="cfg.outputFileName" placeholder="留空则自动生成" @change="commit" />
            </el-form-item>
            <el-form-item label="输出格式">
              <el-select v-model="cfg.outputFormat" placeholder="选择格式" style="width: 100%" @change="commit">
                <el-option label="Markdown (.md)" value="md" />
                <el-option label="纯文本 (.txt)" value="txt" />
                <el-option label="JSON (.json)" value="json" />
              </el-select>
            </el-form-item>
            <el-form-item label="输出变量">
              <el-select
                :model-value="includeNames"
                multiple
                filterable
                allow-create
                default-first-option
                placeholder="留空则自动选择"
                style="width: 100%"
                @update:model-value="(v: any) => (cfg.includeOutputs = v)"
                @change="commit"
              >
                <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
              </el-select>
            </el-form-item>
            <el-form-item label="数组分隔符">
              <el-input v-model="cfg.separator" placeholder="默认：\n\n" @change="commit" />
            </el-form-item>
          </template>

          <!-- ================= 星域·注释 ================= -->
          <template v-else-if="node.type === 'comment'">
            <el-divider content-position="left">注释</el-divider>
            <el-form-item label="注释内容">
              <el-input v-model="cfg.text" type="textarea" :rows="4" @change="commit" />
            </el-form-item>
            <el-form-item label="字体大小">
              <div class="size-row">
                <el-input-number v-model="cfg.fontSize" :min="10" :max="32" @change="commit" />
                <span class="form-tip">当前 {{ cfg.fontSize || 14 }}px</span>
              </div>
            </el-form-item>
            <el-form-item label="文字颜色">
              <el-color-picker v-model="cfg.textColor" @change="commit" />
            </el-form-item>
            <el-form-item label="背景颜色">
              <el-color-picker v-model="cfg.backgroundColor" @change="commit" />
            </el-form-item>
            <div class="form-tip">注释节点不参与执行；画布上可拖右下角调整尺寸</div>
          </template>

          <!-- ================= 星域·中转 ================= -->
          <template v-else-if="node.type === 'transfer'">
            <el-divider content-position="left">中转设置</el-divider>
            <el-form-item label="占位符">
              <el-input v-model="cfg.placeholder" placeholder="未设占位符" @change="commit" />
              <div class="form-tip" v-pre>下游可用同名占位符引用中转的数据</div>
            </el-form-item>
            <el-form-item label="显示模式">
              <el-select v-model="cfg.displayMode" style="width: 100%" @change="commit">
                <el-option label="文本模式" value="text" />
                <el-option label="JSON 模式" value="json" />
              </el-select>
            </el-form-item>
            <el-form-item label="自动转发">
              <el-switch v-model="cfg.autoForward" @change="commit" />
              <div class="form-tip">关闭后数据不再继续向下游传递</div>
            </el-form-item>
            <el-form-item label="输入变量">
              <el-select
                v-model="cfg.inputVariable"
                filterable
                allow-create
                default-first-option
                style="width: 100%"
                @change="commit"
              >
                <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
              </el-select>
            </el-form-item>
            <el-form-item label="占位符与数据">
              <el-button size="small" @click="swapTransferData">
                <i class="fas fa-right-left" /> 将占位符与接收数据互换
              </el-button>
            </el-form-item>
            <el-form-item label="接收的数据">
              <el-input :model-value="String(cfg.receivedData || '')" type="textarea" :rows="4" readonly />
              <div class="form-tip">执行时写入，仅用于查看 · 当前 {{ String(cfg.receivedData || '').length }} 字符</div>
            </el-form-item>
          </template>

          <!-- ================= 星域·文件输出 ================= -->
          <template v-else-if="node.type === 'fileOutput'">
            <el-divider content-position="left">输出设置</el-divider>
            <el-form-item label="保存路径">
              <el-input v-model="cfg.path" placeholder="./exports/" @change="commit" />
              <div class="form-tip">相对本次执行的输出目录（星芒落盘限定在该目录内）</div>
            </el-form-item>
            <el-form-item label="文件名">
              <el-input v-model="cfg.filename" placeholder="output.txt" @change="commit" />
            </el-form-item>
            <el-form-item label="源变量">
              <el-select
                v-model="cfg.sourceVariable"
                filterable
                allow-create
                default-first-option
                style="width: 100%"
                @change="commit"
              >
                <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
              </el-select>
            </el-form-item>
          </template>

          <!-- ================= 星域·文本输入 ================= -->
          <template v-else-if="node.type === 'textInput'">
            <el-divider content-position="left">文本内容</el-divider>
            <el-form-item label="输入提示词">
              <el-input v-model="cfg.text" type="textarea" :rows="4" @change="commit" />
            </el-form-item>
            <el-form-item label="">
              <el-button size="small" @click="saveTextToPromptList"><i class="fas fa-plus" /> 保存到提示词列表</el-button>
              <el-button size="small" @click="importTextInto('text')"><i class="fas fa-file-import" /> 导入文件</el-button>
            </el-form-item>
            <el-form-item label="提示词列表">
              <div class="prompt-list-box">
                <el-checkbox-group v-model="cfg.selectedPrompts">
                  <el-checkbox
                    v-for="(item, idx) in (cfg.promptList || [])"
                    :key="idx"
                    :label="item"
                  >{{ shortText(item) }}</el-checkbox>
                </el-checkbox-group>
                <div v-if="!(cfg.promptList || []).length" class="form-tip">暂无提示词</div>
                <div class="prompt-list-actions">
                  <el-select
                    :model-value="[]"
                    multiple
                    filterable
                    placeholder="从提示词库添加"
                    style="flex: 1"
                    @change="addFromPromptLibrary"
                  >
                    <el-option v-for="p in promptOptions" :key="p.id" :label="p.title" :value="p.id" />
                  </el-select>
                  <el-button size="small" text @click="removeSelectedPrompts">删除选中</el-button>
                  <el-button size="small" text type="danger" @click="clearPromptList">清空</el-button>
                </div>
              </div>
            </el-form-item>
            <el-form-item label="占位符">
              <el-input v-model="cfg.placeholder" placeholder="未设占位符" @change="commit" />
              <div class="form-tip" v-pre>下游可用同名占位符引用这段文本</div>
            </el-form-item>
          </template>

          <!-- ================= 星域·提示词 ================= -->
          <template v-else-if="node.type === 'promptNode'">
            <el-divider content-position="left">提示词</el-divider>
            <el-form-item label="输入提示词">
              <el-input v-model="cfg.text" type="textarea" :rows="4" @change="commit" />
            </el-form-item>
            <el-form-item label="">
              <el-button size="small" @click="saveTextToPromptList"><i class="fas fa-plus" /> 保存到提示词列表</el-button>
              <el-button size="small" @click="importTextInto('text')"><i class="fas fa-file-import" /> 导入文件</el-button>
            </el-form-item>
            <el-form-item label="提示词列表">
              <div class="prompt-list-box">
                <el-checkbox-group v-model="cfg.selectedPrompts">
                  <el-checkbox
                    v-for="(item, idx) in (cfg.promptList || [])"
                    :key="idx"
                    :label="item"
                  >{{ shortText(item) }}</el-checkbox>
                </el-checkbox-group>
                <div v-if="!(cfg.promptList || []).length" class="form-tip">暂无提示词</div>
                <div class="prompt-list-actions">
                  <el-select
                    :model-value="[]"
                    multiple
                    filterable
                    placeholder="从提示词库添加"
                    style="flex: 1"
                    @change="addFromPromptLibrary"
                  >
                    <el-option v-for="p in promptOptions" :key="p.id" :label="p.title" :value="p.id" />
                  </el-select>
                  <el-button size="small" text @click="removeSelectedPrompts">删除选中</el-button>
                  <el-button size="small" text type="danger" @click="clearPromptList">清空</el-button>
                </div>
              </div>
            </el-form-item>
            <el-form-item label="占位符">
              <el-input v-model="cfg.placeholder" placeholder="未设占位符" @change="commit" />
              <div class="form-tip" v-pre>下游可用同名占位符引用这段文本</div>
            </el-form-item>
          </template>

          <!-- ================= 星域·文件输入 ================= -->
          <template v-else-if="node.type === 'fileInput'">
            <el-divider content-position="left">读取设置</el-divider>
            <el-form-item label="文件类型">
              <el-select v-model="cfg.fileType" style="width: 100%" @change="commit">
                <el-option label="文本文件" value="text" />
                <el-option label="JSON 文件" value="json" />
                <el-option label="CSV 文件" value="csv" />
                <el-option label="XML 文件" value="xml" />
                <el-option label="Markdown 文件" value="markdown" />
              </el-select>
            </el-form-item>
            <el-form-item label="文件路径">
              <el-input v-model="cfg.path" placeholder="请输入文件路径..." @change="commit" />
              <el-button size="small" class="mt6" @click="importFileInputContent">
                <i class="fas fa-folder-open" /> 导入文件
              </el-button>
              <div class="form-tip">导入的内容优先于按路径读取</div>
            </el-form-item>
            <el-form-item v-if="cfg.fileName" label="已导入">
              <span class="form-tip">{{ cfg.fileName }}（{{ (cfg.fileContent || '').length }} 字符）</span>
            </el-form-item>
            <el-form-item label="编码格式">
              <el-select v-model="cfg.encoding" style="width: 100%" @change="commit">
                <el-option label="UTF-8" value="utf-8" />
                <el-option label="GBK" value="gbk" />
                <el-option label="ASCII" value="ascii" />
              </el-select>
            </el-form-item>
            <el-form-item label="读取模式">
              <el-select v-model="cfg.readMode" style="width: 100%" @change="commit">
                <el-option label="完整读取" value="full" />
                <el-option label="按行读取" value="lines" />
                <el-option label="分块读取" value="chunks" />
              </el-select>
            </el-form-item>
            <el-form-item label="输出变量">
              <el-input v-model="cfg.outputVariable" placeholder="默认 fileContent" @change="commit" />
            </el-form-item>
          </template>

          <!-- ================= 星域·AI 模型 ================= -->
          <template v-else-if="node.type === 'aiModel'">
            <el-divider content-position="left">模型</el-divider>
            <el-form-item label="AI 模型">
              <ModelSelect :model-value="cfg.configId" @picked="onAiModelPicked" />
            </el-form-item>
            <el-form-item label="加密模式">
              <el-switch v-model="cfg.isEncrypted" @change="onToggleEncryption" />
            </el-form-item>
            <template v-if="!cfg.isEncrypted">
              <el-form-item label="提示词">
                <el-input
                  :ref="el => setPromptRef(el)"
                  v-model="cfg.prompt"
                  type="textarea"
                  :rows="4"
                  placeholder="使用 {{变量名}} 引用上游输出，点击下方变量一键插入"
                  @change="commit"
                />
              </el-form-item>
            </template>
            <template v-else>
              <el-form-item label="加密口令">
                <el-input v-model="cfg.promptPassword" type="password" show-password @change="commit" />
              </el-form-item>
              <el-form-item label="提示词">
                <el-input
                  :ref="el => setPromptRef(el)"
                  v-model="cfg.prompt"
                  type="textarea"
                  :rows="4"
                  placeholder="使用 {{变量名}} 引用上游输出，点击下方变量一键插入"
                  @change="commit"
                />
              </el-form-item>
              <el-form-item label="">
                <el-button size="small" @click="encryptPrompt"><i class="fas fa-lock" /> 生成密文</el-button>
              </el-form-item>
              <el-form-item label="密文提示词">
                <el-input v-model="cfg.encryptedPrompt" type="textarea" :rows="3" placeholder="由明文加密生成，也可直接粘贴密文" @change="commit" />
              </el-form-item>
            </template>
            <el-form-item label="">
              <el-button size="small" @click="importTextInto('aiPrompt')"><i class="fas fa-file-import" /> 导入 MD</el-button>
              <el-popover placement="bottom-start" :width="320" trigger="click">
                <template #reference>
                  <el-button size="small" :disabled="!groupedPromptVars.length">
                    <i class="fas fa-plus" /> 导入占位符
                  </el-button>
                </template>
                <div class="var-pop">
                  <div v-if="!groupedPromptVars.length" class="form-tip">
                    当前节点没有已连接的上游，先从画布把上游节点连进来
                  </div>
                  <div v-for="g in groupedPromptVars" :key="g.nodeId" class="var-pop-group">
                    <div class="var-pop-title">
                      <i class="fas fa-circle-nodes" /> {{ g.label }}
                      <span class="var-pop-type">{{ nodeTypeLabel(g.type) }}</span>
                    </div>
                    <div class="var-pop-vars">
                      <el-tag
                        v-for="v in g.vars"
                        :key="v"
                        size="small"
                        class="var-chip"
                        @click="insertVariable(v)"
                      >{{ chipText(v) }}</el-tag>
                    </div>
                  </div>
                  <div class="form-tip">点击变量插入到提示词光标处；执行时替换为该上游的对应输出</div>
                </div>
              </el-popover>
            </el-form-item>
            <el-form-item label="可用变量">
              <div class="var-chip-row">
                <el-tag
                  v-for="v in promptVariables"
                  :key="v.from + '.' + v.name"
                  size="small"
                  class="var-chip"
                  :title="'来自：' + v.from"
                  @click="insertVariable(v)"
                >{{ chipText(v.name) }}</el-tag>
                <span v-if="!promptVariables.length" class="form-tip">
                  暂无可用变量 —— 占位符只能来自已连接的上游节点，请先在画布上连线
                </span>
              </div>
              <div class="form-tip">只显示实际连线的上游（含多级）；执行时替换为对应输出，未定义的保持原样</div>
            </el-form-item>
            <el-divider content-position="left">生成参数</el-divider>
            <el-form-item label="温度值">
              <el-input-number v-model="cfg.temperature" :min="0" :max="2" :step="0.1" @change="commit" />
            </el-form-item>
            <el-form-item label="最大令牌数">
              <el-input-number v-model="cfg.maxTokens" :min="1" :step="256" @change="commit" />
              <div class="form-tip">留空表示不限制</div>
            </el-form-item>
            <el-form-item label="输入变量">
              <el-select
                v-model="cfg.inputVariable"
                filterable
                allow-create
                default-first-option
                style="width: 100%"
                @change="commit"
              >
                <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
              </el-select>
            </el-form-item>
            <el-form-item label="输出变量">
              <el-input v-model="cfg.outputVariable" placeholder="默认 aiOutput" @change="commit" />
            </el-form-item>

            <el-divider content-position="left">执行控制</el-divider>
            <div class="ai-control-buttons">
              <button
                type="button"
                class="ai-control-btn pause"
                :class="{ active: cfg.controlState === 'paused' }"
                title="分步执行：执行到该节点时挂起，等「继续」再往下走"
                @click="toggleAiPause"
              >
                <i class="fas fa-pause" /> 分步
              </button>
              <button type="button" class="ai-control-btn repeat" title="重复执行该节点" @click="emitNodeAction('repeat')">
                <i class="fas fa-rotate-right" /> 重复
              </button>
              <button type="button" class="ai-control-btn stop" title="停止该节点的生成" @click="emitNodeAction('stop')">
                <i class="fas fa-stop" /> 停止
              </button>
              <button type="button" class="ai-control-btn continue" title="继续执行" @click="emitNodeAction('continue')">
                <i class="fas fa-play" /> 继续
              </button>
            </div>
            <div class="form-tip">停止只中止该节点本次生成，工作流转入分步暂停；重复在执行中即从该节点重跑，未执行时单独跑该节点</div>

            <el-divider content-position="left">生成内容预览</el-divider>
            <div class="ai-output-box">
              <div class="ai-output-head">
                <span class="ai-output-status">{{ genStatusText }}</span>
                <span class="ai-output-count">{{ genText.length }} 字符</span>
              </div>
              <div class="ai-output-controls">
                <el-button size="small" text @click="previewAsMarkdown = !previewAsMarkdown">
                  <i class="fas" :class="previewAsMarkdown ? 'fa-font' : 'fa-code'" />
                  {{ previewAsMarkdown ? '纯文本' : 'MD' }}
                </el-button>
                <el-button size="small" text :disabled="!genText" @click="openModal('ai')">
                  <i class="fas fa-up-right-and-down-left-from-center" /> 放大
                </el-button>
                <el-button size="small" text :disabled="!genText" @click="clearGenerated">
                  <i class="fas fa-trash" /> 清空
                </el-button>
                <el-button size="small" text :disabled="!genText" @click="saveGeneratedToPrompts">
                  <i class="fas fa-bookmark" /> 存提示词库
                </el-button>
              </div>
              <div v-if="previewAsMarkdown && genText" class="ai-output-render" v-html="genRendered" />
              <pre v-else-if="genText" class="ai-output-text">{{ genText }}</pre>
              <div v-else class="ai-output-empty">等待生成...</div>
            </div>
          </template>

          <!-- ================= 星域·文本处理 ================= -->
          <template v-else-if="node.type === 'textProcessor'">
            <el-divider content-position="left">处理设置</el-divider>
            <el-form-item label="处理操作">
              <el-select v-model="cfg.operation" style="width: 100%" @change="commit">
                <el-option label="格式化" value="format" />
                <el-option label="清理文本" value="clean" />
                <el-option label="提取内容" value="extract" />
                <el-option label="替换文本" value="replace" />
                <el-option label="分割文本" value="split" />
                <el-option label="合并文本" value="merge" />
                <el-option label="转换格式" value="transform" />
              </el-select>
              <div class="form-tip">与星域一致：仅「格式化」有内置处理逻辑，其余操作按原文输出</div>
            </el-form-item>
            <el-form-item label="操作参数">
              <el-input
                :model-value="paramsText"
                type="textarea"
                :rows="3"
                placeholder="请输入 JSON 格式的参数..."
                @update:model-value="onParamsInput"
              />
            </el-form-item>
            <el-form-item label="输出格式">
              <el-select v-model="cfg.outputFormat" style="width: 100%" @change="commit">
                <el-option label="纯文本" value="text" />
                <el-option label="JSON" value="json" />
                <el-option label="Markdown" value="markdown" />
                <el-option label="HTML" value="html" />
              </el-select>
            </el-form-item>
            <el-form-item label="保留原格式">
              <el-switch v-model="cfg.preserveFormat" @change="commit" />
            </el-form-item>
            <el-form-item label="源变量">
              <el-select
                v-model="cfg.sourceVariable"
                filterable
                allow-create
                default-first-option
                style="width: 100%"
                @change="commit"
              >
                <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
              </el-select>
            </el-form-item>
          </template>

          <!-- ================= 星域·文本输出 ================= -->
          <template v-else-if="node.type === 'textOutput'">
            <el-divider content-position="left">输出设置</el-divider>
            <el-form-item label="输出格式">
              <el-select v-model="cfg.format" style="width: 100%" @change="commit">
                <el-option label="纯文本" value="text" />
                <el-option label="Markdown" value="markdown" />
                <el-option label="HTML" value="html" />
              </el-select>
            </el-form-item>
            <el-form-item label="文件名">
              <el-input v-model="cfg.filename" placeholder="output.txt" @change="commit" />
            </el-form-item>
            <el-form-item label="源变量">
              <el-select
                v-model="cfg.sourceVariable"
                filterable
                allow-create
                default-first-option
                style="width: 100%"
                @change="commit"
              >
                <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
              </el-select>
            </el-form-item>
            <el-divider content-position="left">内容预览</el-divider>
            <div class="ai-output-box">
              <div class="ai-output-head">
                <span class="ai-output-status">{{ outputStatusText }}</span>
                <span class="ai-output-count">{{ outputPreviewText.length }} 字符</span>
              </div>
              <div class="ai-output-controls">
                <el-button size="small" text @click="outputAsMarkdown = !outputAsMarkdown">
                  <i class="fas" :class="outputAsMarkdown ? 'fa-font' : 'fa-code'" />
                  {{ outputAsMarkdown ? '纯文本' : 'MD' }}
                </el-button>
                <el-button size="small" text :disabled="!outputPreviewText" @click="openModal('text')">
                  <i class="fas fa-up-right-and-down-left-from-center" /> 放大
                </el-button>
                <el-button size="small" text :disabled="!outputPreviewText" @click="clearOutputContent">
                  <i class="fas fa-trash" /> 清空
                </el-button>
                <el-button size="small" text :disabled="!outputPreviewText" @click="saveOutputToPrompts">
                  <i class="fas fa-bookmark" /> 发送到知识库
                </el-button>
              </div>
              <div v-if="outputAsMarkdown && outputPreviewText" class="ai-output-render" v-html="outputRendered" />
              <pre v-else-if="outputPreviewText" class="ai-output-text">{{ outputPreviewText }}</pre>
              <div v-else class="ai-output-empty">等待输入...</div>
            </div>
          </template>

          <!-- ================= 星域·文生图 ================= -->
          <template v-else-if="node.type === 'imageGeneration'">
            <el-divider content-position="left">生成设置</el-divider>
            <el-form-item label="描述文本">
              <el-input v-model="cfg.prompt" type="textarea" :rows="3" placeholder="请输入图片描述..." @change="commit" />
            </el-form-item>
            <el-form-item label="API 提供商">
              <el-select v-model="cfg.apiProvider" style="width: 100%" @change="commit">
                <el-option label="自定义API" value="custom" />
              </el-select>
            </el-form-item>
            <el-form-item label="模型选择">
              <ModelSelect :model-value="cfg.configId" config-type="image" @picked="onMediaModelPicked" />
            </el-form-item>
            <el-form-item label="图片尺寸">
              <div class="size-row">
                <el-input-number v-model="cfg.width" :min="256" :max="2048" :step="64" @change="commit" />
                <span>×</span>
                <el-input-number v-model="cfg.height" :min="256" :max="2048" :step="64" @change="commit" />
              </div>
            </el-form-item>
            <el-form-item label="生成数量">
              <el-input-number v-model="cfg.num_images" :min="1" :max="4" @change="commit" />
            </el-form-item>
            <el-form-item label="风格">
              <el-input v-model="cfg.style" placeholder="可选，如：写实、卡通、油画等" @change="commit" />
            </el-form-item>
            <el-form-item label="输入变量">
              <el-select
                v-model="cfg.inputVariable"
                filterable
                allow-create
                default-first-option
                style="width: 100%"
                @change="commit"
              >
                <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
              </el-select>
            </el-form-item>
          </template>

          <!-- ================= 星域·图片分析 ================= -->
          <template v-else-if="node.type === 'imageAnalyze'">
            <el-divider content-position="left">分析设置</el-divider>
            <el-form-item label="分析说明">
              <el-input v-model="cfg.prompt" type="textarea" :rows="3" placeholder="请输入分析说明..." @change="commit" />
            </el-form-item>
            <el-form-item label="API 提供商">
              <el-select v-model="cfg.apiProvider" style="width: 100%" @change="commit">
                <el-option label="自定义API" value="custom" />
              </el-select>
            </el-form-item>
            <el-form-item label="模型选择">
              <ModelSelect :model-value="cfg.configId" config-type="vision" @picked="onMediaModelPicked" />
            </el-form-item>
            <el-form-item label="输入图片">
              <el-button size="small" @click="importImageInput"><i class="fas fa-image" /> 导入图片</el-button>
              <el-button size="small" :disabled="!cfg.inputImage" @click="clearInputImage">
                <i class="fas fa-times" /> 清除
              </el-button>
              <div v-if="cfg.inputImageName" class="form-tip">已导入：{{ cfg.inputImageName }}</div>
              <div class="form-tip">未导入时从下方源变量取图</div>
            </el-form-item>
            <el-form-item label="输入图片预览">
              <div class="img-preview-box">
                <img v-if="inputImageSrc" :src="inputImageSrc" class="img-preview" alt="输入图片" />
                <div v-else class="img-placeholder">
                  <i class="fas fa-image" />
                  <span>请选择图片文件或等待上游图片输入</span>
                  <small>支持: JPG, PNG, GIF, BMP, WEBP, TIFF, SVG</small>
                </div>
              </div>
            </el-form-item>
            <el-form-item label="源变量">
              <el-select
                v-model="cfg.sourceVariable"
                filterable
                allow-create
                default-first-option
                style="width: 100%"
                @change="commit"
              >
                <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
              </el-select>
            </el-form-item>

            <el-divider content-position="left">分析结果预览</el-divider>
            <el-form-item label="分析状态">
              <span class="status-chip">{{ analysisStatusText }}</span>
            </el-form-item>
            <div class="ai-output-box">
              <div class="ai-output-head">
                <span class="ai-output-status">{{ analysisHeadText }}</span>
                <span class="ai-output-count">{{ analysisText.length }} 字符</span>
              </div>
              <div class="ai-output-controls">
                <el-button size="small" text @click="analysisAsMarkdown = !analysisAsMarkdown">
                  <i class="fas" :class="analysisAsMarkdown ? 'fa-font' : 'fa-code'" />
                  {{ analysisAsMarkdown ? '纯文本' : 'MD' }}
                </el-button>
                <el-button size="small" text :disabled="!analysisText" @click="openModal('analysis')">
                  <i class="fas fa-up-right-and-down-left-from-center" /> 放大
                </el-button>
                <el-button size="small" text :disabled="!analysisText" @click="clearAnalysisResult">
                  <i class="fas fa-trash" /> 清空
                </el-button>
                <el-button size="small" text :disabled="!analysisText" @click="copyAnalysisResult">
                  <i class="fas fa-copy" /> 复制结果
                </el-button>
              </div>
              <div v-if="analysisAsMarkdown && analysisText" class="ai-output-render" v-html="analysisRendered" />
              <pre v-else-if="analysisText" class="ai-output-text">{{ analysisText }}</pre>
              <div v-else class="ai-output-empty">暂无分析结果</div>
            </div>
          </template>

          <!-- ================= 星域·文生视频 ================= -->
          <template v-else-if="node.type === 'videoGeneration'">
            <el-alert type="warning" :closable="false" class="io-alert">
              <template #title>视频服务商尚未接入：需在服务端按选定服务商实现「提交任务 + 轮询」</template>
            </el-alert>
            <el-divider content-position="left">生成设置</el-divider>
            <el-form-item label="描述文本">
              <el-input v-model="cfg.prompt" type="textarea" :rows="3" placeholder="请输入视频描述..." @change="commit" />
            </el-form-item>
            <el-form-item label="API 提供商">
              <el-select v-model="cfg.apiProvider" style="width: 100%" @change="commit">
                <el-option label="自定义API" value="custom" />
              </el-select>
            </el-form-item>
            <el-form-item label="模型选择">
              <ModelSelect :model-value="cfg.configId" config-type="video" @picked="onMediaModelPicked" />
            </el-form-item>
            <el-form-item label="视频尺寸">
              <el-select v-model="cfg.size" style="width: 100%" @change="commit">
                <el-option label="1920x1080 (Full HD)" value="1920x1080" />
                <el-option label="1280x720 (HD)" value="1280x720" />
                <el-option label="1024x576" value="1024x576" />
                <el-option label="3840x2160 (4K)" value="3840x2160" />
              </el-select>
            </el-form-item>
            <el-form-item label="时长(秒)">
              <el-input-number v-model="cfg.duration" :min="1" :max="30" @change="commit" />
            </el-form-item>
            <el-form-item label="帧率">
              <el-select v-model="cfg.fps" style="width: 100%" @change="commit">
                <el-option label="30 FPS" :value="30" />
                <el-option label="60 FPS" :value="60" />
              </el-select>
            </el-form-item>
            <el-form-item label="输出质量">
              <el-select v-model="cfg.quality" style="width: 100%" @change="commit">
                <el-option label="质量优先" value="quality" />
                <el-option label="速度优先" value="speed" />
              </el-select>
            </el-form-item>
            <el-form-item label="音频">
              <el-select v-model="cfg.with_audio" style="width: 100%" @change="commit">
                <el-option label="包含音频" :value="true" />
                <el-option label="无音频" :value="false" />
              </el-select>
            </el-form-item>
            <el-form-item label="风格">
              <el-input v-model="cfg.style" placeholder="可选，如：写实、动画、科幻等" @change="commit" />
            </el-form-item>
            <el-form-item label="输入变量">
              <el-select
                v-model="cfg.inputVariable"
                filterable
                allow-create
                default-first-option
                style="width: 100%"
                @change="commit"
              >
                <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
              </el-select>
            </el-form-item>
          </template>

          <!-- ================= 星域·图生视频 ================= -->
          <template v-else-if="node.type === 'imageToVideo'">
            <el-alert type="warning" :closable="false" class="io-alert">
              <template #title>视频服务商尚未接入：需在服务端按选定服务商实现「提交任务 + 轮询」</template>
            </el-alert>
            <el-divider content-position="left">生成设置</el-divider>
            <el-form-item label="描述文本">
              <el-input v-model="cfg.prompt" type="textarea" :rows="3" placeholder="请输入视频描述..." @change="commit" />
            </el-form-item>
            <el-form-item label="API 提供商">
              <el-select v-model="cfg.apiProvider" style="width: 100%" @change="commit">
                <el-option label="自定义API" value="custom" />
              </el-select>
            </el-form-item>
            <el-form-item label="模型选择">
              <ModelSelect :model-value="cfg.configId" config-type="video" @picked="onMediaModelPicked" />
            </el-form-item>
            <el-form-item label="视频尺寸">
              <div class="size-row">
                <el-input-number v-model="cfg.width" :min="512" :max="1920" :step="64" @change="commit" />
                <span>×</span>
                <el-input-number v-model="cfg.height" :min="512" :max="1080" :step="64" @change="commit" />
              </div>
            </el-form-item>
            <el-form-item label="时长(秒)">
              <el-input-number v-model="cfg.duration" :min="1" :max="30" @change="commit" />
            </el-form-item>
            <el-form-item label="帧率">
              <el-select v-model="cfg.fps" style="width: 100%" @change="commit">
                <el-option label="30 FPS" :value="30" />
                <el-option label="60 FPS" :value="60" />
              </el-select>
            </el-form-item>
            <el-form-item label="风格">
              <el-input v-model="cfg.style" placeholder="可选，如：写实、动画、科幻等" @change="commit" />
            </el-form-item>
            <el-form-item label="输入图片">
              <el-button size="small" @click="importImageInput"><i class="fas fa-image" /> 选择图片</el-button>
              <el-button size="small" :disabled="!cfg.inputImage" @click="clearInputImage">
                <i class="fas fa-times" /> 清除
              </el-button>
              <div v-if="cfg.inputImageName" class="form-tip">已导入：{{ cfg.inputImageName }}</div>
            </el-form-item>
            <el-form-item label="输入图片预览">
              <div class="img-preview-box">
                <img v-if="inputImageSrc" :src="inputImageSrc" class="img-preview" alt="输入图片" />
                <div v-else class="img-placeholder">
                  <i class="fas fa-image" />
                  <span>请选择图片文件或等待上游图片输入</span>
                  <small>支持: JPG, PNG, GIF, BMP, WEBP, TIFF, SVG</small>
                </div>
              </div>
            </el-form-item>
            <el-form-item label="图片源变量">
              <el-select
                v-model="cfg.sourceVariable"
                filterable
                allow-create
                default-first-option
                style="width: 100%"
                @change="commit"
              >
                <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
              </el-select>
              <div class="form-tip">未导入图片时从这里取图</div>
            </el-form-item>

            <el-divider content-position="left">分析结果预览</el-divider>
            <el-form-item label="分析状态">
              <span class="status-chip">{{ analysisStatusText }}</span>
            </el-form-item>
            <div class="ai-output-box">
              <div class="ai-output-head">
                <span class="ai-output-status">{{ videoStatusText }}</span>
                <span class="ai-output-count">{{ videoSrc ? '1 个视频' : '暂无' }}</span>
              </div>
              <div class="ai-output-controls">
                <el-button size="small" text :disabled="!videoSrc" @click="clearVideoResult">
                  <i class="fas fa-trash" /> 清空结果
                </el-button>
              </div>
              <div v-if="videoSrc" class="video-preview-box">
                <video :src="videoSrc" controls preload="metadata" />
              </div>
              <div v-else class="ai-output-empty">生成完成后在此预览</div>
            </div>
          </template>

          <!-- ================= 星域·图片输出 ================= -->
          <template v-else-if="node.type === 'imageOutput'">
            <el-divider content-position="left">输出设置</el-divider>
            <el-form-item label="文件名">
              <el-input v-model="cfg.filename" placeholder="output.png" @change="commit" />
            </el-form-item>
            <el-form-item label="图片格式">
              <el-select v-model="cfg.format" style="width: 100%" @change="commit">
                <el-option label="PNG" value="png" />
                <el-option label="JPG" value="jpg" />
                <el-option label="WEBP" value="webp" />
              </el-select>
            </el-form-item>
            <el-form-item label="源变量">
              <el-select
                v-model="cfg.sourceVariable"
                filterable
                allow-create
                default-first-option
                style="width: 100%"
                @change="commit"
              >
                <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
              </el-select>
            </el-form-item>
            <div class="form-tip">支持 images 数组 / 单张 base64 / 图片 URL；多张时自动加序号</div>
            <el-divider content-position="left">图片预览</el-divider>
            <div class="img-preview-box">
              <img v-if="panelImageSrc" :src="panelImageSrc" class="img-preview" alt="图片预览" />
              <div v-else class="img-placeholder">
                <i class="fas fa-image" />
                <span>暂无图片</span>
                <small>执行后在这里预览（落盘为相对路径，浏览器不能直接读取）</small>
              </div>
            </div>
            <div v-if="savedImageCount" class="form-tip">已保存 {{ savedImageCount }} 张</div>
          </template>

          <!-- ================= 星域·视频输出 ================= -->
          <template v-else-if="node.type === 'videoOutput'">
            <el-divider content-position="left">输出设置</el-divider>
            <el-form-item label="文件名">
              <el-input v-model="cfg.filename" placeholder="output.mp4" @change="commit" />
            </el-form-item>
            <el-form-item label="视频格式">
              <el-select v-model="cfg.format" style="width: 100%" @change="commit">
                <el-option label="MP4" value="mp4" />
                <el-option label="AVI" value="avi" />
                <el-option label="MOV" value="mov" />
              </el-select>
            </el-form-item>
            <el-form-item label="源变量">
              <el-select
                v-model="cfg.sourceVariable"
                filterable
                allow-create
                default-first-option
                style="width: 100%"
                @change="commit"
              >
                <el-option v-for="n in knownVariables" :key="n" :label="n" :value="n" />
              </el-select>
            </el-form-item>
            <el-divider content-position="left">视频预览</el-divider>
            <div class="video-preview-box">
              <video v-if="panelVideoSrc" :src="panelVideoSrc" controls preload="metadata" />
              <div v-else class="img-placeholder">
                <i class="fas fa-video" />
                <span>暂无视频</span>
                <small>执行后在这里预览（外链可直接播放，本地落盘路径不可读）</small>
              </div>
            </div>
          </template>

          <!-- 执行历史 -->
          <div v-if="node.data.executionHistory?.length" class="execution-history">
            <el-collapse>
              <el-collapse-item :title="`执行历史（${node.data.executionHistory.length}）`">
                <div v-for="(h, i) in node.data.executionHistory" :key="i" class="history-item">
                  <div class="history-item-title">
                    <el-tag size="small" :type="h.status === 'success' ? 'success' : h.status === 'error' ? 'danger' : 'info'">
                      {{ h.status }}
                    </el-tag>
                    <span class="history-time">{{ h.timestamp }}</span>
                  </div>
                  <div v-if="h.error" class="error-message">{{ h.error }}</div>
                  <div v-if="h.outputs && Object.keys(h.outputs).length" class="variable-list">
                    <div v-for="(v, k) in h.outputs" :key="k" class="variable-item-vertical">
                      <span class="var-name">{{ k }}</span>
                      <span class="var-value">{{ previewValue(v) }}</span>
                    </div>
                  </div>
                </div>
              </el-collapse-item>
            </el-collapse>
          </div>
        </el-form>
      </div>
    </div>
  </transition>

  <!-- 输出放大弹窗（AI 输出 / 文本输出 / 图片分析 共用，支持编辑） -->
  <el-dialog v-model="modal.visible" :title="modalTitle" width="820px" append-to-body>
    <div class="modal-toolbar">
      <span class="io-count">{{ modalText.length }} 字符</span>
      <el-button size="small" @click="modal.asMarkdown = !modal.asMarkdown">
        <i class="fas" :class="modal.asMarkdown ? 'fa-font' : 'fa-code'" />
        {{ modal.asMarkdown ? '纯文本' : 'MD' }}
      </el-button>
      <el-button v-if="!modal.editing" size="small" @click="startEdit">
        <i class="fas fa-pen" /> 编辑
      </el-button>
      <template v-else>
        <el-button size="small" type="primary" @click="saveEdit"><i class="fas fa-check" /> 保存</el-button>
        <el-button size="small" @click="modal.editing = false"><i class="fas fa-xmark" /> 取消</el-button>
      </template>
      <el-button size="small" @click="copyModalText"><i class="fas fa-copy" /> 复制</el-button>
    </div>
    <el-input v-if="modal.editing" v-model="modal.draft" type="textarea" :rows="18" />
    <div v-else-if="modal.asMarkdown" class="ai-output-render dialog-render" v-html="modalRendered" />
    <el-input v-else :model-value="modalText" type="textarea" :rows="18" readonly />
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { ElInput, ElMessage } from 'element-plus'
import { promptAPI } from '@/api'
import ModelSelect from './ModelSelect.vue'
import { buildModelSnapshot } from '@/services/workflow/modelResolve'
import { encryptContent } from '@/utils/promptCrypto'
import MarkdownIt from 'markdown-it'
import type { ModelSnapshot, WEdge, WNode } from '@/types/workflow'

const props = defineProps<{
  visible: boolean
  node: WNode | null
  knownVariables: string[]
  /** 画布全量节点/连线：用于计算「实际连线的上游」能提供哪些占位符变量 */
  nodes: WNode[]
  edges: WEdge[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'change'): void
  /** 星域「执行控制」按钮：继续 / 重复 / 停止（由编辑器转发给执行器） */
  (e: 'continue-node', id: string): void
  (e: 'repeat-node', id: string): void
  (e: 'stop-node', id: string): void
  /** 分步开关变化，父组件用于同步 store 断点（卡片红点与挂起面板） */
  (e: 'pause-toggle', id: string, paused: boolean): void
}>()

const form = reactive({ label: '', config: {} as Record<string, any> })
const varRows = ref<Array<{ k: string; v: string }>>([])

const promptOptions = ref<Array<{ id: number; title: string }>>([])

/** 下拉数据：提示词库（星域「提示词列表」由此挑选） */
const loadPromptOptions = async () => {
  if (promptOptions.value.length) return
  try {
    const res: any = await promptAPI.getAll()
    promptOptions.value = (res?.data || []).map((p: any) => ({ id: p.id, title: p.title || `#${p.id}` }))
  } catch {
    /* 下拉为空即可 */
  }
}

const cfg = computed(() => {
  if (!props.node) return {} as Record<string, any>
  if (!props.node.data.config) props.node.data.config = {}
  return props.node.data.config
})

watch(
  () => props.visible,
  v => {
    if (v) void loadPromptOptions()
  }
)

watch(
  () => props.node?.id,
  () => {
    if (!props.node) return
    void loadPromptOptions()
    form.label = props.node.data.label || ''
    form.config = props.node.data.config || {}
    const init = form.config.initialVariables || {}
    varRows.value = Object.entries(init).map(([k, v]) => ({ k, v: typeof v === 'string' ? v : JSON.stringify(v) }))
  },
  { immediate: true }
)

/** 把当前文本输出节点的结果存进提示词库（文本输出节点的「存入知识库」动作） */
const saveNodeTextToPrompts = async () => {
  const node = props.node
  if (!node) return
  const content = String(node.data?.result || '')
  if (!content.trim()) {
    ElMessage.warning('还没有可保存的内容')
    return
  }
  const cfgv = node.data?.config || {}
  const title = String(cfgv.filename || node.data?.label || '工作流输出').slice(0, 60)
  try {
    await promptAPI.create({ title, content, category: '工作流输出' } as any)
    ElMessage.success('已存入提示词库')
  } catch {
    /* 拦截器已提示 */
  }
}

const commit = () => {
  if (!props.node) return
  props.node.data.label = form.label
  emit('change')
}

const parseMaybeJson = (raw: string) => {
  const s = String(raw ?? '').trim()
  if (!s) return ''
  if ((s.startsWith('{') && s.endsWith('}')) || (s.startsWith('[') && s.endsWith(']'))) {
    try {
      return JSON.parse(s)
    } catch {
      return raw
    }
  }
  return raw
}

const commitVars = () => {
  const out: Record<string, any> = {}
  varRows.value.forEach(r => {
    const key = r.k.trim()
    if (key) out[key] = parseMaybeJson(r.v)
  })
  cfg.value.initialVariables = out
  commit()
}
const addVar = () => varRows.value.push({ k: '', v: '' })
const removeVar = (i: number) => {
  varRows.value.splice(i, 1)
  commitVars()
}

const toNames = (v: any): string[] => {
  if (!Array.isArray(v)) return []
  return v.map((x: any) => (typeof x === 'string' ? x : x?.variableName)).filter(Boolean)
}
const toRefs = (names: string[]) => names.map(n => ({ variableName: n }))

const ctxNames = computed(() => toNames(cfg.value.contextVariables))
const ctxInputNames = computed(() => toNames(cfg.value.inputVariables))
const mergeNames = computed(() => (Array.isArray(cfg.value.inputVariables) ? cfg.value.inputVariables : []))
const includeNames = computed(() => (Array.isArray(cfg.value.includeOutputs) ? cfg.value.includeOutputs : []))
const extraOutVar = computed(() => (typeof cfg.value.outputVariable === 'string' ? cfg.value.outputVariable : ''))

const addCtxVar = () => {
  cfg.value.contextVariables = [...ctxNames.value.map(n => ({ variableName: n })), { variableName: '' }]
  commit()
}
const setCtxVar = (i: number, v: string) => {
  const arr = [...ctxNames.value]
  arr[i] = v
  cfg.value.contextVariables = toRefs(arr.filter(Boolean))
  commit()
}
const removeCtxVar = (i: number) => {
  const arr = [...ctxNames.value]
  arr.splice(i, 1)
  cfg.value.contextVariables = toRefs(arr)
  commit()
}

/** 数组类变量（循环/上下文的候选） */
const arrayVariables = computed(() =>
  props.knownVariables.filter(n => /plan|plans|chapter|chapters|items|list|results|history|History/i.test(n))
)

const onModelPicked = (p: { configId: number; modelId: string; modelConfig: ModelSnapshot }) => {
  cfg.value.configId = p.configId
  cfg.value.modelId = p.modelId
  cfg.value.modelConfig = p.modelConfig
  commit()
}

/** 星域 AI 模型节点：星域口径存模型名（model），星芒另存 configId 供调用 */
const onAiModelPicked = (p: { configId: number; modelId: string; modelConfig: ModelSnapshot }) => {
  cfg.value.model = p.modelConfig?.modelName || ''
  onModelPicked(p)
}

/** 星域媒体节点：同样把模型名写进 model 字段 */
const onMediaModelPicked = onAiModelPicked

/* ---------------- 星域 AI 节点的执行控制与生成内容预览 ---------------- */

const md = new MarkdownIt({ html: false, linkify: true, breaks: true })
const previewAsMarkdown = ref(false)

/** 生成内容：执行期流式文本优先，其次最终结果 */
const genText = computed(() => {
  const d: any = props.node?.data
  if (!d) return ''
  const stream = typeof d.streamText === 'string' ? d.streamText : ''
  if (stream) return stream
  return typeof d.result === 'string' ? d.result : ''
})
const genRendered = computed(() => md.render(genText.value || ''))
const genStatusText = computed(() => {
  const d: any = props.node?.data
  if (!d) return '等待生成...'
  if (d.status === 'running') return '正在生成...'
  if (d.status === 'error') return '生成失败'
  if (cfg.value.generatedStoppedAt) return '已停止'
  if (d.status === 'success') return '生成完成'
  return '等待生成...'
})

/** 分步：切换星域字段 controlState（执行器与 store 断点都认它） */
const toggleAiPause = () => {
  const paused = cfg.value.controlState !== 'paused'
  cfg.value.controlState = paused ? 'paused' : null
  commit()
  emit('pause-toggle', props.node?.id || '', paused)
}

const emitNodeAction = (kind: 'repeat' | 'stop' | 'continue') => {
  const id = props.node?.id
  if (!id) return
  if (kind === 'repeat') emit('repeat-node', id)
  else if (kind === 'stop') emit('stop-node', id)
  else emit('continue-node', id)
}

const clearGenerated = () => {
  const node = props.node
  if (!node) return
  node.data.result = null
  ;(node.data as any).streamText = ''
  node.data.status = 'idle'
  delete cfg.value.generatedStoppedAt
  commit()
  ElMessage.success('已清空生成内容')
}

const saveGeneratedToPrompts = async () => {
  const content = genText.value
  if (!content.trim()) return
  try {
    await promptAPI.create({
      title: String(cfg.value.prompt || props.node?.data?.label || 'AI 输出').slice(0, 60),
      content,
      category: '工作流输出'
    } as any)
    ElMessage.success('已存入提示词库')
  } catch {
    /* 拦截器已提示 */
  }
}

/* ---------------- 星域节点：面板预览（图片 / 视频 / 分析结果 / 文本输出） ---------------- */

/** 节点内导入的输入图片（dataURL / URL） */
const inputImageSrc = computed(() => String(cfg.value.inputImage || ''))

/** 输出图片预览：优先运行期 images（base64 可直接显示），落盘相对路径浏览器读不到 */
const panelImageSrc = computed(() => {
  const imgs = (props.node?.data as any)?.images
  if (Array.isArray(imgs) && imgs.length) {
    const first = imgs[0] || {}
    if (first.b64) return `data:image/png;base64,${first.b64}`
    if (first.url) return String(first.url)
  }
  const saved = cfg.value.savedImages
  if (Array.isArray(saved) && saved.length) {
    const v = String(saved[0] || '')
    return v.startsWith('http') || v.startsWith('data:') ? v : ''
  }
  return ''
})
const savedImageCount = computed(() => (Array.isArray(cfg.value.savedImages) ? cfg.value.savedImages.length : 0))

/** 视频预览（外链可直接播放） */
const videoSrc = computed(() => {
  const v = (props.node?.data as any)?.videoUrl
  return typeof v === 'string' ? v : ''
})
const panelVideoSrc = computed(() => {
  const v = String(videoSrc.value || (Array.isArray(cfg.value.savedVideos) ? cfg.value.savedVideos[0] : '') || '')
  return v.startsWith('http') || v.startsWith('data:') ? v : ''
})
const videoStatusText = computed(() => {
  const d: any = props.node?.data
  if (d?.status === 'running') return '正在生成...'
  if (videoSrc.value) return '生成完成'
  return '等待生成...'
})

/**
 * 分析状态 / 结果（图片分析、图生视频共用星域字段 analysisStatus / analysisResult）
 * 注意：星域源码里analysisResult 两种形态都出现过——图片分析渲染读 `.content` / `.isComplete`（对象），
 * 而复制/清空处当纯字符串用（且那两处是原版自己的 bug）。这里两种都兼容，
 * 避免导入旧工作流或历史数据时面板空白。
 */
const analysisRaw = computed(() => cfg.value.analysisResult ?? '')
const analysisText = computed(() => {
  const raw = analysisRaw.value
  return typeof raw === 'string' ? raw : String(raw?.content || '')
})
const analysisIsComplete = computed(() => {
  const raw = analysisRaw.value
  if (raw && typeof raw === 'object') return raw.isComplete !== false
  return false
})
const analysisStatusText = computed(() => String(cfg.value.analysisStatus || '等待分析'))
const analysisHeadText = computed(() => {
  if (analysisIsComplete.value) return '分析完成'
  return analysisText.value ? '正在分析...' : '等待分析...'
})
const analysisAsMarkdown = ref(false)
const analysisRendered = computed(() => md.render(analysisText.value || ''))

/** 文本输出的内容预览（星域字段 outputContent） */
const outputPreviewText = computed(() => {
  const c = String(cfg.value.outputContent || '')
  if (c) return c
  const r = props.node?.data?.result
  return typeof r === 'string' ? r : ''
})
const outputAsMarkdown = ref(false)
const outputRendered = computed(() => md.render(outputPreviewText.value || ''))
const outputStatusText = computed(() => {
  const d: any = props.node?.data
  if (d?.status === 'running') return '正在写出...'
  if (outputPreviewText.value) return '已生成'
  return '等待输入...'
})

/* ---------------- 放大弹窗（支持编辑） ---------------- */
const modal = reactive({
  visible: false,
  target: 'ai' as 'ai' | 'text' | 'analysis',
  asMarkdown: false,
  editing: false,
  draft: ''
})
const modalTitle = computed(() => {
  if (modal.target === 'text') return '输出内容'
  if (modal.target === 'analysis') return '分析结果'
  return '生成内容'
})
const modalText = computed(() => {
  if (modal.target === 'text') return outputPreviewText.value
  if (modal.target === 'analysis') return analysisText.value
  return genText.value
})
const modalRendered = computed(() => md.render(modalText.value || ''))

const openModal = (target: 'ai' | 'text' | 'analysis') => {
  modal.target = target
  modal.editing = false
  modal.asMarkdown = false
  modal.draft = ''
  modal.visible = true
}
const startEdit = () => {
  modal.draft = modalText.value
  modal.editing = true
}
const saveEdit = () => {
  const text = modal.draft
  if (modal.target === 'text') cfg.value.outputContent = text
  else if (modal.target === 'analysis') {
    // 星域把内容和完成标记拆成两个平级字段（analysisResult + analysisComplete）
    cfg.value.analysisResult = { content: text, isComplete: true }
    cfg.value.analysisComplete = true
  }
  else {
    const node = props.node
    if (node) {
      node.data.result = text
      ;(node.data as any).streamText = text
    }
  }
  modal.editing = false
  commit()
  ElMessage.success('已保存修改')
}
const copyModalText = async () => {
  if (!modalText.value) return
  try {
    await navigator.clipboard.writeText(modalText.value)
    ElMessage.success('已复制')
  } catch {
    ElMessage.warning('复制失败，请手动选择文本')
  }
}

/* ---------------- 各节点的小动作 ---------------- */
const clearInputImage = () => {
  cfg.value.inputImage = ''
  cfg.value.inputImageName = ''
  cfg.value.analysisStatus = '等待分析'
  commit()
  ElMessage.success('已清除图片')
}

/** 星域：把占位符文本与接收的数据互换 */
const swapTransferData = () => {
  const placeholder = String(cfg.value.placeholder || '')
  const received = String(cfg.value.receivedData || '')
  // 星域：没有接收到数据时互换没有意义，且会把占位符清空
  if (!received.trim()) {
    ElMessage.warning('当前没有接收到数据，无法进行互换')
    return
  }
  cfg.value.placeholder = received
  cfg.value.receivedData = placeholder
  commit()
  ElMessage.success('已互换占位符与接收数据')
}

const copyAnalysisResult = async () => {
  if (!analysisText.value) return
  try {
    await navigator.clipboard.writeText(analysisText.value)
    ElMessage.success('已复制分析结果')
  } catch {
    ElMessage.warning('复制失败，请手动选择文本')
  }
}
const clearAnalysisResult = () => {
  cfg.value.analysisResult = null
  cfg.value.analysisStatus = '等待分析'
  commit()
  ElMessage.success('已清空分析结果')
}
const clearOutputContent = () => {
  cfg.value.outputContent = ''
  const node = props.node
  if (node) {
    node.data.result = null
    ;(node.data as any).streamText = ''
  }
  commit()
  ElMessage.success('已清空输出内容')
}
const saveOutputToPrompts = async () => {
  const content = outputPreviewText.value
  if (!content.trim()) return
  try {
    await promptAPI.create({
      title: String(cfg.value.filename || props.node?.data?.label || '工作流输出').slice(0, 60),
      content,
      category: '工作流输出'
    } as any)
    ElMessage.success('已发送到知识库')
  } catch {
    /* 拦截器已提示 */
  }
}
const clearVideoResult = () => {
  cfg.value.analysisStatus = '等待生成'
  cfg.value.savedVideos = []
  commit()
  ElMessage.success('已清空生成结果')
}

/* ---------------- 星域节点辅助 ---------------- */

/** 提示词列表条目在面板上的截断文本 */
const shortText = (s: any) => {
  const t = String(s ?? '')
  return t.length > 20 ? t.slice(0, 20) + '…' : t
}

/** 把「输入提示词」存进提示词列表（星域：保存到提示词列表） */
const saveTextToPromptList = () => {
  const text = String(cfg.value.text || '').trim()
  if (!text) {
    ElMessage.warning('请先输入提示词内容')
    return
  }
  if (!Array.isArray(cfg.value.promptList)) cfg.value.promptList = []
  if (cfg.value.promptList.includes(text)) {
    ElMessage.info('提示词列表里已有相同内容')
    return
  }
  cfg.value.promptList.push(text)
  if (!Array.isArray(cfg.value.selectedPrompts)) cfg.value.selectedPrompts = []
  cfg.value.selectedPrompts.push(text)
  commit()
  ElMessage.success('已保存到提示词列表')
}

/** 从提示词库挑选条目追加进提示词列表（星域：选取） */
const addFromPromptLibrary = async (ids: number[]) => {
  if (!Array.isArray(ids) || !ids.length) return
  if (!Array.isArray(cfg.value.promptList)) cfg.value.promptList = []
  for (const id of ids) {
    try {
      const res: any = await promptAPI.getOne(Number(id))
      const content = String(res?.data?.content || '')
      if (content && !cfg.value.promptList.includes(content)) cfg.value.promptList.push(content)
    } catch {
      /* 单条失败不影响其它 */
    }
  }
  commit()
  ElMessage.success('已添加到提示词列表')
}

const removeSelectedPrompts = () => {
  const selected: string[] = Array.isArray(cfg.value.selectedPrompts) ? cfg.value.selectedPrompts : []
  if (!selected.length) {
    ElMessage.warning('请先勾选要删除的提示词')
    return
  }
  cfg.value.promptList = (cfg.value.promptList || []).filter((x: any) => !selected.includes(x))
  cfg.value.selectedPrompts = []
  commit()
  ElMessage.success(`已删除 ${selected.length} 条提示词`)
}

const clearPromptList = () => {
  cfg.value.promptList = []
  cfg.value.selectedPrompts = []
  commit()
}

/** 打开系统文件选择框并读取内容（浏览器 FileReader，不经过后端） */
const pickFile = (accept: string, asDataUrl: boolean, onLoad: (text: string, name: string) => void) => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = accept
  input.style.display = 'none'
  document.body.appendChild(input)
  const cleanup = () => input.remove()
  input.onchange = () => {
    const file = input.files?.[0]
    if (!file) return cleanup()
    const reader = new FileReader()
    reader.onload = () => {
      onLoad(String(reader.result || ''), file.name)
      cleanup()
    }
    reader.onerror = () => {
      ElMessage.error('文件读取失败')
      cleanup()
    }
    if (asDataUrl) reader.readAsDataURL(file)
    else reader.readAsText(file, 'utf-8')
  }
  input.click()
}

/** 文本输入 / 提示词 / AI 模型：导入文本文件 */
/* ---------------- AI 模型节点：占位符插入（执行侧 replaceVariables 已支持 {{var}}，这里补 UI 侧） ---------------- */
const promptInputRef = ref<InstanceType<typeof ElInput> | null>(null)
const setPromptRef = (el: any) => {
  // 两个提示词框（明文/加密模式）共用一个 ref 槽，最后渲染的生效
  if (el?.textarea || el?.focus) promptInputRef.value = el
}

/* ---------- 占位符变量：只收「实际连线的上游」---------- */

/**
 * 各节点类型的输出键，只列「提示词模板里有意义」的：
 * 执行器 outputs 为兼容多种取值习惯会带同值冗余键（textInput 输出 5 个键里 3 个同值、2 个元数据），
 * 插入列表照搬会让用户看到 4 个值一样的变量。规则：内容主键 + 值有区分度的键；
 * lastOutput 等通用接口键不进列表，但手打 {{lastOutput}} 依然有效（执行器不受影响）。
 * 动态键按节点配置解析：占位符名、outputVariable、variableName、initialVariables。
 */
const outputKeysOf = (n: WNode): string[] => {
  const c: any = n.data?.config || {}
  const explicitOutVar = () => {
    const v = String(c.outputVariable || '').trim()
    return v ? [v] : []
  }
  const placeholderKey = (fallback: string) =>
    c.placeholder && c.placeholder !== '未设占位符' ? String(c.placeholder) : fallback
  switch (n.type) {
    case 'textInput':
    case 'promptNode':
      return [placeholderKey('textInput')]
    case 'aiModel':
      return ['lastOutput', ...explicitOutVar()]
    case 'aiProcess':
      return ['aiResult', ...explicitOutVar()]
    case 'text':
      return ['text', ...explicitOutVar()]
    case 'fileInput':
      return ['fileContent', ...explicitOutVar()]
    case 'textProcessor':
      return ['textProcessed']
    case 'textOutput':
      return ['textOutput', 'outputFile']
    case 'fileOutput':
      return ['fileContent', 'outputFile']
    case 'imageGeneration':
      return ['imageB64', 'imageUrl']
    case 'imageToVideo':
    case 'videoGeneration':
      return ['videoUrl']
    case 'imageOutput':
      return ['files']
    case 'videoOutput':
      return ['files', 'videoUrl']
    case 'imageAnalyze':
      return ['imageAnalysis']
    case 'transfer':
      return [placeholderKey('transferred')]
    case 'merge':
      return ['merged']
    case 'loop':
      return ['loopResults', 'iterationCount']
    case 'context':
      return c.variableName ? [String(c.variableName)] : []
    case 'start':
      return c.initialVariables && typeof c.initialVariables === 'object' ? Object.keys(c.initialVariables) : []
    case 'comment':
      return ['comment']
    default:
      return []
  }
}

interface VarGroup {
  nodeId: string
  label: string
  type: string
  vars: string[]
}

/**
 * 实际连线可达的上游变量（按来源节点分组）。
 * 从当前节点沿入边向上 DFS：直接上游和多级上游都算「连上了」（数据沿链路下传，
 * 执行顺序上先上游后下游，context 里确实有它们的输出）；未连入的分支一律不给。
 */
const upstreamVarGroups = computed<VarGroup[]>(() => {
  const self = props.node
  if (!self) return []
  const byId = new Map((props.nodes || []).map(n => [n.id, n]))
  const incoming = new Map<string, string[]>()
  for (const e of props.edges || []) {
    if (!incoming.has(e.target)) incoming.set(e.target, [])
    incoming.get(e.target)!.push(e.source)
  }
  const visited = new Set<string>([self.id])
  const stack = [...(incoming.get(self.id) || [])]
  const groups: VarGroup[] = []
  while (stack.length) {
    const id = stack.pop()!
    if (visited.has(id)) continue
    visited.add(id)
    const n = byId.get(id)
    if (!n) continue
    const vars = outputKeysOf(n)
    if (vars.length) {
      groups.push({ nodeId: n.id, label: n.data?.label || n.type || id, type: n.type || '', vars })
    }
    for (const src of incoming.get(id) || []) if (!visited.has(src)) stack.push(src)
  }
  return groups
})

/** chips 平铺展示（带来源标注） */
const promptVariables = computed(() =>
  upstreamVarGroups.value.flatMap(g => g.vars.map(v => ({ name: v, from: g.label })))
)

/** 弹出面板里按来源节点分组、去重后的变量 */
const groupedPromptVars = computed(() =>
  upstreamVarGroups.value.map(g => ({ ...g, vars: Array.from(new Set(g.vars)) })).filter(g => g.vars.length)
)

/** 节点类型中文名（弹窗里标来源用） */
const nodeTypeLabel = (t: string) => {
  const map: Record<string, string> = {
    textInput: '文本输入', promptNode: '提示词', aiModel: 'AI 模型', aiProcess: 'AI 处理',
    text: '文本处理', textProcessor: '文本处理', merge: '合并', loop: '循环', context: '上下文',
    start: '开始', end: '结束', comment: '注释', transfer: '中转', fileInput: '文件输入',
    textOutput: '文本输出', fileOutput: '文件输出', imageGeneration: '图片生成',
    imageToVideo: '图生视频', videoGeneration: '文生视频', imageOutput: '图片输出',
    videoOutput: '视频输出', imageAnalyze: '图片分析'
  }
  return map[t] || t
}

/** 变量 chip 的显示文本（mustache 里的字面 '}}' 会被当成插值结束符，必须放 script） */
const chipText = (name: string) => '{' + '{' + name + '}' + '}'

/** 把 {{变量}} 插到提示词光标处；拿不到光标就追加到末尾 */
const insertVariable = (name: string | { name: string }) => {
  const varName = typeof name === 'string' ? name : name.name
  const snippet = `{{${varName}}}`
  const ta = (promptInputRef.value as any)?.textarea as HTMLTextAreaElement | undefined
  const cur = String(cfg.value.prompt || '')
  if (ta && document.activeElement !== ta) ta.focus()
  if (ta && typeof ta.selectionStart === 'number') {
    const start = ta.selectionStart
    const end = ta.selectionEnd
    cfg.value.prompt = cur.slice(0, start) + snippet + cur.slice(end)
    nextTick(() => {
      ta.focus()
      ta.setSelectionRange(start + snippet.length, start + snippet.length)
    })
  } else {
    cfg.value.prompt = cur + snippet
  }
  commit()
}

const importTextInto = (target: 'text' | 'aiPrompt') => {
  pickFile('.txt,.md,.json,.csv,.xml', false, (text, name) => {
    if (target === 'aiPrompt') cfg.value.prompt = text
    else cfg.value.text = text
    commit()
    ElMessage.success(`已导入 ${name}（${text.length} 字符）`)
  })
}

/** 文件输入节点：导入内容（优先于按路径读取） */
const importFileInputContent = () => {
  pickFile('.txt,.md,.json,.csv,.xml', false, (text, name) => {
    cfg.value.fileContent = text
    cfg.value.fileName = name
    cfg.value.fileSize = text.length
    commit()
    ElMessage.success(`已导入 ${name}`)
  })
}

/** 图片分析节点：导入图片（dataURL） */
const importImageInput = () => {
  pickFile('image/*', true, (dataUrl, name) => {
    cfg.value.inputImage = dataUrl
    cfg.value.inputImageName = name
    commit()
    ElMessage.success(`已导入图片 ${name}`)
  })
}

/** 星域 AI 模型：把明文加密成密文提示词 */
const encryptPrompt = async () => {
  const plain = String(cfg.value.prompt || '')
  const password = String(cfg.value.promptPassword || '')
  if (!plain.trim()) {
    ElMessage.warning('请先输入提示词')
    return
  }
  if (!password) {
    ElMessage.warning('请先设置加密口令')
    return
  }
  try {
    cfg.value.encryptedPrompt = await encryptContent(plain, password)
    commit()
    ElMessage.success('已生成密文提示词')
  } catch {
    ElMessage.error('加密失败')
  }
}

const onToggleEncryption = async () => {
  if (cfg.value.isEncrypted && cfg.value.prompt && !cfg.value.encryptedPrompt) {
    await encryptPrompt()
  }
  commit()
}

/** 文本处理节点：操作参数（JSON 文本编辑） */
const paramsText = computed(() => {
  const p = cfg.value.parameters
  return p && typeof p === 'object' && Object.keys(p).length ? JSON.stringify(p, null, 2) : ''
})

const onParamsInput = (v: string) => {
  const t = (v || '').trim()
  if (!t) {
    cfg.value.parameters = {}
    commit()
    return
  }
  try {
    cfg.value.parameters = JSON.parse(t)
  } catch {
    /* 输入过程中的半成品 JSON 不写回 */
  }
}

const previewValue = (v: any) => {
  if (v === null || v === undefined) return '(空)'
  if (typeof v === 'string') return v.length > 200 ? v.slice(0, 200) + '...' : v
  try {
    const s = JSON.stringify(v, null, 2)
    return s.length > 300 ? s.slice(0, 300) + '...' : s
  } catch {
    return String(v)
  }
}
</script>

<style scoped>
/* 源版 .wf-drawer：固定右上浮层 */
.wf-drawer {
  position: fixed;
  top: 13vh;
  right: 16px;
  height: 80vh;
  width: 360px;
  background: var(--wf-el-bg, #ffffff);
  border: 1px solid var(--wf-el-border, #dcdfe6);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
  z-index: 2022;
  pointer-events: auto;
}
.wf-drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-bottom: 1px solid var(--wf-el-border, #dcdfe6);
  background: var(--wf-el-bg, #ffffff);
  border-radius: 8px 8px 0 0;
}
.wf-drawer-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: var(--wf-el-text-primary, #303133);
}
.wf-drawer-title i {
  color: var(--el-color-primary, #409eff);
}
.wf-drawer-body {
  padding: 12px 12px 16px;
  overflow: auto;
  flex: 1;
  background: var(--wf-el-fill-blank, #ffffff);
}
.slide-fade-enter-active {
  transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
}
.slide-fade-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 1, 1);
}
.slide-fade-enter-from,
.slide-fade-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

.property-form :deep(.el-form-item) {
  margin-bottom: 20px;
}
.property-form :deep(.el-divider) {
  margin: 24px 0 20px;
}
/* ---------- 图片 / 视频预览 ---------- */
.img-preview-box {
  width: 100%;
  min-height: 110px;
  border: 1px solid var(--wf-el-border-light, #ebeef5);
  border-radius: 4px;
  background: var(--wf-el-fill-lighter, #fafafa);
  padding: 8px;
  text-align: center;
}
.img-preview {
  max-width: 100%;
  max-height: 200px;
  object-fit: contain;
  border-radius: 4px;
}
.img-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-height: 96px;
  color: var(--wf-el-text-placeholder, #a8abb2);
  font-size: 12px;
}
.img-placeholder i {
  font-size: 22px;
}
.img-placeholder small {
  font-size: 11px;
  opacity: 0.8;
}
.video-preview-box {
  border-radius: 4px;
  overflow: hidden;
  background: #000;
}
.video-preview-box video {
  display: block;
  width: 100%;
  max-height: 240px;
}
.status-chip {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 10px;
  font-size: 12px;
  color: var(--el-color-warning, #e6a23c);
  background: rgba(230, 162, 60, 0.12);
}
.modal-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.modal-toolbar .io-count {
  margin-right: auto;
  font-size: 12px;
  color: var(--wf-el-text-secondary, #909399);
}
/* ---------- 导入占位符弹窗 ---------- */
.var-pop {
  max-height: 320px;
  overflow-y: auto;
}
.var-pop-group {
  margin-bottom: 10px;
}
.var-pop-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--wf-el-text-primary, #303133);
  margin-bottom: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
}
.var-pop-type {
  font-weight: 400;
  font-size: 11px;
  color: var(--wf-el-text-secondary, #909399);
  background: var(--wf-el-fill-light, #f5f7fa);
  border-radius: 4px;
  padding: 1px 6px;
}
.var-pop-vars {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
/* ---------- 可用变量 chips（占位符插入） ---------- */
.var-chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  line-height: 1.6;
}
.var-chip {
  cursor: pointer;
  user-select: none;
}
.var-chip:hover {
  opacity: 0.8;
}
/* ---------- 星域 AI 节点：执行控制 ---------- */
.ai-control-buttons {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.ai-control-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border: 1px solid var(--wf-el-border, #dcdfe6);
  border-radius: 4px;
  background: var(--wf-el-fill-blank, #fff);
  color: var(--wf-el-text-regular, #606266);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}
.ai-control-btn:hover {
  border-color: var(--el-color-primary, #409eff);
  color: var(--el-color-primary, #409eff);
}
.ai-control-btn.pause.active {
  border-color: var(--el-color-warning, #e6a23c);
  background: rgba(230, 162, 60, 0.12);
  color: var(--el-color-warning, #e6a23c);
}
.ai-control-btn.stop:hover {
  border-color: var(--el-color-danger, #f56c6c);
  color: var(--el-color-danger, #f56c6c);
}
.ai-control-btn.continue:hover {
  border-color: var(--el-color-success, #67c23a);
  color: var(--el-color-success, #67c23a);
}
/* ---------- 星域 AI 节点：生成内容预览 ---------- */
.ai-output-box {
  border: 1px solid var(--wf-el-border-light, #ebeef5);
  border-radius: 4px;
  padding: 8px;
  background: var(--wf-el-fill-lighter, #fafafa);
}
.ai-output-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
  color: var(--wf-el-text-secondary, #909399);
  margin-bottom: 6px;
}
.ai-output-controls {
  display: flex;
  gap: 4px;
  margin-bottom: 6px;
  flex-wrap: wrap;
}
.ai-output-text,
.ai-output-render {
  margin: 0;
  max-height: 220px;
  overflow: auto;
  font-size: 12px;
  line-height: 1.5;
  color: var(--wf-el-text-primary, #303133);
  white-space: pre-wrap;
  word-break: break-word;
}
.ai-output-render :deep(h1),
.ai-output-render :deep(h2),
.ai-output-render :deep(h3) {
  font-size: 14px;
  margin: 6px 0;
}
.ai-output-render :deep(pre) {
  overflow-x: auto;
  background: var(--wf-el-fill, #f0f2f5);
  padding: 6px;
  border-radius: 3px;
}
.ai-output-empty {
  font-size: 12px;
  color: var(--wf-el-text-placeholder, #a8abb2);
  text-align: center;
  padding: 14px 0;
}
.dialog-render {
  max-height: 60vh;
}
.prompt-list-box {
  width: 100%;
}
.prompt-list-actions {
  display: flex;
  gap: 6px;
  align-items: center;
  margin-top: 6px;
}
.mt6 {
  margin-top: 6px;
}
.size-row {
  display: flex;
  gap: 8px;
  align-items: center;
}
.form-tip {
  margin-top: 4px;
  font-size: 12px;
  line-height: 1.6;
  color: var(--wf-el-text-secondary, #909399);
  width: 100%;
}
.io-alert {
  margin-bottom: 16px;
}
.io-line {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  line-height: 1.8;
}
.io-line code {
  padding: 2px 6px;
  background: var(--wf-el-fill-light, #f5f7fa);
  border-radius: 3px;
  font-family: Consolas, Monaco, monospace;
  font-size: 11px;
}
.var-rows {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}
.variable-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.variable-row.two {
  align-items: center;
}
.variable-row.two .el-input {
  flex: 1;
  min-width: 0;
}
.var-select {
  flex: 1;
  min-width: 0;
}
.add-btn {
  width: 100%;
}
.execution-history {
  margin-top: 24px;
}
.history-item {
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--wf-el-border-light, #ebeef5);
}
.history-item:last-child {
  border-bottom: none;
}
.history-item-title {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}
.history-time {
  font-size: 12px;
  color: var(--wf-el-text-secondary, #909399);
}
.error-message {
  margin-top: 8px;
  font-size: 12px;
  color: var(--el-color-danger, #f56c6c);
  background: rgba(245, 108, 108, 0.08);
  padding: 8px;
  border-radius: 4px;
}
.variable-list {
  margin-top: 8px;
  background: var(--wf-el-bg, #ffffff);
  border-radius: 4px;
  padding: 8px;
}
.variable-item-vertical {
  margin-bottom: 8px;
  font-size: 12px;
  line-height: 1.8;
  font-family: Consolas, Monaco, monospace;
}
.var-name {
  color: var(--el-color-primary, #409eff);
  font-weight: 600;
  display: block;
}
.var-value {
  color: var(--wf-el-text-regular, #606266);
  margin-left: 4px;
  word-break: break-word;
  white-space: pre-wrap;
}
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 40px 20px;
  text-align: center;
  color: var(--wf-el-text-secondary, #909399);
}
.empty-state i {
  font-size: 48px;
  color: var(--wf-el-text-placeholder, #c0c4cc);
  margin-bottom: 16px;
}
.empty-state p {
  font-size: 14px;
  color: var(--wf-el-text-secondary, #909399);
  margin-bottom: 0;
}
</style>

<style>
/* EP 自渲染的 alert 内部元素：只能放非 scoped 块才命中。
   项目只引了 EP 亮色变量，暗色下 el-alert 的 light-9 底色会亮得刺眼，这里按抽屉主题覆盖。 */
.wf-drawer .el-alert.io-alert {
  background: var(--wf-el-fill-light, #f5f7fa);
  border: 1px solid var(--wf-el-border-light, #ebeef5);
  padding: 10px 12px;
}
.wf-drawer .el-alert.io-alert .el-alert__content {
  color: var(--wf-el-text-regular, #606266);
}
.wf-drawer .el-alert.io-alert .el-alert__title {
  color: var(--wf-el-text-regular, #606266);
  font-size: 12px;
  line-height: 1.8;
}
.wf-drawer .el-alert.io-alert .el-alert__icon {
  color: var(--el-color-primary, #409eff);
}

/* ---------- 暗色模式补齐（C11） ----------
   项目未引 EP 暗色变量（dark/css-vars.css），--el-* 中性色在暗色下不翻转；
   全局 dark-theme.css 已覆盖 el-input/el-dialog/popper 等外壳，这里补抽屉内
   EP 内部件 + teleport 到 body 后解析不到 --wf-el-* 的弹层内容。 */
/* EP 表单标签 / 文字按钮 */
:root[data-theme='dark'] .wf-drawer .el-form-item__label {
  color: var(--wf-el-text-regular, #cfd3dc);
}
:root[data-theme='dark'] .wf-drawer .el-button.is-text {
  color: var(--wf-el-text-regular, #cfd3dc);
}
:root[data-theme='dark'] .wf-drawer .el-button.is-text:hover {
  color: var(--el-color-primary, #409eff);
  background: var(--wf-el-fill, #303030);
}
/* 分割线文字（content-position 左侧的白底文字胶囊） */
:root[data-theme='dark'] .wf-drawer .el-divider__text {
  background: var(--wf-el-bg, #2a2a2a);
  color: var(--wf-el-text-secondary, #a3a6ad);
}
/* 折叠面板（执行历史） */
:root[data-theme='dark'] .wf-drawer .el-collapse,
:root[data-theme='dark'] .wf-drawer .el-collapse-item__wrap {
  background: var(--wf-el-fill-blank, #1a1a1a);
  border-bottom-color: var(--wf-el-border-light, #363637);
}
:root[data-theme='dark'] .wf-drawer .el-collapse-item__header {
  background: var(--wf-el-fill-blank, #1a1a1a);
  color: var(--wf-el-text-primary, #e5eaf3);
  border-bottom-color: var(--wf-el-border-light, #363637);
}
/* 标签（变量 chip / 执行历史状态） */
:root[data-theme='dark'] .wf-drawer .el-tag {
  --el-tag-bg-color: rgba(64, 158, 255, 0.15);
  --el-tag-border-color: rgba(64, 158, 255, 0.25);
  --el-tag-text-color: #79bbff;
}
:root[data-theme='dark'] .wf-drawer .el-tag--success {
  --el-tag-bg-color: rgba(103, 194, 58, 0.15);
  --el-tag-border-color: rgba(103, 194, 58, 0.25);
  --el-tag-text-color: #95d475;
}
:root[data-theme='dark'] .wf-drawer .el-tag--danger {
  --el-tag-bg-color: rgba(245, 108, 108, 0.15);
  --el-tag-border-color: rgba(245, 108, 108, 0.25);
  --el-tag-text-color: #f89898;
}
:root[data-theme='dark'] .wf-drawer .el-tag--info {
  --el-tag-bg-color: rgba(144, 147, 153, 0.15);
  --el-tag-border-color: rgba(144, 147, 153, 0.25);
  --el-tag-text-color: #a3a6ad;
}
/* 颜色选择器下拉面板（teleport 到 body） */
:root[data-theme='dark'] .el-color-dropdown {
  background: #161b22;
  border: 1px solid rgba(59, 130, 246, 0.24);
}
/* 导入占位符 popover：teleport 到 body 后 --wf-el-* 解析不到，用显式暗色 */
:root[data-theme='dark'] .var-pop-title {
  color: #e5eaf3;
}
:root[data-theme='dark'] .var-pop-type {
  color: #a3a6ad;
  background: #303030;
}
:root[data-theme='dark'] .form-tip {
  color: #a3a6ad;
}
/* 生成内容预览弹窗（append-to-body）：正文文字与代码块 */
:root[data-theme='dark'] .el-dialog .ai-output-render {
  color: #e5eaf3;
}
:root[data-theme='dark'] .el-dialog .ai-output-render pre {
  background: #303030;
}
:root[data-theme='dark'] .el-dialog .io-count {
  color: #a3a6ad;
}
</style>
