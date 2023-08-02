package goconfig_center

import (
	"fmt"
	"github.com/nova2018/goconfig"
	"github.com/spf13/viper"
	"sync"
)

type center struct {
	config     *goconfig.Config
	listOrigin []*viper.Viper
	listDriver []driver
	listCfg    []*viper.Viper
	lock       *sync.Mutex
	configKey  string
}

func New(config *goconfig.Config, key ...string) *center {
	if len(key) == 0 {
		key = []string{KeyConfig}
	}
	return &center{
		config:     config,
		listOrigin: make([]*viper.Viper, 0),
		listDriver: make([]driver, 0),
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
	newDriver := make([]driver, 0, len(cfg.Drivers))
	// 新生成，还未启动监听
	unWatch := make([]driver, 0, len(cfg.Drivers))
	mapHit := make(map[int]bool)
	for i, dc := range cfg.Drivers {
		if !dc.Enable {
			continue
		}
		if !IsSupport(dc.Driver) {
			return fmt.Errorf("driver[%s] is not supported", dc.Driver)
		}
		vv := v.Sub(fmt.Sprintf("drivers[%d]", i))
		if vv == nil {
			continue
		}
		isHit := false
		for j, x := range c.listDriver {
			if mapHit[j] {
				continue
			}
			if x.IsSame(vv) {
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
			newDriver = append(newDriver, dr)
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
		if !dr.Unwatch() {
			return fmt.Errorf("config unwatch failure! [%s]", dr.Name())
		}
	}
	c.listDriver = newDriver
	return nil
}

func (c *center) PushOrigin(v *viper.Viper) {
	c.listOrigin = append(c.listOrigin, v)
	c.config.AddViper(v)
}
