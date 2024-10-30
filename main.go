package main

import (
	"github.com/blinkbean/dingtalk"
	"github.com/spf13/cobra"
	"github.com/zzj0403/bitwardenBak/config"
	backup2 "github.com/zzj0403/bitwardenBak/internal/backup"
	"github.com/zzj0403/bitwardenBak/pkg/ossx"
	"github.com/zzj0403/bitwardenBak/pkg/ossx/ali"
	"log"
)

var (
	oss            ossx.Oss
	ding           *dingtalk.DingTalk
	backup         *backup2.Backup
	configFilePath string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "help",
		Short: "帮助",
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
	var RestoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "执行数据库迁移",
		Run: func(cmd *cobra.Command, args []string) {
			startRestoreCmd()
		},
	}
	rootCmd.AddCommand(BackupCmd, RestoreCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("启动失败%s", err.Error())
	}

}

func InitApp() {
	loadConfig, err := config.LoadConfig(configFilePath)
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
		log.Fatalf("备份失败：%s", err.Error())
		return
	}
	log.Println("备份成功")
}
func startRestoreCmd() {
	InitApp()
	err := backup.RestoreFromOss()
	if err != nil {
		log.Fatalf("恢复失败：%s", err.Error())
		return
	}
	log.Println("恢复成功")

}
