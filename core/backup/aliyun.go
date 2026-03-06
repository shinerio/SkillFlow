package backup

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type AliyunProvider struct {
	client *oss.Client
}

func NewAliyunProvider() *AliyunProvider { return &AliyunProvider{} }

func (a *AliyunProvider) Name() string { return "aliyun" }

func (a *AliyunProvider) RequiredCredentials() []CredentialField {
	return []CredentialField{
		{Key: "access_key_id", Label: "Access Key ID", Secret: false},
		{Key: "access_key_secret", Label: "Access Key Secret", Secret: true},
		{Key: "endpoint", Label: "Endpoint", Placeholder: "oss-cn-hangzhou.aliyuncs.com"},
	}
}

func (a *AliyunProvider) Init(creds map[string]string) error {
	client, err := oss.New(creds["endpoint"], creds["access_key_id"], creds["access_key_secret"])
	if err != nil {
		return err
	}
	a.client = client
	return nil
}

func (a *AliyunProvider) Sync(_ context.Context, localDir, bucket, remotePath string, onProgress func(string)) error {
	b, err := a.client.Bucket(bucket)
	if err != nil {
		return err
	}
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(localDir, path)
		if ShouldSkipBackupPath(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			return nil
		}
		key := remotePath + strings.ReplaceAll(rel, string(filepath.Separator), "/")
		if onProgress != nil {
			onProgress(rel)
		}
		return b.PutObjectFromFile(key, path)
	})
}

func (a *AliyunProvider) Restore(_ context.Context, bucket, remotePath, localDir string) error {
	b, err := a.client.Bucket(bucket)
	if err != nil {
		return err
	}
	marker := ""
	for {
		result, err := b.ListObjects(oss.Prefix(remotePath), oss.Marker(marker))
		if err != nil {
			return err
		}
		for _, obj := range result.Objects {
			rel := strings.TrimPrefix(obj.Key, remotePath)
			if ShouldSkipBackupPath(rel) {
				continue
			}
			local := filepath.Join(localDir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(local), 0755); err != nil {
				return err
			}
			if err := b.GetObjectToFile(obj.Key, local); err != nil {
				return err
			}
		}
		if !result.IsTruncated {
			break
		}
		marker = result.NextMarker
	}
	return nil
}

func (a *AliyunProvider) List(_ context.Context, bucket, remotePath string) ([]RemoteFile, error) {
	b, err := a.client.Bucket(bucket)
	if err != nil {
		return nil, err
	}
	result, err := b.ListObjects(oss.Prefix(remotePath))
	if err != nil {
		return nil, err
	}
	var files []RemoteFile
	for _, obj := range result.Objects {
		rel := strings.TrimPrefix(obj.Key, remotePath)
		if ShouldSkipBackupPath(rel) {
			continue
		}
		files = append(files, RemoteFile{
			Path: rel,
			Size: obj.Size,
		})
	}
	return files, nil
}
