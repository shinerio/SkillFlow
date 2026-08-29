import SkillFlowCore
import SwiftUI

struct BackupView: View {
    @State private var providers: [CloudProviderInfo] = []
    @State private var remoteFiles: [RemoteFile] = []
    @State private var lastChanges: [RemoteFile] = []
    @State private var lastCompletedAt: String = ""
    @State private var gitConflictPending: Bool = false
    @State private var isLoading: Bool = true
    @State private var isBackingUp: Bool = false
    @State private var isRestoring: Bool = false
    @State private var isListingFiles: Bool = false
    @State private var errorMessage: String?
    @State private var statusMessage: String?

    private let client: DaemonClient

    init(client: DaemonClient = DaemonClient()) {
        self.client = client
    }

    var body: some View {
        VStack(spacing: 0) {
            header
            Divider()
            if isLoading {
                ProgressView("Loading backup status...")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                ScrollView {
                    VStack(alignment: .leading, spacing: 24) {
                        actionsSection
                        if gitConflictPending {
                            conflictSection
                        }
                        statusSection
                        changesSection
                        filesSection
                    }
                    .padding(24)
                    .frame(maxWidth: 880, alignment: .leading)
                    .frame(maxWidth: .infinity, alignment: .center)
                }
            }
            Divider()
            footer
        }
        .frame(minWidth: 720, minHeight: 480)
        .task { await load() }
    }

    private var header: some View {
        HStack(spacing: 12) {
            Image(systemName: "externaldrive.badge.icloud")
                .font(.title2)
                .foregroundStyle(.secondary)
            VStack(alignment: .leading, spacing: 2) {
                Text("Cloud Backup")
                    .font(.title2.weight(.semibold))
                Text("Backup, restore, and manage cloud sync.")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
            Spacer()
        }
        .padding(16)
    }

    private var actionsSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            Label("Actions", systemImage: "arrow.up.arrow.down.circle")
                .font(.title3.weight(.semibold))
            HStack(spacing: 12) {
                Button(action: { Task { await backupNow() } }) {
                    Label("Backup Now", systemImage: "arrow.up.to.line")
                }
                .disabled(isBackingUp)
                if isBackingUp {
                    ProgressView().controlSize(.small)
                }

                Button(action: { Task { await restore() } }) {
                    Label("Restore", systemImage: "arrow.down.to.line")
                }
                .disabled(isRestoring)
                if isRestoring {
                    ProgressView().controlSize(.small)
                }

                Button(action: { Task { await listFiles() } }) {
                    Label("List Remote Files", systemImage: "list.bullet")
                }
                .disabled(isListingFiles)
                if isListingFiles {
                    ProgressView().controlSize(.small)
                }
            }
        }
        .padding(18)
        .background(Color(nsColor: .windowBackgroundColor), in: RoundedRectangle(cornerRadius: 14))
    }

    private var conflictSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            Label("Git Conflict Detected", systemImage: "exclamationmark.triangle")
                .font(.title3.weight(.semibold))
                .foregroundStyle(.orange)
            Text("A merge conflict was detected during the last sync. Choose how to resolve it.")
                .font(.callout)
                .foregroundStyle(.secondary)
            HStack(spacing: 12) {
                Button("Use Local") {
                    Task { await resolveConflict(useLocal: true) }
                }
                Button("Use Remote") {
                    Task { await resolveConflict(useLocal: false) }
                }
            }
        }
        .padding(18)
        .background(Color.orange.opacity(0.1), in: RoundedRectangle(cornerRadius: 14))
    }

    private var statusSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Label("Status", systemImage: "checkmark.circle")
                .font(.title3.weight(.semibold))
            if lastCompletedAt.isEmpty {
                Text("No backup has been completed yet.")
                    .foregroundStyle(.secondary)
            } else {
                Text("Last backup: \(lastCompletedAt)")
                    .font(.callout)
            }
        }
        .padding(18)
        .background(Color(nsColor: .windowBackgroundColor), in: RoundedRectangle(cornerRadius: 14))
    }

    private var changesSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Label("Last Backup Changes", systemImage: "arrow.triangle.2.circlepath")
                .font(.title3.weight(.semibold))
            if lastChanges.isEmpty {
                Text("No changes in the last backup.")
                    .foregroundStyle(.secondary)
            } else {
                ForEach(lastChanges) { change in
                    HStack {
                        Image(systemName: change.isDir ? "folder" : "doc")
                            .foregroundStyle(.secondary)
                        Text(change.path)
                            .font(.callout)
                        Spacer()
                        if let action = change.action, !action.isEmpty {
                            Text(action)
                                .font(.caption)
                                .padding(.horizontal, 6)
                                .padding(.vertical, 2)
                                .background(actionColor(action).opacity(0.2), in: Capsule())
                        }
                    }
                }
            }
        }
        .padding(18)
        .background(Color(nsColor: .windowBackgroundColor), in: RoundedRectangle(cornerRadius: 14))
    }

    private var filesSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Label("Remote Files", systemImage: "icloud")
                .font(.title3.weight(.semibold))
            if remoteFiles.isEmpty {
                Text("Click 'List Remote Files' to load.")
                    .foregroundStyle(.secondary)
            } else {
                ForEach(remoteFiles) { file in
                    HStack {
                        Image(systemName: file.isDir ? "folder" : "doc")
                            .foregroundStyle(.secondary)
                        Text(file.path)
                            .font(.callout)
                        Spacer()
                        Text(formatSize(file.size))
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }
        }
        .padding(18)
        .background(Color(nsColor: .windowBackgroundColor), in: RoundedRectangle(cornerRadius: 14))
    }

    private var footer: some View {
        HStack {
            if let statusMessage {
                Text(statusMessage)
                    .foregroundStyle(.secondary)
                    .font(.callout)
            }
            if let errorMessage {
                Text(errorMessage)
                    .foregroundStyle(.red)
                    .font(.callout)
            }
            Spacer()
            Button("Reload") {
                Task { await load() }
            }
        }
        .padding(12)
    }

    private func actionColor(_ action: String) -> Color {
        switch action {
        case "added": .green
        case "modified": .blue
        case "deleted": .red
        default: .gray
        }
    }

    private func formatSize(_ bytes: Int64) -> String {
        if bytes < 1024 { return "\(bytes) B" }
        if bytes < 1024 * 1024 { return "\(bytes / 1024) KB" }
        return "\(bytes / (1024 * 1024)) MB"
    }

    private func load() async {
        isLoading = true
        errorMessage = nil
        do {
            let providerList: [CloudProviderInfo] = try await client.invoke("backup.providers.list")
            providers = providerList
            gitConflictPending = try await client.invoke("backup.gitConflictPending")
            let changes: [RemoteFile] = try await client.invoke("backup.lastChanges")
            lastChanges = changes
            let completedAt: String = try await client.invoke("backup.lastCompletedAt")
            lastCompletedAt = completedAt
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }

    private func backupNow() async {
        isBackingUp = true
        errorMessage = nil
        statusMessage = nil
        do {
            let _: NativeEmptyResult = try await client.invoke("backup.now")
            statusMessage = "Backup completed."
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
        isBackingUp = false
    }

    private func restore() async {
        let alert = NSAlert()
        alert.messageText = "Restore from Cloud?"
        alert.informativeText = "This will overwrite local data with the cloud backup."
        alert.alertStyle = .warning
        alert.addButton(withTitle: "Restore")
        alert.addButton(withTitle: "Cancel")
        guard alert.runModal() == .alertFirstButtonReturn else { return }

        isRestoring = true
        errorMessage = nil
        statusMessage = nil
        do {
            let _: NativeEmptyResult = try await client.invoke("backup.restore")
            statusMessage = "Restore completed."
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
        isRestoring = false
    }

    private func listFiles() async {
        isListingFiles = true
        errorMessage = nil
        do {
            let files: [RemoteFile] = try await client.invoke("backup.listFiles")
            remoteFiles = files
        } catch {
            errorMessage = error.localizedDescription
        }
        isListingFiles = false
    }

    private func resolveConflict(useLocal: Bool) async {
        errorMessage = nil
        statusMessage = nil
        let params = BackupResolveConflictParams(useLocal: useLocal)
        do {
            let _: NativeEmptyResult = try await client.invoke("backup.resolveGitConflict", parameters: params)
            statusMessage = useLocal ? "Conflict resolved: kept local changes." : "Conflict resolved: used remote state."
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
