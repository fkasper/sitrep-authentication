CREATE TABLE settings_by_exercise_identifier (
	id uuid,
	is_default boolean,
	settings MAP<text, text>,
	settings_version text,
	PRIMARY KEY (id)
);
