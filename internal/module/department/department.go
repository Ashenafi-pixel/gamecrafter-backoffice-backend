package department

import (
	"context"
	"fmt"

	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type department struct {
	departmentsStorage storage.Departements
	userStorage        storage.User
	log                *zap.Logger
}

func Init(departmentsStorage storage.Departements, userStorage storage.User, log *zap.Logger) module.Departements {
	return &department{
		departmentsStorage: departmentsStorage,
		log:                log,
		userStorage:        userStorage,
	}
}

func (d *department) CreateDepartement(ctx context.Context, deps dto.CreateDepartementReq) (dto.CreateDepartementRes, error) {
	// validate input
	if err := dto.ValidateCreateDepartement(deps); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateDepartementRes{}, err
	}

	// check if the departement already exist
	_, exist, err := d.departmentsStorage.GetDepartmentsByName(ctx, deps.Name)
	if err != nil {
		return dto.CreateDepartementRes{}, err
	}
	if exist {
		err := fmt.Errorf("department with this name already exist")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateDepartementRes{}, err
	}

	// create departments
	return d.departmentsStorage.CreateDepartement(ctx, deps)
}

func (d *department) GetDepartments(ctx context.Context, depReq dto.GetDepartementsReq) (dto.GetDepartementsRes, error) {
	if depReq.Page <= 0 || depReq.PerPage <= 0 {
		err := errors.ErrInvalidUserInput.Wrap(fmt.Errorf("please provide page and per_page"), "please provide page and per_page")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetDepartementsRes{}, err
	}
	offset := (depReq.Page - 1) * depReq.PerPage
	depReq.Page = offset
	depsRes, _, err := d.departmentsStorage.GetDepartments(ctx, depReq)
	if err != nil {
		return dto.GetDepartementsRes{}, err
	}
	return depsRes, nil
}

func (d *department) UpdateDepartment(ctx context.Context, dep dto.UpdateDepartment) (dto.UpdateDepartment, error) {
	// get department by id
	depRes, exist, err := d.departmentsStorage.GetDepartmentsByID(ctx, dep.ID)
	if err != nil {
		return dto.UpdateDepartment{}, err
	}
	if !exist {
		err := fmt.Errorf("unable to get department ")
		d.log.Error(err.Error(), zap.Any("departmentID", dep.ID.String()))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UpdateDepartment{}, err
	}
	if dep.Name != "" {
		depRes.Name = dep.Name
	}
	if len(dep.Notifications) > 0 {
		// validate department notifications
		for _, notf := range dep.Notifications {
			if notf != constant.BLOCK_TYPE_GAMING && notf != constant.BLOCK_TYPE_FINANCIAL && notf != constant.BLOCK_TYPE_LOGIN && notf != constant.BLOCK_TYPE_COMPLETE {
				err = fmt.Errorf("invalid notification type")
				err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
				return dto.UpdateDepartment{}, err
			}
		}
		depRes.Notifications = dep.Notifications
	}

	return d.departmentsStorage.UpdateDepartment(ctx, dto.UpdateDepartment(depRes))
}

func (d *department) AssignUserToDepartment(ctx context.Context, assignReq dto.AssignDepartmentToUserReq) (dto.AssignDepartmentToUserResp, error) {
	//get department by id
	_, exist, err := d.departmentsStorage.GetDepartmentsByID(ctx, assignReq.DepartmentID)
	if err != nil {
		return dto.AssignDepartmentToUserResp{}, err
	}

	if !exist {
		err := fmt.Errorf("unable to find department")
		d.log.Error(err.Error(), zap.Any("assign_req", assignReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.AssignDepartmentToUserResp{}, err
	}

	// get user by user id
	_, exist, err = d.userStorage.GetUserByID(ctx, assignReq.UserID)
	if err != nil {
		return dto.AssignDepartmentToUserResp{}, err
	}
	if !exist {
		err := fmt.Errorf("unable to find user")
		d.log.Error(err.Error(), zap.Any("assign_req", assignReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.AssignDepartmentToUserResp{}, err
	}
	return d.departmentsStorage.AssignUserToDepartment(ctx, assignReq)
}
