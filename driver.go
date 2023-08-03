package goconfig_center

import (
	"github.com/nova2018/goconfig"
	"github.com/spf13/viper"
)

type Driver interface {
	Name() string
	Watch() bool
	Unwatch() bool
}

type DriverFactory func(config *goconfig.Config, cfg *viper.Viper) (Driver, error)
