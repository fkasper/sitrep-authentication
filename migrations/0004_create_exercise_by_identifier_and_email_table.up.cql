CREATE TABLE exercise_by_identifier_and_email (
	id uuid,
	email text,
	is_authorized boolean,
	is_operator boolean,
  is_active boolean,
  exercise_name varchar,
  user_name varchar,
	PRIMARY KEY (id)
);
