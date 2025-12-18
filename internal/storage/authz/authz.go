package authz

import (
	"context"
	"database/sql"
	"os"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type authz struct {
	db  *gorm.DB
	pdb *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *gorm.DB, log *zap.Logger, pdb *persistencedb.PersistenceDB) storage.Authz {
	a := &authz{
		db:  db,
		log: log,
		pdb: pdb,
	}

	// Skip permission initialization when using server database
	if os.Getenv("SKIP_PERMISSION_INIT") != "true" {
		a.InitPermissions()
	} else {
		log.Info("Skipping permission initialization (using server database)")
	}

	return a
}

func (a *authz) InitPermissions() {
	// Permissions are now managed via database migrations
	// Only initialize the "super" permission for super admin role
	// All other permissions come from: migrations/20250117000001_seed_page_permissions.up.sql
	for _, permission := range dto.PermissionsList {
		// Only handle "super" permission - all others come from migrations
		if permission.Name == "super" {
			var existingPermission dto.Permissions
			if err := a.db.Where("name = ?", permission.Name).First(&existingPermission).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					id := uuid.New()
					p := dto.Permissions{
						ID:          id,
						Name:        permission.Name,
						Description: permission.Description,
						RequiresValue: false,
					}
					if err := a.db.Create(&p).Error; err != nil {
						a.log.Error("Error creating permission", zap.Error(err))
						a.log.Fatal(err.Error())
					}
				} else {
					a.log.Error("Error checking permission existence", zap.Error(err))
					a.log.Error("Continuing without permissions table - this is expected for new installations")
				}
			}
		}
	}

	//get supper admin role if not found create one
	_, err := a.pdb.Queries.GetSupperAdmin(context.Background())
	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error("Error getting super admin role", zap.Error(err))
		a.log.Error("Continuing without super admin role - this is expected for new installations")
		return
	}

	if err != nil {
		// create super admin role
		supA, err := a.pdb.Queries.CreateRole(context.Background(), "super")
		if err != nil {
			a.log.Error("Error creating super admin role", zap.Error(err))
			a.log.Error("Continuing without super admin role - this is expected for new installations")
			return
		}
		// get super admin permission
		p, err := a.pdb.GetPermissonByName(context.Background(), "super")
		if err != nil {
			a.log.Error("Error getting super admin permission", zap.Error(err))
			a.log.Error("Continuing without super admin permission - this is expected for new installations")
			return
		}
		// add permissions role (super admin has unlimited, so value is nil)
		_, err = a.pdb.Queries.AssignPermissionToRole(context.Background(), db.AssignPermissionToRoleParams{
			RoleID:       supA.ID,
			PermissionID: p.ID,
			Value:        nil, // Super admin has unlimited funding
		})
		if err != nil {
			a.log.Error("Error assigning permission to role", zap.Error(err))
			a.log.Error("Continuing without role assignment - this is expected for new installations")
			return
		}
		// create casbin rule to casbin_rule table
		_, err = a.pdb.AddSupperAdminCasbinRule(context.Background(), sql.NullString{String: supA.ID.String(), Valid: true})
		if err != nil {
			a.log.Error("Error adding super admin casbin rule", zap.Error(err))
			a.log.Error("Continuing without casbin rule - this is expected for new installations")
			return
		}
	}
}

func (a *authz) CreateRole(ctx context.Context, req dto.CreateRoleReq) (dto.Role, error) {

	res, err := a.pdb.Queries.CreateRole(ctx, req.Name)
	if err != nil {
		a.log.Error(err.Error(), zap.Any("create_role_req", req))
		err = errors.ErrInvalidUserInput.Wrap(err, "%s", err.Error())
		return dto.Role{}, err
	}
	return dto.Role{
		ID:   res.ID,
		Name: res.Name,
	}, nil
}

// GetPermissionByID returns a permission including requires_value
func (a *authz) GetPermissionByID(ctx context.Context, permissionID uuid.UUID) (dto.Permissions, bool, error) {
	query := `
		SELECT id, name, description, COALESCE(requires_value, FALSE) as requires_value
		FROM permissions
		WHERE id = $1
	`

	var (
		id           uuid.UUID
		name         string
		description  sql.NullString
		requiresValue bool
	)

	err := a.pdb.GetPool().QueryRow(ctx, query, permissionID).Scan(&id, &name, &description, &requiresValue)
	if err != nil {
		if err.Error() == dto.ErrNoRows {
			return dto.Permissions{}, false, nil
		}
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
		return dto.Permissions{}, false, err
	}

	return dto.Permissions{
		ID:            id,
		Name:          name,
		Description:   description.String,
		RequiresValue: requiresValue,
	}, true, nil
}

func (a *authz) AssignPermissionToRole(ctx context.Context, permissionID, roleID uuid.UUID, value *float64, limitType *string, limitPeriod *int) (dto.AssignPermissionToRoleRes, error) {
	// Convert *float64 to *decimal.Decimal-ish value for query (NUMERIC)
	var valueParam interface{}
	if value != nil {
		valueParam = *value
	} else {
		valueParam = nil
	}

	// Insert directly (sqlc models are stale vs new columns)
	query := `
		INSERT INTO role_permissions (role_id, permission_id, value, limit_type, limit_period)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, role_id, permission_id, value, limit_type, limit_period
	`

	var (
		id  uuid.UUID
		rid uuid.UUID
		pid uuid.UUID
		val decimal.NullDecimal
		lt  sql.NullString
		lp  sql.NullInt32
	)

	err := a.pdb.GetPool().QueryRow(ctx, query, roleID, permissionID, valueParam, limitType, limitPeriod).Scan(&id, &rid, &pid, &val, &lt, &lp)
	if err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
		return dto.AssignPermissionToRoleRes{}, err
	}

	// Convert decimal back to float64 for response
	var valueFloat *float64
	if val.Valid {
		f, _ := val.Decimal.Float64()
		valueFloat = &f
	}

	var limitTypeOut *string
	if lt.Valid {
		s := lt.String
		limitTypeOut = &s
	}

	var limitPeriodOut *int
	if lp.Valid {
		i := int(lp.Int32)
		limitPeriodOut = &i
	}

	return dto.AssignPermissionToRoleRes{
		Message: constant.SUCCESS,
		Data: dto.AssignPermissionToRoleData{
			ID:           id,
			RoleID:       rid,
			PermissionID: pid,
			Value:        valueFloat,
			LimitType:    limitTypeOut,
			LimitPeriod:  limitPeriodOut,
		},
	}, nil
}

func (a *authz) GetRoleByName(ctx context.Context, name string) (dto.Role, bool, error) {
	r, err := a.pdb.Queries.GetRoleByName(ctx, name)
	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
		return dto.Role{}, false, err
	}
	if err != nil {
		return dto.Role{}, false, nil
	}
	return dto.Role{
		ID:   r.ID,
		Name: r.Name,
	}, true, nil
}

func (a *authz) RemoveRoleByID(ctx context.Context, roleID uuid.UUID) error {
	if err := a.pdb.Queries.RemoveRole(ctx, roleID); err != nil {
		a.log.Error(err.Error(), zap.Any("roleID", roleID.String()))
		err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
		return err
	}
	return nil
}

func (a *authz) RemoveRolePermissions(ctx context.Context, id uuid.UUID) error {
	if err := a.pdb.Queries.RemoveRolesPermissions(ctx, id); err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, "%s", err.Error())
		return err
	}
	return nil
}

func (a *authz) RemoveRolePermissionsByRoleID(ctx context.Context, id uuid.UUID) error {
	if err := a.pdb.Queries.RemoveRolesPermissionByRoleID(ctx, id); err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, "%s", err.Error())
		return err
	}
	return nil
}

func (a *authz) GetPermissions(ctx context.Context, req dto.GetPermissionReq) ([]dto.Permissions, error) {
	query := `
		SELECT id, name, description, COALESCE(requires_value, FALSE) as requires_value
		FROM permissions
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := a.pdb.GetPool().Query(ctx, query, req.PerPage, req.Page)
	if err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
		return []dto.Permissions{}, err
	}
	defer rows.Close()

	var res []dto.Permissions
	for rows.Next() {
		var (
			id           uuid.UUID
			name         string
			description  sql.NullString
			requiresValue bool
		)
		if err := rows.Scan(&id, &name, &description, &requiresValue); err != nil {
			a.log.Error(err.Error())
			err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
			return []dto.Permissions{}, err
		}

		res = append(res, dto.Permissions{
			ID:            id,
			Name:          name,
			Description:   description.String,
			RequiresValue: requiresValue,
		})
	}
	if err := rows.Err(); err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
		return []dto.Permissions{}, err
	}

	return res, nil
}

func (a *authz) GetRoles(ctx context.Context, getRoleReq dto.GetRoleReq) ([]dto.Role, bool, error) {
	var rolesRes []dto.Role
	roles, err := a.pdb.Queries.GetRoles(ctx, db.GetRolesParams{
		Limit:  int32(getRoleReq.PerPage),
		Offset: int32(getRoleReq.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
		return []dto.Role{}, false, err
	}

	if err != nil {
		return []dto.Role{}, false, nil
	}

	// getPermissionForRoles
	for _, r := range roles {
		var permissions []dto.Permissions
		var permissionsWithValue []dto.PermissionWithValue

		// Load role permissions including requires_value and limit fields
		query := `
			SELECT
				p.id,
				p.name,
				p.description,
				COALESCE(p.requires_value, FALSE) as requires_value,
				rp.value,
				rp.limit_type,
				rp.limit_period
			FROM role_permissions rp
			JOIN permissions p ON p.id = rp.permission_id
			WHERE rp.role_id = $1
			ORDER BY p.name ASC
		`
		rows, err := a.pdb.GetPool().Query(ctx, query, r.ID)
		if err != nil {
			a.log.Error(err.Error())
			err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
			return []dto.Role{}, false, err
		}
		for rows.Next() {
			var (
				pid           uuid.UUID
				name          string
				description   sql.NullString
				requiresValue bool
				val           decimal.NullDecimal
				lt            sql.NullString
				lp            sql.NullInt32
			)
			if err := rows.Scan(&pid, &name, &description, &requiresValue, &val, &lt, &lp); err != nil {
				rows.Close()
				a.log.Error(err.Error())
				err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
				return []dto.Role{}, false, err
			}

			permissions = append(permissions, dto.Permissions{
				ID:            pid,
				Name:          name,
				Description:   description.String,
				RequiresValue: requiresValue,
			})

			var valueFloat *float64
			if val.Valid {
				f, _ := val.Decimal.Float64()
				valueFloat = &f
			}

			var limitTypeOut *string
			if lt.Valid {
				s := lt.String
				limitTypeOut = &s
			}

			var limitPeriodOut *int
			if lp.Valid {
				i := int(lp.Int32)
				limitPeriodOut = &i
			}

			permissionsWithValue = append(permissionsWithValue, dto.PermissionWithValue{
				PermissionID: pid,
				Value:        valueFloat,
				LimitType:    limitTypeOut,
				LimitPeriod:  limitPeriodOut,
			})
		}
		rows.Close()

		rolesRes = append(rolesRes, dto.Role{
			ID:                   r.ID,
			Name:                 r.Name,
			Permissions:          permissions,
			PermissionsWithValue: permissionsWithValue,
		})

	}
	return rolesRes, true, nil
}

func (a *authz) GetRoleByID(ctx context.Context, roleID uuid.UUID) (dto.Role, bool, error) {
	var permissions []dto.Permissions
	var permissionsWithValue []dto.PermissionWithValue
	rl, err := a.pdb.Queries.GetRoleByID(ctx, roleID)
	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
		return dto.Role{}, false, err
	}

	if err != nil {
		return dto.Role{}, false, nil
	}

	query := `
		SELECT
			p.id,
			p.name,
			p.description,
			COALESCE(p.requires_value, FALSE) as requires_value,
			rp.value,
			rp.limit_type,
			rp.limit_period
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.name ASC
	`
	rows, err := a.pdb.GetPool().Query(ctx, query, roleID)
	if err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
		return dto.Role{}, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			pid           uuid.UUID
			name          string
			description   sql.NullString
			requiresValue bool
			val           decimal.NullDecimal
			lt            sql.NullString
			lp            sql.NullInt32
		)
		if err := rows.Scan(&pid, &name, &description, &requiresValue, &val, &lt, &lp); err != nil {
			a.log.Error(err.Error())
			err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
			return dto.Role{}, false, err
		}

		permissions = append(permissions, dto.Permissions{
			ID:            pid,
			Name:          name,
			Description:   description.String,
			RequiresValue: requiresValue,
		})

		var valueFloat *float64
		if val.Valid {
			f, _ := val.Decimal.Float64()
			valueFloat = &f
		}

		var limitTypeOut *string
		if lt.Valid {
			s := lt.String
			limitTypeOut = &s
		}

		var limitPeriodOut *int
		if lp.Valid {
			i := int(lp.Int32)
			limitPeriodOut = &i
		}

		permissionsWithValue = append(permissionsWithValue, dto.PermissionWithValue{
			PermissionID: pid,
			Value:        valueFloat,
			LimitType:    limitTypeOut,
			LimitPeriod:  limitPeriodOut,
		})
	}
	if err := rows.Err(); err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
		return dto.Role{}, false, err
	}

	return dto.Role{
		ID:                   rl.ID,
		Name:                 rl.Name,
		Permissions:          permissions,
		PermissionsWithValue: permissionsWithValue,
	}, true, nil
}

func (a *authz) RemoveRolePermissionFromCasbinRule(ctx context.Context, roleID uuid.UUID) error {
	if err := a.pdb.RemoveFromCasbinRule(ctx, roleID); err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, "%s", err.Error())
		return err
	}
	return nil
}

func (a *authz) GetRolePermissionsByRoleID(ctx context.Context, roleID uuid.UUID) (dto.RolePermissions, bool, error) {
	var permissions []uuid.UUID
	rp, err := a.pdb.Queries.GetRolePermissionsForRole(ctx, roleID)
	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
		return dto.RolePermissions{}, false, err
	}
	if err != nil {
		return dto.RolePermissions{}, false, nil
	}
	for _, p := range rp {
		permissions = append(permissions, p.PermissionID)
	}
	return dto.RolePermissions{
		RoleID:      roleID,
		Permissions: permissions,
	}, true, nil
}

func (a *authz) AddRoleToUser(ctx context.Context, roleID, userID uuid.UUID) (dto.AssignRoleToUserReq, error) {
	_, err := a.pdb.Queries.AddRoleToUser(ctx, db.AddRoleToUserParams{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, "%s", err.Error())
		return dto.AssignRoleToUserReq{}, err
	}
	return dto.AssignRoleToUserReq{
		RoleID: roleID,
		UserID: userID,
	}, nil
}

func (a *authz) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, bool, error) {
	var roles []uuid.UUID
	rs, err := a.pdb.Queries.GetUserRoles(ctx, userID)

	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
		return []uuid.UUID{}, false, err
	}

	if err != nil {
		return []uuid.UUID{}, false, nil
	}

	for _, r := range rs {
		roles = append(roles, r.RoleID)
	}
	return roles, true, nil
}

func (a *authz) GetUserRoleUsingUserIDandRole(ctx context.Context, userID, roleID uuid.UUID) (dto.UserRole, bool, error) {
	r, err := a.pdb.Queries.GetUserRoleByUserIDandRoleID(ctx, db.GetUserRoleByUserIDandRoleIDParams{
		RoleID: roleID,
		UserID: userID,
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error(err.Error(), zap.Any("getRoleReqUserID", userID.String()), zap.Any("getRoleReqRoleID", roleID.String()))
		err = errors.ErrUnableToGet.Wrap(err, "%s", err.Error())
		return dto.UserRole{}, false, err
	}

	if err != nil {
		return dto.UserRole{}, false, nil
	}
	return dto.UserRole{
		UserID: r.UserID,
		RoleID: r.RoleID,
	}, true, nil
}

func (a *authz) RemoveRoleFromUserRoles(ctx context.Context, roleID uuid.UUID) error {
	if err := a.pdb.Queries.RemoveRoleFromUserRoles(ctx, roleID); err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, "%s", err.Error())
		return err
	}
	return nil
}

func (a *authz) RevokeUserRole(ctx context.Context, userID, roleID uuid.UUID) error {
	if err := a.pdb.Queries.RevokeUserRole(ctx, db.RevokeUserRoleParams{
		UserID: userID,
		RoleID: roleID,
	}); err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, "%s", err.Error())
		return err
	}
	return nil

}

func (a *authz) RemoveAllUserRolesExceptSuper(ctx context.Context, userID uuid.UUID) error {
	if err := a.pdb.Queries.RemoveAllUserRolesExceptSuper(ctx, userID); err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, "%s", err.Error())
		return err
	}
	return nil
}

func (a *authz) CheckUserHasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM user_roles ur
			JOIN role_permissions rp ON ur.role_id = rp.role_id
			JOIN permissions p ON rp.permission_id = p.id
			WHERE ur.user_id = $1 AND p.name = $2
		) as has_permission`

	var hasPermission bool
	err := a.pdb.GetPool().QueryRow(ctx, query, userID, permissionName).Scan(&hasPermission)
	if err != nil {
		a.log.Error("Error checking user permission", zap.Error(err), zap.String("user_id", userID.String()), zap.String("permission", permissionName))
		return false, errors.ErrUnableToGet.Wrap(err, "failed to check user permission")
	}
	return hasPermission, nil
}

func (a *authz) GetRoleUsers(ctx context.Context, roleID uuid.UUID) ([]dto.User, error) {
	var res []dto.User
	resp, err := a.pdb.GetRoleUsers(ctx, roleID)

	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, "%s", err.Error())
		return []dto.User{}, err
	}

	for _, r := range resp {
		res = append(res, dto.User{
			ID:             r.ID,
			PhoneNumber:    r.PhoneNumber.String,
			FirstName:      r.FirstName.String,
			LastName:       r.LastName.String,
			Email:          r.Email.String,
			ProfilePicture: r.Profile.String,
			DateOfBirth:    r.DateOfBirth.String,
		})
	}

	return res, nil
}
