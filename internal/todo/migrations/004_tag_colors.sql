ALTER TABLE tags ADD COLUMN color TEXT NOT NULL DEFAULT '#64748b';
UPDATE tags SET color='#b93838' WHERE id='important';
UPDATE tags SET color='#b45309' WHERE id='urgent';
