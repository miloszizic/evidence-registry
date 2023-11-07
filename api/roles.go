package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/miloszizic/der/service"
)

// CreateRoleHandler is an HTTP handler that creates a new role in the system.
func (app *Application) CreateRoleHandler(w http.ResponseWriter, r *http.Request) {
	// parse params from request
	params, err := paramsParser[service.CreateRoleParams](app, r)
	if err != nil {
		app.logger.Errorw("Error parsing params from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	role, err := app.stores.CreateRole(r.Context(), params)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	app.respond(w, r, http.StatusCreated, envelope{"Role": role})
}

// DeleteRoleHandler is an HTTP handler that deletes a role from the system.
func (app *Application) DeleteRoleHandler(w http.ResponseWriter, r *http.Request) {
	roleID, err := app.roleIDParser(r)
	if err != nil {
		app.logger.Errorw("Error parsing role ID from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	exists, err := app.stores.DBStore.RoleExistsByID(r.Context(), roleID)
	if err != nil {
		app.logger.Errorw("Error checking if role exists", "error", err)
		app.respondError(w, r, err)

		return
	}

	if !exists {
		app.logger.Errorw("Role does not exist", "error", err)
		app.respondError(w, r, service.ErrNotFound)

		return
	}

	err = app.stores.DeleteRole(r.Context(), roleID)
	if err != nil {
		app.logger.Errorw("Error deleting role", "error", err)
		app.respondError(w, r, err)

		return
	}

	app.respond(w, r, http.StatusOK, envelope{"message": "role deleted"})
}

// AddPermissionToRoleHandler is an HTTP handler that adds a permission to a role in the system.
func (app *Application) AddPermissionToRoleHandler(w http.ResponseWriter, r *http.Request) {
	roleID, err := app.roleIDParser(r)
	if err != nil {
		app.logger.Errorw("Error parsing role ID from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	permissionID, err := permissionIDParser(r)
	if err != nil {
		app.logger.Errorw("Error parsing permission ID from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	params := &service.RolePermissionParams{
		PermissionID: permissionID,
	}

	err = app.stores.AddPermissionToRole(r.Context(), roleID, *params)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	app.respond(w, r, http.StatusOK, envelope{"message": "permission added to role"})
}

// RemovePermissionFromRoleHandler is an HTTP handler that removes a permission from a role in the system.
func (app *Application) RemovePermissionFromRoleHandler(w http.ResponseWriter, r *http.Request) {
	roleID, err := app.roleIDParser(r)
	if err != nil {
		app.logger.Errorw("Error parsing role ID from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	permissionID, err := permissionIDParser(r)
	if err != nil {
		app.logger.Errorw("Error parsing permission ID from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	params := &service.RolePermissionParams{
		PermissionID: permissionID,
	}

	err = app.stores.RemovePermissionFromRole(r.Context(), roleID, *params)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	app.respond(w, r, http.StatusOK, envelope{"message": "permission removed from role"})
}

// ListRolesHandler is an HTTP handler that returns a list of roles in the system.
func (app *Application) ListRolesHandler(w http.ResponseWriter, r *http.Request) {
	roles, err := app.stores.ListRoles(r.Context())
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	app.respond(w, r, http.StatusOK, envelope{"Roles": roles})
}

// ListPermissionsHandler is an HTTP handler that returns a list of permissions in the system.
func (app *Application) ListPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	permissions, err := app.stores.ListPermissions(r.Context())
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	app.respond(w, r, http.StatusOK, envelope{"Permissions": permissions})
}

// GetRoleParams defines parameters for getting a role.
type GetRoleParams struct {
	ID uuid.UUID `json:"id"`
}

// GetRoleHandler is an HTTP handler that returns details of a specific role.
// The request must include the role's ID as a parameter roleID in URL.
func (app *Application) GetRoleHandler(w http.ResponseWriter, r *http.Request) {
	// parse params from request
	params, err := paramsParser[GetRoleParams](app, r)
	if err != nil {
		app.logger.Errorw("Error parsing params from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	role, err := app.stores.GetCaseByID(r.Context(), params.ID)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	app.respond(w, r, http.StatusOK, envelope{"Role": role})
}

// GetRolePermissionsHandler is an HTTP handler that returns a list of permissions for a specific role.
// The request must include the role's ID as a parameter roleID in URL.
func (app *Application) GetRolePermissionsHandler(w http.ResponseWriter, r *http.Request) {
	roleID, err := app.roleIDParser(r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	permissions, err := app.stores.GetRolePermissions(r.Context(), roleID)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	app.respond(w, r, http.StatusOK, envelope{"Permissions": permissions})
}
