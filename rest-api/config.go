package main

import (
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path"
)

type Config struct {
	SharedSecret   string
	Server         string
	Zone           string
	Domain         string
	NsupdateBinary string
	RecordTTL      int
}

func (conf *Config) LoadConfig(configFile string) {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		homeDir, _ := homedir.Dir()
		defaultConfigPath := path.Join(homeDir, ".goddns")
		_ = os.MkdirAll(defaultConfigPath, os.ModePerm)
		viper.AddConfigPath(defaultConfigPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName("goddns")
	}
	viper.AutomaticEnv()
	if viper.GetBool("verbose") {
		log.SetLevel(log.DebugLevel)
	}
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = viper.SafeWriteConfig()
			if err != nil {
				log.Warningf("error while writing default config file: %v\n", err)
			}
		} else {
			log.Warningf("error while reading config file: %v\n", err)
		}
	}
}
