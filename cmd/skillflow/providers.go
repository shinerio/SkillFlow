package main

import (
	"github.com/shinerio/skillflow/core/backup"
	"github.com/shinerio/skillflow/core/registry"
)

func registerProviders() {
	registry.RegisterCloudProvider(backup.NewAliyunProvider())
	registry.RegisterCloudProvider(backup.NewTencentProvider())
	registry.RegisterCloudProvider(backup.NewHuaweiProvider())
	registry.RegisterCloudProvider(backup.NewGitProvider())
}
