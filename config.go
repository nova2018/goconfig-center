package goconfig_center

const (
	KeyConfig = "config_center"
)

type config struct {
	Enable  bool           `mapstructure:"'enable'" json:"enable,omitempty"`
	Drivers []ConfigDriver `mapstructure:"drivers,omitempty"`
}

type ConfigDriver struct {
	Driver string `mapstructure:"Driver"`
	Enable bool   `mapstructure:"enable"`
}
