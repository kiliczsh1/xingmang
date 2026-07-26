CREATE TABLE IF NOT EXISTS experience_shares (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  summary TEXT,
  content TEXT NOT NULL,
  cover_url TEXT,
  pdf_file_url TEXT,
  pdf_file_name TEXT,
  pdf_file_size INTEGER,
  create_type TEXT DEFAULT 'manual',
  author_id INTEGER,
  author_name TEXT DEFAULT '星芒用户',
  status TEXT DEFAULT 'published',
  pdf_parse_status TEXT,
  pdf_parse_result TEXT,
  source_file_name TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE experience_shares ADD COLUMN cover_url TEXT;
ALTER TABLE experience_shares ADD COLUMN pdf_file_url TEXT;
ALTER TABLE experience_shares ADD COLUMN pdf_file_name TEXT;
ALTER TABLE experience_shares ADD COLUMN pdf_file_size INTEGER;
ALTER TABLE experience_shares ADD COLUMN create_type TEXT DEFAULT 'manual';
ALTER TABLE experience_shares ADD COLUMN author_id INTEGER;
ALTER TABLE experience_shares ADD COLUMN author_name TEXT DEFAULT '星芒用户';
ALTER TABLE experience_shares ADD COLUMN status TEXT DEFAULT 'published';
ALTER TABLE experience_shares ADD COLUMN pdf_parse_status TEXT;
ALTER TABLE experience_shares ADD COLUMN pdf_parse_result TEXT;
ALTER TABLE experience_shares ADD COLUMN source_file_name TEXT;
