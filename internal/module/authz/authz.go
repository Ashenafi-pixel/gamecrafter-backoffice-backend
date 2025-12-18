package authz

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type authz struct {
	Log          *zap.Logger
	authzStorage storage.Authz
	userStorage  storage.User
}

func Init(log *zap.Logger, authzStorage storage.Authz, userStorage storage.User) module.Authz {
	a := &authz{
		authzStorage: authzStorage,
		userStorage:  userStorage,
		Log:          log,
	}

	return a
}

func (a *authz) CreateRole(ctx context.Context, req dto.CreateRoleReq) (dto.Role, error) {
	// check if the role is exist or not
	var permissions []dto.Permissions
	_, exist, err := a.authzStorage.GetRoleByName(ctx, req.Name)
	if err != nil {
		return dto.Role{}, err
	}

	if exist {
		err := fmt.Errorf("role already exist with this role name")
		err = errors.ErrInvalidUserInput.Wrap(err, "%s", err.Error())
		return dto.Role{}, err
	}

	r, err := a.authzStorage.CreateRole(ctx, req)
	if err != nil {
		return dto.Role{}, err
	}

	// validate  permissions

	if len(req.Permissions) > 0 {
		//check if the permissions are exist or not
		for _, p := range req.Permissions {
			rpm, exist, err := a.authzStorage.GetPermissionByID(ctx, p.PermissionID)
			if err != nil {
				a.authzStorage.RemoveRolePermissionsByRoleID(ctx, r.ID)
				return dto.Role{}, err
			}
			if !exist {
				err := fmt.Errorf("permission not found with id %s", p.PermissionID.String())
				a.Log.Error(err.Error(), zap.Any("role_req", req))
				err = errors.ErrInvalidUserInput.Wrap(err, "%s", err.Error())
				continue
			}
			if rpm.Name == "super" {
				err := fmt.Errorf("try to assign super admin %s", p.PermissionID.String())
				a.Log.Error(err.Error(), zap.Any("role_req", req))
				continue
			}
			// assign permission to the role with value + limits
			_, err = a.authzStorage.AssignPermissionToRole(ctx, p.PermissionID, r.ID, p.Value, p.LimitType, p.LimitPeriod)
			permissions = append(permissions, rpm)
			if err != nil {
				a.authzStorage.RemoveRolePermissionsByRoleID(ctx, r.ID)
				return dto.Role{}, err
			}

			// Permissions are now managed directly in role_permissions table
			// No need to sync to casbin_rule table
		}
	}

	return dto.Role{
		ID:          r.ID,
		Name:        r.Name,
		Permissions: permissions,
	}, nil
}

func (a *authz) GetPermissions(ctx context.Context, req dto.GetPermissionReq) ([]dto.Permissions, error) {
	var response []dto.Permissions
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	resp, err := a.authzStorage.GetPermissions(ctx, req)
	if err != nil {
		return []dto.Permissions{}, err
	}
	for _, p := range resp {
		if p.Name != "super" {
			response = append(response, p)
		}
	}
	return response, nil
}

func (a *authz) GetRoles(ctx context.Context, req dto.GetRoleReq) ([]dto.Role, error) {
	var rolesResponse []dto.Role
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	rs, exist, err := a.authzStorage.GetRoles(ctx, req)
	if err != nil {
		return []dto.Role{}, err
	}
	if !exist {
		return []dto.Role{}, nil
	}

	for _, r := range rs {
		if r.Name != "super" {
			rolesResponse = append(rolesResponse, r)

		}
	}

	return rolesResponse, nil
}

func (a *authz) UpdatePermissionsOfRole(ctx context.Context, req dto.UpdatePermissionToRoleReq) (dto.UpdatePermissionToRoleRes, error) {
	var permissions []dto.Permissions

	// check if role exist or not
	rl, exist, err := a.authzStorage.GetRoleByID(ctx, req.RoleID)
	if err != nil {
		return dto.UpdatePermissionToRoleRes{}, err
	}

	if !exist {
		err := fmt.Errorf("role not found")
		err = errors.ErrInvalidUserInput.Wrap(err, "%s", err.Error())
		return dto.UpdatePermissionToRoleRes{}, err
	}

	//get previous permissions

	//remove previous permissions for role
	a.authzStorage.RemoveRolePermissionsByRoleID(ctx, req.RoleID)

	for _, pr := range req.Permissions {

		//check if the permissions exist or not
		rpm, exist, err := a.authzStorage.GetPermissionByID(ctx, pr.PermissionID)
		if err != nil {
			return dto.UpdatePermissionToRoleRes{}, err
		}
		if !exist {
			a.Log.Warn("invalid permission id is given", zap.Any("req", req))
			continue
		}
		if rpm.Name == "super" {
			err := fmt.Errorf("try to assign super admin %s", rpm.ID.String())
			a.Log.Error(err.Error(), zap.Any("role_req", req))
			continue
		}
		// check if the permissions already exist or not

		_, err = a.authzStorage.AssignPermissionToRole(ctx, pr.PermissionID, req.RoleID, pr.Value, pr.LimitType, pr.LimitPeriod)
		permissions = append(permissions, rpm)
		if err != nil {
			a.authzStorage.RemoveRolePermissionsByRoleID(ctx, req.RoleID)
			return dto.UpdatePermissionToRoleRes{}, err
		}

		// Permissions are now managed directly in role_permissions table
		// No need to sync to casbin_rule table

	}
	return dto.UpdatePermissionToRoleRes{
		Message: constant.SUCCESS,
		Role: dto.Role{
			ID:          req.RoleID,
			Name:        rl.Name,
			Permissions: permissions,
		},
	}, nil
}

func (a *authz) RemoveRole(ctx context.Context, roleID uuid.UUID) error {
	rl, exist, err := a.authzStorage.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}

	if !exist {
		err := fmt.Errorf("role not found")
		err = errors.ErrInvalidUserInput.Wrap(err, "%s", err.Error())
		return err
	}
	if rl.Name == "super" {
		err := fmt.Errorf("super admin role can not be removed")
		err = errors.ErrInvalidUserInput.Wrap(err, "%s", err.Error())
		return err
	}
	a.authzStorage.RemoveRoleFromUserRoles(ctx, roleID)
	a.authzStorage.RemoveRolePermissionsByRoleID(ctx, roleID)
	a.authzStorage.RemoveRoleByID(ctx, rl.ID)
	return nil
}

func (a *authz) AssignRoleToUser(ctx context.Context, req dto.AssignRoleToUserReq) (dto.AssignRoleToUserRes, error) {
	var resp dto.AssignRoleToUserRes
	var roles []dto.Role
	//check if role exist or not
	r, exist, err := a.authzStorage.GetRoleByID(ctx, req.RoleID)
	if err != nil {
		return dto.AssignRoleToUserRes{}, err
	}

	if !exist {
		a.Log.Error("role dose not exist", zap.Any("assingRoleReq", req))
		err := fmt.Errorf("role dose not exist with this id")
		err = errors.ErrInvalidUserInput.Wrap(err, "%s", err.Error())
		return dto.AssignRoleToUserRes{}, err
	}
	if r.Name == "super" {
		err := fmt.Errorf("can not assign super admin role")
		err = errors.ErrInvalidUserInput.Wrap(err, "%s", err.Error())
		return dto.AssignRoleToUserRes{}, err
	}

	// If assigning "manager" or "admin role", replace all existing roles (except super)
	if r.Name == "manager" || r.Name == "admin role" {
		// Remove all existing roles for this user (except super)
		err = a.authzStorage.RemoveAllUserRolesExceptSuper(ctx, req.UserID)
		if err != nil {
			a.Log.Error("failed to remove existing roles", zap.Error(err), zap.Any("userID", req.UserID))
			err = errors.ErrUnableToUpdate.Wrap(err, "%s", err.Error())
			return dto.AssignRoleToUserRes{}, err
		}
	} else {
		// For other roles, check if the role is already assigned
		_, exist, err = a.authzStorage.GetUserRoleUsingUserIDandRole(ctx, req.UserID, req.RoleID)
		if err != nil {
			return dto.AssignRoleToUserRes{}, err
		}

		if exist {
			err := fmt.Errorf("role already assigned to the user")
			err = errors.ErrInvalidUserInput.Wrap(err, "%s", err.Error())
			return dto.AssignRoleToUserRes{}, err
		}
	}

	_, err = a.authzStorage.AddRoleToUser(ctx, req.RoleID, req.UserID)

	if err != nil {
		return dto.AssignRoleToUserRes{}, err
	}
	// get user roles
	usrRoles, exist, err := a.authzStorage.GetUserRoles(ctx, req.UserID)
	if err != nil {
		return dto.AssignRoleToUserRes{}, err
	}
	if !exist {
		err := fmt.Errorf("unable to assign role to the user")
		err = errors.ErrInternalServerError.Wrap(err, "%s", err.Error())
		return dto.AssignRoleToUserRes{}, err
	}
	for _, rl := range usrRoles {

		rl, exist, err := a.authzStorage.GetRoleByID(ctx, rl)

		if err != nil {
			return dto.AssignRoleToUserRes{}, err
		}

		if !exist {
			continue
		}
		roles = append(roles, rl)

	}
	resp.UserID = req.UserID
	resp.Roles = roles
	return resp, nil
}

func (a *authz) RevokeUserRole(ctx context.Context, req dto.UserRole) error {
	// check if user role exist
	r, exist, err := a.authzStorage.GetUserRoleUsingUserIDandRole(ctx, req.UserID, req.RoleID)
	if err != nil {
		return err
	}
	if !exist {
		err := fmt.Errorf("user dose not have requested role")
		err = errors.ErrInvalidUserInput.Wrap(err, "%s", err.Error())
		return err
	}

	// get role by id
	rl, exist, err := a.authzStorage.GetRoleByID(ctx, r.RoleID)
	if err != nil {
		return err
	}
	if !exist {
		err := fmt.Errorf("role dose not exist")
		err = errors.ErrInvalidUserInput.Wrap(err, "%s", err.Error())
		return err
	}

	if rl.Name == "super" {
		err := fmt.Errorf("super admin role can not be removed")
		err = errors.ErrInvalidUserInput.Wrap(err, "%s", err.Error())
		return err
	}
	return a.authzStorage.RevokeUserRole(ctx, req.UserID, req.RoleID)

}

func (a *authz) GetRoleUsers(ctx context.Context, roleID uuid.UUID) ([]dto.User, error) {
	return a.authzStorage.GetRoleUsers(ctx, roleID)
}

func (a *authz) GetUserRoles(ctx context.Context, userID uuid.UUID) (dto.UserRolesRes, error) {
	resp := dto.UserRolesRes{}
	var roles []dto.Role
	userRoles, exist, err := a.authzStorage.GetUserRoles(ctx, userID)

	if err != nil {
		return resp, err
	}

	if !exist {
		return resp, nil
	}

	for _, rl := range userRoles {
		rl, exist, err := a.authzStorage.GetRoleByID(ctx, rl)

		if err != nil {
			return resp, err
		}

		if !exist {
			continue
		}

		roles = append(roles, rl)
	}
	resp.UserID = userID
	resp.Roles = roles
	return resp, nil
}

func (a *authz) GetAdmins(ctx context.Context, req dto.GetAdminsReq) ([]dto.Admin, error) {
	return a.userStorage.GetAllAdminUsers(ctx, req)
}

func (a *authz) CheckUserHasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error) {
	return a.authzStorage.CheckUserHasPermission(ctx, userID, permissionName)
}
