package banner

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

type banner struct {
	log           *zap.Logger
	bannerStorage storage.Banner
	bucketName    string
}

func Init(bannerStorage storage.Banner, log *zap.Logger, bucketName string) module.Banner {
	return &banner{
		log:           log,
		bannerStorage: bannerStorage,
		bucketName:    bucketName,
	}
}

func (b *banner) GetAllBanners(ctx context.Context, req dto.GetBannersReq) (dto.GetBannersRes, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}

	banners, err := b.bannerStorage.GetAllBanners(ctx, req)
	if err != nil {
		b.log.Error("unable to get paginated banners", zap.Error(err))
		return dto.GetBannersRes{}, err
	}

	return banners, nil
}

func (b *banner) GetBannerByPage(ctx context.Context, req dto.GetBannerReq) (dto.Banner, error) {
	if req.Page == "" {
		err := errors.ErrInvalidUserInput.New("page parameter is required")
		b.log.Error("invalid request: page parameter is required")
		return dto.Banner{}, err
	}

	banner, _, err := b.bannerStorage.GetBannerByPage(ctx, req)
	if err != nil {
		return dto.Banner{}, err
	}

	return banner, nil
}

func (b *banner) UpdateBanner(ctx context.Context, req dto.UpdateBannerReq) (dto.Banner, error) {
	if req.ID == uuid.Nil {
		err := errors.ErrInvalidUserInput.New("banner ID is required")
		b.log.Error("invalid request: banner ID is required")
		return dto.Banner{}, err
	}

	if err := dto.ValidateUpdateBanner(req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		b.log.Error("invalid request: validation failed", zap.Error(err))
		return dto.Banner{}, err
	}

	existingBanner, err := b.bannerStorage.GetBannerByID(ctx, req.ID)
	if err != nil {
		b.log.Error("unable to get banner by ID", zap.Error(err), zap.String("bannerID", req.ID.String()))
		return dto.Banner{}, err
	}

	if req.PageURL == "" {
		req.PageURL = existingBanner.PageURL
	}
	if req.Headline == "" {
		req.Headline = existingBanner.Headline
	}
	if req.ImageURL == "" {
		req.ImageURL = existingBanner.ImageURL
	}
	if req.Tagline == "" {
		req.Tagline = existingBanner.Tagline
	}

	updatedBanner, err := b.bannerStorage.UpdateBanner(ctx, req)
	if err != nil {
		return dto.Banner{}, err
	}

	return updatedBanner, nil
}

func (b *banner) CreateBanner(ctx context.Context, req dto.CreateBannerReq) (dto.Banner, error) {
	if err := dto.ValidateCreateBanner(req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		b.log.Error("invalid request: validation failed", zap.Error(err))
		return dto.Banner{}, err
	}

	// Check if banner already exists for this page
	_, exists, _ := b.bannerStorage.GetBannerByPage(ctx, dto.GetBannerReq{Page: req.Page})
	if exists {
		err := fmt.Errorf("banner already exists for page: %s", req.Page)
		b.log.Warn("banner already exists", zap.String("page", req.Page))
		err = errors.ErrDataAlredyExist.Wrap(err, err.Error())
		return dto.Banner{}, err
	}

	newBanner, err := b.bannerStorage.CreateBanner(ctx, req)
	if err != nil {
		return dto.Banner{}, err
	}

	return newBanner, nil
}

func (b *banner) DeleteBanner(ctx context.Context, bannerID uuid.UUID) error {
	if bannerID == uuid.Nil {
		err := errors.ErrInvalidUserInput.New("banner ID is required")
		b.log.Error("invalid request: banner ID is required")
		return err
	}

	_, err := b.bannerStorage.GetBannerByID(ctx, bannerID)
	if err != nil {
		return err
	}

	err = b.bannerStorage.DeleteBanner(ctx, bannerID)
	if err != nil {
		return err
	}

	return nil
}

func (b *banner) UploadBannerImage(ctx context.Context, img multipart.File, header *multipart.FileHeader) (dto.UploadBannerImageResp, error) {
	// Extract the original file name and get the extension
	fileExtension := filepath.Ext(header.Filename)
	if fileExtension == "" {
		err := fmt.Errorf("invalid file extension")
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UploadBannerImageResp{}, err
	}

	bannerImageName := uuid.New().String() + fileExtension

	// Create S3 instance
	s3Instance := utils.NewS3Instance(b.log, constant.VALID_IMGS)
	if s3Instance == nil {
		err := fmt.Errorf("unable to create s3 session")
		b.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UploadBannerImageResp{}, err
	}

	_, err := s3Instance.UploadToS3Bucket(b.bucketName, img, bannerImageName, header.Header.Get("Content-Type"))
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UploadBannerImageResp{}, err
	}

	return dto.UploadBannerImageResp{
		Status: constant.SUCCESS,
		Url:    fmt.Sprintf("https://%s.s3.amazonaws.com/%s", b.bucketName, bannerImageName),
	}, nil
}
