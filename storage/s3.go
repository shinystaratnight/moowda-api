package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/labstack/echo"

	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

const requiredConfigs = "key,secret,region,bucket"

// config is like {"key":"access key","secret":"secret key","session":"token", "region":"region", "bucket": "bucket"}
func NewS3Storage(config string) (FileStorage, error) {
	var cf map[string]string

	err := json.Unmarshal([]byte(config), &cf)
	if err != nil {
		return nil, err
	}
	required := strings.Split(requiredConfigs, ",")

	for _, el := range required {
		if _, ok := cf[el]; !ok {
			return nil, fmt.Errorf("storage config has no \"%v\" key", el)
		}
	}

	s := &s3Storage{}

	session, err := awsSession.NewSession(&aws.Config{
		Region:      aws.String(cf["region"]),
		Credentials: credentials.NewStaticCredentials(cf["key"], cf["secret"], cf["session"]),
	})

	s.uploader = s3manager.NewUploader(session)
	s.manager = s3.New(session)
	s.bucket = cf["bucket"]

	if err != nil {
		return nil, err
	}
	return s, nil
}

type Uploader interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

type s3Storage struct {
	uploader Uploader
	manager  *s3.S3
	bucket   string
}

func (s *s3Storage) Store(c echo.Context, name string, sourceFile io.ReadCloser) (string, error) {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		fmt.Print(err)
	}

	key := fmt.Sprintf("%s%s", hex.EncodeToString(randBytes), filepath.Ext(name))

	result, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   sourceFile,
	})
	if err != nil {
		return "", err
	}

	return result.Location, nil
}

func init() {
	Register("s3", NewS3Storage)
}
