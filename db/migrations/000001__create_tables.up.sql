/*
// Use DBML to define your database structure
// Docs: https://dbml.dbdiagram.io/docs

Table user {
  id uuid [primary key]
  login varchar
  username varchar
  email varchar
  password_hash varchar
}

Table project {
  id uuid [primary key]
  name varchar
  description varchar
  owner_id uuid
  created_at timestamp
}

Table project_member {
  id uuid [primary key]
  project_id uuid
  user_id uuid
  role varchar // 'owner', 'member'
  joined_at timestamp
}

Table task {
  id uuid [primary key]
  project_id uuid
  user_id uuid
  title varchar
  description varchar
  importance integer
  status varchar //'waiting', 'in_progress', 'completed'
  created_at timestamp
  deadline timestamp
}

Table note {
  id uuid [primary key]
  project_id uuid
  user_id uuid
  name varchar
  description varchar
  created_at timestamp
}

Ref: user.id < project.owner_id
Ref: project.id < project_member.project_id
Ref: user.id < project_member.user_id
Ref: project.id < task.project_id
Ref: user.id < task.user_id
Ref: project.id < note.project_id
Ref: user.id < note.user_id
*/

CREATE SCHEMA IF NOT EXISTS todo;

CREATE TABLE IF NOT EXISTS todo."user" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  login VARCHAR NOT NULL UNIQUE,
  username VARCHAR NOT NULL,
  email VARCHAR NOT NULL UNIQUE,
  password_hash VARCHAR NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS todo.project (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR NOT NULL,
  description VARCHAR,
  owner_id UUID NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (owner_id) REFERENCES todo."user"(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS todo.project_member (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL,
  user_id UUID NOT NULL,
  role VARCHAR NOT NULL CHECK (role IN ('owner', 'member')),
  joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (project_id) REFERENCES todo.project(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES todo."user"(id) ON DELETE CASCADE,
  UNIQUE(project_id, user_id)
);

CREATE TABLE IF NOT EXISTS todo."task" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL,
  user_id UUID NOT NULL,
  title VARCHAR NOT NULL,
  description VARCHAR,
  importance INTEGER DEFAULT 1 CHECK (importance >= 1 AND importance <= 5),
  status VARCHAR NOT NULL DEFAULT 'waiting' CHECK (status IN ('waiting', 'in_progress', 'completed')),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deadline TIMESTAMP,
  FOREIGN KEY (project_id) REFERENCES todo.project(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES todo."user"(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS todo.note (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL,
  user_id UUID NOT NULL,
  name VARCHAR NOT NULL,
  description VARCHAR,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (project_id) REFERENCES todo.project(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES todo."user"(id) ON DELETE CASCADE
);

-- Индексы для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_project_owner ON todo.project(owner_id);
CREATE INDEX IF NOT EXISTS idx_project_member_project ON todo.project_member(project_id);
CREATE INDEX IF NOT EXISTS idx_project_member_user ON todo.project_member(user_id);
CREATE INDEX IF NOT EXISTS idx_task_project ON todo."task"(project_id);
CREATE INDEX IF NOT EXISTS idx_task_user ON todo."task"(user_id);
CREATE INDEX IF NOT EXISTS idx_task_status ON todo."task"(status);
CREATE INDEX IF NOT EXISTS idx_note_project ON todo.note(project_id);
CREATE INDEX IF NOT EXISTS idx_note_user ON todo.note(user_id);