# Experience Shares Plugin Integration

This plugin package extracts the full "经验分享" module from XingNovel into a reusable bundle.

## Included code

- `frontend/views/ExperienceShares.vue`
- `frontend/views/ExperienceShareEditor.vue`
- `frontend/views/ExperienceShareDetail.vue`
- `frontend/components/SplitRichTextEditor.vue`
- `frontend/components/PdfPageGallery.vue`
- `frontend/components/MarkdownRenderer.vue`
- `frontend/api/experience-shares.ts`
- `frontend/types/experience-share.ts`
- `backend/routes/experienceShares.js`
- `backend/database/experience-shares-migration.sql`

## Host project wiring

### Frontend routes

Register these routes in your Vue Router setup:

- `/experience-shares`
- `/experience-shares/:id`
- `/experience-shares/:id/edit`

### Frontend navigation

Add a left-nav menu item pointing to `/experience-shares`.

If your layout highlights active menu items by exact path, map any `/experience-shares/*` path back to `/experience-shares`.

### Frontend dependencies

Required packages:

- `vue`
- `vue-router`
- `element-plus`
- `@element-plus/icons-vue`
- `axios`
- `markdown-it`
- `highlight.js`
- `pdfjs-dist`

### API layer

Expose the API wrapper from `frontend/api/experience-shares.ts` through your host project's shared API module.

### Type exports

Expose `ExperienceShare` and `ApiResponse` from `frontend/types/experience-share.ts` through your host project's shared types module.

### Backend route mount

In your Express app:

```js
const experienceShareRoutes = require('./routes/experienceShares')
app.use('/api/experience-shares', experienceShareRoutes)
```

### Uploads

Ensure your host already serves:

```js
app.use('/api/uploads', express.static(uploadsDir))
```

The route stores uploaded files under `uploads/experience-shares/`.

### Database migration

Create or migrate the `experience_shares` table using `backend/database/experience-shares-migration.sql`.

## Notes

- The route supports manual content and PDF-import content.
- PDF import relies on `pdfjs-dist/legacy/build/pdf.mjs` on the server.
- The frontend components assume your host project already has the same aliasing strategy or that you will rewrite `@/` imports after copy-in.
