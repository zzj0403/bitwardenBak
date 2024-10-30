package ali

import (
	"bytes"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/schollz/progressbar/v3"
	"github.com/zzj0403/bitwardenBak/pkg/ossx"
	"io"
)

var (
	_alioss ossx.Oss
)

// AliOssConfig holds the configuration for Aliyun OSS.
type OssConfig struct {
	Endpoint        string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
	AccessKeyId     string `mapstructure:"access_key_id" json:"access_key_id" yaml:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret" json:"access_key_secret" yaml:"access_key_secret"`
	BucketName      string `mapstructure:"bucket_name" json:"bucket_name" yaml:"bucket_name"`
	Expired         int64  `mapstructure:"expired" json:"expired" yaml:"expired"`
}

// DefaultAliOss returns a default configuration for Aliyun OSS.
func DefaultAliOss() *OssConfig {
	return &OssConfig{
		Endpoint:        "",
		AccessKeyId:     "",
		AccessKeySecret: "",
		BucketName:      "",
		Expired:         0,
	}
}

// Init initializes the OSS client with the provided configuration.
func Init(conf *OssConfig) ossx.Oss {
	if conf == nil {
		conf = DefaultAliOss()
	}
	if _alioss == nil {
		var err error
		_alioss, err = NewAliOss(conf)
		if err != nil {
			panic(err) // Consider handling the error more gracefully in production code
		}
	}
	return _alioss
}

// AliOss is the implementation of the Oss interface.
type AliOss struct {
	client *oss.Client
	bucket *oss.Bucket
	config *OssConfig
}

// NewAliOss creates a new AliOss instance.
func NewAliOss(conf *OssConfig) (*AliOss, error) {
	client, err := oss.New(conf.Endpoint, conf.AccessKeyId, conf.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("error creating OSS client: %w", err)
	}

	bucket, err := client.Bucket(conf.BucketName)
	if err != nil {
		return nil, fmt.Errorf("error getting OSS bucket: %w", err)
	}

	return &AliOss{
		client: client,
		bucket: bucket,
		config: conf,
	}, nil
}

// PutFile uploads a file to OSS and returns the file's URL.
func (a *AliOss) PutFile(filename string, file io.Reader) (string, error) {
	// 获取文件内容到缓冲区
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return "", fmt.Errorf("error reading file content: %w", err)
	}

	// 创建进度条
	bar := progressbar.NewOptions64(int64(buf.Len()), progressbar.OptionSetDescription("上传中..."))

	// 创建一个管道
	pr, pw := io.Pipe()

	// 启动一个 goroutine 来写入数据并更新进度条
	go func() {
		defer pw.Close()
		for _, b := range buf.Bytes() {
			if _, err := pw.Write([]byte{b}); err != nil {
				return
			}
			bar.Add(1) // 更新进度条
		}
	}()

	// 上传文件到 OSS
	if err := a.bucket.PutObject(filename, pr); err != nil {
		return "", fmt.Errorf("error putting file to OSS: %w", err)
	}

	// 上传完成后换行
	fmt.Println()

	// 获取文件的 URL
	url, err := a.getUrl(filename)
	if err != nil {
		return "", err
	}
	return url, nil
}

// getUrl generates a signed URL for the uploaded file.
func (a *AliOss) getUrl(filename string) (string, error) {
	url, err := a.bucket.SignURL(filename, oss.HTTPGet, a.config.Expired)
	if err != nil {
		return "", fmt.Errorf("error signing URL: %w", err)
	}
	return url, nil
}
