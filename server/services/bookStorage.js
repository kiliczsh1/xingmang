const fs = require('fs');
const path = require('path');
const db = require('../database/db');

const HIDDEN_DIR_NAME = '.xingnovel';
const META_FILE_NAME = 'meta.json';
const INDEX_FILE_NAME = 'index.json';
const BOOK_INFO_FILE_NAME = '作品信息.md';

const isPortablePackage = () => {
  if (process.pkg) return true;
  return process.execPath.includes('XingNovel-node');
};

const getRootDir = () => {
  if (isPortablePackage()) {
    return path.dirname(process.execPath);
  }
  return path.join(__dirname, '..', '..');
};

const getLibraryRoot = () => path.join(getRootDir(), 'data', 'books');
const getLibraryHiddenDir = () => path.join(getLibraryRoot(), HIDDEN_DIR_NAME);
const getIndexPath = () => path.join(getLibraryHiddenDir(), INDEX_FILE_NAME);

const ensureDir = dirPath => {
  fs.mkdirSync(dirPath, { recursive: true });
};

const readJson = (filePath, fallback) => {
  try {
    if (!fs.existsSync(filePath)) return fallback;
    return JSON.parse(fs.readFileSync(filePath, 'utf8'));
  } catch (error) {
    return fallback;
  }
};

const writeJson = (filePath, value) => {
  ensureDir(path.dirname(filePath));
  fs.writeFileSync(filePath, JSON.stringify(value, null, 2), 'utf8');
};

const formatTimestamp = (input = new Date()) => {
  const date = input instanceof Date ? input : new Date(input);
  const pad = num => String(num).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
};

const sanitizeName = value => {
  const normalized = String(value || '')
    .replace(/[\\/:*?"<>|]/g, ' ')
    .replace(/[.\s]+$/g, '')
    .replace(/\s+/g, ' ')
    .trim();
  return normalized || '未命名';
};

const slugify = value => sanitizeName(value).replace(/\s+/g, '-');

const padOrder = (value, length = 3) => String(Math.max(Number(value) || 0, 0) + 1).padStart(length, '0');

const toBookRecord = book => ({
  id: Number(book.id),
  title: book.title || '未命名作品',
  description: book.description || '',
  cover: book.cover || '',
  author: book.author || '',
  category: book.category || '',
  tags: Array.isArray(book.tags) ? book.tags : [],
  status: book.status || 'draft',
  is_public: Boolean(book.is_public),
  created_at: book.created_at || formatTimestamp(),
  updated_at: book.updated_at || formatTimestamp()
});

const createEmptyIndex = () => ({
  version: 1,
  books: {},
  nextVolumeId: 1,
  nextChapterId: 1
});

const ensureLibrary = () => {
  ensureDir(getLibraryRoot());
  ensureDir(getLibraryHiddenDir());
  if (!fs.existsSync(getIndexPath())) {
    writeJson(getIndexPath(), createEmptyIndex());
  }
};

const readIndex = () => {
  ensureLibrary();
  const index = readJson(getIndexPath(), createEmptyIndex());
  index.books = index.books || {};
  index.nextVolumeId = Number(index.nextVolumeId) || 1;
  index.nextChapterId = Number(index.nextChapterId) || 1;
  return index;
};

const writeIndex = index => {
  writeJson(getIndexPath(), index);
};

const getBookDirById = (index, bookId) => {
  const directoryName = index.books[String(bookId)];
  if (!directoryName) return '';
  return path.join(getLibraryRoot(), directoryName);
};

const getMetaPath = bookDir => path.join(bookDir, HIDDEN_DIR_NAME, META_FILE_NAME);

const writeBookInfoMarkdown = meta => {
  const tags = Array.isArray(meta.book.tags) && meta.book.tags.length > 0 ? meta.book.tags.join('、') : '无';
  const lines = [
    `# ${meta.book.title}`,
    '',
    `- 作者：${meta.book.author || '未填写'}`,
    `- 分类：${meta.book.category || '未填写'}`,
    `- 状态：${meta.book.status || 'draft'}`,
    `- 标签：${tags}`,
    `- 创建时间：${meta.book.created_at}`,
    `- 更新时间：${meta.book.updated_at}`,
    ''
  ];

  if (meta.book.description) {
    lines.push('## 简介', '', meta.book.description, '');
  }

  const filePath = path.join(meta.book.base_path, BOOK_INFO_FILE_NAME);
  fs.writeFileSync(filePath, lines.join('\n'), 'utf8');
};

const readBookMeta = bookDir => {
  const meta = readJson(getMetaPath(bookDir), null);
  if (!meta) return null;
  meta.book = meta.book || {};
  meta.book.base_path = bookDir;
  meta.book.tags = Array.isArray(meta.book.tags) ? meta.book.tags : [];
  meta.volumes = (Array.isArray(meta.volumes) ? meta.volumes : []).map(volume => ({
    ...volume,
    parent_id: volume.parent_id ? Number(volume.parent_id) : null
  }));
  meta.chapters = Array.isArray(meta.chapters) ? meta.chapters : [];
  return meta;
};

const writeBookMeta = meta => {
  ensureDir(meta.book.base_path);
  ensureDir(path.join(meta.book.base_path, HIDDEN_DIR_NAME));
  writeJson(getMetaPath(meta.book.base_path), {
    book: {
      ...meta.book,
      base_path: undefined
    },
    volumes: meta.volumes,
    chapters: meta.chapters
  });
  writeBookInfoMarkdown(meta);
};

const getUniqueDirectoryName = (parentDir, rawName, excludeName = '') => {
  const baseName = slugify(rawName);
  let candidate = baseName;
  let counter = 2;
  while (fs.existsSync(path.join(parentDir, candidate)) && candidate !== excludeName) {
    candidate = `${baseName}-${counter}`;
    counter += 1;
  }
  return candidate;
};

const getVolumeFolderName = (meta, volume, excludeName = '') => {
  const label = `${padOrder(volume.order_num)}-${sanitizeName(volume.title)}`;
  const parentDir = getVolumeParentDir(meta, volume.parent_id);
  return getUniqueDirectoryName(parentDir, label, excludeName);
};

const getVolumeById = (meta, volumeId) => {
  if (!volumeId) return null;
  return meta.volumes.find(item => Number(item.id) === Number(volumeId)) || null;
};

const getVolumeParentDir = (meta, parentId) => {
  if (!parentId) {
    return meta.book.base_path;
  }
  const parentVolume = getVolumeById(meta, parentId);
  if (!parentVolume) {
    return meta.book.base_path;
  }
  return path.join(meta.book.base_path, parentVolume.folder_name);
};

const getVolumeDir = (meta, volume) => {
  if (!volume) return meta.book.base_path;
  return path.join(meta.book.base_path, volume.folder_name);
};

const getVolumeDescendantIds = (meta, volumeId) => {
  const descendants = new Set();
  const stack = [Number(volumeId)];
  while (stack.length > 0) {
    const currentId = stack.pop();
    meta.volumes.forEach(volume => {
      if (Number(volume.parent_id) === Number(currentId) && !descendants.has(volume.id)) {
        descendants.add(volume.id);
        stack.push(volume.id);
      }
    });
  }
  return descendants;
};

const getChapterBaseName = chapter => `${padOrder(chapter.order_num, 4)}-${sanitizeName(chapter.title || '新章节')}`;

const getChapterParentDir = (meta, chapter) => {
  if (!chapter.volume_id) {
    return meta.book.base_path;
  }
  const volume = getVolumeById(meta, chapter.volume_id);
  if (!volume) {
    return meta.book.base_path;
  }
  return getVolumeDir(meta, volume);
};

const getUniqueChapterFileName = (meta, chapter, excludeName = '') => {
  const parentDir = getChapterParentDir(meta, chapter);
  ensureDir(parentDir);
  const baseName = getChapterBaseName(chapter);
  let candidate = `${baseName}.md`;
  let counter = 2;
  while (fs.existsSync(path.join(parentDir, candidate)) && candidate !== excludeName) {
    candidate = `${baseName}-${counter}.md`;
    counter += 1;
  }
  return candidate;
};

const writeChapterFile = (meta, chapter) => {
  const parentDir = getChapterParentDir(meta, chapter);
  ensureDir(parentDir);
  const filePath = path.join(parentDir, chapter.file_name);
  fs.writeFileSync(filePath, chapter.content || '', 'utf8');
};

const removeFileIfExists = filePath => {
  if (fs.existsSync(filePath)) {
    fs.unlinkSync(filePath);
  }
};

const removeDirIfEmpty = dirPath => {
  if (!fs.existsSync(dirPath)) return;
  const entries = fs.readdirSync(dirPath).filter(name => name !== HIDDEN_DIR_NAME);
  if (entries.length === 0) {
    fs.rmdirSync(dirPath);
  }
};

const touchBookUpdatedAt = (bookId, updatedAt) => {
  db.prepare('UPDATE books SET updated_at = ? WHERE id = ?').run(updatedAt, bookId);
};

const syncBookMetaFromDb = dbBook => {
  const parsedTags = (() => {
    if (Array.isArray(dbBook.tags)) return dbBook.tags;
    if (!dbBook.tags) return [];
    try {
      return JSON.parse(dbBook.tags);
    } catch (error) {
      return [];
    }
  })();

  return toBookRecord({
    ...dbBook,
    tags: parsedTags,
    is_public: dbBook.is_public === 1 || dbBook.is_public === true
  });
};

const migrateBookFromDatabase = (index, dbBook) => {
  const existingDir = getBookDirById(index, dbBook.id);
  if (existingDir && fs.existsSync(getMetaPath(existingDir))) {
    return;
  }

  const directoryName = existingDir
    ? path.basename(existingDir)
    : getUniqueDirectoryName(getLibraryRoot(), dbBook.title || `作品-${dbBook.id}`);
  const bookDir = path.join(getLibraryRoot(), directoryName);
  ensureDir(bookDir);

  const book = syncBookMetaFromDb(dbBook);
  const dbVolumes = db.prepare('SELECT * FROM volumes WHERE book_id = ? ORDER BY order_num ASC, id ASC').all(dbBook.id);
  const dbChapters = db.prepare('SELECT * FROM chapters WHERE book_id = ? ORDER BY order_num ASC, id ASC').all(dbBook.id);

  const meta = {
    book: {
      ...book,
      base_path: bookDir
    },
    volumes: [],
    chapters: []
  };

  for (const rawVolume of dbVolumes) {
    const volume = {
      id: Number(rawVolume.id),
      book_id: Number(rawVolume.book_id),
      title: rawVolume.title || '新分卷',
      order_num: Number(rawVolume.order_num) || 0,
      parent_id: rawVolume.parent_id ? Number(rawVolume.parent_id) : null,
      created_at: rawVolume.created_at || formatTimestamp(),
      updated_at: rawVolume.updated_at || formatTimestamp(),
      folder_name: ''
    };
    volume.folder_name = getVolumeFolderName(meta, volume);
    ensureDir(path.join(bookDir, volume.folder_name));
    meta.volumes.push(volume);
    index.nextVolumeId = Math.max(index.nextVolumeId, volume.id + 1);
  }

  for (const rawChapter of dbChapters) {
    const chapter = {
      id: Number(rawChapter.id),
      book_id: Number(rawChapter.book_id),
      title: rawChapter.title || '新章节',
      content: rawChapter.content || '',
      summary: rawChapter.summary || '',
      order_num: Number(rawChapter.order_num) || 0,
      type: rawChapter.type || 'chapter',
      volume_id: rawChapter.volume_id ? Number(rawChapter.volume_id) : null,
      created_at: rawChapter.created_at || formatTimestamp(),
      updated_at: rawChapter.updated_at || formatTimestamp(),
      file_name: ''
    };
    chapter.file_name = getUniqueChapterFileName(meta, chapter);
    meta.chapters.push(chapter);
    writeChapterFile(meta, chapter);
    index.nextChapterId = Math.max(index.nextChapterId, chapter.id + 1);
  }

  index.books[String(dbBook.id)] = directoryName;
  writeBookMeta(meta);
};

const ensureMigrated = () => {
  const index = readIndex();
  const books = db.prepare('SELECT * FROM books ORDER BY id ASC').all();
  books.forEach(book => migrateBookFromDatabase(index, book));
  writeIndex(index);
  return index;
};

const loadBookMetaById = bookId => {
  const index = ensureMigrated();
  const bookDir = getBookDirById(index, bookId);
  if (!bookDir) return null;
  return readBookMeta(bookDir);
};

const saveBookMeta = meta => {
  writeBookMeta(meta);
  return meta;
};

const formatBookForApi = meta => ({
  ...meta.book,
  base_path: undefined
});

const formatVolumeForApi = volume => ({ ...volume });

const formatChapterForApi = chapter => ({ ...chapter });

const loadAllBooks = () => {
  const index = ensureMigrated();
  const metas = Object.keys(index.books)
    .map(bookId => readBookMeta(getBookDirById(index, bookId)))
    .filter(Boolean);

  metas.sort((a, b) => new Date(b.book.updated_at).getTime() - new Date(a.book.updated_at).getTime());
  return metas.map(formatBookForApi);
};

const createBook = payload => {
  ensureMigrated();
  const now = formatTimestamp();
  const result = db.prepare(`
    INSERT INTO books (title, description, cover, author, category, tags, status, is_public, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `).run(
    payload.title,
    payload.description || '',
    payload.cover || '',
    payload.author || '',
    payload.category || '',
    JSON.stringify(Array.isArray(payload.tags) ? payload.tags : []),
    payload.status || 'draft',
    payload.is_public ? 1 : 0,
    now,
    now
  );

  const bookId = Number(result.lastInsertRowid);
  const index = readIndex();
  const directoryName = getUniqueDirectoryName(getLibraryRoot(), payload.title || `作品-${bookId}`);
  const bookDir = path.join(getLibraryRoot(), directoryName);
  ensureDir(bookDir);

  const meta = {
    book: {
      id: bookId,
      title: payload.title,
      description: payload.description || '',
      cover: payload.cover || '',
      author: payload.author || '',
      category: payload.category || '',
      tags: Array.isArray(payload.tags) ? payload.tags : [],
      status: payload.status || 'draft',
      is_public: Boolean(payload.is_public),
      created_at: now,
      updated_at: now,
      base_path: bookDir
    },
    volumes: [],
    chapters: []
  };

  index.books[String(bookId)] = directoryName;
  writeIndex(index);
  saveBookMeta(meta);
  return formatBookForApi(meta);
};

const updateBook = (bookId, payload) => {
  const meta = loadBookMetaById(bookId);
  if (!meta) return null;

  const now = formatTimestamp();
  const nextBook = {
    ...meta.book,
    title: payload.title ?? meta.book.title,
    description: payload.description ?? meta.book.description,
    cover: payload.cover ?? meta.book.cover,
    author: payload.author ?? meta.book.author,
    category: payload.category ?? meta.book.category,
    tags: payload.tags != null ? payload.tags : meta.book.tags,
    status: payload.status ?? meta.book.status,
    is_public: payload.is_public != null ? Boolean(payload.is_public) : meta.book.is_public,
    updated_at: now
  };

  const index = readIndex();
  const oldDirectoryName = index.books[String(bookId)];
  const desiredDirectoryName = getUniqueDirectoryName(getLibraryRoot(), nextBook.title || `作品-${bookId}`, oldDirectoryName);
  if (desiredDirectoryName !== oldDirectoryName) {
    const oldPath = meta.book.base_path;
    const nextPath = path.join(getLibraryRoot(), desiredDirectoryName);
    fs.renameSync(oldPath, nextPath);
    index.books[String(bookId)] = desiredDirectoryName;
    writeIndex(index);
    meta.book.base_path = nextPath;
  }

  meta.book = {
    ...nextBook,
    base_path: meta.book.base_path
  };

  db.prepare(`
    UPDATE books SET
      title = ?, description = ?, cover = ?, author = ?, category = ?, tags = ?, status = ?, is_public = ?, updated_at = ?
    WHERE id = ?
  `).run(
    meta.book.title,
    meta.book.description,
    meta.book.cover,
    meta.book.author,
    meta.book.category,
    JSON.stringify(meta.book.tags || []),
    meta.book.status,
    meta.book.is_public ? 1 : 0,
    now,
    bookId
  );

  saveBookMeta(meta);
  return formatBookForApi(meta);
};

const deleteBook = bookId => {
  const meta = loadBookMetaById(bookId);
  const index = readIndex();
  if (meta && fs.existsSync(meta.book.base_path)) {
    fs.rmSync(meta.book.base_path, { recursive: true, force: true });
  }
  delete index.books[String(bookId)];
  writeIndex(index);
  db.prepare('DELETE FROM books WHERE id = ?').run(bookId);
  return true;
};

const listVolumes = bookId => {
  const meta = loadBookMetaById(bookId);
  if (!meta) return null;
  return meta.volumes
    .slice()
    .sort((a, b) => a.order_num - b.order_num || a.id - b.id)
    .map(formatVolumeForApi);
};

const getVolume = volumeId => {
  const index = ensureMigrated();
  for (const bookId of Object.keys(index.books)) {
    const meta = readBookMeta(getBookDirById(index, bookId));
    if (!meta) continue;
    const volume = meta.volumes.find(item => Number(item.id) === Number(volumeId));
    if (volume) {
      return formatVolumeForApi(volume);
    }
  }
  return null;
};

const createVolume = payload => {
  const meta = loadBookMetaById(payload.book_id);
  if (!meta) return null;

  const index = readIndex();
  const now = formatTimestamp();
  const volumeId = index.nextVolumeId;
  index.nextVolumeId += 1;

  const volume = {
    id: volumeId,
    book_id: Number(payload.book_id),
    title: payload.title || '新分卷',
    order_num: Number(payload.order_num) || 0,
    parent_id: payload.parent_id ? Number(payload.parent_id) : null,
    created_at: now,
    updated_at: now,
    folder_name: ''
  };
  volume.folder_name = getVolumeFolderName(meta, volume);
  ensureDir(getVolumeDir(meta, volume));

  meta.volumes.push(volume);
  meta.book.updated_at = now;
  writeIndex(index);
  saveBookMeta(meta);

  db.prepare(`
    INSERT INTO volumes (id, book_id, parent_id, title, order_num, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `).run(volume.id, volume.book_id, volume.parent_id, volume.title, volume.order_num, now, now);
  touchBookUpdatedAt(payload.book_id, now);

  return formatVolumeForApi(volume);
};

const updateVolume = (volumeId, payload) => {
  const index = ensureMigrated();
  for (const bookId of Object.keys(index.books)) {
    const meta = readBookMeta(getBookDirById(index, bookId));
    if (!meta) continue;
    const volume = meta.volumes.find(item => Number(item.id) === Number(volumeId));
    if (!volume) continue;

    const oldFolderName = volume.folder_name;
    const oldFolderPath = getVolumeDir(meta, volume);
    const requestedParentId = payload.parent_id === undefined
      ? volume.parent_id
      : (payload.parent_id ? Number(payload.parent_id) : null);
    const descendantIds = getVolumeDescendantIds(meta, volume.id);
    if (requestedParentId && descendantIds.has(requestedParentId)) {
      throw new Error('不能将文件夹移动到自己的子文件夹中');
    }

    volume.title = payload.title ?? volume.title;
    volume.order_num = payload.order_num ?? volume.order_num;
    volume.parent_id = requestedParentId;
    volume.updated_at = formatTimestamp();
    volume.folder_name = getVolumeFolderName(meta, volume, oldFolderName);

    const nextFolderPath = getVolumeDir(meta, volume);
    if (oldFolderPath !== nextFolderPath && fs.existsSync(oldFolderPath)) {
      ensureDir(path.dirname(nextFolderPath));
      fs.renameSync(oldFolderPath, nextFolderPath);
    }

    meta.chapters
      .filter(chapter => Number(chapter.volume_id) === Number(volumeId))
      .forEach(chapter => {
        chapter.file_name = getUniqueChapterFileName(meta, chapter, chapter.file_name);
      });

    meta.book.updated_at = volume.updated_at;
    saveBookMeta(meta);
    db.prepare('UPDATE volumes SET parent_id = ?, title = ?, order_num = ?, updated_at = ? WHERE id = ?').run(volume.parent_id, volume.title, volume.order_num, volume.updated_at, volume.id);
    touchBookUpdatedAt(volume.book_id, volume.updated_at);
    return formatVolumeForApi(volume);
  }
  return null;
};

const deleteVolume = volumeId => {
  const index = ensureMigrated();
  for (const bookId of Object.keys(index.books)) {
    const meta = readBookMeta(getBookDirById(index, bookId));
    if (!meta) continue;
    const volumeIndex = meta.volumes.findIndex(item => Number(item.id) === Number(volumeId));
    if (volumeIndex === -1) continue;

    const volume = meta.volumes[volumeIndex];
    const hasChildren = meta.volumes.some(item => Number(item.parent_id) === Number(volume.id));
    if (hasChildren) {
      throw new Error('请先删除或移动子文件夹');
    }
    const now = formatTimestamp();
    meta.chapters
      .filter(chapter => Number(chapter.volume_id) === Number(volume.id))
      .forEach(chapter => {
        const oldPath = path.join(getVolumeDir(meta, volume), chapter.file_name);
        chapter.volume_id = null;
        chapter.updated_at = now;
        const nextFileName = getUniqueChapterFileName(meta, chapter, chapter.file_name);
        const nextPath = path.join(meta.book.base_path, nextFileName);
        fs.renameSync(oldPath, nextPath);
        chapter.file_name = nextFileName;
      });

    meta.volumes.splice(volumeIndex, 1);
    meta.book.updated_at = now;
    saveBookMeta(meta);
    removeDirIfEmpty(path.join(meta.book.base_path, volume.folder_name));

    db.prepare('UPDATE chapters SET volume_id = NULL, updated_at = ? WHERE volume_id = ?').run(now, volume.id);
    db.prepare('DELETE FROM volumes WHERE id = ?').run(volume.id);
    touchBookUpdatedAt(volume.book_id, now);
    return true;
  }
  return false;
};

const listChapters = bookId => {
  const meta = loadBookMetaById(bookId);
  if (!meta) return null;
  return meta.chapters
    .slice()
    .sort((a, b) => a.order_num - b.order_num || a.id - b.id)
    .map(formatChapterForApi);
};

const getChapterMeta = chapterId => {
  const index = ensureMigrated();
  for (const bookId of Object.keys(index.books)) {
    const meta = readBookMeta(getBookDirById(index, bookId));
    if (!meta) continue;
    const chapter = meta.chapters.find(item => Number(item.id) === Number(chapterId));
    if (chapter) {
      return { meta, chapter };
    }
  }
  return null;
};

const createChapter = payload => {
  const meta = loadBookMetaById(payload.book_id);
  if (!meta) return null;

  const index = readIndex();
  const now = formatTimestamp();
  const chapterId = index.nextChapterId;
  index.nextChapterId += 1;

  const chapter = {
    id: chapterId,
    book_id: Number(payload.book_id),
    title: payload.title || '新章节',
    content: payload.content || '',
    summary: payload.summary || '',
    order_num: Number(payload.order_num) || 0,
      type: payload.type || 'chapter',
      volume_id: payload.volume_id ? Number(payload.volume_id) : null,
    created_at: now,
    updated_at: now,
    file_name: ''
  };

  chapter.file_name = getUniqueChapterFileName(meta, chapter);
  meta.chapters.push(chapter);
  meta.book.updated_at = now;
  writeChapterFile(meta, chapter);
  writeIndex(index);
  saveBookMeta(meta);

  db.prepare(`
    INSERT INTO chapters (id, book_id, title, content, summary, order_num, type, volume_id, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `).run(
    chapter.id,
    chapter.book_id,
    chapter.title,
    chapter.content,
    chapter.summary,
    chapter.order_num,
    chapter.type,
    chapter.volume_id,
    now,
    now
  );
  touchBookUpdatedAt(payload.book_id, now);

  return formatChapterForApi(chapter);
};

const updateChapter = (chapterId, payload) => {
  const loaded = getChapterMeta(chapterId);
  if (!loaded) return null;

  const { meta, chapter } = loaded;
  const previousPath = path.join(getChapterParentDir(meta, chapter), chapter.file_name);
  const now = formatTimestamp();

  chapter.title = payload.title ?? chapter.title;
  chapter.content = payload.content ?? chapter.content;
  chapter.summary = payload.summary ?? chapter.summary;
  chapter.order_num = payload.order_num ?? chapter.order_num;
  chapter.type = payload.type ?? chapter.type;
  chapter.volume_id = payload.volume_id === undefined
    ? chapter.volume_id
    : (payload.volume_id ? Number(payload.volume_id) : null);
  chapter.updated_at = now;

  const nextParentDir = getChapterParentDir(meta, chapter);
  ensureDir(nextParentDir);
  const nextFileName = getUniqueChapterFileName(meta, chapter, chapter.file_name);
  const nextPath = path.join(nextParentDir, nextFileName);
  if (previousPath !== nextPath && fs.existsSync(previousPath)) {
    fs.renameSync(previousPath, nextPath);
  }
  chapter.file_name = nextFileName;
  writeChapterFile(meta, chapter);

  meta.book.updated_at = now;
  saveBookMeta(meta);
  db.prepare(`
    UPDATE chapters
    SET title = ?, content = ?, summary = ?, order_num = ?, type = ?, volume_id = ?, updated_at = ?
    WHERE id = ?
  `).run(chapter.title, chapter.content, chapter.summary, chapter.order_num, chapter.type, chapter.volume_id, now, chapter.id);
  touchBookUpdatedAt(chapter.book_id, now);

  return formatChapterForApi(chapter);
};

const deleteChapter = chapterId => {
  const loaded = getChapterMeta(chapterId);
  if (!loaded) return false;

  const { meta, chapter } = loaded;
  const chapterPath = path.join(getChapterParentDir(meta, chapter), chapter.file_name);
  removeFileIfExists(chapterPath);
  meta.chapters = meta.chapters.filter(item => Number(item.id) !== Number(chapterId));
  meta.book.updated_at = formatTimestamp();
  saveBookMeta(meta);
  db.prepare('DELETE FROM chapters WHERE id = ?').run(chapterId);
  touchBookUpdatedAt(chapter.book_id, meta.book.updated_at);

  if (chapter.volume_id) {
    const volume = getVolumeById(meta, chapter.volume_id);
    if (volume) {
      removeDirIfEmpty(getVolumeDir(meta, volume));
    }
  }
  return true;
};

const importChapters = (bookId, chapters) => {
  const inserted = [];
  const currentChapters = listChapters(bookId) || [];
  let orderBase = currentChapters.reduce((max, chapter) => Math.max(max, Number(chapter.order_num) || 0), -1) + 1;

  chapters.forEach(rawChapter => {
    const content = String(rawChapter?.content || '').trim();
    if (!content) return;
    const created = createChapter({
      book_id: Number(bookId),
      title: String(rawChapter?.title || '').trim() || `第${orderBase + 1}章`,
      content,
      order_num: orderBase,
      type: 'chapter'
    });
    if (created) {
      inserted.push(created);
      orderBase += 1;
    }
  });

  return inserted;
};

const getChaptersForGraph = (bookId, chapterIds = []) => {
  const chapters = listChapters(bookId) || [];
  if (!Array.isArray(chapterIds) || chapterIds.length === 0) {
    return chapters;
  }
  const chapterIdSet = new Set(chapterIds.map(id => Number(id)));
  return chapters.filter(chapter => chapterIdSet.has(Number(chapter.id)));
};

module.exports = {
  loadAllBooks,
  loadBookMetaById,
  createBook,
  updateBook,
  deleteBook,
  listVolumes,
  getVolume,
  createVolume,
  updateVolume,
  deleteVolume,
  listChapters,
  getChapterMeta,
  createChapter,
  updateChapter,
  deleteChapter,
  importChapters,
  getChaptersForGraph
};
