package backup

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	obs "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

type HuaweiProvider struct {
	client *obs.ObsClient
}

func NewHuaweiProvider() *HuaweiProvider { return &HuaweiProvider{} }

func (h *HuaweiProvider) Name() string { return "huawei" }

func (h *HuaweiProvider) RequiredCredentials() []CredentialField {
	return []CredentialField{
		{Key: "access_key_id", Label: "Access Key ID", Secret: false},
		{Key: "secret_access_key", Label: "Secret Access Key", Secret: true},
		{Key: "endpoint", Label: "Endpoint", Placeholder: "obs.cn-north-4.myhuaweicloud.com"},
	}
}

func (h *HuaweiProvider) Init(creds map[string]string) error {
	client, err := obs.New(creds["access_key_id"], creds["secret_access_key"], creds["endpoint"])
	if err != nil {
		return err
	}
	h.client = client
	return nil
}

func (h *HuaweiProvider) Sync(_ context.Context, localDir, bucket, remotePath string, onProgress func(string)) error {
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
		input := &obs.PutFileInput{}
		input.Bucket = bucket
		input.Key = key
		input.SourceFile = path
		_, err = h.client.PutFile(input)
		return err
	})
}

func (h *HuaweiProvider) Restore(_ context.Context, bucket, remotePath, localDir string) error {
	input := &obs.ListObjectsInput{Bucket: bucket}
	input.Prefix = remotePath
	for {
		result, err := h.client.ListObjects(input)
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
			getInput := &obs.GetObjectInput{}
			getInput.Bucket = bucket
			getInput.Key = obj.Key
			resp, err := h.client.GetObject(getInput)
			if err != nil {
				return err
			}
			f, err := os.Create(local)
			if err != nil {
				resp.Body.Close()
				return err
			}
			_, err = f.ReadFrom(resp.Body)
			resp.Body.Close()
			f.Close()
			if err != nil {
				return err
			}
		}
		if !result.IsTruncated {
			break
		}
		input.Marker = result.NextMarker
	}
	return nil
}

func (h *HuaweiProvider) List(_ context.Context, bucket, remotePath string) ([]RemoteFile, error) {
	input := &obs.ListObjectsInput{Bucket: bucket}
	input.Prefix = remotePath
	var files []RemoteFile
	for {
		result, err := h.client.ListObjects(input)
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
		input.Marker = result.NextMarker
	}
	return files, nil
}
