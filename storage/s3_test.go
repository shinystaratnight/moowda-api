package storage

import (
	"bytes"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

type mockedUploader struct{}

func (u mockedUploader) Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	return &s3manager.UploadOutput{Location: *input.Key}, nil
}

func TestStore(t *testing.T) {
	b := ioutil.NopCloser(bytes.NewReader(make([]byte, 1024)))
	s := s3Storage{uploader: mockedUploader{}}

	url, err := s.Store(nil, "test.png", b)
	assert.Nil(t, err)
	assert.Len(t, url, 36)
}
