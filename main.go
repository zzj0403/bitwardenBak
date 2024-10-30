package main

import (
	"context"
	"github.com/blinkbean/dingtalk"
	"github.com/spf13/cobra"
	"github.com/zzj0403/bitwardenBak/config"
	backup2 "github.com/zzj0403/bitwardenBak/internal/backup"
	"github.com/zzj0403/bitwardenBak/pkg/ossx"
	"github.com/zzj0403/bitwardenBak/pkg/ossx/ali"
	"google.golang.org/appengine/log"
)

var (
	oss    ossx.Oss
	ding   *dingtalk.DingTalk
	backup *backup2.Backup
)

func main() {
	var configFilePath string
	var rootCmd = &cobra.Command{
		Use:   "",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	rootCmd.PersistentFlags().StringVarP(&configFilePath, "config", "c", "config.yaml", "配置文件路径")
	var BackupCmd = &cobra.Command{
		Use:   "backup",
		Short: "执行数据库迁移",
		Run: func(cmd *cobra.Command, args []string) {
			startBackup()
		},
	}
	rootCmd.AddCommand(BackupCmd)
}

func InitApp() {
	loadConfig, err := config.LoadConfig("loadConfig.yaml")
	if err != nil {
		return
	}
	oss = ali.Init(loadConfig.AliOss)

	ding = dingtalk.InitDingTalkWithSecret(loadConfig.Ding.RobotToken, loadConfig.Ding.Secret)
	backup = backup2.NewBackup(loadConfig, oss, ding)
}

func startBackup() {
	InitApp()
	err := backup.BackupToOss()
	if err != nil {
		log.Errorf(context.Background(), "备份失败：%s", err.Error())
		return
	}
}
