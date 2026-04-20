package agentskills

type AllEntry struct {
	Name          string   `json:"name"`
	GroupName     string   `json:"groupName,omitempty"`
	Agents        []string `json:"agents"`
	InstanceCount int      `json:"instanceCount"`
}
