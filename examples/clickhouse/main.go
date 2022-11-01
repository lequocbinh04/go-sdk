/*
 * @author           Viet Tran <viettranx@gmail.com>
 * @copyright        2020 200lab <core@200lab.io>
 * @license          Apache-2.0
 */

package main

import (
	goservice "github.com/lequocbinh04/go-sdk"
	"github.com/lequocbinh04/go-sdk/plugin/storage/sdkclickhouse"
)

func main() {
	service := goservice.New(
		goservice.WithName("demo"),
		goservice.WithVersion("1.0.0"),
		goservice.WithInitRunnable(sdkclickhouse.NewClickHouseDB("clickhouse", "")),
	)

	service.Init()
	_ = service.Start()
}
