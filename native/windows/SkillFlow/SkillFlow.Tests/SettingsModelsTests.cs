using System.Text.Json;
using SkillFlow.Settings;
using Xunit;

namespace SkillFlow.Tests;

public sealed class SettingsModelsTests
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
        PropertyNameCaseInsensitive = true,
    };

    [Fact]
    public void SettingsModelsDecodeCompletePayload()
    {
        var payload = """
        {
          "repoCacheDir": "/tmp/cache",
          "autoUpdateSkills": true,
          "autoPushAgents": ["claude"],
          "launchAtLogin": true,
          "defaultCategory": "general",
          "logLevel": "debug",
          "repoScanMaxDepth": 8,
          "agents": [{
            "name": "claude",
            "scanDirs": ["/tmp/scan"],
            "pushDir": "/tmp/push",
            "memoryPath": "/tmp/memory.md",
            "rulesDir": "/tmp/rules",
            "enabled": true,
            "custom": false
          }],
          "cloud": {
            "provider": "git",
            "enabled": true,
            "bucketName": "repo",
            "remotePath": "skillflow/",
            "credentials": {"repo_url": "https://example.com/repo.git"},
            "syncIntervalMinutes": 15
          },
          "cloudProfiles": {
            "git": {"bucketName": "repo", "remotePath": "skillflow/", "credentials": {"branch": "main"}}
          },
          "proxy": {"mode": "manual", "url": "http://127.0.0.1:7890"},
          "skippedUpdateVersion": "v1.2.3"
        }
        """;

        var settings = JsonSerializer.Deserialize<AppSettings>(payload, JsonOptions)!;

        Assert.Equal("/tmp/cache", settings.RepoCacheDir);
        Assert.True(settings.AutoUpdateSkills);
        Assert.Equal(["claude"], settings.AutoPushAgents);
        Assert.True(settings.LaunchAtLogin);
        Assert.Equal("debug", settings.LogLevel);
        Assert.Equal(8, settings.RepoScanMaxDepth);

        var agent = Assert.Single(settings.Agents);
        Assert.Equal("claude", agent.Name);
        Assert.Equal("/tmp/push", agent.PushDir);
        Assert.Equal(["/tmp/scan"], agent.ScanDirs);
        Assert.True(agent.Enabled);

        Assert.Equal("git", settings.Cloud.Provider);
        Assert.True(settings.Cloud.Enabled);
        Assert.Equal("https://example.com/repo.git", settings.Cloud.Credentials["repo_url"]);
        Assert.Equal(15, settings.Cloud.SyncIntervalMinutes);

        Assert.True(settings.CloudProfiles.ContainsKey("git"));
        Assert.Equal("main", settings.CloudProfiles["git"].Credentials["branch"]);

        Assert.Equal("manual", settings.Proxy.Mode);
        Assert.Equal("http://127.0.0.1:7890", settings.Proxy.Url);

        var encoded = JsonSerializer.Serialize(settings, JsonOptions);
        var decoded = JsonSerializer.Deserialize<AppSettings>(encoded, JsonOptions)!;
        Assert.Equal(settings.RepoCacheDir, decoded.RepoCacheDir);
        Assert.Equal(settings.Cloud.Credentials["repo_url"], decoded.Cloud.Credentials["repo_url"]);
        Assert.Equal(settings.Proxy.Url, decoded.Proxy.Url);
    }

    [Fact]
    public void SettingsModelsDecodeNullCollections()
    {
        var payload = """
        {"repoCacheDir":"","autoUpdateSkills":false,"autoPushAgents":null,"launchAtLogin":false,
        "defaultCategory":"","logLevel":"info","repoScanMaxDepth":5,"agents":null,
        "cloud":{"provider":"","enabled":false,"bucketName":"","remotePath":"skillflow/",
        "credentials":null,"syncIntervalMinutes":0},"cloudProfiles":null,
        "proxy":{"mode":"none","url":""},"skippedUpdateVersion":""}
        """;

        var settings = JsonSerializer.Deserialize<AppSettings>(payload, JsonOptions)!;

        Assert.NotNull(settings.AutoPushAgents);
        Assert.Empty(settings.AutoPushAgents);
        Assert.NotNull(settings.Agents);
        Assert.Empty(settings.Agents);
        Assert.NotNull(settings.Cloud.Credentials);
        Assert.Empty(settings.Cloud.Credentials);
        Assert.NotNull(settings.CloudProfiles);
        Assert.Empty(settings.CloudProfiles);
    }

    [Fact]
    public void AgentSettingsScanDirsTextRoundTrip()
    {
        var agent = new AgentSettings
        {
            ScanDirs = ["/a", "/b", "/c"],
        };

        Assert.Equal("/a, /b, /c", agent.ScanDirsText);

        agent.ScanDirsText = "/x, /y";
        Assert.Equal(["/x", "/y"], agent.ScanDirs);
    }

    [Fact]
    public void ProxyTestParametersSerializesTargetURL()
    {
        var parameters = new ProxyTestParameters
        {
            TargetURL = "https://github.com",
            Proxy = new ProxySettings { Mode = "manual", Url = "http://proxy:8080" },
        };

        var json = JsonSerializer.Serialize(parameters, JsonOptions);
        using var document = JsonDocument.Parse(json);

        Assert.True(document.RootElement.TryGetProperty("targetURL", out var targetURL));
        Assert.Equal("https://github.com", targetURL.GetString());
    }

    [Fact]
    public void AppUpdateInfoDecodesReleaseAndDownloadUrls()
    {
        var payload = """
        {
          "hasUpdate": true,
          "currentVersion": "v1.0.0",
          "latestVersion": "v1.1.0",
          "releaseUrl": "https://github.com/shinerio/skillflow/releases/v1.1.0",
          "downloadUrl": "https://github.com/shinerio/skillflow/releases/download/v1.1.0/skillflow.exe",
          "releaseNotes": "Bug fixes and improvements",
          "canAutoUpdate": true
        }
        """;

        var info = JsonSerializer.Deserialize<AppUpdateInfo>(payload, JsonOptions)!;

        Assert.True(info.HasUpdate);
        Assert.Equal("v1.0.0", info.CurrentVersion);
        Assert.Equal("v1.1.0", info.LatestVersion);
        Assert.Equal("https://github.com/shinerio/skillflow/releases/v1.1.0", info.ReleaseURL);
        Assert.Equal("https://github.com/shinerio/skillflow/releases/download/v1.1.0/skillflow.exe", info.DownloadURL);
        Assert.Equal("Bug fixes and improvements", info.ReleaseNotes);
        Assert.True(info.CanAutoUpdate);
    }
}
