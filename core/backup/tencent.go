package backup

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	cos "github.com/tencentyun/cos-go-sdk-v5"
)

type TencentProvider struct {
	client *cos.Client
}

func NewTencentProvider() *TencentProvider { return &TencentProvider{} }

func (t *TencentProvider) Name() string { return "tencent" }

func (t *TencentProvider) RequiredCredentials() []CredentialField {
	return []CredentialField{
		{Key: "secret_id", Label: "Secret ID", Secret: false},
		{Key: "secret_key", Label: "Secret Key", Secret: true},
		{Key: "bucket_url", Label: "Bucket URL", Placeholder: "https://mybucket-1250000000.cos.ap-guangzhou.myqcloud.com"},
	}
}

func (t *TencentProvider) Init(creds map[string]string) error {
	u, err := url.Parse(creds["bucket_url"])
	if err != nil {
		return err
	}
	t.client = cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  creds["secret_id"],
			SecretKey: creds["secret_key"],
		},
	})
	return nil
}

func (t *TencentProvider) Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(string)) error {
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
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = t.client.Object.Put(ctx, key, f, nil)
		return err
	})
}

func (t *TencentProvider) Restore(ctx context.Context, bucket, remotePath, localDir string) error {
	var marker string
	for {
		result, _, err := t.client.Bucket.Get(ctx, &cos.BucketGetOptions{
			Prefix: remotePath,
			Marker: marker,
		})
		if err != nil {
			return err
		}
		for _, obj := range result.Contents {
			rel := strings.TrimPrefix(obj.Key, remotePath)
			if ShouldSkipBackupPath(rel) {
				continue
			}
			local := filepath.Join(localDir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(local), 0755); err != nil {
				return err
			}
			_, err := t.client.Object.GetToFile(ctx, obj.Key, local, nil)
			if err != nil {
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

func (t *TencentProvider) List(ctx context.Context, bucket, remotePath string) ([]RemoteFile, error) {
	var files []RemoteFile
	var marker string
	for {
		result, _, err := t.client.Bucket.Get(ctx, &cos.BucketGetOptions{
			Prefix: remotePath,
			Marker: marker,
		})
		if err != nil {
			return nil, err
		}
		for _, obj := range result.Contents {
			rel := strings.TrimPrefix(obj.Key, remotePath)
			if ShouldSkipBackupPath(rel) {
				continue
			}
			files = append(files, RemoteFile{
				Path: rel,
				Size: obj.Size,
			})
		}
		if !result.IsTruncated {
			break
		}
		marker = result.NextMarker
	}
	return files, nil
}
