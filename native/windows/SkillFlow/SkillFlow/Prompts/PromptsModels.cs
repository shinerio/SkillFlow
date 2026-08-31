using System.Text.Json.Serialization;

namespace SkillFlow.Prompts;

public sealed class PromptEntry
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("description")]
    public string Description { get; set; } = string.Empty;

    [JsonPropertyName("category")]
    public string Category { get; set; } = string.Empty;

    [JsonPropertyName("path")]
    public string Path { get; set; } = string.Empty;

    [JsonPropertyName("filePath")]
    public string FilePath { get; set; } = string.Empty;

    [JsonPropertyName("content")]
    public string Content { get; set; } = string.Empty;

    [JsonPropertyName("imageURLs")]
    public List<string> ImageURLs { get; set; } = new();

    [JsonPropertyName("webLinks")]
    public List<PromptLink> WebLinks { get; set; } = new();

    [JsonPropertyName("createdAt")]
    public DateTimeOffset CreatedAt { get; set; }

    [JsonPropertyName("updatedAt")]
    public DateTimeOffset UpdatedAt { get; set; }

    public string WebLinksText => WebLinks.Count == 0 ? "" : string.Join(", ", WebLinks.Select(l => $"[{l.Label}]({l.URL})"));
}

public sealed class PromptLink
{
    [JsonPropertyName("label")]
    public string Label { get; set; } = string.Empty;

    [JsonPropertyName("url")]
    public string URL { get; set; } = string.Empty;
}

public sealed class PromptNameParams
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;
}

public sealed class PromptMoveCategoryParams
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("category")]
    public string Category { get; set; } = string.Empty;
}

public sealed class PromptCategoryNameParams
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;
}

public sealed class PromptRenameCategoryParams
{
    [JsonPropertyName("oldName")]
    public string OldName { get; set; } = string.Empty;

    [JsonPropertyName("newName")]
    public string NewName { get; set; } = string.Empty;
}

public sealed class PromptCreateParams
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("description")]
    public string Description { get; set; } = string.Empty;

    [JsonPropertyName("category")]
    public string Category { get; set; } = string.Empty;

    [JsonPropertyName("content")]
    public string Content { get; set; } = string.Empty;

    [JsonPropertyName("imageURLs")]
    public List<string> ImageURLs { get; set; } = new();

    [JsonPropertyName("webLinksMarkdown")]
    public string WebLinksMarkdown { get; set; } = string.Empty;
}

public sealed class PromptUpdateParams
{
    [JsonPropertyName("originalName")]
    public string OriginalName { get; set; } = string.Empty;

    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("description")]
    public string Description { get; set; } = string.Empty;

    [JsonPropertyName("category")]
    public string Category { get; set; } = string.Empty;

    [JsonPropertyName("content")]
    public string Content { get; set; } = string.Empty;

    [JsonPropertyName("imageURLs")]
    public List<string> ImageURLs { get; set; } = new();

    [JsonPropertyName("webLinksMarkdown")]
    public string WebLinksMarkdown { get; set; } = string.Empty;
}

public sealed class PromptExportByNamesParams
{
    [JsonPropertyName("names")]
    public List<string> Names { get; set; } = new();
}
