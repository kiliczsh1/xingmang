# Experience Shares Plugin

This folder packages the XingNovel "经验分享" module and includes a template-adapted single-file entry for `window.XingNovelApp.registerPlugin(...)`.

## Outputs

- `experience-shares.template.js`: single-file plugin entry that matches the provided XingNovel template
- `build-single-file.js`: rebuilds the template entry from the files in this folder
- `frontend/`: Vue pages, supporting components, API wrapper, and types
- `backend/`: Express route and SQL migration
- `docs/`: integration notes
- `.codex-plugin/plugin.json`: plugin manifest

## Template Adaptation

The generated template file preserves the full plugin bundle as embedded base64 assets under `bundle.files`, while exposing a template-compatible plugin registration with metadata and actions.

## Quick Use

1. Load `experience-shares.template.js` in the XingNovel plugin host.
2. Use the `打开经验分享` action to jump to `/experience-shares`.
3. Use the `显示插件信息` action to confirm the bundle is registered.
4. If you need source-level integration, follow `docs/integration.md`.
