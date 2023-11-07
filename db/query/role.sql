-- name: CreateRole :one
INSERT INTO "role" (
  name,
  code
) VALUES (
  $1,
  $2
) RETURNING *;

-- name: RoleExistsByID :one
SELECT EXISTS(SELECT 1 FROM "role" WHERE id = $1);

-- name: RoleExistsByName :one
SELECT EXISTS(SELECT 1 FROM "role" WHERE name = $1);

-- name: GetRoleID :one
SELECT id FROM "role" WHERE name = $1;

-- name: ListRoles :many
SELECT * FROM "role";

-- name: ListPermissions :many
SELECT * FROM "permissions";

-- name: UpdateRole :one
UPDATE "role"
SET
  name = $2,
  code = $3
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM "role" WHERE id = $1;

-- name: GetRoleByName :one
SELECT *
FROM role
WHERE name = $1;

-- name: GetRoleByID :one
SELECT *
FROM role
WHERE id = $1;

-- name: GetPermissionIDByName :one
SELECT id
FROM permissions
WHERE name = $1;


-- name: AddRolePermission :one
INSERT INTO "role_permissions" (
  role_id,
  permission_id
) VALUES (
  $1,
  $2
) RETURNING *;

-- name: AddMultiplePermissionsToRole :many
INSERT INTO "role_permissions" (
  role_id,
  permission_id
)
SELECT
  $1,
  permission_id
FROM unnest($2::uuid[]) AS permission_id
RETURNING *;

-- name: PermissionExists :one
SELECT EXISTS(SELECT 1 FROM "permissions" WHERE id = $1);

-- name: ListRolePermissions :many
SELECT * FROM "role_permissions";


-- name: UpdateRolePermission :one
UPDATE "role_permissions"
SET
  role_id = $2,
  permission_id = $3
WHERE id = $1
RETURNING *;

-- name: DeleteRolePermission :exec
DELETE FROM "role_permissions"
WHERE role_id = $1 AND permission_id = $2;


-- name: GetRolePermissionsByRoleID :many
SELECT *
FROM "role_permissions"
WHERE role_id = $1;

-- name: GetRoleWithPermissions :one
WITH role_perms AS (
  SELECT role.id AS role_id, role.name AS role_name, role.code AS role_code,
         permissions.id AS permission_id, permissions.name AS permission_name, permissions.code AS permission_code
  FROM role
  JOIN role_permissions ON role.id = role_permissions.role_id
  JOIN permissions ON role_permissions.permission_id = permissions.id
  WHERE role.id = $1
)
SELECT role_id, role_name, role_code,
       json_agg(json_build_object('id', permission_id, 'name', permission_name, 'code', permission_code)) AS permissions
FROM role_perms
GROUP BY role_id, role_name, role_code;


-- name: GetRolePermissionsByPermissionID :many
SELECT *
FROM "role_permissions"
WHERE permission_id = $1;
