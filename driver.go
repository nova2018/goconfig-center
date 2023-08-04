package goconfig_center

import (
	"github.com/spf13/viper"
)

type Driver interface {
	Name() string
	GetViper() (*viper.Viper, error)
	OnUpdate() <-chan struct{}
	Close() error
	Prefix() string
}

type DriverFactory func(cfg *viper.Viper) (Driver, error)
