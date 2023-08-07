package goconfig_center

import (
	"fmt"
	"github.com/nova2018/goconfig"
	"github.com/spf13/viper"
	"sync"
)

type Center struct {
	config     *goconfig.Config
	lastViper  *viper.Viper
	listWatch  []*configWatch
	lock       *sync.Mutex
	fLock      *sync.Mutex // lock for flush
	wLock      *sync.Mutex // lock for watch
	configKey  string
	isWatch    bool
	listOrigin []*viper.Viper
}

type driverInfo struct {
	driver Driver
	cfg    *viper.Viper
}

func New(config *goconfig.Config, key ...string) *Center {
	if len(key) == 0 {
		key = []string{KeyConfig}
	}
	return &Center{
		config:    config,
		listWatch: make([]*configWatch, 0),
		lock:      &sync.Mutex{},
		wLock:     &sync.Mutex{},
		fLock:     &sync.Mutex{},
		configKey: key[0],
	}
}

func NewWithViper(v *viper.Viper, key ...string) *Center {
	gConfig := goconfig.New()
	newCenter := New(gConfig, key...)
	newCenter.Attach(v)
	return newCenter
}

func (c *Center) Attach(v *viper.Viper) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.listOrigin == nil {
		c.listOrigin = make([]*viper.Viper, 0, 1)
	}
	c.listOrigin = append(c.listOrigin, v)
	flushConfig(c)
}

func (c *Center) Detach(v *viper.Viper) {
	c.lock.Lock()
	defer c.lock.Unlock()
	listNew := make([]*viper.Viper, 0, len(c.listOrigin))
	for _, x := range c.listOrigin {
		if v != x && v != nil {
			listNew = append(listNew, x)
		}
	}
	c.listOrigin = listNew
	flushConfig(c)
}

func (c *Center) Watch() {
	c.config.OnKeyChange(c.configKey, func() {
		err := c.reload()
		if err != nil {
			print(err)
		}
	})
	err := c.reload()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
	c.isWatch = true
}

func (c *Center) StopWatch() {
	c.isWatch = false
	for _, x := range c.listWatch {
		unWatch(x)
	}
}

func (c *Center) GetConfig() *goconfig.Config {
	return c.config
}

func (c *Center) reload() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	var cfg config

	v := c.config.GetConfig().Sub(c.configKey)
	if v != nil {
		if err := v.Unmarshal(&cfg); err != nil {
			return err
		}
	}
	if !cfg.Enable {
		cfg.Drivers = []ConfigDriver{}
	}

	// 新的driver列表
	newDriver := make([]*configWatch, 0, len(cfg.Drivers))
	// 新生成，还未启动监听
	listUnWatch := make([]*configWatch, 0, len(cfg.Drivers))
	mapHit := make(map[int]bool)
	for i, dc := range cfg.Drivers {
		if !dc.Enable {
			continue
		}
		if !IsSupport(dc.Driver) {
			return fmt.Errorf("driver[%s] is not supported", dc.Driver)
		}
		vv := v.Sub(fmt.Sprintf("drivers.%d", i))
		if vv == nil {
			continue
		}
		isHit := false
		for j, x := range c.listWatch {
			if mapHit[j] {
				continue
			}
			if Equal(vv, x.driver.cfg) {
				mapHit[j] = true
				newDriver = append(newDriver, x)
				isHit = true
				break
			}
		}
		if !isHit {
			dr, err := factory(dc.Driver, vv)
			if err != nil {
				return err
			}
			di := newWatch(c, driverInfo{
				driver: dr,
				cfg:    vv,
			})
			newDriver = append(newDriver, di)
			listUnWatch = append(listUnWatch, di)
		}
	}
	// 启动监听
	for _, dr := range listUnWatch {
		watch(dr)
	}
	// 不再监听
	for i, dr := range c.listWatch {
		if mapHit[i] {
			continue
		}
		unWatch(dr)
	}
	c.listWatch = newDriver
	flushConfig(c)
	return nil
}
