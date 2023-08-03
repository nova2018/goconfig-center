package goconfig_center

import (
	"fmt"
	"github.com/nova2018/goconfig"
	"github.com/spf13/viper"
	"sync"
)

type center struct {
	config     *goconfig.Config
	listDriver []driverInfo
	lock       *sync.Mutex
	configKey  string
}

type driverInfo struct {
	driver Driver
	cfg    *viper.Viper
}

func New(config *goconfig.Config, key ...string) *center {
	if len(key) == 0 {
		key = []string{KeyConfig}
	}
	return &center{
		config:     config,
		listDriver: make([]driverInfo, 0),
		lock:       &sync.Mutex{},
		configKey:  key[0],
	}
}

func (c *center) Watch() {
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
}

func (c *center) reload() error {
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
	newDriver := make([]driverInfo, 0, len(cfg.Drivers))
	// 新生成，还未启动监听
	unWatch := make([]Driver, 0, len(cfg.Drivers))
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
		for j, x := range c.listDriver {
			if mapHit[j] {
				continue
			}
			if goconfig.Equal(vv, x.cfg) {
				mapHit[j] = true
				newDriver = append(newDriver, x)
				isHit = true
				break
			}
		}
		if !isHit {
			dr, err := factory(c.config, dc.Driver, vv)
			if err != nil {
				return err
			}
			newDriver = append(newDriver, driverInfo{
				driver: dr,
				cfg:    vv,
			})
			unWatch = append(unWatch, dr)
		}
	}
	// 启动监听
	for _, dr := range unWatch {
		if !dr.Watch() {
			return fmt.Errorf("config watch failure! [%s]", dr.Name())
		}
	}
	// 不再监听
	for i, dr := range c.listDriver {
		if mapHit[i] {
			continue
		}
		if !dr.driver.Unwatch() {
			return fmt.Errorf("config unwatch failure! [%s]", dr.driver.Name())
		}
	}
	c.listDriver = newDriver
	return nil
}
