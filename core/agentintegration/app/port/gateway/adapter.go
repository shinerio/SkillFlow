package gateway

import (
	"context"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
)

type AgentGateway interface {
	Name() string
	DefaultSkillsDir() string
	Push(ctx context.Context, skills []*skilldomain.InstalledSkill, targetDir string) error
	Pull(ctx context.Context, sourceDir string) ([]*skilldomain.InstalledSkill, error)
}

type MaxDepthPuller interface {
	PullWithMaxDepth(ctx context.Context, sourceDir string, maxDepth int) ([]*skilldomain.InstalledSkill, error)
}

type SkillEnablementApplier interface {
	ApplySkillEnablement(profile agentdomain.AgentProfile, skills []agentdomain.ManagedAgentSkill) error
}
