// api.js

import { getContext } from '../../../extensions.js';
import { getCharaFilename } from '../../../utils.js';
import { world_info, world_names, selected_world_info } from '../../../world-info.js';
import { CONFIG, STATE } from './state.js';
import { logger } from './logger.js';

/**
 * [兼容性 Polyfill] 更新角色主要世界书
 */
export async function charUpdatePrimaryWorld(name) {
    const context = getContext();
    const charId = context.characterId;
    if (charId === undefined || charId === null) return;

    const character = context.characters[charId];
    if (!character) return;

    if (!character.data.extensions) character.data.extensions = {};
    character.data.extensions.world = name;

    const uiSelect = document.getElementById('character_world');
    if (uiSelect) {
        uiSelect.value = name;
        uiSelect.dispatchEvent(new Event('change'));
    }

    const setWorldBtn = document.getElementById('set_character_world');
    if (setWorldBtn) {
        if (name) setWorldBtn.classList.add('world_set');
        else setWorldBtn.classList.remove('world_set');
    }

    if (context.saveCharacterDebounced) {
        context.saveCharacterDebounced();
    }
}

/**
 * [兼容性 Polyfill] 设置角色辅助世界书列表
 */
export function charSetAuxWorlds(fileName, books) {
    const context = getContext();

    if (!world_info.charLore) world_info.charLore = [];

    const idx = world_info.charLore.findIndex(e => e.name === fileName);

    if (books.length === 0) {
        if (idx !== -1) world_info.charLore.splice(idx, 1);
    } else if (idx === -1) {
        world_info.charLore.push({ name: fileName, extraBooks: books });
    } else {
        world_info.charLore[idx].extraBooks = books;
    }

    if (context.saveSettingsDebounced) {
        context.saveSettingsDebounced();
    }
}

/**
 * 处理不同类型的世界书绑定
 */
export async function setCharBindings(type, worldName, isEnabled) {
    const context = getContext();

    // Common character info
    let charId = context.characterId;
    let character = null;
    if (charId !== undefined && charId !== null) {
        character = context.characters[charId];
    }
    const charName = character?.name || 'Unknown';

    if (type === 'primary') {
        if (!character) return;
        const oldPrimary = character.data?.extensions?.world || null;
        const targetName = isEnabled ? worldName : '';
        await charUpdatePrimaryWorld(targetName);
        // Remove from old primary if different
        if (oldPrimary && oldPrimary !== targetName) {
            updateBoundBooksCacheEntry(oldPrimary, {
                primary: { add: [], remove: [charName] }
            });
        }
        // Add to new primary if any
        if (targetName) {
            updateBoundBooksCacheEntry(targetName, {
                primary: { add: [charName], remove: [] }
            });
        }
        return;
    }

    if (type === 'auxiliary') {
        if (!character) return;
        const charAvatar = character.avatar;
        const charFileName = getCharaFilename(null, { manualAvatarKey: charAvatar });

        const charLoreEntry = world_info.charLore?.find(e => e.name === charFileName);
        let currentBooks = charLoreEntry ? [...charLoreEntry.extraBooks] : [];

        if (isEnabled) {
            if (!currentBooks.includes(worldName)) currentBooks.push(worldName);
        } else {
            currentBooks = currentBooks.filter(name => name !== worldName);
        }

        charSetAuxWorlds(charFileName, currentBooks);

        // Only the target worldbook's aux changes
        updateBoundBooksCacheEntry(worldName, {
            aux: { add: isEnabled ? [charName] : [], remove: isEnabled ? [] : [charName] }
        });
        return;
    }

    if (type === 'chat') {
        if (!character) return;
        const chatName = context.chatId?.replace('.jsonl', '') || '';
        // For single character mode, entity is always the character
        const entityName = charName;
        const oldChatWb = context.chatMetadata?.world_info || null;

        if (isEnabled) {
            context.chatMetadata['world_info'] = worldName;
            // Remove from old chat binding if any and different
            if (oldChatWb && oldChatWb !== worldName) {
                updateBoundBooksCacheEntry(oldChatWb, {
                    chat: { add: [], remove: [{ charName: entityName, chatName }] }
                });
            }
            // Add to new chat binding
            updateBoundBooksCacheEntry(worldName, {
                chat: { add: [{ charName: entityName, chatName }], remove: [] }
            });
        } else {
            if (oldChatWb === worldName) {
                delete context.chatMetadata['world_info'];
                updateBoundBooksCacheEntry(worldName, {
                    chat: { add: [], remove: [{ charName: entityName, chatName }] }
                });
            }
        }
        context.saveMetadataDebounced();
        return;
    }

    if (type === 'global') {
        const command = isEnabled
            ? `/world silent=true "${worldName}"`
            : `/world state=off silent=true "${worldName}"`;
        await context.executeSlashCommands(command);
        updateBoundBooksCacheEntry(worldName, { global: isEnabled });
        return;
    }

    logger.warn(`Unknown binding type: ${type}`);
}

function invalidateBoundBooksCache() {
    STATE.boundBooksCacheVersion++;
    STATE.boundBooksCache = null;
}

function updateBoundBooksCacheEntry(bookName, delta) {
    if (!bookName) return;
    if (!STATE.boundBooksCache || STATE.boundBooksCache.version !== STATE.boundBooksCacheVersion) {
        return; // Cache invalid – skip incremental update, will rebuild later.
    }
    const cacheData = STATE.boundBooksCache.data;
    if (!cacheData[bookName]) {
        cacheData[bookName] = { primary: [], aux: [], global: false, chat: [] };
    }
    const entry = cacheData[bookName];

    // Primary updates
    if (delta.primary) {
        if (Array.isArray(delta.primary.add)) {
            for (const name of delta.primary.add) {
                if (!entry.primary.includes(name)) entry.primary.push(name);
            }
        }
        if (Array.isArray(delta.primary.remove)) {
            entry.primary = entry.primary.filter(name => !delta.primary.remove.includes(name));
        }
    }

    // Auxiliary updates
    if (delta.aux) {
        if (Array.isArray(delta.aux.add)) {
            for (const name of delta.aux.add) {
                if (!entry.aux.includes(name)) entry.aux.push(name);
            }
        }
        if (Array.isArray(delta.aux.remove)) {
            entry.aux = entry.aux.filter(name => !delta.aux.remove.includes(name));
        }
    }

    // Global flag update
    if (typeof delta.global === 'boolean') {
        entry.global = delta.global;
    }

    // Chat binding updates
    if (delta.chat) {
        if (Array.isArray(delta.chat.add)) {
            for (const item of delta.chat.add) {
                if (!item || typeof item !== 'object') continue;
                const exists = entry.chat.some(c => c.charName === item.charName && c.chatName === item.chatName);
                if (!exists) entry.chat.push(item);
            }
        }
        if (Array.isArray(delta.chat.remove)) {
            entry.chat = entry.chat.filter(c => !delta.chat.remove.some(r => r && c.charName === r.charName && c.chatName === r.chatName));
        }
    }
}

export const API = {
    async getAllBookNames() {
        return [...(world_names || [])].sort((a, b) => a.localeCompare(b));
    },

    async getCharBindings() {
        const context = getContext();
        const charId = context.characterId;
        if (charId === undefined || charId === null) return { primary: null, additional: [] };

        const character = context.characters[charId];
        if (!character) return { primary: null, additional: [] };

        const primary = character.data?.extensions?.world || null;

        let additional = [];
        const charFileName = getCharaFilename(null, { manualAvatarKey: character.avatar });

        const charLore = world_info.charLore || [];
        const entry = charLore.find(e => e.name === charFileName);
        if (entry && Array.isArray(entry.extraBooks)) {
            additional = [...entry.extraBooks];
        }

        return { primary, additional };
    },

    async getGlobalBindings() {
        return [...(selected_world_info || [])];
    },

    async getChatBinding() {
        const context = getContext();
        return context.chatMetadata?.world_info || null;
    },

    async loadBook(name) {
        const data = await getContext().loadWorldInfo(name);
        if (!data) throw new Error(`Worldbook ${name} not found`);

        const safeEntries = data.entries ? structuredClone(data.entries) : {};
        const entries = Object.values(safeEntries);
        return entries.sort((a, b) => (a.order ?? 0) - (b.order ?? 0));
    },

    async saveBookEntries(name, entriesArray) {
        if (!name || !Array.isArray(entriesArray)) {
            logger.warn("Save aborted: Invalid name or entries.");
            return;
        }

        const oldData = await getContext().loadWorldInfo(name) || { entries: {} };
        const newEntriesObj = {};

        entriesArray.forEach(entry => {
            const uid = entry.uid;
            const oldEntry = (oldData.entries && oldData.entries[uid]) ? oldData.entries[uid] : {};
            const safeEntry = structuredClone(entry);
            newEntriesObj[uid] = { ...oldEntry, ...safeEntry };
        });

        const newData = { ...oldData, entries: newEntriesObj };
        await getContext().saveWorldInfo(name, newData, false);
    },

    async createEntry(name, newEntriesArray) {
        const currentEntries = await this.loadBook(name);
        const combined = [...newEntriesArray, ...currentEntries];
        await this.saveBookEntries(name, combined);
    },

    async deleteEntries(name, uidsToDelete) {
        let currentEntries = await this.loadBook(name);
        currentEntries = currentEntries.filter(e => !uidsToDelete.includes(e.uid));
        await this.saveBookEntries(name, currentEntries);
    },

    async getAllBoundBookNames(forceRefresh = false) {
        // Return cached result if valid
        if (!forceRefresh && STATE.boundBooksCache && STATE.boundBooksCache.version === STATE.boundBooksCacheVersion) {
            return STATE.boundBooksCache.data;
        }

        const context = getContext();
        if (typeof context.getCharacters === 'function') {
            await context.getCharacters();
        }

        const characters = context.characters || [];
        const boundMap = {};

        const initBook = (name) => {
            if (name && !boundMap[name]) {
                boundMap[name] = { primary: [], aux: [], global: false, chat: [] };
            }
        };

        // 处理 Persona (用户画像) 绑定的世界书
        const pus = context.powerUserSettings || {};
        if (pus.persona_descriptions) {
            Object.keys(pus.persona_descriptions).forEach(avatarId => {
                const pData = pus.persona_descriptions[avatarId];
                if (pData && pData.lorebook) {
                    initBook(pData.lorebook);
                    const niceName = (pus.personas && pus.personas[avatarId]) ? pus.personas[avatarId] : avatarId;
                    const displayName = `👤用户: ${niceName}`;
                    if (!boundMap[pData.lorebook].primary.includes(displayName)) {
                        boundMap[pData.lorebook].primary.push(displayName);
                    }
                }
            });
        }

        // 恢复角色的分批获取逻辑，只获取角色卡信息，彻底抛弃导致严重卡顿的"历史聊天获取"
        const batchSize = 30;
        for (let i = 0; i < characters.length; i += batchSize) {
            const chunk = characters.slice(i, i + batchSize);
            await Promise.all(chunk.map(async (charItem) => {
                if (!charItem || !charItem.avatar) return;

                let charData = charItem;
                // 如果内存中只是浅层对象，则请求完整数据以获取绑定的世界书
                if (charItem.shallow || !charItem.data) {
                    try {
                        const res = await fetch('/api/characters/get', {
                            method: 'POST',
                            headers: context.getRequestHeaders(),
                            body: JSON.stringify({ avatar_url: charItem.avatar })
                        });
                        if (res.ok) charData = await res.json();
                    } catch (e) {
                        console.warn(`Failed to load full data for ${charItem.name}`, e);
                    }
                }

                const dataFields = charData.data || charData;
                const curCharName = charData.name || charItem.name || '未知角色';

                if (dataFields) {
                    const primaryWb = dataFields.extensions?.world || dataFields.world || dataFields.world_info || dataFields.lorebook || dataFields.character_book?.name || (typeof dataFields.character_book === 'string' ? dataFields.character_book : null);
                    if (primaryWb && typeof primaryWb === 'string') {
                        initBook(primaryWb);
                        if (!boundMap[primaryWb].primary.includes(curCharName)) {
                            boundMap[primaryWb].primary.push(curCharName);
                        }
                    }
                }
            }));
        }

        // 缓存结果
        STATE.boundBooksCache = { data: boundMap, version: STATE.boundBooksCacheVersion };
        return boundMap;
    },

    getMetadata() {
        const context = getContext();
        return context.extensionSettings[CONFIG.settingsKey] || {};
    },
    async saveMetadata(data) {
        const context = getContext();
        context.extensionSettings[CONFIG.settingsKey] = data;
        context.saveSettingsDebounced();
    },

    async createWorldbook(name) {
        await getContext().saveWorldInfo(name, { entries: {} }, true);
        await getContext().updateWorldInfoList();
    },
    async deleteWorldbook(name) {
        await fetch('/api/worldinfo/delete', {
            method: 'POST',
            headers: getContext().getRequestHeaders(),
            body: JSON.stringify({ name }),
        });
        await getContext().updateWorldInfoList();
        invalidateBoundBooksCache();
    },
    async renameWorldbook(oldName, newName) {
        const data = await getContext().loadWorldInfo(oldName);
        if (data) {
            // Show loading overlay
            const loader = document.createElement('div');
            loader.className = 'wb-sort-modal-overlay';
            loader.style.zIndex = '25000';
            loader.innerHTML = `
                <div class="wb-export-card" style="width:320px;text-align:center;padding:24px 16px;">
                    <div style="font-size:2em;color:var(--SmartThemeAccentColor);margin-bottom:12px;"><i class="fa-solid fa-circle-notch fa-spin"></i></div>
                    <div style="font-weight:600">正在重命名世界书并同步绑定信息中，请稍候...</div>
                </div>`;
            document.body.appendChild(loader);

            try {
                await getContext().saveWorldInfo(newName, data, true);

                const context = getContext();

                // Migrate current character's primary binding if set
                const { primary } = await this.getCharBindings();
                if (primary === oldName) {
                    const charId = context.characterId;
                    if (charId != null && context.characters[charId]) {
                        context.characters[charId].data.extensions = context.characters[charId].data.extensions || {};
                        context.characters[charId].data.extensions.world = newName;
                        if (context.saveCharacterDebounced) context.saveCharacterDebounced();
                    }
                }

                // Migrate auxiliary bindings globally (all characters)
                const charLore = world_info.charLore || [];
                let auxModified = false;
                for (const entry of charLore) {
                    if (Array.isArray(entry.extraBooks)) {
                        const idx = entry.extraBooks.indexOf(oldName);
                        if (idx !== -1) {
                            entry.extraBooks[idx] = newName;
                            auxModified = true;
                        }
                    }
                }
                if (auxModified) {
                    world_info.charLore = charLore;
                    context.saveSettingsDebounced();
                }

                // Migrate global bindings
                const globalBindings = selected_world_info || [];
                const gIdx = globalBindings.indexOf(oldName);
                if (gIdx !== -1) {
                    globalBindings[gIdx] = newName;
                    world_info.globalSelect = globalBindings;
                    context.saveSettingsDebounced();
                }

                // Migrate chat binding
                if (context.chatMetadata?.world_info === oldName) {
                    context.chatMetadata.world_info = newName;
                    context.saveMetadataDebounced();
                }

                // Migrate cache entry to new name before deletion, if present
                if (STATE.boundBooksCache && STATE.boundBooksCache.data[oldName]) {
                    STATE.boundBooksCache.data[newName] = STATE.boundBooksCache.data[oldName];
                    delete STATE.boundBooksCache.data[oldName];
                }
                await this.deleteWorldbook(oldName);
                invalidateBoundBooksCache();
            } catch (e) {
                logger.error("Rename failed:", e);
                toastr.error("重命名过程中出现错误: " + e.message);
            } finally {
                loader.remove();
            }
        }
    }
};