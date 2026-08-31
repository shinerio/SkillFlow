using System.Text.Json.Serialization;

namespace SkillFlow.Settings;

public sealed class AppSettings
{
    [JsonPropertyName("repoCacheDir")]
    public string RepoCacheDir { get; set; } = string.Empty;

    [JsonPropertyName("autoUpdateSkills")]
    public bool AutoUpdateSkills { get; set; }

    [JsonPropertyName("autoPushAgents")]
    public List<string> AutoPushAgents { get; set; } = new();

    [JsonPropertyName("launchAtLogin")]
    public bool LaunchAtLogin { get; set; }

    [JsonPropertyName("defaultCategory")]
    public string DefaultCategory { get; set; } = string.Empty;

    [JsonPropertyName("logLevel")]
    public string LogLevel { get; set; } = "info";

    [JsonPropertyName("repoScanMaxDepth")]
    public int RepoScanMaxDepth { get; set; } = 5;

    [JsonPropertyName("agents")]
    public List<AgentSettings> Agents { get; set; } = new();

    [JsonPropertyName("cloud")]
    public CloudSettings Cloud { get; set; } = new();

    [JsonPropertyName("cloudProfiles")]
    public Dictionary<string, CloudProviderProfile> CloudProfiles { get; set; } = new();

    [JsonPropertyName("proxy")]
    public ProxySettings Proxy { get; set; } = new();

    [JsonPropertyName("skippedUpdateVersion")]
    public string SkippedUpdateVersion { get; set; } = string.Empty;
}

public sealed class AgentSettings
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("scanDirs")]
    public List<string> ScanDirs { get; set; } = new();

    [JsonPropertyName("pushDir")]
    public string PushDir { get; set; } = string.Empty;

    [JsonPropertyName("memoryPath")]
    public string MemoryPath { get; set; } = string.Empty;

    [JsonPropertyName("rulesDir")]
    public string RulesDir { get; set; } = string.Empty;

    [JsonPropertyName("enabled")]
    public bool Enabled { get; set; }

    [JsonPropertyName("custom")]
    public bool Custom { get; set; }

    [JsonIgnore]
    public string ScanDirsText
    {
        get => string.Join(", ", ScanDirs);
        set => ScanDirs = value
            .Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries)
            .ToList();
    }
}

public sealed class CloudSettings
{
    [JsonPropertyName("provider")]
    public string Provider { get; set; } = string.Empty;

    [JsonPropertyName("enabled")]
    public bool Enabled { get; set; }

    [JsonPropertyName("bucketName")]
    public string BucketName { get; set; } = string.Empty;

    [JsonPropertyName("remotePath")]
    public string RemotePath { get; set; } = "skillflow/";

    [JsonPropertyName("credentials")]
    public Dictionary<string, string> Credentials { get; set; } = new();

    [JsonPropertyName("syncIntervalMinutes")]
    public int SyncIntervalMinutes { get; set; }
}

public sealed class CloudProviderProfile
{
    [JsonPropertyName("bucketName")]
    public string BucketName { get; set; } = string.Empty;

    [JsonPropertyName("remotePath")]
    public string RemotePath { get; set; } = "skillflow/";

    [JsonPropertyName("credentials")]
    public Dictionary<string, string> Credentials { get; set; } = new();
}

public sealed class ProxySettings
{
    [JsonPropertyName("mode")]
    public string Mode { get; set; } = "none";

    [JsonPropertyName("url")]
    public string Url { get; set; } = string.Empty;
}

public sealed class CloudProviderInfo
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("fields")]
    public List<string> Fields { get; set; } = new();
}

public sealed class ProxyTestParameters
{
    [JsonPropertyName("targetURL")]
    public string TargetURL { get; set; } = string.Empty;

    [JsonPropertyName("proxy")]
    public ProxySettings Proxy { get; set; } = new();
}

public sealed class ProxyConnectionTestResult
{
    [JsonPropertyName("targetURL")]
    public string TargetURL { get; set; } = string.Empty;

    [JsonPropertyName("success")]
    public bool Success { get; set; }

    [JsonPropertyName("statusCode")]
    public int StatusCode { get; set; }

    [JsonPropertyName("elapsedMs")]
    public long ElapsedMs { get; set; }

    [JsonPropertyName("message")]
    public string Message { get; set; } = string.Empty;
}

public sealed class AppUpdateInfo
{
    [JsonPropertyName("hasUpdate")]
    public bool HasUpdate { get; set; }

    [JsonPropertyName("currentVersion")]
    public string CurrentVersion { get; set; } = string.Empty;

    [JsonPropertyName("latestVersion")]
    public string LatestVersion { get; set; } = string.Empty;

    [JsonPropertyName("releaseUrl")]
    public string ReleaseURL { get; set; } = string.Empty;

    [JsonPropertyName("downloadUrl")]
    public string DownloadURL { get; set; } = string.Empty;

    [JsonPropertyName("releaseNotes")]
    public string ReleaseNotes { get; set; } = string.Empty;

    [JsonPropertyName("canAutoUpdate")]
    public bool CanAutoUpdate { get; set; }
}

public sealed class CustomAgentAddParams
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("pushDir")]
    public string PushDir { get; set; } = string.Empty;
}

public sealed class AgentNameParams
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;
}
