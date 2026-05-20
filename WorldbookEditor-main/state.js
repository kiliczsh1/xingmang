// state.js

export const CONFIG = {
    id: 'enhanced-wb-panel-v6',
    btnId: 'wb-menu-btn-v6',
    settingsKey: 'WorldbookEditor_Metadata',
    showLogButton: false, // 开发者日志按钮开关，默认关闭
    colors: {
        accent: '#7c5cbd',
    }
};

export const THEME_KEY = 'wb_ctx_theme';

export const STATE = {
    currentView: 'editor', // 'editor' | 'binding' | 'manage'
    currentBookName: null,

    isInitialized: false,
    isManageDirty: true,

    entries: [],
    allBookNames: [],
    metadata: {},

    manageFilters: {
        global: true,
        primary: true,
        aux: true,
        chat: true
    },
    manageTab: 'worldbook',
    manageExpandedGroups: new Set(),

    boundBooksSet: {},

    bindings: {
        char: { primary: null, additional: [] },
        global: [],
        chat: null
    },

    debouncer: null,

    boundBooksCacheVersion: 0,
    boundBooksCache: null
};

export const WI_POSITION_MAP = {
    0: 'before_character_definition',
    1: 'after_character_definition',
    2: 'before_author_note',
    3: 'after_author_note',
    4: 'at_depth',
    5: 'before_example_messages',
    6: 'after_example_messages'
};

export const WI_POSITION_MAP_REV = Object.fromEntries(
    Object.entries(WI_POSITION_MAP).map(([k, v]) => [v, parseInt(k)])
);