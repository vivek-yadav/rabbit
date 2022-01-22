package settings

var RunningConfig Config

type Config struct {
	Service ServiceConfig `mapstructure:"service"`
	Db      DbConfig      `mapstructure:"db"`
}

type ServiceConfig struct {
	Port string `mapstructure:"port"`
}
type DbConfig struct {
	Url string `mapstructure:"url"`
}
