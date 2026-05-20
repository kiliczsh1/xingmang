// index.js

import { event_types, eventSource } from '../../../../script.js';
import { world_names } from '../../../world-info.js';
import { CONFIG } from './state.js';
import { Actions } from './actions.js';
import { UI } from './ui.js';
import { logger } from './logger.js';

jQuery(async () => {

    const injectButton = () => {
        if (document.getElementById(CONFIG.btnId)) return;
        const container = document.querySelector('#options .options-content');

        if (container) {
            const targetClasses = 'interactable';
            const html = `
                <a id="${CONFIG.btnId}" class="${targetClasses}" title="世界书管理" tabindex="0">
                    <i class="fa-lg fa-solid fa-book-journal-whills"></i>
                    <span>世界书</span>
                </a>
            `;

            $(container).append(html);

            $(`#${CONFIG.btnId}`).on('click', (e) => {
                e.preventDefault();
                $('#options').hide();
                UI.open();
            });

            logger.info("Button injected successfully into .options-content.");
        } else {
            logger.warn("Target container #options .options-content not found. Button injection skipped.");
        }
    };

    injectButton();

    const performInit = async () => {
        try {
            await Actions.init();
            logger.info("Pre-loading complete.");
        } catch (e) {
            logger.error("Pre-loading failed:", e);
        }
    };

    if (typeof world_names === 'undefined') {
        logger.info("Waiting for APP_READY...");
        eventSource.on(event_types.APP_READY, performInit);
    } else {
        performInit();
    }

    logger.info("Enhanced Script Loaded");
});