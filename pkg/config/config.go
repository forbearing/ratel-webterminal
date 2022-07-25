package config

import (
	"os"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Config = &RatelTerminalConf{}

type RatelTerminalConf struct {
	Port           int    `mapstructure:"port"`
	BindAddress    string `mapstructure:"bindAddress"`
	KubeConfigFile string `mapstructure:"kubeConfigFile"`
	LogLevel       string `mapstructure:"logLevel"`
	LogFormat      string `mapstructure:"logFormat"`
	LogFile        string `mapstructure:"logFile"`
}

func Init(filename string) error {
	// if config file not exist, return error.
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return err
	}

	viper.SetConfigFile(filename)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(Config); err != nil {
		return err
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Debug(e.Op.String, e.Name)
		if err = viper.Unmarshal(Config); err != nil {
			return
		}
	})

	if err != nil {
		return err
	}
	return nil
}
