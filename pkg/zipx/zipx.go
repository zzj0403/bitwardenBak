package zipx

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
)

// ZipDirectory 压缩指定的目录
func ZipDirectory(source string, target string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			fmt.Printf("failed to close zip writer: %v\n", err)
		}
	}()

	var totalFiles int64
	filepath.Walk(source, func(_ string, fi os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking through files: %w", err)
		}
		if !fi.IsDir() {
			totalFiles++
		}
		return nil
	})

	bar := progressbar.NewOptions64(totalFiles, progressbar.OptionSetDescription("压缩中..."))
	err = filepath.Walk(source, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking through files: %w", err)
		}

		relPath, err := filepath.Rel(filepath.Dir(source), file)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		if fi.IsDir() {
			if _, err := zipWriter.Create(relPath + "/"); err != nil {
				return fmt.Errorf("failed to create directory entry in zip: %w", err)
			}
			return nil
		}

		if err := addFileToZip(zipWriter, relPath, file); err != nil {
			return err
		}

		bar.Add(1)
		return nil
	})
	fmt.Println()
	return err
}

// addFileToZip 将文件添加到 ZIP
func addFileToZip(zipWriter *zip.Writer, relPath string, file string) error {
	zipFileWriter, err := zipWriter.Create(relPath)
	if err != nil {
		return fmt.Errorf("failed to create zip entry for %s: %w", relPath, err)
	}

	srcFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", file, err)
	}
	defer srcFile.Close()

	if _, err = io.Copy(zipFileWriter, srcFile); err != nil {
		return fmt.Errorf("failed to write file %s to zip: %w", file, err)
	}

	return nil
}

// UnzipFile 解压指定的 ZIP 文件到目标目录
func UnzipFile(source string, target string) error {
	zipFile, err := zip.OpenReader(source)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer zipFile.Close()

	bar := progressbar.NewOptions64(int64(len(zipFile.File)), progressbar.OptionSetDescription("解压中..."))

	for _, file := range zipFile.File {
		if err := extractFile(file, target); err != nil {
			return err
		}
		bar.Add(1)
	}

	// 在进度条完成后输出换行
	fmt.Println()
	return nil
}

// extractFile 从 ZIP 文件中提取单个文件
func extractFile(file *zip.File, target string) error {
	filePath := filepath.Join(target, file.Name)

	if file.FileInfo().IsDir() {
		return os.MkdirAll(filePath, os.ModePerm)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", filePath, err)
	}

	outFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("failed to open file %s for writing: %w", filePath, err)
	}
	defer outFile.Close()

	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file %s in zip: %w", file.Name, err)
	}
	defer rc.Close()

	if _, err = io.Copy(outFile, rc); err != nil {
		return fmt.Errorf("failed to copy content to %s: %w", filePath, err)
	}

	return nil
}

// ZipDirectoryToIo 压缩指定的目录，并返回压缩文件的 io.Reader
func ZipDirectoryToIo(source string, password string) (io.Reader, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// 计算总文件数以设置进度条
	totalFiles, err := countFiles(source)
	if err != nil {
		return nil, err
	}

	bar := progressbar.NewOptions64(totalFiles, progressbar.OptionSetDescription("压缩中..."))

	// 压缩文件
	if err := filepath.Walk(source, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(filepath.Dir(source), file)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			if _, err := zipWriter.Create(relPath + "/"); err != nil {
				return fmt.Errorf("failed to create zip entry for directory %s: %w", relPath, err)
			}
			return nil
		}

		if err := addFileToZip(zipWriter, relPath, file); err != nil {
			return err
		}

		// 更新进度条
		bar.Add(1)
		return nil
	}); err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	// 在进度条完成后输出换行
	if err := bar.Finish(); err != nil {
		return nil, err
	}

	fmt.Println() // 确保输出换行

	// 加密压缩文件内容
	if password != "" {
		encryptedContent, err := encryptFileContent(&buf, password)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(encryptedContent), nil
	}
	return &buf, nil // 返回压缩后的内容
}

// countFiles 计算指定目录中的文件数量
func countFiles(source string) (int64, error) {
	var totalFiles int64
	err := filepath.Walk(source, func(_ string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			totalFiles++
		}
		return nil
	})
	return totalFiles, err
}

// GenerateRandomPassword 生成随机密码
func GenerateRandomPassword(length int) (string, error) {
	passwordBytes := make([]byte, length)
	if _, err := rand.Read(passwordBytes); err != nil {
		return "", fmt.Errorf("failed to generate random password: %w", err)
	}
	return base64.StdEncoding.EncodeToString(passwordBytes)[:length], nil
}

// DecryptFileContent 解密文件内容
func DecryptFileContent(encryptedData []byte, password string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(password))
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	if len(encryptedData) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := encryptedData[:aes.BlockSize]
	ciphertext := encryptedData[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}

// / encryptFileContent 使用 AES 加密文件内容
func encryptFileContent(file io.Reader, password string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(password))
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// 生成随机初始化向量 (IV)
	ciphertext := make([]byte, aes.BlockSize+len(password))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to read random bytes: %w", err)
	}

	// 加密
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(password))

	return ciphertext, nil
}
