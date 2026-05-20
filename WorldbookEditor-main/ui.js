// ui.js

import { getContext } from '../../../extensions.js';
import { CONFIG, STATE, THEME_KEY, WI_POSITION_MAP, WI_POSITION_MAP_REV } from './state.js';
import { API } from './api.js';
import { Actions } from './actions.js';
import { logger } from './logger.js';

export const UI = {
    // 防抖辅助函数
    debounce: function(fn, delay) {
        let timer;
        return (...args) => {
            clearTimeout(timer);
            timer = setTimeout(() => fn.apply(this, args), delay);
        };
    },

    updateGlider(tabName) {
        const glider = document.querySelector('.wb-tab-glider');
        const targetTab = document.querySelector(`.wb-tab[data-tab="${tabName}"]`);
        if (glider && targetTab) {
            const width = targetTab.offsetWidth;
            const left = targetTab.offsetLeft;
            glider.style.width = `${width}px`;
            glider.style.transform = `translateX(${left}px)`;
        }
    },

    updateManageGlider(tabName) {
        const glider = document.querySelector('.wb-m-tab-glider');
        const targetTab = document.querySelector(`.wb-m-tab[data-tab="${tabName}"]`);
        if (glider && targetTab) {
            const width = targetTab.offsetWidth;
            const left = targetTab.offsetLeft;
            glider.style.width = `${width}px`;
            glider.style.transform = `translateX(${left}px)`;
        }
    },
    
    centerDialog(el) {
        if (!el) return;
        const winW = window.innerWidth;
        const winH = window.innerHeight;
        el.style.maxHeight = (winH - 40) + 'px';
        el.style.overflow = 'hidden';
        const elW = el.offsetWidth;
        const elH = el.offsetHeight;
        el.style.left = Math.max(0, (winW - elW) / 2) + 'px';
        el.style.top = Math.max(0, (winH - elH) / 2) + 'px';
        el.style.position = 'fixed';
        el.style.margin = '0';
        el.style.transform = 'none';
    },

    setupModalPositioning(el, overlay) {
        requestAnimationFrame(() => this.centerDialog(el));
        const resizeHandler = () => this.centerDialog(el);
        window.addEventListener('resize', resizeHandler);
        const originalRemove = overlay.remove.bind(overlay);
        overlay.remove = () => {
            window.removeEventListener('resize', resizeHandler);
            originalRemove();
        };
    },

    async open() {
        if (document.getElementById(CONFIG.id)) return;

        const panel = document.createElement('div');
        panel.id = CONFIG.id;
        panel.innerHTML = `
            <div class="wb-header-bar">
                <div class="wb-tabs">
                    <div class="wb-tab-glider"></div>
                    <div class="wb-tab active" data-tab="editor"><i class="fa-solid fa-pen-to-square"></i> 编辑世界书</div>
                    <div class="wb-tab" data-tab="binding"><i class="fa-solid fa-link"></i> 绑定世界书</div>
                    <div class="wb-tab" data-tab="manage"><i class="fa-solid fa-list-check"></i> 管理世界书</div>
                </div>
                <div id="wb-close" class="wb-header-close" title="关闭"><i class="fa-solid fa-xmark"></i></div>
            </div>

            <div class="wb-content">
                <div id="wb-loading-layer" style="position:absolute;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.5);z-index:100;display:flex;align-items:center;justify-content:center;backdrop-filter:blur(2px);">
                    <div style="font-size:2em;color:#fff"><i class="fa-solid fa-circle-notch fa-spin"></i></div>
                </div>
                
                <!-- 视图 1: 编辑器 -->
                <div id="wb-view-editor" class="wb-view-section">
                    <div class="wb-book-bar">
                        <select id="wb-book-selector" style="flex:1;"><option>加载中...</option></select>
                        <div class="wb-menu-wrapper">
                            <button class="wb-btn-circle" title="分析与统计" id="btn-wb-analysis"><i class="fa-solid fa-coins"></i></button>
                            <div class="wb-menu-dropdown" id="wb-analysis-menu">
                                <div class="wb-menu-item" data-type="stats"><i class="fa-solid fa-chart-pie"></i> 世界书统计与分析</div>
                                <div class="wb-menu-item" data-type="context"><i class="fa-solid fa-align-left"></i> 世界书实际上下文</div>
                                <div class="wb-menu-item" data-type="export_txt"><i class="fa-solid fa-file-lines"></i> 导出世界书为TXT</div>
                            </div>
                        </div>
                        <div class="wb-menu-wrapper">
                            <button class="wb-btn-circle" title="更多操作" id="btn-wb-menu-trigger"><i class="fa-solid fa-magic-wand-sparkles interactable"></i></button>
                            <div class="wb-menu-dropdown" id="wb-main-menu">
                                <div class="wb-menu-item" data-action="import"><i class="fa-solid fa-file-import"></i> 导入世界书</div>
                                <div class="wb-menu-item" data-action="export"><i class="fa-solid fa-file-export"></i> 导出世界书</div>
                                <div class="wb-menu-item" data-action="create"><i class="fa-solid fa-plus"></i> 新建世界书</div>
                                <div class="wb-menu-item" data-action="rename"><i class="fa-solid fa-pen"></i> 重命名世界书</div>
                                <div class="wb-menu-item danger" data-action="delete"><i class="fa-solid fa-trash"></i> 删除世界书</div>
                            </div>
                        </div>
                        <div class="wb-menu-wrapper">
                            <button class="wb-btn-circle" title="切换主题" id="btn-wb-menu-theme"><i id="btn-wb-menu-theme-toggle" class="fa-solid fa-moon interactable"></i></button>
                        </div>
                        <input type="file" id="wb-import-file" accept=".json,.wb" style="display:none">
                    </div>
                    <div class="wb-stat-line">
                        <div class="wb-stat-group">
                            <div id="wb-warning-stat" class="wb-warning-badge hidden" title="点击查看问题条目"><i class="fa-solid fa-circle-exclamation"></i> <span id="wb-warning-count">0</span></div>
                            <div id="wb-exclude-recursion-stat" class="wb-exclude-recursion-badge hidden" title="点击查看未设不可递归的条目"><i class="fa-solid fa-triangle-exclamation"></i> <span id="wb-exclude-recursion-count">0</span></div>
                            <div class="wb-stat-item" id="wb-display-count">0 条目</div>
                        </div>
                    </div>
                    <div class="wb-tool-bar">
                        <input class="wb-input-dark" id="wb-search-entry" style="flex:1; width:100%; border-radius:15px; padding-left:15px;" placeholder="搜索条目...">
                        <button class="wb-btn-circle interactable" id="btn-group-sort" title="分组排序管理"><i class="fa-solid fa-arrow-down-9-1"></i></button>
                        <button class="wb-btn-circle" id="btn-sort-priority" title="列表按优先级重排"><i class="fa-solid fa-filter"></i></button>
                        <button class="wb-btn-circle" id="btn-add-entry" title="新建条目"><i class="fa-solid fa-plus"></i></button>
                    </div>
                    <div class="wb-list" id="wb-entry-list"></div>
                </div>

                <!-- 视图 2: 绑定管理 -->
                <div id="wb-view-binding" class="wb-view-section wb-hidden">
                    <div class="wb-bind-grid">
                        <div class="wb-bind-card"><div class="wb-bind-title"><span><i class="fa-solid fa-globe"></i> 全局世界书</span></div><div class="wb-scroll-list" id="wb-bind-global-list"></div></div>
                        <div class="wb-bind-card"><div class="wb-bind-title"><span><i class="fa-solid fa-user-tag"></i> 角色世界书</span></div><div class="wb-bind-label"> 主要世界书</div><div id="wb-bind-char-primary" style="position:relative"></div><div class="wb-bind-label">附加世界书</div><div class="wb-scroll-list" id="wb-bind-char-list"></div></div>
                        <div class="wb-bind-card"><div class="wb-bind-title"><span><i class="fa-solid fa-comments"></i> 聊天世界书</span></div><div id="wb-bind-chat" style="position:relative"></div></div>
                    </div>
                    <div id="wb-footer-info" class="wb-footer-info"></div>
                </div>

                <!-- 视图 3: 管理 -->
                <div id="wb-view-manage" class="wb-view-section wb-hidden">
                    <div class="wb-manage-container">
                        <div class="wb-tool-bar" style="flex-wrap: nowrap; gap: 10px; align-items: center; justify-content: flex-start;">
                            <input class="wb-input-dark" id="wb-manage-search" style="flex:1; min-width:100px; border-radius:15px;padding-left:15px" placeholder="🔍 搜索世界书...">
                            <div class="wb-manage-tabs" id="wb-manage-mode-tabs">
                                <div class="wb-m-tab-glider"></div>
                                <div class="wb-m-tab ${STATE.manageTab === 'worldbook' ? 'active' : ''}" data-tab="worldbook">世界书</div>
                                <div class="wb-m-tab ${STATE.manageTab === 'custom' ? 'active' : ''}" data-tab="custom">自定义</div>
                            </div>
                        </div>
                        <div class="wb-manage-content" id="wb-manage-content"></div>
                    </div>
                </div>
            </div>
        `;
        document.body.appendChild(panel);

        const $ = (sel) => panel.querySelector(sel);
        const $$ = (sel) => panel.querySelectorAll(sel);

        // ESC 快捷键：优先关闭子弹窗，如果没有子弹窗则关闭主面板
        const escHandler = (e) => {
            if (e.key === 'Escape') {
                // 检查是否有子弹窗（模态框、弹窗、下拉菜单等）
                const subOverlays = document.querySelectorAll('.wb-modal-overlay:not(.wb-main-panel-overlay), .wb-sort-modal-overlay, #wb-content-popup-overlay, #wb-active-dropdown, #wb-gr-tag-menu-overlay');
                const openDropdowns = document.querySelectorAll('.wb-menu-dropdown.show, .wb-gr-dropdown.show');

                if (subOverlays.length > 0 || openDropdowns.length > 0) {
                    // 有关闭子弹窗或下拉菜单，只关闭这些
                    subOverlays.forEach(el => el.remove());
                    openDropdowns.forEach(el => el.classList.remove('show'));
                } else {
                    // 没有子弹窗，关闭主面板
                    const mainPanel = document.getElementById(CONFIG.id);
                    if (mainPanel) {
                        Actions.flushPendingSave().then(() => mainPanel.remove());
                    }
                }
            }
        };
        document.addEventListener('keydown', escHandler);

        $('#wb-close').onclick = async () => {
            document.removeEventListener('keydown', escHandler);
            await Actions.flushPendingSave();
            panel.remove();
        };
        $$('.wb-tab').forEach(el => el.onclick = () => Actions.switchView(el.dataset.tab));
        $('#wb-book-selector').addEventListener('change', (e) => Actions.loadBook(e.target.value));
        $('#wb-search-entry').oninput = this.debounce((e) => this.renderList(e.target.value), 250);
        $('#btn-add-entry').onclick = () => Actions.addNewEntry();
        $('#btn-group-sort').onclick = () => this.openSortingModal();
        $('#btn-sort-priority').onclick = () => Actions.sortByPriority();
        $('#btn-wb-menu-theme').onclick = () => Actions.switchTheme();

        const savedTheme = localStorage.getItem(THEME_KEY) || 'light';
        Actions.applyTheme(savedTheme);

        const analysisBtn = $('#btn-wb-analysis');
        const analysisMenu = $('#wb-analysis-menu');
        analysisBtn.onclick = (e) => {
            e.stopPropagation();
            const isShow = analysisMenu.classList.contains('show');
            document.querySelectorAll('.wb-menu-dropdown.show').forEach(el => el.classList.remove('show'));
            if (!isShow) analysisMenu.classList.add('show');
        };
        analysisMenu.querySelectorAll('.wb-menu-item').forEach(item => {
            item.onclick = (e) => {
                e.stopPropagation();
                analysisMenu.classList.remove('show');
                const type = item.dataset.type;
                if (type === 'stats') this.openAnalysisModal();
                else if (type === 'context') this.openContextPreviewModal();
                else if (type === 'export_txt') Actions.actionExportTxt();
            };
        });

        const menuTrigger = $('#btn-wb-menu-trigger');
        const menuDropdown = $('#wb-main-menu');
        menuTrigger.onclick = (e) => {
            e.stopPropagation();
            const isShow = menuDropdown.classList.contains('show');
            document.querySelectorAll('.wb-menu-dropdown, .wb-gr-dropdown').forEach(el => el.classList.remove('show'));
            if (!isShow) menuDropdown.classList.add('show');
        };
        menuDropdown.querySelectorAll('.wb-menu-item').forEach(item => {
            item.onclick = async (e) => {
                e.stopPropagation();
                menuDropdown.classList.remove('show');
                const action = item.dataset.action;
                if (action === 'import') Actions.actionImport();
                else if (action === 'export') Actions.actionExport();
                else if (action === 'create') Actions.actionCreateNew();
                else if (action === 'rename') Actions.actionRename();
                else if (action === 'delete') Actions.actionDelete();
            };
        });

        document.addEventListener('click', (e) => {
            if (menuDropdown.classList.contains('show') && !menuTrigger.contains(e.target) && !menuDropdown.contains(e.target)) menuDropdown.classList.remove('show');
            if (analysisMenu.classList.contains('show') && !analysisBtn.contains(e.target) && !analysisMenu.contains(e.target)) analysisMenu.classList.remove('show');
        });

        const fileInput = $('#wb-import-file');
        fileInput.onchange = (e) => {
            if (e.target.files.length > 0) {
                Actions.actionHandleImport(e.target.files[0]);
                fileInput.value = '';
            }
        };

        $('#wb-entry-list').addEventListener('wb-reorder', (e) => Actions.reorderEntry(e.detail.from, e.detail.to));
        $('#wb-manage-search').oninput = this.debounce((e) => this.renderManageView(e.target.value), 250);

        // 动态添加自定义分组管理按钮
        (() => {
            const searchInput = $('#wb-manage-search');
            const tabs = $('#wb-manage-mode-tabs');
            if (searchInput && tabs) {
                const btn = document.createElement('button');
                btn.id = 'btn-custom-group-add';
                btn.className = 'wb-btn-rect';
                btn.style.display = 'none';
                btn.style.padding = '6px 15px';
                btn.style.fontSize = '0.9em';
                btn.innerHTML = '<i class="fa-solid fa-tags"></i> 管理标签';
                searchInput.insertAdjacentElement('afterend', btn);
            }
        })();

        $('#btn-custom-group-add').onclick = () => this.openCustomGroupModal();

        $$('#wb-manage-mode-tabs .wb-m-tab').forEach(tab => {
            tab.onclick = () => {
                if (STATE.manageTab === tab.dataset.tab) return;
                $$('#wb-manage-mode-tabs .wb-m-tab').forEach(t => t.classList.remove('active'));
                tab.classList.add('active');
                STATE.manageTab = tab.dataset.tab;
                const manageBtn = document.getElementById('btn-custom-group-add');
                if (manageBtn) {
                    manageBtn.style.display = tab.dataset.tab === 'custom' ? 'inline-block' : 'none';
                }
                this.updateManageGlider(tab.dataset.tab);
                this.renderManageView($('#wb-manage-search').value);
            };
        });

        const activeTab = $('#wb-manage-mode-tabs .wb-m-tab.active');
        if (activeTab) {
            const manageBtn = document.getElementById('btn-custom-group-add');
            if (manageBtn) {
                manageBtn.style.display = activeTab.dataset.tab === 'custom' ? 'inline-block' : 'none';
            }
        }

        const loader = document.getElementById('wb-loading-layer');

        try {
            await Actions.refreshAllContext();
            STATE.isManageDirty = true;

            const charPrimary = STATE.bindings.char.primary;
            const chatBook = STATE.bindings.chat;
            let targetBook = null;

            if (charPrimary && STATE.allBookNames.includes(charPrimary)) targetBook = charPrimary;
            else if (chatBook && STATE.allBookNames.includes(chatBook)) targetBook = chatBook;
            else if (STATE.allBookNames.length > 0) targetBook = STATE.allBookNames[0];

            this.renderBookSelector();
            this.updateHeaderInfo();

            if (targetBook) await Actions.loadBook(targetBook);
            else this.renderList();

        } catch (e) {
            logger.error("Panel Init Error:", e);
            toastr.error("初始化面板数据失败");
        } finally {
            if (loader) loader.style.display = 'none';
        }

        this.updateGlider('editor');
        this.updateManageGlider(STATE.manageTab);
        setTimeout(() => {
            const glider = panel.querySelector('.wb-tab-glider');
            if (glider) glider.classList.add('wb-glider-animating');
        }, 50);

        Actions.switchView('editor');
        this.initEntryListDelegation();

        // 首次打开显示ESC键提示弹窗（延迟显示，确保主面板完全加载）
        // 移动设备检测：如果是移动设备则跳过首次打开弹窗
        const isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent) || window.innerWidth <= 768;
        const firstOpenKey = '[SiilyTavern]世界书管理器-首次打开';
        if (!localStorage.getItem(firstOpenKey)) {
            if (isMobile) {
                // 移动设备直接标记为已打开，不显示弹窗
                localStorage.setItem(firstOpenKey, 'true');
            } else {
                setTimeout(() => {
                    this.showFirstOpenHintModal();
                    localStorage.setItem(firstOpenKey, 'true');
                }, 50);
            }
        }
    },

    showFirstOpenHintModal() {
        const overlay = document.createElement('div');
        overlay.className = 'wb-sort-modal-overlay';

        overlay.innerHTML = `
            <div class="wb-sort-modal" style="background:#fff; border-radius:12px; padding:30px; min-width:320px; max-width:400px; box-shadow:0 4px 20px rgba(0,0,0,0.15);">
                <div style="display:flex; flex-direction:column; align-items:center; justify-content:center; min-height:120px; margin-bottom:30px;">
                    <div style="font-size:16px; font-weight:700; color:#374151; text-align:center; line-height:1.6;">
                        电脑用户按 <span style="color:#ef4444;">ESC</span> 键可以直接关闭插件任意弹窗
                    </div>
                </div>
                <div style="display:flex; justify-content:center;">
                    <button class="wb-btn-first-open-hint" style="padding:12px 40px; background:#000000; color:#fff; border:none; border-radius:8px; font-size:14px; font-weight:500; cursor:pointer; transition:background 0.2s;">
                        已阅
                    </button>
                </div>
            </div>
        `;

        document.body.appendChild(overlay);
        this.setupModalPositioning(overlay.querySelector('.wb-sort-modal'), overlay);

        const close = () => overlay.remove();
        overlay.querySelector('.wb-btn-first-open-hint').onclick = close;
        overlay.onclick = (e) => { if (e.target === overlay) close(); };
    },

    showTagFilterHintModal() {
        const overlay = document.createElement('div');
        overlay.className = 'wb-sort-modal-overlay';

        overlay.innerHTML = `
            <div class="wb-sort-modal" style="background:#fff; border-radius:12px; padding:30px; min-width:320px; max-width:450px; box-shadow:0 4px 20px rgba(0,0,0,0.15);">
                <div style="display:flex; flex-direction:column; align-items:center; justify-content:center; min-height:100px; margin-bottom:25px;">
                    <div style="font-size:14px; font-weight:600; color:#374151; text-align:center; line-height:1.8;">
                        若您当前未设置标签分组，可以去【管理世界书】→【自定义】→【管理标签】中为世界书打上标签进行分组，届时您可以直接按标签分组快速筛选世界书
                    </div>
                </div>
                <div style="display:flex; justify-content:center;">
                    <button class="wb-btn-tag-filter-hint" style="padding:10px 35px; background:#000000; color:#fff; border:none; border-radius:8px; font-size:14px; font-weight:500; cursor:pointer; transition:background 0.2s;">
                        已阅
                    </button>
                </div>
            </div>
        `;

        document.body.appendChild(overlay);
        this.setupModalPositioning(overlay.querySelector('.wb-sort-modal'), overlay);

        const close = () => overlay.remove();
        overlay.querySelector('.wb-btn-tag-filter-hint').onclick = close;
        overlay.onclick = (e) => { if (e.target === overlay) close(); };
    },

    renderBookSelector() {
        const selector = document.getElementById('wb-book-selector');
        if (!selector) return;
        const { char, global, chat } = STATE.bindings;
        const allNames = STATE.allBookNames;
        const charBooks = new Set([char.primary, ...char.additional].filter(Boolean));
        const globalBooks = new Set(global);
        const chatBook = chat;

        let html = '';
        if (char.primary) html += `<optgroup label="主要世界书"><option value="${char.primary}">${char.primary}</option></optgroup>`;
        
        const additionalBooks = char.additional.filter(name => name && name !== char.primary);
        if (additionalBooks.length > 0) {
            html += `<optgroup label="附加世界书">`;
            additionalBooks.forEach(name => html += `<option value="${name}">${name}</option>`);
            html += `</optgroup>`;
        }
        if (globalBooks.size > 0) {
            html += `<optgroup label="全局启用">`;
            globalBooks.forEach(name => html += `<option value="${name}">${name}</option>`);
            html += `</optgroup>`;
        }
        if (chatBook) html += `<optgroup label="当前聊天"><option value="${chatBook}">${chatBook}</option></optgroup>`;
        
        html += `<optgroup label="其他">`;
        allNames.forEach(name => html += `<option value="${name}">${name}</option>`);
        html += `</optgroup>`;

        selector.innerHTML = html;
        if (STATE.currentBookName) selector.value = STATE.currentBookName;
        this.applyCustomDropdown('wb-book-selector');
    },

    renderBindingView() {
        const allNames = STATE.allBookNames;
        const { char, global, chat } = STATE.bindings;
        const view = document.getElementById('wb-view-binding');
        if (!view) return;

        const createOpts = (selectedVal) => {
            let html = '<option value="">(无)</option>';
            allNames.forEach(name => {
                const sel = name === selectedVal ? 'selected' : '';
                html += `<option value="${name}" ${sel}>${name}</option>`;
            });
            return html;
        };

        const createMultiSelect = (containerSelector, initialSelectedArray, dataClass) => {
            const container = view.querySelector(containerSelector);
            if (!container) return;
            container.innerHTML = '';
            container.className = 'wb-multi-select';
            const selectedSet = new Set(initialSelectedArray.filter(n => allNames.includes(n)));
            const dom = document.createElement('div');
            dom.innerHTML = `
                <div class="wb-ms-tags"></div>
                <div class="wb-ms-dropdown">
                    <div class="wb-ms-search"><input type="text" placeholder="搜索世界书..."></div>
                    <div class="wb-ms-list"></div>
                </div>
            `;
            container.appendChild(dom);
            const tagsEl = dom.querySelector('.wb-ms-tags');
            const dropEl = dom.querySelector('.wb-ms-dropdown');
            const inputEl = dom.querySelector('input');
            const listEl = dom.querySelector('.wb-ms-list');
            const searchBox = dom.querySelector('.wb-ms-search');
            searchBox.style.position = 'relative';

            // Tag filter button
            const tagFilter = document.createElement('div');
            tagFilter.className = 'wb-ms-tag-filter';
            tagFilter.title = '按标签筛选';
            tagFilter.innerHTML = '<i class="fa-solid fa-tag"></i>';
            searchBox.appendChild(tagFilter);

            let currentTag = null;

            const updateTagFilterButton = () => {
                if (currentTag) {
                    tagFilter.innerHTML = `<span class="wb-ms-tag-pill">${currentTag}<span class="wb-ms-tag-close">&times;</span></span>`;
                    const closeEl = tagFilter.querySelector('.wb-ms-tag-close');
                    if (closeEl) {
                        closeEl.onclick = (e) => {
                            e.stopPropagation();
                            currentTag = null;
                            updateTagFilterButton();
                            applyFilters();
                        };
                    }
                } else {
                    tagFilter.innerHTML = '<i class="fa-solid fa-tag"></i>';
                }
            };

            const applyFilters = () => {
                const term = inputEl.value.toLowerCase();
                const metadata = STATE.metadata || {};
                listEl.querySelectorAll('.wb-ms-item').forEach(item => {
                    const textMatch = item.textContent.toLowerCase().includes(term);
                    const bookName = item.dataset.value;
                    const tags = metadata[bookName]?.tags || [];
                    const tagMatch = !currentTag || tags.includes(currentTag);
                    item.classList.toggle('hidden', !(textMatch && tagMatch));
                });
            };

            const refresh = () => {
                tagsEl.innerHTML = '';
                if (selectedSet.size === 0) tagsEl.innerHTML = `<div class="wb-ms-placeholder">点击选择世界书...</div>`;
                else {
                    selectedSet.forEach(name => {
                        const tag = document.createElement('div');
                        tag.className = 'wb-ms-tag';
                        tag.dataset.val = name;
                        tag.dataset.bindType = dataClass;
                        tag.innerHTML = `<span>${name}</span><span class="wb-ms-tag-close" style="margin-left:6px;cursor:pointer;font-weight:bold">×</span>`;
                        tag.querySelector('.wb-ms-tag-close').onclick = (e) => {
                            e.stopPropagation();
                            selectedSet.delete(name);
                            refresh();
                            Actions.saveBindings();
                        };
                        tagsEl.appendChild(tag);
                    });
                }
                listEl.innerHTML = '';
                // 过滤掉已选中的选项，只显示未选中的，并按字母顺序排序
                const availableOptions = [...allNames].filter(name => !selectedSet.has(name)).sort((a, b) => a.localeCompare(b));
                if (availableOptions.length === 0) {
                    listEl.innerHTML = `<div style="padding:10px;color:#666;text-align:center">没有可用选项</div>`;
                } else {
                    availableOptions.forEach(name => {
                        const item = document.createElement('div');
                        item.className = 'wb-ms-item';
                        item.textContent = name;
                        item.dataset.value = name;
                        item.onclick = () => {
                            selectedSet.add(name);
                            inputEl.value = '';
                            refresh();
                            Actions.saveBindings();
                            // 多选组件：添加选项后不关闭下拉框
                        };
                        listEl.appendChild(item);
                    });
                    applyFilters();
                }
            };

            const showTagMenu = () => {
                // 首次点击标签筛选按钮时显示提示
                const tagFilterHintKey = '[SillyTavern]世界书管理器-标签筛选提示';
                if (!localStorage.getItem(tagFilterHintKey)) {
                    localStorage.setItem(tagFilterHintKey, 'true');
                    this.showTagFilterHintModal();
                }

                const existingMenu = document.getElementById('wb-ms-tag-menu-overlay');
                if (existingMenu) {
                    existingMenu.remove();
                    return;
                }
                const menu = document.createElement('div');
                menu.id = 'wb-ms-tag-menu-overlay';
                menu.className = 'wb-ms-tag-menu-overlay';
                const meta = STATE.metadata || {};
                const allTags = [...new Set(Object.values(meta).flatMap(m => m.tags || []))].sort();
                const allOption = document.createElement('div');
                allOption.className = 'wb-ms-tag-option';
                allOption.textContent = '全部';
                allOption.onclick = (e) => {
                    e.stopPropagation();
                    currentTag = null;
                    updateTagFilterButton();
                    applyFilters();
                    menu.remove();
                };
                menu.appendChild(allOption);
                allTags.forEach(tag => {
                    const opt = document.createElement('div');
                    opt.className = 'wb-ms-tag-option';
                    opt.textContent = tag;
                    opt.onclick = (e) => {
                        e.stopPropagation();
                        currentTag = tag;
                        updateTagFilterButton();
                        applyFilters();
                        menu.remove();
                    };
                    menu.appendChild(opt);
                });
                document.body.appendChild(menu);
                const rect = tagFilter.getBoundingClientRect();
                menu.style.position = 'fixed';
                menu.style.top = `${rect.bottom + 4}px`;
                menu.style.right = `${window.innerWidth - rect.right}px`;
                menu.style.left = 'auto';
                const closeMenu = (e) => {
                    if (!menu.contains(e.target)) {
                        menu.remove();
                        document.removeEventListener('click', closeMenu);
                    }
                };
                setTimeout(() => document.addEventListener('click', closeMenu), 0);
            };

            tagFilter.onclick = (e) => {
                e.stopPropagation();
                showTagMenu();
            };

            tagsEl.onclick = () => {
                const isVisible = dropEl.classList.contains('show');
                document.querySelectorAll('.wb-ms-dropdown.show').forEach(el => el.classList.remove('show'));
                if (!isVisible) {
                    dropEl.classList.add('show');
                    // 不自动聚焦，避免移动端弹出虚拟键盘
                    // const isTouchDevice = window.matchMedia('(pointer: coarse)').matches;
                    // if (!isTouchDevice) inputEl.focus();
                }
            };
            inputEl.oninput = applyFilters;
            document.addEventListener('click', (e) => { if (!dom.contains(e.target)) dropEl.classList.remove('show'); });
            refresh();
        };

        createMultiSelect('#wb-bind-global-list', global, 'wb-bind-global');
        this.createSingleSelect('wb-bind-char-primary', allNames, char.primary, 'wb-bind-char-primary');
        createMultiSelect('#wb-bind-char-list', char.additional, 'wb-bind-char-add');
        this.createSingleSelect('wb-bind-chat', allNames, chat, 'wb-bind-chat');
    },

    // 带搜索功能的单选组件
    createSingleSelect(containerId, options, selectedValue, dataClass) {
        const view = document.getElementById('wb-view-binding');
        if (!view) return;
        const container = view.querySelector(`#${containerId}`);
        if (!container) return;

        container.innerHTML = '';
        container.className = 'wb-multi-select';

        const dom = document.createElement('div');
        dom.innerHTML = `
            <div class="wb-ms-tags"></div>
            <div class="wb-ms-dropdown">
                <div class="wb-ms-search"><input type="text" placeholder="搜索世界书..."></div>
                <div class="wb-ms-list"></div>
            </div>
        `;
        container.appendChild(dom);

        const tagsEl = dom.querySelector('.wb-ms-tags');
        const dropEl = dom.querySelector('.wb-ms-dropdown');
        const inputEl = dom.querySelector('input');
        const listEl = dom.querySelector('.wb-ms-list');
        const searchBox = dom.querySelector('.wb-ms-search');
        searchBox.style.position = 'relative';

        // 标签筛选按钮
        const tagFilter = document.createElement('div');
        tagFilter.className = 'wb-ms-tag-filter';
        tagFilter.title = '按标签筛选';
        tagFilter.innerHTML = '<i class="fa-solid fa-tag"></i>';
        searchBox.appendChild(tagFilter);

        let currentTag = null;

        const updateTagFilterButton = () => {
            if (currentTag) {
                tagFilter.innerHTML = `<span class="wb-ms-tag-pill">${currentTag}<span class="wb-ms-tag-close">&times;</span></span>`;
                const closeEl = tagFilter.querySelector('.wb-ms-tag-close');
                if (closeEl) {
                    closeEl.onclick = (e) => {
                        e.stopPropagation();
                        currentTag = null;
                        updateTagFilterButton();
                        applyFilters();
                    };
                }
            } else {
                tagFilter.innerHTML = '<i class="fa-solid fa-tag"></i>';
            }
        };

        const showTagMenu = () => {
            // 首次点击标签筛选按钮时显示提示
            const tagFilterHintKey = '[SillyTavern]世界书管理器-标签筛选提示';
            if (!localStorage.getItem(tagFilterHintKey)) {
                localStorage.setItem(tagFilterHintKey, 'true');
                this.showTagFilterHintModal();
            }

            const existingMenu = document.getElementById('wb-ms-tag-menu-overlay');
            if (existingMenu) {
                existingMenu.remove();
                return;
            }
            const menu = document.createElement('div');
            menu.id = 'wb-ms-tag-menu-overlay';
            menu.className = 'wb-ms-tag-menu-overlay';
            const meta = STATE.metadata || {};
            const allTags = [...new Set(Object.values(meta).flatMap(m => m.tags || []))].sort();
            const allOption = document.createElement('div');
            allOption.className = 'wb-ms-tag-option';
            allOption.textContent = '全部';
            allOption.onclick = (e) => {
                e.stopPropagation();
                currentTag = null;
                updateTagFilterButton();
                applyFilters();
                menu.remove();
            };
            menu.appendChild(allOption);
            allTags.forEach(tag => {
                const opt = document.createElement('div');
                opt.className = 'wb-ms-tag-option';
                opt.textContent = tag;
                opt.onclick = (e) => {
                    e.stopPropagation();
                    currentTag = tag;
                    updateTagFilterButton();
                    applyFilters();
                    menu.remove();
                };
                menu.appendChild(opt);
            });
            document.body.appendChild(menu);
            const rect = tagFilter.getBoundingClientRect();
            menu.style.position = 'fixed';
            menu.style.top = `${rect.bottom + 4}px`;
            menu.style.right = `${window.innerWidth - rect.right}px`;
            menu.style.left = 'auto';
            const closeMenu = (e) => {
                if (!menu.contains(e.target)) {
                    menu.remove();
                    document.removeEventListener('click', closeMenu);
                }
            };
            setTimeout(() => document.addEventListener('click', closeMenu), 0);
        };

        tagFilter.onclick = (e) => {
            e.stopPropagation();
            showTagMenu();
        };

        let currentValue = selectedValue || '';
        // 同步容器 value 属性，以便 saveBindings() 可以读取
        container.value = currentValue;

        const updateDisplay = () => {
            tagsEl.innerHTML = '';
            if (currentValue) {
                // 有选中值：添加 has-value 类，显示名称和关闭图标
                tagsEl.classList.add('has-value');
                const nameSpan = document.createElement('span');
                nameSpan.className = 'wb-ms-tag-name';
                nameSpan.textContent = currentValue;
                tagsEl.appendChild(nameSpan);

                const closeIcon = document.createElement('span');
                closeIcon.className = 'wb-ms-close-icon';
                closeIcon.innerHTML = '<i class="fa-solid fa-times"></i>';
                closeIcon.onclick = (e) => {
                    e.stopPropagation();
                    currentValue = '';
                    container.value = currentValue;
                    updateDisplay();
                    Actions.saveBindings();
                };
                tagsEl.appendChild(closeIcon);
            } else {
                // 无选中值：移除 has-value 类，显示占位文本
                tagsEl.classList.remove('has-value');
                tagsEl.innerHTML = `<div class="wb-ms-placeholder">点击选择世界书...</div>`;
            }

            // 过滤掉当前选中的选项，只显示未选中的，并按字母顺序排序
            const availableOptions = [...options].filter(name => name !== currentValue).sort((a, b) => a.localeCompare(b));

            listEl.innerHTML = '';
            availableOptions.forEach(name => {
                const item = document.createElement('div');
                item.className = 'wb-ms-item';
                item.textContent = name;
                item.dataset.value = name;
                item.onclick = () => {
                    // 点击选项：选中并关闭下拉框
                    currentValue = name;
                    container.value = currentValue;
                    inputEl.value = '';
                    updateDisplay();
                    dropEl.classList.remove('show');
                    Actions.saveBindings();
                };
                listEl.appendChild(item);
            });
            applyFilters();
        };

        const applyFilters = () => {
            const term = inputEl.value.toLowerCase();
            const metadata = STATE.metadata || {};
            listEl.querySelectorAll('.wb-ms-item').forEach(item => {
                const textMatch = item.textContent.toLowerCase().includes(term);
                const bookName = item.dataset.value;
                const tags = metadata[bookName]?.tags || [];
                const tagMatch = !currentTag || tags.includes(currentTag);
                item.classList.toggle('hidden', !(textMatch && tagMatch));
            });
        };

        tagsEl.onclick = () => {
            const isVisible = dropEl.classList.contains('show');
            document.querySelectorAll('.wb-ms-dropdown.show').forEach(el => el.classList.remove('show'));
            if (!isVisible) {
                dropEl.classList.add('show');
                // 不自动聚焦，避免移动端弹出虚拟键盘
                // const isTouchDevice = window.matchMedia('(pointer: coarse)').matches;
                // if (!isTouchDevice) inputEl.focus();
            }
        };

        inputEl.oninput = applyFilters;
        document.addEventListener('click', (e) => {
            if (!dom.contains(e.target)) dropEl.classList.remove('show');
        });

        updateDisplay();
    },

    updateHeaderInfo() {
        this.renderGlobalStats();
        const selector = document.getElementById('wb-book-selector');
        if (selector && STATE.currentBookName) selector.value = STATE.currentBookName;

        const footerEl = document.getElementById('wb-footer-info');
        if (footerEl) {
            const context = getContext();
            const charId = context.characterId;
            const charName = (context.characters && context.characters[charId]) ? context.characters[charId].name : '无';
            const avatarImgEl = document.getElementById('avatar_load_preview');
            const avatarHtml = (avatarImgEl && avatarImgEl.src) ? `<img src="${avatarImgEl.src}" class="wb-footer-avatar">` : '';
            const chatName = context.chatId ? String(context.chatId).replace(/\.json$/i, '') : '无';
            footerEl.innerHTML = `<div>当前角色为${avatarHtml}<strong>${charName}</strong></div><div>当前聊天为 <strong>${chatName}</strong></div>`;
        }
    },

    getWarningList() {
        return STATE.entries.filter(entry => entry.disable === false && entry.constant === false && !(entry.key?.length > 0));
    },

    getExcludeRecursionWarningList() {
        return STATE.entries.filter(entry => entry.disable === false && entry.constant === false && entry.excludeRecursion !== true);
    },

    renderGlobalStats() {
        const countEl = document.getElementById('wb-display-count');
        const warningEl = document.getElementById('wb-warning-stat');
        const warningNumEl = document.getElementById('wb-warning-count');
        const erStatEl = document.getElementById('wb-exclude-recursion-stat');
        const erCountEl = document.getElementById('wb-exclude-recursion-count');
        if (countEl) {
            let blueTokens = 0, greenTokens = 0;
            STATE.entries.forEach(entry => {
                if (entry.disable === false) {
                    const t = Actions.getTokenCount(entry.content);
                    if (entry.constant === true) blueTokens += t;
                    else greenTokens += t;
                }
            });
            countEl.innerHTML = `<span style="margin-right:5px">${STATE.entries.length} 条目 | ${blueTokens + greenTokens} Tokens</span><span style="font-size:0.9em; color:#6b7280">( <span class="wb-text-blue" title="蓝灯">${blueTokens}</span> + <span class="wb-text-green" title="绿灯">${greenTokens}</span> )</span>`;
        }
        if (warningEl && warningNumEl) {
            const warnings = this.getWarningList();
            if (warnings.length > 0) {
                warningEl.classList.remove('hidden');
                warningNumEl.textContent = warnings.length;
                warningEl.onclick = () => this.openWarningListModal();
            } else warningEl.classList.add('hidden');
        }
        if (erStatEl && erCountEl) {
            const erWarnings = this.getExcludeRecursionWarningList();
            if (erWarnings.length > 0) {
                erStatEl.classList.remove('hidden');
                erCountEl.textContent = erWarnings.length;
                erStatEl.onclick = () => this.openExcludeRecursionModal();
            } else erStatEl.classList.add('hidden');
        }
    },

    updateCardStatus(uid) {
        const entry = STATE.entries.find(e => e.uid === uid);
        const card = document.querySelector(`.wb-card[data-uid="${uid}"]`);
        if (!entry || !card) return;

        card.classList.remove('disabled', 'type-green', 'type-blue');

        if (entry.disable) {
            card.classList.add('disabled');
        } else {
            if (entry.constant) card.classList.add('type-blue');
            else card.classList.add('type-green');
        }

        const tokenEl = card.querySelector('.wb-token-display');
        if (tokenEl) tokenEl.textContent = Actions.getTokenCount(entry.content);
        const warnContainer = card.querySelector('.wb-warning-container');
        if (warnContainer) {
            const showWarning = entry.disable === false && entry.constant === false && !(entry.key?.length > 0);
            warnContainer.innerHTML = showWarning ? `<i class="fa-solid fa-circle-exclamation wb-keyword-warning" style="color:#ef4444!important; margin-right:6px!important; cursor:pointer!important;" data-wb-tooltip="警告：绿灯条目已启用但未设置关键词，将无法触发"></i>` : '';
            // 为警告图标添加点击事件
            const warningIcon = warnContainer.querySelector('.wb-keyword-warning');
            if (warningIcon) {
                warningIcon.onclick = (e) => {
                    e.stopPropagation();
                    this.openWarningListModal();
                };
            }
        }
        const erWarningContainer = card.querySelector('.wb-exclude-recursion-warning-container');
        if (erWarningContainer) {
            const showERWarning = entry.disable === false && entry.constant === false && entry.excludeRecursion !== true;
            erWarningContainer.innerHTML = showERWarning ? `<i class="fa-solid fa-triangle-exclamation wb-exclude-recursion-warning" style="color:#eab308!important; margin-right:6px!important; cursor:pointer!important;" data-wb-tooltip="警告：此条目可能引发递归冲突，建议启用「不可递归」"></i>` : '';
            // 为递归警告图标添加点击事件
            const erWarningIcon = erWarningContainer.querySelector('.wb-exclude-recursion-warning');
            if (erWarningIcon) {
                erWarningIcon.onclick = (e) => {
                    e.stopPropagation();
                    this.openExcludeRecursionModal();
                };
            }
        }
    },

    renderList(filterText = '') {
        const list = document.getElementById('wb-entry-list');
        if (!list) return;
        const term = filterText.toLowerCase();

        // 首次渲染（list为空）或搜索词为空时，重建完整DOM
        const isFirstRender = list.children.length === 0;
        const isClearSearch = term === '';

        if (isFirstRender || isClearSearch) {
            list.innerHTML = '';
            STATE.entries.forEach((entry, index) => {
                const card = this.createCard(entry, index);
                // 添加可搜索的文本属性（包含所有可搜索内容）
                const comment = entry.comment || '';
                const content = entry.content || '';
                card.dataset.searchText = `${comment} ${content}`.toLowerCase();
                list.appendChild(card);
                this.applyCustomDropdown(`wb-pos-${entry.uid}`);
            });
            return;
        }

        // === 增量过滤：仅切换 hidden 类 ===
        Array.from(list.children).forEach(card => {
            const searchText = card.dataset.searchText || '';
            const hasMatch = searchText.includes(term);
            card.classList.toggle('hidden', !hasMatch);
        });
    },

    createCard(entry, index) {
        const context = getContext();
        const currentAnDepth = (context.chatMetadata?.note_depth) ?? (context.extensionSettings?.note?.defaultDepth) ?? 4;
        const isEnabled = !entry.disable;
        const isConstant = !!entry.constant;
        const keys = entry.key || [];

        const card = document.createElement('div');
        let typeClass = '';
        if (isEnabled) typeClass = isConstant ? 'type-blue' : 'type-green';

        card.className = `wb-card ${isEnabled ? '' : 'disabled'} ${typeClass}`;
        card.dataset.uid = entry.uid;
        card.dataset.index = index;
        card.draggable = false;
        // 预计算搜索文本
        card.dataset.searchText = `${entry.comment || ''} ${entry.content || ''}`.toLowerCase();

        const escapeHtml = (str) => (str || '').replace(/[&<>"']/g, m => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;','\'':'&#039;'}[m]));
        const curPosInt = typeof entry.position === 'number' ? entry.position : 1;
        const curPosStr = WI_POSITION_MAP[curPosInt] || 'after_character_definition';

        const corePositions = ['before_character_definition', 'after_character_definition', 'at_depth'];
        const allPosOptions = [
            { v: 'before_character_definition', t: '角色定义之前' },
            { v: 'after_character_definition', t: '角色定义之后' },
            { v: 'before_example_messages', t: '示例消息之前' },
            { v: 'after_example_messages', t: '示例消息之后' },
            { v: 'before_author_note', t: `作者注释之前` },
            { v: 'after_author_note', t: `作者注释之后` },
            { v: 'at_depth', t: '@D' }
        ];

        const showCoreOnly = corePositions.includes(curPosStr);
        const hasKeys = keys.length > 0;
        const showWarning = isEnabled && !isConstant && !hasKeys;
        const warningIcon = showWarning ? `<i class="fa-solid fa-circle-exclamation wb-keyword-warning" style="color:#ef4444!important; margin-right:6px!important; cursor:pointer!important;" data-wb-tooltip="警告：绿灯条目已启用但未设置关键词，将无法触发"></i>` : '';
        const showExcludeRecursionWarning = isEnabled && !isConstant && entry.excludeRecursion !== true;
        const excludeRecursionWarningIcon = showExcludeRecursionWarning ? `<i class="fa-solid fa-triangle-exclamation wb-exclude-recursion-warning" style="color:#eab308!important; margin-right:6px!important; cursor:pointer!important;" data-wb-tooltip="警告：此条目可能引发递归冲突，建议启用「不可递归」"></i>` : '';

        let optionsHtml = '';
        allPosOptions.forEach(opt => {
            if (showCoreOnly && !corePositions.includes(opt.v)) return;
            const selected = opt.v === curPosStr ? 'selected' : '';
            optionsHtml += `<option value="${opt.v}" ${selected}>${opt.t}</option>`;
        });

        card.innerHTML = `
            <div class="wb-card-header">
                <!-- 启用开关 -->
                <div class="wb-header-item wb-item-enable">
                    <div class="wb-ctrl-group"><label class="wb-switch"><input type="checkbox" class="inp-enable" ${isEnabled ? 'checked' : ''}><span class="wb-slider purple"></span></label></div>
                </div>
                <!-- 类型开关 -->
                <div class="wb-header-item wb-item-type">
                    <div class="wb-ctrl-group"><label class="wb-switch"><input type="checkbox" class="inp-type" ${isConstant ? 'checked' : ''}><span class="wb-slider blue"></span></label></div>
                </div>
                <!-- 标题输入和警告图标 -->
                <div class="wb-header-item wb-item-title">
                    <div class="wb-title-wrapper">
                        <input class="wb-inp-title inp-name" value="${escapeHtml(entry.comment)}" placeholder="条目名称 (Comment)">
                        <div class="wb-warning-container" style="margin-right:4px!important;">${warningIcon}</div>
                        <div class="wb-exclude-recursion-warning-container">${excludeRecursionWarningIcon}</div>
                    </div>
                </div>
                <!-- 位置和顺序 -->
                <div class="wb-header-item wb-item-pos-order">
                    <div class="wb-pos-wrapper ${curPosStr === 'at_depth' ? 'show-depth' : ''}">
                        <select id="wb-pos-${entry.uid}" class="wb-input-dark inp-pos" style="font-size:0.85em">${optionsHtml}</select>
                        <input type="number" class="wb-inp-num inp-pos-depth" style="display: ${curPosStr === 'at_depth' ? 'block' : 'none'};" placeholder="D" value="${entry.depth ?? 4}">
                    </div>
                    <div class="wb-ctrl-group order-group" title="顺序"><span>顺序</span><input type="number" class="wb-inp-num inp-order" style="width:65px;height:24px;font-size:0.85em" value="${entry.order ?? 0}"></div>
                </div>
                <!-- Token 显示 -->
                <div class="wb-header-item wb-item-token">
                    <div class="wb-input-dark wb-token-display" title="Tokens">${Actions.getTokenCount(entry.content)}</div>
                </div>
                <!-- 预览和删除按钮 -->
                <div class="wb-header-item wb-item-actions">
                    <i class="fa-solid fa-eye btn-preview" style="cursor:pointer!important;padding:5px!important;color:#6b7280!important;" title="编辑内容"></i>
                    <i class="fa-solid fa-trash btn-delete" style="cursor:pointer!important;padding:5px!important;margin-left:5px!important;color:#6b7280!important;" title="删除条目"></i>
                </div>
            </div>
        `;

        // 为警告图标添加点击事件
        const keywordWarningIcon = card.querySelector('.wb-keyword-warning');
        if (keywordWarningIcon) {
            keywordWarningIcon.onclick = (e) => {
                e.stopPropagation();
                this.openWarningListModal();
            };
        }
        const erWarningIcon = card.querySelector('.wb-exclude-recursion-warning');
        if (erWarningIcon) {
            erWarningIcon.onclick = (e) => {
                e.stopPropagation();
                this.openExcludeRecursionModal();
            };
        }

        // 事件委托：不在卡片上绑定，改为在父容器统一处理

        return card;
    },

    openContentPopup(entry, triggerBtn) {
        const old = document.getElementById('wb-content-popup-overlay');
        if (old) old.remove();

        const overlay = document.createElement('div');
        overlay.id = 'wb-content-popup-overlay';
        overlay.className = 'wb-modal-overlay';
        const popup = document.createElement('div');
        popup.className = 'wb-content-popup';

        let tempContent = entry.content || '';
        let tempKeys = (entry.key || []).map(k => String(k).replace(/，/g, ',')).join(',');
        const escapeHtml = (str) => (str || '').replace(/[&<>\"']/g, m => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;','\'':'&#039;'}[m]));

        popup.innerHTML = `
            <div class="wb-popup-header"><span>${entry.comment || '未命名条目'}</span></div>
            <input class="wb-popup-input-keys" placeholder="关键词 (英文逗号分隔)" value="${escapeHtml(tempKeys)}">
            <textarea class="wb-popup-textarea" placeholder="在此编辑内容...">${escapeHtml(tempContent)}</textarea>
            <div class="wb-popup-footer"><button class="wb-btn-black btn-cancel">取消</button><button class="wb-btn-black btn-save">保存</button></div>
        `;
        overlay.appendChild(popup);
        document.body.appendChild(overlay);

        const keysInput = popup.querySelector('.wb-popup-input-keys');
        const textarea = popup.querySelector('.wb-popup-textarea');
        textarea.oninput = (e) => { tempContent = e.target.value; };
        keysInput.oninput = (e) => { tempKeys = e.target.value; };

        const close = () => overlay.remove();
        popup.querySelector('.btn-cancel').onclick = close;
        popup.querySelector('.btn-save').onclick = () => {
            Actions.updateEntry(entry.uid, d => d.content = tempContent);
            const finalKeys = tempKeys.replace(/，/g, ',').split(',').map(s => s.trim()).filter(Boolean);
            Actions.updateEntry(entry.uid, d => { d.key = finalKeys; });
            this.updateCardStatus(entry.uid);
            this.renderGlobalStats();
            close();
        };
        overlay.onclick = (e) => { if (e.target === overlay) close(); };
    },

    compareGroupNames(a, b) {
        const systemOrder = {
            '角色定义之前': 10, '角色定义': 20, '角色定义之后': 30, '普通': 40,
            '[InitVar]1st': 45, '作者注释之前': 50, '作者注释': 60, '作者注释之后': 70
        };
        const weightA = systemOrder[a] || 9999;
        const weightB = systemOrder[b] || 9999;
        if (weightA !== 9999 || weightB !== 9999) return weightA - weightB;

        const isAD = (str) => str.startsWith('@D');
        if (isAD(a) && isAD(b)) {
            const numA = parseInt(a.replace(/[^0-9]/g, '')) || 0;
            const numB = parseInt(b.replace(/[^0-9]/g, '')) || 0;
            return numA - numB;
        }
        if (isAD(a)) return 1;
        if (isAD(b)) return -1;
        return a.localeCompare(b);
    },

    renderManageView(filterText = '') {
        const container = document.getElementById('wb-manage-content');
        if (!container) return;
        const term = filterText.toLowerCase();

        // 检测筛选条件是否变化
        const currentFilters = JSON.stringify(STATE.manageFilters);
        const filtersChanged = currentFilters !== (this._lastManageFilters || '');
        this._lastManageFilters = currentFilters;

        // 快速增量搜索路径：仅当没有脏标记、未切换标签、且筛选条件未变化时使用
        if (!STATE.isManageDirty && container.children.length > 0 && this._lastManageTab === STATE.manageTab && !filtersChanged) {
            container.querySelectorAll('.wb-group').forEach(groupDiv => {
                let visibleCount = 0;
                groupDiv.querySelectorAll('.wb-manage-card').forEach(card => {
                    const searchText = card.dataset.searchText || '';
                    const match = searchText.includes(term);
                    card.style.display = match ? '' : 'none';
                    if (match) visibleCount++;
                });
                groupDiv.style.display = visibleCount > 0 ? '' : 'none';
                const countEl = groupDiv.querySelector('.wb-group-count');
                if (countEl) countEl.textContent = visibleCount;
            });
            return;
        }

        // 下面是完整的重新渲染逻辑
        this._lastManageTab = STATE.manageTab;
        STATE.isManageDirty = false;

        container.classList.toggle('mode-custom', STATE.manageTab === 'custom');
        container.classList.toggle('mode-worldbook', STATE.manageTab === 'worldbook');

        const renderGroup = (groupName, items, isCustomTab = false, customBoundMap = null) => {
            if (!isCustomTab && items.length === 0) return;

            if (STATE.manageTab === 'worldbook' || STATE.manageTab === 'custom') {
                items.sort((a, b) => {
                    const pinA = STATE.metadata[a]?.pinned ? 1 : 0;
                    const pinB = STATE.metadata[b]?.pinned ? 1 : 0;
                    return pinB - pinA || a.localeCompare(b);
                });
            } else {
                items.sort((a, b) => a.localeCompare(b));
            }

            const groupDiv = document.createElement('div');
            groupDiv.className = 'wb-group';
            const isSystem = groupName === '已绑定' || groupName === '未绑定' || groupName === '未标签化';
            const shouldExpand = term.length > 0 || STATE.manageExpandedGroups.has(groupName);

            groupDiv.innerHTML = `<div class="wb-group-header ${shouldExpand ? 'expanded' : ''}"><span class="wb-group-title ${isSystem ? 'system' : ''}">${groupName}</span><div style="display:flex;align-items:center"><span class="wb-group-count">${items.length}</span><i class="fa-solid fa-chevron-right wb-group-arrow"></i></div></div><div class="wb-group-body ${shouldExpand ? 'show' : ''}"></div>`;
            const header = groupDiv.querySelector('.wb-group-header');
            const body = groupDiv.querySelector('.wb-group-body');

            header.onclick = (e) => {
                const isExpanding = !header.classList.contains('expanded');
                header.classList.toggle('expanded', isExpanding);
                body.classList.toggle('show', isExpanding);
                if (isExpanding) STATE.manageExpandedGroups.add(groupName);
                else STATE.manageExpandedGroups.delete(groupName);
            };

            items.forEach(itemName => {
                if (STATE.manageTab === 'worldbook' || STATE.manageTab === 'custom' || isCustomTab) {
                    const bookName = itemName;
                    const meta = STATE.metadata[bookName] || {};
                    const bindInfo = customBoundMap ? (customBoundMap[bookName] || { primary: [], aux: [], global: false, chat: [] }) : (STATE.boundBooksSet[bookName] || { primary: [], aux: [], global: false, chat: [] });

                    let isVisibleInBindGroup = false;
                    let subTitleHtml = '';

                    if (STATE.manageTab === 'custom' || isCustomTab) {
                        // 自定义标签页不显示标签信息栏
                    } else if (groupName === '主要世界书') {
                        // 只要有绑定角色就显示
                        if (bindInfo.primary && bindInfo.primary.length > 0) {
                            isVisibleInBindGroup = true;
                            // 使用竖线分割角色名称
                            const charNames = bindInfo.primary.join(' | ');
                            subTitleHtml = `<div class="wb-card-subtitle bind-info-row wb-bind-multiple" style="font-size:0.8em; color:#70a1ff; margin-top:4px; cursor:pointer;" title="绑定角色: ${charNames}">${charNames}</div>`;
                        } else {
                            isVisibleInBindGroup = false;
                        }
                        if (!isVisibleInBindGroup) return;
                    } else if (groupName === '非主要世界书') {
                        // 直接显示，不需要副标题
                        isVisibleInBindGroup = true;
                    }

                    const card = document.createElement('div');
                    card.className = 'wb-manage-card';
                    // 给卡片附加上用于快速过滤的文本缓存
                    card.dataset.searchText = `${bookName} ${meta.note || ''}`.toLowerCase();
                    if (meta.pinned) card.style.borderLeft = `3px solid ${CONFIG.colors.accent}`;

                    let iconsHtml = '';
                    iconsHtml += `<div class="wb-icon-action btn-view" title="跳转到编辑"><i class="fa-solid fa-eye"></i></div>`;
                    iconsHtml += `<div class="wb-icon-action btn-del" title="删除世界书"><i class="fa-solid fa-trash"></i></div>`;

                    if (isCustomTab) {
                        if (groupName !== '未标签化') {
                            iconsHtml += `<div class="wb-icon-action btn-tag" title="从当前标签分组移除"><i class="fa-solid fa-tag"></i></div>`;
                        }
                    } else {
                        iconsHtml += `<div class="wb-icon-action btn-pin ${meta.pinned ? 'pinned' : ''}" title="${meta.pinned ? '取消顶置' : '组内顶置'}"><i class="fa-solid fa-thumbtack"></i></div>`;
                    }

                    let titleHtml = `<span class="wb-card-title">${bookName}</span>${subTitleHtml}`;
                    card.innerHTML = `<div class="wb-card-top"><div class="wb-card-info">${titleHtml}</div><div class="wb-manage-icons">${iconsHtml}</div></div>`;

                    const q = (s) => card.querySelector(s);
                    const subEl = q('.wb-card-subtitle');
                    if (subEl && STATE.manageTab !== 'custom' && !isCustomTab) {
                        subEl.onclick = () => this.showBindingDetailsModal(bookName, bindInfo);
                        subEl.onmouseover = () => subEl.style.opacity = '0.8';
                        subEl.onmouseout = () => subEl.style.opacity = '1';
                    }

                    q('.btn-view').onclick = () => Actions.jumpToEditor(bookName);
                    q('.btn-del').onclick = () => Actions.deleteBookDirectly(bookName);
                    if (isCustomTab) {
                        q('.btn-tag').onclick = (e) => {
                            e.stopPropagation();
                            if (confirm(`是否将世界书 "${bookName}" 移除标签分组 "${groupName}"？`)) {
                                Actions.removeTagFromWorldbook(bookName, groupName);
                            }
                        };
                    } else {
                        q('.btn-pin').onclick = () => Actions.togglePin(bookName);
                    }

                    body.appendChild(card);
                }
            });
            container.appendChild(groupDiv);
        };

        if (STATE.manageTab === 'custom') {
            const groups = {};
            // 收集所有在元数据中出现的标签，无论对应世界书是否存在
            Object.entries(STATE.metadata).forEach(([bookName, meta]) => {
                const tags = meta.tags || [];
                tags.forEach(tag => {
                    if (!groups[tag]) groups[tag] = [];
                    groups[tag].push(bookName);
                });
            });

            container.innerHTML = '';
            const groupNames = Object.keys(groups).sort((a, b) => a.localeCompare(b));
            if (groupNames.length > 0) {
                groupNames.forEach(tagName => {
                    const allBooksWithTag = groups[tagName];
                    // 只显示在当前世界书列表中的书
                    const validBooks = allBooksWithTag.filter(name => STATE.allBookNames.includes(name));
                    // 只要该标签在元数据中存在（即 allBooksWithTag 非空），就显示分组
                    renderGroup(tagName, validBooks, true, STATE.boundBooksSet);
                });
            } else {
                container.innerHTML += `<div style="text-align:center; color:#9ca3af; padding:40px;">暂无自定义分组<br><span style="font-size:0.9em">请点击上方"管理标签"添加</span></div>`;
            }

            // 未标签化的书：所有无标签的有效世界书
            const ungrouped = STATE.allBookNames.filter(name => {
                const meta = STATE.metadata[name] || {};
                const tags = meta.tags || [];
                return tags.length === 0;
            });
            if (ungrouped.length > 0) renderGroup('未标签化', ungrouped, true, STATE.boundBooksSet);
            return;
        }

        const boundMap = STATE.boundBooksSet || {};
        const groups = { '主要世界书': [], '非主要世界书': [] };
        if (STATE.manageTab === 'worldbook') {
            Actions.getExistingGroups().forEach(g => groups[g] = []);
            STATE.allBookNames.forEach(name => {
                const meta = STATE.metadata[name] || {};
                if (term && !name.toLowerCase().includes(term) && !(meta.note || '').toLowerCase().includes(term)) return;
                let gName = meta.group;
                const bindInfo = boundMap[name];

                // 判断是否被任何角色作为主要世界书绑定
                const isPrimaryBound = bindInfo && bindInfo.primary && bindInfo.primary.length > 0;

                // 如果未分组，则归入我们定义好的两个大类
                if (!gName || gName === '未分组' || gName === '已绑定角色' || gName === '未绑定角色' || gName === '已绑定' || gName === '未绑定') {
                    gName = isPrimaryBound ? '主要世界书' : '非主要世界书';
                }

                if (!groups[gName]) groups[gName] = [];
                groups[gName].push(name);
            });
        }

        container.innerHTML = '';
        if (groups['非主要世界书'] && groups['非主要世界书'].length > 0) renderGroup('非主要世界书', groups['非主要世界书']);
        Object.keys(groups).sort(this.compareGroupNames.bind(this)).forEach(g => {
            if (g !== '主要世界书' && g !== '非主要世界书') renderGroup(g, groups[g]);
        });
        if (groups['主要世界书'] && groups['主要世界书'].length > 0) renderGroup('主要世界书', groups['主要世界书']);

        // Old-style global settings: append directly to container, no negative margin
        const config = Actions.getGlobalConfig();
        const settingsDiv = document.createElement('div');
        settingsDiv.className = 'wb-manage-settings';
        settingsDiv.innerHTML = `
            <div class="wb-setting-row">
                <div>
                    <div class="wb-setting-label">联级删除主要世界书</div>
                    <div class="wb-setting-desc">删除角色卡时，询问是否也删除其绑定的主要世界书</div>
                </div>
                <div class="wb-ctrl-group">
                    <label class="wb-switch"><input type="checkbox" id="wb-setting-del-wb" ${config.deleteWbWithChar ? 'checked' : ''}><span class="wb-slider purple"></span></label>
                </div>
            </div>`;
        container.appendChild(settingsDiv);

        // 绑定事件
        settingsDiv.querySelector('#wb-setting-del-wb').onchange = async (e) => {
            await Actions.saveGlobalConfig({ deleteWbWithChar: e.target.checked });
        };

        // 如果在重建时本身就带有搜索词，立刻执行一遍快速隐藏
        if (term) {
            container.querySelectorAll('.wb-group').forEach(groupDiv => {
                let visibleCount = 0;
                groupDiv.querySelectorAll('.wb-manage-card').forEach(card => {
                    const searchText = card.dataset.searchText || '';
                    const match = searchText.includes(term);
                    card.style.display = match ? '' : 'none';
                    if (match) visibleCount++;
                });
                groupDiv.style.display = visibleCount > 0 ? '' : 'none';
                const countEl = groupDiv.querySelector('.wb-group-count');
                if (countEl) countEl.textContent = visibleCount;
            });
        }
    },

    applyCustomDropdown(selectId) {
        const originalSelect = document.getElementById(selectId);
        if (!originalSelect) return;
        let trigger = document.getElementById(`wb-trigger-${selectId}`);
        if (originalSelect.style.display !== 'none') {
            originalSelect.style.display = 'none';
            if (trigger) trigger.remove();
            trigger = document.createElement('div');
            trigger.id = `wb-trigger-${selectId}`;
            trigger.className = 'wb-gr-trigger';
            originalSelect.parentNode.insertBefore(trigger, originalSelect.nextSibling);
            trigger.onclick = (e) => { e.stopPropagation(); this.toggleCustomDropdown(selectId, trigger); };
        }
        const update = () => {
            const selectedOpt = originalSelect.options[originalSelect.selectedIndex];
            trigger.innerHTML = `<span style="overflow:hidden; text-overflow:ellipsis; white-space:nowrap; flex:1; text-align:left;">${selectedOpt ? selectedOpt.text : '请选择...'}</span>`;
        };
        update();
        originalSelect.addEventListener('change', update);
    },

    toggleCustomDropdown(selectId, triggerElem) {
        const existing = document.getElementById('wb-active-dropdown');
        if (existing) {
            const isSame = existing.dataset.source === selectId;
            existing.remove();
            if (isSame) return;
        }
        const originalSelect = document.getElementById(selectId);
        const dropdown = document.createElement('div');
        dropdown.id = 'wb-active-dropdown';
        dropdown.className = 'wb-gr-dropdown show';
        dropdown.dataset.source = selectId;

        // 判断是否为位置选择器（位置选择器不需要搜索框和标签筛选）
        const isPositionSelector = selectId.startsWith('wb-pos-');

        let searchBox, searchInput, tagFilter;

        // 只有非位置选择器才显示搜索框和标签筛选
        if (!isPositionSelector) {
            searchBox = document.createElement('div');
            searchBox.className = 'wb-gr-search-box';
            searchBox.style.position = 'relative';
            searchInput = document.createElement('input');
            searchInput.className = 'wb-gr-search-input';
            searchInput.placeholder = '搜索世界书...';
            searchInput.onclick = (e) => e.stopPropagation();
            searchBox.appendChild(searchInput);

            // Tag filter button
            tagFilter = document.createElement('div');
            tagFilter.className = 'wb-gr-tag-filter';
            tagFilter.title = '按标签筛选';
            tagFilter.innerHTML = '<i class="fa-solid fa-tag"></i>';
            searchBox.appendChild(tagFilter);
        }

        const optionsContainer = document.createElement('div');
        optionsContainer.className = 'wb-gr-options-container';

        let currentTag = null;

        const updateTagFilterButton = () => {
            if (!tagFilter) return;
            if (currentTag) {
                tagFilter.innerHTML = `<span class="wb-gr-tag-pill">${currentTag}<span class="wb-gr-tag-close">&times;</span></span>`;
                const closeEl = tagFilter.querySelector('.wb-gr-tag-close');
                if (closeEl) {
                    closeEl.onclick = (e) => {
                        e.stopPropagation();
                        currentTag = null;
                        updateTagFilterButton();
                        applyFilters();
                    };
                }
            } else {
                tagFilter.innerHTML = '<i class="fa-solid fa-tag"></i>';
            }
        };

        const applyFilters = () => {
            if (!searchInput) return;
            const term = searchInput.value.toLowerCase();
            const metadata = STATE.metadata || {};
            optionsContainer.querySelectorAll('.wb-gr-option').forEach(o => {
                const textMatch = o.textContent.toLowerCase().includes(term);
                const bookName = o.dataset.value;
                const tags = metadata[bookName]?.tags || [];
                const tagMatch = !currentTag || tags.includes(currentTag);
                o.classList.toggle('hidden', !(textMatch && tagMatch));
            });
        };

        const createOption = (optNode) => {
            const div = document.createElement('div');
            div.className = 'wb-gr-option';
            div.textContent = optNode.text;
            div.dataset.value = optNode.value;
            if (optNode.selected) div.classList.add('selected');
            div.onclick = (e) => {
                e.stopPropagation();
                originalSelect.value = optNode.value;
                originalSelect.dispatchEvent(new Event('change', { bubbles: true }));
                dropdown.remove();
            };
            optionsContainer.appendChild(div);
        };

        const showTagMenu = () => {
            if (!tagFilter) return;
            // 首次点击标签筛选按钮时显示提示
            const tagFilterHintKey = '[SillyTavern]世界书管理器-标签筛选提示';
            if (!localStorage.getItem(tagFilterHintKey)) {
                localStorage.setItem(tagFilterHintKey, 'true');
                this.showTagFilterHintModal();
            }

            const existingMenu = document.getElementById('wb-gr-tag-menu-overlay');
            if (existingMenu) {
                existingMenu.remove();
                return;
            }
            const menu = document.createElement('div');
            menu.id = 'wb-gr-tag-menu-overlay';
            menu.className = 'wb-gr-tag-menu-overlay';
            const meta = STATE.metadata || {};
            const allTags = [...new Set(Object.values(meta).flatMap(m => m.tags || []))].sort();
            const allOption = document.createElement('div');
            allOption.className = 'wb-gr-tag-option';
            allOption.textContent = '全部';
            allOption.onclick = (e) => {
                e.stopPropagation();
                currentTag = null;
                updateTagFilterButton();
                applyFilters();
                menu.remove();
            };
            menu.appendChild(allOption);
            allTags.forEach(tag => {
                const opt = document.createElement('div');
                opt.className = 'wb-gr-tag-option';
                opt.textContent = tag;
                opt.onclick = (e) => {
                    e.stopPropagation();
                    currentTag = tag;
                    updateTagFilterButton();
                    applyFilters();
                    menu.remove();
                };
                menu.appendChild(opt);
            });
            document.body.appendChild(menu);
            const rect = tagFilter.getBoundingClientRect();
            menu.style.position = 'fixed';
            menu.style.top = `${rect.bottom + 4}px`;
            menu.style.right = `${window.innerWidth - rect.right}px`;
            menu.style.left = 'auto';
            const closeMenu = (e) => {
                if (!menu.contains(e.target)) {
                    menu.remove();
                    document.removeEventListener('click', closeMenu);
                }
            };
            setTimeout(() => document.addEventListener('click', closeMenu), 0);
        };

        if (tagFilter) {
            tagFilter.onclick = (e) => {
                e.stopPropagation();
                showTagMenu();
            };
        }

        Array.from(originalSelect.children).forEach(child => {
            if (child.tagName === 'OPTGROUP') {
                const label = document.createElement('div');
                label.className = 'wb-gr-group-label';
                label.textContent = child.label;
                optionsContainer.appendChild(label);
                Array.from(child.children).forEach(createOption);
            } else if (child.tagName === 'OPTION') createOption(child);
        });

        // 只有非位置选择器才显示搜索框
        if (searchBox) {
            dropdown.appendChild(searchBox);
        }
        dropdown.appendChild(optionsContainer);
        document.body.appendChild(dropdown);

        const rect = triggerElem.getBoundingClientRect();
        dropdown.style.top = `${rect.bottom + 5}px`;
        dropdown.style.left = `${rect.left}px`;
        dropdown.style.width = `${rect.width}px`;

        // 只有非位置选择器才设置搜索输入事件
        if (searchInput) {
            // 不自动聚焦搜索框，避免移动端弹出虚拟键盘
            // const isTouchDevice = window.matchMedia('(pointer: coarse)').matches;
            // if (!isTouchDevice) searchInput.focus();

            searchInput.oninput = applyFilters;
            applyFilters();
        }

        const closeHandler = (e) => {
            if (!dropdown.contains(e.target) && e.target !== triggerElem) {
                dropdown.remove();
                document.removeEventListener('click', closeHandler);
            }
        };
        setTimeout(() => document.addEventListener('click', closeHandler), 0);
    },

    openSortingModal() {
        document.getElementById('btn-group-sort')?.blur();
        if (!STATE.currentBookName) return toastr.warning("请先选择一本世界书");

        const groups = {};
        const groupKeys = [];
        const priorityMap = {
            'before_character_definition': 10, 'after_character_definition': 20,
            'before_author_note': 30, 'after_author_note': 40,
            'at_depth': 50, 'before_example_messages': 60, 'after_example_messages': 70
        };
        const typeLabels = {
            'before_character_definition': '角色定义之前', 'after_character_definition': '角色定义之后',
            'before_example_messages': '示例消息之前', 'after_example_messages': '示例消息之后',
            'before_author_note': '作者注释之前', 'after_author_note': '作者注释之后', 'at_depth': '@D'
        };

        const sortedEntries = [...STATE.entries].sort((a, b) => (a.order ?? 0) - (b.order ?? 0));
        sortedEntries.forEach(entry => {
            const posInt = typeof entry.position === 'number' ? entry.position : 1;
            const posStr = WI_POSITION_MAP[posInt] || 'after_character_definition';
            let key = posStr === 'at_depth' ? `at_depth_${entry.depth ?? 0}` : posStr;
            let label = posStr === 'at_depth' ? `@D ${entry.depth ?? 0}` : (typeLabels[key] || key);
            const rawType = posStr;
            const depthVal = entry.depth ?? 0;

            if (!groups[key]) {
                groups[key] = { label, items: [], rawType, depthVal };
                groupKeys.push(key);
            }
            groups[key].items.push(entry);
        });

        groupKeys.sort((keyA, keyB) => {
            const gA = groups[keyA];
            const gB = groups[keyB];
            const pA = priorityMap[gA.rawType] ?? 999;
            const pB = priorityMap[gB.rawType] ?? 999;
            if (pA !== pB) return pA - pB;
            if (gA.rawType === 'at_depth') return gB.depthVal - gA.depthVal;
            return 0;
        });

        const overlay = document.createElement('div');
        overlay.className = 'wb-sort-modal-overlay';
        overlay.innerHTML = `<div class="wb-sort-modal"><div class="wb-sort-header"><span><i class="fa-solid fa-arrow-down-9-1"></i> 分组排序管理</span><div style="cursor:pointer" id="wb-sort-close"><i class="fa-solid fa-xmark"></i></div></div><div class="wb-sort-body" id="wb-sort-body"></div><div class="wb-sort-footer" style="display:flex; justify-content:center; gap:15px;"><button class="wb-btn-rect" id="wb-sort-cancel" style="font-size:0.9em;padding:8px 20px; background:#fff; color:#000; border:1px solid #e5e7eb;">取消</button><button class="wb-btn-rect" id="wb-sort-save" style="font-size:0.9em;padding:8px 20px">保存</button></div></div>`;
        document.body.appendChild(overlay);

        const bodyEl = overlay.querySelector('#wb-sort-body');
        const isDark = () => document.body.getAttribute('data-theme') === 'dark';
        const getBg = (i) => isDark() ? `hsl(${(i * 137.5) % 360}, 20%, 18%)` : `hsl(${(i * 137.5) % 360}, 70%, 95%)`;
        const getBdr = (i) => isDark() ? `hsl(${(i * 137.5) % 360}, 40%, 28%)` : `hsl(${(i * 137.5) % 360}, 60%, 80%)`;
        const getTxt = (i) => isDark() ? `hsl(${(i * 137.5) % 360}, 80%, 85%)` : `hsl(${(i * 137.5) % 360}, 80%, 30%)`;

        groupKeys.forEach((key, i) => {
            const group = groups[key];
            const container = document.createElement('div');
            container.className = 'wb-sort-group-container';
            container.style.backgroundColor = getBg(i);
            container.style.borderColor = getBdr(i);

            container.innerHTML = `
                <div class="wb-sort-group-title" style="color:${getTxt(i)}"><span>${group.label} <span style="font-weight:normal;font-size:0.8em;opacity:0.8">(${group.items.length})</span></span><i class="fa-solid fa-chevron-down wb-sort-arrow"></i></div>
                <div class="wb-sort-group-list" data-group-key="${key}"></div>`;

            const titleEl = container.querySelector('.wb-sort-group-title');
            const listEl = container.querySelector('.wb-sort-group-list');

            titleEl.onclick = () => {
                const isCollapsed = listEl.classList.contains('collapsed');
                if (isCollapsed) { listEl.classList.remove('collapsed'); titleEl.classList.remove('collapsed'); } 
                else { listEl.classList.add('collapsed'); titleEl.classList.add('collapsed'); }
            };

            const itemsHtml = group.items.map(entry => {
                const safeTitle = (entry.comment || '无标题').replace(/&/g, '&amp;').replace(/</g, '&lt;');
                return `
                <div class="wb-sort-item" data-uid="${entry.uid}" data-group="${key}" draggable="true">
                    <div class="wb-sort-item-order">${entry.order ?? 0}</div>
                    <div style="flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;pointer-events:none;">${safeTitle}</div>
                    <div class="wb-sort-handle"><i class="fa-solid fa-bars" style="color:#ccc; pointer-events:none;"></i></div>
                </div>`;
            }).join('');

            listEl.innerHTML = itemsHtml;
            this.initSortableGroup(listEl, key);
            bodyEl.appendChild(container);
        });

        overlay.querySelector('#wb-sort-close').onclick = () => overlay.remove();
        overlay.querySelector('#wb-sort-cancel').onclick = () => overlay.remove();
        overlay.querySelector('#wb-sort-save').onclick = async () => {
            await API.saveBookEntries(STATE.currentBookName, STATE.entries);
            Actions.sortByPriority();
            overlay.remove();
        };
    },

    initSortableGroup(listEl, groupKey) {
        const updateOrder = () => {
            [...listEl.querySelectorAll('.wb-sort-item')].forEach((el, idx) => {
                const newOrder = idx + 1;
                el.querySelector('.wb-sort-item-order').textContent = newOrder;
                const entry = STATE.entries.find(e => e.uid === Number(el.dataset.uid));
                if (entry) { entry.order = newOrder; }
            });
        };

        listEl.addEventListener('dragstart', (e) => {
            const item = e.target.closest('.wb-sort-item');
            if (!item) return;
            e.dataTransfer.effectAllowed = 'move';
            e.dataTransfer.setData('text/uid', item.dataset.uid);
            e.dataTransfer.setData('text/group', groupKey);
            item.classList.add('pc-dragging');
        });

        listEl.addEventListener('dragend', (e) => {
            const item = e.target.closest('.wb-sort-item');
            if (item) item.classList.remove('pc-dragging');
            updateOrder();
        });

        listEl.addEventListener('dragover', (e) => {
            e.preventDefault(); 
            const dragging = listEl.querySelector('.pc-dragging');
            if (!dragging) return;
            const siblings = [...listEl.querySelectorAll('.wb-sort-item:not(.pc-dragging)')];
            const next = siblings.find(s => {
                const rect = s.getBoundingClientRect();
                return e.clientY <= rect.top + rect.height / 2;
            });
            listEl.insertBefore(dragging, next);
        });

        let touchItem = null;
        let touchTimer = null;
        let startX = 0, startY = 0;
        const TOUCH_TOLERANCE = 10;
        let rAF = null; 

        listEl.addEventListener('touchstart', (e) => {
            const handle = e.target.closest('.wb-sort-handle');
            if (!handle) return;
            e.preventDefault();
            const item = handle.closest('.wb-sort-item');
            if (!item) return;
            const touch = e.touches[0];
            startX = touch.clientX;
            startY = touch.clientY;
            touchItem = item;
            touchTimer = setTimeout(() => {
                if (touchItem) {
                    touchItem.classList.add('mobile-dragging');
                    if (navigator.vibrate) navigator.vibrate(50);
                    document.body.style.overflow = 'hidden'; 
                }
            }, 150);
        }, { passive: false });

        listEl.addEventListener('touchmove', (e) => {
            if (!touchItem) return;
            const touch = e.touches[0];
            if (!touchItem.classList.contains('mobile-dragging')) {
                const diffX = Math.abs(touch.clientX - startX);
                const diffY = Math.abs(touch.clientY - startY);
                if (diffX > TOUCH_TOLERANCE || diffY > TOUCH_TOLERANCE) {
                    clearTimeout(touchTimer);
                    touchItem = null;
                }
                return;
            }
            e.preventDefault();
            if (rAF) return;
            rAF = requestAnimationFrame(() => {
                rAF = null; 
                const target = document.elementFromPoint(touch.clientX, touch.clientY);
                if (!target) return;
                const swapItem = target.closest('.wb-sort-item');
                if (swapItem && swapItem !== touchItem && listEl.contains(swapItem)) {
                    const rect = swapItem.getBoundingClientRect();
                    const next = (touch.clientY - rect.top) / rect.height > 0.5;
                    const sibling = next ? swapItem.nextSibling : swapItem;
                    if (sibling !== touchItem && sibling !== touchItem.nextSibling) listEl.insertBefore(touchItem, sibling);
                }
            });
        }, { passive: false });

        const endDrag = () => {
            if (touchTimer) clearTimeout(touchTimer);
            if (touchItem) {
                touchItem.classList.remove('mobile-dragging');
                touchItem = null;
                document.body.style.overflow = ''; 
                updateOrder();
            }
        };

        listEl.addEventListener('touchend', endDrag);
        listEl.addEventListener('touchcancel', endDrag);
    },

    openAnalysisModal() {
        if (!STATE.currentBookName) return toastr.warning("请先选择一本世界书");

        let showAll = false;
        const overlay = document.createElement('div');
        overlay.className = 'wb-sort-modal-overlay';
        overlay.innerHTML = `
            <div class="wb-sort-modal" id="wb-analysis-box" style="width:550px; height:auto; max-height:85vh;">
                <div class="wb-sort-header" style=" padding: 15px 20px;">
                    <span style="font-size:1.1em; display:flex; align-items:center; gap:10px;"><i class="fa-solid fa-chart-pie" style="color:#374151;"></i><span id="wb-analysis-title">${STATE.currentBookName}</span></span>
                    <div style="display:flex; gap:15px; align-items:center;">
                        <i class="fa-solid fa-repeat wb-action-icon" id="wb-analysis-toggle" title="切换：仅已启用 / 所有条目"></i>
                        <div style="cursor:pointer" class="wb-close-modal"><i class="fa-solid fa-xmark"></i></div>
                    </div>
                </div>
                <div class="wb-sort-body" style="padding:0; overflow:hidden !important;"><div id="wb-analysis-content" class="wb-stats-container"></div></div>
            </div>`;
        document.body.appendChild(overlay);
        this.setupModalPositioning(overlay.querySelector('#wb-analysis-box'), overlay);

        const render = () => {
            const sourceEntries = STATE.entries;
            const targetEntries = showAll ? sourceEntries : sourceEntries.filter(e => e.disable === false);
            const titleEl = overlay.querySelector('#wb-analysis-title');
            titleEl.innerHTML = `${STATE.currentBookName} <span style="font-size:0.8em; font-weight:normal;">(${showAll ? '所有条目' : '仅已启用'})</span>`;

            if (targetEntries.length === 0) {
                overlay.querySelector('#wb-analysis-content').innerHTML = `<div style="text-align:center; color:#9ca3af; padding:40px;">暂无数据</div>`;
                return;
            }

            // 统计所有条目的tokens（用于右上角显示）
            let allEnabledTokens = 0, allDisabledTokens = 0;
            sourceEntries.forEach(entry => {
                const t = Actions.getTokenCount(entry.content);
                if (entry.disable) { allDisabledTokens += t; } else { allEnabledTokens += t; }
            });

            let blueTokens = 0, greenTokens = 0, blueCount = 0, greenCount = 0;
            const rankList = [];

            targetEntries.forEach(entry => {
                const t = Actions.getTokenCount(entry.content);
                const isBlue = !!entry.constant;
                const isDisabled = !!entry.disable;
                if (isBlue) { blueTokens += t; blueCount++; } else { greenTokens += t; greenCount++; }
                rankList.push({ name: entry.comment || '未命名', tokens: t, isBlue: isBlue, isDisabled: isDisabled, uid: entry.uid });
            });

            const totalTokens = blueTokens + greenTokens;
            const totalCount = blueCount + greenCount;
            const bluePercent = totalTokens > 0 ? (blueTokens / totalTokens * 100).toFixed(1) : 0;
            const greenPercent = totalTokens > 0 ? (greenTokens / totalTokens * 100).toFixed(1) : 0;
            const blueCountPercent = totalCount > 0 ? (blueCount / totalCount * 100).toFixed(1) : 0;
            const greenCountPercent = totalCount > 0 ? (greenCount / totalCount * 100).toFixed(1) : 0;

            rankList.sort((a, b) => {
                if (a.isBlue !== b.isBlue) return a.isBlue ? -1 : 1; 
                return b.tokens - a.tokens; 
            });

            // 右上角显示的总 tokens：根据 showAll 状态决定显示已启用还是全部
            const displayTotalTokens = showAll ? (allEnabledTokens + allDisabledTokens) : allEnabledTokens;

            const progressHtml = `
                <div class="wb-stats-row">
                    <div class="wb-stats-label"><span>Token 占比</span><span class="wb-stats-total">${displayTotalTokens}</span></div>
                    <div class="wb-progress-bar">
                        <div class="wb-bar-seg wb-bg-blue" style="width:${bluePercent}%">${blueTokens > 0 ? blueTokens : ''}</div>
                        <div class="wb-bar-seg wb-bg-green" style="width:${greenPercent}%">${greenTokens > 0 ? greenTokens : ''}</div>
                    </div>
                    <div class="wb-bar-legend">
                        <span><span class="wb-legend-dot wb-dot-blue"></span>蓝灯: ${bluePercent}%</span>
                        <span><span class="wb-legend-dot wb-dot-green"></span>绿灯: ${greenPercent}%</span>
                    </div>
                </div>`;

            const pieGradient = `conic-gradient(#3b82f6 0% ${blueCountPercent}%, #22c55e ${blueCountPercent}% 100%)`;
            const pieHtml = `
                <div class="wb-pie-row">
                    <div class="wb-pie-chart" style="background: ${pieGradient};"></div>
                    <div class="wb-pie-legend">
                        <div class="wb-pie-legend-item"><span class="wb-legend-dot wb-dot-blue"></span> 蓝灯条目: <strong>${blueCount}</strong> <span style="font-size:0.9em;color:#6b7280;margin-left:4px">(${blueCountPercent}%)</span></div>
                        <div class="wb-pie-legend-item"><span class="wb-legend-dot wb-dot-green"></span> 绿灯条目: <strong>${greenCount}</strong> <span style="font-size:0.9em;color:#6b7280;margin-left:4px">(${greenCountPercent}%)</span></div>
                        <div class="wb-pie-sub">共 ${totalCount} 条</div>
                    </div>
                </div>`;

            let rankHtmlItems = '';
            rankList.forEach(item => {
                const percent = totalTokens > 0 ? (item.tokens / totalTokens * 100).toFixed(1) : 0;
                const isDark = document.body.getAttribute('data-theme') === 'dark';
                let barColor, trackColor;
                if (isDark) {
                    barColor = item.isBlue ? 'rgba(59, 130, 246, 0.25)' : 'rgba(34, 197, 94, 0.25)';
                    trackColor = '#2b2b2b'; 
                } else {
                    barColor = item.isBlue ? '#dbeafe' : '#dcfce7';
                    trackColor = '#f8fafc';
                }
                const bgStyle = `background: linear-gradient(to right, ${barColor} ${percent}%, ${trackColor} ${percent}%);`;
                const disabledStyle = item.isDisabled ? 'font-style:italic; text-decoration:underline;' : '';

                rankHtmlItems += `
                    <div class="wb-rank-pill" style="${bgStyle}">
                        <div class="wb-rank-pill-name" title="${item.name}" style="${disabledStyle}">${item.name}</div><div class="wb-rank-pill-val" style="${disabledStyle}">${item.tokens}</div>
                    </div>`;
            });

            // 右下角显示：根据 showAll 状态决定显示格式
            const displayRankTotal = showAll
                ? `${allEnabledTokens} + <span style="font-style:italic;">${allDisabledTokens}</span>`
                : `${allEnabledTokens}`;

            const rankHtml = `
                <div class="wb-stats-row" style="flex:1; min-height:0;">
                    <div class="wb-stats-label"><span>Token 排行 (蓝前绿后)</span><span class="wb-stats-total">${displayRankTotal}</span></div>
                    <div class="wb-rank-list">${rankHtmlItems}</div>
                </div>`;

            overlay.querySelector('#wb-analysis-content').innerHTML = progressHtml + pieHtml + rankHtml;
        };

        const toggleBtn = overlay.querySelector('#wb-analysis-toggle');
        toggleBtn.onclick = () => { showAll = !showAll; render(); };
        overlay.querySelector('.wb-close-modal').onclick = () => overlay.remove();
        overlay.onclick = (e) => { if (e.target === overlay) overlay.remove(); };

        render();
    },

    openLogViewerModal() {
        const escapeHtml = (str) => (str || '').replace(/[&<>"']/g, m => ({
            '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;'
        })[m]);

        const overlay = document.createElement('div');
        overlay.className = 'wb-sort-modal-overlay';
        overlay.innerHTML = `
            <div class="wb-sort-modal" style="width:700px; max-height:90vh; display:flex; flex-direction:column; background:#fff; border-radius:12px; overflow:hidden;">
                <div class="wb-sort-header" style="padding:15px; border-bottom:1px solid #e5e7eb; display:flex; justify-content:space-between; align-items:center;">
                    <span style="font-weight:bold; font-size:1.1em;"><i class="fa-solid fa-file-lines"></i> 插件日志</span>
                    <div style="display:flex; gap:10px; align-items:center;">
                        <button id="btn-copy-logs" class="wb-btn-rect" style="font-size:0.9em;padding:6px 15px; background:#fff; color:#374151; border:1px solid #d1d5db;">复制</button>
                        <button id="btn-download-logs" class="wb-btn-rect" style="font-size:0.9em;padding:6px 15px; background:#fff; color:#374151; border:1px solid #d1d5db;">下载</button>
                        <div class="wb-close-modal" style="cursor:pointer"><i class="fa-solid fa-xmark"></i></div>
                    </div>
                </div>
                <div class="wb-sort-body" style="padding:0; overflow:hidden;">
                    <div class="wb-log-filters" style="display:flex; gap:15px; padding:10px; background:rgba(0,0,0,0.05); border-bottom:1px solid var(--SmartThemeBorderColor);">
                        <label style="cursor:pointer; display:flex; align-items:center; gap:5px;"><input type="radio" name="log-level" value="all" checked> 全部</label>
                        <label style="cursor:pointer; display:flex; align-items:center; gap:5px;"><input type="radio" name="log-level" value="info"> 信息</label>
                        <label style="cursor:pointer; display:flex; align-items:center; gap:5px;"><input type="radio" name="log-level" value="warn"> 警告</label>
                        <label style="cursor:pointer; display:flex; align-items:center; gap:5px;"><input type="radio" name="log-level" value="error"> 错误</label>
                    </div>
                    <div class="wb-log-list" style="flex:1; overflow-y:auto; padding:10px; background:#fff;"></div>
                </div>
            </div>
        `;
        document.body.appendChild(overlay);
        this.setupModalPositioning(overlay.querySelector('.wb-sort-modal'), overlay);

        const logListEl = overlay.querySelector('.wb-log-list');
        const levelRadios = overlay.querySelectorAll('input[name="log-level"]');

        const renderLogs = (filterLevel) => {
            const logs = logger.getLogs(filterLevel === 'all' ? null : filterLevel);
            logListEl.innerHTML = '';
            if (logs.length === 0) {
                logListEl.innerHTML = '<div style="text-align:center; color:#9ca3af; padding:20px;">暂无日志</div>';
                return;
            }
            logs.forEach(log => {
                const entry = document.createElement('div');
                entry.style.padding = '6px 10px';
                entry.style.borderBottom = '1px solid #f3f4f6';
                entry.style.fontSize = '0.9em';
                entry.style.fontFamily = 'monospace';
                let levelColor = '#000';
                if (log.level === 'error') levelColor = '#ef4444';
                else if (log.level === 'warn') levelColor = '#f59e0b';
                else if (log.level === 'info') levelColor = '#3b82f6';
                else if (log.level === 'debug') levelColor = '#6b7280';

                let displayMsg = log.message;
                if (log.args && log.args.length > 0) {
                    displayMsg += ' ' + log.args.map(arg =>
                        typeof arg === 'object' ? JSON.stringify(arg) : String(arg)
                    ).join(' ');
                }

                entry.innerHTML = `<span style="color:#9ca3af; margin-right:8px;">${log.timestamp}</span><span style="color:${levelColor}; font-weight:bold; margin-right:8px;">[${log.level.toUpperCase()}]</span><span>${escapeHtml(displayMsg)}</span>`;
                logListEl.appendChild(entry);
            });
        };

        // Initial render
        renderLogs('all');

        // Filter change
        levelRadios.forEach(radio => {
            radio.addEventListener('change', (e) => {
                if (e.target.checked) renderLogs(e.target.value);
            });
        });

        // Copy button
        overlay.querySelector('#btn-copy-logs').onclick = () => {
            const filterLevel = overlay.querySelector('input[name="log-level"]:checked').value;
            const logs = logger.getLogs(filterLevel === 'all' ? null : filterLevel);
            const text = logs.map(log => {
                let line = `[${log.timestamp}] [${log.level.toUpperCase()}] ${log.message}`;
                if (log.args && log.args.length > 0) {
                    line += ' ' + log.args.map(arg =>
                        typeof arg === 'object' ? JSON.stringify(arg) : String(arg)
                    ).join(' ');
                }
                return line;
            }).join('\n');
            navigator.clipboard.writeText(text).then(() => toastr.success('日志已复制到剪贴板')).catch(err => toastr.error('复制失败: ' + err));
        };

        // Download button
        overlay.querySelector('#btn-download-logs').onclick = () => {
            const filterLevel = overlay.querySelector('input[name="log-level"]:checked').value;
            const logs = logger.getLogs(filterLevel === 'all' ? null : filterLevel);
            const text = logs.map(log => {
                let line = `[${log.timestamp}] [${log.level.toUpperCase()}] ${log.message}`;
                if (log.args && log.args.length > 0) {
                    line += ' ' + log.args.map(arg =>
                        typeof arg === 'object' ? JSON.stringify(arg) : String(arg)
                    ).join(' ');
                }
                return line;
            }).join('\n');
            const blob = new Blob([text], { type: 'text/plain' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `worldbook_logs_${new Date().toISOString().slice(0,19).replace(/[:T]/g, '_')}.txt`;
            a.click();
            URL.revokeObjectURL(url);
            toastr.success('日志已下载');
        };

        // Close handlers
        overlay.querySelector('.wb-close-modal').onclick = () => overlay.remove();
        overlay.onclick = (e) => { if (e.target === overlay) overlay.remove(); };
    },

    openWarningListModal() {
        const warnings = this.getWarningList();
        if (warnings.length === 0) return toastr.info("没有警告条目");

        const escapeHtml = (str) => (str || '').replace(/[&<>"']/g, m => ({
            '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;'
        })[m]);

        const overlay = document.createElement('div');
        overlay.className = 'wb-sort-modal-overlay';

        let listHtml = '';
        warnings.forEach(entry => {
            listHtml += `
            <div class="wb-warning-item" data-uid="${entry.uid}">
                <div class="wb-warning-item-info">
                    <div class="wb-warning-icon">
                        <i class="fa-solid fa-circle-exclamation"></i>
                    </div>
                    <div class="wb-warning-item-content">
                        <span class="wb-warning-item-name">${escapeHtml(entry.comment) || '未命名条目'}</span>
                    </div>
                </div>
                <button class="wb-warning-item-action" data-edit="${entry.uid}" type="button" title="编辑">
                    <i class="fa-solid fa-pen-to-square"></i>
                </button>
            </div>`;
        });

        overlay.innerHTML = `
            <div class="wb-warning-modal">
                <div class="wb-warning-header">
                    <div class="wb-warning-title">
                        <i class="fa-solid fa-triangle-exclamation"></i>
                        <span>关键词缺失警告</span>
                        <span class="wb-warning-badge">${warnings.length}</span>
                    </div>
                    <button class="wb-warning-close" aria-label="关闭">
                        <i class="fa-solid fa-xmark"></i>
                    </button>
                </div>
                <div class="wb-warning-body">
                    <div class="wb-warning-alert">
                        <i class="fa-solid fa-circle-info"></i>
                        <span>这些绿灯条目已启用但未设置关键词，在聊天中永远不会被触发。您可以点击编辑按钮为每个条目添加关键词。</span>
                    </div>
                    <div class="wb-warning-list">
                        ${listHtml}
                    </div>
                </div>
            </div>
        `;

        document.body.appendChild(overlay);
        this.setupModalPositioning(overlay.querySelector('.wb-warning-modal'), overlay);

        overlay.addEventListener('click', (e) => {
            if (e.target.dataset.edit) {
                const entry = STATE.entries.find(en => en.uid === Number(e.target.dataset.edit));
                if (entry) {
                    this.openContentPopup(entry);
                    overlay.remove();
                }
            }
        });

        overlay.querySelector('.wb-warning-close').onclick = () => overlay.remove();
        overlay.onclick = (e) => { if (e.target === overlay) overlay.remove(); };
    },

    openExcludeRecursionModal() {
        const badEntries = this.getExcludeRecursionWarningList();
        if (badEntries.length === 0) return toastr.info("没有需要修复的条目");

        const escapeHtml = (str) => (str || '').replace(/[&<>"']/g, m => ({
            '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;'
        })[m]);

        const overlay = document.createElement('div');
        overlay.className = 'wb-sort-modal-overlay';

        let listHtml = '';
        badEntries.forEach(entry => {
            listHtml += `
            <div class="wb-warning-item" data-uid="${entry.uid}">
                <div class="wb-warning-item-info">
                    <div class="wb-warning-icon">
                        <i class="fa-solid fa-triangle-exclamation"></i>
                    </div>
                    <div class="wb-warning-item-content">
                        <span class="wb-warning-item-name">${escapeHtml(entry.comment) || '未命名条目'}</span>
                    </div>
                </div>
            </div>`;
        });

        overlay.innerHTML = `
            <div class="wb-warning-modal wb-warning-modal-warm">
                <div class="wb-warning-header">
                    <div class="wb-warning-title">
                        <i class="fa-solid fa-triangle-exclamation"></i>
                        <span>不可递归警告</span>
                        <span class="wb-warning-badge">${badEntries.length}</span>
                    </div>
                    <button class="wb-warning-close" aria-label="关闭">
                        <i class="fa-solid fa-xmark"></i>
                    </button>
                </div>
                <div class="wb-warning-body">
                    <div class="wb-warning-alert">
                        <i class="fa-solid fa-shield-halved"></i>
                        <span>以下这些绿灯条目未启用「不可递归」，可能会被世界书的其他条目始终激活触发。</span>
                    </div>
                    <div class="wb-warning-list">
                        ${listHtml}
                    </div>
                </div>
                <div class="wb-warning-footer">
                    <button class="wb-warning-btn-primary" id="btn-ers-fix-all">
                        <i class="fa-solid fa-check"></i>
                        <span>一键启用不可递归</span>
                    </button>
                </div>
            </div>
        `;

        document.body.appendChild(overlay);
        this.setupModalPositioning(overlay.querySelector('.wb-warning-modal'), overlay);

        const fixAllBtn = overlay.querySelector('#btn-ers-fix-all');
        fixAllBtn.onclick = async () => {
            if (!confirm(`确定要将 ${badEntries.length} 个条目的「不可递归」设为 True 吗？`)) return;
            try {
                await Actions.batchSetExcludeRecursion(badEntries.map(e => e.uid));
                toastr.success('已批量更新');
                overlay.remove();
                this.renderGlobalStats();
                badEntries.forEach(e => this.updateCardStatus(e.uid));
            } catch (e) {
                toastr.error('批量更新失败: ' + e.message);
            }
        };

        overlay.querySelector('.wb-warning-close').onclick = () => overlay.remove();
        overlay.onclick = (e) => { if (e.target === overlay) overlay.remove(); };
    },

    async openContextPreviewModal() {
        if (!STATE.currentBookName) return toastr.warning("请先选择一本世界书");

        const loadingToast = toastr.info("正在分析上下文...", "请稍候", { timeOut: 0, extendedTimeOut: 0 });

        try {
            const context = getContext();
            const charId = context.characterId;
            const charData = context.characters[charId] || {};
            let fullText = (charData.description || '') + '\n' + (charData.persona || '') + '\n';
            const chat = context.chat || [];
            const recentChat = chat.slice(-30);
            fullText += recentChat.map(c => (c.name || '') + ': ' + (c.mes || '')).join('\n');
            const searchContext = fullText.toLowerCase();

            let activatedEntries = STATE.entries.filter(entry => {
                if (entry.disable) return false;
                if (entry.constant) return true;
                if (!entry.key || entry.key.length === 0) return false;
                return entry.key.some(k => {
                    const keyStr = String(k).trim();
                    if (!keyStr) return false;
                    if (keyStr.startsWith('/') && keyStr.endsWith('/') && keyStr.length > 2) {
                        try {
                            const regexBody = keyStr.substring(1, keyStr.lastIndexOf('/'));
                            const flags = keyStr.substring(keyStr.lastIndexOf('/') + 1) + 'i';
                            const regex = new RegExp(regexBody, flags);
                            return regex.test(fullText);
                        } catch (e) { return false; }
                    } else {
                        return searchContext.includes(keyStr.toLowerCase());
                    }
                });
            });

            toastr.clear(loadingToast);

            activatedEntries.sort((a, b) => {
                const scoreA = Actions.getEntrySortScore(a);
                const scoreB = Actions.getEntrySortScore(b);
                if (scoreA !== scoreB) return scoreB - scoreA; 
                const orderA = a.order ?? 0;
                const orderB = b.order ?? 0;
                return (orderA - orderB) || (a.uid - b.uid);
            });

            let sidebarHtml = '';
            let contentHtml = '';
            const originalContentMap = new Map();

            const posMapping = {
                0: '角色定义之前', 1: '角色定义之后', 2: 'AN 之前', 3: 'AN 之后',
                4: '@D', 5: '示例消息之前', 6: '示例消息之后'
            };

            if (activatedEntries.length === 0) {
                sidebarHtml = `<div style="padding:20px 15px;color:#9ca3af;text-align:center;font-size:0.9em;">无激活条目</div>`;
                contentHtml = `
                    <div style="display:flex;height:100%;align-items:center;justify-content:center;color:#9ca3af;flex-direction:column">
                        <i class="fa-solid fa-ghost" style="font-size:3em;margin-bottom:15px;opacity:0.5"></i>
                        <div>当前上下文未激活任何条目</div>
                    </div>`;
            } else {
                activatedEntries.forEach((entry, idx) => {
                    const title = entry.comment || (entry.key && entry.key.length ? entry.key[0] : `Entry #${entry.uid}`);
                    const isConstant = !!entry.constant;
                    const itemTypeClass = isConstant ? 'type-blue' : 'type-green';
                    const barColorClass = isConstant ? 'wb-bar-blue' : 'wb-bar-green';

                    let posVal = typeof entry.position === 'number' ? entry.position : 1;
                    let posText = posMapping[posVal] || '未知位置';
                    if (posVal === 4) posText = `@D ${entry.depth ?? 4}`;

                    const typeLabel = isConstant ? '蓝灯' : '绿灯';
                    const orderVal = entry.order ?? 0;
                    const tooltipText = `${typeLabel} ${posText} ${orderVal}`;
                    const colorMode = isConstant ? 'blue' : 'green';

                    const rawContent = (entry.content || '').replace(/</g, '&lt;');
                    originalContentMap.set(`ctx-block-${idx}`, { title, content: rawContent });

                    sidebarHtml += `
                        <div class="wb-ctx-sidebar-item ${itemTypeClass}" data-target="ctx-block-${idx}" id="sidebar-item-${idx}" title="${tooltipText}" data-color-mode="${colorMode}">
                            <div class="wb-ctx-bar ${barColorClass}"></div>
                            <div class="wb-ctx-info"><span class="wb-ctx-name">${title}</span></div>
                        </div>`;

                    contentHtml += `
                        <div id="ctx-block-${idx}" class="wb-ctx-block" data-idx="${idx}">
                            <div class="wb-ctx-block-title">
                                <span class="title-text">${title}</span>
                                <span style="font-size:0.8em; font-weight:normal; color:#9ca3af; margin-left:auto; font-family: 'Segoe UI', sans-serif;">${posText}</span>
                            </div>
                            <div class="wb-ctx-block-content">${rawContent}</div>
                        </div>`;
                });
            }

            const overlay = document.createElement('div');
            overlay.className = 'wb-sort-modal-overlay';
            overlay.style.zIndex = '22000';

            const isSidebarCollapsed = localStorage.getItem('wb_ctx_sidebar_collapsed') === 'true';

            overlay.innerHTML = `
                <div class="wb-sort-modal" style="width:1000px; height:85vh; max-width:95vw; border-radius:12px; overflow:hidden; display:flex; flex-direction:column; box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);">
                    <div class="wb-sort-header" style="padding:10px 20px; height:60px;">
                        <span style="font-size:1.1em; font-weight:bold; color:#111827; display:flex; align-items:center; gap:15px;">
                            <i class="fa-solid fa-align-left" id="wb-ctx-toggle-sidebar" style="cursor:pointer; color:#6b7280; transition:0.2s" title="切换侧边栏"></i>
                            <span class="wb-ctx-header-title-text">实际上下文预览</span>
                        </span>
                        <div style="display:flex; align-items:center;">
                            <div class="wb-ctx-search-container">
                                <i class="fa-solid fa-magnifying-glass" style="color:#9ca3af; font-size:0.9em;"></i>
                                <input type="text" class="wb-ctx-search-input" placeholder="检索关键词...">
                                <div class="wb-ctx-nav-controls">
                                    <div class="wb-ctx-nav-btn" id="wb-search-up"><i class="fa-solid fa-arrow-up"></i></div>
                                    <div class="wb-ctx-nav-btn" id="wb-search-down"><i class="fa-solid fa-arrow-down"></i></div>
                                    <div class="wb-ctx-nav-info">0/0</div>
                                </div>
                            </div>
                            <i class="fa-solid fa-heading" id="wb-ctx-toggle-clean" style="cursor:pointer; color:#9ca3af; font-size:1.2em; padding:5px; margin-left:10px;" title="切换纯净模式 (仅显示内容)"></i>
                            <div class="wb-close-btn" style="cursor:pointer; color:#9ca3af; font-size:1.2em; padding:5px; margin-left:10px;"><i class="fa-solid fa-xmark"></i></div>
                        </div>
                    </div>
                    <div class="wb-ctx-layout-container">
                        <div class="wb-ctx-sidebar-panel ${isSidebarCollapsed ? 'collapsed' : ''}" id="wb-ctx-sidebar">${sidebarHtml}</div>
                        <div class="wb-ctx-viewer-panel" id="wb-ctx-viewer">${contentHtml}</div>
                    </div>
                </div>`;

            document.body.appendChild(overlay);
            this.setupModalPositioning(overlay.querySelector('.wb-sort-modal'), overlay);

            const sidebar = overlay.querySelector('#wb-ctx-sidebar');
            const viewer = overlay.querySelector('#wb-ctx-viewer');
            const sidebarItems = Array.from(sidebar.querySelectorAll('.wb-ctx-sidebar-item'));
            const blocks = Array.from(viewer.querySelectorAll('.wb-ctx-block'));
            const toggleBtn = overlay.querySelector('#wb-ctx-toggle-sidebar');
            const searchInput = overlay.querySelector('.wb-ctx-search-input');
            const navControls = overlay.querySelector('.wb-ctx-nav-controls');
            const navInfo = overlay.querySelector('.wb-ctx-nav-info');
            const btnUp = overlay.querySelector('#wb-search-up');
            const btnDown = overlay.querySelector('#wb-search-down');
            const cleanBtn = overlay.querySelector('#wb-ctx-toggle-clean');

            cleanBtn.onclick = () => {
                viewer.classList.toggle('wb-clean-mode');
                const isClean = viewer.classList.contains('wb-clean-mode');
                cleanBtn.style.color = isClean ? '#3b82f6' : '#9ca3af';
            };

            toggleBtn.onclick = () => {
                sidebar.classList.toggle('collapsed');
                const isCollapsed = sidebar.classList.contains('collapsed');
                toggleBtn.style.color = isCollapsed ? '#d1d5db' : '#6b7280';
                localStorage.setItem('wb_ctx_sidebar_collapsed', isCollapsed);
            };

            const scrollToBlock = (targetId) => {
                const targetEl = viewer.querySelector(`#${targetId}`);
                if (targetEl) {
                    const topPos = targetEl.offsetTop - 20;
                    viewer.scrollTo({ top: topPos, behavior: 'smooth' });
                }
            };

            sidebarItems.forEach(item => {
                item.onclick = () => {
                    sidebarItems.forEach(i => i.classList.remove('active'));
                    item.classList.add('active');
                    scrollToBlock(item.dataset.target);
                };
            });

            let scrollTimeout;
            viewer.addEventListener('scroll', () => {
                if (scrollTimeout) clearTimeout(scrollTimeout);
                scrollTimeout = setTimeout(() => {
                    const viewerTop = viewer.scrollTop;
                    let activeId = null;
                    const visibleBlocks = blocks.filter(b => b.style.display !== 'none');

                    for (let block of visibleBlocks) {
                        if (block.offsetTop <= viewerTop + 100) { activeId = block.id; }
                        else {
                            if (!activeId) activeId = block.id;
                            break;
                        }
                    }

                    if (activeId) {
                        sidebarItems.forEach(i => {
                            if (i.dataset.target === activeId) i.classList.add('active');
                            else i.classList.remove('active');
                        });
                        // 禁用侧边栏自动滚动，避免造成抖动
                    }
                }, 100);
            });
            viewer.dispatchEvent(new Event('scroll'));

            let searchDebounce;
            let currentMatches = []; 
            let currentMatchIndex = -1;

            const updateNavInfo = () => {
                if (currentMatches.length > 0) {
                    navControls.classList.add('show');
                    navInfo.textContent = `${currentMatchIndex + 1}/${currentMatches.length}`;
                } else {
                    navControls.classList.remove('show');
                    navInfo.textContent = "0/0";
                }
            };

            const jumpToMatch = (index) => {
                if (index < 0 || index >= currentMatches.length) return;
                currentMatchIndex = index;
                currentMatches.forEach(el => el.classList.remove('active'));
                const target = currentMatches[index];
                target.classList.add('active');

                // 修复二：弃用引发抖动的 scrollIntoView，改用安全的相对定位滚动
                let offsetTop = 0;
                let el = target;
                while(el && el !== viewer) {
                    offsetTop += el.offsetTop;
                    el = el.offsetParent;
                }
                const topPos = offsetTop - viewer.offsetHeight / 2 + target.offsetHeight / 2;
                viewer.scrollTo({ top: Math.max(0, topPos), behavior: 'smooth' });

                updateNavInfo();
            };

            searchInput.addEventListener('input', (e) => {
                const term = e.target.value.trim();
                if (searchDebounce) clearTimeout(searchDebounce);
                searchDebounce = setTimeout(() => {
                    currentMatches = [];
                    currentMatchIndex = -1;

                    if (!term) {
                        blocks.forEach(block => {
                            const data = originalContentMap.get(block.id);
                            if (data) {
                                block.querySelector('.wb-ctx-block-content').innerHTML = data.content;
                                block.querySelector('.title-text').innerHTML = data.title;
                            }
                            block.classList.remove('filtered-out');
                        });
                        sidebarItems.forEach(item => item.classList.remove('filtered-out'));
                        navControls.classList.remove('show');
                        viewer.dispatchEvent(new Event('scroll'));
                        return;
                    }

                    const regex = new RegExp(`(${term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');

                    blocks.forEach((block, i) => {
                        const data = originalContentMap.get(block.id);
                        if (!data) return;
                        const titleMatch = regex.test(data.title);
                        const contentMatch = regex.test(data.content);
                        const hasMatch = titleMatch || contentMatch;

                        if (hasMatch) {
                            block.classList.remove('filtered-out');
                            sidebarItems[i].classList.remove('filtered-out');
                            if (contentMatch) block.querySelector('.wb-ctx-block-content').innerHTML = data.content.replace(regex, '<span class="wb-search-highlight">$1</span>');
                            else block.querySelector('.wb-ctx-block-content').innerHTML = data.content;

                            if (titleMatch) block.querySelector('.title-text').innerHTML = data.title.replace(regex, '<span class="wb-search-highlight">$1</span>');
                            else block.querySelector('.title-text').innerHTML = data.title;
                        } else {
                            block.classList.add('filtered-out');
                            sidebarItems[i].classList.add('filtered-out');
                        }
                    });

                    currentMatches = Array.from(viewer.querySelectorAll('.wb-search-highlight'));
                    if (currentMatches.length > 0) jumpToMatch(0);
                    else updateNavInfo();

                }, 300); 
            });

            btnUp.onclick = () => {
                let next = currentMatchIndex - 1;
                if (next < 0) next = currentMatches.length - 1; 
                jumpToMatch(next);
            };
            btnDown.onclick = () => {
                let next = currentMatchIndex + 1;
                if (next >= currentMatches.length) next = 0; 
                jumpToMatch(next);
            };

            const close = () => overlay.remove();
            overlay.querySelector('.wb-close-btn').onclick = close;
            overlay.onclick = (e) => { if (e.target === overlay) close(); };

        } catch (e) {
            toastr.clear(loadingToast);
            logger.error("计算上下文失败:", e);
            toastr.error("计算上下文失败: " + e.message);
        }
    },

    showDeleteWbConfirmModal(bookName, onConfirm, onDisable) {
        const overlay = document.createElement('div');
        overlay.className = 'wb-sort-modal-overlay';
        overlay.style.zIndex = '25000';
        overlay.innerHTML = `
            <div class="wb-sort-modal" id="wb-del-confirm-box" style="width:400px; height:auto; border-radius:12px; overflow:hidden;">
                <div style="padding:20px; text-align:center;">
                    <div style="font-size:3em; color:#f59e0b; margin-bottom:10px;"><i class="fa-solid fa-triangle-exclamation"></i></div>
                    <h3 style="margin:0 0 10px 0; color:#1f2937;">关联删除</h3>
                    <p style="color:#4b5563;">是否同时删除角色绑定的主要世界书<br><strong>${bookName}</strong>?</p>
                    <div style="margin-top:15px; border-top:1px solid #f3f4f6; padding-top:10px;">
                         <button class="wb-btn-modal btn-disable" style="color:#9ca3af; background:none; border:none; cursor:pointer; text-decoration:underline; font-size:0.9em;">禁用该功能</button>
                    </div>
                </div>
                <div style="background:#f9fafb; padding:15px; display:flex; justify-content:center; gap:20px; border-top:1px solid #e5e7eb;">
                    <button class="wb-btn-modal btn-cancel" style="padding:8px 25px; border-radius:6px; border:1px solid #000; background:#000; color:#fff; cursor:pointer;">取消</button>
                    <button class="wb-btn-modal btn-confirm" style="padding:8px 25px; border-radius:6px; border:none; background:#ef4444; color:#fff; cursor:pointer;">删除</button>
                </div>
            </div>`;
        document.body.appendChild(overlay);
        this.setupModalPositioning(overlay.querySelector('.wb-sort-modal'), overlay);
        overlay.querySelector('.btn-cancel').onclick = () => overlay.remove();
        overlay.querySelector('.btn-confirm').onclick = () => {
            onConfirm();
            toastr.success(`已删除世界书：${bookName}`);
            overlay.remove();
        };
        overlay.querySelector('.btn-disable').onclick = () => { onDisable(); overlay.remove(); };
    },

    showBindingDetailsModal(bookName, bindInfo) {
        // HTML转义辅助函数
        const escapeHtml = (str) => (str || '').replace(/[&<>"']/g, m => ({
            '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;'
        })[m]);

        // 创建遮罩层
        const overlay = document.createElement('div');
        overlay.className = 'wb-sort-modal-overlay';

        // 创建弹窗
        const modal = document.createElement('div');
        modal.className = 'wb-sort-modal';
        modal.style.cssText = 'width:380px;max-width:90vw;max-height:calc(100vh - 40px);overflow-y:auto;border-radius:16px;box-shadow:0 20px 25px -5px rgba(0,0,0,0.1);position:relative;';

        // 构建内容
        let html = '';

        // 弹窗头部：标题和右上角关闭按钮
        html += `<div style="padding:16px 20px;border-bottom:1px solid #e5e7eb;display:flex;justify-content:space-between;align-items:center;">`;
        html += `<h3 style="margin:0;color:#000;font-size:18px;font-weight:700;">${escapeHtml(bookName)}</h3>`;
        html += `<button class="wb-binding-close-btn" style="background:none;border:none;cursor:pointer;padding:4px 8px;color:#6b7280;font-size:18px;line-height:1;" title="关闭"><i class="fa-solid fa-xmark"></i></button>`;
        html += `</div>`;

        html += `<div style="padding:20px;display:flex;flex-direction:column;gap:16px;font-size:14px;">`;

        // 全局绑定
        if (bindInfo.global) {
            html += `<div><div style="color:#000;font-weight:700;">绑定到全局</div></div>`;
        }

        // 主要世界书绑定
        if (bindInfo.primary && bindInfo.primary.length > 0) {
            const primaryText = bindInfo.primary.map(p => escapeHtml(p)).join(' 和 ');
            html += `<div><div style="color:#000;font-weight:700;margin-bottom:4px;">主要</div><div style="color:#9ca3af;">${primaryText}</div></div>`;
        }

        // 附加世界书绑定
        if (bindInfo.aux && bindInfo.aux.length > 0) {
            const auxText = bindInfo.aux.map(a => escapeHtml(a)).join(' 和 ');
            html += `<div><div style="color:#000;font-weight:700;margin-bottom:4px;">附加</div><div style="color:#9ca3af;">${auxText}</div></div>`;
        }

        // 聊天世界书绑定
        if (bindInfo.chat && bindInfo.chat.length > 0) {
            const chatMap = {};
            bindInfo.chat.forEach(item => {
                if (typeof item === 'string') {
                    chatMap[''] = chatMap[''] || [];
                    chatMap[''].push(item);
                } else {
                    const key = item.charName || '';
                    if (!chatMap[key]) chatMap[key] = [];
                    chatMap[key].push(item.chatName);
                }
            });

            const chatSections = [];
            Object.entries(chatMap).forEach(([charName, chats]) => {
                const uniqueChats = [...new Set(chats)];
                if (charName) {
                    const chatDisplay = uniqueChats.map(c => escapeHtml(c)).join(' 和 ');
                    chatSections.push(`<div style="color:#9ca3af;">${escapeHtml(charName)} <i style="font-style:italic;">的</i> ${chatDisplay}</div>`);
                } else {
                    uniqueChats.forEach(chat => {
                        chatSections.push(`<div style="color:#9ca3af;">${escapeHtml(chat)}</div>`);
                    });
                }
            });

            html += `<div><div style="color:#000;font-weight:700;margin-bottom:4px;">聊天</div>${chatSections.join('')}</div>`;
        }

        html += `</div>`;

        modal.innerHTML = html;
        overlay.appendChild(modal);
        document.body.appendChild(overlay);

        // 定位弹窗
        this.setupModalPositioning(modal, overlay);

        // 深色主题支持
        const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        if (isDark) {
            modal.style.background = '#1f2937';
            modal.style.color = '#d1d5db';
            modal.style.border = '1px solid #374151';
            modal.querySelector('h3').style.color = '#f3f4f6';
            modal.children[0].style.borderBottomColor = '#374151';
            modal.querySelector('.wb-binding-close-btn').style.color = '#9ca3af';
        }

        // 关闭事件
        const close = () => overlay.remove();
        modal.querySelector('.wb-binding-close-btn').onclick = close;
        overlay.onclick = (e) => { if(e.target === overlay) close(); };
    },

    showCharBindingDetailsModal(charName, bindData) {
        const overlay = document.createElement('div');
        overlay.className = 'wb-sort-modal-overlay';
        overlay.style.zIndex = '25000';

        let html = `<div class="wb-sort-modal" style="width:400px; border-radius:12px; overflow:hidden;">
            <div style="background:var(--SmartThemeBlurTintColor, #1f2937); padding:15px; text-align:center;">
                <h3 style="margin:0; color:var(--SmartThemeQuoteColor, #70a1ff); font-size:18px;">${charName}</h3>
                <div style="font-size:12px; color:gray; margin-top:4px;">角色绑定详情</div>
            </div>
            <div style="padding:20px; text-align:center; display:flex; flex-direction:column; gap:16px; font-size:14px; color:var(--SmartThemeBodyColor, #d1d5db);">
        `;

        if (bindData.primary) html += `<div><strong style="color:var(--SmartThemeQuoteColor, #70a1ff);">👤 已绑定主要世界书：</strong><br><span style="color:#9ca3af">${bindData.primary}</span></div>`;
        else html += `<div><strong style="color:var(--SmartThemeQuoteColor, #70a1ff);">👤 已绑定主要世界书：</strong><br><span style="color:#9ca3af">无</span></div>`;
        
        if (bindData.aux.length > 0) html += `<div><strong style="color:var(--SmartThemeQuoteColor, #70a1ff);">🔗 已绑定附加世界书：</strong><br><span style="color:#9ca3af">${bindData.aux.join(', ')}</span></div>`;
        else html += `<div><strong style="color:var(--SmartThemeQuoteColor, #70a1ff);">🔗 已绑定附加世界书：</strong><br><span style="color:#9ca3af">无</span></div>`;

        if (bindData.chats.length > 0) {
            const chatStrs = bindData.chats.map(c => `${c.chatName} 已绑定 ${c.wbName}`);
            html += `<div><strong style="color:var(--SmartThemeQuoteColor, #70a1ff);">💬 聊天绑定情况：</strong><br><span style="color:#9ca3af">${chatStrs.join('<br>')}</span></div>`;
        } else html += `<div><strong style="color:var(--SmartThemeQuoteColor, #70a1ff);">💬 聊天绑定情况：</strong><br><span style="color:#9ca3af">无</span></div>`;

        html += `
            </div>
            <div style="background:var(--SmartThemeBlurTintColor, #f9fafb); padding:15px; display:flex; justify-content:center; border-top:1px solid var(--SmartThemeBorderColor, #e5e7eb);">
                <button class="wb-btn-modal btn-cancel" style="padding:8px 35px; border-radius:6px; background:var(--SmartThemeBotMesColor, #374151); color:var(--SmartThemeBodyColor, #fff); border:1px solid var(--SmartThemeBorderColor, gray); cursor:pointer;">关闭详情</button>
            </div>
        </div>`;

        overlay.innerHTML = html;
        document.body.appendChild(overlay);
        this.setupModalPositioning(overlay.querySelector('.wb-sort-modal'), overlay);

        const close = () => overlay.remove();
        overlay.querySelector('.btn-cancel').onclick = close;
        overlay.onclick = (e) => { if(e.target === overlay) close(); };
    },
    
    openCustomGroupModal() {
        const overlay = document.createElement('div');
        overlay.className = 'wb-sort-modal-overlay';
        overlay.style.zIndex = '25000';

        let searchTerm = '';
        let filterPrimary = true; 
        const tagColors = ['#bae6fd', '#bbf7d0', '#fbcfe8', '#fef08a', '#e9d5ff'];

        overlay.innerHTML = `
            <div class="wb-sort-modal" id="wb-custom-group-modal" style="width:600px; max-height:85vh; border-radius:12px; overflow:hidden; display:flex; flex-direction:column; background:var(--SmartThemeBlurTintColor, #1f2937);">
                <div style="padding:15px 20px;">
                    <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:15px;">
                        <h3 style="margin:0; font-size:1.1em; color:#111827;">自定义分组管理</h3>
                        <div style="cursor:pointer; color:#9ca3af;" class="btn-close-custom"><i class="fa-solid fa-xmark"></i></div>
                    </div>
                    <div style="display:flex; flex-direction:column; gap:10px;">
                        <div style="position:relative;">
                            <input type="text" id="custom-search-input" placeholder="🔍 搜索世界书..." style="width:100%; padding:8px 15px; padding-left:15px; border-radius:15px; border:1px solid var(--SmartThemeBorderColor, gray); background:rgba(0,0,0,0.2); color:inherit;">
                        </div>
                        <label style="display:flex; align-items:center; gap:8px; cursor:pointer; color:var(--SmartThemeBodyColor, #d1d5db); font-size:0.9em;">
                            <input type="checkbox" id="custom-filter-primary" ${filterPrimary ? 'checked' : ''}> 过滤掉主要世界书
                        </label>
                    </div>
                </div>
                <div id="custom-list-container" style="flex:1; overflow-y:auto; padding:15px; display:flex; flex-direction:column; gap:8px;"></div>
            </div>
        `;

        document.body.appendChild(overlay);
        this.setupModalPositioning(overlay.querySelector('.wb-sort-modal'), overlay);

        const listContainer = overlay.querySelector('#custom-list-container');
        const searchInput = overlay.querySelector('#custom-search-input');
        const filterCheck = overlay.querySelector('#custom-filter-primary');

        const renderList = () => {
            listContainer.innerHTML = '';

            let items = STATE.allBookNames.map(name => {
                const meta = STATE.metadata[name] || {};
                const tags = meta.tags || [];
                const bindInfo = STATE.boundBooksSet[name] || { primary: [] };
                const isPrimary = bindInfo.primary.length > 0;
                return { name, tags, isPrimary };
            });

            items = items.filter(item => {
                if (searchTerm && !item.name.toLowerCase().includes(searchTerm.toLowerCase())) return false;
                if (filterPrimary) { if (item.isPrimary) return false; }
                return true;
            });

            items.sort((a, b) => {
                if (a.tags.length !== b.tags.length) return b.tags.length - a.tags.length;
                const tagsStrA = a.tags.join('');
                const tagsStrB = b.tags.join('');
                if (tagsStrA !== tagsStrB) return tagsStrA.localeCompare(tagsStrB);
                return a.name.localeCompare(b.name);
            });

            if (items.length === 0) {
                listContainer.innerHTML = '<div style="text-align:center; color:#9ca3af; padding:20px;">没有符合条件的世界书</div>';
                return;
            }

            items.forEach(item => {
                let tagsHtml = '';
                item.tags.forEach((tag, idx) => {
                    const color = tagColors[idx % 5];
                    tagsHtml += `<span style="background:${color}; color:#1f2937; font-size:0.8em; padding:3px 8px; border-radius:12px; margin-right:6px; display:inline-block; cursor:pointer; font-weight:600;" class="wb-tag-badge" data-book="${item.name}" data-idx="${idx}" title="点击删除该标签">${tag}</span>`;
                });

                const cardHtml = `
                    <div class="wb-manage-card" style="display:flex; justify-content:space-between; align-items:center; padding:12px 15px; margin:0; border:1px solid var(--SmartThemeBorderColor, rgba(255,255,255,0.1)); border-radius:8px; background:rgba(0,0,0,0.15);">
                        <div style="font-weight:bold; color:var(--SmartThemeBodyColor, #e5e7eb); flex:1; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; font-size:1.05em;" title="${item.name}">
                            ${item.name}
                            ${item.isPrimary ? '<span style="font-size:0.7em; color:#ef4444; margin-left:5px; font-weight:normal;">[主要]</span>' : ''}
                        </div>
                        <div style="display:flex; align-items:center;">
                            <div style="margin-right:10px; display:flex; align-items:center;">${tagsHtml}</div>
                            <div class="wb-icon-action btn-add-tag" data-book="${item.name}" title="添加标签" style="color:var(--SmartThemeQuoteColor, #7c5cbd);"><i class="fa-solid fa-pen-to-square"></i></div>
                        </div>
                    </div>
                `;
                listContainer.insertAdjacentHTML('beforeend', cardHtml);
            });

            listContainer.querySelectorAll('.btn-add-tag').forEach(btn => {
                btn.onclick = async () => {
                    const bookName = btn.dataset.book;
                    const meta = STATE.metadata[bookName] || {};
                    const currentTags = meta.tags || [];

                    if (currentTags.length >= 5) return toastr.warning("每个世界书最多只能添加 5 个标签！");
                    const newTag = prompt(`为 ${bookName} 添加新标签:\n(当前已添加 ${currentTags.length}/5 个)`);
                    if (newTag && newTag.trim()) {
                        const trimmedTag = newTag.trim();
                        if (currentTags.includes(trimmedTag)) return toastr.warning("该标签已存在！");
                        await Actions.addCustomTag(bookName, trimmedTag);
                        renderList(); 
                    }
                };
            });

            listContainer.querySelectorAll('.wb-tag-badge').forEach(badge => {
                badge.onclick = async () => {
                    const bookName = badge.dataset.book;
                    const index = parseInt(badge.dataset.idx);
                    if (confirm(`确定要移除标签 "${badge.innerText}" 吗？`)) {
                        await Actions.removeCustomTag(bookName, index);
                        renderList(); 
                    }
                };
            });
        };

        searchInput.oninput = (e) => { searchTerm = e.target.value; renderList(); };
        filterCheck.onchange = (e) => { filterPrimary = e.target.checked; renderList(); };
        renderList();

        const close = () => {
            overlay.remove();
            if (STATE.manageTab === 'custom') this.renderManageView(document.getElementById('wb-manage-search').value);
        };
        overlay.querySelector('.btn-close-custom').onclick = close;
        overlay.onclick = (e) => { if(e.target === overlay) close(); };
    },

    initTooltips() {
        if (this._tooltipInited) return;
        this._tooltipInited = true;
        const tipEl = document.createElement('div');
        tipEl.className = 'wb-tooltip';
        document.body.appendChild(tipEl);

        const show = (text, x, y, colorMode) => {
            tipEl.textContent = text;
            tipEl.classList.remove('blue', 'green');
            if (colorMode) tipEl.classList.add(colorMode);
            tipEl.classList.add('show');

            const rect = tipEl.getBoundingClientRect();
            let left = x + 15;
            let top = y + 15;
            if (left + rect.width > window.innerWidth) left = x - rect.width - 5;
            if (top + rect.height > window.innerHeight) top = y - rect.height - 5;

            tipEl.style.left = left + 'px';
            tipEl.style.top = top + 'px';
        };
        const hide = () => { tipEl.classList.remove('show', 'blue', 'green'); };

        let isTouchInteraction = false;
        document.body.addEventListener('mouseover', (e) => {
            if (isTouchInteraction) return; 
            const container = e.target.closest(`#${CONFIG.id}, .wb-modal-overlay, .wb-sort-modal-overlay`);
            if (!container) return;

            const target = e.target.closest('[title], [data-wb-tooltip]');
            if (target) {
                const text = target.getAttribute('title') || target.getAttribute('data-wb-tooltip');
                const colorMode = target.dataset.colorMode;
                if (target.getAttribute('title')) { target.setAttribute('data-wb-tooltip', text); target.removeAttribute('title'); }
                if (text) show(text, e.clientX, e.clientY, colorMode);
            }
        });
        document.body.addEventListener('mouseout', hide);

        let touchTimer = null;
        document.body.addEventListener('touchstart', (e) => {
            isTouchInteraction = true;
            hide();
            const container = e.target.closest(`#${CONFIG.id}, .wb-modal-overlay, .wb-sort-modal-overlay`);
            if (!container) return;

            const target = e.target.closest('[title], [data-wb-tooltip]');
            if (!target) return;

            const text = target.getAttribute('title') || target.getAttribute('data-wb-tooltip');
            const colorMode = target.dataset.colorMode;
            if (target.getAttribute('title')) { target.setAttribute('data-wb-tooltip', text); target.removeAttribute('title'); }

            if (text) {
                touchTimer = setTimeout(() => {
                    const touch = e.touches[0];
                    show(text, touch.clientX, touch.clientY, colorMode);
                }, 500);
            }
        }, { passive: true });

        const cancelTouch = () => {
            if (touchTimer) { clearTimeout(touchTimer); touchTimer = null; }
            setTimeout(() => { isTouchInteraction = false; }, 500);
            hide();
        };

        document.body.addEventListener('touchend', cancelTouch);
        document.body.addEventListener('touchmove', () => { if (touchTimer) clearTimeout(touchTimer); });
    },

    // 事件委托初始化：为动态生成的卡片统一监听事件
    initEntryListDelegation() {
        const list = document.getElementById('wb-entry-list');
        if (!list) return;

        // 监听输入框变化
        list.addEventListener('input', (e) => {
            const card = e.target.closest('.wb-card');
            if (!card) return;
            const uid = Number(card.dataset.uid);

            if (e.target.classList.contains('inp-name')) {
                Actions.updateEntry(uid, d => d.comment = e.target.value);
            } else if (e.target.classList.contains('inp-pos-depth')) {
                Actions.updateEntry(uid, d => d.depth = Number(e.target.value));
            } else if (e.target.classList.contains('inp-order')) {
                Actions.updateEntry(uid, d => d.order = Number(e.target.value));
            }
        });

        // 监听复选框和下拉框变化
        list.addEventListener('change', (e) => {
            const card = e.target.closest('.wb-card');
            if (!card) return;
            const uid = Number(card.dataset.uid);

            if (e.target.classList.contains('inp-enable')) {
                card.classList.toggle('disabled', !e.target.checked);
                Actions.updateEntry(uid, d => d.disable = !e.target.checked);
            } else if (e.target.classList.contains('inp-type')) {
                Actions.updateEntry(uid, d => {
                    d.constant = e.target.checked;
                    if (d.constant) d.selective = false; else d.selective = true;
                });
            } else if (e.target.classList.contains('inp-pos')) {
                const val = e.target.value;
                const depthInput = card.querySelector('.inp-pos-depth');
                const posWrapper = card.querySelector('.wb-pos-wrapper');
                if (depthInput) depthInput.style.display = val === 'at_depth' ? 'block' : 'none';
                if (posWrapper) posWrapper.classList.toggle('show-depth', val === 'at_depth');
                const intVal = WI_POSITION_MAP_REV[val] ?? 1;
                Actions.updateEntry(uid, d => d.position = intVal);
            }
        });

        // 监听点击事件
        list.addEventListener('click', (e) => {
            const card = e.target.closest('.wb-card');
            if (!card) return;
            const uid = Number(card.dataset.uid);

            if (e.target.classList.contains('btn-delete')) {
                if (confirm("确定要删除此条目吗？")) {
                    Actions.deleteEntry(uid);
                }
            } else if (e.target.classList.contains('btn-preview')) {
                const entry = STATE.entries.find(e => e.uid === uid);
                if (entry) this.openContentPopup(entry, e.target);
            }
        });
    }
};

export function adjustSortModalLayout() {
    const modal = document.querySelector('.wb-sort-modal');
    if (!modal) return;
    const vh = window.innerHeight;
    const maxLimit = vh * 0.85;

    modal.style.height = 'auto';
    modal.style.maxHeight = 'none';
    let targetHeight = modal.offsetHeight;

    if (targetHeight > maxLimit) {
        targetHeight = maxLimit;
        modal.style.height = targetHeight + 'px';
    } else {
        modal.style.height = 'auto';
    }

    modal.style.maxHeight = maxLimit + 'px';
    const topPosition = (vh - targetHeight) / 2;
    modal.style.top = `${Math.max(20, topPosition)}px`;
    modal.style.left = '50%';
    modal.style.transform = 'translateX(-50%)'; 
    modal.style.margin = '0'; 
}