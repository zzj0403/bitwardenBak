package config

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/zzj0403/bitwardenBak/pkg/ossx/ali"
)

type Config struct {
	BackupDir string         `mapstructure:"backup_dir" json:"backup_dir" yaml:"backup_dir"`
	AliOss    *ali.OssConfig `mapstructure:"ali_oss" json:"ali_oss" yaml:"ali_oss"`
	Ding      *DingTalk      `mapstructure:"ding" json:"ding" yaml:"ding"`
}

type DingTalk struct {
	RobotToken string `mapstructure:"robot_token" json:"robot_token" yaml:"robot_token"`
	Secret     string `mapstructure:"secret" json:"secret" yaml:"secret"`
	KeyWord    string `mapstructure:"key_word" json:"key_word" yaml:"key_word"`
}

func LoadConfig(confFile string) (config *Config, err error) {
	v := viper.New()
	v.SetConfigFile(confFile)
	v.SetConfigType("yaml")
	if err = v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("read config failed: %s \n", err))
	}

	if err = v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("Fatal error marshal config file: %s \n", err)
	}
	return config, err
}
