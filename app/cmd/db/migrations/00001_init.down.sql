-- Drop dependents before the tables they reference, otherwise the foreign keys
-- reject the drop.
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS users_permissions;
DROP TABLE IF EXISTS users_roles;
DROP TABLE IF EXISTS roles_permissions;
DROP TABLE IF EXISTS images;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
