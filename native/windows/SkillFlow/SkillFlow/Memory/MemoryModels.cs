using System.Text.Json.Serialization;

namespace SkillFlow.Memory;

public sealed class MainMemory
{
    [JsonPropertyName("content")]
    public string Content { get; set; } = string.Empty;

    [JsonPropertyName("updatedAt")]
    public string UpdatedAt { get; set; } = string.Empty;
}

public sealed class ModuleMemory
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("content")]
    public string Content { get; set; } = string.Empty;

    [JsonPropertyName("enabled")]
    public bool Enabled { get; set; }

    [JsonPropertyName("updatedAt")]
    public string UpdatedAt { get; set; } = string.Empty;
}

public sealed class MemoryPushConfig
{
    [JsonPropertyName("agentType")]
    public string AgentType { get; set; } = string.Empty;

    [JsonPropertyName("mode")]
    public string Mode { get; set; } = string.Empty;

    [JsonPropertyName("autoPush")]
    public bool AutoPush { get; set; }
}

public sealed class PushStatus
{
    [JsonPropertyName("agentType")]
    public string AgentType { get; set; } = string.Empty;

    [JsonPropertyName("status")]
    public string Status { get; set; } = string.Empty;
}

public sealed class PushResult
{
    [JsonPropertyName("agentType")]
    public string AgentType { get; set; } = string.Empty;

    [JsonPropertyName("success")]
    public bool Success { get; set; }

    [JsonPropertyName("error")]
    public string? Error { get; set; }
}

public sealed class MemoryContentParams
{
    [JsonPropertyName("content")]
    public string Content { get; set; } = string.Empty;
}

public sealed class ModuleMemoryContentParams
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("content")]
    public string Content { get; set; } = string.Empty;
}

public sealed class ModuleMemoryNameParams
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;
}

public sealed class ModuleMemoryEnabledParams
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("enabled")]
    public bool Enabled { get; set; }
}

public sealed class MemoryOpenEditorParams
{
    [JsonPropertyName("memoryType")]
    public string MemoryType { get; set; } = string.Empty;

    [JsonPropertyName("moduleName")]
    public string ModuleName { get; set; } = string.Empty;
}

public sealed class MemorySavePushConfigParams
{
    [JsonPropertyName("agentType")]
    public string AgentType { get; set; } = string.Empty;

    [JsonPropertyName("mode")]
    public string Mode { get; set; } = string.Empty;

    [JsonPropertyName("autoPush")]
    public bool AutoPush { get; set; }
}

public sealed class MemoryPushSelectedParams
{
    [JsonPropertyName("agentTypes")]
    public List<string> AgentTypes { get; set; } = new();

    [JsonPropertyName("moduleNames")]
    public List<string> ModuleNames { get; set; } = new();

    [JsonPropertyName("mode")]
    public string Mode { get; set; } = string.Empty;
}
