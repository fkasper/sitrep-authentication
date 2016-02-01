CREATE TABLE exercise_permissions_level(
  user_email varchar,
  exercise_identifier uuid,
  is_oc boolean,
  is_admin boolean,
  is_trainee boolean,
  is_authorized boolean,
  is_invisible boolean,
  role_description text,
  PRIMARY KEY (user_email, exercise_identifier)
);
