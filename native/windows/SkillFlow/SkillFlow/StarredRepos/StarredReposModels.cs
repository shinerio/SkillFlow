using System.Text.Json.Serialization;

namespace SkillFlow.StarredRepos;

public sealed class StarRepo
{
    [JsonPropertyName("url")]
    public string Url { get; set; } = string.Empty;

    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("source")]
    public string Source { get; set; } = string.Empty;

    [JsonPropertyName("localDir")]
    public string LocalDir { get; set; } = string.Empty;

    [JsonPropertyName("lastSync")]
    public DateTimeOffset LastSync { get; set; }

    [JsonPropertyName("syncError")]
    public string SyncError { get; set; } = string.Empty;

    public string SyncStatusText => string.IsNullOrEmpty(SyncError)
        ? LastSync.ToString("yyyy-MM-dd HH:mm")
        : SyncError;
}

public sealed class StarSkillEntry
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("path")]
    public string Path { get; set; } = string.Empty;

    [JsonPropertyName("subPath")]
    public string SubPath { get; set; } = string.Empty;

    [JsonPropertyName("repoUrl")]
    public string RepoUrl { get; set; } = string.Empty;

    [JsonPropertyName("repoName")]
    public string RepoName { get; set; } = string.Empty;

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

    public string StatusText
    {
        get
        {
            if (Installed) return "Installed";
            if (Imported) return "Imported";
            return "New";
        }
    }
}

public sealed class StarredRepoURLParams
{
    [JsonPropertyName("repoURL")]
    public string RepoURL { get; set; } = string.Empty;
}

public sealed class StarredRepoCredentialsParams
{
    [JsonPropertyName("repoURL")]
    public string RepoURL { get; set; } = string.Empty;

    [JsonPropertyName("username")]
    public string Username { get; set; } = string.Empty;

    [JsonPropertyName("password")]
    public string Password { get; set; } = string.Empty;
}

public sealed class StarredImportParams
{
    [JsonPropertyName("skillPaths")]
    public List<string> SkillPaths { get; set; } = new();

    [JsonPropertyName("repoURL")]
    public string RepoURL { get; set; } = string.Empty;

    [JsonPropertyName("category")]
    public string Category { get; set; } = string.Empty;
}

public sealed class StarredPushParams
{
    [JsonPropertyName("skillPaths")]
    public List<string> SkillPaths { get; set; } = new();

    [JsonPropertyName("agentNames")]
    public List<string> AgentNames { get; set; } = new();
}
