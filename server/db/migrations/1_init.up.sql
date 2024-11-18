CREATE TABLE users
(
  id UUID NOT NULL,
  email VARCHAR NOT NULL,
  password VARCHAR NOT NULL,
  first_name VARCHAR(50),
  last_name VARCHAR(50),
  status VARCHAR(50) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
  
  CONSTRAINT pk_user_id PRIMARY KEY(id)
);

CREATE TABLE roles
(
  id UUID NOT NULL,
  name VARCHAR(100) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  updated_by VARCHAR,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT pk_role_id PRIMARY KEY(id)
);

CREATE TABLE users_roles
(
  user_id UUID,
  role_id UUID,

  CONSTRAINT pk_user_role_id PRIMARY KEY(user_id, role_id)
);

CREATE TABLE permissions
(
  id UUID NOT NULL,
  name VARCHAR(200) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  updated_by VARCHAR,
  is_deleted BOOLEAN DEFAULT FALSE,

  CONSTRAINT pk_permission_id PRIMARY KEY(id)
);

CREATE TABLE roles_permissions
(
  role_id UUID,
  permission_id UUID,

  CONSTRAINT pk_role_permission_id PRIMARY KEY(role_id, permission_id)
);

CREATE TABLE categories
(
  id UUID NOT NULL,
  name VARCHAR(100) NOT NULL,
  image_url VARCHAR NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT pk_category_id PRIMARY KEY(id)
);

CREATE TABLE posts
(
  id UUID NOT NULL,
  title VARCHAR(100) NOT NULL,
  content VARCHAR NOT NULL,
  category_id UUID NOT NULL,
  creator_user_id UUID NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  updated_by VARCHAR,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT pk_post_id PRIMARY KEY(id),
  CONSTRAINT fk_category_id FOREIGN KEY(category_id) REFERENCES categories(id)
);

CREATE TABLE images
(
  id UUID NOT NULL,
  url VARCHAR NOT NULL,
  owner_id UUID,
  owner_type VARCHAR,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  updated_by VARCHAR,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT pk_image_id PRIMARY KEY(id)
);
