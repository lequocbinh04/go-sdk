package aws

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	s32 "github.com/aws/aws-sdk-go/service/s3"
)

func (s *s3) ListImageWithPaging(ctx context.Context, prefixPath string) ([]string, error) {
	objs, err := s.service.ListObjectsV2(&s32.ListObjectsV2Input{
		Bucket: &s.cfg.s3Bucket,
		Prefix: aws.String(prefixPath),
	})
	if err != nil {
		return nil, err
	}
	objKey := make([]string, len(objs.Contents))
	for i, item := range objs.Contents {
		objKey[i] = *item.Key
	}
	return objKey, nil
}
