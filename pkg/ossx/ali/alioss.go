package ali

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/zzj0403/bitwardenBak/pkg/ossx"
	"io"
	"sort"
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
	if err := a.bucket.PutObject(filename, file, oss.Progress(&OssProgressListener{})); err != nil {
		return "", fmt.Errorf("error putting file to OSS: %w", err)
	}
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

// 定义进度条监听器。
type OssProgressListener struct {
}

// 定义进度变更事件处理函数。
func (listener *OssProgressListener) ProgressChanged(event *oss.ProgressEvent) {
	switch event.EventType {
	case oss.TransferStartedEvent:
		fmt.Printf("Transfer Started, ConsumedBytes: %d, TotalBytes %d.\n",
			event.ConsumedBytes, event.TotalBytes)
	case oss.TransferDataEvent:
		fmt.Printf("\rTransfer Data, ConsumedBytes: %d, TotalBytes %d, %d%%.",
			event.ConsumedBytes, event.TotalBytes, event.ConsumedBytes*100/event.TotalBytes)
	case oss.TransferCompletedEvent:
		fmt.Printf("\nTransfer Completed, ConsumedBytes: %d, TotalBytes %d.\n",
			event.ConsumedBytes, event.TotalBytes)
	case oss.TransferFailedEvent:
		fmt.Printf("\nTransfer Failed, ConsumedBytes: %d, TotalBytes %d.\n",
			event.ConsumedBytes, event.TotalBytes)
	default:
	}
}

// 获取指定目录下所有子目录的信息
func (a *AliOss) DirFilesList(dir string) (FilesList []oss.ObjectProperties, err error) {
	continueToken := ""
	prefix := dir
	lsRes, err := a.bucket.ListObjectsV2(oss.Prefix(prefix), oss.ContinuationToken(continueToken))
	if err != nil {
		return nil, err
	}
	// 打印列举结果。
	for _, object := range lsRes.Objects {
		FilesList = append(FilesList, object)
	}
	// 如果结果被截断，则更新继续标记并继续循环
	if lsRes.IsTruncated {
		continueToken = lsRes.NextContinuationToken
	} else {
		return FilesList, nil
	}
	// 根据最后修改时间排序
	sort.Slice(FilesList, func(i, j int) bool {
		return FilesList[i].LastModified.After(FilesList[j].LastModified)
	})

	// 只取最新的 10 个文件
	if len(FilesList) > 10 {
		FilesList = FilesList[:10]
	}

	return FilesList, err
}

// 下载指定文件到本地
func (a *AliOss) DownloadFile(filename string, localPath string) error {
	err := a.bucket.GetObjectToFile(filename, localPath)
	if err != nil {
		return fmt.Errorf("error downloading file from OSS: %w", err)
	}
	return nil
}
