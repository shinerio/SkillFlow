import Foundation

struct InstalledSkill: Codable, Identifiable, Equatable {
    let id: String
    let name: String
    let path: String
    let category: String
    let source: String
    let sourceSha: String
    let latestSha: String
    let updatable: Bool
    let pushed: Bool
    let pushedAgents: [String]
}

struct SkillsImportLocalParams: Encodable {
    let dir: String
    let category: String
}

struct SkillsDeleteParams: Encodable {
    let skillID: String
}

struct SkillsDeleteBatchParams: Encodable {
    let skillIDs: [String]
}

struct SkillsMoveCategoryParams: Encodable {
    let skillID: String
    let category: String
}

struct SkillsPushParams: Encodable {
    let skillIDs: [String]
    let agentNames: [String]
}

struct AgentConfig: Codable, Identifiable, Equatable, Hashable {
    let name: String
    let scanDirs: [String]
    let pushDir: String
    let enabled: Bool
    let custom: Bool
    let memoryPath: String
    let rulesDir: String

    var id: String { name }
}
