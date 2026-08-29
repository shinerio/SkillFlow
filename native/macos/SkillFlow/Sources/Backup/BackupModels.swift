import Foundation

struct RemoteFile: Codable, Identifiable, Equatable {
    let path: String
    let size: Int64
    let isDir: Bool
    let action: String?

    var id: String { path }
}

struct BackupResolveConflictParams: Encodable {
    let useLocal: Bool
}
