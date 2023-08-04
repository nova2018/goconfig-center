package goconfig_center

import (
	"github.com/spf13/viper"
)

func factory(drName string, cfg *viper.Viper) (Driver, error) {
	if fn, ok := mapDriver[drName]; ok {
		return fn(cfg)
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
