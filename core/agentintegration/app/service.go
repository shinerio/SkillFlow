package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	gatewayport "github.com/shinerio/skillflow/core/agentintegration/app/port/gateway"
	"github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/shared/logicalkey"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
)

type SkillStatus struct {
	LogicalKey   string
	Source       string
	Installed    bool
	Imported     bool
	Updatable    bool
	ContentKeys  []string
	Pushed       bool
	PushedAgents []string
}

type GatewayResolver func(profile domain.AgentProfile) gatewayport.AgentGateway

type Service struct {
	resolveGateway GatewayResolver
}

func NewService(resolveGateway GatewayResolver) *Service {
	return &Service{resolveGateway: resolveGateway}
}

func (s *Service) EnabledProfiles(profiles []domain.AgentProfile) []domain.AgentProfile {
	result := make([]domain.AgentProfile, 0, len(profiles))
	for _, profile := range profiles {
		if profile.Enabled {
			result = append(result, profile)
		}
	}
	return result
}

func FindProfile(profiles []domain.AgentProfile, agentName string) (domain.AgentProfile, bool) {
	for _, profile := range profiles {
		if profile.Name == agentName {
			return profile, true
		}
	}
	return domain.AgentProfile{}, false
}

func (s *Service) CheckMissingPushDirs(profiles []domain.AgentProfile, agentNames []string) ([]domain.MissingPushDir, error) {
	missing := make([]domain.MissingPushDir, 0, len(agentNames))
	for _, agentName := range agentNames {
		profile, ok := FindProfile(profiles, agentName)
		if !ok || strings.TrimSpace(profile.PushDir) == "" {
			continue
		}
		if _, err := os.Stat(profile.PushDir); os.IsNotExist(err) {
			missing = append(missing, domain.MissingPushDir{Name: profile.Name, Dir: profile.PushDir})
		} else if err != nil {
			return nil, err
		}
	}
	return missing, nil
}

func (s *Service) PushSkills(ctx context.Context, profiles []domain.AgentProfile, agentNames []string, skills []*skilldomain.InstalledSkill, force bool) ([]domain.PushConflict, error) {
	conflicts := make([]domain.PushConflict, 0)
	for _, agentName := range agentNames {
		profile, ok := FindProfile(profiles, agentName)
		if !ok {
			continue
		}
		if strings.TrimSpace(profile.PushDir) == "" {
			return nil, fmt.Errorf("智能体 %s 未配置推送路径", agentName)
		}
		gateway, err := s.gateway(profile)
		if err != nil {
			return nil, err
		}

		toPush := make([]*skilldomain.InstalledSkill, 0, len(skills))
		for _, skill := range skills {
			if skill == nil {
				continue
			}
			targetPath := filepath.Join(profile.PushDir, skill.Name)
			if _, err := os.Stat(targetPath); err == nil {
				if force {
					if err := os.RemoveAll(targetPath); err != nil {
						return nil, err
					}
					toPush = append(toPush, skill)
					continue
				}
				conflicts = append(conflicts, domain.PushConflict{
					SkillID:    skill.ID,
					SkillName:  skill.Name,
					SkillPath:  skill.Path,
					AgentName:  profile.Name,
					TargetPath: targetPath,
				})
				continue
			} else if !os.IsNotExist(err) {
				return nil, err
			}
			toPush = append(toPush, skill)
		}

		if len(toPush) == 0 {
			continue
		}
		if err := gateway.Push(ctx, toPush, profile.PushDir); err != nil {
			return nil, err
		}
	}
	return conflicts, nil
}

func (s *Service) DeletePushedSkill(profile domain.AgentProfile, skillPath string) error {
	if strings.TrimSpace(profile.PushDir) == "" {
		return fmt.Errorf("智能体 %s 未配置推送路径", profile.Name)
	}
	rel, err := filepath.Rel(profile.PushDir, skillPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("无法删除不在推送路径下的 Skill")
	}
	if err := os.RemoveAll(skillPath); err != nil {
		return fmt.Errorf("删除失败: %w", err)
	}
	return nil
}

func (s *Service) BuildPresenceIndex(ctx context.Context, profiles []domain.AgentProfile, idx *skillquery.InstalledIndex, maxDepth int) (*domain.AgentPresenceIndex, error) {
	presence := domain.NewAgentPresenceIndex()
	for _, profile := range profiles {
		if strings.TrimSpace(profile.PushDir) == "" {
			continue
		}
		if _, err := os.Stat(profile.PushDir); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		gateway, err := s.gateway(profile)
		if err != nil {
			return nil, err
		}
		pushed, err := pullAgentSkills(ctx, gateway, profile.PushDir, maxDepth)
		if err != nil {
			return nil, err
		}
		for _, candidate := range pushed {
			if candidate == nil {
				continue
			}
			presence.Add(profile.Name, agentPresenceKeys(candidate, idx)...)
		}
	}
	return presence, nil
}

func (s *Service) ScanAgentSkills(ctx context.Context, profile domain.AgentProfile, idx *skillquery.InstalledIndex, presence *domain.AgentPresenceIndex, maxDepth int) ([]domain.AgentSkillCandidate, error) {
	gateway, err := s.gateway(profile)
	if err != nil {
		return nil, err
	}
	scanned, err := scanAgentSkillsRaw(ctx, gateway, profile.ScanDirs, maxDepth)
	if err != nil {
		return nil, err
	}
	return resolveAgentSkillCandidates(scanned, idx, presence), nil
}

func (s *Service) ListAgentSkills(ctx context.Context, profile domain.AgentProfile, idx *skillquery.InstalledIndex, presence *domain.AgentPresenceIndex, maxDepth int) ([]domain.AgentSkillEntry, error) {
	gateway, err := s.gateway(profile)
	if err != nil {
		return nil, err
	}

	var pushSkills []*skilldomain.InstalledSkill
	if strings.TrimSpace(profile.PushDir) != "" {
		if _, statErr := os.Stat(profile.PushDir); statErr == nil {
			if pulled, pullErr := pullAgentSkills(ctx, gateway, profile.PushDir, maxDepth); pullErr == nil {
				pushSkills = pulled
			}
		}
	}

	scanSkills, err := scanAgentSkillsRaw(ctx, gateway, profile.ScanDirs, maxDepth)
	if err != nil {
		return nil, err
	}
	return aggregateAgentSkillEntries(pushSkills, scanSkills, idx, presence), nil
}

func (s *Service) ListManagedSkills(ctx context.Context, profile domain.AgentProfile, management domain.SkillManagementConfig, maxDepth int) ([]domain.ManagedAgentSkill, error) {
	gateway, err := s.gateway(profile)
	if err != nil {
		return nil, err
	}
	dirs := compactKeys(append([]string{profile.PushDir}, profile.ScanDirs...)...)
	rawSkills, err := scanAgentSkillsRaw(ctx, gateway, dirs, maxDepth)
	if err != nil {
		return nil, err
	}
	return buildManagedSkills(profile.Name, rawSkills, management), nil
}

func (s *Service) ApplyManagedSkillEnablement(ctx context.Context, profile domain.AgentProfile, management domain.SkillManagementConfig, maxDepth int) ([]domain.ManagedAgentSkill, error) {
	gateway, err := s.gateway(profile)
	if err != nil {
		return nil, err
	}
	applier, ok := gateway.(gatewayport.SkillEnablementApplier)
	if !ok {
		return nil, fmt.Errorf("agent skill enablement is not supported: %s", profile.Name)
	}
	managedSkills, err := s.ListManagedSkills(ctx, profile, management, maxDepth)
	if err != nil {
		return nil, err
	}
	if err := applier.ApplySkillEnablement(profile, managedSkills); err != nil {
		return nil, err
	}
	return managedSkills, nil
}

func (s *Service) RefreshPushedCopies(ctx context.Context, profiles []domain.AgentProfile, skill *skilldomain.InstalledSkill) error {
	if skill == nil {
		return nil
	}
	for _, profile := range profiles {
		if strings.TrimSpace(profile.PushDir) == "" {
			continue
		}
		targetPath := filepath.Join(profile.PushDir, skill.Name)
		if _, err := os.Stat(targetPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if err := os.RemoveAll(targetPath); err != nil {
			return err
		}
		gateway, err := s.gateway(profile)
		if err != nil {
			return err
		}
		if err := gateway.Push(ctx, []*skilldomain.InstalledSkill{skill}, profile.PushDir); err != nil {
			return err
		}
	}
	return nil
}

func ResolveSkillStatus(name, logicalKey string, idx *skillquery.InstalledIndex, presence *domain.AgentPresenceIndex, aliasKeys ...string) SkillStatus {
	status := idx.Resolve(name, logicalKey)
	resolvedKey := coalesceLogicalKey(status.LogicalKey, logicalKey)
	lookupKeys := []string{resolvedKey}
	for _, aliasKey := range aliasKeys {
		if strings.TrimSpace(aliasKey) != "" {
			lookupKeys = append(lookupKeys, aliasKey)
		}
	}
	pushedAgents := presence.Agents(lookupKeys...)
	return SkillStatus{
		LogicalKey:   coalesceLogicalKey(resolvedKey, logicalKey),
		Source:       status.Source,
		Installed:    status.Installed,
		Imported:     status.Imported,
		Updatable:    status.Updatable,
		ContentKeys:  append([]string(nil), status.ContentKeys...),
		Pushed:       len(pushedAgents) > 0,
		PushedAgents: pushedAgents,
	}
}

func SelectAgentSkillCandidates(candidates []domain.AgentSkillCandidate, selectedPaths []string) []domain.AgentSkillCandidate {
	selectedSet := make(map[string]struct{}, len(selectedPaths))
	for _, selectedPath := range selectedPaths {
		selectedSet[selectedPath] = struct{}{}
	}

	result := make([]domain.AgentSkillCandidate, 0, len(selectedPaths))
	for _, candidate := range candidates {
		if _, ok := selectedSet[candidate.Path]; ok {
			result = append(result, candidate)
		}
	}
	return result
}

func (s *Service) gateway(profile domain.AgentProfile) (gatewayport.AgentGateway, error) {
	if s.resolveGateway == nil {
		return nil, fmt.Errorf("agent gateway resolver is not configured")
	}
	gateway := s.resolveGateway(profile)
	if gateway == nil {
		return nil, fmt.Errorf("agent gateway not found: %s", profile.Name)
	}
	return gateway, nil
}

func pullAgentSkills(ctx context.Context, gateway gatewayport.AgentGateway, dir string, maxDepth int) ([]*skilldomain.InstalledSkill, error) {
	if depthAware, ok := gateway.(gatewayport.MaxDepthPuller); ok {
		return depthAware.PullWithMaxDepth(ctx, dir, maxDepth)
	}
	return gateway.Pull(ctx, dir)
}

func scanAgentSkillsRaw(ctx context.Context, gateway gatewayport.AgentGateway, scanDirs []string, maxDepth int) ([]*skilldomain.InstalledSkill, error) {
	var result []*skilldomain.InstalledSkill
	seenPaths := map[string]struct{}{}
	for _, dir := range scanDirs {
		if dir == "" {
			continue
		}
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		skills, err := pullAgentSkills(ctx, gateway, dir, maxDepth)
		if err != nil {
			return nil, err
		}
		for _, skill := range skills {
			if skill == nil || strings.TrimSpace(skill.Path) == "" {
				continue
			}
			if _, ok := seenPaths[skill.Path]; ok {
				continue
			}
			seenPaths[skill.Path] = struct{}{}
			result = append(result, skill)
		}
	}
	return result, nil
}

func resolveAgentSkillCandidates(candidates []*skilldomain.InstalledSkill, idx *skillquery.InstalledIndex, presence *domain.AgentPresenceIndex) []domain.AgentSkillCandidate {
	byKey := map[string]domain.AgentSkillCandidate{}
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		logicalKey, err := logicalkey.ContentFromDir(candidate.Path)
		if err != nil {
			logicalKey = ""
		}
		state := ResolveSkillStatus(candidate.Name, logicalKey, idx, presence)
		updatable := state.Updatable
		if state.Installed && logicalKey != "" && !containsContentKey(state.ContentKeys, logicalKey) {
			updatable = true
		}
		resolved := domain.AgentSkillCandidate{
			Name:         candidate.Name,
			Path:         candidate.Path,
			Source:       state.Source,
			LogicalKey:   coalesceLogicalKey(state.LogicalKey, logicalKey),
			Enabled:      true,
			Installed:    state.Installed,
			Imported:     state.Imported,
			Updatable:    updatable,
			Pushed:       state.Pushed,
			PushedAgents: state.PushedAgents,
		}
		groupKey := agentGroupKey(resolved.Name, resolved.LogicalKey, resolved.Path)
		if existing, ok := byKey[groupKey]; ok {
			byKey[groupKey] = mergeAgentCandidate(existing, resolved)
			continue
		}
		byKey[groupKey] = resolved
	}

	result := make([]domain.AgentSkillCandidate, 0, len(byKey))
	for _, candidate := range byKey {
		result = append(result, candidate)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name == result[j].Name {
			return result[i].Path < result[j].Path
		}
		return result[i].Name < result[j].Name
	})
	return result
}

func aggregateAgentSkillEntries(pushSkills, scanSkills []*skilldomain.InstalledSkill, idx *skillquery.InstalledIndex, presence *domain.AgentPresenceIndex) []domain.AgentSkillEntry {
	entries := map[string]domain.AgentSkillEntry{}

	for _, candidate := range resolveAgentSkillCandidates(pushSkills, idx, presence) {
		groupKey := agentGroupKey(candidate.Name, candidate.LogicalKey, candidate.Path)
		entry := entries[groupKey]
		entry.Name = candidate.Name
		entry.Path = candidate.Path
		entry.Source = coalesceSource(candidate.Source, entry.Source)
		entry.LogicalKey = coalesceLogicalKey(candidate.LogicalKey, entry.LogicalKey)
		entry.Enabled = true
		entry.Installed = entry.Installed || candidate.Installed
		entry.Imported = entry.Imported || candidate.Imported
		entry.Updatable = entry.Updatable || candidate.Updatable
		entry.Pushed = true
		entry.PushedAgents = mergeAgentLists(entry.PushedAgents, candidate.PushedAgents)
		entries[groupKey] = entry
	}

	for _, candidate := range resolveAgentSkillCandidates(scanSkills, idx, presence) {
		groupKey := agentGroupKey(candidate.Name, candidate.LogicalKey, candidate.Path)
		entry := entries[groupKey]
		if entry.Path == "" || !entry.Pushed {
			entry.Path = candidate.Path
		}
		if entry.Name == "" {
			entry.Name = candidate.Name
		}
		entry.Source = coalesceSource(candidate.Source, entry.Source)
		entry.LogicalKey = coalesceLogicalKey(candidate.LogicalKey, entry.LogicalKey)
		if !entry.Pushed {
			entry.Enabled = true
		}
		entry.Installed = entry.Installed || candidate.Installed
		entry.Imported = entry.Imported || candidate.Imported
		entry.Updatable = entry.Updatable || candidate.Updatable
		entry.PushedAgents = mergeAgentLists(entry.PushedAgents, candidate.PushedAgents)
		entry.SeenInAgentScan = true
		entries[groupKey] = entry
	}

	result := make([]domain.AgentSkillEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name == result[j].Name {
			return result[i].Path < result[j].Path
		}
		return result[i].Name < result[j].Name
	})
	return result
}

func agentPresenceKeys(candidate *skilldomain.InstalledSkill, idx *skillquery.InstalledIndex) []string {
	if candidate == nil {
		return nil
	}
	contentKey, err := logicalkey.ContentFromDir(candidate.Path)
	if err != nil {
		contentKey = ""
	}
	status := idx.Resolve(candidate.Name, contentKey)
	return compactKeys(status.LogicalKey, contentKey)
}

func mergeAgentCandidate(existing, incoming domain.AgentSkillCandidate) domain.AgentSkillCandidate {
	if existing.Path == "" {
		existing.Path = incoming.Path
	}
	if existing.LogicalKey == "" {
		existing.LogicalKey = incoming.LogicalKey
	}
	if existing.Source == "" {
		existing.Source = incoming.Source
	}
	if existing.GroupName == "" {
		existing.GroupName = incoming.GroupName
	}
	existing.Enabled = existing.Enabled || incoming.Enabled
	existing.Installed = existing.Installed || incoming.Installed
	existing.Imported = existing.Imported || incoming.Imported
	existing.Updatable = existing.Updatable || incoming.Updatable
	existing.Pushed = existing.Pushed || incoming.Pushed
	existing.PushedAgents = mergeAgentLists(existing.PushedAgents, incoming.PushedAgents)
	return existing
}

func mergeAgentLists(primary, secondary []string) []string {
	merged := append([]string{}, primary...)
	for _, agentName := range secondary {
		merged = appendUniqueAgent(merged, agentName)
	}
	sort.Strings(merged)
	return merged
}

func appendUniqueAgent(agents []string, agentName string) []string {
	for _, existing := range agents {
		if existing == agentName {
			return agents
		}
	}
	return append(agents, agentName)
}

func compactKeys(keys ...string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	return result
}

func containsContentKey(keys []string, want string) bool {
	for _, key := range keys {
		if key == want {
			return true
		}
	}
	return false
}

func coalesceLogicalKey(primary, secondary string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return secondary
}

func coalesceSource(primary, secondary string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return secondary
}

func agentGroupKey(name, logicalKey, path string) string {
	if strings.TrimSpace(logicalKey) != "" {
		return "logical:" + logicalKey
	}
	if trimmedName := strings.ToLower(strings.TrimSpace(name)); trimmedName != "" {
		return "fallback:" + trimmedName
	}
	return "path:" + path
}

func buildManagedSkills(agentName string, rawSkills []*skilldomain.InstalledSkill, management domain.SkillManagementConfig) []domain.ManagedAgentSkill {
	grouped := map[string]*domain.ManagedAgentSkill{}
	for _, rawSkill := range rawSkills {
		if rawSkill == nil {
			continue
		}
		skillName := strings.TrimSpace(rawSkill.Name)
		skillPath := strings.TrimSpace(rawSkill.Path)
		if skillName == "" || skillPath == "" {
			continue
		}
		entry := grouped[skillName]
		if entry == nil {
			resolved := domain.ResolveManagedSkillState(agentName, skillName, management)
			entry = &domain.ManagedAgentSkill{
				Name:      skillName,
				GroupName: resolved.GroupName,
				Enabled:   !resolved.Disabled,
			}
			grouped[skillName] = entry
		}
		entry.Paths = appendUniquePath(entry.Paths, skillPath)
		entry.Agents = appendUniqueAgent(entry.Agents, agentName)
	}
	result := make([]domain.ManagedAgentSkill, 0, len(grouped))
	for _, entry := range grouped {
		result = append(result, *entry)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].GroupName == result[j].GroupName {
			return result[i].Name < result[j].Name
		}
		return result[i].GroupName < result[j].GroupName
	})
	return result
}

func appendUniquePath(paths []string, path string) []string {
	for _, existing := range paths {
		if existing == path {
			return paths
		}
	}
	return append(paths, path)
}
