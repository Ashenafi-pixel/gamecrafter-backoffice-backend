package banner

import (
	"context"
	"database/sql"
	"math"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type banner struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Banner {
	return &banner{
		db:  db,
		log: log,
	}
}

func (b *banner) GetAllBanners(ctx context.Context, req dto.GetBannersReq) (dto.GetBannersRes, error) {
	offset := (req.Page - 1) * req.PerPage

	banners, err := b.db.Queries.GetAllBanners(ctx, db.GetAllBannersParams{
		Limit:  int32(req.PerPage),
		Offset: int32(offset),
	})
	if err != nil {
		b.log.Error("unable to get paginated banners", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get paginated banners")
		return dto.GetBannersRes{}, err
	}

	var result dto.GetBannersRes
	var total int

	if len(banners) > 0 {
		total = int(banners[0].Total)
	}

	result.TotalCount = total
	result.CurrentPage = req.Page
	result.TotalPages = int(math.Ceil(float64(total) / float64(req.PerPage)))

	result.Banners = make([]dto.Banner, len(banners))
	for i, banner := range banners {
		result.Banners[i] = dto.Banner{
			ID:        banner.ID,
			Page:      banner.Page,
			PageURL:   banner.PageUrl,
			ImageURL:  banner.ImageUrl,
			Headline:  banner.Headline,
			Tagline:   banner.Tagline.String,
			UpdatedAt: dto.GetTimePtrFromNullTime(banner.UpdatedAt),
		}
	}

	return result, nil
}

func (b *banner) GetBannerByPage(ctx context.Context, req dto.GetBannerReq) (dto.Banner, bool, error) {
	banner, err := b.db.Queries.GetBannerByPage(ctx, req.Page)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.Banner{}, false, nil
		}
		b.log.Error("unable to get banner by page", zap.Error(err), zap.String("page", req.Page))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get banner by page")
		return dto.Banner{}, false, err
	}

	return dto.Banner{
		ID:        banner.ID,
		Page:      banner.Page,
		PageURL:   banner.PageUrl,
		ImageURL:  banner.ImageUrl,
		Headline:  banner.Headline,
		Tagline:   banner.Tagline.String,
		UpdatedAt: dto.GetTimePtrFromNullTime(banner.UpdatedAt),
	}, true, nil
}

func (b *banner) GetBannerByID(ctx context.Context, bannerID uuid.UUID) (dto.Banner, error) {
	banner, err := b.db.Queries.GetBannerByID(ctx, bannerID)
	if err != nil {
		if err == sql.ErrNoRows {
			err := errors.ErrResourceNotFound.New("no banner found with ID: %s", bannerID.String())
			b.log.Warn("no banner found", zap.String("bannerID", bannerID.String()))
			return dto.Banner{}, err
		}
		b.log.Error("unable to get banner by ID", zap.Error(err), zap.String("bannerID", bannerID.String()))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get banner by ID")
		return dto.Banner{}, err
	}

	return dto.Banner{
		ID:        banner.ID,
		Page:      banner.Page,
		PageURL:   banner.PageUrl,
		ImageURL:  banner.ImageUrl,
		Headline:  banner.Headline,
		Tagline:   banner.Tagline.String,
		UpdatedAt: dto.GetTimePtrFromNullTime(banner.UpdatedAt),
	}, nil
}

func (b *banner) UpdateBanner(ctx context.Context, req dto.UpdateBannerReq) (dto.Banner, error) {
	updatedBanner, err := b.db.Queries.UpdateBanner(ctx, db.UpdateBannerParams{
		ID:       req.ID,
		PageUrl:  req.PageURL,
		Headline: req.Headline,
		Tagline:  sql.NullString{String: req.Tagline, Valid: req.Tagline != ""},
		ImageUrl: req.ImageURL,
	})

	if err != nil {
		b.log.Error("unable to update banner", zap.Error(err), zap.String("bannerID", req.ID.String()))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to update banner")
		return dto.Banner{}, err
	}

	return dto.Banner{
		ID:        updatedBanner.ID,
		Page:      updatedBanner.Page,
		PageURL:   updatedBanner.PageUrl,
		ImageURL:  updatedBanner.ImageUrl,
		Headline:  updatedBanner.Headline,
		Tagline:   updatedBanner.Tagline.String,
		UpdatedAt: dto.GetTimePtrFromNullTime(updatedBanner.UpdatedAt),
	}, nil
}

func (b *banner) CreateBanner(ctx context.Context, req dto.CreateBannerReq) (dto.Banner, error) {
	createParams := db.CreateBannerParams{
		Page:     req.Page,
		PageUrl:  req.PageURL,
		ImageUrl: req.ImageURL,
		Headline: req.Headline,
		Tagline:  sql.NullString{String: req.Tagline, Valid: req.Tagline != ""},
	}

	newBanner, err := b.db.Queries.CreateBanner(ctx, createParams)
	if err != nil {
		b.log.Error("unable to create banner", zap.Error(err), zap.Any("request", req))
		err = errors.ErrUnableTocreate.Wrap(err, "unable to create banner")
		return dto.Banner{}, err
	}

	return dto.Banner{
		ID:        newBanner.ID,
		Page:      newBanner.Page,
		PageURL:   newBanner.PageUrl,
		ImageURL:  newBanner.ImageUrl,
		Headline:  newBanner.Headline,
		Tagline:   newBanner.Tagline.String,
		UpdatedAt: dto.GetTimePtrFromNullTime(newBanner.UpdatedAt),
	}, nil
}

func (b *banner) DeleteBanner(ctx context.Context, bannerID uuid.UUID) error {
	err := b.db.Queries.DeleteBanner(ctx, bannerID)
	if err != nil {
		b.log.Error("unable to delete banner", zap.Error(err), zap.String("bannerID", bannerID.String()))
		err = errors.ErrDBDelError.Wrap(err, "unable to delete banner")
		return err
	}

	b.log.Info("banner deleted successfully", zap.String("bannerID", bannerID.String()))
	return nil
}
