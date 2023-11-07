-- name: CreateUser :one
INSERT INTO "app_users" (
  username,
  first_name,
  last_name,
  email,
  password,
  role_id
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UpdateUserPassword :one
UPDATE app_users
SET password = $1
WHERE id = $2
RETURNING *;

-- name: AddRoleToUser :one
UPDATE app_users
SET role_id = $1
WHERE id = $2
RETURNING *;

-- name: GetUser :one
SELECT * FROM "app_users" WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM "app_users" WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM "app_users" WHERE id = $1;


-- name: GetUserByUsername :one
SELECT * FROM "app_users" WHERE username = $1;

-- name: ListUsers :many
SELECT * FROM "app_users";

-- name: UpdateUser :one
UPDATE app_users
SET
  username = $1,
  first_name = $2,
  last_name = $3,
  email = $4,
  password = $5,
  role_id = $6
WHERE id = $7
RETURNING *;


-- name: DeleteUser :exec
DELETE FROM "app_users" WHERE id = $1;

-- name: UserExists :one
SELECT EXISTS(SELECT 1 FROM "app_users" WHERE username = $1);

-- name: UserExistsByID :one
SELECT EXISTS(SELECT 1 FROM "app_users" WHERE id = $1);

-- name: CreateUserCase :one
INSERT INTO "user_cases" (
  user_id,
  case_id
) VALUES (
  $1, $2
) RETURNING *;

-- name: GetUserCaseID :one
SELECT id FROM "user_cases" WHERE case_id = $1;

-- name: GetUserWithRole :one
SELECT u.*, r.name as role_name
FROM app_users AS u
LEFT JOIN role AS r ON u.role_id = r.id
WHERE u.username = $1;

-- name: GetRolePermissions :many
SELECT p.name
FROM role_permissions AS rp
INNER JOIN permissions AS p ON rp.permission_id = p.id
WHERE rp.role_id = $1;

-- name: GetPermissionsForRole :many
SELECT permissions.name
FROM permissions
JOIN role_permissions ON role_permissions.permission_id = permissions.id
WHERE role_permissions.role_id = $1;

-- name: CreatePermission :one
INSERT INTO permissions (name)
VALUES ($1)
RETURNING *;

-- name: AssignRoleToUser :exec
UPDATE app_users
SET role_id = $1
WHERE id = $2;

-- name: GetUsers :many
SELECT * FROM app_users;

-- name: GetUsersWithRoles :many
SELECT app_users.*, role.name AS role_name
FROM app_users
INNER JOIN role ON app_users.role_id = role.id;
