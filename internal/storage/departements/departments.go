package departements

import (
	"context"
	"database/sql"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/db"
	"github.com/joshjones612/egyptkingcrash/internal/constant/persistencedb"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"go.uber.org/zap"
)

type departements struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Departements {
	return &departements{
		db:  db,
		log: log,
	}
}

func (d *departements) CreateDepartement(ctx context.Context, dep dto.CreateDepartementReq) (dto.CreateDepartementRes, error) {

	depRes, err := d.db.Queries.CreateDepartment(ctx, db.CreateDepartmentParams{
		Name:          dep.Name,
		Notifications: dep.Notifications,
		CreatedAt:     sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		d.log.Error(err.Error(), zap.Any("dep_req", dep))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateDepartementRes{}, err
	}
	return dto.CreateDepartementRes{
		ID:            depRes.ID,
		Name:          depRes.Name,
		Notifications: depRes.Notifications,
		CreatedAt:     time.Now(),
	}, nil
}

func (d *departements) GetDepartmentsByName(ctx context.Context, name string) (dto.CreateDepartementRes, bool, error) {
	deps, err := d.db.Queries.GetDepartementByName(ctx, name)
	if err != nil && err.Error() != dto.ErrNoRows {
		d.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.CreateDepartementRes{}, false, err
	}
	if err != nil {
		return dto.CreateDepartementRes{}, false, nil
	}
	return dto.CreateDepartementRes{
		ID:            deps.ID,
		Name:          deps.Name,
		Notifications: deps.Notifications,
		CreatedAt:     deps.CreatedAt.Time,
	}, true, nil
}

func (d *departements) GetDepartmentsByID(ctx context.Context, id uuid.UUID) (dto.CreateDepartementRes, bool, error) {
	deps, err := d.db.Queries.GetDepartementByID(ctx, id)
	if err != nil && err.Error() != dto.ErrNoRows {
		d.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.CreateDepartementRes{}, false, err
	}
	if err != nil {
		return dto.CreateDepartementRes{}, false, nil
	}
	return dto.CreateDepartementRes{
		ID:            deps.ID,
		Name:          deps.Name,
		Notifications: deps.Notifications,
		CreatedAt:     deps.CreatedAt.Time,
	}, true, nil
}

func (d *departements) GetDepartments(ctx context.Context, getDepReq dto.GetDepartementsReq) (dto.GetDepartementsRes, bool, error) {
	var depsRes dto.GetDepartementsRes

	deps, err := d.db.Queries.GetAllDepatments(ctx, db.GetAllDepatmentsParams{
		Limit:  int32(getDepReq.PerPage),
		Offset: int32(getDepReq.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		d.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetDepartementsRes{}, false, err
	}

	if err != nil {
		return dto.GetDepartementsRes{}, false, nil
	}
	for _, dep := range deps {
		ps := float64(dep.Total / int64(getDepReq.PerPage))
		total := int(math.Ceil(ps))
		depsRes.Departements = append(depsRes.Departements, dto.Departement{
			ID:            dep.ID,
			Name:          dep.Name,
			Notifications: dep.Notifications,
			CreatedAt:     dep.CreatedAt.Time,
		})
		depsRes.TotalPages = total
	}

	return depsRes, true, nil
}

func (d *departements) UpdateDepartment(ctx context.Context, dep dto.UpdateDepartment) (dto.UpdateDepartment, error) {
	updatedDep, err := d.db.Queries.UpdateDepartments(ctx, db.UpdateDepartmentsParams{
		Name:          dep.Name,
		Notifications: dep.Notifications,
		ID:            dep.ID,
	})
	if err != nil {
		d.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.UpdateDepartment{}, err
	}
	return dto.UpdateDepartment{
		ID:            updatedDep.ID,
		Name:          updatedDep.Name,
		Notifications: updatedDep.Notifications,
		CreatedAt:     updatedDep.CreatedAt.Time,
	}, nil
}

func (d *departements) AssignUserToDepartment(ctx context.Context, assignDep dto.AssignDepartmentToUserReq) (dto.AssignDepartmentToUserResp, error) {
	assignRes, err := d.db.Queries.AssignUserToDepartment(ctx, db.AssignUserToDepartmentParams{
		UserID:       assignDep.UserID,
		DepartmentID: assignDep.DepartmentID,
	})
	if err != nil {
		d.log.Error(err.Error(), zap.Any("assignDepReq", assignDep))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.AssignDepartmentToUserResp{}, err
	}
	return dto.AssignDepartmentToUserResp{
		Message:      constant.SUCCESS,
		ID:           assignRes.ID,
		UserID:       assignRes.UserID,
		DepartmentID: assignRes.DepartmentID,
	}, nil
}
