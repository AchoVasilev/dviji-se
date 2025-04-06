CREATE TABLE users
(
  id UUID NOT NULL,
  email VARCHAR NOT NULL UNIQUE,
  password VARCHAR NOT NULL,
  first_name VARCHAR(50),
  last_name VARCHAR(50),
  status VARCHAR(50) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
  
  CONSTRAINT pk_users_id PRIMARY KEY(id)
);

CREATE TABLE roles
(
  id UUID NOT NULL,
  name VARCHAR(100) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  updated_by VARCHAR,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT pk_roles_id PRIMARY KEY(id)
);

CREATE TABLE permissions
(
  id UUID NOT NULL,
  name VARCHAR(100) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  updated_by VARCHAR,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT pk_permissions_id PRIMARY KEY(id)
);

CREATE TABLE users_roles
(
  user_id UUID,
  role_id UUID,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  updated_by VARCHAR,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT fk_user_id FOREIGN KEY(user_id) REFERENCES users(id),
  CONSTRAINT fk_role_id FOREIGN KEY(role_id) REFERENCES roles(id),

  CONSTRAINT pk_users_roles_id PRIMARY KEY(user_id, role_id)
);

CREATE TABLE users_permissions
(
  user_id UUID,
  permission_id UUID,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  updated_by VARCHAR,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT fk_user_id FOREIGN KEY(user_id) REFERENCES users(id),
  CONSTRAINT fk_permission_id FOREIGN KEY(permission_id) REFERENCES permissions(id),

  CONSTRAINT pk_users_permissions_id PRIMARY KEY(user_id, permission_id)
);

CREATE TABLE roles_permissions
(
  role_id UUID,
  permission_id UUID,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  updated_by VARCHAR,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT fk_role_id FOREIGN KEY(role_id) REFERENCES roles(id),
  CONSTRAINT fk_permission_id FOREIGN KEY(permission_id) REFERENCES permissions(id),

  CONSTRAINT pk_roles_permissions PRIMARY KEY(role_id, permission_id)
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
  resource_id UUID,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  updated_by VARCHAR,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT pk_image_id PRIMARY KEY(id)
);
