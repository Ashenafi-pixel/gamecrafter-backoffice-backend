package company

import (
	"context"
	"database/sql"
	"github.com/jackc/pgtype"
	"math"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/db"
	"github.com/joshjones612/egyptkingcrash/internal/constant/persistencedb"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"go.uber.org/zap"
)

type company struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Company {
	return &company{
		db:  db,
		log: log,
	}
}

func (c *company) CreateCompany(ctx context.Context, req dto.CreateCompanyReq) (dto.CreateCompanyRes, error) {
	var maxLoginAttempt, passwordExpiry, lockoutDuration sql.NullInt32
	var maintenanceMode, requireTwoFactorAuth sql.NullBool
	var ipList []pgtype.Inet

	maxLoginAttempt = sql.NullInt32{Valid: false}
	if req.MaximumLoginAttempt > 0 {
		maxLoginAttempt = sql.NullInt32{Int32: int32(req.MaximumLoginAttempt), Valid: true}
	}
	passwordExpiry = sql.NullInt32{Valid: false}
	if req.PasswordExpiry > 0 {
		passwordExpiry = sql.NullInt32{Int32: int32(req.PasswordExpiry), Valid: true}
	}
	lockoutDuration = sql.NullInt32{Valid: false}
	if req.LockoutDuration > 0 {
		lockoutDuration = sql.NullInt32{Int32: int32(req.LockoutDuration), Valid: true}
	}

	maintenanceMode = sql.NullBool{Valid: true, Bool: req.MaintenanceMode}
	requireTwoFactorAuth = sql.NullBool{Valid: true, Bool: req.RequireTwoFactorAuthentication}

	if req.IpList != nil {
		ipListPg, err := dto.ToPgtypeInetArray(req.IpList)
		ipList = ipListPg
		if err != nil {
			c.log.Error("invalid IP in list", zap.Error(err))
			return dto.CreateCompanyRes{}, errors.ErrInvalidUserInput.Wrap(err, "invalid IP in list")
		}
	} else {
		ipList = []pgtype.Inet{}
	}

	company, err := c.db.Queries.CreateCompany(ctx, db.CreateCompanyParams{
		SiteName:                       req.SiteName,
		SupportEmail:                   req.SupportEmail,
		SupportPhone:                   req.SupportPhone,
		MaintenanceMode:                maintenanceMode,
		MaximumLoginAttempt:            maxLoginAttempt,
		PasswordExpiry:                 passwordExpiry,
		LockoutDuration:                lockoutDuration,
		RequireTwoFactorAuthentication: requireTwoFactorAuth,
		IpList:                         ipList,
		CreatedBy:                      req.CreatedBy,
	})
	if err != nil {
		c.log.Error(err.Error(), zap.Any("dep_req", company))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.CreateCompanyRes{}, err
	}

	return dto.CreateCompanyRes{
		ID:                             company.ID,
		SiteName:                       company.SiteName,
		SupportEmail:                   company.SupportEmail,
		SupportPhone:                   company.SupportPhone,
		MaintenanceMode:                dto.GetBoolValue(company.MaintenanceMode),
		MaximumLoginAttempt:            dto.GetIntValue(company.MaximumLoginAttempt),
		PasswordExpiry:                 dto.GetIntValue(company.PasswordExpiry),
		LockoutDuration:                dto.GetIntValue(company.LockoutDuration),
		RequireTwoFactorAuthentication: dto.GetBoolValue(company.RequireTwoFactorAuthentication),
		IpList:                         company.IpList,
		CreatedBy:                      company.CreatedBy.String(),
		CreatedAt:                      company.CreatedAt,
		UpdatedAt:                      company.UpdatedAt,
		DeletedAt:                      dto.GetTimePtrFromNullTime(company.DeletedAt),
	}, nil
}

func (c *company) GetCompanyByID(ctx context.Context, id uuid.UUID) (dto.Company, bool, error) {
	company, err := c.db.Queries.GetCompanyByID(ctx, id)
	if err != nil {
		c.log.Error("unable to get company by ID", zap.String("id", id.String()), zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.Company{}, false, err
	}

	return dto.Company{
		ID:                             company.ID,
		SiteName:                       company.SiteName,
		SupportEmail:                   company.SupportEmail,
		SupportPhone:                   company.SupportPhone,
		MaintenanceMode:                dto.GetBoolValue(company.MaintenanceMode),
		MaximumLoginAttempt:            dto.GetIntValue(company.MaximumLoginAttempt),
		PasswordExpiry:                 dto.GetIntValue(company.PasswordExpiry),
		LockoutDuration:                dto.GetIntValue(company.LockoutDuration),
		RequireTwoFactorAuthentication: dto.GetBoolValue(company.RequireTwoFactorAuthentication),
		IpList:                         company.IpList,
		CreatedBy:                      company.CreatedBy.String(),
		CreatedAt:                      company.CreatedAt,
		UpdatedAt:                      company.UpdatedAt,
		DeletedAt:                      dto.GetTimePtrFromNullTime(company.DeletedAt),
	}, true, nil
}

func (c *company) GetCompanies(ctx context.Context, req dto.GetCompaniesReq) (dto.GetCompaniesRes, error) {
	var res dto.GetCompaniesRes

	companies, err := c.db.Queries.GetCompanies(ctx, db.GetCompaniesParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		c.log.Error("unable to get companies", zap.Int("page", req.Page), zap.Int("per_page", req.PerPage), zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetCompaniesRes{}, err
	}
	if err != nil {
		return dto.GetCompaniesRes{}, nil
	}

	var total int
	if len(companies) > 0 {
		total = int(companies[0].Total)
	}
	if req.PerPage > 0 {
		res.TotalPages = int(math.Ceil(float64(total) / float64(req.PerPage)))
	}

	for _, comp := range companies {
		res.Companies = append(res.Companies, dto.Company{
			ID:                             comp.ID,
			SiteName:                       comp.SiteName,
			SupportEmail:                   comp.SupportEmail,
			SupportPhone:                   comp.SupportPhone,
			MaintenanceMode:                comp.MaintenanceMode.Bool,
			MaximumLoginAttempt:            int(comp.MaximumLoginAttempt.Int32),
			PasswordExpiry:                 int(comp.PasswordExpiry.Int32),
			LockoutDuration:                int(comp.LockoutDuration.Int32),
			RequireTwoFactorAuthentication: comp.RequireTwoFactorAuthentication.Bool,
			IpList:                         comp.IpList,
			CreatedBy:                      comp.CreatedBy.String(),
			CreatedAt:                      comp.CreatedAt,
			UpdatedAt:                      comp.UpdatedAt,
			DeletedAt:                      dto.GetTimePtrFromNullTime(comp.DeletedAt),
		})
	}

	return res, nil
}

func (c *company) UpdateCompany(ctx context.Context, req dto.UpdateCompanyReq) (dto.UpdateCompanyRes, error) {
	company, err := c.db.Queries.UpdateCompany(ctx, db.UpdateCompanyParams{
		ID:                             req.ID,
		SiteName:                       req.SiteName,
		SupportEmail:                   req.SupportEmail,
		SupportPhone:                   req.SupportPhone,
		MaintenanceMode:                sql.NullBool{Bool: req.MaintenanceMode, Valid: true},
		MaximumLoginAttempt:            sql.NullInt32{Int32: int32(req.MaximumLoginAttempt), Valid: true},
		PasswordExpiry:                 sql.NullInt32{Int32: int32(req.PasswordExpiry), Valid: true},
		LockoutDuration:                sql.NullInt32{Int32: int32(req.LockoutDuration), Valid: true},
		RequireTwoFactorAuthentication: sql.NullBool{Bool: req.RequireTwoFactorAuthentication, Valid: true},
	})
	if err != nil {
		c.log.Error("unable to update company", zap.String("id", req.ID.String()), zap.Error(err))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.UpdateCompanyRes{}, err
	}

	return dto.UpdateCompanyRes{
		ID:                             company.ID,
		SiteName:                       company.SiteName,
		SupportEmail:                   company.SupportEmail,
		SupportPhone:                   company.SupportPhone,
		MaintenanceMode:                company.MaintenanceMode.Bool,
		MaximumLoginAttempt:            int(company.MaximumLoginAttempt.Int32),
		PasswordExpiry:                 int(company.PasswordExpiry.Int32),
		LockoutDuration:                int(company.LockoutDuration.Int32),
		RequireTwoFactorAuthentication: company.RequireTwoFactorAuthentication.Bool,
		IpList:                         company.IpList,
		CreatedBy:                      company.CreatedBy.String(),
		CreatedAt:                      company.CreatedAt,
		UpdatedAt:                      company.UpdatedAt,
		DeletedAt:                      dto.GetTimePtrFromNullTime(company.DeletedAt),
	}, nil
}

func (c *company) DeleteCompany(ctx context.Context, id uuid.UUID) error {
	err := c.db.Queries.DeleteCompany(ctx, id)
	if err != nil {
		c.log.Error("unable to delete company", zap.String("id", id.String()), zap.Error(err))
		err = errors.ErrDBDelError.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (c *company) AddIP(ctx context.Context, companyID uuid.UUID, ip string) (dto.UpdateCompanyRes, error) {
	company, err := c.db.Queries.AddIPAddressToCompany(ctx, db.AddIPAddressToCompanyParams{
		ID:            companyID,
		ArrayPosition: ip,
	})
	if err != nil {
		c.log.Error("unable to add IP to company", zap.String("company_id", companyID.String()), zap.String("ip", ip), zap.Error(err))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.UpdateCompanyRes{}, err
	}

	return dto.UpdateCompanyRes{
		ID:                             company.ID,
		SiteName:                       company.SiteName,
		SupportEmail:                   company.SupportEmail,
		SupportPhone:                   company.SupportPhone,
		MaintenanceMode:                company.MaintenanceMode.Bool,
		MaximumLoginAttempt:            int(company.MaximumLoginAttempt.Int32),
		PasswordExpiry:                 int(company.PasswordExpiry.Int32),
		LockoutDuration:                int(company.LockoutDuration.Int32),
		RequireTwoFactorAuthentication: company.RequireTwoFactorAuthentication.Bool,
		IpList:                         company.IpList,
		CreatedBy:                      company.CreatedBy.String(),
		CreatedAt:                      company.CreatedAt,
		UpdatedAt:                      company.UpdatedAt,
		DeletedAt:                      dto.GetTimePtrFromNullTime(company.DeletedAt),
	}, nil
}
