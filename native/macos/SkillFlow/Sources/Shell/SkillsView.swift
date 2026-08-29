import AppKit
import SkillFlowCore
import SwiftUI

struct SkillsView: View {
    @State private var skills: [InstalledSkill] = []
    @State private var categories: [String] = []
    @State private var agents: [AgentConfig] = []
    @State private var selectedCategory: String? = nil
    @State private var searchText: String = ""
    @State private var sortAscending: Bool = true
    @State private var selectedSkillIDs: Set<String> = []
    @State private var isLoading: Bool = true
    @State private var errorMessage: String?
    @State private var statusMessage: String?
    @State private var isImporting: Bool = false
    @State private var isPushing: Bool = false
    @State private var isCheckingUpdates: Bool = false
    @State private var showPushSheet: Bool = false

    private let client: DaemonClient

    init(client: DaemonClient = DaemonClient()) {
        self.client = client
    }

    var body: some View {
        VStack(spacing: 0) {
            header
            Divider()
            if isLoading {
                ProgressView("Loading skills...")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                content
            }
            Divider()
            footer
        }
        .frame(minWidth: 720, minHeight: 480)
        .task { await load() }
        .sheet(isPresented: $showPushSheet) {
            PushSheet(
                skills: filteredSkills.filter { selectedSkillIDs.contains($0.id) },
                agents: agents,
                isPushing: $isPushing,
                onPush: { agentNames in
                    await push(agentNames: agentNames)
                }
            )
        }
    }

    // MARK: - Header

    private var header: some View {
        HStack(spacing: 12) {
            Image(systemName: "square.grid.2x2")
                .font(.title2)
                .foregroundStyle(.secondary)
            VStack(alignment: .leading, spacing: 2) {
                Text("My Skills")
                    .font(.title2.weight(.semibold))
                Text("\(skills.count) skill\(skills.count == 1 ? "" : "s")")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
            Spacer()
            Button(action: { Task { await checkUpdates() } }) {
                Label("Check Updates", systemImage: "arrow.triangle.2.circlepath")
            }
            .disabled(isCheckingUpdates || skills.isEmpty)
            if isCheckingUpdates {
                ProgressView().controlSize(.small)
            }
            Button(action: { Task { await importLocal() } }) {
                Label("Import", systemImage: "square.and.arrow.down")
            }
            .disabled(isImporting)
            if isImporting {
                ProgressView().controlSize(.small)
            }
        }
        .padding(16)
    }

    // MARK: - Content

    private var content: some View {
        HStack(spacing: 0) {
            categorySidebar
            Divider()
            skillList
        }
    }

    private var categorySidebar: some View {
        VStack(alignment: .leading, spacing: 0) {
            List(selection: $selectedCategory) {
                Text("All").tag(String?.none)
                ForEach(categories, id: \.self) { category in
                    Text(category).tag(Optional(category))
                }
            }
            .frame(width: 200)
        }
    }

    private var skillList: some View {
        VStack(spacing: 0) {
            toolbar
            Divider()
            if filteredSkills.isEmpty {
                emptyState
            } else {
                skillTable
            }
        }
    }

    private var toolbar: some View {
        HStack(spacing: 12) {
            TextField("Search skills...", text: $searchText)
                .textFieldStyle(.roundedBorder)
                .frame(maxWidth: 300)
            Spacer()
            Button(action: {
                sortAscending.toggle()
            }) {
                Label(sortAscending ? "A-Z" : "Z-A", systemImage: "arrow.up.arrow.down")
            }
            Menu {
                Button("Delete Selected") {
                    Task { await deleteSelected() }
                }
                .disabled(selectedSkillIDs.isEmpty)
                Button("Push Selected...") {
                    showPushSheet = true
                }
                .disabled(selectedSkillIDs.isEmpty || agents.isEmpty)
            } label: {
                Label("Actions", systemImage: "ellipsis.circle")
            }
            .disabled(selectedSkillIDs.isEmpty)
        }
        .padding(12)
    }

    private var skillTable: some View {
        Table(filteredSkills, selection: $selectedSkillIDs) {
            TableColumn("Name") { skill in
                HStack(spacing: 8) {
                    Image(systemName: skill.updatable ? "arrow.up.circle.fill" : "doc.text")
                        .foregroundStyle(skill.updatable ? .orange : .secondary)
                    Text(skill.name)
                    if skill.pushed {
                        Image(systemName: "checkmark.circle.fill")
                            .foregroundStyle(.green)
                            .font(.caption)
                    }
                }
            }
            TableColumn("Category") { skill in
                Text(skill.category)
                    .foregroundStyle(.secondary)
            }
            TableColumn("Source") { skill in
                Text(skill.source)
                    .foregroundStyle(.secondary)
            }
            TableColumn("Agents") { skill in
                if skill.pushedAgents.isEmpty {
                    Text("—")
                        .foregroundStyle(.tertiary)
                } else {
                    Text(skill.pushedAgents.joined(separator: ", "))
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }
        }
    }

    private var emptyState: some View {
        VStack(spacing: 12) {
            Image(systemName: "square.grid.2x2")
                .font(.system(size: 36))
                .foregroundStyle(.tertiary)
            Text(searchText.isEmpty ? "No skills installed" : "No skills match your search")
                .font(.headline)
            if searchText.isEmpty {
                Button("Import a skill folder") {
                    Task { await importLocal() }
                }
                .buttonStyle(.borderedProminent)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    // MARK: - Footer

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

    // MARK: - Computed

    private var filteredSkills: [InstalledSkill] {
        var result = skills
        if let selectedCategory {
            result = result.filter { $0.category == selectedCategory }
        }
        if !searchText.isEmpty {
            result = result.filter { $0.name.localizedCaseInsensitiveContains(searchText) }
        }
        result.sort { sortAscending ? $0.name < $1.name : $0.name > $1.name }
        return result
    }

    // MARK: - Actions

    private func load() async {
        isLoading = true
        errorMessage = nil
        do {
            let skillsList: [InstalledSkill] = try await client.invoke("skills.list")
            let categoriesList: [String] = try await client.invoke("skills.categories.list")
            let agentsList: [AgentConfig] = try await client.invoke("agents.listEnabled")
            skills = skillsList
            categories = categoriesList
            agents = agentsList
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }

    private func importLocal() async {
        let panel = NSOpenPanel()
        panel.canChooseFiles = false
        panel.canChooseDirectories = true
        panel.allowsMultipleSelection = false
        guard panel.runModal() == .OK, let url = panel.url else { return }

        isImporting = true
        errorMessage = nil
        statusMessage = nil
        let params = SkillsImportLocalParams(dir: url.path, category: selectedCategory ?? "Uncategorized")
        do {
            let _: NativeEmptyResult = try await client.invoke("skills.importLocal", parameters: params)
            statusMessage = "Imported skill from \(url.lastPathComponent)."
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
        isImporting = false
    }

    private func deleteSelected() async {
        let ids = Array(selectedSkillIDs)
        guard !ids.isEmpty else { return }

        let alert = NSAlert()
        alert.messageText = "Delete \(ids.count) skill\(ids.count == 1 ? "" : "s")?"
        alert.informativeText = "This action cannot be undone."
        alert.alertStyle = .warning
        alert.addButton(withTitle: "Delete")
        alert.addButton(withTitle: "Cancel")
        guard alert.runModal() == .alertFirstButtonReturn else { return }

        errorMessage = nil
        statusMessage = nil
        do {
            if ids.count == 1 {
                let _: NativeEmptyResult = try await client.invoke("skills.delete", parameters: SkillsDeleteParams(skillID: ids[0]))
            } else {
                let _: NativeEmptyResult = try await client.invoke("skills.deleteBatch", parameters: SkillsDeleteBatchParams(skillIDs: ids))
            }
            statusMessage = "Deleted \(ids.count) skill\(ids.count == 1 ? "" : "s")."
            selectedSkillIDs.removeAll()
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func checkUpdates() async {
        isCheckingUpdates = true
        errorMessage = nil
        statusMessage = nil
        do {
            let _: NativeEmptyResult = try await client.invoke("skills.updateCheck")
            let updatable = skills.filter { $0.updatable }
            statusMessage = updatable.isEmpty ? "All skills are up to date." : "\(updatable.count) skill\(updatable.count == 1 ? "" : "s") can be updated."
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
        isCheckingUpdates = false
    }

    private func push(agentNames: [String]) async {
        let ids = Array(selectedSkillIDs)
        guard !ids.isEmpty, !agentNames.isEmpty else { return }

        isPushing = true
        errorMessage = nil
        statusMessage = nil
        let params = SkillsPushParams(skillIDs: ids, agentNames: agentNames)
        do {
            let _: NativeEmptyResult = try await client.invoke("skills.push", parameters: params)
            statusMessage = "Pushed \(ids.count) skill\(ids.count == 1 ? "" : "s") to \(agentNames.count) agent\(agentNames.count == 1 ? "" : "s")."
            showPushSheet = false
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
        isPushing = false
    }
}

// MARK: - Push Sheet

private struct PushSheet: View {
    let skills: [InstalledSkill]
    let agents: [AgentConfig]
    @Binding var isPushing: Bool
    let onPush: ([String]) async -> Void

    @State private var selectedAgents: Set<String> = []

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Push Skills to Agents")
                .font(.headline)

            VStack(alignment: .leading, spacing: 4) {
                Text("Skills:")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                ForEach(skills) { skill in
                    Text("• \(skill.name)")
                        .font(.callout)
                }
            }

            VStack(alignment: .leading, spacing: 4) {
                Text("Select target agents:")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                ForEach(agents) { agent in
                    Toggle(agent.name, isOn: Binding(
                        get: { selectedAgents.contains(agent.name) },
                        set: { isSelected in
                            if isSelected {
                                selectedAgents.insert(agent.name)
                            } else {
                                selectedAgents.remove(agent.name)
                            }
                        }
                    ))
                }
            }

            HStack {
                Spacer()
                Button("Cancel") {
                    selectedAgents.removeAll()
                }
                Button("Push") {
                    Task { await onPush(Array(selectedAgents)) }
                }
                .keyboardShortcut(.return)
                .disabled(selectedAgents.isEmpty || isPushing)
            }
        }
        .padding(24)
        .frame(width: 400)
    }
}
