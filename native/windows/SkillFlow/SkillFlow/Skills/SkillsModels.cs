using System.Text.Json.Serialization;

namespace SkillFlow.Skills;

public sealed class InstalledSkill
{
    [JsonPropertyName("id")]
    public string Id { get; set; } = string.Empty;

    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("path")]
    public string Path { get; set; } = string.Empty;

    [JsonPropertyName("category")]
    public string Category { get; set; } = string.Empty;

    [JsonPropertyName("source")]
    public string Source { get; set; } = string.Empty;

    [JsonPropertyName("sourceSha")]
    public string SourceSha { get; set; } = string.Empty;

    [JsonPropertyName("latestSha")]
    public string LatestSha { get; set; } = string.Empty;

    [JsonPropertyName("updatable")]
    public bool Updatable { get; set; }

    [JsonPropertyName("pushed")]
    public bool Pushed { get; set; }

    [JsonPropertyName("pushedAgents")]
    public List<string> PushedAgents { get; set; } = new();

    public string PushedAgentsText => PushedAgents.Count == 0 ? "—" : string.Join(", ", PushedAgents);
    public string UpdatableGlyph => Updatable ? "\xE72C" : "\xE8A5";
}

public sealed class SkillsImportLocalParams
{
    [JsonPropertyName("dir")]
    public string Dir { get; set; } = string.Empty;

    [JsonPropertyName("category")]
    public string Category { get; set; } = string.Empty;
}

public sealed class SkillsDeleteParams
{
    [JsonPropertyName("skillID")]
    public string SkillID { get; set; } = string.Empty;
}

public sealed class SkillsDeleteBatchParams
{
    [JsonPropertyName("skillIDs")]
    public List<string> SkillIDs { get; set; } = new();
}

public sealed class SkillsMoveCategoryParams
{
    [JsonPropertyName("skillID")]
    public string SkillID { get; set; } = string.Empty;

    [JsonPropertyName("category")]
    public string Category { get; set; } = string.Empty;
}

public sealed class SkillsPushParams
{
    [JsonPropertyName("skillIDs")]
    public List<string> SkillIDs { get; set; } = new();

    [JsonPropertyName("agentNames")]
    public List<string> AgentNames { get; set; } = new();
}

public sealed class AgentInfo
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("scanDirs")]
    public List<string> ScanDirs { get; set; } = new();

    [JsonPropertyName("pushDir")]
    public string PushDir { get; set; } = string.Empty;

    [JsonPropertyName("enabled")]
    public bool Enabled { get; set; }

    [JsonPropertyName("custom")]
    public bool Custom { get; set; }

    public string EnabledGlyph => Enabled ? "\xE73E" : "\xE711";
}
