import Foundation

struct StarRepo: Codable, Identifiable, Equatable {
    let url: String
    let name: String
    let source: String
    let localDir: String
    let lastSync: Date
    let syncError: String?

    var id: String { url }

    var syncStatusText: String {
        if let error = syncError, !error.isEmpty { return error }
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd HH:mm"
        return formatter.string(from: lastSync)
    }
}

struct StarSkillEntry: Codable, Identifiable, Equatable {
    let name: String
    let path: String
    let subPath: String
    let repoUrl: String
    let repoName: String
    let source: String
    let logicalKey: String
    let installed: Bool
    let imported: Bool
    let updatable: Bool
    let pushed: Bool
    let pushedAgents: [String]

    var id: String { path }

    var statusText: String {
        if installed { return "Installed" }
        if imported { return "Imported" }
        return "New"
    }
}

struct StarredRepoURLParams: Encodable {
    let repoURL: String
}

struct StarredRepoCredentialsParams: Encodable {
    let repoURL: String
    let username: String
    let password: String
}

struct StarredImportParams: Encodable {
    let skillPaths: [String]
    let repoURL: String
    let category: String
}

struct StarredPushParams: Encodable {
    let skillPaths: [String]
    let agentNames: [String]
}
