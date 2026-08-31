import AppKit
import SwiftUI

struct SettingsView: View {
    @State
    private var draft = AppSettings()
    @State
    private var providers: [CloudProviderInfo] = []
    @State
    private var isLoading = true
    @State
    private var isSaving = false
    @State
    private var isTestingProxy = false
    @State
    private var isCheckingUpdate = false
    @State
    private var errorMessage: String?
    @State
    private var statusMessage: String?
    @State
    private var updateInfo: AppUpdateInfo?
    @State
    private var showAddCustomAgent: Bool = false

    private let client: DaemonClient

    init(client: DaemonClient = DaemonClient()) {
        self.client = client
    }

    var body: some View {
        VStack(spacing: 0) {
            header
            Divider()
            if isLoading {
                ProgressView("Loading settings...")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                ScrollView {
                    VStack(alignment: .leading, spacing: 24) {
                        generalSection
                        agentsSection
                        cloudSection
                        networkSection
                        maintenanceSection
                    }
                    .padding(24)
                    .frame(maxWidth: 880, alignment: .leading)
                    .frame(maxWidth: .infinity, alignment: .center)
                }
            }
            Divider()
            footer
        }
        .frame(minWidth: 720, minHeight: 560)
        .task {
            await load()
        }
        .sheet(isPresented: $showAddCustomAgent) {
            AddCustomAgentSheet(client: client) { _ in
                Task { await load() }
            }
        }
    }

    private var header: some View {
        HStack(spacing: 12) {
            Image(systemName: "gearshape")
                .font(.title2)
                .foregroundStyle(.secondary)
            VStack(alignment: .leading, spacing: 2) {
                Text("Settings")
                    .font(.title2.weight(.semibold))
                Text("Manage SkillFlow preferences and system integrations.")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
            Spacer()
            Button("Reload") {
                Task { await load() }
            }
            .disabled(isLoading || isSaving)
            Button("Save") {
                Task { await save() }
            }
            .keyboardShortcut("s", modifiers: [.command])
            .disabled(isLoading || isSaving)
        }
        .padding(20)
    }

    private var generalSection: some View {
        SettingsSection(
            title: "General",
            systemImage: "slider.horizontal.3",
            content: {
                Picker("Log level", selection: $draft.logLevel) {
                    Text("Debug").tag("debug")
                    Text("Info").tag("info")
                    Text("Error").tag("error")
                }
                .pickerStyle(.segmented)
                Toggle("Launch at login", isOn: $draft.launchAtLogin)
                HStack {
                    TextField("Repository cache directory", text: $draft.repoCacheDir)
                    Button("Choose...") {
                        chooseDirectory { path in
                            if let path {
                                draft.repoCacheDir = path
                            }
                        }
                    }
                }
                Stepper(
                    "Repository scan depth: \(draft.repoScanMaxDepth)",
                    value: $draft.repoScanMaxDepth,
                    in: 1...20
                )
                Toggle("Automatically update skills", isOn: $draft.autoUpdateSkills)
            }
        )
    }

    private var agentsSection: some View {
        SettingsSection(
            title: "Agents",
            systemImage: "person.2",
            content: {
                if draft.agents.isEmpty {
                    Text("No agents are configured in the daemon.")
                        .foregroundStyle(.secondary)
                } else {
                    ForEach(draft.agents) { agent in
                        if let index = draft.agents.firstIndex(where: { $0.id == agent.id }) {
                            AgentSettingsRow(agent: $draft.agents[index]) {
                                Task { await removeCustomAgent(name: agent.name) }
                            }
                        }
                    }
                }
                Button(action: { showAddCustomAgent = true }) {
                    Label("Add Custom Agent", systemImage: "plus")
                }
                .buttonStyle(.bordered)
            }
        )
    }

    private var cloudSection: some View {
        SettingsSection(
            title: "Cloud Backup",
            systemImage: "externaldrive.badge.icloud",
            content: {
                Picker("Provider", selection: $draft.cloud.provider) {
                    ForEach(providers) { provider in
                        Text(provider.name.capitalized).tag(provider.name)
                    }
                }
                Toggle("Enable scheduled cloud backup", isOn: $draft.cloud.enabled)
                TextField("Bucket or repository", text: $draft.cloud.bucketName)
                TextField("Remote path", text: $draft.cloud.remotePath)
                Stepper(
                    "Sync interval: \(draft.cloud.syncIntervalMinutes) minutes",
                    value: $draft.cloud.syncIntervalMinutes,
                    in: 0...1440,
                    step: 5
                )
                if let provider = providers.first(where: { $0.name == draft.cloud.provider }) {
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Credentials")
                            .font(.headline)
                        ForEach(provider.fields, id: \.self) { field in
                            CredentialField(
                                title: field.replacingOccurrences(of: "_", with: " ").capitalized,
                                value: credentialBinding(field)
                            )
                        }
                    }
                }
            }
        )
    }

    private var networkSection: some View {
        SettingsSection(
            title: "Network",
            systemImage: "network",
            content: {
                Picker("Proxy mode", selection: $draft.proxy.mode) {
                    Text("Direct").tag("none")
                    Text("System").tag("system")
                    Text("Manual").tag("manual")
                }
                .pickerStyle(.segmented)
                if draft.proxy.mode == "manual" {
                    TextField("Proxy URL", text: $draft.proxy.url)
                        .textFieldStyle(.roundedBorder)
                }
                HStack {
                    Button("Test Connection") {
                        Task { await testProxy() }
                    }
                    .disabled(isTestingProxy)
                    if isTestingProxy {
                        ProgressView()
                            .controlSize(.small)
                    }
                }
                if let errorMessage {
                    Text(errorMessage)
                        .foregroundStyle(.red)
                        .font(.callout)
                }
            }
        )
    }

    private var maintenanceSection: some View {
        SettingsSection(
            title: "Maintenance",
            systemImage: "wrench.and.screwdriver",
            content: {
                HStack {
                    Button("Check for Updates") {
                        Task { await checkUpdate() }
                    }
                    .disabled(isCheckingUpdate)
                    if isCheckingUpdate {
                        ProgressView()
                            .controlSize(.small)
                    }
                    Button("Open Log Directory") {
                        Task { await openLogDirectory() }
                    }
                    Button("Open App Data Directory") {
                        Task { await openAppDataDir() }
                    }
                }
                if let updateInfo {
                    VStack(alignment: .leading, spacing: 4) {
                        Text(updateInfo.hasUpdate ? "Update available" : "You are up to date")
                            .font(.headline)
                        Text("Current: \(updateInfo.currentVersion) · Latest: \(updateInfo.latestVersion)")
                            .font(.callout)
                            .foregroundStyle(.secondary)
                        if !updateInfo.releaseNotes.isEmpty {
                            Text(updateInfo.releaseNotes)
                                .font(.caption)
                                .foregroundStyle(.secondary)
                                .lineLimit(5)
                        }
                    }
                }
                if let statusMessage {
                    Text(statusMessage)
                        .foregroundStyle(.secondary)
                        .font(.callout)
                }
            }
        )
    }

    private var footer: some View {
        HStack {
            if isSaving {
                ProgressView()
                    .controlSize(.small)
                Text("Saving settings...")
                    .foregroundStyle(.secondary)
            }
            Spacer()
            Button("Save") {
                Task { await save() }
            }
            .keyboardShortcut("s", modifiers: [.command])
            .disabled(isLoading || isSaving)
        }
        .padding(16)
    }

    private func load() async {
        isLoading = true
        errorMessage = nil
        statusMessage = nil
        do {
            let config: AppSettings = try await client.invoke("settings.get")
            let providerList: [CloudProviderInfo] = try await client.invoke("backup.providers.list")
            draft = config
            providers = providerList
            if draft.cloud.provider.isEmpty, let firstProvider = providerList.first {
                draft.cloud.provider = firstProvider.name
            }
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }

    private func save() async {
        isSaving = true
        errorMessage = nil
        statusMessage = nil
        if !draft.cloud.provider.isEmpty {
            draft.cloudProfiles[draft.cloud.provider] = CloudProviderProfile(
                bucketName: draft.cloud.bucketName,
                remotePath: draft.cloud.remotePath,
                credentials: draft.cloud.credentials
            )
        }
        do {
            let _: NativeEmptyResult = try await client.invoke("settings.save", parameters: draft)
            statusMessage = "Settings saved."
        } catch {
            errorMessage = error.localizedDescription
        }
        isSaving = false
    }

    private func testProxy() async {
        isTestingProxy = true
        errorMessage = nil
        statusMessage = nil
        do {
            let result: ProxyConnectionTestResult = try await client.invoke(
                "proxy.test",
                parameters: ProxyTestParameters(targetURL: "https://github.com", proxy: draft.proxy)
            )
            statusMessage = "\(result.success ? "Connected" : "Failed") in \(result.elapsedMs) ms · \(result.message)"
        } catch {
            errorMessage = error.localizedDescription
        }
        isTestingProxy = false
    }

    private func checkUpdate() async {
        isCheckingUpdate = true
        errorMessage = nil
        statusMessage = nil
        do {
            updateInfo = try await client.invoke("app.update.check")
        } catch {
            errorMessage = error.localizedDescription
        }
        isCheckingUpdate = false
    }

    private func openLogDirectory() async {
        errorMessage = nil
        statusMessage = nil
        do {
            let _: NativeEmptyResult = try await client.invoke("logs.openDir")
            statusMessage = "Log directory opened."
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func openAppDataDir() async {
        errorMessage = nil
        statusMessage = nil
        do {
            let _: NativeEmptyResult = try await client.invoke("app.openAppDataDir")
            statusMessage = "App data directory opened."
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func removeCustomAgent(name: String) async {
        let alert = NSAlert()
        alert.messageText = "Remove Custom Agent"
        alert.informativeText = "Remove \"\(name)\" from the agent list? This action cannot be undone."
        alert.alertStyle = .warning
        alert.addButton(withTitle: "Remove")
        alert.addButton(withTitle: "Cancel")
        guard alert.runModal() == .alertFirstButtonReturn else { return }

        errorMessage = nil
        statusMessage = nil
        do {
            let _: NativeEmptyResult = try await client.invoke("agents.removeCustom", parameters: AgentNameParams(name: name))
            statusMessage = "Removed custom agent: \(name)."
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func credentialBinding(_ field: String) -> Binding<String> {
        Binding(
            get: { draft.cloud.credentials[field] ?? "" },
            set: { draft.cloud.credentials[field] = $0 }
        )
    }

    private func chooseDirectory(_ completion: @escaping (String?) -> Void) {
        let panel = NSOpenPanel()
        panel.canChooseFiles = false
        panel.canChooseDirectories = true
        panel.allowsMultipleSelection = false
        if panel.runModal() == .OK {
            completion(panel.url?.path)
        } else {
            completion(nil)
        }
    }
}

private struct AgentSettingsRow: View {
    @Binding
    var agent: AgentSettings
    let onRemove: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Toggle(agent.name, isOn: $agent.enabled)
                    .font(.headline)
                Spacer()
                if agent.custom {
                    Text("Custom")
                        .font(.caption)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(.quaternary, in: Capsule())
                    Button(role: .destructive, action: onRemove) {
                        Image(systemName: "trash")
                    }
                    .buttonStyle(.borderless)
                }
            }
            TextField("Push directory", text: $agent.pushDir)
            TextField("Scan directories (comma separated)", text: scanDirectories)
            TextField("Memory path", text: $agent.memoryPath)
            TextField("Rules directory", text: $agent.rulesDir)
        }
        .padding(12)
        .background(Color(nsColor: .controlBackgroundColor), in: RoundedRectangle(cornerRadius: 10))
    }

    private var scanDirectories: Binding<String> {
        Binding(
            get: { agent.scanDirs.joined(separator: ", ") },
            set: { value in
                agent.scanDirs = value
                    .split(separator: ",")
                    .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
                    .filter { !$0.isEmpty }
            }
        )
    }
}

private struct CredentialField: View {
    let title: String
    @Binding
    var value: String

    var body: some View {
        HStack {
            Text(title)
                .frame(width: 180, alignment: .leading)
            TextField(title, text: $value)
        }
    }
}

private struct SettingsSection<Content: View>: View {
    let title: String
    let systemImage: String
    let content: Content

    init(
        title: String,
        systemImage: String,
        @ViewBuilder content: () -> Content
    ) {
        self.title = title
        self.systemImage = systemImage
        self.content = content()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            Label(title, systemImage: systemImage)
                .font(.title3.weight(.semibold))
            content
                .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(18)
        .background(Color(nsColor: .windowBackgroundColor), in: RoundedRectangle(cornerRadius: 14))
    }
}

// MARK: - Add Custom Agent Sheet

private struct AddCustomAgentSheet: View {
    let client: DaemonClient
    let onCreated: (String) -> Void

    @State private var name: String = ""
    @State private var pushDir: String = ""
    @State private var isLoading: Bool = false
    @State private var errorMessage: String?
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Add Custom Agent").font(.headline)
            VStack(alignment: .leading, spacing: 4) {
                Text("Agent Name").font(.subheadline).foregroundStyle(.secondary)
                TextField("e.g. my-tool", text: $name)
                    .textFieldStyle(.roundedBorder)
            }
            VStack(alignment: .leading, spacing: 4) {
                Text("Push Directory").font(.subheadline).foregroundStyle(.secondary)
                HStack {
                    TextField("/path/to/push/dir", text: $pushDir)
                        .textFieldStyle(.roundedBorder)
                    Button("Choose...") {
                        chooseDirectory { path in
                            if let path { pushDir = path }
                        }
                    }
                }
            }
            if let errorMessage {
                Text(errorMessage).foregroundStyle(.red).font(.caption)
            }
            HStack {
                Button("Cancel") { dismiss() }
                Spacer()
                Button("Add") {
                    Task { await addAgent() }
                }
                .disabled(name.isEmpty || pushDir.isEmpty || isLoading)
                .buttonStyle(.borderedProminent)
            }
        }
        .padding(24)
        .frame(width: 440)
    }

    private func addAgent() async {
        isLoading = true
        errorMessage = nil
        let trimmedName = name.trimmingCharacters(in: .whitespaces)
        let trimmedDir = pushDir.trimmingCharacters(in: .whitespaces)
        do {
            let _: NativeEmptyResult = try await client.invoke("agents.addCustom", parameters: CustomAgentAddParams(name: trimmedName, pushDir: trimmedDir))
            onCreated(trimmedName)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }

    private func chooseDirectory(_ completion: @escaping (String?) -> Void) {
        let panel = NSOpenPanel()
        panel.canChooseFiles = false
        panel.canChooseDirectories = true
        panel.allowsMultipleSelection = false
        if panel.runModal() == .OK {
            completion(panel.url?.path)
        } else {
            completion(nil)
        }
    }
}
