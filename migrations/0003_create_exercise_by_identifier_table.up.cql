CREATE TABLE exercise_by_identifier (
	id uuid,
	active_until timestamp,
	exercise_description text,
	exercise_name text,
	has_activation boolean,
	is_active boolean,
	PRIMARY KEY (id)
);
