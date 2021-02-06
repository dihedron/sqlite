CREATE TABLE IF NOT EXISTS pairs (
	key         TEXT PRIMARY KEY,
	value       TEXT NOT NULL,
	created_at  TEXT NOT NULL,
	updated_at  TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS pairs_value ON pairs(value);

INSERT INTO pairs (key,value,created_at,updated_at)
VALUES
	("key01","value01",datetime('now'),datetime('now')),
	("key02","value02",datetime('now'),datetime('now')),
	("key03","value03",datetime('now'),datetime('now')),
	("key04","value04",datetime('now'),datetime('now')),
	("key05","value05",datetime('now'),datetime('now')),
	("key06","value06",datetime('now'),datetime('now')),
	("key07","value07",datetime('now'),datetime('now')),
	("key08","value08",datetime('now'),datetime('now'));
