package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/miloszizic/der/db"
)

// CreateRoleParams holds the parameters for creating a new role.
type CreateRoleParams struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// CreateRole creates a new role in the database.
func (s *Stores) CreateRole(ctx context.Context, params CreateRoleParams) (*Role, error) {
	// check if a role already exists
	exists, err := s.DBStore.RoleExistsByName(ctx, params.Name)
	if err != nil {
		return nil, fmt.Errorf("error checking if role exists: %w", err)
	}

	if exists {
		return nil, ErrAlreadyExists
	}
	roleParams := db.CreateRoleParams{
		Name: params.Name,
		Code: params.Code,
	}
	DBRole, err := s.DBStore.CreateRole(ctx, roleParams)
	if err != nil {
		return nil, fmt.Errorf("error creating role: %w", err)
	}
	role := Role{
		ID:   DBRole.ID,
		Name: DBRole.Name,
		Code: DBRole.Code,
	}
	return &role, nil
}

// DeleteRole deletes a role from the database.
func (s *Stores) DeleteRole(ctx context.Context, ID uuid.UUID) error {
	// check if a role exists
	exists, err := s.DBStore.RoleExistsByID(ctx, ID)
	if err != nil {
		return fmt.Errorf("error checking if role exists: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	err = s.DBStore.DeleteRole(ctx, ID)
	if err != nil {
		return fmt.Errorf("error deleting role: %w", err)
	}
	return nil
}

// Role holds the details of a role in the service layer.
type Role struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

// ListRoles returns a list of roles from the database.
func (s *Stores) ListRoles(ctx context.Context) ([]Role, error) {
	DBRoles, err := s.DBStore.ListRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing roles: %w", err)
	}

	roles := make([]Role, 0, len(DBRoles))

	for _, DBRole := range DBRoles {
		role := Role{
			ID:   DBRole.ID,
			Name: DBRole.Name,
			Code: DBRole.Code,
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// Permission holds the details of a permission in the service layer.
type Permission struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

// ListPermissions returns a list of permissions from the database.
func (s *Stores) ListPermissions(ctx context.Context) ([]Permission, error) {
	DBPermissions, err := s.DBStore.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing permissions: %w", err)
	}

	permissions := make([]Permission, 0, len(DBPermissions))

	for _, DBPermission := range DBPermissions {
		permission := Permission{
			ID:   DBPermission.ID,
			Name: DBPermission.Name,
			Code: DBPermission.Code,
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// GetRoleByName returns a role from the database.
func (s *Stores) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	DBRole, err := s.DBStore.GetRoleByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error getting role: %w", err)
	}

	role := Role{
		ID:   DBRole.ID,
		Name: DBRole.Name,
		Code: DBRole.Code,
	}

	return &role, nil
}

// UpdateRoleParams holds the parameters for updating a role.
type UpdateRoleParams struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

// UpdateRole updates a role in the database.
func (s *Stores) UpdateRole(ctx context.Context, params UpdateRoleParams) (*Role, error) {
	// check if a role exists
	exists, err := s.DBStore.RoleExistsByID(ctx, params.ID)
	if err != nil {
		return nil, fmt.Errorf("error checking if role exists: %w", err)
	}

	if !exists {
		return nil, ErrNotFound
	}

	updateParas := db.UpdateRoleParams{
		ID:   params.ID,
		Name: params.Name,
		Code: params.Code,
	}
	DBRole, err := s.DBStore.UpdateRole(ctx, updateParas)
	if err != nil {
		return nil, fmt.Errorf("error updating role: %w", err)
	}

	role := Role{
		ID:   DBRole.ID,
		Name: DBRole.Name,
		Code: DBRole.Code,
	}

	return &role, nil
}

// RolePermissionParams holds the parameters for adding a permission to a role.
type RolePermissionParams struct {
	PermissionID uuid.UUID `json:"permission_id"`
}

// AddPermissionToRole adds a permission to a role in the database.
func (s *Stores) AddPermissionToRole(ctx context.Context, roleID uuid.UUID, params RolePermissionParams) error {
	// check if a role exists
	exists, err := s.DBStore.RoleExistsByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("error checking if role exists: %w", err)
	}

	if !exists {
		return ErrNotFound
	}

	exists, err = s.DBStore.PermissionExists(ctx, params.PermissionID)
	if err != nil {
		return fmt.Errorf("error checking if permission exists: %w", err)
	}

	if !exists {
		return ErrNotFound
	}

	permissionParams := db.AddRolePermissionParams{
		RoleID:       roleID,
		PermissionID: params.PermissionID,
	}

	_, err = s.DBStore.AddRolePermission(ctx, permissionParams)
	if err != nil {
		return fmt.Errorf("error adding permission to role: %w", err)
	}

	return nil
}

// RemovePermissionFromRole removes a permission from a role in the database.
func (s *Stores) RemovePermissionFromRole(ctx context.Context, roleID uuid.UUID, params RolePermissionParams) error {
	// check if a role exists
	exists, err := s.DBStore.RoleExistsByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("error checking if role exists: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	exists, err = s.DBStore.PermissionExists(ctx, params.PermissionID)
	if err != nil {
		return fmt.Errorf("error checking if permission exists: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	permissionParams := db.DeleteRolePermissionParams{
		RoleID:       roleID,
		PermissionID: params.PermissionID,
	}
	err = s.DBStore.DeleteRolePermission(ctx, permissionParams)
	if err != nil {
		return fmt.Errorf("error removing permission from role: %w", err)
	}

	return nil
}

// RoleWithPermissions holds the details of a role with its permissions in the service layer.
type RoleWithPermissions struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Code        string       `json:"code"`
	Permissions []Permission `json:"permissions"`
}

// GetRolePermissions returns a list of permissions for a role from the database.
func (s *Stores) GetRolePermissions(ctx context.Context, roleID uuid.UUID) (*RoleWithPermissions, error) {
	// check if a role exists
	exists, err := s.DBStore.RoleExistsByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("error checking if role exists: %w", err)
	}
	if !exists {
		return nil, ErrNotFound
	}

	DBRolePermissions, err := s.DBStore.GetRoleWithPermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("error getting role permissions: %w", err)
	}

	var permissions []Permission
	err = json.Unmarshal(DBRolePermissions.Permissions, &permissions)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling permissions: %w", err)
	}

	return &RoleWithPermissions{
		ID:          DBRolePermissions.RoleID,
		Name:        DBRolePermissions.RoleName,
		Code:        DBRolePermissions.RoleCode,
		Permissions: permissions,
	}, nil
}
