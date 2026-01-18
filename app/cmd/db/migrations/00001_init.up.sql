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
  slug VARCHAR(100) NOT NULL,
  image_url VARCHAR NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT pk_category_id PRIMARY KEY(id),
  CONSTRAINT uq_categories_slug UNIQUE (slug)
);

CREATE TABLE posts
(
  id UUID NOT NULL,
  title VARCHAR(100) NOT NULL,
  slug VARCHAR(255) NOT NULL,
  content VARCHAR NOT NULL,
  excerpt VARCHAR(500),
  cover_image_url VARCHAR(500),
  status VARCHAR(20) NOT NULL DEFAULT 'created',
  published_at TIMESTAMPTZ,
  meta_description VARCHAR(160),
  reading_time_minutes INTEGER DEFAULT 0,
  category_id UUID NOT NULL,
  creator_user_id UUID NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  updated_by VARCHAR,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT pk_post_id PRIMARY KEY(id),
  CONSTRAINT fk_category_id FOREIGN KEY(category_id) REFERENCES categories(id),
  CONSTRAINT uq_posts_slug UNIQUE (slug),
  CONSTRAINT chk_posts_status CHECK (status IN ('created', 'draft', 'published', 'archived'))
);

CREATE INDEX idx_posts_status ON posts (status, published_at DESC) WHERE is_deleted = FALSE;

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

CREATE TABLE password_reset_tokens
(
  id UUID NOT NULL,
  user_id UUID NOT NULL,
  token_hash VARCHAR(64) NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),

  CONSTRAINT pk_password_reset_tokens_id PRIMARY KEY(id),
  CONSTRAINT fk_password_reset_tokens_user_id FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE INDEX idx_password_reset_tokens_hash ON password_reset_tokens (token_hash) WHERE used_at IS NULL;
CREATE INDEX idx_password_reset_tokens_user ON password_reset_tokens (user_id, expires_at DESC);
