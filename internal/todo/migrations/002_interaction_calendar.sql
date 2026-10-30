CREATE TABLE task_expansion(task_id TEXT PRIMARY KEY NOT NULL REFERENCES tasks(id) ON DELETE CASCADE, expanded INTEGER NOT NULL CHECK(expanded IN (0,1)));
CREATE TABLE local_calendar_config(id INTEGER PRIMARY KEY CHECK(id=1),sources_json TEXT NOT NULL DEFAULT '[]',last_sync INTEGER NOT NULL DEFAULT 0);
INSERT INTO local_calendar_config(id) VALUES(1);
CREATE TABLE local_calendar_events(id TEXT PRIMARY KEY NOT NULL,calendar_id TEXT NOT NULL,start_at INTEGER NOT NULL,end_at INTEGER NOT NULL,event_json TEXT NOT NULL);
CREATE INDEX idx_local_calendar_range ON local_calendar_events(start_at,end_at,calendar_id);
