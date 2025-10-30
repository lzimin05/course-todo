/*
// Use DBML to define your database structure
// Docs: https://dbml.dbdiagram.io/docs

Table user {
  id integer [primary key]
  login varchar
  username varchar
  email varchar
  password_hash varchar
}

Table task {
  id integer [primary key]
  user_id integer
  title varchar
  description varchar
  importance integer
  status varchar //'waiting', 'in_progress', 'completed'
  created_at data
  deadline data
}

Ref user: user.id < task.user_id 
*/

CREATE SCHEMA IF NOT EXISTS todo;

CREATE TABLE IF NOT EXISTS todo."user" (
  id UUID PRIMARY KEY,
  login VARCHAR NOT NULL UNIQUE,
  username VARCHAR NOT NULL,
  email VARCHAR NOT NULL UNIQUE,
  password_hash VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS todo."task" (
  id UUID PRIMARY KEY,
  user_id UUID,
  title VARCHAR NOT NULL,
  description VARCHAR,
  importance INTEGER,
  status VARCHAR NOT NULL, -- 'waiting', 'in_progress', 'completed'
  created_at TIMESTAMP,
  deadline TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES todo."user"(id)
);

CREATE TABLE IF NOT EXISTS todo.note (
  id UUID PRIMARY KEY,
  user_id UUID,
  name VARCHAR NOT NULL,
  description VARCHAR,
  created_at TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES todo."user"(id)
)