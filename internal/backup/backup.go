package backup

import (
	"fmt"
	"github.com/blinkbean/dingtalk"
	"github.com/manifoldco/promptui"
	"github.com/zzj0403/bitwardenBak/config"
	"github.com/zzj0403/bitwardenBak/pkg/ossx"
	"github.com/zzj0403/bitwardenBak/pkg/utlis"
	"github.com/zzj0403/bitwardenBak/pkg/zipx"
	"log"
	"path/filepath"
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
	// 压缩文件
	// 获取output名字
	outputFilename := b.getOutputFilename("backup")
	log.Printf("正在压缩目录%s", outputFilename)
	io, err := zipx.ZipDirectoryToIo(b.config.BackupDir)
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
	message := b.genMarkDownMessage(url)
	return b.ding.SendMarkDownMessage("备份bitwarden成功", message)

}

// genMarkDownMessage 生成钉钉消息的 Markdown 格式
func (b *Backup) genMarkDownMessage(url string) string {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("备份成功！\n - 下载链接: [点击下载](%s)\n - 备份时间: %s", url, currentTime)
}
func (b *Backup) getOutputFilename(outPut string) string {
	// 获取当前时间拼成2024-10-30_15:07_bitwarden.zip
	// 获取当前时间
	currentTime := time.Now()
	// 格式化为所需的字符串格式
	filename := fmt.Sprintf("%s/%s_bitwarden.zip", outPut, currentTime.Format("2006-01-02_15:04"))
	return filename
}

func (b *Backup) RestoreFromOss() error {
	// 从oss下载文件
	list, err := b.myOss.DirFilesList("backup")
	if err != nil {
		return err
	} else if len(list) == 0 {
		return fmt.Errorf("没有找到备份文件")
	}
	// 选择文件
	for i, file := range list {
		fmt.Printf("[%d] %s\n", i, file.Key)
	}
	// 创建提示框，显示文件列表
	var fileNames []string
	for _, file := range list {
		fileNames = append(fileNames, file.Key)
	}

	prompt := promptui.Select{
		Label: "选择要下载的文件",
		Items: fileNames,
	}

	_, result, err := prompt.Run()
	log.Printf("您选择了文件: %s", result)
	// 下载文件
	err = utlis.EnsureDirExists(b.config.TmpDir)
	if err != nil {
		log.Fatal(err)
		return err
	}
	localFilePath := filepath.Join(b.config.TmpDir, filepath.Base(result))
	log.Printf("文件将存储在: %s", localFilePath)
	err = b.myOss.DownloadFile(result, localFilePath)
	// 解压文件
	err = zipx.UnzipFile(localFilePath, b.config.TmpDir)
	if err != nil {
		return err
	}
	// 发送钉钉消息
	return nil
}
