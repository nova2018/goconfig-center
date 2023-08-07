package goconfig_center

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"sync"
)

type fileConfig struct {
	ConfigDriver `mapstructure:",squash"`
	Path         string `mapstructure:"path"`
	Prefix       string `mapstructure:"prefix"`
	Type         string `mapstructure:"type"`
}

type fileDriver struct {
	cfg      *fileConfig
	viper    *viper.Viper
	close    bool
	onUpdate chan struct{}
	lock     *sync.Mutex
}

func (r *fileDriver) Name() string {
	return r.cfg.Driver
}

func (r *fileDriver) GetViper() (*viper.Viper, error) {
	return r.viper, nil
}

func (r *fileDriver) OnUpdate() <-chan struct{} {
	if r.close {
		return nil
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.onUpdate == nil {
		r.onUpdate = make(chan struct{})
		r.viper.OnConfigChange(func(e fsnotify.Event) {
			r.lock.Lock()
			if r.onUpdate != nil {
				r.onUpdate <- struct{}{}
			}
			r.lock.Unlock()
		})
		r.viper.WatchConfig()
	}
	return r.onUpdate
}

func (r *fileDriver) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.close {
		return nil
	}
	r.close = true
	if r.onUpdate != nil {
		close(r.onUpdate)
		r.onUpdate = nil
		r.viper.OnConfigChange(nil)
	}
	return nil
}

func (r *fileDriver) Prefix() string {
	return r.cfg.Prefix
}

func fileFactory(cfg *viper.Viper) (Driver, error) {
	var c fileConfig
	if err := cfg.Unmarshal(&c); err != nil {
		return nil, err
	}
	v := viper.New()
	if c.Type != "" {
		v.SetConfigType(c.Type)
	}
	v.AddConfigPath(c.Path)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return &fileDriver{
		cfg:   &c,
		viper: v,
		lock:  &sync.Mutex{},
	}, nil
}

func init() {
	Register("file", fileFactory)
}
