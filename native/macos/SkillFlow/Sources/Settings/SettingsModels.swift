import Foundation

public struct AppSettings: Codable, Equatable {
    public var repoCacheDir: String
    public var autoUpdateSkills: Bool
    public var autoPushAgents: [String]
    public var launchAtLogin: Bool
    public var defaultCategory: String
    public var logLevel: String
    public var repoScanMaxDepth: Int
    public var agents: [AgentSettings]
    public var cloud: CloudSettings
    public var cloudProfiles: [String: CloudProviderProfile]
    public var proxy: ProxySettings
    public var skippedUpdateVersion: String

    public init(
        repoCacheDir: String = "",
        autoUpdateSkills: Bool = false,
        autoPushAgents: [String] = [],
        launchAtLogin: Bool = false,
        defaultCategory: String = "",
        logLevel: String = "info",
        repoScanMaxDepth: Int = 5,
        agents: [AgentSettings] = [],
        cloud: CloudSettings = CloudSettings(),
        cloudProfiles: [String: CloudProviderProfile] = [:],
        proxy: ProxySettings = ProxySettings(),
        skippedUpdateVersion: String = ""
    ) {
        self.repoCacheDir = repoCacheDir
        self.autoUpdateSkills = autoUpdateSkills
        self.autoPushAgents = autoPushAgents
        self.launchAtLogin = launchAtLogin
        self.defaultCategory = defaultCategory
        self.logLevel = logLevel
        self.repoScanMaxDepth = repoScanMaxDepth
        self.agents = agents
        self.cloud = cloud
        self.cloudProfiles = cloudProfiles
        self.proxy = proxy
        self.skippedUpdateVersion = skippedUpdateVersion
    }

    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        repoCacheDir = try container.decodeIfPresent(String.self, forKey: .repoCacheDir) ?? ""
        autoUpdateSkills = try container.decodeIfPresent(Bool.self, forKey: .autoUpdateSkills) ?? false
        autoPushAgents = try container.decodeIfPresent([String].self, forKey: .autoPushAgents) ?? []
        launchAtLogin = try container.decodeIfPresent(Bool.self, forKey: .launchAtLogin) ?? false
        defaultCategory = try container.decodeIfPresent(String.self, forKey: .defaultCategory) ?? ""
        logLevel = try container.decodeIfPresent(String.self, forKey: .logLevel) ?? "info"
        repoScanMaxDepth = try container.decodeIfPresent(Int.self, forKey: .repoScanMaxDepth) ?? 5
        agents = try container.decodeIfPresent([AgentSettings].self, forKey: .agents) ?? []
        cloud = try container.decodeIfPresent(CloudSettings.self, forKey: .cloud) ?? CloudSettings()
        cloudProfiles = try container.decodeIfPresent([String: CloudProviderProfile].self, forKey: .cloudProfiles) ?? [:]
        proxy = try container.decodeIfPresent(ProxySettings.self, forKey: .proxy) ?? ProxySettings()
        skippedUpdateVersion = try container.decodeIfPresent(String.self, forKey: .skippedUpdateVersion) ?? ""
    }
}

public struct AgentSettings: Codable, Equatable, Identifiable {
    public var name: String
    public var scanDirs: [String]
    public var pushDir: String
    public var memoryPath: String
    public var rulesDir: String
    public var enabled: Bool
    public var custom: Bool

    public var id: String { name }

    public init(
        name: String = "",
        scanDirs: [String] = [],
        pushDir: String = "",
        memoryPath: String = "",
        rulesDir: String = "",
        enabled: Bool = false,
        custom: Bool = false
    ) {
        self.name = name
        self.scanDirs = scanDirs
        self.pushDir = pushDir
        self.memoryPath = memoryPath
        self.rulesDir = rulesDir
        self.enabled = enabled
        self.custom = custom
    }

    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        name = try container.decodeIfPresent(String.self, forKey: .name) ?? ""
        scanDirs = try container.decodeIfPresent([String].self, forKey: .scanDirs) ?? []
        pushDir = try container.decodeIfPresent(String.self, forKey: .pushDir) ?? ""
        memoryPath = try container.decodeIfPresent(String.self, forKey: .memoryPath) ?? ""
        rulesDir = try container.decodeIfPresent(String.self, forKey: .rulesDir) ?? ""
        enabled = try container.decodeIfPresent(Bool.self, forKey: .enabled) ?? false
        custom = try container.decodeIfPresent(Bool.self, forKey: .custom) ?? false
    }
}

public struct CloudSettings: Codable, Equatable {
    public var provider: String
    public var enabled: Bool
    public var bucketName: String
    public var remotePath: String
    public var credentials: [String: String]
    public var syncIntervalMinutes: Int

    public init(
        provider: String = "",
        enabled: Bool = false,
        bucketName: String = "",
        remotePath: String = "skillflow/",
        credentials: [String: String] = [:],
        syncIntervalMinutes: Int = 0
    ) {
        self.provider = provider
        self.enabled = enabled
        self.bucketName = bucketName
        self.remotePath = remotePath
        self.credentials = credentials
        self.syncIntervalMinutes = syncIntervalMinutes
    }

    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        provider = try container.decodeIfPresent(String.self, forKey: .provider) ?? ""
        enabled = try container.decodeIfPresent(Bool.self, forKey: .enabled) ?? false
        bucketName = try container.decodeIfPresent(String.self, forKey: .bucketName) ?? ""
        remotePath = try container.decodeIfPresent(String.self, forKey: .remotePath) ?? "skillflow/"
        credentials = try container.decodeIfPresent([String: String].self, forKey: .credentials) ?? [:]
        syncIntervalMinutes = try container.decodeIfPresent(Int.self, forKey: .syncIntervalMinutes) ?? 0
    }
}

public struct CloudProviderProfile: Codable, Equatable {
    public var bucketName: String
    public var remotePath: String
    public var credentials: [String: String]

    public init(bucketName: String = "", remotePath: String = "skillflow/", credentials: [String: String] = [:]) {
        self.bucketName = bucketName
        self.remotePath = remotePath
        self.credentials = credentials
    }

    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        bucketName = try container.decodeIfPresent(String.self, forKey: .bucketName) ?? ""
        remotePath = try container.decodeIfPresent(String.self, forKey: .remotePath) ?? "skillflow/"
        credentials = try container.decodeIfPresent([String: String].self, forKey: .credentials) ?? [:]
    }
}

public struct ProxySettings: Codable, Equatable {
    public var mode: String
    public var url: String

    public init(mode: String = "none", url: String = "") {
        self.mode = mode
        self.url = url
    }

    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        mode = try container.decodeIfPresent(String.self, forKey: .mode) ?? "none"
        url = try container.decodeIfPresent(String.self, forKey: .url) ?? ""
    }
}

public struct CloudProviderInfo: Codable, Equatable, Identifiable {
    public let name: String
    public let fields: [String]

    public var id: String { name }
}

public struct ProxyTestParameters: Encodable {
    public let targetURL: String
    public let proxy: ProxySettings

    public init(targetURL: String, proxy: ProxySettings) {
        self.targetURL = targetURL
        self.proxy = proxy
    }
}

public struct ProxyConnectionTestResult: Codable, Equatable {
    public let targetURL: String
    public let success: Bool
    public let statusCode: Int
    public let elapsedMs: Int
    public let message: String
}

public struct AppUpdateInfo: Codable, Equatable {
    public let hasUpdate: Bool
    public let currentVersion: String
    public let latestVersion: String
    public let releaseURL: String
    public let downloadURL: String
    public let releaseNotes: String
    public let canAutoUpdate: Bool

    enum CodingKeys: String, CodingKey {
        case hasUpdate
        case currentVersion
        case latestVersion
        case releaseURL = "releaseUrl"
        case downloadURL = "downloadUrl"
        case releaseNotes
        case canAutoUpdate
    }
}

public struct NativeEmptyResult: Codable, Equatable {}
