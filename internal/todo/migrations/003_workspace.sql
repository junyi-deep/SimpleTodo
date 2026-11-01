ALTER TABLE tasks ADD COLUMN starts_at INTEGER NOT NULL DEFAULT 0;
ALTER TABLE tasks ADD COLUMN ends_at INTEGER;
ALTER TABLE tasks ADD COLUMN tags_json TEXT NOT NULL DEFAULT '[]';
UPDATE tasks SET starts_at=created_at;
CREATE INDEX idx_tasks_times ON tasks(starts_at,ends_at,status);
CREATE TABLE tags(id TEXT PRIMARY KEY NOT NULL,name TEXT NOT NULL UNIQUE,sort_order INTEGER NOT NULL);
INSERT INTO tags VALUES('important','重要',0),('urgent','紧急',1);
CREATE TABLE notes(id TEXT PRIMARY KEY NOT NULL,title TEXT NOT NULL,content TEXT NOT NULL DEFAULT '',created_at INTEGER NOT NULL,updated_at INTEGER NOT NULL,version INTEGER NOT NULL DEFAULT 1);
CREATE INDEX idx_notes_updated ON notes(updated_at DESC,id);
