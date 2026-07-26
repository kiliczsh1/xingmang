# Experience Shares Plugin

This folder packages the XingNovel "经验分享" board as a reusable plugin bundle.

## Purpose

Use this package when you want to transplant the experience-sharing module into another Vue + Express project without re-collecting its code manually.

## Structure

- `frontend/`: Vue pages, supporting components, API wrapper, and types
- `backend/`: Express route and SQL migration
- `docs/`: integration notes
- `.codex-plugin/plugin.json`: plugin manifest

## Quick use

1. Copy the frontend files into your host app.
2. Register the three routes and the left-nav entry.
3. Copy the backend route and mount `/api/experience-shares`.
4. Run the SQL migration for `experience_shares`.
5. Confirm `pdfjs-dist` and upload static serving are enabled.

See `docs/integration.md` for details.
