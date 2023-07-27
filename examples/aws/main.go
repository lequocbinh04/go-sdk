package main

import (
	"context"
	"fmt"
	goservice "github.com/lequocbinh04/go-sdk"
	"github.com/lequocbinh04/go-sdk/plugin/aws"
)

func main() {
	service := goservice.New(
		goservice.WithName("demo"),
		goservice.WithVersion("1.0.0"),
		goservice.WithInitRunnable(aws.New("aws")),
	)
	_ = service.Init()

	s3 := service.MustGet("aws").(aws.S3)
	url, err := s3.ListImageWithPaging(context.Background(), "image/space/64c2457721ce22a8d87865dc")

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(url)
}
