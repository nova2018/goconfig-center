package goconfig_center

import (
	"fmt"
	"github.com/spf13/viper"
	"sync"
	"time"
)

type remoteConfig struct {
	ConfigDriver `mapstructure:",squash"`
	Endpoint     string `mapstructure:"endpoint"`
	Path         string `mapstructure:"path"`
	Prefix       string `mapstructure:"prefix"`
	Type         string `mapstructure:"type"`
}

type remoteDriver struct {
	cfg      *remoteConfig
	viper    *viper.Viper
	close    bool
	onUpdate chan struct{}
	lock     *sync.Mutex
	once     *sync.Once
}

func (r *remoteDriver) Name() string {
	return r.cfg.Driver
}

func (r *remoteDriver) GetViper() (*viper.Viper, error) {
	return r.viper, nil
}

func (r *remoteDriver) OnUpdate() <-chan struct{} {
	if r.close {
		return nil
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.onUpdate == nil {
		r.onUpdate = make(chan struct{})
		go func() {
			ticker := time.Tick(time.Second)
			for !r.close {
				select {
				case <-ticker:
					if r.viper == nil {
						break
					}
					e := r.viper.WatchRemoteConfig()
					if e != nil {
						fmt.Printf("%s viper remote listen failure! err=%v\n", time.Now().Format("2006-01-02 15:04:05"), e)
						continue
					}
					r.lock.Lock()
					if r.onUpdate != nil {
						r.onUpdate <- struct{}{}
					}
					r.lock.Unlock()
				}
			}
		}()
	}
	return r.onUpdate
}

func (r *remoteDriver) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.close {
		return nil
	}
	r.close = true
	if r.onUpdate != nil {
		close(r.onUpdate)
		r.onUpdate = nil
	}
	return nil
}

func (r *remoteDriver) Prefix() string {
	return r.cfg.Prefix
}

func remoteFactory(cfg *viper.Viper) (Driver, error) {
	var c remoteConfig
	if err := cfg.Unmarshal(&c); err != nil {
		return nil, err
	}
	v := viper.New()
	if err := v.AddRemoteProvider(c.Driver, c.Endpoint, c.Path); err != nil {
		return nil, err
	}
	v.SetConfigType(c.Type)
	err := v.ReadRemoteConfig()
	if err != nil {
		return nil, err
	}
	return &remoteDriver{
		cfg:   &c,
		viper: v,
		lock:  &sync.Mutex{},
		once:  &sync.Once{},
	}, nil
}

func init() {
	Register("etcd", remoteFactory)
	Register("etcd3", remoteFactory)
	Register("consul", remoteFactory)
	Register("firestore", remoteFactory)
}
