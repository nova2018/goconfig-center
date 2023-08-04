package goconfig_center

import (
	"github.com/spf13/viper"
	"log"
	"time"
)

func newWatch(c *Center, info driverInfo) *configWatch {
	return &configWatch{
		driver: info,
		center: c,
	}
}

type configWatch struct {
	lastHash string
	driver   driverInfo
	center   *Center
	closed   bool
}

func (w *configWatch) isUpdate() bool {
	newViper, err := w.driver.driver.GetViper()
	if err != nil {
		log.Fatalf("watch config failure! driver=%s err=%+v", w.driver.driver.Name(), err)
		return false
	}
	newHash := genHash(newViper)
	if newHash != w.lastHash {
		w.lastHash = newHash
		return true
	}
	return false
}

func (w *configWatch) reload() {
	newViper, _ := w.driver.driver.GetViper()
	w.lastHash = genHash(newViper)
}

func watch(cWatch *configWatch) {
	if cWatch.closed || cWatch.driver.driver == nil || cWatch.center == nil {
		return
	}

	go func() {
		driver := cWatch.driver.driver
		onChange := driver.OnUpdate()
		for !cWatch.closed {
			select {
			case <-onChange:
				cWatch.center.wLock.Lock()
				if cWatch.isUpdate() {
					flushConfig(cWatch.center)
				}
				cWatch.center.wLock.Unlock()
			}
		}
	}()
}

func unWatch(cWatch *configWatch) {
	cWatch.closed = true
	_ = cWatch.driver.driver.Close()
}

func scanWatch(c *Center) {
	go func() {
		ticker := time.Tick(time.Second)
		for c.isWatch {
			select {
			case <-ticker:
				c.wLock.Lock()
				isUpdate := false
				for _, x := range c.listWatch {
					if x.isUpdate() {
						isUpdate = true
						break
					}
				}
				if isUpdate {
					flushConfig(c)
				}
				c.wLock.Unlock()
			}
		}
	}()
}

func flushConfig(c *Center) {
	c.fLock.Lock()
	defer c.fLock.Unlock()
	newViper := viper.New()

	for _, w := range c.listOrigin {
		if w != nil {
			mergeConfig(newViper, w)
		}
	}

	lockEnable := c.wLock.TryLock()
	for _, w := range c.listWatch {
		vv, _ := w.driver.driver.GetViper()
		if vv != nil {
			mergeConfig(newViper, vv, w.driver.driver.Prefix())
		}
	}
	if lockEnable {
		c.wLock.Unlock()
	}

	hash := genHash(c.lastViper)
	newHash := genHash(newViper)
	if hash == newHash {
		// 无变化，则跳过
		return
	}

	lockEnable = c.wLock.TryLock()
	for _, w := range c.listWatch {
		w.reload()
	}
	if lockEnable {
		c.wLock.Unlock()
	}
	c.config.SetConfig(newViper)
	c.lastViper = newViper
}
