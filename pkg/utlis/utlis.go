package utlis

import (
	"fmt"
	"log"
	"os"
)

// EnsureDirExists 检查目录是否存在，如果不存在则创建
func EnsureDirExists(dir string) error {
	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// 目录不存在，创建目录
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("创建目录 %s 时出错: %w", dir, err)
		}
		log.Printf("目录 %s 已创建\n", dir)
	} else if err != nil {
		return fmt.Errorf("检查目录 %s 时出错: %w", dir, err)
	}
	return nil
}
