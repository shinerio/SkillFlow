import SkillFlowCore
import SwiftUI

struct MemoryView: View {
    @State private var mainMemory: MainMemory?
    @State private var modules: [ModuleMemory] = []
    @State private var pushConfigs: [MemoryPushConfig] = []
    @State private var pushStatuses: [PushStatus] = []
    @State private var isLoading: Bool = true
    @State private var errorMessage: String?
    @State private var statusMessage: String?
    @State private var isPushingAll: Bool = false

    private let client: DaemonClient

    init(client: DaemonClient = DaemonClient()) {
        self.client = client
    }

    var body: some View {
        VStack(spacing: 0) {
            header
            Divider()
            if isLoading {
                ProgressView("Loading memory...")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                ScrollView {
                    VStack(alignment: .leading, spacing: 24) {
                        mainMemorySection
                        modulesSection
                        pushConfigSection
                        pushStatusSection
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
            Image(systemName: "brain")
                .font(.title2)
                .foregroundStyle(.secondary)
            VStack(alignment: .leading, spacing: 2) {
                Text("My Memory")
                    .font(.title2.weight(.semibold))
                Text("Manage main memory, modules, and push configuration.")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
            Spacer()
            Button(action: { Task { await pushAll() } }) {
                Label("Push All", systemImage: "arrow.up.circle")
            }
            .disabled(isPushingAll)
            if isPushingAll {
                ProgressView().controlSize(.small)
            }
        }
        .padding(16)
    }

    private var mainMemorySection: some View {
        VStack(alignment: .leading, spacing: 12) {
            Label("Main Memory", systemImage: "doc.text")
                .font(.title3.weight(.semibold))
            if let mainMemory {
                MemoryEditor(
                    title: "Main",
                    initialContent: mainMemory.content,
                    onSave: { content in
                        let params = MemoryContentParams(content: content)
                        let _: NativeEmptyResult = try await client.invoke("memory.main.save", parameters: params)
                    },
                    onOpenEditor: {
                        let params = MemoryOpenEditorParams(memoryType: "main", moduleName: "")
                        let _: NativeEmptyResult = try await client.invoke("memory.openInEditor", parameters: params)
                    },
                    client: client,
                    onSaved: { await load() }
                )
            }
        }
        .padding(18)
        .background(Color(nsColor: .windowBackgroundColor), in: RoundedRectangle(cornerRadius: 14))
    }

    private var modulesSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Label("Module Memories", systemImage: "square.stack.3d.up")
                    .font(.title3.weight(.semibold))
                Spacer()
                Button(action: { Task { await createModule() } }) {
                    Label("New Module", systemImage: "plus")
                }
            }
            if modules.isEmpty {
                Text("No module memories configured.")
                    .foregroundStyle(.secondary)
            } else {
                ForEach(modules) { module in
                    ModuleMemoryRow(
                        module: module,
                        onSave: { content in
                            let params = ModuleMemoryContentParams(name: module.name, content: content)
                            let _: NativeEmptyResult = try await client.invoke("memory.modules.save", parameters: params)
                        },
                        onToggle: { enabled in
                            let params = ModuleMemoryEnabledParams(name: module.name, enabled: enabled)
                            let _: NativeEmptyResult = try await client.invoke("memory.modules.setEnabled", parameters: params)
                        },
                        onDelete: {
                            let params = ModuleMemoryNameParams(name: module.name)
                            let _: NativeEmptyResult = try await client.invoke("memory.modules.delete", parameters: params)
                        },
                        onOpenEditor: {
                            let params = MemoryOpenEditorParams(memoryType: "module", moduleName: module.name)
                            let _: NativeEmptyResult = try await client.invoke("memory.openInEditor", parameters: params)
                        },
                        client: client,
                        onChanged: { await load() }
                    )
                }
            }
        }
        .padding(18)
        .background(Color(nsColor: .windowBackgroundColor), in: RoundedRectangle(cornerRadius: 14))
    }

    private var pushConfigSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            Label("Push Configuration", systemImage: "arrow.up.arrow.down.circle")
                .font(.title3.weight(.semibold))
            if pushConfigs.isEmpty {
                Text("No push configurations found.")
                    .foregroundStyle(.secondary)
            } else {
                ForEach(pushConfigs) { config in
                    HStack {
                        Text(config.agentType)
                            .font(.headline)
                        Spacer()
                        Picker("Mode", selection: .constant(config.mode)) {
                            Text("Merge").tag("merge")
                            Text("Takeover").tag("takeover")
                        }
                        .pickerStyle(.segmented)
                        .frame(width: 200)
                        .disabled(true)
                        Toggle("Auto Push", isOn: .constant(config.autoPush))
                            .disabled(true)
                    }
                    .padding(10)
                    .background(Color(nsColor: .controlBackgroundColor), in: RoundedRectangle(cornerRadius: 8))
                }
            }
        }
        .padding(18)
        .background(Color(nsColor: .windowBackgroundColor), in: RoundedRectangle(cornerRadius: 14))
    }

    private var pushStatusSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            Label("Push Status", systemImage: "checkmark.circle")
                .font(.title3.weight(.semibold))
            if pushStatuses.isEmpty {
                Text("No push status available.")
                    .foregroundStyle(.secondary)
            } else {
                ForEach(pushStatuses) { status in
                    HStack {
                        Text(status.agentType)
                        Spacer()
                        Text(status.status)
                            .font(.caption)
                            .padding(.horizontal, 8)
                            .padding(.vertical, 2)
                            .background(statusColor(status.status).opacity(0.2), in: Capsule())
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

    private func statusColor(_ status: String) -> Color {
        switch status {
        case "synced": .green
        case "pendingPush": .orange
        default: .gray
        }
    }

    private func load() async {
        isLoading = true
        errorMessage = nil
        do {
            let main: MainMemory = try await client.invoke("memory.main.get")
            mainMemory = main
            let mods: [ModuleMemory] = try await client.invoke("memory.modules.list")
            modules = mods
            let configs: [MemoryPushConfig] = try await client.invoke("memory.pushConfig.getAll")
            pushConfigs = configs
            let statuses: [PushStatus] = try await client.invoke("memory.pushStatus.getAll")
            pushStatuses = statuses
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }

    private func pushAll() async {
        isPushingAll = true
        errorMessage = nil
        statusMessage = nil
        do {
            let results: [PushResult] = try await client.invoke("memory.pushAll")
            let successCount = results.filter { $0.success }.count
            statusMessage = "Pushed to \(successCount) of \(results.count) agent\(results.count == 1 ? "" : "s")."
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
        isPushingAll = false
    }

    private func createModule() async {
        let alert = NSAlert()
        alert.messageText = "New Module Memory"
        alert.informativeText = "Enter a name for the module:"
        alert.addButton(withTitle: "Create")
        alert.addButton(withTitle: "Cancel")

        let input = NSTextField(frame: NSRect(x: 0, y: 0, width: 300, height: 24))
        alert.accessoryView = input

        guard alert.runModal() == .alertFirstButtonReturn else { return }
        let name = input.stringValue.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !name.isEmpty else { return }

        do {
            let params = ModuleMemoryContentParams(name: name, content: "")
            let _: ModuleMemory = try await client.invoke("memory.modules.create", parameters: params)
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

private struct MemoryEditor: View {
    let title: String
    let initialContent: String
    let onSave: (String) async throws -> Void
    let onOpenEditor: () async throws -> Void
    let client: DaemonClient
    let onSaved: () async -> Void

    @State private var content: String = ""
    @State private var isSaving: Bool = false
    @State private var error: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            TextEditor(text: $content)
                .font(.body.monospaced())
                .frame(minHeight: 120)
                .border(Color(nsColor: .separatorColor))
            HStack {
                if let error {
                    Text(error)
                        .foregroundStyle(.red)
                        .font(.caption)
                }
                Spacer()
                Button("Open in Editor") {
                    Task { try? await onOpenEditor() }
                }
                Button("Save") {
                    Task {
                        isSaving = true
                        error = nil
                        do {
                            try await onSave(content)
                            await onSaved()
                        } catch {
                            self.error = error.localizedDescription
                        }
                        isSaving = false
                    }
                }
                .disabled(isSaving || content == initialContent)
            }
        }
        .onAppear { content = initialContent }
    }
}

private struct ModuleMemoryRow: View {
    let module: ModuleMemory
    let onSave: (String) async throws -> Void
    let onToggle: (Bool) async throws -> Void
    let onDelete: () async throws -> Void
    let onOpenEditor: () async throws -> Void
    let client: DaemonClient
    let onChanged: () async -> Void

    @State private var isExpanded: Bool = false
    @State private var content: String = ""
    @State private var isSaving: Bool = false
    @State private var error: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Toggle("", isOn: Binding(
                    get: { module.enabled },
                    set: { newValue in
                        Task { try? await onToggle(newValue); await onChanged() }
                    }
                ))
                .labelsHidden()
                Button(action: { isExpanded.toggle() }) {
                    HStack {
                        Image(systemName: isExpanded ? "chevron.down" : "chevron.right")
                            .font(.caption)
                        Text(module.name)
                            .font(.headline)
                    }
                }
                .buttonStyle(.plain)
                Spacer()
                Button(action: { Task { try? await onOpenEditor() } }) {
                    Image(systemName: "pencil.and.list.clipboard")
                }
                .buttonStyle(.borderless)
                Button(role: .destructive, action: { Task { try? await onDelete(); await onChanged() } }) {
                    Image(systemName: "trash")
                }
                .buttonStyle(.borderless)
            }
            if isExpanded {
                TextEditor(text: $content)
                    .font(.body.monospaced())
                    .frame(minHeight: 100)
                    .border(Color(nsColor: .separatorColor))
                HStack {
                    if let error {
                        Text(error)
                            .foregroundStyle(.red)
                            .font(.caption)
                    }
                    Spacer()
                    Button("Save") {
                        Task {
                            isSaving = true
                            error = nil
                            do {
                                try await onSave(content)
                                await onChanged()
                            } catch {
                                self.error = error.localizedDescription
                            }
                            isSaving = false
                        }
                    }
                    .disabled(isSaving)
                }
            }
        }
        .padding(12)
        .background(Color(nsColor: .controlBackgroundColor), in: RoundedRectangle(cornerRadius: 10))
        .onAppear { content = module.content }
    }
}
