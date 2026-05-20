<template>
  <div class="workflow-workbench-page">
    <template v-if="workflow">
      <header class="editor-toolbar">
        <div class="toolbar-main">
          <button type="button" class="toolbar-back" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </button>

          <div class="toolbar-title">
            <h1>{{ workflow.name }}</h1>
            <p>
              <span class="toolbar-canvas-label">工作流画布</span>
              <span class="toolbar-divider">·</span>
              <span>{{ workflow.description || '现在支持滚轮缩放和节点拖拽，鼠标停在画布上即可直接缩放。' }}</span>
            </p>
          </div>
        </div>

        <div class="toolbar-side">
          <div class="mode-switch-group">
            <button
              type="button"
              class="mode-switch-btn"
              :class="{ active: editorMode === 'visual' }"
              @click="editorMode = 'visual'"
            >
              <el-icon><View /></el-icon>
              <span>可视化</span>
            </button>
            <button
              type="button"
              class="mode-switch-btn"
              :class="{ active: editorMode === 'code' }"
              @click="editorMode = 'code'"
            >
              <el-icon><Edit /></el-icon>
              <span>源码</span>
            </button>
          </div>

          <div class="workspace-tools toolbar-workspace-tools">
            <div class="workspace-stats">
              <span>{{ stageNodes.length }} 节点</span>
              <span>{{ stageEdges.length }} 连线</span>
            </div>

            <div class="zoom-controls">
              <button type="button" class="zoom-btn" @click="adjustZoom(-0.1)">-</button>
              <button type="button" class="zoom-value-btn" @click="resetZoom">{{ zoomLabel }}</button>
              <button type="button" class="zoom-btn" @click="adjustZoom(0.1)">+</button>
            </div>
          </div>

          <div class="toolbar-actions">
            <el-tooltip
              v-for="action in toolbarActions"
              :key="action.key"
              :content="action.label"
              placement="bottom"
            >
              <button
                type="button"
                class="toolbar-icon-btn"
                :class="action.tone"
                @click="handleToolbarAction(action.key)"
              >
                <el-icon><component :is="action.icon" /></el-icon>
              </button>
            </el-tooltip>
          </div>
        </div>
      </header>

      <section v-if="editorMode === 'visual'" class="editor-layout" :style="editorLayoutStyle">
        <aside class="sidebar" :class="{ collapsed: leftSidebarCollapsed }">
          <div class="panel-toggle-wrap">
            <button type="button" class="panel-toggle-btn" @click="toggleLeftSidebar">
              <el-icon><component :is="leftSidebarCollapsed ? ArrowRight : ArrowLeft" /></el-icon>
            </button>
          </div>

          <template v-if="!leftSidebarCollapsed">
            <div class="sidebar-head">
              <h2>节点库</h2>
              <span>节点分类</span>
            </div>

            <div class="sidebar-list">
              <div
                v-for="group in groupedNodeLibrary"
                :key="group.category"
                class="library-group"
              >
                <div class="library-group-title">
                  <el-icon><Folder /></el-icon>
                  <span>{{ group.category }}</span>
                </div>
                <article
                  v-for="node in group.nodes"
                  :key="node.id"
                  class="library-item"
                  :style="{ '--accent': node.accent }"
                  draggable="true"
                  @dblclick="openLibraryNodePreview(node)"
                  @dragstart="handleLibraryDragStart($event, node.id)"
                  @dragend="handleLibraryDragEnd"
                >
                  <div class="library-item-title">
                    <span class="library-icon">
                      <el-icon><component :is="node.icon" /></el-icon>
                    </span>
                    <strong>{{ node.label }}</strong>
                  </div>
                </article>
              </div>
            </div>
          </template>
          <div v-else class="collapsed-rail-label">节点库</div>
        </aside>

        <main class="workspace">
          <div class="workspace-body" :style="workspaceBodyStyle">
            <div
              ref="workspaceScrollRef"
              class="workspace-scroll"
              :class="{ panning: isPanning, 'drop-ready': isLibraryDragging }"
              :style="gridStyle"
              @wheel="handleCanvasWheel"
              @pointerdown="startCanvasPan"
              @dragover.prevent="handleLibraryDragOver"
              @drop="handleLibraryDrop"
            >
              <div
                class="stage-camera"
                :style="{
                  transform: `translate(${camera.x}px, ${camera.y}px)`
                }"
              >
                <div
                  class="stage"
                  :style="{
                    width: `${stageBounds.width}px`,
                    height: `${stageBounds.height}px`,
                    transform: `scale(${zoom})`
                  }"
                >
                  <div class="stage-grid"></div>

                  <svg
                    class="edge-layer"
                    :width="stageBounds.width"
                    :height="stageBounds.height"
                    :viewBox="`0 0 ${stageBounds.width} ${stageBounds.height}`"
                  >
                    <defs>
                      <marker
                        id="workflow-arrow"
                        viewBox="0 0 10 10"
                        refX="9"
                        refY="5"
                        markerWidth="7"
                        markerHeight="7"
                        orient="auto-start-reverse"
                      >
                        <path d="M 0 0 L 10 5 L 0 10 z" fill="#7dd3fc" />
                      </marker>
                    </defs>

                    <g v-for="edge in stageEdges" :key="edge.id">
                      <path
                        :d="buildEdgePath(edge)"
                        class="edge-hit-path"
                        @contextmenu.prevent.stop="openStageEdgeContextMenu($event, edge.id)"
                      />
                      <path
                        :d="buildEdgePath(edge)"
                        class="edge-path"
                        :class="{ loop: edge.sourceHandle, running: runningEdgeIds.has(edge.id) }"
                        marker-end="url(#workflow-arrow)"
                      />
                    </g>

                    <path
                      v-if="draftEdgePath"
                      :d="draftEdgePath"
                      class="edge-path preview"
                    />
                  </svg>

                  <article
                    v-for="node in stageNodes"
                    :key="node.id"
                    class="stage-node"
                    :class="{
                      dragging: draggingNodeId === node.id,
                      selected: selectedNodeId === node.id,
                      'connect-ready': Boolean(edgeDragState) && edgeDragState?.source !== node.id,
                      'connect-hover': hoveredEdgeTargetId === node.id,
                      running: runningNodeId === node.id
                    }"
                    :style="getStageNodeStyle(node)"
                    @pointerdown="startNodeDrag($event, node.id)"
                    @pointerenter="handleEdgeTargetEnter(node.id)"
                    @pointerleave="handleEdgeTargetLeave(node.id)"
                    @click.stop="selectNode(node.id)"
                    @contextmenu.prevent.stop="openStageNodeContextMenu($event, node.id)"
                  >
                    <span
                      class="handle target"
                      @pointerdown.stop.prevent
                      @pointerup.stop="completeEdgeDrag(node.id)"
                    ></span>

                    <template v-if="node.type === 'loop'">
                      <span
                        class="handle source handle-main"
                        @pointerdown.stop.prevent="startEdgeDrag($event, node.id)"
                      ></span>
                      <span
                        class="handle source handle-loop-body"
                        @pointerdown.stop.prevent="startEdgeDrag($event, node.id, 'loop-body')"
                      ></span>
                      <span
                        class="handle source handle-loop-done"
                        @pointerdown.stop.prevent="startEdgeDrag($event, node.id, 'loop-done')"
                      ></span>
                    </template>
                    <template v-else>
                      <span
                        class="handle source"
                        @pointerdown.stop.prevent="startEdgeDrag($event, node.id)"
                      ></span>
                    </template>

                    <div class="stage-node-header">
                      <span class="stage-node-icon">
                        <el-icon><component :is="node.icon" /></el-icon>
                      </span>
                      <h3>{{ node.title }}</h3>
                    </div>

                    <p>{{ node.subtitle }}</p>

                    <div class="node-tags">
                      <span v-for="item in node.configPreview" :key="item">{{ item }}</span>
                    </div>

                    <span
                      class="resize-handle"
                      @pointerdown.stop.prevent="startNodeResize($event, node.id)"
                    ></span>
                  </article>
                </div>
              </div>
            </div>

            <div
              v-if="stageNodeContextMenu"
              class="stage-node-context-menu"
              :style="{
                left: `${stageNodeContextMenu.x}px`,
                top: `${stageNodeContextMenu.y}px`
              }"
            >
              <button type="button" class="stage-node-context-item" @click="copyStageNode(stageNodeContextMenu.nodeId)">
                复制
              </button>
              <button
                type="button"
                class="stage-node-context-item danger"
                @click="deleteStageNode(stageNodeContextMenu.nodeId)"
              >
                删除
              </button>
            </div>

            <div
              v-if="stageEdgeContextMenu"
              class="stage-node-context-menu stage-edge-context-menu"
              :style="{
                left: `${stageEdgeContextMenu.x}px`,
                top: `${stageEdgeContextMenu.y}px`
              }"
            >
              <button
                type="button"
                class="stage-node-context-item danger"
                @click="deleteStageEdge(stageEdgeContextMenu.edgeId)"
              >
                取消连线
              </button>
            </div>

            <aside class="inspector" :class="{ collapsed: rightInspectorCollapsed }">
              <div class="panel-toggle-wrap inspector-toggle-wrap">
                <button type="button" class="panel-toggle-btn" @click="toggleRightInspector">
                  <el-icon><component :is="rightInspectorCollapsed ? ArrowLeft : ArrowRight" /></el-icon>
                </button>
              </div>

              <template v-if="!rightInspectorCollapsed">
                <template v-if="selectedNode">
                  <div class="inspector-head">
                    <div class="inspector-title">
                      <span class="inspector-icon" :style="{ '--accent': selectedNode.accent }">
                        <el-icon><component :is="selectedNode.icon" /></el-icon>
                      </span>
                      <div class="inspector-title-meta">
                        <el-popover trigger="click" placement="bottom-start" :width="260">
                          <template #reference>
                            <button type="button" class="node-title-trigger inspector-title-trigger" aria-label="编辑节点名称">
                              <span>{{ selectedNode.title || '未命名节点' }}</span>
                            </button>
                          </template>
                          <div class="node-title-editor">
                            <span>节点名称</span>
                            <el-input v-model="selectedNode.title" placeholder="请输入节点名称" />
                          </div>
                        </el-popover>
                        <p>{{ selectedNode.id }}</p>
                      </div>
                    </div>
                    <el-tag size="small" effect="dark">{{ selectedNode.type }}</el-tag>
                  </div>

                  <template v-if="selectedStartNode">
                    <div class="inspector-section">
                      <span class="section-label">节点名称</span>
                      <el-popover trigger="click" placement="bottom-start" :width="260">
                        <template #reference>
                          <button type="button" class="node-title-trigger" aria-label="编辑节点名称">
                            <span>{{ selectedStartNode.title || '未命名节点' }}</span>
                          </button>
                        </template>
                        <div class="node-title-editor">
                          <span>节点名称</span>
                          <el-input v-model="selectedStartNode.title" placeholder="请输入节点名称" />
                        </div>
                      </el-popover>
                    </div>

                  </template>

                  <template v-else-if="selectedAiNode">
                    <div class="inspector-section">
                      <span class="section-label">节点名称</span>
                      <el-popover trigger="click" placement="bottom-start" :width="260">
                        <template #reference>
                          <button type="button" class="node-title-trigger" aria-label="编辑节点名称">
                            <span>{{ selectedAiNode.title || '未命名节点' }}</span>
                          </button>
                        </template>
                        <div class="node-title-editor">
                          <span>节点名称</span>
                          <el-input v-model="selectedAiNode.title" placeholder="请输入节点名称" />
                        </div>
                      </el-popover>
                    </div>

                    <div class="inspector-section">
                      <span class="section-label">AI模型选择</span>
                      <el-select
                        v-model="selectedAiNode.aiConfig!.modelId"
                        placeholder="请选择模型"
                        @change="handleAiModelChange(selectedAiNode, $event)"
                      >
                        <el-option
                          v-for="item in aiModelOptions"
                          :key="item.value"
                          :label="item.label"
                          :value="item.value"
                        />
                      </el-select>
                    </div>

                    <div class="inspector-section ai-compact-section ai-compact-params-anchor">
                      <span class="section-label">模型参数</span>
                      <div class="ai-config-pills">
                        <el-popover trigger="click" placement="bottom-start" :width="280">
                          <template #reference>
                            <button type="button" class="ai-config-pill" aria-label="调整模型温度">
                              <span>温度</span>
                              <strong>{{ Number(selectedAiNode.aiConfig!.temperature ?? 0).toFixed(1) }}</strong>
                            </button>
                          </template>
                          <div class="ai-config-editor ai-temp-editor">
                            <div class="ai-config-editor-head">
                              <span>模型温度</span>
                              <strong>{{ Number(selectedAiNode.aiConfig!.temperature ?? 0).toFixed(1) }}</strong>
                            </div>
                            <el-slider
                              v-model="selectedAiNode.aiConfig!.temperature"
                              :min="0"
                              :max="2"
                              :step="0.1"
                              show-input
                            />
                          </div>
                        </el-popover>

                        <el-popover trigger="click" placement="bottom-start" :width="220">
                          <template #reference>
                            <button type="button" class="ai-config-pill" aria-label="调整最大输出 Token">
                              <span>Token</span>
                              <strong>{{ selectedAiNode.aiConfig!.maxTokens }}</strong>
                            </button>
                          </template>
                          <div class="ai-config-editor">
                            <div class="ai-config-editor-head">
                              <span>最大输出 Token</span>
                              <strong>{{ selectedAiNode.aiConfig!.maxTokens }}</strong>
                            </div>
                            <el-input-number
                              v-model="selectedAiNode.aiConfig!.maxTokens"
                              :min="1"
                              :max="32000"
                              :step="256"
                              controls-position="right"
                            />
                          </div>
                        </el-popover>
                      </div>
                    </div>

                    <div class="inspector-section ai-prompt-editor-section" :class="{ collapsed: selectedAiNode?.aiConfig?.promptEditorCollapsed }">
                      <div class="section-label-row" @click="togglePromptEditorCollapsed(selectedAiNode)">
                        <span class="section-label">提示词编辑</span>
                        <el-icon class="collapse-icon"><ArrowDown /></el-icon>
                      </div>
                      <div v-show="!selectedAiNode?.aiConfig?.promptEditorCollapsed" class="inspector-field-stack ai-compact-stack">
                        <div class="ai-prompt-editor-shell">
                          <div class="ai-prompt-overlay" aria-hidden="true">
                            <div
                              class="ai-prompt-overlay-content"
                              :style="{
                                transform: `translate(${-aiPromptScroll.left}px, ${-aiPromptScroll.top}px)`
                              }"
                            >
                              <template v-if="selectedAiPromptDisplaySegments.length">
                                <template v-for="segment in selectedAiPromptDisplaySegments" :key="segment.key">
                                  <span
                                    v-if="segment.type === 'text'"
                                    class="ai-prompt-overlay-text"
                                  >{{ segment.text }}</span>
                                  <span
                                    v-else
                                    class="ai-prompt-inline-chip"
                                    :class="segment.tokenType === 'prompt' ? 'prompt' : 'field'"
                                  >
                                    <span>{{ segment.tokenType === 'prompt' ? '提示词' : '变量' }}</span>
                                    <strong>{{ segment.name }}</strong>
                                  </span>
                                </template>
                              </template>
                              <span v-else class="ai-prompt-overlay-placeholder">
                                请输入提示词，可插入 `[[workflow.字段名]]` 或 `[[workflow.提示词名]]`
                              </span>
                            </div>
                          </div>
                          <el-input
                            ref="aiPromptEditorRef"
                            v-model="selectedAiNode.aiConfig!.promptText"
                            type="textarea"
                            :rows="8"
                            placeholder="请输入提示词，可插入 [[workflow.字段名]] 或 [[workflow.提示词名]]"
                            @click="syncAiPromptSelection"
                            @keyup="syncAiPromptSelection"
                            @select="syncAiPromptSelection"
                          />
                        </div>
                        <div class="ai-prompt-action-row">
                          <button type="button" class="ai-prompt-action-btn" @click="openWorkflowPromptDialog">
                            插入提示词
                          </button>
                          <el-popover
                            placement="bottom"
                            trigger="click"
                            :width="320"
                            popper-class="workflow-reference-popover"
                          >
                            <template #reference>
                              <button type="button" class="ai-prompt-action-btn">插入占位符</button>
                            </template>
                            <div class="workflow-reference-menu">
                              <template v-if="selectedAiWorkflowReferenceGroups.length">
                                <button
                                  v-for="item in selectedAiWorkflowReferenceGroups.flatMap((g) => g.items)"
                                  :key="item.key"
                                  type="button"
                                  class="workflow-reference-item"
                                  @click="insertAiWorkflowReference(item.name)"
                                >
                                  {{ item.label }}
                                </button>
                              </template>
                              <p v-else class="workflow-reference-empty">
                                无可用字段
                              </p>
                            </div>
                          </el-popover>
                        </div>
                      </div>
                    </div>

                    <div class="inspector-section">
                      <span class="section-label">模型温度</span>
                      <el-slider
                        v-model="selectedAiNode.aiConfig!.temperature"
                        :min="0"
                        :max="2"
                        :step="0.1"
                        show-input
                      />
                    </div>

                    <div class="inspector-section">
                      <span class="section-label">模型最大输出token</span>
                      <el-input-number
                        v-model="selectedAiNode.aiConfig!.maxTokens"
                        :min="1"
                        :max="32000"
                        :step="256"
                        controls-position="right"
                      />
                    </div>


                    <div class="inspector-section">
                      <span class="section-label">额外输入</span>
                      <el-input
                        v-model="selectedAiNode.aiConfig!.extraInput"
                        placeholder="输入补充说明、变量提示或额外约束"
                      />
                    </div>

                    <div class="inspector-section">
                      <div class="section-label-row">
                        <span class="section-label">输出预览</span>
                        <div class="output-preview-actions">
                          <el-tooltip v-if="!selectedAiNode?.aiConfig?.stepMode" content="分步模式" placement="top">
                            <button type="button" class="output-action-btn" :class="{ active: selectedAiNode?.aiConfig?.stepMode }" @click="toggleStepMode(selectedAiNode)">
                              <el-icon><VideoPause /></el-icon>
                            </button>
                          </el-tooltip>
                          <el-tooltip v-else content="关闭分步" placement="top">
                            <button type="button" class="output-action-btn active" @click="toggleStepMode(selectedAiNode)">
                              <el-icon><VideoPlay /></el-icon>
                            </button>
                          </el-tooltip>
                          <el-tooltip v-if="waitingForContinue && waitingNodeId === selectedAiNode?.id" content="继续执行" placement="top">
                            <button type="button" class="output-action-btn continue-btn" @click="continueExecution">
                              <el-icon><VideoPlay /></el-icon>
                            </button>
                          </el-tooltip>
                          <el-tooltip content="停止生成" placement="top">
                            <button type="button" class="output-action-btn" @click="stopAiGeneration(selectedAiNode.id)" :disabled="!isGeneratingOutput(selectedAiNode.id)">
                              <el-icon><VideoPause /></el-icon>
                            </button>
                          </el-tooltip>
                          <el-tooltip content="清空内容" placement="top">
                            <button type="button" class="output-action-btn" @click="clearAiOutput(selectedAiNode.id)">
                              <el-icon><Delete /></el-icon>
                            </button>
                          </el-tooltip>
                          <el-tooltip content="放大预览" placement="top">
                            <button type="button" class="output-action-btn" @click="openOutputPreviewDialog">
                              <el-icon><ZoomIn /></el-icon>
                            </button>
                          </el-tooltip>
                        </div>
                      </div>
                      <div ref="aiOutputPreviewRef" class="ai-output-preview">
                        <div v-if="workflowExecutionContext[`ai_output_${selectedAiNode.id}`]" class="output-content">
                          <MarkdownRenderer :content="workflowExecutionContext[`ai_output_${selectedAiNode.id}`]" />
                        </div>
                        <div v-else class="output-placeholder">
                          <el-icon><Loading /></el-icon>
                          <span>运行工作流后显示输出结果</span>
                        </div>
                      </div>
                    </div>

                    <div class="inspector-section">
                      <div class="section-label-row">
                        <span class="section-label">输出字段</span>
                        <el-tooltip placement="top" effect="dark">
                          <template #content>
                            <div class="output-field-tooltip">
                              <p>用于下一个节点的占位符：</p>
                              <code>[[workflow.ai_output_{{ selectedAiNode.id.slice(0, 6) }}]]</code>
                            </div>
                          </template>
                          <el-icon class="section-help-icon"><QuestionFilled /></el-icon>
                        </el-tooltip>
                      </div>
                      <div class="context-fields-preview">
                        <div class="output-fields-list">
                          <div class="output-field-item">
                            <div class="field-name">
                              <el-icon><Connection /></el-icon>
                              <span class="field-id">ai_output_{{ selectedAiNode.id.slice(0, 6) }}</span>
                            </div>
                            <div class="field-value">
                              {{ workflowExecutionContext[`ai_output_${selectedAiNode.id}`] ? '已生成' : '(待生成)' }}
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </template>

                  <template v-else-if="selectedPromptSelectorNode">
                    <div class="inspector-section">
                      <span class="section-label">节点名称</span>
                      <el-popover trigger="click" placement="bottom-start" :width="260">
                        <template #reference>
                          <button type="button" class="node-title-trigger" aria-label="编辑节点名称">
                            <span>{{ selectedPromptSelectorNode.title || '未命名节点' }}</span>
                          </button>
                        </template>
                        <div class="node-title-editor">
                          <span>节点名称</span>
                          <el-input v-model="selectedPromptSelectorNode.title" placeholder="请输入节点名称" />
                        </div>
                      </el-popover>
                    </div>

                    <div class="inspector-section">
                      <span class="section-label">提示词选择</span>
                      <div class="inspector-field-stack ai-compact-stack">
                        <el-select
                          v-model="selectedPromptSelectorNode.promptConfig!.promptCategory"
                          placeholder="先选分类"
                          @change="handlePromptSelectorCategoryChange(selectedPromptSelectorNode, $event)"
                        >
                          <el-option
                            v-for="item in promptCategoryOptions"
                            :key="item.value"
                            :label="item.label"
                            :value="item.value"
                          />
                        </el-select>
                        <el-select
                          v-model="selectedPromptSelectorNode.promptConfig!.promptId"
                          placeholder="再选提示词"
                          @change="handlePromptSelectorPromptChange(selectedPromptSelectorNode, $event)"
                        >
                          <el-option
                            v-for="item in selectedPromptSelectorOptions"
                            :key="item.value"
                            :label="item.label"
                            :value="item.value"
                          />
                        </el-select>
                      </div>
                    </div>

                    <div class="inspector-section">
                      <span class="section-label">占位符输入</span>
                      <div v-if="selectedPromptSelectorFields.length" class="inspector-field-stack">
                        <div
                          v-for="field in selectedPromptSelectorFields"
                          :key="field.name"
                          class="prompt-field-editor"
                        >
                          <span class="prompt-field-label">
                            {{ field.label || field.name }}
                            <small v-if="field.required">必填</small>
                          </span>

                          <el-select
                            v-if="field.type === 'select'"
                            v-model="selectedPromptSelectorNode.promptConfig!.fieldValues![field.name]"
                            placeholder="请选择"
                          >
                            <el-option
                              v-for="option in field.options"
                              :key="option"
                              :label="option"
                              :value="option"
                            />
                          </el-select>

                          <el-input
                            v-else-if="field.type === 'textarea'"
                            v-model="selectedPromptSelectorNode.promptConfig!.fieldValues![field.name]"
                            type="textarea"
                            :rows="3"
                            :placeholder="field.description || `请输入${field.label || field.name}`"
                          />

                          <el-input
                            v-else
                            v-model="selectedPromptSelectorNode.promptConfig!.fieldValues![field.name]"
                            :placeholder="field.description || `请输入${field.label || field.name}`"
                          />

                          <p v-if="field.description" class="prompt-field-desc">{{ field.description }}</p>
                        </div>
                      </div>
                      <p v-else>当前提示词没有配置占位符字段。</p>
                    </div>

                    <div class="inspector-section">
                      <span class="section-label">节点说明</span>
                      <p>{{ selectedPromptSelectorNode.description }}</p>
                    </div>

                    <div class="inspector-section">
                      <span class="section-label">主要输出</span>
                      <div class="inspector-chip-list">
                        <span class="inspector-chip" v-for="output in selectedPromptSelectorNode.outputs" :key="output">
                          {{ output }}
                        </span>
                      </div>
                    </div>
                  </template>

                  <template v-else-if="selectedTextInputNode">
                    <div class="inspector-section">
                      <span class="section-label">节点名称</span>
                      <el-popover trigger="click" placement="bottom-start" :width="260">
                        <template #reference>
                          <button type="button" class="node-title-trigger" aria-label="编辑节点名称">
                            <span>{{ selectedTextInputNode.title || '未命名节点' }}</span>
                          </button>
                        </template>
                        <div class="node-title-editor">
                          <span>节点名称</span>
                          <el-input v-model="selectedTextInputNode.title" placeholder="请输入节点名称" />
                        </div>
                      </el-popover>
                    </div>

                    <div class="inspector-section">
                      <span class="section-label">节点说明</span>
                      <el-input
                        v-model="selectedTextInputNode.subtitle"
                        type="textarea"
                        :rows="2"
                        placeholder="请输入节点说明"
                      />
                    </div>

                    <div class="inspector-section">
                      <div class="section-label-row">
                        <span class="section-label">文本字段</span>
                        <el-dropdown trigger="click" @command="(type: 'text' | 'variable') => addTextInputField(selectedTextInputNode, type)">
                          <el-button size="small" type="primary" :icon="Plus">
                            添加字段
                          </el-button>
                          <template #dropdown>
                            <el-dropdown-menu>
                              <el-dropdown-item command="text">文本字段</el-dropdown-item>
                              <el-dropdown-item command="variable">变量引用</el-dropdown-item>
                            </el-dropdown-menu>
                          </template>
                        </el-dropdown>
                      </div>
                      
                      <div class="text-input-fields-list">
                        <div 
                          v-for="field in selectedTextInputNode.textInputConfig!.fields" 
                          :key="field.id" 
                          class="text-input-field-item"
                        >
                          <div class="field-header">
                            <el-input 
                              v-model="field.name" 
                              placeholder="字段名称" 
                              class="field-name-input"
                            />
                            <el-tag 
                              :type="field.type === 'text' ? 'primary' : 'success'" 
                              size="small"
                            >
                              {{ field.type === 'text' ? '文本' : '变量' }}
                            </el-tag>
                            <el-popover 
                              trigger="click" 
                              placement="left-start" 
                              :width="360"
                              :teleported="true"
                              popper-class="field-config-popper"
                            >
                              <template #reference>
                                <button type="button" class="field-config-btn">
                                  <el-icon><Setting /></el-icon>
                                </button>
                              </template>
                              <div class="field-config-popover">
                                <div class="field-config-row">
                                  <span>字段类型</span>
                                  <el-select v-model="field.type" size="small">
                                    <el-option label="文本" value="text" />
                                    <el-option label="变量引用" value="variable" />
                                  </el-select>
                                </div>
                                <div class="field-config-row">
                                  <span>占位符</span>
                                  <el-input v-model="field.placeholder" placeholder="占位符提示" size="small" />
                                </div>
                                <div class="field-config-row" v-if="field.type === 'text'">
                                  <span>文件导入</span>
                                  <div class="field-file-import">
                                    <el-upload
                                      :auto-upload="false"
                                      :limit="1"
                                      :on-change="(file: any) => handleTextInputFileImport(file, field.id)"
                                      accept=".txt,.md,.json"
                                    >
                                      <el-button size="small" :icon="Files">选择文件</el-button>
                                    </el-upload>
                                    <div v-if="field.importedFile" class="imported-file-info">
                                      <el-icon><Files /></el-icon>
                                      <span>{{ field.importedFile.name }}</span>
                                      <button
                                        type="button"
                                        class="clear-import-btn"
                                        @click="clearTextInputImportedFile(field.id)"
                                      >
                                        <el-icon><Close /></el-icon>
                                      </button>
                                    </div>
                                  </div>
                                </div>
                              </div>
                            </el-popover>
                            <button 
                              type="button" 
                              class="field-delete-btn"
                              @click="removeTextInputField(selectedTextInputNode, field.id)"
                            >
                              <el-icon><Delete /></el-icon>
                            </button>
                          </div>
                          <el-input
                            v-model="field.content"
                            type="textarea"
                            :rows="3"
                            :placeholder="field.placeholder || '请输入内容...'"
                            class="field-content-input"
                          />
                        </div>
                      </div>
                    </div>

                    <div class="inspector-section">
                      <div class="section-label-row">
                        <span class="section-label">输出字段</span>
                        <el-tooltip placement="top" effect="dark">
                          <template #content>
                            <div class="output-field-tooltip">
                              <p>用于下一个节点的占位符：</p>
                              <code>[[workflow.text_output_{{ selectedTextInputNode.id.slice(0, 6) }}]]</code>
                            </div>
                          </template>
                          <el-icon class="section-help-icon"><QuestionFilled /></el-icon>
                        </el-tooltip>
                      </div>
                      <div class="context-fields-preview">
                        <div class="output-fields-list">
                          <div 
                            v-for="field in selectedTextInputNode.textInputConfig!.fields" 
                            :key="field.id" 
                            class="output-field-item"
                          >
                            <div class="field-name">
                              <el-icon><Connection /></el-icon>
                              <span class="field-id">{{ field.name }}</span>
                            </div>
                            <div class="field-value">
                              {{ field.content ? `${field.content.length} 字` : '(待输入)' }}
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </template>

                  <div v-else class="inspector-section">
                    <span class="section-label">节点说明</span>
                    <p>{{ selectedNode.description }}</p>
                  </div>

                  <div v-if="!selectedAiNode && !selectedPromptSelectorNode" class="inspector-section">
                    <span class="section-label">主要输出</span>
                    <div class="inspector-chip-list">
                      <span class="inspector-chip" v-for="output in selectedNode.outputs" :key="output">
                        {{ output }}
                      </span>
                    </div>
                  </div>

                  <div v-if="!selectedAiNode && !selectedPromptSelectorNode" class="inspector-section">
                    <div class="inspector-section-head">
                      <span class="section-label">配置摘要</span>
                      <el-popover trigger="hover" placement="top-start" :width="320">
                        <template #reference>
                          <button type="button" class="inspector-help-btn" aria-label="配置摘要说明">
                            <el-icon><QuestionFilled /></el-icon>
                          </button>
                        </template>
                        <div class="inspector-help-details">
                          <div v-for="item in selectedNode.configDetails" :key="item.label" class="inspector-help-detail">
                            <span>{{ item.label }}</span>
                            <strong>{{ item.value }}</strong>
                          </div>
                        </div>
                      </el-popover>
                    </div>
                  </div>
                </template>

                <div v-else class="inspector-empty">
                  <el-icon><Connection /></el-icon>
                  <h3>选择一个节点</h3>
                  <p>点击画布中的节点，这里会显示对应的配置信息和输出说明。</p>
                </div>
              </template>
              <div v-else class="collapsed-rail-label">信息栏</div>
            </aside>
          </div>
        </main>
      </section>

      <section v-else class="code-editor-layout">
        <div class="code-editor-left">
          <div class="code-editor-container">
            <div class="code-editor-header">
              <h3>工作流源码</h3>
              <div class="code-editor-actions">
                <el-button size="small" @click="formatCodeSource">
                  <el-icon><Edit /></el-icon>
                  格式化
                </el-button>
                <el-button type="primary" size="small" @click="applyCodeSource">
                  <el-icon><Check /></el-icon>
                  应用更改
                </el-button>
              </div>
            </div>
            <div class="code-editor-body">
              <textarea
                v-model="codeSource"
                class="code-textarea"
                placeholder="在此编辑工作流JSON源码..."
                spellcheck="false"
              ></textarea>
            </div>
            <div v-if="codeError" class="code-error-bar">
              <el-icon><Close /></el-icon>
              <span>{{ codeError }}</span>
            </div>
          </div>
        </div>
        <div class="code-editor-right">
          <div class="workflow-generator-container">
            <div class="generator-header">
              <h3>工作流生成器</h3>
              <div class="generator-header-actions">
                <el-button type="primary" link size="small" @click="showWorkflowPromptEditor">
                  <el-icon><Setting /></el-icon>
                  修改提示词
                </el-button>
                <el-tag size="small" type="info">AI辅助</el-tag>
              </div>
            </div>
            <div class="generator-body">
              <div class="generator-prompt-section">
                <div class="prompt-label">
                  <el-icon><ChatDotRound /></el-icon>
                  <span>描述你想要的工作流</span>
                </div>
                <el-input
                  v-model="workflowGeneratorInput"
                  type="textarea"
                  :rows="6"
                  placeholder="例如：创建一个小说章节生成工作流，包含大纲输入、角色选择、AI生成和结果输出..."
                  class="generator-input"
                />
              </div>
              <div class="generator-options">
                <div class="option-item">
                  <span class="option-label">目标模型</span>
                  <el-select v-model="workflowGeneratorModel" placeholder="选择模型" size="small" style="width: 160px;">
                    <el-option
                      v-for="model in availableModels"
                      :key="model.id"
                      :label="model.name"
                      :value="model.id"
                    />
                  </el-select>
                </div>
              </div>
            </div>
            <div class="generator-footer">
              <el-button
                type="primary"
                :loading="workflowGenerating"
                :disabled="!workflowGeneratorInput.trim()"
                @click="generateWorkflowFromPrompt"
              >
                <el-icon><Lightning /></el-icon>
                生成工作流
              </el-button>
            </div>
          </div>
        </div>
      </section>

      <el-dialog
        v-model="workflowPromptEditorVisible"
        title="修改工作流生成器提示词"
        width="700px"
        align-center
        destroy-on-close
        class="workflow-prompt-editor-dialog"
      >
        <div class="prompt-editor-content">
          <div class="prompt-editor-hint">
            <el-alert
              title="提示词说明"
              type="info"
              :closable="false"
              show-icon
            >
              <template #default>
                <p>此提示词将作为系统指令发送给AI，指导AI生成符合规范的工作流JSON代码。</p>
                <p style="margin-top: 8px;">提示词中应包含：节点类型说明、JSON结构规范、生成规则等。</p>
              </template>
            </el-alert>
          </div>
          <div class="prompt-editor-body">
            <el-input
              v-model="editableWorkflowPrompt"
              type="textarea"
              :rows="18"
              placeholder="请输入提示词..."
              class="prompt-textarea"
            />
          </div>
        </div>
        <template #footer>
          <div class="prompt-editor-footer">
            <el-button @click="resetWorkflowPrompt">
              <el-icon><Refresh /></el-icon>
              恢复默认
            </el-button>
            <el-button @click="workflowPromptEditorVisible = false">取消</el-button>
            <el-button type="primary" @click="saveWorkflowPrompt">
              <el-icon><Check /></el-icon>
              保存
            </el-button>
          </div>
        </template>
      </el-dialog>

      <el-dialog
        v-model="libraryPreviewVisible"
        width="560px"
        align-center
        destroy-on-close
        class="library-preview-dialog"
      >
        <template #header>
          <div v-if="previewLibraryNode" class="inspector-head library-preview-head">
            <div class="inspector-title">
              <span class="inspector-icon" :style="{ '--accent': previewLibraryNode.accent }">
                <el-icon><component :is="previewLibraryNode.icon" /></el-icon>
              </span>
              <div>
                <h3>{{ previewLibraryTemplate?.title || previewLibraryNode.label }}</h3>
                <p>{{ previewLibraryNode.label }}</p>
              </div>
            </div>
            <el-tag size="small" effect="dark">{{ previewLibraryNode.type }}</el-tag>
          </div>
        </template>

        <div v-if="previewLibraryNode" class="library-preview-body">
          <template v-if="previewLibraryAiConfig">
            <div class="inspector-section library-preview-first">
              <span class="section-label">节点名称</span>
              <p>{{ previewLibraryTemplate?.title || previewLibraryNode.label }}</p>
            </div>

            <div class="inspector-section">
              <span class="section-label">节点说明</span>
              <p>{{ previewLibraryTemplate?.description || previewLibraryNode.description }}</p>
            </div>

            <div class="inspector-section">
              <span class="section-label">AI模型选择</span>
              <div class="config-list">
                <div class="config-item">
                  <span>当前模型</span>
                  <strong>{{ previewLibraryModelLabel }}</strong>
                </div>
              </div>
            </div>

            <div class="inspector-section">
              <span class="section-label">模型温度</span>
              <div class="config-list">
                <div class="config-item">
                  <span>temperature</span>
                  <strong>{{ previewLibraryAiConfig.temperature }}</strong>
                </div>
              </div>
            </div>

            <div class="inspector-section">
              <span class="section-label">模型最大输出 Token</span>
              <div class="config-list">
                <div class="config-item">
                  <span>maxTokens</span>
                  <strong>{{ previewLibraryAiConfig.maxTokens }}</strong>
                </div>
              </div>
            </div>


            <div class="inspector-section">
              <span class="section-label">额外输入</span>
              <p>{{ previewLibraryAiConfig.extraInput || '未设置' }}</p>
            </div>
          </template>

          <template v-else-if="previewLibraryPromptConfig">
            <div class="inspector-section library-preview-first">
              <span class="section-label">节点名称</span>
              <p>{{ previewLibraryTemplate?.title || previewLibraryNode.label }}</p>
            </div>

            <div class="inspector-section">
              <span class="section-label">提示词选择</span>
              <div class="config-list">
                <div class="config-item">
                  <span>分类</span>
                  <strong>{{ previewLibraryPromptConfig.promptCategory || '未设置' }}</strong>
                </div>
                <div class="config-item">
                  <span>提示词</span>
                  <strong>{{ getPromptLabel(previewLibraryPromptConfig.promptId) || '未设置' }}</strong>
                </div>
              </div>
            </div>

            <div class="inspector-section">
              <span class="section-label">占位符输入</span>
              <div v-if="previewLibraryPromptFields.length" class="config-list">
                <div v-for="field in previewLibraryPromptFields" :key="field.name" class="config-item">
                  <span>{{ field.label || field.name }}</span>
                  <strong>{{ previewLibraryPromptConfig.fieldValues?.[field.name] || '未填写' }}</strong>
                </div>
              </div>
              <p v-else>当前提示词没有配置占位符字段。</p>
            </div>

            <div class="inspector-section">
              <span class="section-label">节点说明</span>
              <p>{{ previewLibraryTemplate?.description || previewLibraryNode.description }}</p>
            </div>

            <div class="inspector-section">
              <span class="section-label">主要输出</span>
              <div class="inspector-chip-list">
                <span
                  v-for="output in previewLibraryTemplate?.outputs || []"
                  :key="`${previewLibraryNode.id}-${output}`"
                  class="inspector-chip"
                >
                  {{ output }}
                </span>
              </div>
            </div>
          </template>

          <template v-else>
            <div class="inspector-section library-preview-first">
              <span class="section-label">节点名称</span>
              <p>{{ previewLibraryTemplate?.title || previewLibraryNode.label }}</p>
            </div>

            <div class="inspector-section">
              <span class="section-label">节点说明</span>
              <p>{{ previewLibraryTemplate?.description || previewLibraryNode.description }}</p>
            </div>

            <div class="inspector-section" v-if="previewLibraryNode.type !== 'start' && (previewLibraryTemplate?.outputs?.length || 0)">
              <span class="section-label">主要输出</span>
              <div class="inspector-chip-list">
                <span
                  v-for="output in previewLibraryTemplate?.outputs || []"
                  :key="`${previewLibraryNode.id}-${output}`"
                  class="inspector-chip"
                >
                  {{ output }}
                </span>
              </div>
            </div>

            <div class="inspector-section" v-if="previewLibraryTemplate?.configPreview?.length">
              <span class="section-label">配置预览</span>
              <div class="inspector-chip-list">
                <span
                  v-for="item in previewLibraryTemplate?.configPreview || []"
                  :key="`${previewLibraryNode.id}-preview-${item}`"
                  class="inspector-chip"
                >
                  {{ item }}
                </span>
              </div>
            </div>

            <div class="inspector-section" v-if="previewLibraryTemplate?.configDetails?.length">
              <div class="inspector-section-head">
                <span class="section-label">配置摘要</span>
                <el-popover trigger="hover" placement="top-start" :width="320">
                  <template #reference>
                    <button type="button" class="inspector-help-btn" aria-label="配置摘要说明">
                      <el-icon><QuestionFilled /></el-icon>
                    </button>
                  </template>
                  <div class="inspector-help-details">
                    <div
                      v-for="item in previewLibraryTemplate?.configDetails || []"
                      :key="`${previewLibraryNode.id}-${item.label}`"
                      class="inspector-help-detail"
                    >
                      <span>{{ item.label }}</span>
                      <strong>{{ item.value }}</strong>
                    </div>
                  </div>
                </el-popover>
              </div>
            </div>
          </template>
        </div>
      </el-dialog>

      <el-dialog
        v-model="workflowPromptDialogVisible"
        width="720px"
        align-center
        destroy-on-close
        class="workflow-prompt-dialog"
        modal-class="workflow-prompt-dialog-modal"
      >
        <template #header>
          <div class="workflow-prompt-dialog-head">
            <h3>插入提示词引用</h3>
            <p>选择后会插入 `[[workflow.提示词名]]`，编辑区里会以压缩标签显示。</p>
          </div>
        </template>

        <div class="workflow-prompt-dialog-body">
          <div v-if="workflowPromptCategories.length" class="workflow-prompt-category-panel">
            <div class="workflow-prompt-category-head">
              <div>
                <strong>提示词分类</strong>
                <p>选择一个分类后，下方会只显示对应提示词。</p>
              </div>
              <span class="workflow-prompt-category-count">{{ workflowPromptCategories.length }} 个分类</span>
            </div>

            <div class="workflow-prompt-category-bar">
              <button
                v-for="category in workflowPromptCategories"
                :key="category"
                type="button"
                class="workflow-prompt-category-btn"
                :class="{ active: selectedWorkflowPromptCategory === category }"
                @click="selectedWorkflowPromptCategory = category"
              >
                <span class="workflow-prompt-category-dot"></span>
                <span>{{ category }}</span>
              </button>
            </div>
          </div>

          <div v-if="filteredWorkflowPromptOptions.length" class="workflow-prompt-list">
            <button
              v-for="prompt in filteredWorkflowPromptOptions"
              :key="prompt.id"
              type="button"
              class="workflow-prompt-item"
              @click="handleInsertWorkflowPrompt(prompt)"
            >
              <div class="workflow-prompt-item-head">
                <strong>{{ prompt.name }}</strong>
                <span>{{ prompt.category || '未分类' }}</span>
              </div>
              <p>{{ prompt.content?.slice(0, 120) }}{{ (prompt.content?.length || 0) > 120 ? '...' : '' }}</p>
            </button>
          </div>

          <div v-else class="workflow-prompt-empty">
            当前分类下还没有提示词。
          </div>
        </div>
      </el-dialog>

      <el-dialog
        v-model="workflowLogsDialogVisible"
        width="800px"
        align-center
        destroy-on-close
        class="workflow-logs-dialog"
        modal-class="workflow-logs-dialog-modal"
      >
        <template #header>
          <div class="workflow-logs-dialog-head">
            <h3>API 调用记录</h3>
            <div class="logs-actions">
              <span class="logs-count">共 {{ workflowLogs.length }} 条记录</span>
              <button type="button" class="clear-logs-btn" @click="clearWorkflowLogs" :disabled="workflowLogs.length === 0">
                <el-icon><Delete /></el-icon>
                清空
              </button>
            </div>
          </div>
        </template>
        <div class="workflow-logs-dialog-body">
          <div v-if="!workflowLogs.length" class="workflow-logs-empty">
            <el-icon><Tickets /></el-icon>
            <p>暂无 API 调用记录</p>
            <p class="hint">运行工作流后，AI API 调用记录会显示在这里</p>
          </div>
          <div v-else class="workflow-logs-list">
            <div
              v-for="log in workflowLogs"
              :key="log.id"
              class="workflow-log-item"
              :class="log.level"
              @click="viewLogDetail(log)"
            >
              <div class="log-item-header">
                <span class="log-node-title">{{ log.nodeTitle || '系统' }}</span>
                <span class="log-time">{{ log.timestamp }}</span>
              </div>
              <div class="log-item-body">
                <div class="log-meta">
                  <el-tag v-if="log.modelName" size="small" type="info">{{ log.modelName }}</el-tag>
                  <el-tag :type="log.level === 'success' ? 'success' : log.level === 'error' ? 'danger' : 'info'" size="small">
                    {{ log.level === 'success' ? '成功' : log.level === 'error' ? '失败' : log.level === 'warning' ? '警告' : '信息' }}
                  </el-tag>
                  <span v-if="log.inputTokens || log.outputTokens" class="log-tokens">
                    输入: {{ log.inputTokens || 0 }} / 输出: {{ log.outputTokens || 0 }} tokens
                  </span>
                </div>
                <div class="log-preview">
                  {{ log.outputContent?.substring(0, 100) || log.message }}{{ (log.outputContent?.length || 0) > 100 ? '...' : '' }}
                </div>
              </div>
              <div class="log-item-actions">
                <el-button size="small" type="primary" @click.stop="viewLogDetail(log)">
                  <el-icon><View /></el-icon>
                  查看详情
                </el-button>
              </div>
            </div>
          </div>
        </div>
        <template #footer>
          <el-button type="primary" @click="workflowLogsDialogVisible = false">关闭</el-button>
        </template>
      </el-dialog>

      <el-dialog
        v-model="logDetailDialogVisible"
        :title="selectedLogRecord?.nodeTitle || '调用详情'"
        width="900px"
        class="log-detail-dialog"
        modal-class="log-detail-dialog-modal"
        destroy-on-close
      >
        <div class="log-detail-content" v-if="selectedLogRecord">
          <div class="log-detail-meta">
            <div class="meta-item">
              <span class="meta-label">节点名称</span>
              <span class="meta-value">{{ selectedLogRecord.nodeTitle || '系统' }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">调用时间</span>
              <span class="meta-value">{{ selectedLogRecord.timestamp }}</span>
            </div>
            <div v-if="selectedLogRecord.modelName" class="meta-item">
              <span class="meta-label">使用模型</span>
              <span class="meta-value">{{ selectedLogRecord.modelName }}</span>
            </div>
            <div v-if="selectedLogRecord.inputTokens || selectedLogRecord.outputTokens" class="meta-item">
              <span class="meta-label">Token 消耗</span>
              <span class="meta-value">输入: {{ selectedLogRecord.inputTokens || 0 }} / 输出: {{ selectedLogRecord.outputTokens || 0 }}</span>
            </div>
          </div>
          <div class="conversation-messages">
            <div v-if="selectedLogRecord.promptContent" class="message-bubble user">
              <div class="message-sender">
                <el-icon class="sender-icon"><User /></el-icon>
                <span class="sender-name">提示词</span>
              </div>
              <div class="message-content">
                <pre>{{ selectedLogRecord.promptContent }}</pre>
              </div>
            </div>
            <div v-if="selectedLogRecord.outputContent" class="message-bubble assistant">
              <div class="message-sender">
                <el-icon class="sender-icon"><ChatDotRound /></el-icon>
                <span class="sender-name">AI 输出</span>
              </div>
              <div class="message-content">
                <MarkdownRenderer :content="selectedLogRecord.outputContent" />
              </div>
            </div>
            <div v-if="!selectedLogRecord.promptContent && !selectedLogRecord.outputContent" class="message-bubble system">
              <div class="message-sender">
                <el-icon class="sender-icon"><Setting /></el-icon>
                <span class="sender-name">系统</span>
              </div>
              <div class="message-content">
                <pre>{{ selectedLogRecord.message }}</pre>
              </div>
            </div>
          </div>
        </div>
        <template #footer>
          <el-button @click="logDetailDialogVisible = false">关闭</el-button>
        </template>
      </el-dialog>

      <el-dialog
        v-model="outputPreviewDialogVisible"
        width="800px"
        align-center
        destroy-on-close
        class="output-preview-dialog"
        modal-class="output-preview-dialog-modal"
      >
        <template #header>
          <div class="output-preview-dialog-head">
            <h3>输出预览</h3>
            <div class="output-preview-dialog-actions">
              <button type="button" class="output-action-btn" @click="toggleOutputEditMode" :disabled="!canEditOutput">
                <el-icon><Edit /></el-icon>
                {{ isOutputEditMode ? '完成' : '编辑' }}
              </button>
              <button type="button" class="output-action-btn" @click="selectedAiNode && clearAiOutput(selectedAiNode.id)">
                <el-icon><Delete /></el-icon>
                清空
              </button>
            </div>
          </div>
        </template>
        <div class="output-preview-dialog-body">
          <template v-if="selectedAiNode">
            <div v-if="isOutputEditMode" class="output-edit-area">
              <el-input
                v-model="workflowExecutionContext[`ai_output_${selectedAiNode.id}`]"
                type="textarea"
                :rows="16"
                placeholder="编辑输出内容..."
                class="output-editor-textarea"
              />
            </div>
            <div v-else-if="workflowExecutionContext[`ai_output_${selectedAiNode.id}`]" class="output-preview-content">
              <MarkdownRenderer :content="workflowExecutionContext[`ai_output_${selectedAiNode.id}`]" />
            </div>
            <div v-else class="output-preview-empty">
              <el-icon><Tickets /></el-icon>
              <p>暂无输出内容</p>
            </div>
          </template>
        </div>
      </el-dialog>

    </template>

    <div v-else class="missing-state">
      <el-result
        icon="warning"
        title="未找到该工作流"
        sub-title="这个工作流可能已被删除，或者当前浏览器里还没有对应数据。"
      >
        <template #extra>
          <el-button type="primary" @click="goBack">返回工作流列表</el-button>
        </template>
      </el-result>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  ArrowDown,
  ArrowLeft,
  ArrowRight,
  ArrowUp,
  ChatDotRound,
  Check,
  CircleCheckFilled,
  Close,
  Coin,
  Connection,
  CopyDocument,
  Delete,
  DocumentAdd,
  Edit,
  Files,
  Flag,
  Folder,
  Lightning,
  Plus,
  QuestionFilled,
  Refresh,
  Setting,
  Tickets,
  User,
  VideoPause,
  VideoPlay,
  View,
  ZoomIn
} from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import { aiAPI, configAPI, promptAPI } from '@/api'
import type { ApiModel, Prompt } from '@/types'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import {
  createWorkflow,
  deleteWorkflow,
  getWorkflowById,
  getWorkflowSceneById,
  loadWorkflowNodeLibrary,
  saveWorkflowNodeLibrary,
  saveWorkflowScene,
  type WorkflowNodeLibraryTemplate,
  type WorkflowSceneAIConfig,
  type WorkflowSceneEdge,
  type WorkflowSceneFieldConfig,
  type WorkflowSceneNode,
  type WorkflowScenePromptConfig,
  type WorkflowSceneStartConfig,
  type WorkflowSceneTextInputConfig,
  type WorkflowItem
} from '@/utils/workflow'

interface ToolbarAction {
  key: string
  label: string
  icon: object
  tone: 'neutral' | 'success' | 'danger'
}

type WorkflowNodeType = 'start' | 'aiProcess' | 'loop' | 'text' | 'end' | 'textInput'

interface NodeLibraryItem {
  id: string
  type: WorkflowNodeType
  variant?: string
  category: string
  label: string
  description: string
  outputs: string
  accent: string
  icon: object
  template: StageNodeDraft
  generated?: boolean
}

interface StageNode {
  id: string
  type: WorkflowNodeType
  variant?: string
  title: string
  subtitle: string
  description: string
  outputs: string[]
  configPreview: string[]
  configDetails: Array<{
    label: string
    value: string
  }>
  position: {
    x: number
    y: number
  }
  width: number
  height: number
  accent: string
  icon: object
  startConfig?: WorkflowSceneStartConfig
  aiConfig?: AIProcessConfig
  promptConfig?: PromptSelectorConfig
  textInputConfig?: TextInputConfig
}

type AIProcessConfig = WorkflowSceneAIConfig
type PromptSelectorConfig = WorkflowScenePromptConfig
type PromptField = NonNullable<Prompt['fields']>[number]

type StageEdge = WorkflowSceneEdge

interface AiPromptDisplaySegment {
  key: string
  type: 'text' | 'token'
  text: string
  name?: string
  tokenType?: 'field' | 'prompt'
}

interface AiWorkflowReferenceItem {
  key: string
  name: string
  label: string
  token: string
}

interface AiWorkflowReferenceGroup {
  key: string
  label: string
  typeLabel: string
  items: AiWorkflowReferenceItem[]
}

interface DragState {
  nodeId: string
  pointerStartX: number
  pointerStartY: number
  originX: number
  originY: number
  currentX: number
  currentY: number
}

interface ResizeState {
  nodeId: string
  pointerStartX: number
  pointerStartY: number
  originWidth: number
  originHeight: number
  currentWidth: number
  currentHeight: number
}

interface PanState {
  pointerStartX: number
  pointerStartY: number
  originX: number
  originY: number
}

interface EdgeDragState {
  source: string
  sourceHandle?: StageEdge['sourceHandle']
  pointerX: number
  pointerY: number
}

interface CanvasPoint {
  x: number
  y: number
}

interface StageNodeContextMenuState {
  nodeId: string
  x: number
  y: number
}

interface StageEdgeContextMenuState {
  edgeId: string
  x: number
  y: number
}

interface StageNodeDraft {
  title: string
  subtitle: string
  description: string
  outputs: string[]
  configPreview: string[]
  configDetails: Array<{
    label: string
    value: string
  }>
  width: number
  height: number
  startConfig?: WorkflowSceneStartConfig
  aiConfig?: AIProcessConfig
  promptConfig?: PromptSelectorConfig
  textInputConfig?: TextInputConfig
}

interface TextInputConfig {
  fields: TextInputField[]
}

interface TextInputField {
  id: string
  name: string
  type: 'text' | 'variable'
  content: string
  placeholder: string
  importedFile?: {
    name: string
    type: 'txt' | 'md' | 'json'
    content: string
  }
}

type TextInputStageNode = StageNode & {
  type: 'textInput'
  textInputConfig: TextInputConfig
}

const route = useRoute()
const router = useRouter()

const baseStageSize = {
  width: 3600,
  height: 2400
}
const STAGE_PADDING = 420

const MIN_ZOOM = 0.2
const MAX_ZOOM = 1.6

const toolbarActions: ToolbarAction[] = [
  { key: 'save', label: '保存', icon: Files, tone: 'neutral' },
  { key: 'run', label: '运行', icon: VideoPlay, tone: 'success' },
  { key: 'stop', label: '停止', icon: VideoPause, tone: 'danger' },
  { key: 'add-node', label: '添加节点', icon: DocumentAdd, tone: 'neutral' },
  { key: 'logs', label: '日志', icon: Tickets, tone: 'neutral' },
  { key: 'settings', label: '设置', icon: Setting, tone: 'neutral' }
]

type EditorMode = 'visual' | 'code'
const editorMode = ref<EditorMode>('visual')

const availableModels = ref<ApiModel[]>([])
const availablePrompts = ref<Prompt[]>([])
const customNodeLibrary = ref<WorkflowNodeLibraryTemplate[]>([])

const loadStoredPromptCategories = () => {
  const stored = localStorage.getItem('prompt_categories')

  if (!stored) {
    return []
  }

  try {
    const parsed = JSON.parse(stored)
    return Array.isArray(parsed) ? parsed.map(item => String(item).trim()).filter(Boolean) : []
  } catch {
    return []
  }
}

const NODE_TYPE_VISUALS: Record<
  WorkflowNodeType,
  {
    label: string
    description: string
    outputs: string
    accent: string
    icon: object
  }
> = {
  start: {
    label: '开始节点',
    description: '初始化变量并作为流程起点。',
    outputs: 'initialVariables',
    accent: '#22c55e',
    icon: Flag
  },
  aiProcess: {
    label: 'AI处理节点',
    description: '调用模型处理上下文并输出结果。',
    outputs: 'aiResult, lastAIOutput',
    accent: '#3b82f6',
    icon: ChatDotRound
  },
  loop: {
    label: '循环节点',
    description: '支持数组循环和固定次数循环。',
    outputs: 'currentIteration, currentItem, loopResult',
    accent: '#8b5cf6',
    icon: Refresh
  },
  text: {
    label: '文本节点',
    description: '输出固定文本或模板文本内容。',
    outputs: 'textResult',
    accent: '#ec4899',
    icon: Files
  },
  end: {
    label: '结束节点',
    description: '定义输出格式并结束流程。',
    outputs: 'outputFile',
    accent: '#ef4444',
    icon: CircleCheckFilled
  },
  textInput: {
    label: '文本输入节点',
    description: '手动输入文本内容或从文件导入文本。',
    outputs: 'textResult',
    accent: '#14b8a6',
    icon: Edit
  }
}

const legacyNodeLibrary = [
  {
    type: 'start',
    label: '开始节点',
    description: '初始化变量并作为流程起点。',
    outputs: 'initialVariables',
    accent: '#22c55e',
    icon: Flag
  },
  {
    type: 'aiProcess',
    label: 'AI处理节点',
    description: '调用模型处理上下文并输出结果。',
    outputs: 'aiResult, lastAIOutput',
    accent: '#3b82f6',
    icon: ChatDotRound
  },
  {
    type: 'loop',
    label: '循环节点',
    description: '支持数组循环和固定次数循环。',
    outputs: 'currentIteration, currentItem, loopResult',
    accent: '#8b5cf6',
    icon: Refresh
  },
  {
    type: 'text',
    label: '文本节点',
    description: '输出固定文本或模板文本内容。',
    outputs: 'textResult',
    accent: '#ec4899',
    icon: Files
  },
  {
    type: 'end',
    label: '结束节点',
    description: '定义输出格式并结束流程。',
    outputs: 'outputFile',
    accent: '#ef4444',
    icon: CircleCheckFilled
  },
  {
    type: 'textInput',
    label: '文本输入节点',
    description: '手动输入文本内容或从文件导入文本。',
    outputs: 'textResult',
    accent: '#14b8a6',
    icon: Edit
  }
]

const workflowId = computed(() => Number(route.params.id))

const workflow = ref<WorkflowItem | null>(null)

const loadWorkflow = async () => {
  if (Number.isNaN(workflowId.value)) {
    workflow.value = null
    return
  }
  workflow.value = await getWorkflowById(workflowId.value)
}

const createInitialStageNodes = (): StageNode[] => []

const stageNodes = ref<StageNode[]>(createInitialStageNodes())
const zoom = ref(1)
const camera = ref({
  x: 220,
  y: 120
})
const workspaceScrollRef = ref<HTMLDivElement | null>(null)
const dragState = ref<DragState | null>(null)
const resizeState = ref<ResizeState | null>(null)
const panState = ref<PanState | null>(null)
const edgeDragState = ref<EdgeDragState | null>(null)
const libraryDragPoint = ref<CanvasPoint | null>(null)
const hoveredEdgeTargetId = ref('')
const selectedNodeId = ref('')
const leftSidebarCollapsed = ref(false)
const rightInspectorCollapsed = ref(false)
const draggingLibraryId = ref('')
const libraryPreviewVisible = ref(false)
const previewLibraryNode = ref<NodeLibraryItem | null>(null)
const stageNodeContextMenu = ref<StageNodeContextMenuState | null>(null)
const stageEdgeContextMenu = ref<StageEdgeContextMenuState | null>(null)
const aiPromptEditorRef = ref<any>(null)
const aiOutputPreviewRef = ref<HTMLElement | null>(null)
const aiPromptTextareaScrollHandler = ref<(() => void) | null>(null)
const aiPromptSelection = ref({
  start: 0,
  end: 0
})
const aiPromptScroll = ref({
  top: 0,
  left: 0
})
const workflowPromptDialogVisible = ref(false)
const selectedWorkflowPromptCategory = ref('')
const dragPreviewPosition = ref<CanvasPoint | null>(null)

const workflowLogsDialogVisible = ref(false)
const workflowLogs = ref<Array<{
  id: string
  timestamp: string
  level: 'info' | 'success' | 'error' | 'warning'
  message: string
  nodeId?: string
  nodeTitle?: string
  modelName?: string
  promptContent?: string
  outputContent?: string
  inputTokens?: number
  outputTokens?: number
}>>([])
const selectedLogRecord = ref<typeof workflowLogs.value[0] | null>(null)
const logDetailDialogVisible = ref(false)
const isWorkflowRunning = ref(false)
const runningNodeId = ref<string | null>(null)
const runningEdgeIds = ref<Set<string>>(new Set())
const workflowExecutionContext = ref<Record<string, any>>({})
const workflowFinalOutput = ref<string | null>(null)
const aiGenerationControllers = ref<Map<string, AbortController>>(new Map())
const waitingForContinue = ref(false)
const waitingNodeId = ref<string | null>(null)
let continueResolver: (() => void) | null = null
let pendingDragPreviewPosition: CanvasPoint | null = null
let dragPreviewFrameId: number | null = null
const draggingNodeId = computed(() => dragState.value?.nodeId ?? '')
const isPanning = computed(() => panState.value !== null)
const isLibraryDragging = computed(() => draggingLibraryId.value !== '')
const isCustomLibraryNode = (node: NodeLibraryItem) => Boolean(node.generated)
const isPromptSelectorNode = (node: { type: string; variant?: string } | null | undefined) =>
  node?.type === 'text' && node.variant === 'promptSelector'
const isTextInputNode = (node: StageNode | null | undefined): node is TextInputStageNode =>
  node?.type === 'textInput' && Boolean(node.textInputConfig)
const previewLibraryAiConfig = computed(() => {
  if (previewLibraryNode.value?.type !== 'aiProcess') {
    return null
  }

  return previewLibraryTemplate.value?.aiConfig ?? createDefaultAiConfig()
})
const previewLibraryModelLabel = computed(() => {
  if (!previewLibraryAiConfig.value) {
    return ''
  }

  const fallback = previewLibraryAiConfig.value.model?.trim() || '未设置'
  const modelId = previewLibraryAiConfig.value.modelId

  if (modelId === null) {
    return fallback
  }

  return availableModels.value.find(item => item.id === modelId)?.name ?? fallback
})
const previewLibraryPromptConfig = computed(() => {
  if (!isPromptSelectorNode(previewLibraryNode.value)) {
    return null
  }

  return previewLibraryTemplate.value?.promptConfig ?? createDefaultPromptSelectorConfig()
})
const previewLibraryPromptFields = computed(() => getPromptFields(previewLibraryPromptConfig.value?.promptId ?? null))
const selectedNode = computed(() => stageNodes.value.find(node => node.id === selectedNodeId.value) ?? null)
const selectedStartNode = computed(() => {
  if (selectedNode.value?.type !== 'start') {
    return null
  }

  return selectedNode.value
})
const selectedAiNode = computed(() => {
  if (selectedNode.value?.type !== 'aiProcess') {
    return null
  }

  return selectedNode.value
})
const workflowPromptCategories = computed(() =>
  Array.from(
    new Set(
      availablePrompts.value
        .map(prompt => String(prompt.category || '').trim())
        .filter(Boolean)
    )
  )
)
const workflowPromptNameSet = computed(
  () =>
    new Set(
      availablePrompts.value
        .map(prompt => String(prompt.name || '').trim())
        .filter(Boolean)
    )
)
const filteredWorkflowPromptOptions = computed(() => {
  const category = selectedWorkflowPromptCategory.value

  return availablePrompts.value.filter(prompt => {
    const promptCategory = String(prompt.category || '').trim()
    return category ? promptCategory === category : true
  })
})
const selectedAiIncomingNodes = computed(() => {
  if (!selectedAiNode.value) {
    return []
  }

  const sourceIds = Array.from(
    new Set(stageEdges.value.filter(edge => edge.target === selectedAiNode.value?.id).map(edge => edge.source))
  )

  return sourceIds
    .map(sourceId => stageNodes.value.find(node => node.id === sourceId) ?? null)
    .filter((node): node is StageNode => Boolean(node))
})
const selectedAiWorkflowReferenceGroups = computed<AiWorkflowReferenceGroup[]>(() =>
  selectedAiIncomingNodes.value
    .map(node => ({
      key: node.id,
      label: node.title || node.id,
      typeLabel: NODE_TYPE_VISUALS[node.type]?.label || node.type,
      items: getWorkflowReferenceItems(node)
    }))
    .filter(group => group.items.length > 0)
)
const selectedAiPromptDisplaySegments = computed(() =>
  parseAiPromptDisplaySegments(selectedAiNode.value?.aiConfig?.promptText ?? '')
)
const selectedPromptSelectorNode = computed(() => {
  if (!isPromptSelectorNode(selectedNode.value)) {
    return null
  }

  return selectedNode.value
})
const selectedTextInputNode = computed(() => {
  if (!isTextInputNode(selectedNode.value)) {
    return null
  }

  return selectedNode.value
})
const selectedPromptSelectorOptions = computed(() => {
  const category = selectedPromptSelectorNode.value?.promptConfig?.promptCategory ?? ''

  return availablePrompts.value
    .filter(prompt => String(prompt.category || '').trim() === category)
    .map(prompt => ({
      label: prompt.name,
      value: prompt.id
    }))
})
const selectedPromptSelectorFields = computed(() =>
  getPromptFields(selectedPromptSelectorNode.value?.promptConfig?.promptId ?? null)
)
const previewLibraryTemplate = computed<StageNodeDraft | null>(() => {
  if (!previewLibraryNode.value) {
    return null
  }

  return previewLibraryNode.value.template
})
const aiModelOptions = computed(() =>
  availableModels.value.map(model => ({
    label: model.provider_name ? `${model.name} · ${model.provider_name}` : model.name,
    value: model.id
  }))
)
const promptCategoryOptions = computed(() => {
  const names = new Set(loadStoredPromptCategories())

  availablePrompts.value.forEach(prompt => {
    const category = String(prompt.category || '').trim()
    if (category) {
      names.add(category)
    }
  })

  return Array.from(names).map(name => ({
    label: name,
    value: name
  }))
})
const stageBounds = computed(() => {
  if (stageNodes.value.length === 0) {
    return {
      minX: -STAGE_PADDING,
      minY: -STAGE_PADDING,
      maxX: baseStageSize.width,
      maxY: baseStageSize.height,
      width: baseStageSize.width + STAGE_PADDING,
      height: baseStageSize.height + STAGE_PADDING
    }
  }

  const minNodeX = Math.min(...stageNodes.value.map(node => node.position.x))
  const minNodeY = Math.min(...stageNodes.value.map(node => node.position.y))
  const maxNodeX = Math.max(...stageNodes.value.map(node => node.position.x + node.width))
  const maxNodeY = Math.max(...stageNodes.value.map(node => node.position.y + node.height))

  const minX = Math.min(-STAGE_PADDING, minNodeX - STAGE_PADDING)
  const minY = Math.min(-STAGE_PADDING, minNodeY - STAGE_PADDING)
  const maxX = Math.max(baseStageSize.width, maxNodeX + STAGE_PADDING)
  const maxY = Math.max(baseStageSize.height, maxNodeY + STAGE_PADDING)

  return {
    minX,
    minY,
    maxX,
    maxY,
    width: maxX - minX,
    height: maxY - minY
  }
})

watch(
  () => ({
    minX: stageBounds.value.minX,
    minY: stageBounds.value.minY
  }),
  (nextBounds, prevBounds) => {
    if (!prevBounds) {
      return
    }

    camera.value.x += (nextBounds.minX - prevBounds.minX) * zoom.value
    camera.value.y += (nextBounds.minY - prevBounds.minY) * zoom.value
  },
  {
    flush: 'sync'
  }
)

watch(
  () => selectedAiNode.value?.id ?? '',
  () => {
    if (!selectedAiNode.value) {
      return
    }

    normalizeAiConfig(selectedAiNode.value)
    syncAiNodePresentation(selectedAiNode.value)
  },
  {
    immediate: true
  }
)

watch(
  () =>
    selectedAiNode.value
      ? [
          selectedAiNode.value.aiConfig?.modelId ?? null,
          selectedAiNode.value.aiConfig?.temperature ?? null,
          selectedAiNode.value.aiConfig?.maxTokens ?? null,
          selectedAiNode.value.aiConfig?.promptText ?? '',
          selectedAiNode.value.aiConfig?.extraInput ?? ''
        ]
      : null,
  () => {
    if (!selectedAiNode.value) {
      return
    }

    syncAiNodePresentation(selectedAiNode.value)
  }
)

watch(
  () => selectedPromptSelectorNode.value?.id ?? '',
  () => {
    if (!selectedPromptSelectorNode.value) {
      return
    }

    normalizePromptSelectorConfig(selectedPromptSelectorNode.value)
    syncPromptSelectorPresentation(selectedPromptSelectorNode.value)
  },
  {
    immediate: true
  }
)

watch(
  () =>
    selectedPromptSelectorNode.value
      ? [
          selectedPromptSelectorNode.value.promptConfig?.promptCategory ?? '',
          selectedPromptSelectorNode.value.promptConfig?.promptId ?? null,
          JSON.stringify(selectedPromptSelectorNode.value.promptConfig?.fieldValues ?? {})
        ]
      : null,
  () => {
    if (!selectedPromptSelectorNode.value) {
      return
    }

    syncPromptSelectorPresentation(selectedPromptSelectorNode.value)
  }
)

watch(
  () => selectedTextInputNode.value?.id ?? '',
  () => {
    if (!selectedTextInputNode.value) {
      return
    }

    syncTextInputPresentation(selectedTextInputNode.value)
  },
  {
    immediate: true
  }
)

watch(
  () => selectedTextInputNode.value?.textInputConfig?.fields ?? [],
  () => {
    if (!selectedTextInputNode.value) {
      return
    }

    syncTextInputPresentation(selectedTextInputNode.value)
  },
  {
    deep: true
  }
)

watch(
  workflowPromptCategories,
  categories => {
    if (!categories.length) {
      selectedWorkflowPromptCategory.value = ''
      return
    }

    if (!categories.includes(selectedWorkflowPromptCategory.value)) {
      selectedWorkflowPromptCategory.value = categories[0]
    }
  },
  {
    immediate: true
  }
)

watch(
  () => selectedAiNode.value ? workflowExecutionContext.value[`ai_output_${selectedAiNode.value.id}`] : null,
  () => {
    if (aiOutputPreviewRef.value) {
      nextTick(() => {
        aiOutputPreviewRef.value?.scrollTo({
          top: aiOutputPreviewRef.value.scrollHeight,
          behavior: 'smooth'
        })
      })
    }
  }
)

const zoomLabel = computed(() => `${Math.round(zoom.value * 100)}%`)
const editorLayoutStyle = computed(() => ({
  gridTemplateColumns: `${leftSidebarCollapsed.value ? 48 : 248}px minmax(0, 1fr)`
}))
const workspaceBodyStyle = computed(() => ({
  gridTemplateColumns: `minmax(0, 1fr) ${rightInspectorCollapsed.value ? 48 : 288}px`
}))
const gridStyle = computed(() => ({
  backgroundPosition: `${camera.value.x}px ${camera.value.y}px`,
  backgroundSize: `${28 * zoom.value}px ${28 * zoom.value}px`
}))

const stageEdges = ref<StageEdge[]>([])

const codeSource = ref('')
const codeError = ref('')

const workflowGeneratorInput = ref('')
const workflowGeneratorModel = ref<number | string>('')
const workflowGenerating = ref(false)
const workflowPromptEditorVisible = ref(false)
const editableWorkflowPrompt = ref('')

const getPreferredModel = () => {
  return availableModels.value.find(model => Boolean(model.is_default)) ?? availableModels.value[0] ?? null
}

const getPromptCategories = () => {
  return promptCategoryOptions.value.map(option => option.value)
}

const getPromptOptionsByCategory = (category: string) => {
  return availablePrompts.value.filter(prompt => String(prompt.category || '').trim() === category)
}

const createDefaultAiConfig = (): AIProcessConfig => {
  const defaultModel = getPreferredModel()

  return {
    modelId: defaultModel?.id ?? null,
    model: defaultModel?.model ?? '',
    temperature: defaultModel?.temperature ?? 0.8,
    maxTokens: defaultModel?.max_tokens ?? 4096,
    promptCategory: '',
    promptId: null,
    promptText: '',
    extraInput: ''
  }
}

const createDefaultPromptSelectorConfig = (): PromptSelectorConfig => {
  const categories = getPromptCategories()
  const promptCategory = categories[0] ?? ''
  const promptId = getPromptOptionsByCategory(promptCategory)[0]?.id ?? null

  return {
    promptCategory,
    promptId,
    fieldValues: {}
  }
}

const createDefaultTextInputConfig = (): TextInputConfig => {
  return {
    fields: [
      {
        id: `field-${Date.now()}`,
        name: 'text_1',
        type: 'text',
        content: '',
        placeholder: '请输入文本内容...'
      }
    ]
  }
}

const generateFieldId = () => `field-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`

const cloneTextInputImportedFile = (file: TextInputField['importedFile']) =>
  file
    ? {
        name: file.name,
        type: file.type,
        content: file.content
      }
    : undefined

const cloneTextInputField = (field: TextInputField): TextInputField => ({
  id: field.id,
  name: field.name,
  type: field.type,
  content: field.content,
  placeholder: field.placeholder,
  importedFile: cloneTextInputImportedFile(field.importedFile)
})

const cloneTextInputConfig = (config: TextInputConfig): TextInputConfig => ({
  fields: config.fields.map(field => cloneTextInputField(field))
})

const normalizeTextInputConfig = (config?: WorkflowSceneTextInputConfig): TextInputConfig => {
  if (Array.isArray(config?.fields) && config.fields.length > 0) {
    return {
      fields: config.fields.map(field => ({
        id: field.id ?? generateFieldId(),
        name: field.name ?? 'text_1',
        type: field.type ?? 'text',
        content: field.content ?? '',
        placeholder: field.placeholder ?? '',
        importedFile: field.importedFile
          ? {
              name: field.importedFile.name,
              type: field.importedFile.type,
              content: field.importedFile.content
            }
          : undefined
      }))
    }
  }

  if (config && ('text' in config || 'placeholder' in config || 'importedFile' in config)) {
    return {
      fields: [
        {
          id: generateFieldId(),
          name: config.outputField || 'text_1',
          type: 'text',
          content: config.text ?? '',
          placeholder: config.placeholder ?? '',
          importedFile: config.importedFile
            ? {
                name: config.importedFile.name,
                type: config.importedFile.type,
                content: config.importedFile.content
              }
            : undefined
        }
      ]
    }
  }

  return createDefaultTextInputConfig()
}

const addTextInputField = (node: StageNode | null | undefined, type: 'text' | 'variable' = 'text') => {
  if (!isTextInputNode(node)) {
    return
  }

  const fieldCount = node.textInputConfig.fields.length
  const newField: TextInputField = {
    id: generateFieldId(),
    name: `text_${fieldCount + 1}`,
    type,
    content: '',
    placeholder: type === 'text' ? '请输入文本内容...' : '请输入变量引用...'
  }

  node.textInputConfig.fields.push(newField)
  syncTextInputPresentation(node)
}

const removeTextInputField = (node: StageNode | null | undefined, fieldId: string) => {
  if (!isTextInputNode(node)) {
    return
  }

  node.textInputConfig.fields = node.textInputConfig.fields.filter(f => f.id !== fieldId)
  if (node.textInputConfig.fields.length === 0) {
    addTextInputField(node, 'text')
  }
  syncTextInputPresentation(node)
}

const updateTextInputField = (node: StageNode | null | undefined, fieldId: string, updates: Partial<TextInputField>) => {
  if (!isTextInputNode(node)) {
    return
  }

  const field = node.textInputConfig.fields.find(f => f.id === fieldId)
  if (field) {
    Object.assign(field, updates)
    syncTextInputPresentation(node)
  }
}

const normalizeStartField = (field: Partial<WorkflowSceneFieldConfig>): WorkflowSceneFieldConfig => ({
  name: field.name ?? '',
  label: field.label ?? '',
  type: field.type ?? 'text',
  options: Array.isArray(field.options) ? [...field.options] : [],
  description: field.description ?? '',
  required: field.required ?? false
})

const getModelLabel = (modelId: number | null, fallback = '') => {
  if (modelId === null) {
    return fallback || '?????'
  }

  const model = availableModels.value.find(item => item.id === modelId)
  return model?.name ?? (fallback || '?????')
}

const getPromptLabel = (promptId: number | null) => {
  if (promptId === null) {
    return '??????'
  }

  return availablePrompts.value.find(item => item.id === promptId)?.name ?? '??????'
}

const getPromptById = (promptId: number | null) => {
  if (promptId === null) {
    return null
  }

  return availablePrompts.value.find(item => item.id === promptId) ?? null
}

const extractPromptPlaceholders = (content: string) => {
  const matches = content.matchAll(/\{\{\s*([^{}]+?)\s*\}\}/g)
  const uniqueNames = Array.from(
    new Set(
      Array.from(matches)
        .map(match => String(match[1] ?? '').trim())
        .filter(Boolean)
    )
  )

  return uniqueNames.map(
    name =>
      ({
        name,
        label: name,
        type: 'text',
        options: [],
        description: '',
        required: false
      }) satisfies PromptField
  )
}

const getPromptFields = (promptId: number | null): PromptField[] => {
  const prompt = getPromptById(promptId)

  if (!prompt) {
    return []
  }

  if (Array.isArray(prompt.fields) && prompt.fields.length > 0) {
    return prompt.fields
      .map(
        field =>
          ({
            name: String(field.name ?? '').trim(),
            label: String(field.label ?? field.name ?? '').trim(),
            type: field.type === 'textarea' || field.type === 'select' ? field.type : 'text',
            options: Array.isArray(field.options)
              ? field.options.map(option => String(option).trim()).filter(Boolean)
              : [],
            description: String(field.description ?? '').trim(),
            required: Boolean(field.required)
          }) satisfies PromptField
      )
      .filter(field => field.name)
  }

  return extractPromptPlaceholders(prompt.content)
}

const buildWorkflowReferenceToken = (name: string) => `[[workflow.${name}]]`

const getAiPromptTextarea = () => {
  const instance = aiPromptEditorRef.value

  if (!instance) {
    return null
  }

  if (instance.textarea instanceof HTMLTextAreaElement) {
    return instance.textarea
  }

  return instance.$el?.querySelector?.('textarea') ?? null
}

const syncAiPromptScroll = () => {
  const textarea = getAiPromptTextarea()

  aiPromptScroll.value = {
    top: textarea?.scrollTop ?? 0,
    left: textarea?.scrollLeft ?? 0
  }
}

const detachAiPromptTextareaScroll = () => {
  const textarea = getAiPromptTextarea()

  if (textarea && aiPromptTextareaScrollHandler.value) {
    textarea.removeEventListener('scroll', aiPromptTextareaScrollHandler.value)
  }

  aiPromptTextareaScrollHandler.value = null
}

const attachAiPromptTextareaScroll = () => {
  const textarea = getAiPromptTextarea()

  if (!textarea) {
    return
  }

  detachAiPromptTextareaScroll()
  const handler = () => syncAiPromptScroll()
  textarea.addEventListener('scroll', handler)
  aiPromptTextareaScrollHandler.value = handler
  syncAiPromptScroll()
}

const syncAiPromptSelection = () => {
  const textarea = getAiPromptTextarea()

  if (!textarea) {
    return
  }

  aiPromptSelection.value = {
    start: textarea.selectionStart ?? 0,
    end: textarea.selectionEnd ?? 0
  }
  syncAiPromptScroll()
}

const parseAiPromptDisplaySegments = (content: string): AiPromptDisplaySegment[] => {
  const text = String(content || '')
  const segments: AiPromptDisplaySegment[] = []
  const pattern = /\[\[\s*workflow\.([^[\]]+?)\s*\]\]/g
  let lastIndex = 0
  let tokenIndex = 0

  for (const match of text.matchAll(pattern)) {
    const rawToken = String(match[0] ?? '')
    const tokenName = String(match[1] ?? '').trim()
    const start = match.index ?? 0

    if (start > lastIndex) {
      segments.push({
        key: `text-${lastIndex}`,
        type: 'text',
        text: text.slice(lastIndex, start)
      })
    }

    segments.push({
      key: `token-${tokenIndex}`,
      type: 'token',
      text: rawToken,
      name: tokenName,
      tokenType: workflowPromptNameSet.value.has(tokenName) ? 'prompt' : 'field'
    })

    lastIndex = start + rawToken.length
    tokenIndex += 1
  }

  if (lastIndex < text.length) {
    segments.push({
      key: `text-${lastIndex}`,
      type: 'text',
      text: text.slice(lastIndex)
    })
  }

  return segments
}

const normalizeWorkflowReferenceName = (value: string) => String(value || '').trim().replace(/^\{\{|\}\}$/g, '')

const getWorkflowReferenceItems = (node: StageNode): AiWorkflowReferenceItem[] => {
  const seen = new Set<string>()
  const items: AiWorkflowReferenceItem[] = []
  const pushItem = (name: string, label = name) => {
    const normalizedName = normalizeWorkflowReferenceName(name)

    if (!normalizedName || seen.has(normalizedName)) {
      return
    }

    seen.add(normalizedName)
    items.push({
      key: `${node.id}-${normalizedName}`,
      name: normalizedName,
      label: label.trim() || normalizedName,
      token: buildWorkflowReferenceToken(normalizedName)
    })
  }

  if (node.type === 'start') {
    return items
  }

  if (node.type === 'aiProcess') {
    const outputFieldName = `ai_output_${node.id}`
    pushItem(outputFieldName, `AI输出 (${node.id.slice(0, 6)})`)
    return items
  }

  if (node.type === 'textInput' && node.textInputConfig?.fields) {
    node.textInputConfig.fields.forEach(field => {
      pushItem(field.name, `${field.name} (${field.type === 'text' ? '文本' : '变量'})`)
    })
    return items
  }

  node.outputs.forEach(output => pushItem(output))
  return items
}

const insertAiPromptTextAtCursor = async (insertText: string) => {
  if (!selectedAiNode.value?.aiConfig || !insertText.trim()) {
    return
  }

  const currentValue = selectedAiNode.value.aiConfig.promptText ?? ''
  const textarea = getAiPromptTextarea()
  const start = textarea?.selectionStart ?? aiPromptSelection.value.start
  const end = textarea?.selectionEnd ?? aiPromptSelection.value.end

  selectedAiNode.value.aiConfig.promptText =
    currentValue.slice(0, start) + insertText + currentValue.slice(end)
  syncAiNodePresentation(selectedAiNode.value)

  const cursor = start + insertText.length

  await nextTick()

  const nextTextarea = getAiPromptTextarea()
  if (nextTextarea) {
    nextTextarea.focus()
    nextTextarea.setSelectionRange(cursor, cursor)
  }

  aiPromptSelection.value = {
    start: cursor,
    end: cursor
  }
}

const insertAiWorkflowReference = async (name: string) => {
  const normalizedName = normalizeWorkflowReferenceName(name)

  if (!normalizedName) {
    return
  }

  await insertAiPromptTextAtCursor(buildWorkflowReferenceToken(normalizedName))
}

const openWorkflowPromptDialog = () => {
  if (!availablePrompts.value.length) {
    ElMessage.warning('当前还没有可选提示词')
    return
  }

  if (workflowPromptCategories.value.length && !workflowPromptCategories.value.includes(selectedWorkflowPromptCategory.value)) {
    selectedWorkflowPromptCategory.value = workflowPromptCategories.value[0] ?? ''
  }

  workflowPromptDialogVisible.value = true
}

const handleInsertWorkflowPrompt = async (prompt: Prompt) => {
  const promptName = String(prompt.name || '').trim()

  if (!promptName) {
    return
  }

  await insertAiWorkflowReference(promptName)
  workflowPromptDialogVisible.value = false
}

const outputPreviewDialogVisible = ref(false)
const isOutputEditMode = ref(false)

watch(
  () => selectedAiNode.value?.id ?? '',
  async () => {
    detachAiPromptTextareaScroll()

    if (!selectedAiNode.value) {
      aiPromptScroll.value = {
        top: 0,
        left: 0
      }
      return
    }

    await nextTick()
    attachAiPromptTextareaScroll()
    syncAiPromptSelection()
    isOutputEditMode.value = false
  },
  {
    immediate: true
  }
)

const clearAiOutput = (nodeId: string) => {
  const outputKey = `ai_output_${nodeId}`
  delete workflowExecutionContext.value[outputKey]
}

const isGeneratingOutput = (nodeId: string): boolean => {
  return aiGenerationControllers.value.has(nodeId)
}

const stopAiGeneration = (nodeId: string) => {
  const controller = aiGenerationControllers.value.get(nodeId)
  if (controller) {
    controller.abort()
    aiGenerationControllers.value.delete(nodeId)
    ElMessage.success('已停止生成')
  }
}

const toggleStepMode = (node: StageNode | null) => {
  if (!node?.aiConfig) return
  node.aiConfig.stepMode = !node.aiConfig.stepMode
}

const togglePromptEditorCollapsed = (node: StageNode | null) => {
  if (!node?.aiConfig) return
  node.aiConfig.promptEditorCollapsed = !node.aiConfig.promptEditorCollapsed
}

const continueExecution = () => {
  if (continueResolver) {
    continueResolver()
    ElMessage.success('继续执行工作流')
  }
}

const canEditOutput = computed(() => {
  if (!selectedAiNode.value) return false
  if (isWorkflowRunning.value && !waitingForContinue.value) return false
  return true
})

const toggleOutputEditMode = () => {
  isOutputEditMode.value = !isOutputEditMode.value
}

const openOutputPreviewDialog = () => {
  if (!selectedAiNode.value) {
    return
  }
  outputPreviewDialogVisible.value = true
}

const normalizeStartConfig = (node: StageNode) => {
  if (node.type !== 'start') {
    return
  }

  if (!node.startConfig) {
    node.startConfig = {
      fields: [],
      fieldValues: {}
    }
    return
  }

  if (!Array.isArray(node.startConfig.fields)) {
    node.startConfig.fields = []
  }

  if (!node.startConfig.fieldValues || typeof node.startConfig.fieldValues !== 'object') {
    node.startConfig.fieldValues = {}
  }

  node.startConfig.fields = node.startConfig.fields.map(field => normalizeStartField(field))

  const fieldNames = new Set(node.startConfig.fields.map(f => f.name))
  const nextValues: Record<string, string> = {}
  for (const fieldName of fieldNames) {
    nextValues[fieldName] = String(node.startConfig.fieldValues[fieldName] ?? '')
  }
  node.startConfig.fieldValues = nextValues
}

const syncStartNodePresentation = (node: StageNode) => {
  if (node.type !== 'start' || !node.startConfig) {
    return
  }

  const fields = node.startConfig.fields
  const fieldValues = node.startConfig.fieldValues ?? {}
  const filledCount = fields.filter(f => String(fieldValues[f.name] ?? '').trim()).length

  node.configPreview = fields.length > 0
    ? fields.slice(0, 3).map(f => `${f.label}: ${String(fieldValues[f.name] ?? '').slice(0, 10) || '未填写'}`)
    : ['未配置字段']

  node.configDetails = fields.map(f => ({
    label: f.label,
    value: String(fieldValues[f.name] ?? '') || '未填写'
  }))

  if (fields.length > 3) {
    node.configPreview.push(`... 还有 ${fields.length - 3} 个字段`)
  }

  node.subtitle = `共 ${fields.length} 个输入字段`
  node.description = `定义工作流的初始变量，当前已填写 ${filledCount}/${fields.length} 个字段`
}

const normalizePromptSelectorConfig = (node: StageNode) => {
  if (!isPromptSelectorNode(node)) {
    return
  }

  if (!node.promptConfig) {
    node.promptConfig = createDefaultPromptSelectorConfig()
  }

  const categories = getPromptCategories()
  if (!categories.includes(node.promptConfig.promptCategory)) {
    node.promptConfig.promptCategory = categories[0] ?? ''
  }

  const categoryPrompts = getPromptOptionsByCategory(node.promptConfig.promptCategory)
  if (!categoryPrompts.some(prompt => prompt.id === node.promptConfig?.promptId)) {
    node.promptConfig.promptId = categoryPrompts[0]?.id ?? null
  }

  const fields = getPromptFields(node.promptConfig.promptId)
  const nextValues = Object.fromEntries(
    fields.map(field => [field.name, String(node.promptConfig?.fieldValues?.[field.name] ?? '')])
  )
  node.promptConfig.fieldValues = nextValues
}

const normalizeAiConfig = (node: StageNode) => {
  if (node.type !== 'aiProcess') {
    return
  }

  if (!node.aiConfig) {
    node.aiConfig = createDefaultAiConfig()
    return
  }

  const config = node.aiConfig
  const selectedModel = availableModels.value.find(model => model.id === config.modelId)
  const fallbackModel = getPreferredModel()
  const legacyPromptContent = getPromptById(config.promptId)?.content ?? ''

  if (!selectedModel && fallbackModel) {
    config.modelId = fallbackModel.id
    config.model = fallbackModel.model
    config.temperature = fallbackModel.temperature
    config.maxTokens = fallbackModel.max_tokens
  } else if (selectedModel) {
    config.model = selectedModel.model
  }

  config.promptCategory = String(config.promptCategory ?? '')
  config.promptId = typeof config.promptId === 'number' ? config.promptId : null

  if (typeof config.promptText !== 'string') {
    config.promptText = ''
  }

  if (!config.promptText.trim() && legacyPromptContent.trim()) {
    config.promptText = legacyPromptContent
  }
}

const handleAiModelChange = (node: StageNode, modelId: number | null) => {
  if (node.type !== 'aiProcess' || !node.aiConfig) {
    return
  }

  const model = availableModels.value.find(item => item.id === modelId)
  node.aiConfig.modelId = model?.id ?? null
  node.aiConfig.model = model?.model ?? ''

  if (model) {
    node.aiConfig.temperature = model.temperature
    node.aiConfig.maxTokens = model.max_tokens
  }

  syncAiNodePresentation(node)
}

const handlePromptSelectorCategoryChange = (node: StageNode, category: string) => {
  if (!isPromptSelectorNode(node) || !node.promptConfig) {
    return
  }

  node.promptConfig.promptCategory = category
  node.promptConfig.promptId = getPromptOptionsByCategory(category)[0]?.id ?? null
  node.promptConfig.fieldValues = {}
  syncPromptSelectorPresentation(node)
}

const handlePromptSelectorPromptChange = (node: StageNode, promptId: number | null) => {
  if (!isPromptSelectorNode(node) || !node.promptConfig) {
    return
  }

  node.promptConfig.promptId = promptId
  normalizePromptSelectorConfig(node)
  syncPromptSelectorPresentation(node)
}

const syncAiNodePresentation = (node: StageNode) => {
  if (node.type !== 'aiProcess' || !node.aiConfig) {
    return
  }

  const modelLabel = getModelLabel(node.aiConfig.modelId, node.aiConfig.model)
  const promptText = String(node.aiConfig.promptText ?? '').trim()
  const promptPreview = promptText
    ? promptText.replace(/\s+/g, ' ').slice(0, 20)
    : '未设置提示词'
  const promptLength = promptText.length

  node.configPreview = [
    modelLabel,
    `?? ${node.aiConfig.temperature}`,
    promptText ? `提示词 ${promptLength} 字` : '未设置提示词'
  ]
  node.configDetails = [
    { label: '??', value: modelLabel },
    { label: '???? Token', value: String(node.aiConfig.maxTokens) },
    { label: '提示词', value: promptPreview },
    { label: '提示词长度', value: `${promptLength} 字` }
  ]
}

const syncPromptSelectorPresentation = (node: StageNode) => {
  if (!isPromptSelectorNode(node) || !node.promptConfig) {
    return
  }

  const categoryLabel = node.promptConfig.promptCategory || '未设置分类'
  const promptLabel = getPromptLabel(node.promptConfig.promptId)
  const fields = getPromptFields(node.promptConfig.promptId)
  const filledFieldCount = fields.filter(field => String(node.promptConfig?.fieldValues?.[field.name] ?? '').trim()).length

  node.subtitle = '单独选择提示词分类和提示词内容。'
  node.description = '作为独立的内容节点使用，用于在流程中明确指定当前要引用的提示词配置。'
  node.outputs = ['selectedPrompt', 'selectedPromptContent']
  node.configPreview = [categoryLabel, promptLabel, `字段 ${filledFieldCount}/${fields.length}`]
  node.configDetails = [
    { label: '提示词分类', value: categoryLabel },
    { label: '提示词名称', value: promptLabel },
    { label: '占位符填写', value: `${filledFieldCount}/${fields.length}` }
  ]
}

const syncTextInputPresentation = (node: StageNode | null | undefined) => {
  if (!isTextInputNode(node)) {
    return
  }

  const fields = node.textInputConfig.fields || []
  const textFields = fields.filter(f => f.type === 'text')
  const variableFields = fields.filter(f => f.type === 'variable')

  node.subtitle = '手动输入文本内容或从文件导入。'
  node.description = '支持多个文本字段和变量引用，可从 TXT/MD/JSON 文件导入内容。'
  node.outputs = ['textResult']
  node.configPreview = [
    textFields.length > 0 ? `文本 ${textFields.length}` : '',
    variableFields.length > 0 ? `变量 ${variableFields.length}` : ''
  ].filter(Boolean)
  node.configDetails = [
    { label: '文本字段', value: String(textFields.length) },
    { label: '变量字段', value: String(variableFields.length) },
    { label: '输出变量', value: 'textResult' }
  ]
}

const handleTextInputFileImport = (file: any, fieldId?: string) => {
  const node = selectedTextInputNode.value
  if (!node) {
    return
  }

  const rawFile = file.raw
  if (!rawFile) {
    return
  }

  const fileType = rawFile.name.split('.').pop()?.toLowerCase() as 'txt' | 'md' | 'json'
  if (!['txt', 'md', 'json'].includes(fileType)) {
    ElMessage.error('仅支持 TXT、MD 或 JSON 格式文件')
    return
  }

  const reader = new FileReader()
  reader.onload = (e) => {
    const content = e.target?.result as string
    
    if (fieldId) {
      const field = node.textInputConfig.fields.find(f => f.id === fieldId)
      if (field) {
        field.content = content
        field.importedFile = {
          name: rawFile.name,
          type: fileType,
          content
        }
      }
    } else {
      const firstField = node.textInputConfig.fields[0]
      if (firstField) {
        firstField.content = content
        firstField.importedFile = {
          name: rawFile.name,
          type: fileType,
          content
        }
      }
    }
    syncTextInputPresentation(node)
  }
  reader.onerror = () => {
    ElMessage.error('文件读取失败')
  }
  reader.readAsText(rawFile)
}

const clearTextInputImportedFile = (fieldId: string) => {
  const node = selectedTextInputNode.value
  if (!node) {
    return
  }

  const field = node.textInputConfig.fields.find(f => f.id === fieldId)
  if (field) {
    field.importedFile = undefined
    field.content = ''
    syncTextInputPresentation(node)
  }
}

const fetchWorkflowResources = async () => {
  try {
    const [modelsResponse, promptsResponse] = await Promise.all([configAPI.getAll(), promptAPI.getAll()])

    availableModels.value = modelsResponse.success && modelsResponse.data ? modelsResponse.data.filter(m => m.enabled !== 0) : []
    availablePrompts.value = promptsResponse.success && promptsResponse.data ? promptsResponse.data : []
  } catch {
    availableModels.value = []
    availablePrompts.value = []
  }

  stageNodes.value.forEach(node => {
    if (node.type === 'start') {
      return
    }

    if (node.type === 'aiProcess') {
      normalizeAiConfig(node)
      syncAiNodePresentation(node)
      return
    }

    if (isPromptSelectorNode(node)) {
      normalizePromptSelectorConfig(node)
      syncPromptSelectorPresentation(node)
    }
  })
}

const handleGlobalPointerDown = (event: PointerEvent) => {
  const target = event.target as HTMLElement | null

  if (
    target?.closest('.stage-node-context-menu') ||
    target?.closest('.stage-edge-context-menu') ||
    target?.closest('.stage-node')
  ) {
    return
  }

  closeStageNodeContextMenu()
  closeStageEdgeContextMenu()
}

const handleGlobalViewportChange = () => {
  closeStageNodeContextMenu()
  closeStageEdgeContextMenu()
}

onMounted(async () => {
  await loadWorkflow()
  customNodeLibrary.value = await loadWorkflowNodeLibrary()
  await loadSavedWorkflowScene()
  await fetchWorkflowResources()
  window.addEventListener('pointerdown', handleGlobalPointerDown)
  window.addEventListener('resize', handleGlobalViewportChange)
  window.addEventListener('scroll', handleGlobalViewportChange, true)
})

const getNodeHeight = (node: StageNode) => node.height

const getTargetPoint = (node: StageNode) => {
  const position = getRenderedNodePosition(node)

  return {
    x: position.x - stageBounds.value.minX,
    y: position.y - stageBounds.value.minY + getNodeHeight(node) / 2
  }
}

const getSourcePoint = (node: StageNode, handle?: StageEdge['sourceHandle']) => {
  const position = getRenderedNodePosition(node)
  const x = position.x - stageBounds.value.minX + node.width

  if (node.type === 'loop' && handle === 'loop-body') {
    return { x, y: position.y - stageBounds.value.minY + 48 }
  }

  if (node.type === 'loop' && handle === 'loop-done') {
    return { x, y: position.y - stageBounds.value.minY + 112 }
  }

  return {
    x,
    y: position.y - stageBounds.value.minY + getNodeHeight(node) / 2
  }
}

const buildCurvePath = (
  source: {
    x: number
    y: number
  },
  target: {
    x: number
    y: number
  }
) => {
  const curve = Math.max(80, Math.abs(target.x - source.x) * 0.35)

  return `M ${source.x} ${source.y} C ${source.x + curve} ${source.y}, ${target.x - curve} ${target.y}, ${target.x} ${target.y}`
}

const scheduleDragPreviewPosition = (position: CanvasPoint) => {
  pendingDragPreviewPosition = position

  if (dragPreviewFrameId !== null) {
    return
  }

  dragPreviewFrameId = window.requestAnimationFrame(() => {
    dragPreviewFrameId = null

    if (pendingDragPreviewPosition) {
      dragPreviewPosition.value = pendingDragPreviewPosition
    }
  })
}

const clearDragPreviewPosition = () => {
  if (dragPreviewFrameId !== null) {
    window.cancelAnimationFrame(dragPreviewFrameId)
    dragPreviewFrameId = null
  }

  pendingDragPreviewPosition = null
  dragPreviewPosition.value = null
}

const getRenderedNodePosition = (node: StageNode) => {
  if (dragState.value?.nodeId === node.id && dragPreviewPosition.value) {
    return dragPreviewPosition.value
  }

  return node.position
}

const getStageNodeStyle = (node: StageNode) => {
  const position = getRenderedNodePosition(node)

  return {
    left: `${position.x - stageBounds.value.minX}px`,
    top: `${position.y - stageBounds.value.minY}px`,
    width: `${node.width}px`,
    '--node-height': `${node.height}px`,
    '--accent': node.accent
  }
}

const buildEdgePath = (edge: StageEdge) => {
  const sourceNode = stageNodes.value.find(node => node.id === edge.source)
  const targetNode = stageNodes.value.find(node => node.id === edge.target)

  if (!sourceNode || !targetNode) {
    return ''
  }

  return buildCurvePath(getSourcePoint(sourceNode, edge.sourceHandle), getTargetPoint(targetNode))
}

const draftEdgePath = computed(() => {
  if (!edgeDragState.value) {
    return ''
  }

  const sourceNode = stageNodes.value.find(node => node.id === edgeDragState.value?.source)

  if (!sourceNode) {
    return ''
  }

  return buildCurvePath(getSourcePoint(sourceNode, edgeDragState.value.sourceHandle), {
    x: edgeDragState.value.pointerX - stageBounds.value.minX,
    y: edgeDragState.value.pointerY - stageBounds.value.minY
  })
})

const clampZoom = (value: number) => {
  return Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, Number(value.toFixed(2))))
}

const adjustZoom = (delta: number) => {
  zoom.value = clampZoom(zoom.value + delta)
}

const resetZoom = () => {
  zoom.value = 1
}

const handleCanvasWheel = (event: WheelEvent) => {
  event.preventDefault()
  closeStageNodeContextMenu()
  closeStageEdgeContextMenu()
  const nextZoom = clampZoom(zoom.value + (event.deltaY < 0 ? 0.08 : -0.08))

  if (nextZoom === zoom.value) {
    return
  }

  const viewport = event.currentTarget as HTMLDivElement | null

  if (!viewport) {
    zoom.value = nextZoom
    return
  }

  const rect = viewport.getBoundingClientRect()
  const pointerX = event.clientX - rect.left
  const pointerY = event.clientY - rect.top
  const worldPoint = getCanvasWorldPoint(event, viewport)
  const worldX = worldPoint.x - stageBounds.value.minX
  const worldY = worldPoint.y - stageBounds.value.minY

  zoom.value = nextZoom
  camera.value.x = pointerX - worldX * nextZoom
  camera.value.y = pointerY - worldY * nextZoom
}

const handleLibraryDragStart = (event: DragEvent, libraryId: string) => {
  draggingLibraryId.value = libraryId
  libraryDragPoint.value = null

  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'copy'
    event.dataTransfer.setData('application/x-workflow-node', libraryId)
  }
}

const handleLibraryDragEnd = () => {
  draggingLibraryId.value = ''
  libraryDragPoint.value = null
}

const handleLibraryDragOver = (event: DragEvent) => {
  const viewport = event.currentTarget as HTMLDivElement | null

  if (!draggingLibraryId.value || !viewport) {
    return
  }

  libraryDragPoint.value = getCanvasWorldPoint(event, viewport)

  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'copy'
  }
}

const handleLibraryDrop = (event: DragEvent) => {
  event.preventDefault()
  closeStageNodeContextMenu()
  closeStageEdgeContextMenu()

  const viewport = event.currentTarget as HTMLDivElement | null
  const libraryId =
    event.dataTransfer?.getData('application/x-workflow-node') || draggingLibraryId.value

  draggingLibraryId.value = ''

  if (!viewport || !libraryId) {
    return
  }

  const worldPoint = libraryDragPoint.value ?? getCanvasWorldPoint(event, viewport)
  const nextNode = createStageNodeFromLibrary(libraryId, worldPoint)
  libraryDragPoint.value = null

  if (!nextNode) {
    return
  }

  stageNodes.value.push(nextNode)
  selectedNodeId.value = nextNode.id
}

const startEdgeDrag = (event: PointerEvent, nodeId: string, sourceHandle?: StageEdge['sourceHandle']) => {
  if (event.button !== 0) {
    return
  }

  closeStageNodeContextMenu()
  closeStageEdgeContextMenu()
  const viewport = workspaceScrollRef.value

  if (!viewport) {
    return
  }

  const worldPoint = getCanvasWorldPoint(event, viewport)
  selectedNodeId.value = nodeId
  edgeDragState.value = {
    source: nodeId,
    sourceHandle,
    pointerX: worldPoint.x,
    pointerY: worldPoint.y
  }
  hoveredEdgeTargetId.value = ''

  document.body.style.userSelect = 'none'
  window.addEventListener('pointermove', handleEdgeDragMove)
  window.addEventListener('pointerup', finishEdgeDrag)
}

const handleEdgeDragMove = (event: PointerEvent) => {
  const viewport = workspaceScrollRef.value

  if (!edgeDragState.value || !viewport) {
    return
  }

  const worldPoint = getCanvasWorldPoint(event, viewport)
  edgeDragState.value.pointerX = worldPoint.x
  edgeDragState.value.pointerY = worldPoint.y
}

const createEdgeId = (source: string, target: string, sourceHandle?: StageEdge['sourceHandle']) => {
  const handlePart = sourceHandle ? `-${sourceHandle}` : ''
  return `edge-${source}${handlePart}-${target}`
}

const completeEdgeDrag = (targetNodeId: string) => {
  const currentDrag = edgeDragState.value

  if (!currentDrag) {
    return
  }

  if (currentDrag.source === targetNodeId) {
    stopEdgeDrag()
    return
  }

  const alreadyExists = stageEdges.value.some(
    edge =>
      edge.source === currentDrag.source &&
      edge.target === targetNodeId &&
      edge.sourceHandle === currentDrag.sourceHandle
  )

  if (!alreadyExists) {
    stageEdges.value.push({
      id: createEdgeId(currentDrag.source, targetNodeId, currentDrag.sourceHandle),
      source: currentDrag.source,
      target: targetNodeId,
      sourceHandle: currentDrag.sourceHandle
    })
  }

  stopEdgeDrag()
}

const handleEdgeTargetEnter = (nodeId: string) => {
  if (!edgeDragState.value || edgeDragState.value.source === nodeId) {
    return
  }

  hoveredEdgeTargetId.value = nodeId
}

const handleEdgeTargetLeave = (nodeId: string) => {
  if (hoveredEdgeTargetId.value === nodeId) {
    hoveredEdgeTargetId.value = ''
  }
}

const finishEdgeDrag = () => {
  if (hoveredEdgeTargetId.value) {
    completeEdgeDrag(hoveredEdgeTargetId.value)
    return
  }

  stopEdgeDrag()
}

const stopEdgeDrag = () => {
  edgeDragState.value = null
  hoveredEdgeTargetId.value = ''
  window.removeEventListener('pointermove', handleEdgeDragMove)
  window.removeEventListener('pointerup', finishEdgeDrag)

  if (!dragState.value && !panState.value) {
    document.body.style.userSelect = ''
  }
}

const updateNodePosition = (nodeId: string, x: number, y: number) => {
  const node = stageNodes.value.find(item => item.id === nodeId)

  if (!node) {
    return
  }

  node.position.x = x
  node.position.y = y
}

const closeStageNodeContextMenu = () => {
  stageNodeContextMenu.value = null
}

const closeStageEdgeContextMenu = () => {
  stageEdgeContextMenu.value = null
}

const openStageNodeContextMenu = (event: MouseEvent, nodeId: string) => {
  closeStageEdgeContextMenu()
  selectedNodeId.value = nodeId
  const menuWidth = 132
  const menuHeight = 96
  const offset = 8

  stageNodeContextMenu.value = {
    nodeId,
    x: Math.min(event.clientX, window.innerWidth - menuWidth - offset),
    y: Math.min(event.clientY, window.innerHeight - menuHeight - offset)
  }
}

const openStageEdgeContextMenu = (event: MouseEvent, edgeId: string) => {
  closeStageNodeContextMenu()
  const menuWidth = 132
  const menuHeight = 56
  const offset = 8

  stageEdgeContextMenu.value = {
    edgeId,
    x: Math.min(event.clientX, window.innerWidth - menuWidth - offset),
    y: Math.min(event.clientY, window.innerHeight - menuHeight - offset)
  }
}

const cloneStageNode = (node: StageNode): StageNode => ({
  id: generateNodeId(node.type),
  type: node.type,
  variant: node.variant,
  title: node.title.endsWith('副本') ? node.title : `${node.title} 副本`,
  subtitle: node.subtitle,
  description: node.description,
  outputs: [...node.outputs],
  configPreview: [...node.configPreview],
  configDetails: node.configDetails.map(item => ({
    label: item.label,
    value: item.value
  })),
  position: {
    x: node.position.x + 36,
    y: node.position.y + 36
  },
  width: node.width,
  height: node.height,
  accent: node.accent,
  icon: node.icon,
  aiConfig: node.aiConfig
    ? {
        modelId: node.aiConfig.modelId,
        model: node.aiConfig.model,
        temperature: node.aiConfig.temperature,
        maxTokens: node.aiConfig.maxTokens,
        promptCategory: node.aiConfig.promptCategory,
        promptId: node.aiConfig.promptId,
        promptText: node.aiConfig.promptText,
        extraInput: node.aiConfig.extraInput
      }
    : undefined,
  promptConfig: node.promptConfig
    ? {
        promptCategory: node.promptConfig.promptCategory,
        promptId: node.promptConfig.promptId,
        fieldValues: {
          ...(node.promptConfig.fieldValues ?? {})
        }
      }
    : undefined,
  textInputConfig: node.textInputConfig
    ? cloneTextInputConfig(node.textInputConfig)
    : undefined
})

const copyStageNode = (nodeId: string) => {
  const node = stageNodes.value.find(item => item.id === nodeId)

  if (!node) {
    closeStageNodeContextMenu()
    return
  }

  const duplicatedNode = cloneStageNode(node)
  stageNodes.value.push(duplicatedNode)
  selectedNodeId.value = duplicatedNode.id
  closeStageNodeContextMenu()
  ElMessage.success(`已复制节点：${node.title}`)
}

const deleteStageNode = (nodeId: string) => {
  const node = stageNodes.value.find(item => item.id === nodeId)

  if (!node) {
    closeStageNodeContextMenu()
    return
  }

  stageNodes.value = stageNodes.value.filter(item => item.id !== nodeId)
  stageEdges.value = stageEdges.value.filter(edge => edge.source !== nodeId && edge.target !== nodeId)

  if (selectedNodeId.value === nodeId) {
    selectedNodeId.value = ''
  }
  closeStageNodeContextMenu()
  ElMessage.success(`已删除节点：${node.title}`)
}

const deleteStageEdge = (edgeId: string) => {
  const edge = stageEdges.value.find(item => item.id === edgeId)

  if (!edge) {
    closeStageEdgeContextMenu()
    return
  }

  stageEdges.value = stageEdges.value.filter(item => item.id !== edgeId)
  closeStageEdgeContextMenu()
  ElMessage.success('已取消连线')
}

const generateNodeId = (type: string) => {
  const prefixMap: Record<string, string> = {
    start: 'start',
    aiProcess: 'ai',
    context: 'ctx',
    loop: 'loop',
    text: 'text',
    end: 'end'
  }

  const prefix = prefixMap[type] ?? 'node'
  const usedIds = new Set(stageNodes.value.map(node => node.id))
  let index = 1

  while (usedIds.has(`${prefix}-${index}`)) {
    index += 1
  }

  return `${prefix}-${index}`
}

const createNodeDraft = (type: string): StageNodeDraft | null => {
  switch (type) {
    case 'start':
      return {
        title: '开始节点',
        subtitle: '初始化流程变量并设定入口参数。',
        description: '负责声明工作流运行时需要的初始变量、上下文和触发条件。',
        outputs: ['initialVariables'],
        configPreview: ['变量初始化', '入口参数'],
        configDetails: [
          { label: '触发方式', value: '手动触发' },
          { label: '初始化变量', value: 'chapterTitle, outline, stylePreset' }
        ],
        width: 149,
        height: 83
      }
    case 'aiProcess':
      return {
        title: 'AI处理节点',
        subtitle: '调用模型完成分析、生成或改写。',
        description: '将当前上下文和提示词发送给模型，并输出结构化结果或文本结果。',
        outputs: ['aiResult', 'lastAIOutput'],
        configPreview: ['模型调用', '提示词', '结果输出'],
        configDetails: [
          { label: '模型', value: '默认模型' },
          { label: '提示词来源', value: '流程上下文 + 当前节点配置' }
        ],
        width: 168,
        height: 89
      }
    case 'loop':
      return {
        title: '循环节点',
        subtitle: '遍历数组或按固定次数重复执行。',
        description: '支持数组迭代、次数循环和循环体输出汇总，适合分段生成长内容。',
        outputs: ['currentIteration', 'currentItem', 'loopResult'],
        configPreview: ['循环体', '迭代变量', '汇总结果'],
        configDetails: [
          { label: '循环模式', value: '遍历数组' },
          { label: '当前项变量', value: 'currentItem' }
        ],
        width: 172,
        height: 98
      }
    case 'text':
      return {
        title: '文本节点',
        subtitle: '输出固定文本或模板化内容。',
        description: '可以作为提示词片段、静态正文、固定说明或模板内容的拼装节点。',
        outputs: ['textResult'],
        configPreview: ['文本模板', '静态输出'],
        configDetails: [
          { label: '文本类型', value: '模板文本' },
          { label: '输出变量', value: 'textResult' }
        ],
        width: 151,
        height: 84
      }
    case 'end':
      return {
        title: '结束节点',
        subtitle: '收束流程并定义最终输出。',
        description: '负责汇总流程结果、整理输出格式，并作为整个工作流的最终结束点。',
        outputs: ['outputFile'],
        configPreview: ['结果汇总', '流程结束'],
        configDetails: [
          { label: '输出格式', value: '文本 / JSON' },
          { label: '结束行为', value: '返回最终结果' }
        ],
        width: 146,
        height: 83
      }
    case 'textInput':
      return {
        title: '文本输入节点',
        subtitle: '手动输入文本内容或从文件导入。',
        description: '支持多个文本字段和变量引用，可从 TXT/MD/JSON 文件导入内容。',
        outputs: ['textResult'],
        configPreview: ['文本输入', '变量引用'],
        configDetails: [
          { label: '字段数量', value: '1' },
          { label: '输出变量', value: 'textResult' }
        ],
        width: 156,
        height: 84,
        textInputConfig: createDefaultTextInputConfig()
      }
    default:
      return null
  }
}

const createPromptSelectorDraft = (): StageNodeDraft => {
  const promptConfig = createDefaultPromptSelectorConfig()
  const promptCategory = promptConfig.promptCategory || '未设置分类'
  const promptName = getPromptLabel(promptConfig.promptId) || '未设置提示词'

  return {
    title: '提示词选择节点',
    subtitle: '单独选择提示词分类和提示词内容。',
    description: '作为独立的内容节点使用，用于在流程中明确指定当前要引用的提示词配置。',
    outputs: ['selectedPrompt', 'selectedPromptContent'],
    configPreview: ['提示词分类', '提示词条目'],
    configDetails: [
      { label: '提示词分类', value: promptCategory },
      { label: '提示词名称', value: promptName }
    ],
    width: 158,
    height: 88,
    promptConfig
  }
}

const normalizeNodeType = (type: string): WorkflowNodeType => {
  return type in NODE_TYPE_VISUALS ? (type as WorkflowNodeType) : 'aiProcess'
}

const getNodeVisual = (type: string) => {
  return NODE_TYPE_VISUALS[normalizeNodeType(type)]
}

const cloneDraft = (draft: StageNodeDraft): StageNodeDraft => ({
  title: draft.title,
  subtitle: draft.subtitle,
  description: draft.description,
  outputs: [...draft.outputs],
  configPreview: [...draft.configPreview],
  configDetails: draft.configDetails.map(item => ({
    label: item.label,
    value: item.value
  })),
  width: draft.width,
  height: draft.height,
  aiConfig: draft.aiConfig
    ? {
        modelId: draft.aiConfig.modelId,
        model: draft.aiConfig.model,
        temperature: draft.aiConfig.temperature,
        maxTokens: draft.aiConfig.maxTokens,
        promptCategory: draft.aiConfig.promptCategory,
        promptId: draft.aiConfig.promptId,
        promptText: draft.aiConfig.promptText,
        extraInput: draft.aiConfig.extraInput
      }
    : undefined,
  promptConfig: draft.promptConfig
    ? {
        promptCategory: draft.promptConfig.promptCategory,
        promptId: draft.promptConfig.promptId,
        fieldValues: {
          ...(draft.promptConfig.fieldValues ?? {})
        }
      }
    : undefined,
  textInputConfig: draft.textInputConfig
    ? cloneTextInputConfig(draft.textInputConfig)
    : undefined
})

const NODE_CATEGORY_MAP: Record<WorkflowNodeType, string> = {
  start: '输入节点',
  aiProcess: '处理节点',
  loop: '功能节点',
  text: '内容节点',
  end: '输出节点',
  textInput: '输入节点'
}

const buildBaseNodeLibrary = (): NodeLibraryItem[] => {
  const baseItems: NodeLibraryItem[] = (Object.keys(NODE_TYPE_VISUALS) as WorkflowNodeType[]).map(type => {
    const visual = NODE_TYPE_VISUALS[type]
    const draft = createNodeDraft(type) ?? createNodeDraft('aiProcess')
    return {
      id: `builtin-${type}`,
      type,
      category: NODE_CATEGORY_MAP[type],
      label: visual.label,
      description: visual.description,
      outputs: visual.outputs,
      accent: visual.accent,
      icon: visual.icon,
      template: cloneDraft({
        ...(draft as StageNodeDraft),
        aiConfig: type === 'aiProcess' ? createDefaultAiConfig() : undefined
      })
    }
  })

  const textVisual = NODE_TYPE_VISUALS.text

  baseItems.splice(5, 0, {
    id: 'builtin-prompt-selector',
    type: 'text',
    variant: 'promptSelector',
    category: '内容节点',
    label: '提示词选择节点',
    description: '把提示词选择拆成独立内容节点，便于单独拖拽、连接和复用。',
    outputs: 'selectedPrompt, selectedPromptContent',
    accent: textVisual.accent,
    icon: textVisual.icon,
    template: cloneDraft(createPromptSelectorDraft())
  } satisfies NodeLibraryItem)

  return baseItems
}

const hydrateCustomLibraryItem = (item: WorkflowNodeLibraryTemplate): NodeLibraryItem => {
  const type = normalizeNodeType(item.type)
  const visual = getNodeVisual(type)
  const fallbackDraft =
    item.variant === 'promptSelector'
      ? createPromptSelectorDraft()
      : (createNodeDraft(type) ?? createNodeDraft('aiProcess'))
  const template = item.template ?? (fallbackDraft as StageNodeDraft)

  return {
    id: item.id,
    type,
    variant: item.variant,
    category: item.category || '生成节点',
    label: item.label || template.title || visual.label,
    description: item.description || template.description || visual.description,
    outputs: item.outputs || template.outputs.join(', ') || visual.outputs,
    accent: visual.accent,
    icon: visual.icon,
    generated: Boolean(item.generated),
    template: cloneDraft({
      title: template.title || (fallbackDraft as StageNodeDraft).title,
      subtitle: template.subtitle || (fallbackDraft as StageNodeDraft).subtitle,
      description: template.description || (fallbackDraft as StageNodeDraft).description,
      outputs: Array.isArray(template.outputs) && template.outputs.length > 0
        ? template.outputs
        : (fallbackDraft as StageNodeDraft).outputs,
      configPreview: Array.isArray(template.configPreview) && template.configPreview.length > 0
        ? template.configPreview
        : (fallbackDraft as StageNodeDraft).configPreview,
      configDetails: Array.isArray(template.configDetails) && template.configDetails.length > 0
        ? template.configDetails
        : (fallbackDraft as StageNodeDraft).configDetails,
      width: Number(template.width) || (fallbackDraft as StageNodeDraft).width,
      height: Number(template.height) || (fallbackDraft as StageNodeDraft).height,
      startConfig:
        type === 'start'
          ? {
              fields: Array.isArray(template.startConfig?.fields)
                ? template.startConfig.fields.map(field => normalizeStartField(field))
                : [],
              fieldValues: {
                ...(template.startConfig?.fieldValues ?? {})
              }
            }
          : undefined,
      aiConfig:
        type === 'aiProcess'
          ? {
              ...createDefaultAiConfig(),
              ...(template.aiConfig ?? {})
            }
          : undefined,
      promptConfig:
        item.variant === 'promptSelector'
          ? {
              ...createDefaultPromptSelectorConfig(),
              ...(template.promptConfig ?? {}),
              fieldValues: {
                ...(createDefaultPromptSelectorConfig().fieldValues ?? {}),
                ...(template.promptConfig?.fieldValues ?? {})
              }
            }
          : undefined
    })
  }
}

const nodeLibrary = computed<NodeLibraryItem[]>(() => [
  ...buildBaseNodeLibrary(),
  ...customNodeLibrary.value.map(hydrateCustomLibraryItem)
])

const NODE_CATEGORY_ORDER = ['输入节点', '处理节点', '输出节点', '内容节点', '功能节点']

const groupedNodeLibrary = computed(() => {
  const groups = new Map<string, NodeLibraryItem[]>()
  nodeLibrary.value.forEach(node => {
    const category = node.category || '其他'
    if (!groups.has(category)) {
      groups.set(category, [])
    }
    groups.get(category)!.push(node)
  })
  return NODE_CATEGORY_ORDER.filter(category => groups.has(category)).map(category => ({
    category,
    nodes: groups.get(category)!
  }))
})

watch(
  [() => previewLibraryNode.value?.id ?? '', nodeLibrary],
  ([previewId]) => {
    if (!previewId) {
      return
    }

    const nextPreviewNode = nodeLibrary.value.find(item => item.id === previewId) ?? null
    if (nextPreviewNode) {
      previewLibraryNode.value = nextPreviewNode
    }
  }
)

const persistCustomNodeLibrary = async () => {
  await saveWorkflowNodeLibrary(customNodeLibrary.value)
}

const createStageNodeFromLibrary = (
  libraryId: string,
  position: {
    x: number
    y: number
  }
) => {
  const libraryItem = nodeLibrary.value.find(node => node.id === libraryId)

  if (!libraryItem) {
    return null
  }

  const draft = cloneDraft(libraryItem.template)
  const type = libraryItem.type

  return {
    id: generateNodeId(type),
    type,
    variant: libraryItem.variant,
    title: draft.title,
    subtitle: draft.subtitle,
    description: draft.description,
    outputs: draft.outputs,
    configPreview: draft.configPreview,
    configDetails: draft.configDetails,
    position: {
      x: position.x - draft.width / 2,
      y: position.y - draft.height / 2
    },
    width: draft.width,
    height: draft.height,
    accent: libraryItem.accent,
    icon: libraryItem.icon,
    startConfig:
      type === 'start'
        ? {
            fields: (draft.startConfig?.fields ?? []).map(field => normalizeStartField(field)),
            fieldValues: {
              ...(draft.startConfig?.fieldValues ?? {})
            }
          }
        : undefined,
    aiConfig:
      type === 'aiProcess'
        ? {
            ...createDefaultAiConfig(),
            ...(draft.aiConfig ?? {})
          }
        : undefined,
    promptConfig:
      libraryItem.variant === 'promptSelector'
        ? {
            ...createDefaultPromptSelectorConfig(),
            ...(draft.promptConfig ?? {}),
            fieldValues: {
              ...(createDefaultPromptSelectorConfig().fieldValues ?? {}),
              ...(draft.promptConfig?.fieldValues ?? {})
            }
          }
        : undefined,
    textInputConfig:
      type === 'textInput'
        ? {
            ...createDefaultTextInputConfig(),
            ...(draft.textInputConfig ?? {})
          }
        : undefined
  } satisfies StageNode
}

const getLibraryNodeCode = (node: NodeLibraryItem) => {
  const codePayload = {
    id: node.id,
    type: node.type,
    variant: node.variant,
    category: node.category,
    label: node.label,
    description: node.description,
    outputs: node.outputs,
    accent: node.accent,
    generated: Boolean(node.generated),
    template: {
      title: node.template.title,
      subtitle: node.template.subtitle,
      description: node.template.description,
      outputs: [...node.template.outputs],
      configPreview: [...node.template.configPreview],
      configDetails: node.template.configDetails.map(item => ({
        label: item.label,
        value: item.value
      })),
      width: node.template.width,
      height: node.template.height,
      startConfig: node.template.startConfig
        ? {
            fields: node.template.startConfig.fields.map(field => ({
              name: field.name,
              label: field.label,
              type: field.type,
              options: [...field.options],
              description: field.description,
              required: field.required
            })),
            fieldValues: {
              ...(node.template.startConfig.fieldValues ?? {})
            }
          }
        : undefined,
      aiConfig: node.template.aiConfig
        ? {
            modelId: node.template.aiConfig.modelId,
            model: node.template.aiConfig.model,
            temperature: node.template.aiConfig.temperature,
            maxTokens: node.template.aiConfig.maxTokens,
            promptCategory: node.template.aiConfig.promptCategory,
            promptId: node.template.aiConfig.promptId,
            promptText: node.template.aiConfig.promptText,
            extraInput: node.template.aiConfig.extraInput
          }
        : undefined,
      promptConfig: node.template.promptConfig
        ? {
            promptCategory: node.template.promptConfig.promptCategory,
            promptId: node.template.promptConfig.promptId,
            fieldValues: {
              ...(node.template.promptConfig.fieldValues ?? {})
            }
          }
        : undefined
    }
  }

  return `const nodeTemplate = ${JSON.stringify(codePayload, null, 2)}`
}

const copyTextToClipboard = async (content: string, successMessage: string) => {
  try {
    await navigator.clipboard.writeText(content)
    ElMessage.success(successMessage)
  } catch {
    const textarea = document.createElement('textarea')
    textarea.value = content
    textarea.setAttribute('readonly', 'true')
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    textarea.style.pointerEvents = 'none'
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    ElMessage.success(successMessage)
  }
}

const copyLibraryNodeCode = async (node: NodeLibraryItem) => {
  const code = getLibraryNodeCode(node)
  await copyTextToClipboard(code, `已复制节点代码：${node.label}`)
}

const openLibraryNodePreview = (node: NodeLibraryItem) => {
  previewLibraryNode.value = node
  libraryPreviewVisible.value = true
}

const removeLibraryNode = (node: NodeLibraryItem) => {
  if (!isCustomLibraryNode(node)) {
    return
  }

  customNodeLibrary.value = customNodeLibrary.value.filter(item => item.id !== node.id)
  persistCustomNodeLibrary()
  ElMessage.success(`已删除节点模板：${node.label}`)
}

const hydrateStageNode = (node: WorkflowSceneNode): StageNode | null => {
  const type = normalizeNodeType(node.type)
  const libraryItem =
    nodeLibrary.value.find(item => item.type === type && item.variant === node.variant) ??
    nodeLibrary.value.find(item => item.type === type)
  const fallbackDraft = node.variant === 'promptSelector' ? createPromptSelectorDraft() : createNodeDraft(type)

  if (!libraryItem || !fallbackDraft) {
    return null
  }

  return {
    id: node.id,
    type,
    variant: node.variant,
    title: node.title,
    subtitle: node.subtitle,
    description: node.description,
    outputs: Array.isArray(node.outputs) ? node.outputs : fallbackDraft.outputs,
    configPreview: Array.isArray(node.configPreview) ? node.configPreview : fallbackDraft.configPreview,
    configDetails: Array.isArray(node.configDetails) ? node.configDetails : fallbackDraft.configDetails,
    position: {
      x: Number(node.position?.x) || 0,
      y: Number(node.position?.y) || 0
    },
    width: Number(node.width) || fallbackDraft.width,
    height: Number(node.height) || fallbackDraft.height,
    accent: libraryItem.accent,
    icon: libraryItem.icon,
    startConfig:
      type === 'start'
        ? {
            fields: Array.isArray(node.startConfig?.fields)
              ? node.startConfig.fields.map(field => normalizeStartField(field))
              : [],
            fieldValues: {
              ...(node.startConfig?.fieldValues ?? {})
            }
          }
        : undefined,
    aiConfig:
      type === 'aiProcess'
        ? {
            ...createDefaultAiConfig(),
            ...(node.aiConfig
              ? {
                  modelId: node.aiConfig.modelId ?? null,
                  model: node.aiConfig.model ?? '',
                  temperature: node.aiConfig.temperature ?? 0.8,
                  maxTokens: node.aiConfig.maxTokens ?? 4096,
                  promptCategory: node.aiConfig.promptCategory ?? '',
                  promptId: node.aiConfig.promptId ?? null,
                  promptText: node.aiConfig.promptText ?? '',
                  extraInput: node.aiConfig.extraInput ?? ''
                }
              : {})
          }
        : undefined,
    promptConfig:
      node.variant === 'promptSelector'
        ? {
            ...createDefaultPromptSelectorConfig(),
            ...(node.promptConfig
              ? {
                  promptCategory: node.promptConfig.promptCategory ?? '',
                  promptId: node.promptConfig.promptId ?? null,
                  fieldValues: {
                    ...(node.promptConfig.fieldValues ?? {})
                  }
                }
              : {})
          }
        : undefined,
    textInputConfig:
      type === 'textInput'
        ? {
            ...normalizeTextInputConfig(node.textInputConfig)
          }
        : undefined
  }
}

const serializeStageNode = (node: StageNode): WorkflowSceneNode => ({
  id: node.id,
  type: node.type,
  variant: node.variant,
  title: node.title,
  subtitle: node.subtitle,
  description: node.description,
  outputs: [...node.outputs],
  configPreview: [...node.configPreview],
  configDetails: node.configDetails.map(item => ({
    label: item.label,
    value: item.value
  })),
  position: {
    x: node.position.x,
    y: node.position.y
  },
  width: node.width,
  height: node.height,
  startConfig: node.startConfig
    ? {
        fields: node.startConfig.fields.map(field => ({
          name: field.name,
          label: field.label,
          type: field.type,
          options: [...field.options],
          description: field.description,
          required: field.required
        })),
        fieldValues: {
          ...(node.startConfig.fieldValues ?? {})
        }
      }
    : undefined,
  aiConfig: node.aiConfig
    ? {
        modelId: node.aiConfig.modelId,
        model: node.aiConfig.model,
        temperature: node.aiConfig.temperature,
        maxTokens: node.aiConfig.maxTokens,
        promptCategory: node.aiConfig.promptCategory,
        promptId: node.aiConfig.promptId,
        promptText: node.aiConfig.promptText,
        extraInput: node.aiConfig.extraInput
      }
    : undefined,
  promptConfig: node.promptConfig
    ? {
        promptCategory: node.promptConfig.promptCategory,
        promptId: node.promptConfig.promptId,
        fieldValues: {
          ...(node.promptConfig.fieldValues ?? {})
        }
      }
    : undefined,
  textInputConfig: node.textInputConfig
    ? {
        fields: node.textInputConfig.fields.map(field => ({
          id: field.id,
          name: field.name,
          type: field.type,
          content: field.content,
          placeholder: field.placeholder,
          importedFile: field.importedFile
            ? {
                name: field.importedFile.name,
                type: field.importedFile.type,
                content: field.importedFile.content
              }
            : undefined
        }))
      }
    : undefined
})

const loadSavedWorkflowScene = async () => {
  if (Number.isNaN(workflowId.value)) {
    return
  }

  const scene = await getWorkflowSceneById(workflowId.value)

  if (!scene) {
    return
  }

  stageNodes.value = Array.isArray(scene.nodes)
    ? scene.nodes.map(hydrateStageNode).filter((node): node is StageNode => node !== null)
    : []
  stageEdges.value = Array.isArray(scene.edges) ? scene.edges : []

  if (scene.viewport) {
    zoom.value = clampZoom(scene.viewport.zoom)
    camera.value = {
      x: Number(scene.viewport.camera?.x) || 220,
      y: Number(scene.viewport.camera?.y) || 120
    }
  }
}

const getCanvasWorldPoint = (event: DragEvent | WheelEvent | PointerEvent, viewport: HTMLDivElement) => {
  const rect = viewport.getBoundingClientRect()
  const pointerX = event.clientX - rect.left
  const pointerY = event.clientY - rect.top

  return {
    x: (pointerX - camera.value.x) / zoom.value + stageBounds.value.minX,
    y: (pointerY - camera.value.y) / zoom.value + stageBounds.value.minY
  }
}

const handleNodeDragMove = (event: PointerEvent) => {
  if (!dragState.value) {
    return
  }

  const deltaX = (event.clientX - dragState.value.pointerStartX) / zoom.value
  const deltaY = (event.clientY - dragState.value.pointerStartY) / zoom.value

  const nextPosition = {
    x: dragState.value.originX + deltaX,
    y: dragState.value.originY + deltaY
  }

  dragState.value.currentX = nextPosition.x
  dragState.value.currentY = nextPosition.y
  scheduleDragPreviewPosition(nextPosition)
}

const stopNodeDrag = () => {
  if (dragState.value) {
    const finalPosition = pendingDragPreviewPosition ?? dragPreviewPosition.value

    if (finalPosition) {
      updateNodePosition(dragState.value.nodeId, finalPosition.x, finalPosition.y)
    }
  }

  clearDragPreviewPosition()
  dragState.value = null
  window.removeEventListener('pointermove', handleNodeDragMove)
  window.removeEventListener('pointerup', stopNodeDrag)

  if (!edgeDragState.value && !panState.value && !resizeState.value) {
    document.body.style.userSelect = ''
  }
}

const MIN_NODE_WIDTH = 100
const MIN_NODE_HEIGHT = 60

const startNodeResize = (event: PointerEvent, nodeId: string) => {
  event.stopPropagation()
  event.preventDefault()

  const node = stageNodes.value.find(item => item.id === nodeId)
  if (!node) return

  resizeState.value = {
    nodeId,
    pointerStartX: event.clientX,
    pointerStartY: event.clientY,
    originWidth: node.width,
    originHeight: node.height,
    currentWidth: node.width,
    currentHeight: node.height
  }

  document.body.style.userSelect = 'none'
  window.addEventListener('pointermove', handleNodeResizeMove)
  window.addEventListener('pointerup', stopNodeResize)
}

const handleNodeResizeMove = (event: PointerEvent) => {
  if (!resizeState.value) return

  const deltaX = event.clientX - resizeState.value.pointerStartX
  const deltaY = event.clientY - resizeState.value.pointerStartY

  const newWidth = Math.max(MIN_NODE_WIDTH, resizeState.value.originWidth + deltaX)
  const newHeight = Math.max(MIN_NODE_HEIGHT, resizeState.value.originHeight + deltaY)

  resizeState.value.currentWidth = newWidth
  resizeState.value.currentHeight = newHeight

  const node = stageNodes.value.find(item => item.id === resizeState.value!.nodeId)
  if (node) {
    node.width = newWidth
    node.height = newHeight
  }
}

const stopNodeResize = () => {
  resizeState.value = null
  window.removeEventListener('pointermove', handleNodeResizeMove)
  window.removeEventListener('pointerup', stopNodeResize)

  if (!edgeDragState.value && !panState.value && !dragState.value) {
    document.body.style.userSelect = ''
  }
}

const startNodeDrag = (event: PointerEvent, nodeId: string) => {
  if (event.button !== 0) {
    return
  }

  event.stopPropagation()
  closeStageNodeContextMenu()
  closeStageEdgeContextMenu()
  selectedNodeId.value = nodeId

  const node = stageNodes.value.find(item => item.id === nodeId)

  if (!node) {
    return
  }

  dragState.value = {
    nodeId,
    pointerStartX: event.clientX,
    pointerStartY: event.clientY,
    originX: node.position.x,
    originY: node.position.y,
    currentX: node.position.x,
    currentY: node.position.y
  }

  dragPreviewPosition.value = {
    x: node.position.x,
    y: node.position.y
  }
  document.body.style.userSelect = 'none'
  window.addEventListener('pointermove', handleNodeDragMove)
  window.addEventListener('pointerup', stopNodeDrag)
}

const selectNode = (nodeId: string) => {
  closeStageNodeContextMenu()
  closeStageEdgeContextMenu()
  selectedNodeId.value = nodeId
}

const toggleLeftSidebar = () => {
  leftSidebarCollapsed.value = !leftSidebarCollapsed.value
}

const toggleRightInspector = () => {
  rightInspectorCollapsed.value = !rightInspectorCollapsed.value
}

const handleSaveWorkflow = async () => {
  if (!workflow.value) {
    ElMessage.error('当前工作流不存在，无法保存')
    return
  }

  stageNodes.value.forEach(node => {
    if (node.type === 'start') {
      normalizeStartConfig(node)
      syncStartNodePresentation(node)
      return
    }

    if (node.type === 'aiProcess') {
      normalizeAiConfig(node)
      syncAiNodePresentation(node)
      return
    }

    if (isPromptSelectorNode(node)) {
      normalizePromptSelectorConfig(node)
      syncPromptSelectorPresentation(node)
    }
  })

  await saveWorkflowScene(workflow.value.id, {
    nodes: stageNodes.value.map(serializeStageNode),
    edges: stageEdges.value.map(edge => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      sourceHandle: edge.sourceHandle
    })),
    viewport: {
      zoom: zoom.value,
      camera: {
        x: camera.value.x,
        y: camera.value.y
      }
    }
  })

  ElMessage.success('工作流已保存')
}

const addWorkflowLog = (
  level: 'info' | 'success' | 'error' | 'warning',
  message: string,
  nodeId?: string,
  nodeTitle?: string,
  extra?: {
    modelName?: string
    promptContent?: string
    outputContent?: string
    inputTokens?: number
    outputTokens?: number
  }
) => {
  const now = new Date()
  const timestamp = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`
  workflowLogs.value.push({
    id: `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`,
    timestamp,
    level,
    message,
    nodeId,
    nodeTitle,
    modelName: extra?.modelName,
    promptContent: extra?.promptContent,
    outputContent: extra?.outputContent,
    inputTokens: extra?.inputTokens,
    outputTokens: extra?.outputTokens
  })
}

const clearWorkflowLogs = () => {
  workflowLogs.value = []
}

const viewLogDetail = (log: typeof workflowLogs.value[0]) => {
  selectedLogRecord.value = log
  logDetailDialogVisible.value = true
}

const getNextNodes = (nodeId: string) => {
  return stageEdges.value
    .filter(edge => edge.source === nodeId)
    .map(edge => edge.target)
}

const resolvePromptFieldPlaceholders = (promptContent: string): string => {
  if (!promptContent) return ''
  
  let result = promptContent
  const context = workflowExecutionContext.value
  
  const fieldRegex = /\$\{([^}]+)\}/g
  let fieldMatch: RegExpExecArray | null
  
  while ((fieldMatch = fieldRegex.exec(promptContent)) !== null) {
    const fullMatch = fieldMatch[0]
    const fieldName = fieldMatch[1].trim()
    const fieldValue = context[fieldName]
    
    if (fieldValue !== undefined && fieldValue !== null) {
      result = result.replace(fullMatch, String(fieldValue))
    }
  }
  
  return result
}

const resolveWorkflowPlaceholders = (text: string): string => {
  if (!text) return ''
  
  let result = text
  
  const workflowPlaceholderRegex = /\[\[workflow\.([^\]]+)\]\]/g
  let match: RegExpExecArray | null
  
  while ((match = workflowPlaceholderRegex.exec(text)) !== null) {
    const fullMatch = match[0]
    const placeholderName = match[1].trim()
    
    const contextValue = workflowExecutionContext.value[placeholderName]
    if (contextValue !== undefined && contextValue !== null) {
      result = result.replace(fullMatch, String(contextValue))
      continue
    }
    
    const promptByName = availablePrompts.value.find(
      p => p.name && p.name.trim().toLowerCase() === placeholderName.toLowerCase()
    )
    if (promptByName && promptByName.content) {
      const resolvedPromptContent = resolvePromptFieldPlaceholders(promptByName.content)
      result = result.replace(fullMatch, resolvedPromptContent)
      continue
    }
    
    result = result.replace(fullMatch, `[未找到: ${placeholderName}]`)
  }
  
  return result
}

const buildAiPrompt = (node: StageNode): string => {
  const aiConfig = node.aiConfig
  if (!aiConfig) return ''
  
  let prompt = aiConfig.promptText || ''
  
  prompt = resolveWorkflowPlaceholders(prompt)
  
  const context = workflowExecutionContext.value
  Object.keys(context).forEach(key => {
    const regex = new RegExp(`{{${key}}}`, 'g')
    prompt = prompt.replace(regex, context[key] || '')
  })
  
  if (aiConfig.extraInput) {
    prompt += `\n\n${aiConfig.extraInput}`
  }
  
  return prompt
}

const parseSseResponse = (sseText: string): string => {
  const lines = sseText.split('\n')
  const contents: string[] = []
  
  for (const line of lines) {
    const trimmed = line.trim()
    if (!trimmed || !trimmed.startsWith('data:')) continue
    
    const message = trimmed.replace(/^data:\s*/, '')
    if (message === '[DONE]') break
    
    try {
      const parsed = JSON.parse(message)
      if (parsed.content) {
        contents.push(parsed.content)
      }
    } catch {
      // 忽略解析错误
    }
  }
  
  return contents.join('')
}

const executeAiNode = async (node: StageNode, outputKey: string): Promise<string> => {
  const aiConfig = node.aiConfig
  if (!aiConfig) {
    throw new Error('AI 节点配置为空')
  }
  
  const prompt = buildAiPrompt(node)
  if (!prompt.trim()) {
    throw new Error('AI 提示词为空')
  }
  
  const modelConfig = availableModels.value.find(m => m.id === aiConfig.modelId)
  if (!modelConfig || !modelConfig.api_key || !modelConfig.api_url) {
    throw new Error('未找到有效的 API 配置')
  }
  
  const messages = [
    { role: 'user' as const, content: prompt }
  ]
  
  const requestBody = {
    model: modelConfig.model,
    messages,
    temperature: modelConfig.temperature,
    max_tokens: modelConfig.max_tokens,
    stream: true
  }
  
  const abortController = new AbortController()
  aiGenerationControllers.value.set(node.id, abortController)
  
  try {
    const response = await fetch(modelConfig.api_url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${modelConfig.api_key}`
      },
      body: JSON.stringify(requestBody),
      signal: abortController.signal
    })
    
    if (!response.ok) {
      throw new Error(`API 请求失败: ${response.status}`)
    }
    
    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('无法获取响应流')
    }
    
    const decoder = new TextDecoder()
    let fullContent = ''
    workflowExecutionContext.value[outputKey] = ''
    
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      
      const chunk = decoder.decode(value, { stream: true })
      const lines = chunk.split('\n')
      
      for (const line of lines) {
        const trimmedLine = line.trim()
        if (!trimmedLine || !trimmedLine.startsWith('data:')) continue
        
        const message = trimmedLine.replace(/^data:\s*/, '')
        if (message === '[DONE]') continue
        
        try {
          const parsed = JSON.parse(message)
          if (parsed.error) {
            throw new Error(parsed.error.message || parsed.error)
          }
          const content = parsed.choices?.[0]?.delta?.content || ''
          if (content) {
            fullContent += content
            workflowExecutionContext.value[outputKey] = fullContent
          }
        } catch {
          // 忽略解析错误
        }
      }
    }
    
    return fullContent
  } catch (error: any) {
    if (error.name === 'AbortError') {
      return workflowExecutionContext.value[outputKey] || ''
    }
    throw error
  } finally {
    aiGenerationControllers.value.delete(node.id)
  }
}

const getIncomingEdges = (nodeId: string) => {
  return stageEdges.value.filter(edge => edge.target === nodeId)
}

const addRunningEdges = (nodeId: string) => {
  const incomingEdges = getIncomingEdges(nodeId)
  incomingEdges.forEach(edge => {
    runningEdgeIds.value.add(edge.id)
  })
}

const removeRunningEdges = (nodeId: string) => {
  const incomingEdges = getIncomingEdges(nodeId)
  incomingEdges.forEach(edge => {
    runningEdgeIds.value.delete(edge.id)
  })
}

const executeWorkflow = async () => {
  if (isWorkflowRunning.value) {
    ElMessage.warning('工作流正在运行中')
    return
  }

  if (!stageNodes.value.length) {
    ElMessage.warning('请先添加节点')
    return
  }

  const startNode = stageNodes.value.find(node => node.type === 'start')
  if (!startNode) {
    ElMessage.warning('请添加开始节点')
    return
  }

  isWorkflowRunning.value = true
  runningEdgeIds.value = new Set()
  clearWorkflowLogs()

  try {
    const executed = new Set<string>()
    const queue: string[] = [startNode.id]

    while (queue.length > 0 && isWorkflowRunning.value) {
      const nodeId = queue.shift()!
      if (executed.has(nodeId)) continue

      const node = stageNodes.value.find(n => n.id === nodeId)
      if (!node) continue

      executed.add(nodeId)
      runningNodeId.value = nodeId
      addRunningEdges(nodeId)

      await new Promise(resolve => setTimeout(resolve, 500))

      try {
        switch (node.type) {
          case 'start': {
            const startConfig = node.startConfig
            if (startConfig?.fields && startConfig.fieldValues) {
              for (const field of startConfig.fields) {
                const value = startConfig.fieldValues[field.name] || ''
                workflowExecutionContext.value[field.name] = value
              }
            }
            break
          }
          case 'aiProcess': {
            const outputKey = `ai_output_${nodeId}`
            const aiConfig = node.aiConfig
            const modelConfig = aiConfig ? availableModels.value.find(m => m.id === aiConfig.modelId) : null
            const promptContent = buildAiPrompt(node)
            
            try {
              const result = await executeAiNode(node, outputKey)
              addWorkflowLog('success', 'AI 节点执行完成', nodeId, node.title, {
                modelName: modelConfig?.name || modelConfig?.model,
                promptContent,
                outputContent: result
              })
              
              if (node.aiConfig?.stepMode) {
                waitingForContinue.value = true
                waitingNodeId.value = nodeId
                await new Promise<void>(resolve => {
                  continueResolver = resolve
                })
                waitingForContinue.value = false
                waitingNodeId.value = null
                continueResolver = null
              }
            } catch (error: any) {
              addWorkflowLog('error', `AI 调用失败：${error.message}`, nodeId, node.title, {
                modelName: modelConfig?.name || modelConfig?.model,
                promptContent
              })
              throw error
            }
            break
          }
          case 'loop':
            await new Promise(resolve => setTimeout(resolve, 600))
            break
          case 'text':
            break
          case 'textInput': {
            if (node.textInputConfig?.fields) {
              for (const field of node.textInputConfig.fields) {
                workflowExecutionContext.value[field.name] = field.content || ''
              }
            }
            break
          }
          case 'end': {
            const outputs: string[] = []
            Object.keys(workflowExecutionContext.value).forEach(key => {
              const value = workflowExecutionContext.value[key]
              if (value) {
                outputs.push(`${key}: ${value}`)
              }
            })
            workflowFinalOutput.value = outputs.join('\n\n')
            removeRunningEdges(nodeId)
            isWorkflowRunning.value = false
            runningNodeId.value = null
            return
          }
        }
      } catch (nodeError: any) {
        throw nodeError
      }

      removeRunningEdges(nodeId)
      runningNodeId.value = null

      const nextNodeIds = getNextNodes(nodeId)
      for (const nextId of nextNodeIds) {
        if (!executed.has(nextId)) {
          queue.push(nextId)
        }
      }

      await new Promise(resolve => setTimeout(resolve, 200))
    }

    if (isWorkflowRunning.value) {
      addWorkflowLog('success', '工作流执行完成')
    }
  } catch (error) {
    addWorkflowLog('error', `执行出错: ${error}`)
  } finally {
    isWorkflowRunning.value = false
    runningNodeId.value = null
    runningEdgeIds.value = new Set()
  }
}

const stopWorkflow = () => {
  if (!isWorkflowRunning.value) {
    ElMessage.info('工作流未在运行')
    return
  }
  isWorkflowRunning.value = false
  runningNodeId.value = null
  runningEdgeIds.value = new Set()
  waitingForContinue.value = false
  waitingNodeId.value = null
  if (continueResolver) {
    continueResolver()
    continueResolver = null
  }
  ElMessage.info('工作流已停止')
}

const copyOutput = async () => {
  if (!workflowFinalOutput.value) {
    ElMessage.warning('没有可复制的内容')
    return
  }
  
  try {
    await navigator.clipboard.writeText(workflowFinalOutput.value)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

const handleToolbarAction = (actionKey: string) => {
  switch (actionKey) {
    case 'save':
      handleSaveWorkflow()
      return
    case 'run':
      executeWorkflow()
      return
    case 'stop':
      stopWorkflow()
      return
    case 'logs':
      workflowLogsDialogVisible.value = true
      return
    default:
      ElMessage.info('该功能还在开发中')
  }
}

const handleCanvasPanMove = (event: PointerEvent) => {
  if (!panState.value) {
    return
  }

  camera.value.x = panState.value.originX + (event.clientX - panState.value.pointerStartX)
  camera.value.y = panState.value.originY + (event.clientY - panState.value.pointerStartY)
}

const stopCanvasPan = () => {
  panState.value = null
  window.removeEventListener('pointermove', handleCanvasPanMove)
  window.removeEventListener('pointerup', stopCanvasPan)

  if (!dragState.value && !edgeDragState.value) {
    document.body.style.userSelect = ''
  }
}

const startCanvasPan = (event: PointerEvent) => {
  if (event.button !== 0) {
    return
  }

  closeStageNodeContextMenu()
  closeStageEdgeContextMenu()
  const target = event.target as HTMLElement | null

  if (target?.closest('.stage-node')) {
    return
  }

  panState.value = {
    pointerStartX: event.clientX,
    pointerStartY: event.clientY,
    originX: camera.value.x,
    originY: camera.value.y
  }

  document.body.style.userSelect = 'none'
  window.addEventListener('pointermove', handleCanvasPanMove)
  window.addEventListener('pointerup', stopCanvasPan)
}

const goBack = () => {
  router.push('/workflow')
}

const syncCodeSourceFromStage = () => {
  const sceneData = {
    nodes: stageNodes.value.map(node => ({
      id: node.id,
      type: node.type,
      variant: node.variant,
      title: node.title,
      subtitle: node.subtitle,
      description: node.description,
      outputs: node.outputs,
      configPreview: node.configPreview,
      configDetails: node.configDetails,
      position: node.position,
      width: node.width,
      height: node.height,
      accent: node.accent,
      startConfig: node.startConfig,
      aiConfig: node.aiConfig,
      promptConfig: node.promptConfig
    })),
    edges: stageEdges.value,
    viewport: {
      zoom: zoom.value,
      camera: {
        x: camera.value.x,
        y: camera.value.y
      }
    }
  }
  codeSource.value = JSON.stringify(sceneData, null, 2)
  codeError.value = ''
}

const formatCodeSource = () => {
  try {
    const parsed = JSON.parse(codeSource.value)
    codeSource.value = JSON.stringify(parsed, null, 2)
    codeError.value = ''
  } catch (e) {
    codeError.value = `JSON格式错误: ${e instanceof Error ? e.message : String(e)}`
  }
}

const applyCodeSource = () => {
  try {
    const parsed = JSON.parse(codeSource.value)
    
    if (!Array.isArray(parsed.nodes)) {
      throw new Error('缺少nodes数组')
    }
    if (!Array.isArray(parsed.edges)) {
      throw new Error('缺少edges数组')
    }

    stageNodes.value = parsed.nodes.map((node: any) => ({
      ...node,
      icon: NODE_TYPE_VISUALS[node.type as WorkflowNodeType]?.icon || Flag
    }))
    stageEdges.value = parsed.edges

    if (parsed.viewport) {
      zoom.value = parsed.viewport.zoom || 1
      if (parsed.viewport.camera) {
        camera.value.x = parsed.viewport.camera.x || 0
        camera.value.y = parsed.viewport.camera.y || 0
      }
    }

    codeError.value = ''
    ElMessage.success('源码已应用')
    editorMode.value = 'visual'
  } catch (e) {
    codeError.value = `应用失败: ${e instanceof Error ? e.message : String(e)}`
  }
}

const DEFAULT_WORKFLOW_GENERATOR_PROMPT = `你是一个专业的工作流JSON生成器。用户会描述他们需要的工作流功能，你需要根据描述生成符合规范的工作流JSON代码。

## 工作流JSON结构规范

工作流由节点和连线组成

## 可用节点类型

{
  "nodes": [
    {
      "id": "start-1",
      "type": "start",
      "title": "开始节点",
      "subtitle": "初始化流程变量并设定入口参数。",
      "description": "负责声明工作流运行时需要的初始变量、上下文和触发条件。",
      "outputs": [
        "initialVariables"
      ],
      "configPreview": [
        "变量初始化",
        "入口参数"
      ],
      "configDetails": [
        {
          "label": "触发方式",
          "value": "手动触发"
        },
        {
          "label": "字段配置",
          "value": "未配置"
        },
        {
          "label": "占位符统计",
          "value": "0 个"
        }
      ],
      "position": {
        "x": -608.9971466064453,
        "y": -449.9943161010742
      },
      "width": 248,
      "height": 138,
      "accent": "#22c55e",
      "startConfig": {
        "fields": [],
        "fieldValues": {}
      }
    },
    {
      "id": "ai-1",
      "type": "aiProcess",
      "title": "AI处理节点",
      "subtitle": "调用模型完成分析、生成或改写。",
      "description": "将当前上下文和提示词发送给模型，并输出结构化结果或文本结果。",
      "outputs": [
        "aiResult",
        "lastAIOutput"
      ],
      "configPreview": [
        "DeepSeek-V3.2（chat）",
        "?? 0.8",
        "未设置提示词"
      ],
      "configDetails": [
        {
          "label": "??",
          "value": "DeepSeek-V3.2（chat）"
        },
        {
          "label": "???? Token",
          "value": "8000"
        },
        {
          "label": "提示词",
          "value": "未设置提示词"
        },
        {
          "label": "提示词长度",
          "value": "0 字"
        }
      ],
      "position": {
        "x": ,
        "y": 
      },
      "width": 280,
      "height": 148,
      "accent": "#3b82f6",
      "aiConfig": {
        "modelId": 12,
        "model": "deepseek-chat",
        "temperature": 0.8,
        "maxTokens": 8000,
        "promptCategory": "",
        "promptId": null,
        "promptText": "",
        "extraInput": ""
      }
    },
    {
      "id": "text-1",
      "type": "text",
      "variant": "promptSelector",
      "title": "提示词选择节点",
      "subtitle": "单独选择提示词分类和提示词内容。",
      "description": "作为独立的内容节点使用，用于在流程中明确指定当前要引用的提示词配置。",
      "outputs": [
        "selectedPrompt",
        "selectedPromptContent"
      ],
      "configPreview": [
        "简介",
        "【老菠萝】导语仿写生成",
        "字段 0/1"
      ],
      "configDetails": [
        {
          "label": "提示词分类",
          "value": "简介"
        },
        {
          "label": "提示词名称",
          "value": "【老菠萝】导语仿写生成"
        },
        {
          "label": "占位符填写",
          "value": "0/1"
        }
      ],
      "position": {
        "x": ,
        "y": 
      },
      "width": 264,
      "height": 146,
      "accent": "#ec4899",
      "promptConfig": {
        "promptCategory": "简介",
        "promptId": 415,
        "fieldValues": {
          "选中内容": ""
        }
      }
    },
    {
      "id": "end-1",
      "type": "end",
      "title": "结束节点",
      "subtitle": "收束流程并定义最终输出。",
      "description": "负责汇总流程结果、整理输出格式，并作为整个工作流的最终结束点。",
      "outputs": [
        "outputFile"
      ],
      "configPreview": [
        "结果汇总",
        "流程结束"
      ],
      "configDetails": [
        {
          "label": "输出格式",
          "value": "文本 / JSON"
        },
        {
          "label": "结束行为",
          "value": "返回最终结果"
        }
      ],
      "position": {
        "x": ,
        "y": 
      },
      "width": 244,
      "height": 138,
      "accent": "#ef4444"
    },
    {
      "id": "loop-1",
      "type": "loop",
      "title": "循环节点",
      "subtitle": "遍历数组或按固定次数重复执行。",
      "description": "支持数组迭代、次数循环和循环体输出汇总，适合分段生成长内容。",
      "outputs": [
        "currentIteration",
        "currentItem",
        "loopResult"
      ],
      "configPreview": [
        "循环体",
        "迭代变量",
        "汇总结果"
      ],
      "configDetails": [
        {
          "label": "循环模式",
          "value": "遍历数组"
        },
        {
          "label": "当前项变量",
          "value": "currentItem"
        }
      ],
      "position": {
        "x": ,
        "y": 
      },
      "width": 286,
      "height": 164,
      "accent": "#8b5cf6"
    }
  ],
  "edges": [],
  "viewport": {
    "zoom": 1,
    "camera": {
      "x": ,
      "y": 
    }
  }
}

## 节点基本结构

每个节点需要包含：
- id: 唯一标识符，格式如 "node_1", "node_2"
- type: 节点类型
- position: { x, y } 位置坐标（以 "x": 220, "y": 120为中心）
- data: 节点配置数据
"viewport": {
    "zoom": 0.6,
    "camera": {
      "x": -111.34947248186386,
      "y": -140.60009964652681
    }
  }

## 连线结构

每条连线需要包含：
- id: 唯一标识符，格式如 "edge_1", "edge_2"
- source: 源节点id
- target: 目标节点id
- sourceHandle: 源节点连接点（可选）
- targetHandle: 目标节点连接点（可选）

## 生成规则

1. 始终以start节点开始，end节点结束
2. 节点位置要合理分布，避免重叠
3. 连线要正确连接各节点
4. 根据用户需求选择合适的节点类型组合
5. 只输出JSON代码，不要有其他解释文字
6. 确保生成的JSON格式正确，可直接解析使用`

const currentWorkflowGeneratorPrompt = ref(DEFAULT_WORKFLOW_GENERATOR_PROMPT)

const showWorkflowPromptEditor = () => {
  editableWorkflowPrompt.value = currentWorkflowGeneratorPrompt.value
  workflowPromptEditorVisible.value = true
}

const saveWorkflowPrompt = () => {
  currentWorkflowGeneratorPrompt.value = editableWorkflowPrompt.value
  workflowPromptEditorVisible.value = false
  ElMessage.success('提示词已保存')
}

const resetWorkflowPrompt = () => {
  editableWorkflowPrompt.value = DEFAULT_WORKFLOW_GENERATOR_PROMPT
  ElMessage.info('已恢复默认提示词')
}

const generateWorkflowFromPrompt = async () => {
  if (!workflowGeneratorInput.value.trim()) return

  workflowGenerating.value = true
  await nextTick()

  try {
    const selectedModelData = availableModels.value.find(m => m.id === workflowGeneratorModel.value)

    if (!selectedModelData) {
      ElMessage.error('请选择模型')
      workflowGenerating.value = false
      return
    }

    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        messages: [
          { role: 'user', content: workflowGeneratorInput.value }
        ],
        configId: selectedModelData.id,
        systemPrompts: [currentWorkflowGeneratorPrompt.value]
      })
    })

    if (!response.ok) {
      throw new Error('AI调用失败')
    }

    const reader = response.body?.getReader()
    const decoder = new TextDecoder()
    let aiResponse = ''

    while (reader) {
      const { done, value } = await reader.read()
      if (done) break

      const text = decoder.decode(value)
      const textLines = text.split('\n')

      for (const line of textLines) {
        if (line.startsWith('data: ')) {
          const data = line.substring(6)
          if (data === '[DONE]') continue

          try {
            const json = JSON.parse(data)

            if (json.error) {
              ElMessage.error(`AI调用失败: ${json.error}`)
              throw new Error(json.error)
            }

            if (json.content) {
              aiResponse += json.content
            }
          } catch (e) {
            if (e instanceof Error && e.message.includes('AI调用失败')) {
              throw e
            }
            console.error('解析流式数据失败:', e)
          }
        }
      }
    }

    if (aiResponse) {
      const jsonMatch = aiResponse.match(/```json\s*([\s\S]*?)\s*```/) || aiResponse.match(/```\s*([\s\S]*?)\s*```/)
      const jsonStr = jsonMatch ? jsonMatch[1] : aiResponse

      try {
        const parsed = JSON.parse(jsonStr.trim())
        codeSource.value = JSON.stringify(parsed, null, 2)
        codeError.value = ''
        ElMessage.success('工作流生成成功！请检查后点击"应用更改"')
      } catch {
        codeSource.value = jsonStr.trim()
        codeError.value = '生成的内容不是有效的JSON，请手动修正'
        ElMessage.warning('生成的内容格式有误，请手动修正')
      }
    }
  } catch (e) {
    console.error('生成工作流失败:', e)
    ElMessage.error(`生成失败: ${e instanceof Error ? e.message : String(e)}`)
  } finally {
    workflowGenerating.value = false
  }
}

watch(editorMode, (newMode) => {
  if (newMode === 'code') {
    syncCodeSourceFromStage()
    const defaultModel = getPreferredModel()
    if (defaultModel && !workflowGeneratorModel.value) {
      workflowGeneratorModel.value = defaultModel.id
    }
  }
})

onBeforeUnmount(() => {
  detachAiPromptTextareaScroll()
  stopEdgeDrag()
  stopNodeDrag()
  stopCanvasPan()
  window.removeEventListener('pointerdown', handleGlobalPointerDown)
  window.removeEventListener('resize', handleGlobalViewportChange)
  window.removeEventListener('scroll', handleGlobalViewportChange, true)
})
</script>

<style scoped>
.workflow-workbench-page {
  height: 100vh;
  background: #111827;
  color: #e5e7eb;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.editor-toolbar {
  min-height: 76px;
  padding: 10px 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid #1f2937;
  background: #0f172a;
}

.toolbar-main {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-back {
  width: 40px;
  height: 40px;
  border: 1px solid #334155;
  border-radius: 10px;
  background: #111827;
  color: #e5e7eb;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.toolbar-title {
  min-width: 0;
}

.toolbar-title h1 {
  margin: 0;
  font-size: 22px;
  color: #f9fafb;
}

.toolbar-title p {
  margin: 4px 0 0;
  font-size: 13px;
  color: #94a3b8;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.toolbar-canvas-label {
  color: #e2e8f0;
  font-weight: 600;
}

.toolbar-divider {
  color: #475569;
}

.toolbar-side {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  flex-wrap: wrap;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-workspace-tools {
  justify-content: flex-end;
}

.toolbar-icon-btn {
  width: 40px;
  height: 40px;
  border: 1px solid #334155;
  border-radius: 10px;
  background: #111827;
  color: #e5e7eb;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.toolbar-icon-btn.success {
  color: #4ade80;
}

.toolbar-icon-btn.danger {
  color: #fb7185;
}

.mode-switch-group {
  display: flex;
  align-items: center;
  background: #1f2937;
  border-radius: 10px;
  padding: 3px;
  gap: 2px;
}

.mode-switch-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #94a3b8;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.mode-switch-btn:hover {
  color: #e2e8f0;
  background: rgba(255, 255, 255, 0.05);
}

.mode-switch-btn.active {
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  color: #fff;
  box-shadow: 0 2px 8px rgba(8, 198, 190, 0.3);
}

.mode-switch-btn .el-icon {
  font-size: 16px;
}

.code-editor-layout {
  height: calc(100vh - 76px);
  display: flex;
  flex-direction: row;
  background: #0b1220;
  gap: 0;
}

.code-editor-left {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  border-right: 1px solid #1f2937;
}

.code-editor-right {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.code-editor-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  border: 1px solid #1f2937;
  border-radius: 12px;
  margin: 16px;
  background: #111827;
  overflow: hidden;
}

.workflow-generator-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  border: 1px solid #1f2937;
  border-radius: 12px;
  margin: 16px;
  background: #111827;
  overflow: hidden;
}

.generator-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 18px;
  border-bottom: 1px solid #1f2937;
  background: #0f172a;
}

.generator-header h3 {
  margin: 0;
  font-size: 16px;
  color: #f9fafb;
  font-weight: 600;
}

.generator-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 16px;
  gap: 16px;
  overflow-y: auto;
}

.generator-prompt-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.prompt-label {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #e5e7eb;
  font-size: 14px;
  font-weight: 500;
}

.prompt-label .el-icon {
  color: #08c6be;
}

.generator-input {
  width: 100%;
}

.generator-input :deep(.el-textarea__inner) {
  background: #0b1220;
  border-color: #1f2937;
  color: #e5e7eb;
  font-size: 14px;
  line-height: 1.6;
}

.generator-input :deep(.el-textarea__inner:focus) {
  border-color: #08c6be;
  box-shadow: 0 0 0 2px rgba(8, 198, 190, 0.15);
}

.generator-input :deep(.el-textarea__inner::placeholder) {
  color: #64748b;
}

.generator-options {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.option-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.option-label {
  color: #94a3b8;
  font-size: 13px;
  min-width: 70px;
}

.generator-footer {
  padding: 16px 18px;
  border-top: 1px solid #1f2937;
  background: #0f172a;
  display: flex;
  justify-content: center;
}

.generator-footer .el-button {
  min-width: 160px;
}

.generator-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.workflow-prompt-editor-dialog .prompt-editor-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.workflow-prompt-editor-dialog .prompt-editor-hint {
  margin-bottom: 8px;
}

.workflow-prompt-editor-dialog .prompt-editor-body {
  flex: 1;
}

.workflow-prompt-editor-dialog .prompt-textarea :deep(.el-textarea__inner) {
  background: #0b1220;
  border-color: #1f2937;
  color: #e5e7eb;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
}

.workflow-prompt-editor-dialog .prompt-textarea :deep(.el-textarea__inner:focus) {
  border-color: #08c6be;
}

.workflow-prompt-editor-dialog .prompt-editor-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.code-editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 18px;
  border-bottom: 1px solid #1f2937;
  background: #0f172a;
}

.code-editor-header h3 {
  margin: 0;
  font-size: 16px;
  color: #f9fafb;
  font-weight: 600;
}

.code-editor-actions {
  display: flex;
  gap: 10px;
}

.code-editor-body {
  flex: 1;
  min-height: 0;
  padding: 16px;
}

.code-textarea {
  width: 100%;
  height: 100%;
  min-height: 400px;
  padding: 16px;
  border: 1px solid #1f2937;
  border-radius: 8px;
  background: #0b1220;
  color: #e5e7eb;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  resize: none;
  outline: none;
  box-sizing: border-box;
}

.code-textarea:focus {
  border-color: #08c6be;
  box-shadow: 0 0 0 2px rgba(8, 198, 190, 0.15);
}

.code-textarea::placeholder {
  color: #64748b;
}

.code-error-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 18px;
  background: rgba(239, 68, 68, 0.1);
  border-top: 1px solid rgba(239, 68, 68, 0.3);
  color: #f87171;
  font-size: 13px;
}

.code-error-bar .el-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.editor-layout {
  height: calc(100vh - 76px);
  display: grid;
  grid-template-columns: 248px minmax(0, 1fr);
  min-height: 0;
  overflow: hidden;
}

.sidebar {
  background: #0b1220;
  border-right: 1px solid #1f2937;
  display: flex;
  flex-direction: column;
  min-height: 0;
  position: relative;
  overflow: hidden;
}

.sidebar.collapsed {
  overflow: hidden;
}

.sidebar-head {
  padding: 20px 18px 14px;
  border-bottom: 1px solid #1f2937;
  flex-shrink: 0;
}

.sidebar-head h2 {
  margin: 0;
  font-size: 18px;
  color: #f9fafb;
}

.sidebar-head span {
  display: inline-block;
  margin-top: 6px;
  font-size: 12px;
  color: #94a3b8;
}

.sidebar-list {
  flex: 1;
  min-height: 0;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
}

.library-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.library-group-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #e5e7eb;
  padding: 8px 10px;
  background: #1f2937;
  border-radius: 8px;
  margin-bottom: 2px;
}

.library-group-title .el-icon {
  color: #fbbf24;
}

.library-item {
  padding: 10px 12px;
  border: 1px solid #1f2937;
  border-left: 3px solid var(--accent);
  border-radius: 8px;
  background: #111827;
  cursor: grab;
  transition:
    transform 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease;
}

.library-item:hover {
  transform: translateY(-1px);
  border-color: color-mix(in srgb, var(--accent) 42%, #334155);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.28);
}

.library-item:active {
  cursor: grabbing;
}

.library-item-title {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.library-icon {
  width: 24px;
  height: 24px;
  flex-shrink: 0;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--accent);
  background: color-mix(in srgb, var(--accent) 18%, transparent);
  border: 1px solid color-mix(in srgb, var(--accent) 30%, #1f2937);
}

.library-item-title strong {
  font-size: 13px;
  color: #f9fafb;
  font-weight: 500;
}

.library-item p {
  margin: 10px 0 0;
  font-size: 13px;
  line-height: 1.6;
  color: #cbd5e1;
}

.library-item-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.library-copy-btn {
  height: 28px;
  padding: 0 10px;
  border: 1px solid #334155;
  border-radius: 8px;
  background: #0f172a;
  color: #cbd5e1;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
  cursor: pointer;
  transition:
    border-color 0.18s ease,
    color 0.18s ease,
    background-color 0.18s ease;
}

.library-copy-btn:hover {
  border-color: color-mix(in srgb, var(--accent) 45%, #38bdf8);
  color: #f8fafc;
  background: #111827;
}

.library-copy-btn.danger:hover {
  border-color: rgba(251, 113, 133, 0.58);
  color: #fff1f2;
  background: rgba(159, 18, 57, 0.16);
}

.library-copy-btn span {
  font-size: 12px;
  line-height: 1;
}

.workspace {
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: #111827;
  overflow: hidden;
}

.workspace-body {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 288px;
}

.workspace-stats {
  display: flex;
  align-items: center;
  gap: 8px;
}

.workspace-tools {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.workspace-stats span {
  padding: 6px 10px;
  border-radius: 999px;
  background: #0f172a;
  border: 1px solid #334155;
  font-size: 12px;
  color: #cbd5e1;
}

.workspace-scroll {
  flex: 1;
  overflow: hidden;
  padding: 18px;
  position: relative;
  cursor: grab;
  background-color: #1f2937;
  background-image:
    linear-gradient(rgba(71, 85, 105, 0.24) 1px, transparent 1px),
    linear-gradient(90deg, rgba(71, 85, 105, 0.24) 1px, transparent 1px);
}

.workspace-scroll.panning {
  cursor: grabbing;
}

.workspace-scroll.drop-ready {
  box-shadow: inset 0 0 0 1px rgba(125, 211, 252, 0.3);
}

.stage-camera {
  position: absolute;
  left: 0;
  top: 0;
}

.stage {
  position: absolute;
  left: 0;
  top: 0;
  transform-origin: top left;
}

.stage-grid {
  display: none;
}

.edge-layer {
  position: absolute;
  inset: 0;
}

.edge-path {
  fill: none;
  stroke: #7dd3fc;
  stroke-width: 3;
  pointer-events: none;
}

.edge-hit-path {
  fill: none;
  stroke: transparent;
  stroke-width: 18;
  pointer-events: stroke;
  cursor: pointer;
}

.edge-path.loop {
  stroke: #c4b5fd;
}

.edge-path.preview {
  stroke: #38bdf8;
  stroke-dasharray: 10 8;
  opacity: 0.9;
}

.edge-path.running {
  stroke: #22c55e;
  stroke-width: 3;
  stroke-dasharray: 12 8;
  animation: edge-flow 0.6s linear infinite;
  filter: drop-shadow(0 0 6px rgba(34, 197, 94, 0.6));
}

@keyframes edge-flow {
  from {
    stroke-dashoffset: 20;
  }
  to {
    stroke-dashoffset: 0;
  }
}

.stage-node {
  position: absolute;
  height: var(--node-height, 132px);
  padding: 8px 10px;
  border-radius: 10px;
  border: 1px solid #334155;
  border-left: 3px solid var(--accent);
  background: #0f172a;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.22);
  overflow: hidden;
  cursor: grab;
  touch-action: none;
  user-select: none;
}

.stage-node.dragging {
  cursor: grabbing;
  box-shadow: 0 16px 30px rgba(14, 165, 233, 0.18);
}

.stage-node.selected {
  border-color: color-mix(in srgb, var(--accent) 55%, #334155);
  box-shadow:
    0 0 0 1px color-mix(in srgb, var(--accent) 45%, transparent),
    0 16px 30px rgba(14, 165, 233, 0.14);
}

.stage-node.connect-ready {
  box-shadow:
    0 0 0 1px rgba(125, 211, 252, 0.18),
    0 12px 28px rgba(56, 189, 248, 0.1);
}

.stage-node.connect-hover {
  border-color: rgba(56, 189, 248, 0.78);
  box-shadow:
    0 0 0 1px rgba(56, 189, 248, 0.44),
    0 18px 36px rgba(56, 189, 248, 0.18);
}

.stage-node.running {
  border-color: #22c55e;
  border-left-color: #22c55e;
  box-shadow:
    0 0 0 1px rgba(34, 197, 94, 0.35),
    0 18px 36px rgba(34, 197, 94, 0.15);
  animation: node-pulse 1.5s ease-in-out infinite;
}

@keyframes node-pulse {
  0%,
  100% {
    box-shadow:
      0 0 0 1px rgba(34, 197, 94, 0.35),
      0 18px 36px rgba(34, 197, 94, 0.15);
  }
  50% {
    box-shadow:
      0 0 0 2px rgba(34, 197, 94, 0.5),
      0 18px 36px rgba(34, 197, 94, 0.2);
  }
}

.stage-node-header {
  display: flex;
  align-items: center;
  gap: 6px;
}

.stage-node-icon {
  width: 20px;
  height: 20px;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--accent);
  background: color-mix(in srgb, var(--accent) 18%, transparent);
  border: 1px solid color-mix(in srgb, var(--accent) 30%, #334155);
  font-size: 12px;
  flex-shrink: 0;
}

.stage-node h3 {
  margin: 0;
  font-size: 13px;
  color: #f9fafb;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.stage-node p {
  margin: 4px 0 0;
  font-size: 11px;
  line-height: 1.4;
  color: #94a3b8;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.stage-node-context-menu {
  position: fixed;
  z-index: 40;
  min-width: 132px;
  padding: 6px;
  border: 1px solid #334155;
  border-radius: 12px;
  background: rgba(15, 23, 42, 0.96);
  box-shadow: 0 18px 36px rgba(2, 6, 23, 0.42);
  backdrop-filter: blur(12px);
  display: grid;
  gap: 4px;
}

.stage-node-context-item {
  height: 34px;
  padding: 0 12px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #e5e7eb;
  text-align: left;
  font-size: 13px;
  cursor: pointer;
  transition:
    background-color 0.18s ease,
    color 0.18s ease;
}

.stage-node-context-item:hover {
  background: #1e293b;
  color: #f8fafc;
}

.stage-node-context-item.danger:hover {
  background: rgba(127, 29, 29, 0.28);
  color: #fecaca;
}

.node-tags {
  margin-top: 6px;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.node-tags span {
  padding: 2px 6px;
  border-radius: 4px;
  background: #1e293b;
  font-size: 10px;
  color: #94a3b8;
}

.handle {
  position: absolute;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: var(--accent);
  border: 2px solid #111827;
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease,
    background-color 0.18s ease;
}

.handle::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: transparent;
}

.handle.target {
  left: -7px;
  top: 50%;
  transform: translateY(-50%);
  background: #e2e8f0;
  cursor: copy;
}

.handle.source {
  right: -7px;
  top: 50%;
  transform: translateY(-50%);
  cursor: crosshair;
}

.handle:hover {
  box-shadow: 0 0 0 6px rgba(125, 211, 252, 0.2);
  transform: translateY(-50%) scale(1.2);
}

.handle.target:hover {
  transform: translateY(-50%) scale(1.2);
}

.handle.source:hover {
  transform: translateY(-50%) scale(1.2);
}

.handle-main {
  top: 50%;
  transform: translateY(-50%);
}

.handle-loop-body {
  top: 28px;
  transform: none;
}

.handle-loop-body:hover {
  transform: scale(1.2);
}

.handle-loop-done {
  top: 60px;
  transform: none;
}

.handle-loop-done:hover {
  transform: scale(1.2);
}

.resize-handle {
  position: absolute;
  right: 2px;
  bottom: 2px;
  width: 10px;
  height: 10px;
  cursor: nwse-resize;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.resize-handle::before,
.resize-handle::after {
  content: '';
  position: absolute;
  background: #64748b;
  border-radius: 1px;
}

.resize-handle::before {
  right: 0;
  bottom: 3px;
  width: 8px;
  height: 2px;
  transform: rotate(-45deg);
  transform-origin: right bottom;
}

.resize-handle::after {
  right: 0;
  bottom: 0;
  width: 5px;
  height: 2px;
  transform: rotate(-45deg);
  transform-origin: right bottom;
}

.stage-node:hover .resize-handle,
.stage-node.selected .resize-handle {
  opacity: 1;
}

.resize-handle:hover::before,
.resize-handle:hover::after {
  background: #3b82f6;
}

.missing-state {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
}

.zoom-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.zoom-btn,
.zoom-value-btn {
  height: 34px;
  border: 1px solid #334155;
  border-radius: 8px;
  background: #0f172a;
  color: #e5e7eb;
  cursor: pointer;
}

.zoom-btn {
  width: 34px;
}

.zoom-value-btn {
  min-width: 72px;
  padding: 0 12px;
}

.inspector {
  border-left: 1px solid #1f2937;
  background: #0b1220;
  padding: 14px;
  overflow-y: auto;
  position: relative;
}

.inspector.collapsed {
  overflow: hidden;
  padding: 0;
}

.inspector-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.inspector-title {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.inspector-title-meta {
  min-width: 0;
  display: grid;
  gap: 4px;
}

.inspector-icon {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--accent);
  background: color-mix(in srgb, var(--accent) 18%, transparent);
  border: 1px solid color-mix(in srgb, var(--accent) 28%, #334155);
}

.inspector-title h3 {
  margin: 0;
  font-size: 18px;
  color: #f9fafb;
}

.inspector-title p {
  margin: 0;
  font-size: 11px;
  color: #94a3b8;
  line-height: 1.4;
}

.inspector-section {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid #1f2937;
}

.inspector-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.section-label {
  display: inline-block;
  margin-bottom: 8px;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #7dd3fc;
}

.section-label-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
}

.section-label-row .section-label {
  margin-bottom: 0;
}

.section-help-icon {
  font-size: 14px;
  color: #6b7280;
  cursor: help;
  transition: color 0.2s ease;
}

.section-help-icon:hover {
  color: #9ca3af;
}

.output-field-tooltip {
  font-size: 13px;
  line-height: 1.5;
}

.output-field-tooltip p {
  margin: 0 0 4px 0;
}

.output-field-tooltip code {
  display: block;
  padding: 6px 10px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  color: #22d3ee;
}

.inspector-section-head .section-label {
  margin-bottom: 0;
}

.inspector-section p {
  margin: 0;
  font-size: 13px;
  line-height: 1.7;
  color: #cbd5e1;
}

.section-intro {
  display: none;
}

.inspector-help-btn {
  width: 22px;
  height: 22px;
  margin-right: auto;
  padding: 0;
  border: none;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: color-mix(in srgb, #7dd3fc 18%, transparent);
  color: #cfeeff;
  cursor: help;
  transition:
    background-color 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease;
}

.inspector-help-btn:hover {
  background: color-mix(in srgb, #7dd3fc 28%, transparent);
  color: #ffffff;
  transform: translateY(-1px);
}

.inspector-help-btn:focus-visible {
  outline: 2px solid color-mix(in srgb, #7dd3fc 72%, #ffffff);
  outline-offset: 2px;
}

.inspector-help-content {
  display: grid;
  gap: 8px;
  color: #475569;
}

.inspector-help-content p {
  margin: 0;
}

.inspector-help-content code {
  padding: 2px 6px;
  border-radius: 6px;
  background: #e2e8f0;
  color: #0f172a;
  font-size: 12px;
}

.inspector-help-details {
  display: grid;
  gap: 10px;
}

.inspector-help-detail {
  padding: 10px 12px;
  border-radius: 10px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  display: grid;
  gap: 4px;
}

.inspector-help-detail span {
  font-size: 12px;
  color: #64748b;
}

.inspector-help-detail strong {
  font-size: 13px;
  line-height: 1.5;
  color: #0f172a;
}

.inspector-section:has(.node-title-trigger) {
  display: none;
}

.node-title-trigger {
  width: auto;
  max-width: 152px;
  min-height: 28px;
  padding: 0 10px;
  border: 1px solid #334155;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #0f172a;
  color: #f8fafc;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    transform 0.2s ease;
}

.node-title-trigger:hover {
  border-color: #38bdf8;
  background: #111827;
  transform: translateY(-1px);
}

.node-title-trigger span {
  max-width: 132px;
  overflow: hidden;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  font-weight: 500;
}

.inspector-title-trigger {
  max-width: 168px;
  min-height: 30px;
  padding: 0 12px;
  justify-content: flex-start;
}

.inspector-title-trigger span {
  max-width: 142px;
  text-align: left;
  font-size: 13px;
  font-weight: 600;
}

.node-title-editor {
  display: grid;
  gap: 8px;
}

.node-title-editor span {
  font-size: 12px;
  color: #64748b;
}

.ai-compact-section {
  margin-top: 12px;
  padding-top: 12px;
}

.ai-compact-stack {
  gap: 8px;
}

.ai-output-preview {
  border: 1px solid #334155;
  border-radius: 8px;
  background: #0f1419;
  min-height: 120px;
  max-height: 300px;
  overflow: auto;
}

.ai-output-preview .output-content {
  padding: 12px;
}

.ai-output-preview .output-content :deep(.markdown-body) {
  color: #e5e7eb;
  font-size: 13px;
  line-height: 1.6;
}

.ai-output-preview .output-content :deep(.markdown-body h1),
.ai-output-preview .output-content :deep(.markdown-body h2),
.ai-output-preview .output-content :deep(.markdown-body h3),
.ai-output-preview .output-content :deep(.markdown-body h4) {
  color: #f3f4f6;
  margin-top: 16px;
  margin-bottom: 8px;
}

.ai-output-preview .output-content :deep(.markdown-body p) {
  margin: 8px 0;
}

.ai-output-preview .output-content :deep(.markdown-body code) {
  background: #1f2937;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.ai-output-preview .output-content :deep(.markdown-body pre) {
  background: #1f2937;
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
}

.ai-output-preview .output-content :deep(.markdown-body pre code) {
  background: transparent;
  padding: 0;
}

.ai-output-preview .output-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 120px;
  color: #6b7280;
  font-size: 13px;
}

.ai-output-preview .output-placeholder .el-icon {
  font-size: 24px;
}

.context-fields-preview {
  background: #0c0c0c;
  border: 1px solid #1f2937;
  border-radius: 8px;
  min-height: 80px;
}

.output-fields-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
}

.output-field-item {
  background: #111827;
  border: 1px solid #1f2937;
  border-radius: 6px;
  padding: 10px 12px;
}

.output-field-hint {
  margin: 0;
  padding: 8px 12px;
  font-size: 11px;
  color: #6b7280;
  border-top: 1px solid #1f2937;
}

.context-fields-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px;
  color: #6b7280;
  font-size: 13px;
}

.context-fields-empty .el-icon {
  font-size: 20px;
}

.context-fields-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  max-height: 200px;
  overflow-y: auto;
}

.context-fields-list::-webkit-scrollbar {
  width: 4px;
}

.context-fields-list::-webkit-scrollbar-track {
  background: transparent;
}

.context-fields-list::-webkit-scrollbar-thumb {
  background: #374151;
  border-radius: 2px;
}

.context-field-item {
  background: #111827;
  border: 1px solid #1f2937;
  border-radius: 6px;
  padding: 10px 12px;
}

.field-name {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
  font-size: 12px;
  font-weight: 500;
  color: #60a5fa;
}

.field-name .el-icon {
  font-size: 14px;
}

.field-id {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  color: #22d3ee;
  font-weight: 600;
}

.field-value {
  font-size: 12px;
  color: #9ca3af;
  line-height: 1.5;
  word-break: break-word;
}

.output-preview-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.output-action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.2s ease;
}

.output-action-btn:hover {
  background: rgba(148, 163, 184, 0.15);
  color: #e5e7eb;
}

.output-action-btn.active {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
}

.output-action-btn.continue-btn {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
  animation: pulse-continue 1.5s ease-in-out infinite;
}

@keyframes pulse-continue {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.4);
  }
  50% {
    box-shadow: 0 0 0 6px rgba(34, 197, 94, 0);
  }
}

.output-preview-dialog-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.output-preview-dialog-head h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #f3f4f6;
}

.output-preview-dialog-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.output-preview-dialog-actions .output-action-btn {
  width: auto;
  padding: 0 12px;
  gap: 4px;
  font-size: 13px;
}

.output-preview-dialog-body {
  min-height: 300px;
  max-height: 60vh;
  overflow: auto;
}

.output-preview-content {
  padding: 16px;
  background: #0f1419;
  border-radius: 8px;
  border: 1px solid #334155;
}

.output-preview-content :deep(.markdown-body) {
  color: #e5e7eb;
  font-size: 14px;
  line-height: 1.7;
}

.output-preview-content :deep(.markdown-body h1),
.output-preview-content :deep(.markdown-body h2),
.output-preview-content :deep(.markdown-body h3),
.output-preview-content :deep(.markdown-body h4) {
  color: #f3f4f6;
  margin-top: 20px;
  margin-bottom: 12px;
}

.output-preview-content :deep(.markdown-body p) {
  margin: 12px 0;
}

.output-preview-content :deep(.markdown-body code) {
  background: #1f2937;
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 13px;
}

.output-preview-content :deep(.markdown-body pre) {
  background: #1f2937;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
}

.output-preview-content :deep(.markdown-body pre code) {
  background: transparent;
  padding: 0;
}

.output-edit-area {
  padding: 16px;
  background: #0f1419;
  border-radius: 8px;
  border: 1px solid #334155;
}

.output-editor-textarea :deep(.el-textarea__inner) {
  background: #1f2937;
  border-color: #374151;
  color: #e5e7eb;
  font-size: 14px;
  line-height: 1.7;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.output-editor-textarea :deep(.el-textarea__inner::placeholder) {
  color: #6b7280;
}

.output-editor-textarea :deep(.el-textarea__inner:focus) {
  border-color: #3b82f6;
}

.output-preview-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  height: 300px;
  color: #6b7280;
  font-size: 14px;
}

.output-preview-empty .el-icon {
  font-size: 32px;
}

.ai-prompt-editor-shell {
  position: relative;
  border-radius: 12px;
  overflow: hidden;
}

.ai-prompt-editor-shell :deep(.el-textarea) {
  display: block;
}

.ai-prompt-editor-shell :deep(.el-textarea__inner) {
  position: relative;
  z-index: 2;
  background: transparent;
  color: transparent;
  caret-color: #e5e7eb;
  line-height: 1.7;
  resize: none;
}

.ai-prompt-editor-shell :deep(.el-textarea__inner::selection) {
  background: rgba(56, 189, 248, 0.28);
}

.ai-prompt-editor-shell :deep(.el-textarea__inner::placeholder) {
  color: transparent;
}

.ai-prompt-overlay {
  position: absolute;
  inset: 1px;
  z-index: 1;
  overflow: hidden;
  border-radius: 11px;
  pointer-events: none;
}

.ai-prompt-overlay-content {
  min-height: 100%;
  padding: 9px 11px;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 14px;
  line-height: 1.7;
  color: #e5e7eb;
}

.ai-prompt-overlay-text {
  color: #dbe4f0;
}

.ai-prompt-overlay-placeholder {
  color: #64748b;
}

.ai-prompt-inline-chip {
  margin: 0 3px;
  padding: 2px 8px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  vertical-align: middle;
  background: #0f172a;
  border: 1px solid #334155;
  color: #e5e7eb;
}

.ai-prompt-inline-chip span {
  font-size: 11px;
  color: #94a3b8;
}

.ai-prompt-inline-chip strong {
  font-size: 12px;
  font-weight: 600;
  color: #f8fafc;
}

.ai-prompt-inline-chip.field {
  border-color: rgba(125, 211, 252, 0.36);
  background: rgba(14, 116, 144, 0.18);
}

.ai-prompt-inline-chip.prompt {
  border-color: rgba(253, 186, 116, 0.36);
  background: rgba(154, 52, 18, 0.2);
}

.ai-prompt-action-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.ai-prompt-action-btn {
  min-height: 30px;
  padding: 0 12px;
  border: 1px solid #334155;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #0f172a;
  color: #e5e7eb;
  cursor: pointer;
  font-size: 12px;
  line-height: 1;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease;
}

.ai-prompt-action-btn:hover {
  border-color: #38bdf8;
  background: #111827;
  color: #f8fafc;
}

.ai-config-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.ai-config-pill {
  min-height: 32px;
  padding: 0 12px;
  border: 1px solid #334155;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #0f172a;
  color: #e5e7eb;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    transform 0.2s ease;
}

.ai-config-pill:hover {
  border-color: #38bdf8;
  background: #111827;
  transform: translateY(-1px);
}

.ai-config-pill span {
  font-size: 12px;
  color: #94a3b8;
}

.ai-config-pill strong {
  font-size: 12px;
  font-weight: 600;
  color: #f8fafc;
}

.ai-config-editor {
  display: grid;
  gap: 12px;
}

.text-input-file-import {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.text-input-fields-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 8px;
}

.text-input-field-item {
  padding: 12px;
  border: 1px solid #1f2937;
  border-radius: 8px;
  background: #0f172a;
}

.field-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.field-name-input {
  flex: 1;
}

.field-name-input :deep(.el-input__wrapper) {
  background: #111827;
  border-color: #1f2937;
}

.field-config-btn {
  padding: 6px;
  border: 1px solid #1f2937;
  border-radius: 6px;
  background: #111827;
  color: #6b7280;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.field-config-btn:hover {
  background: #1f2937;
  color: #e5e7eb;
}

.field-delete-btn {
  padding: 6px;
  border: 1px solid #1f2937;
  border-radius: 6px;
  background: #111827;
  color: #6b7280;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.field-delete-btn:hover {
  background: #7f1d1d;
  border-color: #991b1b;
  color: #fca5a5;
}

.field-content-input :deep(.el-textarea__inner) {
  background: #111827;
  border-color: #1f2937;
  color: #e5e7eb;
}

.field-config-popover {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.field-config-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.field-config-row span {
  font-size: 13px;
  color: #94a3b8;
  white-space: nowrap;
}

.field-config-row .el-select,
.field-config-row .el-input {
  flex: 1;
}

.field-file-import {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.imported-file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border: 1px solid #1f2937;
  border-radius: 8px;
  background: #111827;
  font-size: 13px;
  color: #e5e7eb;
}

.imported-file-info .el-icon {
  color: #3b82f6;
}

.clear-import-btn {
  margin-left: auto;
  padding: 4px;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: #6b7280;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.clear-import-btn:hover {
  background: #1f2937;
  color: #ef4444;
}

.ai-config-editor-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  font-size: 12px;
  color: #64748b;
}

.ai-config-editor-head strong {
  font-size: 13px;
  color: #0f172a;
}

.ai-config-editor :deep(.el-slider) {
  padding-right: 0;
}

.ai-config-editor :deep(.el-input-number) {
  width: 100%;
}

.ai-temp-editor :deep(.el-slider) {
  display: flex;
  align-items: center;
  gap: 12px;
}

.ai-temp-editor :deep(.el-slider__runway) {
  flex: 1;
  margin: 0;
}

.ai-temp-editor :deep(.el-slider__input) {
  flex: 0 0 84px;
  width: 84px;
  margin-top: 0;
}

.ai-temp-editor :deep(.el-slider__input .el-input-number) {
  width: 84px;
}

.ai-temp-editor :deep(.el-input-number__increase),
.ai-temp-editor :deep(.el-input-number__decrease) {
  width: 20px;
}

.ai-prompt-editor-section .section-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  user-select: none;
  padding: 4px 0;
  margin-bottom: 8px;
}

.ai-prompt-editor-section .section-label-row:hover {
  opacity: 0.8;
}

.ai-prompt-editor-section .collapse-icon {
  transition: transform 0.2s ease;
  font-size: 12px;
  color: #94a3b8;
}

.ai-prompt-editor-section.collapsed .collapse-icon {
  transform: rotate(-90deg);
}

.ai-prompt-editor-section + .inspector-section,
.ai-prompt-editor-section + .inspector-section + .inspector-section,
.ai-prompt-editor-section + .inspector-section + .inspector-section + .inspector-section {
  display: none;
}

.inspector:has(.ai-compact-params-anchor) .inspector-section {
  margin-top: 12px;
  padding-top: 12px;
}

.inspector:has(.ai-compact-params-anchor) .section-label {
  margin-bottom: 6px;
}

.inspector:has(.ai-compact-params-anchor) .inspector-field-stack {
  gap: 8px;
}

.inspector-field-stack {
  display: grid;
  gap: 10px;
}

.start-fields-list {
  display: grid;
  gap: 14px;
}

.start-field-item {
  padding: 14px;
  border-radius: 14px;
  border: 1px solid #1f2937;
  background: #111827;
  display: grid;
  gap: 12px;
}

.start-field-summary-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.start-field-toggle {
  width: 28px;
  height: 28px;
  padding: 0;
  border: 1px solid #334155;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #0f172a;
  color: #cbd5e1;
  cursor: pointer;
  flex-shrink: 0;
}

.start-field-summary {
  min-width: 0;
  flex: 1;
  display: block;
}

.start-field-copy-btn {
  width: 28px;
  height: 28px;
  padding: 0;
  border: 1px solid #334155;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #0f172a;
  color: #7dd3fc;
  cursor: pointer;
  flex-shrink: 0;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease;
}

.start-field-copy-btn:hover {
  border-color: #38bdf8;
  background: #111827;
  color: #e0f2fe;
}

.start-field-summary strong {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  color: #f8fafc;
}

.start-field-editor-body {
  display: grid;
  gap: 12px;
}

.start-field-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.start-field-name {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
}

.start-field-editor-body .start-field-name :deep(.el-button) {
  display: none;
}

.start-field-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.start-field-form {
  display: grid;
  gap: 10px;
}

.start-field-row {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(120px, 1fr);
  gap: 10px;
}

.start-field-options {
  display: grid;
  gap: 8px;
}

.start-field-options-label {
  font-size: 12px;
  color: #94a3b8;
}

.required-tag {
  flex-shrink: 0;
}

.start-preview-list {
  display: grid;
  gap: 12px;
}

.start-preview-item {
  display: grid;
  gap: 8px;
}

.start-preview-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #cbd5e1;
}

.start-preview-label small {
  color: #fca5a5;
}

.start-preview-desc {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
  color: #94a3b8;
}

.prompt-field-editor {
  display: grid;
  gap: 8px;
}

.prompt-field-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #cbd5e1;
}

.prompt-field-label small {
  color: #fca5a5;
}

.prompt-field-desc {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
  color: #94a3b8;
}

.inspector-section :deep(.el-select),
.inspector-section :deep(.el-input),
.inspector-section :deep(.el-input-number) {
  width: 100%;
}

.inspector-section :deep(.el-slider) {
  padding-right: 12px;
}

.inspector-section :deep(.el-input__wrapper),
.inspector-section :deep(.el-select__wrapper),
.inspector-section :deep(.el-textarea__inner),
.inspector-section :deep(.el-input-number),
.inspector-section :deep(.el-input-number__increase),
.inspector-section :deep(.el-input-number__decrease) {
  background: #0f172a;
  box-shadow: 0 0 0 1px #334155 inset;
  color: #e5e7eb;
}

.inspector-section :deep(.el-input__inner),
.inspector-section :deep(.el-select__selected-item),
.inspector-section :deep(.el-textarea__inner),
.inspector-section :deep(.el-input-number .el-input__inner) {
  color: #e5e7eb;
}

.inspector-section :deep(.el-input__inner::placeholder),
.inspector-section :deep(.el-textarea__inner::placeholder) {
  color: #64748b;
}

.inspector-section :deep(.el-input-number__increase),
.inspector-section :deep(.el-input-number__decrease) {
  border-color: #334155;
}

.inspector-section :deep(.el-select__wrapper.is-focused),
.inspector-section :deep(.el-input__wrapper.is-focus),
.inspector-section :deep(.el-textarea__inner:focus),
.inspector-section :deep(.el-input-number.is-controls-right .el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px rgba(56, 189, 248, 0.45) inset;
}

.inspector-chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.inspector-chip {
  padding: 6px 10px;
  border-radius: 999px;
  background: #111827;
  border: 1px solid #334155;
  font-size: 12px;
  color: #e5e7eb;
}

.config-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.config-item {
  padding: 10px 12px;
  border-radius: 10px;
  background: #111827;
  border: 1px solid #1f2937;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.config-item span {
  font-size: 12px;
  color: #94a3b8;
}

.config-item strong {
  font-size: 13px;
  font-weight: 600;
  color: #f8fafc;
  line-height: 1.6;
}

.inspector-empty {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  color: #94a3b8;
}

.inspector-empty .el-icon {
  font-size: 34px;
  margin-bottom: 14px;
  color: #7dd3fc;
}

.inspector-empty h3 {
  margin: 0 0 8px;
  color: #f8fafc;
  font-size: 18px;
}

.inspector-empty p {
  margin: 0;
  line-height: 1.7;
}

:deep(.workflow-reference-popover) {
  border: 1px solid #1f2937 !important;
  border-radius: 10px !important;
  background: #0b1220 !important;
  box-shadow: 0 16px 32px rgba(0, 0, 0, 0.45) !important;
  max-height: 400px !important;
  overflow: hidden !important;
}

:deep(.workflow-reference-popover .el-popper__arrow) {
  display: none !important;
}

.workflow-reference-menu {
  display: grid;
  gap: 4px;
  padding: 6px;
  max-height: 320px;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: #374151 transparent;
}

.workflow-reference-menu::-webkit-scrollbar {
  width: 4px;
}

.workflow-reference-menu::-webkit-scrollbar-track {
  background: transparent;
}

.workflow-reference-menu::-webkit-scrollbar-thumb {
  background: #374151;
  border-radius: 2px;
}

.workflow-reference-item {
  width: 100%;
  padding: 6px 10px;
  border: 1px solid #1f2937;
  border-radius: 5px;
  background: #111827;
  color: #cbd5e1;
  cursor: pointer;
  text-align: left;
  font-size: 12px;
  transition:
    border-color 0.15s ease,
    background-color 0.15s ease;
}

.workflow-reference-item:hover {
  border-color: #3b82f6;
  background: #1e293b;
}

.workflow-reference-empty {
  margin: 0;
  font-size: 12px;
  color: #64748b;
  padding: 16px 8px;
  text-align: center;
}

.workflow-prompt-dialog-head {
  padding: 16px 18px;
  border: 1px solid #000000;
  border-radius: 16px;
  background: #000000;
}

.workflow-prompt-dialog-head h3 {
  margin: 0;
  font-size: 18px;
  color: #dde5ee;
}

.workflow-prompt-dialog-head p {
  margin: 6px 0 0;
  font-size: 12px;
  line-height: 1.6;
  color: #8c98a7;
}

.workflow-prompt-dialog-body {
  display: grid;
  gap: 16px;
}

.workflow-prompt-category-panel {
  padding: 14px;
  border: 1px solid #000000;
  border-radius: 18px;
  display: grid;
  gap: 14px;
  background:
    linear-gradient(180deg, rgba(40, 52, 65, 0.28) 0%, rgba(25, 33, 42, 0.96) 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.03),
    0 12px 30px rgba(5, 10, 18, 0.18);
}

.workflow-prompt-category-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.workflow-prompt-category-head strong {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: #e6ecf3;
}

.workflow-prompt-category-head p {
  margin: 6px 0 0;
  font-size: 12px;
  line-height: 1.6;
  color: #8793a3;
}

.workflow-prompt-category-count {
  flex-shrink: 0;
  min-height: 28px;
  padding: 0 10px;
  border: 1px solid #000000;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  background: rgba(22, 29, 36, 0.82);
  color: #9eabb9;
  font-size: 11px;
}

.workflow-prompt-category-bar {
  display: flex;
  flex-wrap: nowrap;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 2px;
  scrollbar-width: thin;
  scrollbar-color: rgba(110, 127, 145, 0.45) transparent;
}

.workflow-prompt-category-bar::-webkit-scrollbar {
  height: 6px;
}

.workflow-prompt-category-bar::-webkit-scrollbar-track {
  background: transparent;
}

.workflow-prompt-category-bar::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: rgba(110, 127, 145, 0.35);
}

.workflow-prompt-category-bar::-webkit-scrollbar-thumb:hover {
  background: rgba(130, 148, 167, 0.45);
}

.workflow-prompt-category-btn {
  flex-shrink: 0;
  min-height: 38px;
  padding: 0 14px;
  border: 1px solid #000000;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background:
    linear-gradient(180deg, rgba(34, 43, 54, 0.96) 0%, rgba(24, 31, 39, 0.96) 100%);
  color: #c3ccd7;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.02);
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease,
    box-shadow 0.2s ease;
}

.workflow-prompt-category-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: rgba(150, 166, 182, 0.52);
  box-shadow: 0 0 0 4px rgba(150, 166, 182, 0.08);
}

.workflow-prompt-category-btn:hover,
.workflow-prompt-category-btn.active {
  border-color: rgba(112, 141, 166, 0.88);
  background:
    linear-gradient(180deg, rgba(45, 58, 72, 0.98) 0%, rgba(31, 40, 49, 0.98) 100%);
  color: #eef3f8;
  transform: translateY(-1px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.04),
    0 10px 20px rgba(7, 13, 22, 0.18);
}

.workflow-prompt-category-btn:hover .workflow-prompt-category-dot,
.workflow-prompt-category-btn.active .workflow-prompt-category-dot {
  background: #8eb2cd;
  box-shadow: 0 0 0 4px rgba(142, 178, 205, 0.16);
}

.workflow-prompt-list {
  display: grid;
  gap: 10px;
  max-height: 420px;
  overflow-y: auto;
}

.workflow-prompt-item {
  padding: 14px;
  border: 1px solid #2f3b49;
  border-radius: 14px;
  display: grid;
  gap: 8px;
  background: #1a222c;
  color: #d6dde6;
  text-align: left;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    transform 0.2s ease;
}

.workflow-prompt-item:hover {
  border-color: #5d7c93;
  background: #212b35;
  transform: translateY(-1px);
}

.workflow-prompt-item-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.workflow-prompt-item-head strong {
  font-size: 14px;
  color: #e5ebf2;
}

.workflow-prompt-item-head span {
  flex-shrink: 0;
  font-size: 11px;
  color: #9ec2df;
}

.workflow-prompt-item p {
  margin: 0;
  font-size: 13px;
  line-height: 1.7;
  color: #8c98a7;
}

.workflow-prompt-empty {
  min-height: 160px;
  border: 1px dashed #33404b;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(24, 31, 40, 0.72);
  color: #8c98a7;
  font-size: 13px;
}

.panel-toggle-wrap {
  position: absolute;
  top: 14px;
  right: 10px;
  z-index: 5;
}

.inspector-toggle-wrap {
  top: 14px;
  left: 10px;
  right: auto;
}

.panel-toggle-btn {
  width: 28px;
  height: 28px;
  border: 1px solid #334155;
  border-radius: 8px;
  background: #111827;
  color: #cbd5e1;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.collapsed-rail-label {
  height: 100%;
  writing-mode: vertical-rl;
  text-orientation: mixed;
  letter-spacing: 0.14em;
  color: #94a3b8;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  user-select: none;
}

.library-preview-body {
  display: grid;
  gap: 0;
  background: #0b1220;
  color: #cbd5e1;
  border-radius: 16px;
}

.library-preview-body .inspector-section {
  background: #0b1220;
}

.library-preview-head :deep(.el-tag) {
  --el-tag-bg-color: #111827;
  --el-tag-border-color: #334155;
  --el-tag-text-color: #e5e7eb;
}

.library-preview-first {
  margin-top: 0;
  padding-top: 0;
  border-top: none;
}

:global(.workflow-prompt-dialog-modal) {
  background: rgba(0, 0, 0, 0.82) !important;
  backdrop-filter: blur(4px);
}

:global(.workflow-prompt-dialog-modal) .el-overlay-dialog {
  background: transparent !important;
}

:global(.output-preview-dialog-modal) {
  background: rgba(0, 0, 0, 0.82) !important;
  backdrop-filter: blur(4px);
}

:global(.output-preview-dialog-modal) .el-overlay-dialog {
  background: transparent !important;
}

:global(.workflow-prompt-dialog) .el-dialog {
  --el-bg-color-overlay: #000000;
  --el-dialog-bg-color: #000000;
  --el-fill-color-light: #0c0c0c;
  --el-fill-color-blank: #000000;
  --el-border-color: #000000;
  --el-border-color-light: #000000;
  --el-text-color-primary: #dde5ee;
  --el-text-color-regular: #c7d0da;
  --el-text-color-secondary: #8c98a7;
  border: 1px solid #000000;
  border-radius: 20px;
  background: #000000 !important;
  box-shadow: 0 24px 56px rgba(3, 8, 15, 0.34);
  overflow: hidden;
  padding: 0;
}

:global(.workflow-prompt-dialog) .el-dialog__header.show-close,
:global(.workflow-prompt-dialog) .el-dialog__header {
  margin: 0;
  padding: 20px 22px 0;
  background: #000000 !important;
}

:global(.workflow-prompt-dialog) .el-dialog__body {
  padding: 20px 22px 22px;
  background: #000000 !important;
}

:global(.output-preview-dialog) .el-dialog {
  --el-bg-color-overlay: #000000;
  --el-dialog-bg-color: #000000;
  --el-fill-color-light: #0c0c0c;
  --el-fill-color-blank: #000000;
  --el-border-color: #000000;
  --el-border-color-light: #000000;
  --el-text-color-primary: #dde5ee;
  --el-text-color-regular: #c7d0da;
  --el-text-color-secondary: #8c98a7;
  border: 1px solid #000000;
  border-radius: 20px;
  background: #000000 !important;
  box-shadow: 0 24px 56px rgba(3, 8, 15, 0.34);
  overflow: hidden;
  padding: 0;
}

:global(.output-preview-dialog) .el-dialog__header.show-close,
:global(.output-preview-dialog) .el-dialog__header {
  margin: 0;
  padding: 20px 22px 0;
  background: #000000 !important;
}

:global(.output-preview-dialog) .el-dialog__body {
  padding: 20px 22px 22px;
  background: #000000 !important;
}

:global(.workflow-prompt-dialog) .el-dialog__footer {
  margin: 0;
  padding: 0 22px 22px;
  background: #000000 !important;
}

:global(.workflow-prompt-dialog) .el-overlay,
:global(.workflow-prompt-dialog) .el-overlay-dialog {
  background: transparent !important;
}

:global(.workflow-prompt-dialog) .el-dialog__headerbtn {
  background: transparent;
}

:global(.workflow-prompt-dialog) .el-dialog__headerbtn .el-dialog__close {
  color: #8c98a7;
}

:global(.workflow-prompt-dialog) .el-dialog__headerbtn:hover {
  background: rgba(55, 69, 84, 0.52);
  border-radius: 10px;
}

:global(.workflow-prompt-dialog) .el-dialog__headerbtn:hover .el-dialog__close {
  color: #e3e9f0;
}

.library-preview-dialog :deep(.el-dialog) {
  --el-bg-color-overlay: #0b1220;
  --el-dialog-bg-color: #0b1220;
  --el-fill-color-light: #111827;
  --el-fill-color-blank: #0b1220;
  --el-border-color: #1e293b;
  --el-border-color-light: #1e293b;
  --el-text-color-primary: #f8fafc;
  --el-text-color-regular: #cbd5e1;
  --el-text-color-secondary: #94a3b8;
  border: 1px solid #1e293b;
  border-radius: 20px;
  background: #0b1220 !important;
  box-shadow: 0 24px 64px rgba(2, 6, 23, 0.62);
  overflow: hidden;
}

.library-preview-dialog :deep(.el-overlay-dialog) {
  background: transparent;
}

.library-preview-dialog :deep(.el-dialog__header) {
  margin: 0;
  padding: 20px 22px 0;
  background: #0b1220 !important;
}

.library-preview-dialog :deep(.el-dialog__body) {
  padding: 20px 22px 22px;
  background: #0b1220 !important;
}

.library-preview-dialog :deep(.el-dialog__footer) {
  background: #0b1220 !important;
}

.library-preview-dialog :deep(.el-dialog__headerbtn .el-dialog__close) {
  color: #94a3b8;
}

.library-preview-dialog :deep(.el-dialog__headerbtn) {
  background: transparent;
}

.library-preview-dialog :deep(.el-dialog__headerbtn:hover) {
  background: rgba(30, 41, 59, 0.72);
  border-radius: 10px;
}

.library-preview-dialog :deep(.el-dialog__headerbtn:hover .el-dialog__close) {
  color: #e2e8f0;
}

@media (max-width: 1024px) {
  .editor-layout {
    grid-template-columns: 1fr;
  }

  .sidebar {
    border-right: none;
    border-bottom: 1px solid #1f2937;
  }

  .sidebar-list {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 12px;
  }

  .library-group {
    display: contents;
  }

  .library-group-title {
    grid-column: 1 / -1;
  }

  .workspace-body {
    grid-template-columns: 1fr;
  }

  .inspector {
    border-left: none;
    border-top: 1px solid #1f2937;
  }
}

@media (max-width: 768px) {
  .editor-toolbar {
    padding: 12px;
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-side {
    justify-content: space-between;
  }

  .toolbar-title p {
    gap: 6px;
  }

  .toolbar-actions {
    flex-wrap: wrap;
  }

  .workspace-tools {
    width: 100%;
  }
}

:global(.workflow-logs-dialog-modal) {
  background: rgba(0, 0, 0, 0.82) !important;
  backdrop-filter: blur(4px);
}

:global(.workflow-logs-dialog-modal) .el-overlay-dialog {
  background: transparent !important;
}

:global(.workflow-logs-dialog) .el-dialog {
  --el-bg-color-overlay: #000000;
  --el-dialog-bg-color: #000000;
  --el-fill-color-light: #0c0c0c;
  --el-fill-color-blank: #000000;
  --el-border-color: #000000;
  --el-border-color-light: #000000;
  --el-text-color-primary: #dde5ee;
  --el-text-color-regular: #c7d0da;
  --el-text-color-secondary: #8c98a7;
  border: 1px solid #000000;
  border-radius: 20px;
  background: #000000 !important;
  box-shadow: 0 24px 56px rgba(3, 8, 15, 0.34);
  overflow: hidden;
  padding: 0;
}

:global(.workflow-logs-dialog) .el-dialog__header.show-close,
:global(.workflow-logs-dialog) .el-dialog__header {
  margin: 0;
  padding: 20px 22px 0;
  background: #000000 !important;
}

:global(.workflow-logs-dialog) .el-dialog__body {
  padding: 20px 22px 22px;
  background: #000000 !important;
}

.workflow-logs-dialog-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.workflow-logs-dialog-head h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #e5e7eb;
}

.clear-logs-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid #374151;
  border-radius: 8px;
  background: #111827;
  color: #9ca3af;
  font-size: 13px;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    background-color 0.15s ease,
    color 0.15s ease;
}

.clear-logs-btn:hover {
  border-color: #ef4444;
  background: #7f1d1d;
  color: #fca5a5;
}

.workflow-logs-dialog-body {
  max-height: 480px;
  overflow: hidden;
}

.workflow-logs-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  gap: 12px;
  color: #6b7280;
}

.workflow-logs-empty .el-icon {
  font-size: 48px;
}

.workflow-logs-empty p {
  margin: 0;
  font-size: 13px;
}

.workflow-logs-list {
  display: grid;
  gap: 8px;
  max-height: 460px;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: #374151 transparent;
}

.workflow-logs-list::-webkit-scrollbar {
  width: 6px;
}

.workflow-logs-list::-webkit-scrollbar-track {
  background: transparent;
}

.workflow-logs-list::-webkit-scrollbar-thumb {
  background: #374151;
  border-radius: 3px;
}

.workflow-log-item {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px 16px;
  border-radius: 10px;
  background: #111827;
  border: 1px solid #1f2937;
  cursor: pointer;
  transition: all 0.2s ease;
}

.workflow-log-item:hover {
  background: #1a2332;
  border-color: #374151;
}

.workflow-log-item.info {
  border-left: 3px solid #3b82f6;
}

.workflow-log-item.success {
  border-left: 3px solid #22c55e;
}

.workflow-log-item.error {
  border-left: 3px solid #ef4444;
}

.workflow-log-item.warning {
  border-left: 3px solid #f59e0b;
}

.log-item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.log-node-title {
  font-size: 14px;
  font-weight: 600;
  color: #e5e7eb;
}

.log-time {
  font-size: 12px;
  color: #6b7280;
}

.log-item-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.log-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.log-tokens {
  font-size: 12px;
  color: #9ca3af;
}

.log-preview {
  font-size: 13px;
  color: #9ca3af;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.log-item-actions {
  display: flex;
  justify-content: flex-end;
}

.logs-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logs-count {
  font-size: 13px;
  color: #9ca3af;
}

.workflow-logs-empty .hint {
  font-size: 12px;
  color: #6b7280;
  margin-top: 4px;
}

:global(.log-detail-dialog-modal) {
  background: rgba(0, 0, 0, 0.82) !important;
  backdrop-filter: blur(4px);
}

:global(.log-detail-dialog-modal) .el-overlay-dialog {
  background: transparent !important;
}

:global(.log-detail-dialog) .el-dialog {
  --el-bg-color-overlay: #000000;
  --el-dialog-bg-color: #0a0a0a;
  --el-fill-color-light: #0c0c0c;
  --el-fill-color-blank: #0a0a0a;
  --el-border-color: #1f2937;
  --el-border-color-light: #1f2937;
  --el-text-color-primary: #dde5ee;
  --el-text-color-regular: #c7d0da;
  --el-text-color-secondary: #8c98a7;
  border: 1px solid #1f2937;
  border-radius: 16px;
  background: #0a0a0a !important;
  box-shadow: 0 24px 56px rgba(0, 0, 0, 0.5);
  overflow: hidden;
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  margin: 0;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

:global(.log-detail-dialog) .el-dialog__header {
  margin: 0;
  padding: 16px 20px;
  background: #0a0a0a !important;
  border-bottom: 1px solid #1f2937;
  flex-shrink: 0;
}

:global(.log-detail-dialog) .el-dialog__title {
  color: #e5e7eb;
  font-weight: 600;
}

:global(.log-detail-dialog) .el-dialog__headerbtn .el-dialog__close {
  color: #6b7280;
}

:global(.log-detail-dialog) .el-dialog__headerbtn:hover .el-dialog__close {
  color: #e5e7eb;
}

:global(.log-detail-dialog) .el-dialog__body {
  background: #0a0a0a !important;
  flex: 1;
  overflow: hidden;
  padding: 0;
}

:global(.log-detail-dialog) .el-dialog__footer {
  background: #0a0a0a !important;
  border-top: 1px solid #1f2937;
  padding: 12px 20px;
  flex-shrink: 0;
}

.log-detail-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 20px;
  max-height: 60vh;
  overflow-y: auto;
}

.log-detail-content::-webkit-scrollbar {
  width: 6px;
}

.log-detail-content::-webkit-scrollbar-track {
  background: #111827;
  border-radius: 3px;
}

.log-detail-content::-webkit-scrollbar-thumb {
  background: #374151;
  border-radius: 3px;
}

.log-detail-content::-webkit-scrollbar-thumb:hover {
  background: #4b5563;
}

.log-detail-meta {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  padding: 16px;
  background: #111827;
  border-radius: 8px;
  border: 1px solid #1f2937;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.meta-label {
  font-size: 12px;
  color: #6b7280;
}

.meta-value {
  font-size: 14px;
  color: #e5e7eb;
}

.log-detail-section h4 {
  margin: 0 0 10px 0;
  font-size: 14px;
  font-weight: 600;
  color: #e5e7eb;
}

.log-detail-block {
  padding: 16px;
  background: #111827;
  border-radius: 8px;
  border: 1px solid #1f2937;
  overflow: auto;
  max-height: 300px;
}

.log-detail-block pre {
  margin: 0;
  font-size: 13px;
  color: #d1d5db;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.log-time {
  font-size: 11px;
  color: #6b7280;
}

.log-content {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 13px;
  color: #d1d5db;
}

.log-node {
  color: #60a5fa;
  font-weight: 500;
}

.log-message {
  min-width: 0;
}

.conversation-messages {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 16px;
}

.message-bubble {
  border-radius: 12px;
  padding: 14px 16px;
  max-width: 100%;
}

.message-bubble.user {
  background: #1e3a5f;
  border: 1px solid #2563eb;
  margin-left: 0;
  margin-right: 40px;
}

.message-bubble.assistant {
  background: #1a2332;
  border: 1px solid #374151;
  margin-left: 40px;
  margin-right: 0;
}

.message-bubble.system {
  background: #2d1f3d;
  border: 1px solid #7c3aed;
  margin-left: 20px;
  margin-right: 20px;
}

.message-sender {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.sender-icon {
  font-size: 16px;
}

.message-bubble.user .sender-icon {
  color: #60a5fa;
}

.message-bubble.assistant .sender-icon {
  color: #22c55e;
}

.message-bubble.system .sender-icon {
  color: #a78bfa;
}

.sender-name {
  font-size: 13px;
  font-weight: 600;
  color: #e5e7eb;
}

.message-content {
  font-size: 14px;
  line-height: 1.6;
  color: #d1d5db;
}

.message-content pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
}

.message-content :deep(.markdown-body) {
  color: #d1d5db;
  font-size: 14px;
  line-height: 1.6;
}

.message-content :deep(.markdown-body p) {
  margin: 8px 0;
}

.message-content :deep(.markdown-body code) {
  background: #111827;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
}

.message-content :deep(.markdown-body pre) {
  background: #111827;
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
}
</style>

<style>
.field-config-popper {
  background: #1e293b !important;
  border: 1px solid #334155 !important;
  border-radius: 8px !important;
  padding: 16px !important;
}

.field-config-popper .el-popper__arrow::before {
  background: #1e293b !important;
  border-color: #334155 !important;
}

.field-config-popper .field-config-row span {
  color: #94a3b8 !important;
}

.field-config-popper .el-input__wrapper {
  background: #0f172a !important;
  border-color: #1f2937 !important;
}

.field-config-popper .el-input__inner {
  color: #e5e7eb !important;
}
</style>
