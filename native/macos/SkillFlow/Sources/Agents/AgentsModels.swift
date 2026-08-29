import Foundation

struct AgentSkillEntry: Codable, Identifiable, Equatable {
    let name: String
    let path: String
    let source: String
    let logicalKey: String
    let installed: Bool
    let imported: Bool
    let updatable: Bool
    let pushed: Bool
    let pushedAgents: [String]
    let seenInAgentScan: Bool

    var id: String { path }
}

struct AgentSkillCandidate: Codable, Identifiable, Equatable {
    let name: String
    let path: String
    let source: String
    let logicalKey: String
    let installed: Bool
    let imported: Bool
    let updatable: Bool
    let pushed: Bool

    var id: String { path }
}

struct AgentMemoryPreview: Codable, Equatable {
    let agentName: String
    let memoryPath: String
    let rulesDir: String
    let mainExists: Bool
    let mainContent: String
    let rulesDirExists: Bool
    let rules: [AgentMemoryRule]
}

struct AgentMemoryRule: Codable, Identifiable, Equatable {
    let name: String
    let path: String
    let content: String
    let managed: Bool

    var id: String { path }
}

struct AgentPullParams: Encodable {
    let agentName: String
    let skillPaths: [String]
    let category: String
}

struct AgentDeleteSkillParams: Encodable {
    let agentName: String
    let skillPath: String
}
