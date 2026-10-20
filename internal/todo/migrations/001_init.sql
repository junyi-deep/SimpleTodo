CREATE TABLE schema_migrations(version INTEGER PRIMARY KEY, checksum TEXT NOT NULL, applied_at INTEGER NOT NULL);
CREATE TABLE app_meta(key TEXT PRIMARY KEY NOT NULL,value TEXT NOT NULL);
INSERT INTO app_meta VALUES('revision','0');
CREATE TABLE repeat_series(id TEXT PRIMARY KEY NOT NULL,stopped_from TEXT,version INTEGER NOT NULL DEFAULT 1);
CREATE TABLE repeat_rules(id TEXT PRIMARY KEY NOT NULL,series_id TEXT NOT NULL REFERENCES repeat_series(id),effective_from TEXT NOT NULL,effective_to TEXT,anchor_date TEXT NOT NULL,frequency TEXT NOT NULL CHECK(frequency IN ('daily','weekly','monthly')),interval_n INTEGER NOT NULL CHECK(interval_n BETWEEN 1 AND 99),weekdays TEXT NOT NULL,month_day INTEGER NOT NULL,template_tree_json TEXT NOT NULL,generated_through TEXT NOT NULL,version INTEGER NOT NULL DEFAULT 1);
CREATE TABLE tasks(
 id TEXT PRIMARY KEY NOT NULL,parent_id TEXT,root_id TEXT NOT NULL,title TEXT NOT NULL CHECK(length(trim(title)) BETWEEN 1 AND 500),description TEXT CHECK(length(description)<=20000),planned_date TEXT,original_planned_date TEXT,status TEXT NOT NULL CHECK(status IN ('pending','completed','skipped')),color TEXT,sort_order INTEGER NOT NULL,completed_at INTEGER,auto_postpone_policy TEXT,series_id TEXT REFERENCES repeat_series(id),rule_id TEXT REFERENCES repeat_rules(id),occurrence_date TEXT,is_exception INTEGER NOT NULL DEFAULT 0,remind_at INTEGER,reminder_local_time TEXT,reminder_timezone TEXT,reminder_generation INTEGER NOT NULL DEFAULT 0,version INTEGER NOT NULL,created_at INTEGER NOT NULL,updated_at INTEGER NOT NULL,
 FOREIGN KEY(root_id) REFERENCES tasks(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
 FOREIGN KEY(root_id,parent_id) REFERENCES tasks(root_id,id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
 CHECK((parent_id IS NULL AND root_id=id) OR (parent_id IS NOT NULL AND root_id<>id)),
 CHECK(parent_id IS NULL OR parent_id<>id),
 CHECK((status='completed' AND completed_at IS NOT NULL) OR (status<>'completed' AND completed_at IS NULL)),
 CHECK((series_id IS NULL AND rule_id IS NULL AND occurrence_date IS NULL) OR (parent_id IS NULL AND series_id IS NOT NULL AND rule_id IS NOT NULL AND occurrence_date IS NOT NULL)),
 CHECK(status<>'skipped' OR (parent_id IS NULL AND series_id IS NOT NULL)),
 CHECK(parent_id IS NULL OR (description IS NULL AND planned_date IS NULL AND original_planned_date IS NULL AND color IS NULL AND auto_postpone_policy IS NULL AND series_id IS NULL AND rule_id IS NULL AND occurrence_date IS NULL AND remind_at IS NULL AND reminder_local_time IS NULL AND reminder_timezone IS NULL AND is_exception=0 AND reminder_generation=0))
);
CREATE UNIQUE INDEX ux_tasks_root_node ON tasks(root_id,id);
CREATE INDEX idx_tasks_children_order ON tasks(parent_id,sort_order,id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_tasks_root_schedule ON tasks(planned_date,status DESC,sort_order,id) WHERE parent_id IS NULL AND status<>'skipped';
CREATE INDEX idx_tasks_root_pending ON tasks(planned_date,sort_order,id) WHERE parent_id IS NULL AND status='pending';
CREATE UNIQUE INDEX ux_tasks_series_occurrence ON tasks(series_id,occurrence_date) WHERE parent_id IS NULL AND series_id IS NOT NULL;
CREATE INDEX idx_tasks_rule ON tasks(rule_id,occurrence_date,id) WHERE rule_id IS NOT NULL;
CREATE UNIQUE INDEX ux_repeat_rules_series_from ON repeat_rules(series_id,effective_from);
CREATE TABLE focus_sessions(id TEXT PRIMARY KEY NOT NULL,task_id TEXT REFERENCES tasks(id) ON DELETE SET NULL,task_title_snapshot TEXT NOT NULL,duration_seconds INTEGER NOT NULL,remaining_seconds INTEGER NOT NULL,deadline_at INTEGER NOT NULL,state TEXT NOT NULL CHECK(state IN ('running','paused','completed','stopped')),started_at INTEGER NOT NULL,finished_at INTEGER);
CREATE UNIQUE INDEX ux_focus_one_active ON focus_sessions((1)) WHERE state IN ('running','paused');
CREATE INDEX idx_focus_task ON focus_sessions(task_id) WHERE task_id IS NOT NULL;
CREATE TABLE countdowns(id TEXT PRIMARY KEY NOT NULL,name TEXT NOT NULL,all_day INTEGER NOT NULL,anchor_local_date TEXT NOT NULL,anchor_local_time TEXT NOT NULL,target_at INTEGER NOT NULL,timezone TEXT NOT NULL,repeat_frequency TEXT NOT NULL,enabled INTEGER NOT NULL,reminder_offset_seconds INTEGER NOT NULL,generation INTEGER NOT NULL);
CREATE INDEX idx_countdowns_enabled_target ON countdowns(target_at,id) WHERE enabled=1;
CREATE TABLE notification_jobs(id TEXT PRIMARY KEY NOT NULL,source_type TEXT NOT NULL,source_id TEXT NOT NULL,generation INTEGER NOT NULL,cycle_key TEXT NOT NULL,fire_at INTEGER NOT NULL,state TEXT NOT NULL,claimed_at INTEGER,submitted_at INTEGER,error_code TEXT,title TEXT NOT NULL);
CREATE UNIQUE INDEX ux_notification_source_cycle ON notification_jobs(source_type,source_id,generation,cycle_key);
CREATE INDEX idx_notification_state_fire ON notification_jobs(state,fire_at,id);
CREATE TABLE settings(key TEXT PRIMARY KEY NOT NULL,value_json TEXT NOT NULL,version INTEGER NOT NULL DEFAULT 1);
CREATE TABLE window_states(window_key TEXT PRIMARY KEY NOT NULL,state_json TEXT NOT NULL,version INTEGER NOT NULL DEFAULT 1);
CREATE TABLE holiday_days(date TEXT PRIMARY KEY NOT NULL,name TEXT NOT NULL,kind TEXT NOT NULL,dataset_version TEXT NOT NULL,source_label TEXT NOT NULL);
CREATE TABLE command_receipts(command_id TEXT PRIMARY KEY NOT NULL,request_hash TEXT NOT NULL,result_json TEXT NOT NULL,created_at INTEGER NOT NULL);
CREATE INDEX idx_command_receipts_created ON command_receipts(created_at);
