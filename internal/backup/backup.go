package backup

import (
	"fmt"
	"github.com/blinkbean/dingtalk"
	"github.com/zzj0403/bitwardenBak/config"
	"github.com/zzj0403/bitwardenBak/pkg/ossx"
	"github.com/zzj0403/bitwardenBak/pkg/zipx"
	"log"
	"time"
)

type Backup struct {
	config *config.Config
	myOss  ossx.Oss
	ding   *dingtalk.DingTalk
}

func NewBackup(config *config.Config, myOss ossx.Oss, ding *dingtalk.DingTalk) *Backup {
	return &Backup{
		config: config,
		myOss:  myOss,
		ding:   ding,
	}
}

func (b *Backup) BackupToOss() error {
	password, err := zipx.GenerateRandomPassword(16)
	outputFilename := b.getOutputFilename("backup")
	if err != nil {
		return err
	}
	log.Printf("正在压缩目录%s", outputFilename)
	io, err := zipx.ZipDirectoryToIo(b.config.BackupDir, password)
	if err != nil {
		log.Printf("压缩目录失败: %v", err)
		return b.ding.SendMarkDownMessage("备份bitwarden失败", "压缩目录失败，请检查日志。")
	}
	// 上传到oss
	url, err := b.myOss.PutFile(outputFilename, io)
	if err != nil {
		log.Printf("上传到 OSS 失败: %v", err)
		return b.ding.SendMarkDownMessage("备份bitwarden失败", "压缩目录失败，请检查日志。")

	}
	// 发送钉钉消息
	message := b.genMarkDownMessage(url, password)
	return b.ding.SendMarkDownMessage("备份bitwarden成功", message)

}

// genMarkDownMessage 生成钉钉消息的 Markdown 格式
func (b *Backup) genMarkDownMessage(url string, password string) string {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("备份成功！\n - 下载链接: [点击下载](%s)\n - 备份时间: %s \n - 解压密码：%s", url, currentTime, password)
}
func (b *Backup) getOutputFilename(outPut string) string {
	// 获取当前时间拼成2024-10-30_15:07_bitwarden.zip
	// 获取当前时间
	currentTime := time.Now()
	// 格式化为所需的字符串格式
	filename := fmt.Sprintf("%s/%s_bitwarden.zip", outPut, currentTime.Format("2006-01-02_15:04"))
	return filename
}
