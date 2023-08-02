package goconfig_center

import (
	"github.com/nova2018/goconfig"
	"github.com/spf13/viper"
)

type driver interface {
	Name() string
	Viper() *viper.Viper
	IsSame(*viper.Viper) bool
	Watch() bool
	Unwatch() bool
	Prefix() string
}

type DriverFactory func(config *goconfig.Config, cfg *viper.Viper) (driver, error)
