import SkillFlow
import XCTest

final class SettingsModelsTests: XCTestCase {
    func testSettingsModelsDecodeCompletePayload() throws {
        let payload = """
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
        """

        let settings = try JSONDecoder().decode(AppSettings.self, from: Data(payload.utf8))
        XCTAssertEqual(settings.repoCacheDir, "/tmp/cache")
        XCTAssertEqual(settings.agents.first?.pushDir, "/tmp/push")
        XCTAssertEqual(settings.cloud.credentials["repo_url"], "https://example.com/repo.git")
        XCTAssertEqual(settings.cloudProfiles["git"]?.credentials["branch"], "main")
        XCTAssertEqual(settings.proxy.url, "http://127.0.0.1:7890")

        let encoded = try JSONEncoder().encode(settings)
        let decoded = try JSONDecoder().decode(AppSettings.self, from: encoded)
        XCTAssertEqual(settings, decoded)
    }

    func testSettingsModelsDecodeNullCollections() throws {
        let payload = """
        {"repoCacheDir":"","autoUpdateSkills":false,"autoPushAgents":null,"launchAtLogin":false,
        "defaultCategory":"","logLevel":"info","repoScanMaxDepth":5,"agents":null,
        "cloud":{"provider":"","enabled":false,"bucketName":"","remotePath":"skillflow/",
        "credentials":null,"syncIntervalMinutes":0},"cloudProfiles":null,
        "proxy":{"mode":"none","url":""},"skippedUpdateVersion":""}
        """

        let settings = try JSONDecoder().decode(AppSettings.self, from: Data(payload.utf8))
        XCTAssertEqual(settings.autoPushAgents, [])
        XCTAssertEqual(settings.agents, [])
        XCTAssertEqual(settings.cloud.credentials, [:])
        XCTAssertEqual(settings.cloudProfiles, [:])
    }
}
