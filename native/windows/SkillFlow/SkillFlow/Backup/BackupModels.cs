using System.Text.Json.Serialization;

namespace SkillFlow.Backup;

public sealed class RemoteFile
{
    [JsonPropertyName("path")]
    public string Path { get; set; } = string.Empty;

    [JsonPropertyName("size")]
    public long Size { get; set; }

    [JsonPropertyName("isDir")]
    public bool IsDir { get; set; }

    [JsonPropertyName("action")]
    public string? Action { get; set; }
}

public sealed class BackupResolveConflictParams
{
    [JsonPropertyName("useLocal")]
    public bool UseLocal { get; set; }
}
