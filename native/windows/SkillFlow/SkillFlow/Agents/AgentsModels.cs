using System.Text.Json.Serialization;

namespace SkillFlow.Agents;

public sealed class AgentSkillEntry
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("path")]
    public string Path { get; set; } = string.Empty;

    [JsonPropertyName("source")]
    public string Source { get; set; } = string.Empty;

    [JsonPropertyName("logicalKey")]
    public string LogicalKey { get; set; } = string.Empty;

    [JsonPropertyName("installed")]
    public bool Installed { get; set; }

    [JsonPropertyName("imported")]
    public bool Imported { get; set; }

    [JsonPropertyName("updatable")]
    public bool Updatable { get; set; }

    [JsonPropertyName("pushed")]
    public bool Pushed { get; set; }

    [JsonPropertyName("pushedAgents")]
    public List<string> PushedAgents { get; set; } = new();

    [JsonPropertyName("seenInAgentScan")]
    public bool SeenInAgentScan { get; set; }
}

public sealed class AgentSkillCandidate
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("path")]
    public string Path { get; set; } = string.Empty;

    [JsonPropertyName("source")]
    public string Source { get; set; } = string.Empty;

    [JsonPropertyName("logicalKey")]
    public string LogicalKey { get; set; } = string.Empty;

    [JsonPropertyName("installed")]
    public bool Installed { get; set; }

    [JsonPropertyName("imported")]
    public bool Imported { get; set; }

    [JsonPropertyName("updatable")]
    public bool Updatable { get; set; }

    [JsonPropertyName("pushed")]
    public bool Pushed { get; set; }
}

public sealed class AgentMemoryPreview
{
    [JsonPropertyName("agentName")]
    public string AgentName { get; set; } = string.Empty;

    [JsonPropertyName("memoryPath")]
    public string MemoryPath { get; set; } = string.Empty;

    [JsonPropertyName("rulesDir")]
    public string RulesDir { get; set; } = string.Empty;

    [JsonPropertyName("mainExists")]
    public bool MainExists { get; set; }

    [JsonPropertyName("mainContent")]
    public string MainContent { get; set; } = string.Empty;

    [JsonPropertyName("rulesDirExists")]
    public bool RulesDirExists { get; set; }

    [JsonPropertyName("rules")]
    public List<AgentMemoryRule> Rules { get; set; } = new();
}

public sealed class AgentMemoryRule
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("path")]
    public string Path { get; set; } = string.Empty;

    [JsonPropertyName("content")]
    public string Content { get; set; } = string.Empty;

    [JsonPropertyName("managed")]
    public bool Managed { get; set; }
}

public sealed class AgentNameParams
{
    [JsonPropertyName("agentName")]
    public string AgentName { get; set; } = string.Empty;
}

public sealed class AgentPullParams
{
    [JsonPropertyName("agentName")]
    public string AgentName { get; set; } = string.Empty;

    [JsonPropertyName("skillPaths")]
    public List<string> SkillPaths { get; set; } = new();

    [JsonPropertyName("category")]
    public string Category { get; set; } = string.Empty;
}

public sealed class AgentDeleteSkillParams
{
    [JsonPropertyName("agentName")]
    public string AgentName { get; set; } = string.Empty;

    [JsonPropertyName("skillPath")]
    public string SkillPath { get; set; } = string.Empty;
}
