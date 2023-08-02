package goconfig_center

import (
	"github.com/nova2018/goconfig"
	"github.com/spf13/viper"
)

func factory(config *goconfig.Config, drName string, cfg *viper.Viper) (driver, error) {
	if fn, ok := mapDriver[drName]; ok {
		return fn(config, cfg)
	}
	return nil, ErrDriverNotFound
}

var (
	mapDriver = map[string]DriverFactory{}
)

func Register(driverName string, fn DriverFactory) {
	mapDriver[driverName] = fn
}

func IsSupport(driverName string) bool {
	_, ok := mapDriver[driverName]
	return ok
}
