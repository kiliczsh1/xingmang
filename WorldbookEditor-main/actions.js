// actions.js

import { getContext } from '../../../extensions.js';
import { event_types, eventSource } from '../../../../script.js';
import { getCharaFilename } from '../../../utils.js';
import { CONFIG, STATE, THEME_KEY } from './state.js';
import { API, setCharBindings, charSetAuxWorlds } from './api.js';
import { UI } from './ui.js';
import { logger } from './logger.js';

export const Actions = {
    async flushPendingSave() {
        if (STATE.debouncer) {
            clearTimeout(STATE.debouncer);
            STATE.debouncer = null;
            if (STATE.currentBookName && Array.isArray(STATE.entries)) {
                await API.saveBookEntries(STATE.currentBookName, STATE.entries);
            }
        }
    },

    async loadBoundBooksSetAsync() {
        return await API.getAllBoundBookNames();
    },

    getEntrySortScore(entry) {
        const context = getContext();
        const anDepth = (context.chatMetadata && context.chatMetadata['note_depth'])
            ?? (context.extensionSettings && context.extensionSettings.note && context.extensionSettings.note.defaultDepth)
            ?? 4;
        const pos = typeof entry.position === 'number' ? entry.position : 1;

        if (pos === 0) return 100000; 
        if (pos === 1) return 90000;  
        if (pos === 5) return 80000;  
        if (pos === 6) return 70000;  
        if (pos === 4) return entry.depth ?? 4;
        if (pos === 2) return anDepth + 0.6;
        if (pos === 3) return anDepth + 0.4;
        return -9999;
    },

    async init() {
        if (STATE.isInitialized) return;
        UI.initTooltips();
        this.registerCharDeleteListener();

        const es = eventSource;
        const et = event_types;

        es.on(et.SETTINGS_UPDATED, () => { if (document.getElementById(CONFIG.id)) this.refreshAllContext(); });
        es.on(et.WORLDINFO_UPDATED, (name, data) => { if (STATE.currentBookName === name) this.loadBook(name); });
        es.on(et.CHAT_CHANGED, () => { if (document.getElementById(CONFIG.id)) this.refreshAllContext(); });
        es.on(et.CHARACTER_SELECTED, () => {
            setTimeout(() => {
                if (document.getElementById(CONFIG.id)) this.refreshAllContext();
                else this.refreshAllContext();
            }, 100);
        });
        es.on(et.CHARACTER_EDITED, () => { if (document.getElementById(CONFIG.id)) this.refreshAllContext(); });

        STATE.isInitialized = true;
        await this.refreshAllContext();
        logger.info("Initialization complete.");
    },

    async refreshAllContext() {
        try {
            // 快速加载核心数据（不包含耗时的 boundBooksSet）
            const [all, char, glob, chat] = await Promise.all([
                API.getAllBookNames(),
                API.getCharBindings(),
                API.getGlobalBindings(),
                API.getChatBinding()
            ]);

            STATE.allBookNames = all.sort((a, b) => a.localeCompare(b));
            STATE.bindings.char = char;
            STATE.bindings.global = glob;
            STATE.bindings.chat = chat;
            STATE.metadata = API.getMetadata();

            // 清理已删除世界书的孤儿元数据
            const metaKeys = Object.keys(STATE.metadata);
            let needsSave = false;
            metaKeys.forEach(key => {
                if (!STATE.allBookNames.includes(key)) {
                    delete STATE.metadata[key];
                    needsSave = true;
                }
            });
            if (needsSave) {
                await API.saveMetadata(STATE.metadata);
            }

            // 后台异步加载绑定映射，避免阻塞UI
            this.loadBoundBooksSetAsync().then(boundSet => {
                STATE.boundBooksSet = boundSet;
                // 如果当前在管理视图，刷新以显示正确的绑定状态
                if (STATE.currentView === 'manage') {
                    UI.renderManageView();
                }
            }).catch(e => {
                logger.error("Async load of bound books set failed:", e);
            });

            UI.renderBookSelector();
        } catch (e) {
            logger.error("Failed to refresh context:", e);
        }
    },

    async switchView(viewName) {
        await this.flushPendingSave();
        UI.updateGlider(viewName);
        document.querySelectorAll('.wb-tab').forEach(el => {
            el.classList.toggle('active', el.dataset.tab === viewName);
        });

        setTimeout(() => {
            STATE.currentView = viewName;
            document.querySelectorAll('.wb-view-section').forEach(el => el.classList.add('wb-hidden'));
            const targetView = document.getElementById(`wb-view-${viewName}`);
            if (targetView) targetView.classList.remove('wb-hidden');

            if (viewName === 'binding') {
                UI.renderBindingView();
            } else if (viewName === 'manage') {
                // 修复一：确保视图显示后再计算滑块位置
                UI.updateManageGlider(STATE.manageTab);

                if (STATE.isManageDirty) {
                    UI.renderManageView();
                    // 这里原本直接修改，移到 renderManageView 内部管理
                }
            } else if (viewName === 'editor') {
                if (STATE.currentBookName && !STATE.allBookNames.includes(STATE.currentBookName)) {
                    STATE.currentBookName = null;
                    STATE.entries = [];
                    UI.renderList();
                }
                UI.renderBookSelector();
                UI.updateHeaderInfo();
            }
        }, 10);
    },

    async loadBook(name) {
        if (!name) return;
        await this.flushPendingSave();
        STATE.currentBookName = name;

        try {
            const loadedEntries = await API.loadBook(name);
            if (STATE.currentBookName !== name) return;

            STATE.entries = loadedEntries;
            STATE.entries.sort((a, b) => {
                const scoreA = this.getEntrySortScore(a);
                const scoreB = this.getEntrySortScore(b);
                if (scoreA !== scoreB) return scoreB - scoreA;
                return (a.order ?? 0) - (b.order ?? 0) || a.uid - b.uid;
            });

            UI.updateHeaderInfo();
            UI.renderList();

            const selector = document.getElementById('wb-book-selector');
            if (selector) selector.value = name;
        } catch (e) {
            if (STATE.currentBookName === name) {
                logger.error("Load book failed", e);
                toastr.error(`无法加载世界书 "${name}"`);
            }
        }
    },

    updateEntry(uid, updater) {
        const entry = STATE.entries.find(e => e.uid === uid);
        if (!entry) return;

        updater(entry);
        UI.updateCardStatus && UI.updateCardStatus(uid);
        UI.renderGlobalStats();

        // 更新搜索文本缓存
        const card = document.querySelector(`.wb-card[data-uid="${uid}"]`);
        if (card) {
            const comment = entry.comment || '';
            const content = entry.content || '';
            card.dataset.searchText = `${comment} ${content}`.toLowerCase();
            // 如果当前有搜索词，重新验证显隐状态
            const searchInput = document.getElementById('wb-search-entry');
            if (searchInput && searchInput.value) {
                const term = searchInput.value.toLowerCase();
                const hasMatch = card.dataset.searchText.includes(term);
                card.classList.toggle('hidden', !hasMatch);
            }
        }

        if (STATE.debouncer) clearTimeout(STATE.debouncer);
        const targetBookName = STATE.currentBookName;
        const targetEntries = STATE.entries;

        STATE.debouncer = setTimeout(() => {
            STATE.debouncer = null;
            if (targetBookName && targetEntries) {
                API.saveBookEntries(targetBookName, targetEntries);
            }
        }, 1500); // 延长防抖以减少 I/O
    },
    
    async addNewEntry() {
        if (!STATE.currentBookName) return toastr.warning("请先选择一本世界书");
        const maxUid = STATE.entries.reduce((max, e) => Math.max(max, Number(e.uid) || 0), -1);
        const newUid = maxUid + 1;

        const newEntry = {
            uid: newUid,
            comment: '新建条目', disable: false, content: '',
            constant: true, key: [], order: 0, position: 0, depth: 4, probability: 100, selective: true
        };
        await API.createEntry(STATE.currentBookName, [newEntry]);
        await this.loadBook(STATE.currentBookName);
    },

    async deleteEntry(uid) {
        if (!confirm("确定要删除此条目吗？")) return;
        await API.deleteEntries(STATE.currentBookName, [uid]);
        await this.loadBook(STATE.currentBookName);
    },

    sortByPriority() {
        STATE.entries.sort((a, b) => {
            const scoreA = this.getEntrySortScore(a);
            const scoreB = this.getEntrySortScore(b);
            if (scoreA !== scoreB) return scoreB - scoreA;
            const orderA = a.order ?? 0;
            const orderB = b.order ?? 0;
            if (orderA !== orderB) return orderA - orderB;
            return a.uid - b.uid;
        });

        UI.renderList();
        API.saveBookEntries(STATE.currentBookName, STATE.entries);
        toastr.success(`已重新按上下文逻辑重排`);
    },

    async batchSetExcludeRecursion(uids, value = true) {
        if (!STATE.currentBookName) throw new Error('No book loaded');
        let modified = false;
        STATE.entries.forEach(entry => {
            if (uids.includes(entry.uid) && entry.excludeRecursion !== value) {
                entry.excludeRecursion = value;
                modified = true;
            }
        });
        if (modified) {
            await API.saveBookEntries(STATE.currentBookName, STATE.entries);
            UI.renderGlobalStats();
            uids.forEach(uid => UI.updateCardStatus(uid));
        }
    },
    
    applyTheme(theme) {
      const themeBtn = document.getElementById('btn-wb-menu-theme-toggle');
      if (theme === 'light') {
          themeBtn.classList.replace('fa-moon', 'fa-sun');
          document.body.setAttribute('data-theme', 'light');
      } else {
          themeBtn.classList.replace('fa-sun', 'fa-moon');
          document.body.setAttribute('data-theme', 'dark');
      }
    },

    switchTheme() {
      const currentTheme = localStorage.getItem(THEME_KEY) === 'light' ? 'dark' : 'light';
      this.applyTheme(currentTheme);
      localStorage.setItem(THEME_KEY, currentTheme);
    },

    async saveBindings() {
        const view = document.getElementById('wb-view-binding');
        const charPrimary = view.querySelector('#wb-bind-char-primary').value;
        const charAddTags = view.querySelectorAll('.wb-ms-tag[data-bind-type="wb-bind-char-add"]');
        const charAdditional = Array.from(charAddTags).map(el => el.dataset.val);
        const globalTags = view.querySelectorAll('.wb-ms-tag[data-bind-type="wb-bind-global"]');
        const globalBooks = Array.from(globalTags).map(el => el.dataset.val);
        const chatBook = view.querySelector('#wb-bind-chat').value;

        try {
            // Get current global bindings to compute changes
            const currentGlobal = await API.getGlobalBindings();
            const toRemove = currentGlobal.filter(b => !globalBooks.includes(b));
            const toAdd = globalBooks.filter(b => !currentGlobal.includes(b));

            // Optimistically update local state and UI for immediate feedback
            STATE.bindings.char.primary = charPrimary || null;
            STATE.bindings.char.additional = charAdditional;
            STATE.bindings.global = globalBooks;
            STATE.bindings.chat = chatBook || null;

            UI.renderBookSelector();
            UI.renderBindingView();

            // Build all async operations in parallel
            const promises = [];

            // Primary binding
            promises.push(setCharBindings('primary', charPrimary || '', !!charPrimary));

            // Auxiliary binding (synchronous but triggers async save)
            const context = getContext();
            const charId = context.characterId;
            if (charId || charId === 0) {
                const charAvatar = context.characters[charId]?.avatar;
                const charFileName = getCharaFilename(null, { manualAvatarKey: charAvatar });
                charSetAuxWorlds(charFileName, charAdditional);
            }

            // Global bindings - execute in parallel
            toRemove.forEach(book => promises.push(setCharBindings('global', book, false)));
            toAdd.forEach(book => promises.push(setCharBindings('global', book, true)));

            // Chat binding
            promises.push(setCharBindings('chat', chatBook || '', !!chatBook));

            // Wait for all operations to settle
            await Promise.all(promises);

            // Do not force refresh here; rely on optimistic UI and eventual SETTINGS_UPDATED event
            toastr.success("绑定设置已保存");
        } catch (e) {
            toastr.error('保存失败: ' + e.message);
        }
    },

    getTokenCount(text) {
        if (!text) return 0;
        try {
            return getContext().getTokenCount(text); 
        } catch (e) {
            return Math.ceil(text.length / 3);
        }
    },
    
    getExistingGroups() {
        const groups = new Set();
        Object.values(STATE.metadata).forEach(m => {
            if (m.group && m.group !== '未分组') groups.add(m.group);
        });
        return Array.from(groups).sort();
    },

    async reorderEntry(fromIndex, toIndex) {
        if (fromIndex === toIndex) return;
        const [item] = STATE.entries.splice(fromIndex, 1);
        STATE.entries.splice(toIndex, 0, item);
        UI.renderList();
        await API.saveBookEntries(STATE.currentBookName, STATE.entries);
    },

    async updateMeta(bookName, updater) {
        if (!STATE.metadata[bookName]) { STATE.metadata[bookName] = { group: '', note: '' }; }
        updater(STATE.metadata[bookName]);
        await API.saveMetadata(STATE.metadata);
    },
    async setBookGroup(bookName, groupName) {
        await this.updateMeta(bookName, (meta) => { meta.group = groupName; });
        UI.renderManageView();
    },
    updateNote(bookName, note) { this.updateMeta(bookName, (meta) => { meta.note = note; }); },
    async togglePin(bookName) {
        await this.updateMeta(bookName, (meta) => { meta.pinned = !meta.pinned; });
        STATE.isManageDirty = true; // 强制完整重排以应用新的置顶顺序
        UI.renderManageView();
    },
    async addCustomTag(bookName, tagText) {
        await this.updateMeta(bookName, (meta) => {
            if (!meta.tags) meta.tags = [];
            if (meta.tags.length < 5 && !meta.tags.includes(tagText)) meta.tags.push(tagText);
        });
    },
    async removeCustomTag(bookName, index) {
        await this.updateMeta(bookName, (meta) => {
            if (meta.tags && meta.tags.length > index) meta.tags.splice(index, 1);
        });
    },

    async deleteBookDirectly(bookName) {
        if (!confirm(`确定要永久删除世界书 "${bookName}" 吗？`)) return;
        try {
            if (STATE.currentBookName === bookName && STATE.debouncer) {
                clearTimeout(STATE.debouncer);
                STATE.debouncer = null;
            }
            await API.deleteWorldbook(bookName);
            // 删除对应的元数据条目
            if (STATE.metadata[bookName]) {
                delete STATE.metadata[bookName];
                await API.saveMetadata(STATE.metadata);
            }
            if (STATE.currentBookName === bookName) {
                STATE.currentBookName = null;
                STATE.entries = [];
            }
            await this.refreshAllContext();
            STATE.isManageDirty = true;
            UI.renderManageView();
        } catch (e) { toastr.error("删除失败: " + e.message); }
    },

    async jumpToEditor(bookName) {
        await this.loadBook(bookName);
        this.switchView('editor');
    },

    async toggleBindState(bookName, targetCharName, isUnbind) {
        const context = getContext();
        const currentChar = context.characters[context.characterId]?.name;

        if (isUnbind) {
            if (!confirm(`确定要解除世界书 "${bookName}" 与角色 "${targetCharName}" 的绑定吗？`)) return;
            try {
                if (currentChar === targetCharName) await setCharBindings('primary', bookName, false);
                await this.refreshAllContext();
                STATE.isManageDirty = true;
                UI.renderManageView();
            } catch (e) { toastr.error("解绑失败: " + e.message); }
        } else {
            if (!currentChar) return toastr.warning("当前没有加载任何角色，无法绑定。");
            if (!confirm(`确定要将世界书 "${bookName}" 绑定为当前角色 "${currentChar}" 的主要世界书吗？`)) return;
            try {
                await setCharBindings('primary', bookName, true);
                await this.refreshAllContext();
                STATE.isManageDirty = true;
                if (bookName) await this.loadBook(bookName);
                UI.renderManageView();
            } catch (e) { toastr.error("绑定失败: " + e.message); }
        }
    },

    async actionImport() { document.getElementById('wb-import-file').click(); },
    
    async actionHandleImport(file) {
        if (!file) return;
        const reader = new FileReader();
        reader.onload = async (e) => {
            try {
                const content = JSON.parse(e.target.result);
                let entries = content.entries ? Object.values(content.entries) : content;
                if (!Array.isArray(entries)) entries = [];

                let bookName = file.name.replace(/\.(json|wb)$/i, '');
                const name = prompt("请输入导入后的世界书名称:", bookName);
                if (!name) return;

                if (STATE.allBookNames.includes(name)) {
                    if (!confirm(`世界书 "${name}" 已存在，是否覆盖？`)) return;
                }
                if (!STATE.allBookNames.includes(name)) await API.createWorldbook(name);
                await API.saveBookEntries(name, entries);
                toastr.success(`导入成功: ${name}`);
                await this.refreshAllContext();
                await this.loadBook(name);
            } catch (err) { toastr.error("导入失败: " + err.message); }
        };
        reader.readAsText(file);
    },

    async actionExport() {
        if (!STATE.currentBookName) return toastr.warning("请先选择一本世界书");
        try {
            const entries = await API.loadBook(STATE.currentBookName);
            const entriesObj = {};
            entries.forEach(entry => { entriesObj[entry.uid] = entry; });
            const exportData = { entries: entriesObj };
            const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${STATE.currentBookName}.json`;
            a.click();
            URL.revokeObjectURL(url);
        } catch (e) { toastr.error("导出失败: " + e.message); }
    },

    async actionExportTxt() {
        if (!STATE.currentBookName) return toastr.warning("请先选择一本世界书");

        const overlay = document.createElement('div');
        overlay.className = 'wb-sort-modal-overlay';
        overlay.style.zIndex = '25000';
        overlay.innerHTML = `
            <div class="wb-export-card">
                <div class="wb-export-header"><div class="wb-export-title">导出世界书为 TXT</div><div class="wb-export-close">×</div></div>
                <div class="wb-export-section"><div class="wb-export-label">导出所有条目</div><div class="wb-export-grid"><button class="wb-export-btn" data-type="all-title">含标题</button><button class="wb-export-btn" data-type="all-no-title">不含标题</button></div></div>
                <div class="wb-export-section"><div class="wb-export-label">仅导出已启用条目</div><div class="wb-export-grid"><button class="wb-export-btn" data-type="enabled-title">含标题</button><button class="wb-export-btn" data-type="enabled-no-title">不含标题</button></div></div>
            </div>`;
        document.body.appendChild(overlay);

        const processExport = (type) => {
            try {
                let targetEntries = [...STATE.entries];
                if (type.startsWith('enabled')) targetEntries = targetEntries.filter(e => !e.disable);
                targetEntries.sort((a, b) => {
                    const scoreA = this.getEntrySortScore(a);
                    const scoreB = this.getEntrySortScore(b);
                    if (scoreA !== scoreB) return scoreB - scoreA;
                    return (a.order ?? 0) - (b.order ?? 0) || a.uid - b.uid;
                });

                if (targetEntries.length === 0) return toastr.warning("没有符合条件的条目可导出");
                const includeTitle = !type.includes('no-title');
                let txtContent = "";
                targetEntries.forEach(entry => {
                    const title = entry.comment || '无标题条目';
                    const content = entry.content || '';
                    if (includeTitle) txtContent += `#### ${title}\n${content}\n\n`;
                    else txtContent += `${content}\n\n`;
                });

                const scopeName = type.startsWith('enabled') ? '仅启用' : '所有';
                const formatName = includeTitle ? '含标题' : '无标题';
                const fileName = `${STATE.currentBookName}_${scopeName}_${formatName}.txt`;
                const blob = new Blob([txtContent], { type: 'text/plain' });
                const url = URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = fileName;
                a.click();
                URL.revokeObjectURL(url);
                toastr.success(`导出成功: ${fileName}`);
                overlay.remove();
            } catch (e) { toastr.error("导出失败: " + e.message); }
        };

        overlay.querySelector('.wb-export-close').onclick = () => overlay.remove();
        overlay.querySelectorAll('.wb-export-btn').forEach(btn => { btn.onclick = () => processExport(btn.dataset.type); });
        overlay.onclick = (e) => { if (e.target === overlay) overlay.remove(); };
    },

    async actionCreateNew() {
        const name = prompt("请输入新世界书名称:");
        if (!name) return;
        if (STATE.allBookNames.includes(name)) return toastr.warning("该名称已存在");
        try {
            await API.createWorldbook(name);
            await this.refreshAllContext();
            await this.loadBook(name);
        } catch (e) { toastr.error("创建失败: " + e.message); }
    },

    async actionDelete() {
        if (!STATE.currentBookName) return;
        if (!confirm(`确定要永久删除世界书 "${STATE.currentBookName}" 吗？`)) return;
        try {
            if (STATE.debouncer) { clearTimeout(STATE.debouncer); STATE.debouncer = null; }
            await API.deleteWorldbook(STATE.currentBookName);
            STATE.currentBookName = null;
            STATE.entries = [];
            await this.refreshAllContext();
            await this.init(); 
        } catch (e) { toastr.error("删除失败: " + e.message); }
    },

    async actionRename() {
        if (!STATE.currentBookName) return;
        const newName = prompt("重命名世界书为:", STATE.currentBookName);
        if (!newName || newName === STATE.currentBookName) return;
        if (STATE.allBookNames.includes(newName)) return toastr.warning("目标名称已存在");

        try {
            await this.flushPendingSave();
            await API.renameWorldbook(STATE.currentBookName, newName);
            await this.refreshAllContext();
            await this.loadBook(newName);
        } catch (e) { toastr.error("重命名失败: " + e.message); }
    },

    getGlobalConfig() {
        const allMeta = API.getMetadata() || {};
        const config = allMeta['__GLOBAL_CONFIG__'] || {};
        if (config.deleteWbWithChar === undefined) config.deleteWbWithChar = true;
        if (config.showLogButton === undefined) config.showLogButton = false;
        return config;
    },
    
    async saveGlobalConfig(newConfig) {
        const allMeta = API.getMetadata() || {};
        allMeta['__GLOBAL_CONFIG__'] = { ...allMeta['__GLOBAL_CONFIG__'], ...newConfig };
        await API.saveMetadata(allMeta);
    },
    
    registerCharDeleteListener() {
        const es = eventSource;
        const et = event_types;
        if (!es) return;

        es.on(et.CHARACTER_DELETED, async (data) => {
             const config = this.getGlobalConfig();
             if (!config.deleteWbWithChar) return;

             const charName = data.character?.name || data.name;
             if (!charName) return;

             const map = await API.getAllBoundBookNames();
             let bookName = null;
             // 适配新版 getAllBoundBookNames 返回结构：{ wbName: { primary: [], aux: [], global: false, chat: [] } }
             for (const [wb, bindInfo] of Object.entries(map)) {
                 // 只检查 primary 绑定：角色作为该世界书的主要绑定者
                 if (bindInfo.primary && bindInfo.primary.includes(charName)) {
                     bookName = wb;
                     break;
                 }
             }

             if (bookName) {
                 UI.showDeleteWbConfirmModal(bookName, async () => {
                     await API.deleteWorldbook(bookName);
                 }, async () => {
                     await this.saveGlobalConfig({ deleteWbWithChar: false });
                     if (STATE.currentView === 'manage') UI.renderManageView();
                 });
             }
        });
    },

    async removeTagFromWorldbook(bookName, tagToRemove) {
        const meta = STATE.metadata[bookName] || {};
        const tags = meta.tags || [];
        const newTags = tags.filter(tag => tag !== tagToRemove);
        await this.updateMeta(bookName, (meta) => { meta.tags = newTags; });
        STATE.isManageDirty = true;
        const searchInput = document.getElementById('wb-manage-search');
        if (searchInput) {
            UI.renderManageView(searchInput.value);
        } else {
            UI.renderManageView();
        }
    }
};