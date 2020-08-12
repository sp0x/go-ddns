package config

import (
	"fmt"
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
	DnsProvider    string
}

func (conf *Config) Load(configFile string) {
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
	log.SetLevel(log.InfoLevel)
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

	conf.Secret = viper.GetString("api_key")
	conf.Server = viper.GetString("nsupdate.server")
	conf.Zone = viper.GetString("zone")
	conf.Domain = viper.GetString("domain")
	conf.NsupdateBinary = viper.GetString("nsupdate.path")
	conf.RecordTTL = viper.GetInt("ttl")
	conf.DnsProvider = viper.GetString("provider")
	validateDnsProvider(conf)
}

func validateDnsProvider(c *Config) {
	dns := c.DnsProvider
	switch dns {
	case "nsupdate":
		if c.NsupdateBinary == "" {
			c.NsupdateBinary = findNsupdate()
		}
		if c.NsupdateBinary == "" {
			fmt.Print("nsupdate binary is not set")
			os.Exit(1)
		}
	case "google":
		break
	default:
		fmt.Printf("dns provider `%s` is not supported", c.DnsProvider)
		os.Exit(1)
	}
}

func findNsupdate() string {
	//maybe use the $PATH to resolve the absolute path to nsupdate?
	return "nsupdate"
}

func (conf *Config) setDefaults() {
	viper.SetDefault("ttl", 300)
	viper.SetDefault("provider", "nsupdate")
}
