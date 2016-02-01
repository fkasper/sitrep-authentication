CREATE TABLE users_by_jwt (
	jwt text,
	encryption_key text,
	user_email text,
	user_name text,
	PRIMARY KEY (jwt)
);
