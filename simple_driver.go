package goconfig_center

import (
	"github.com/nova2018/goconfig"
	"github.com/spf13/viper"
)

type simpleDriver struct {
	cfg    *viper.Viper
	viper  *viper.Viper
	config *goconfig.Config
	name   string
	prefix string
}

func (e *simpleDriver) Name() string {
	return e.name
}

func (e *simpleDriver) Viper() *viper.Viper {
	return e.viper
}

func (e *simpleDriver) Watch() bool {
	e.config.AddWatchViper(goconfig.WatchRemote, e.Viper(), e.Prefix())
	return true
}

func (e *simpleDriver) Unwatch() bool {
	e.config.DelViper(e.Viper())
	return true
}

func (e *simpleDriver) Prefix() string {
	return e.prefix
}

func (e *simpleDriver) IsSame(v *viper.Viper) bool {
	return goconfig.Equal(e.cfg, v)
}

type simpleConfig struct {
	ConfigDriver `mapstructure:",squash"`
	Endpoint     string `mapstructure:"endpoint"`
	Path         string `mapstructure:"path"`
	Prefix       string `mapstructure:"prefix"`
	Type         string `mapstructure:"type"`
}

func simpleFactory(config *goconfig.Config, cfg *viper.Viper) (Driver, error) {
	var c simpleConfig
	if err := cfg.Unmarshal(&c); err != nil {
		return nil, err
	}

	v := viper.New()
	if err := v.AddRemoteProvider(c.Driver, c.Endpoint, c.Path); err != nil {
		return nil, err
	}
	v.SetConfigType(c.Type)

	return &simpleDriver{
		cfg:    cfg,
		viper:  v,
		config: config,
		prefix: c.Prefix,
		name:   c.Driver,
	}, nil

}

func init() {
	Register("etcd", simpleFactory)
	Register("etcd3", simpleFactory)
	Register("consul", simpleFactory)
	Register("firestore", simpleFactory)
}
