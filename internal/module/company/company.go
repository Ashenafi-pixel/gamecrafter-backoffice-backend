package company

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type company struct {
	log            *zap.Logger
	companyStorage storage.Company
}

func Init(companyStorage storage.Company, log *zap.Logger) module.Company {
	return &company{
		log:            log,
		companyStorage: companyStorage,
	}
}

func (c *company) CreateCompany(ctx context.Context, req dto.CreateCompanyReq) (dto.CreateCompanyRes, error) {
	if err := dto.ValidateCreateCompany(req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateCompanyRes{}, err
	}

	return c.companyStorage.CreateCompany(ctx, req)
}

func (c *company) GetCompanyByID(ctx context.Context, id uuid.UUID) (dto.Company, error) {
	if id == uuid.Nil {
		err := errors.ErrInvalidUserInput.New("invalid company ID")
		return dto.Company{}, err
	}

	company, _, err := c.companyStorage.GetCompanyByID(ctx, id)
	if err != nil {
		return dto.Company{}, err
	}

	return company, nil
}

func (c *company) GetCompanies(ctx context.Context, req dto.GetCompaniesReq) (dto.GetCompaniesRes, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}

	offset := (req.Page - 1) * req.PerPage

	req.Page = offset

	companiesRes, err := c.companyStorage.GetCompanies(ctx, req)
	if err != nil {
		return dto.GetCompaniesRes{}, err
	}
	return companiesRes, nil
}

func (c *company) UpdateCompany(ctx context.Context, req dto.UpdateCompanyReq) (dto.UpdateCompanyRes, error) {
	cmp, exist, err := c.companyStorage.GetCompanyByID(ctx, req.ID)
	if err != nil {
		return dto.UpdateCompanyRes{}, err
	}
	if !exist {
		err := fmt.Errorf("unable to get Company ")
		c.log.Error(err.Error(), zap.Any("companyID", req.ID.String()))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UpdateCompanyRes{}, err
	}
	if err := dto.ValidateUpdateCompany(req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UpdateCompanyRes{}, err
	}

	if req.SiteName == "" {
		req.SiteName = cmp.SiteName
	}
	if req.SupportEmail == "" {
		req.SupportEmail = cmp.SupportEmail
	}
	if req.SupportPhone == "" {
		req.SupportPhone = cmp.SupportPhone
	}

	if req.MaximumLoginAttempt == 0 {
		req.MaximumLoginAttempt = cmp.MaximumLoginAttempt
	}

	if req.PasswordExpiry == 0 {
		req.PasswordExpiry = cmp.PasswordExpiry
	}
	if req.LockoutDuration == 0 {
		req.LockoutDuration = cmp.LockoutDuration
	}
	if !req.RequireTwoFactorAuthentication {
		req.RequireTwoFactorAuthentication = cmp.RequireTwoFactorAuthentication
	}
	if !req.MaintenanceMode {
		req.MaintenanceMode = cmp.MaintenanceMode
	}

	return c.companyStorage.UpdateCompany(ctx, req)
}

func (c *company) DeleteCompany(ctx context.Context, companyID uuid.UUID) error {
	if companyID == uuid.Nil {
		err := errors.ErrInvalidUserInput.New("invalid company ID")
		return err
	}

	if err := c.companyStorage.DeleteCompany(ctx, companyID); err != nil {
		return err
	}

	return nil
}

func (c *company) AddIP(ctx context.Context, companyID uuid.UUID, ip string) (dto.UpdateCompanyRes, error) {
	req := dto.AddCompanyIPReq{IpAddr: ip}
	if err := dto.ValidateAddCompanyIP(req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UpdateCompanyRes{}, err
	}
	return c.companyStorage.AddIP(ctx, companyID, ip)
}
