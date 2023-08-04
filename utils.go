package goconfig_center

import (
	"bytes"
	"crypto/md5"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"log"
)

func Equal(v1, v2 *viper.Viper) bool {
	if v1 == v2 {
		return true
	}
	if v1 == nil || v2 == nil {
		return v1 == v2
	}
	return genHash(v1) == genHash(v2)
}

func genHash(v *viper.Viper) string {
	if v == nil {
		// 如果为空则返回空字符串
		return ""
	}
	c := v.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("unable to marshal config to YAML: %v", err)
	}
	hash := md5.Sum(bs)
	return string(hash[:])
}

func mergeConfig(viper, sourceViper *viper.Viper, prefix ...string) {
	if len(prefix) == 0 {
		prefix = []string{""}
	}
	viper.SetConfigType("yml")
	cfg := sourceViper.AllSettings()
	if prefix[0] != "" {
		cfg = map[string]interface{}{
			prefix[0]: cfg,
		}
	}
	bs, err := yaml.Marshal(cfg)
	if err != nil {
		log.Fatalf("unable to marshal config to YAML: %v", err)
	}
	_ = viper.MergeConfig(bytes.NewBuffer(bs))
}
