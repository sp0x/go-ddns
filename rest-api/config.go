package main

import (
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path"
)

type Config struct {
	Secret         string
	Server         string
	Zone           string
	Domain         string
	NsupdateBinary string
	RecordTTL      int
}

func (conf *Config) loadConfig(configFile string) {
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
	conf.setDefaults()
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

	conf.Secret = viper.GetString("secret")
	conf.Server = viper.GetString("nsupdate.server")
	conf.Zone = viper.GetString("zone")
	conf.Domain = viper.GetString("domain")
	conf.NsupdateBinary = viper.GetString("nsupdate.path")
	conf.RecordTTL = viper.GetInt("ttl")
}

func (conf *Config) setDefaults() {
	viper.SetDefault("ttl", 300)
}
