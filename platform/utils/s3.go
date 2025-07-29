package utils

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type S3 struct {
	S3Client        *s3.S3
	log             *zap.Logger
	validExtentions []string
}

func NewS3Instance(log *zap.Logger, validExtentions []string) *S3 {
	awsAccessKey := viper.GetString("aws.bucket.accessKey")
	awsSecretAccessKey := viper.GetString("aws.bucket.secretAccessKey")
	awsRegion := viper.GetString("aws.bucket.region")
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretAccessKey, ""),
	})
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	s3Client := s3.New(sess)
	return &S3{
		S3Client:        s3Client,
		log:             log,
		validExtentions: validExtentions,
	}
}
func (c *S3) UploadToS3Bucket(bucketName string, file multipart.File, filename, contentType string) (string, error) {
	sanitizedFileName := sanitizeFileName(filename)
	fileExtention := filepath.Ext(sanitizedFileName)
	if !c.isValidFile(fileExtention) {
		c.log.Error("invalid file extentions ", zap.Any("extentions", fileExtention))
		err := fmt.Errorf("invalid Extentions %s", fileExtention)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return "", err
	}
   tag:="public"
	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(sanitizedFileName),
		Body:        file,
		ContentType: aws.String(contentType),
		Tagging: &tag,
	}
	_, err := c.S3Client.PutObject(input)
	if err != nil {
		c.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return "", err
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, sanitizedFileName), nil

}
func sanitizeFileName(filename string) string {
	return strings.Replace(filename, " ", "_", -1)
}
func (c *S3) isValidFile(extention string) bool {
	for _, ext := range c.validExtentions {
		if strings.ToLower(extention) == ext {
			return true
		}
	}
	return false
}
func  GetFromS3Bucket(bucketName, filename string) string {
	sanitizedFileName := sanitizeFileName(filename)
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, sanitizedFileName)
}
