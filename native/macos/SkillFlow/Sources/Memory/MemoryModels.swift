import Foundation

struct MainMemory: Codable, Equatable {
    let content: String
    let updatedAt: String
}

struct ModuleMemory: Codable, Identifiable, Equatable {
    let name: String
    let content: String
    let enabled: Bool
    let updatedAt: String

    var id: String { name }
}

struct MemoryPushConfig: Codable, Identifiable, Equatable {
    let agentType: String
    let mode: String
    let autoPush: Bool

    var id: String { agentType }
}

struct ModulePushTargets: Codable, Identifiable, Equatable {
    let moduleName: String
    let pushTargets: [String]

    var id: String { moduleName }
}

struct PushStatus: Codable, Identifiable, Equatable {
    let agentType: String
    let status: String

    var id: String { agentType }
}

struct PushResult: Codable, Equatable {
    let agentType: String
    let success: Bool
    let error: String?
}

struct MemoryContentParams: Encodable {
    let content: String
}

struct ModuleMemoryContentParams: Encodable {
    let name: String
    let content: String
}

struct ModuleMemoryNameParams: Encodable {
    let name: String
}

struct ModuleMemoryEnabledParams: Encodable {
    let name: String
    let enabled: Bool
}

struct MemoryAgentTypeParams: Encodable {
    let agentType: String
}

struct MemorySavePushConfigParams: Encodable {
    let agentType: String
    let mode: String
    let autoPush: Bool
}

struct MemoryModulePushTargetsParams: Encodable {
    let moduleName: String
    let pushTargets: [String]
}

struct MemoryPushSelectedParams: Encodable {
    let agentTypes: [String]
    let moduleNames: [String]
    let mode: String
}

struct MemoryOpenEditorParams: Encodable {
    let memoryType: String
    let moduleName: String
}
