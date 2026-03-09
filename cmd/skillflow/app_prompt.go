package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shinerio/skillflow/core/prompt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) promptStorage() (*prompt.Storage, string, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return nil, "", err
	}
	root := filepath.Join(a.backupRootDir(cfg), "prompts")
	return prompt.NewStorage(root), root, nil
}

func (a *App) ListPrompts() ([]*prompt.Prompt, error) {
	store, _, err := a.promptStorage()
	if err != nil {
		return nil, err
	}
	return store.ListAll()
}

func (a *App) ListPromptCategories() ([]string, error) {
	store, _, err := a.promptStorage()
	if err != nil {
		return nil, err
	}
	categories, err := store.ListCategories()
	if err != nil {
		return nil, err
	}
	hasDefault := false
	for _, category := range categories {
		if category == prompt.DefaultCategoryName {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		categories = append([]string{prompt.DefaultCategoryName}, categories...)
	}
	return categories, nil
}

func (a *App) CreatePrompt(name, description, category, content string) (*prompt.Prompt, error) {
	store, root, err := a.promptStorage()
	if err != nil {
		a.logErrorf("prompt create failed: load storage failed: %v", err)
		return nil, err
	}
	a.logInfof("prompt create started: prompt=%s category=%s root=%s", name, category, root)
	item, err := store.Create(name, description, category, content)
	if err != nil {
		a.logErrorf("prompt create failed: prompt=%s category=%s root=%s err=%v", name, category, root, err)
		return nil, err
	}
	a.logInfof("prompt create completed: prompt=%s category=%s file=%s", item.Name, item.Category, item.FilePath)
	a.scheduleAutoBackup()
	return item, nil
}

func (a *App) UpdatePrompt(originalName, name, description, category, content string) (*prompt.Prompt, error) {
	store, root, err := a.promptStorage()
	if err != nil {
		a.logErrorf("prompt update failed: prompt=%s load storage failed: %v", originalName, err)
		return nil, err
	}
	a.logInfof("prompt update started: prompt=%s next=%s category=%s root=%s", originalName, name, category, root)
	item, err := store.Update(originalName, name, description, category, content)
	if err != nil {
		a.logErrorf("prompt update failed: prompt=%s next=%s category=%s root=%s err=%v", originalName, name, category, root, err)
		return nil, err
	}
	a.logInfof("prompt update completed: prompt=%s category=%s file=%s", item.Name, item.Category, item.FilePath)
	a.scheduleAutoBackup()
	return item, nil
}

func (a *App) MovePromptCategory(name string, category string) error {
	store, root, err := a.promptStorage()
	if err != nil {
		a.logErrorf("prompt move category failed: prompt=%s category=%s load storage failed: %v", name, category, err)
		return err
	}
	a.logInfof("prompt move category started: prompt=%s category=%s root=%s", name, category, root)
	if err := store.MoveCategory(name, category); err != nil {
		a.logErrorf("prompt move category failed: prompt=%s category=%s root=%s err=%v", name, category, root, err)
		return err
	}
	a.logInfof("prompt move category completed: prompt=%s category=%s", name, category)
	a.scheduleAutoBackup()
	return nil
}

func (a *App) DeletePrompt(name string) error {
	store, root, err := a.promptStorage()
	if err != nil {
		a.logErrorf("prompt delete failed: prompt=%s load storage failed: %v", name, err)
		return err
	}
	a.logInfof("prompt delete started: prompt=%s root=%s", name, root)
	if err := store.Delete(name); err != nil {
		a.logErrorf("prompt delete failed: prompt=%s root=%s err=%v", name, root, err)
		return err
	}
	a.logInfof("prompt delete completed: prompt=%s", name)
	a.scheduleAutoBackup()
	return nil
}

func (a *App) CreatePromptCategory(name string) error {
	store, root, err := a.promptStorage()
	if err != nil {
		a.logErrorf("prompt category create failed: category=%s load storage failed: %v", name, err)
		return err
	}
	a.logInfof("prompt category create started: category=%s root=%s", name, root)
	if err := store.CreateCategory(name); err != nil {
		a.logErrorf("prompt category create failed: category=%s root=%s err=%v", name, root, err)
		return err
	}
	a.logInfof("prompt category create completed: category=%s", name)
	a.scheduleAutoBackup()
	return nil
}

func (a *App) RenamePromptCategory(oldName, newName string) error {
	if oldName == prompt.DefaultCategoryName {
		return fmt.Errorf("默认分类不可重命名")
	}
	if newName == prompt.DefaultCategoryName {
		return fmt.Errorf("不能重命名为默认分类")
	}
	store, root, err := a.promptStorage()
	if err != nil {
		a.logErrorf("prompt category rename failed: category=%s next=%s load storage failed: %v", oldName, newName, err)
		return err
	}
	a.logInfof("prompt category rename started: category=%s next=%s root=%s", oldName, newName, root)
	if err := store.RenameCategory(oldName, newName); err != nil {
		a.logErrorf("prompt category rename failed: category=%s next=%s root=%s err=%v", oldName, newName, root, err)
		return err
	}
	a.logInfof("prompt category rename completed: category=%s next=%s", oldName, newName)
	a.scheduleAutoBackup()
	return nil
}

func (a *App) DeletePromptCategory(name string) error {
	if name == prompt.DefaultCategoryName {
		return fmt.Errorf("默认分类不可删除")
	}
	store, root, err := a.promptStorage()
	if err != nil {
		a.logErrorf("prompt category delete failed: category=%s load storage failed: %v", name, err)
		return err
	}
	a.logInfof("prompt category delete started: category=%s root=%s", name, root)
	if err := store.DeleteCategory(name); err != nil {
		a.logErrorf("prompt category delete failed: category=%s root=%s err=%v", name, root, err)
		return err
	}
	a.logInfof("prompt category delete completed: category=%s", name)
	a.scheduleAutoBackup()
	return nil
}

func (a *App) ImportPrompts() (int, error) {
	store, root, err := a.promptStorage()
	if err != nil {
		a.logErrorf("prompt import failed: load storage failed: %v", err)
		return 0, err
	}
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "导入提示词",
		DefaultDirectory: nearestExistingDirectory(root),
		Filters: []runtime.FileFilter{{
			DisplayName: "JSON Files (*.json)",
			Pattern:     "*.json",
		}},
	})
	if err != nil {
		a.logErrorf("prompt import failed: open dialog err=%v", err)
		return 0, err
	}
	if filePath == "" {
		return 0, nil
	}
	a.logInfof("prompt import started: file=%s", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		a.logErrorf("prompt import failed: file=%s err=%v", filePath, err)
		return 0, err
	}
	count, err := store.ImportJSON(data)
	if err != nil {
		a.logErrorf("prompt import failed: file=%s err=%v", filePath, err)
		return 0, err
	}
	a.logInfof("prompt import completed: file=%s count=%d", filePath, count)
	a.scheduleAutoBackup()
	return count, nil
}

func (a *App) ExportPrompts() (string, error) {
	store, root, err := a.promptStorage()
	if err != nil {
		a.logErrorf("prompt export failed: load storage failed: %v", err)
		return "", err
	}
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:                "导出提示词",
		DefaultDirectory:     nearestExistingDirectory(root),
		DefaultFilename:      fmt.Sprintf("skillflow-prompts-%s.json", time.Now().Format("20060102-150405")),
		CanCreateDirectories: true,
		Filters: []runtime.FileFilter{{
			DisplayName: "JSON Files (*.json)",
			Pattern:     "*.json",
		}},
	})
	if err != nil {
		a.logErrorf("prompt export failed: save dialog err=%v", err)
		return "", err
	}
	if filePath == "" {
		return "", nil
	}
	a.logInfof("prompt export started: file=%s", filePath)
	data, err := store.ExportJSON()
	if err != nil {
		a.logErrorf("prompt export failed: file=%s err=%v", filePath, err)
		return "", err
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		a.logErrorf("prompt export failed: file=%s err=%v", filePath, err)
		return "", err
	}
	a.logInfof("prompt export completed: file=%s", filePath)
	return filePath, nil
}

func (a *App) PromptRootDir() (string, error) {
	_, root, err := a.promptStorage()
	if err != nil {
		return "", fmt.Errorf("load prompt root failed: %w", err)
	}
	return root, nil
}
