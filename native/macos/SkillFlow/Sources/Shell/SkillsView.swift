import AppKit
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
    @State private var showAddCategory: Bool = false
    @State private var showMoveSheet: Bool = false
    @State private var skillToMove: InstalledSkill?

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
        .sheet(isPresented: $showAddCategory) {
            AddSkillCategorySheet(client: client) { _ in
                Task { await load() }
            }
        }
        .sheet(isPresented: $showMoveSheet) {
            if let skill = skillToMove {
                MoveCategorySheet(skill: skill, categories: categories) { category in
                    Task { await moveSkillCategory(skill, to: category) }
                    showMoveSheet = false
                }
            }
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
            HStack {
                Text("Categories").fontWeight(.semibold)
                Spacer()
                Button(action: { showAddCategory = true }) {
                    Image(systemName: "plus")
                }
                .buttonStyle(.borderless)
            }
            .padding(8)
            List(selection: $selectedCategory) {
                Text("All").tag(String?.none)
                ForEach(categories, id: \.self) { category in
                    Text(category).tag(Optional(category))
                        .contextMenu {
                            Button("Rename") {
                                Task { await renameCategory(category) }
                            }
                            Button("Delete", role: .destructive) {
                                Task { await deleteCategory(category) }
                            }
                        }
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
                if selectedSkillIDs.count == 1,
                   let id = selectedSkillIDs.first,
                   let skill = skills.first(where: { $0.id == id }) {
                    if skill.updatable {
                        Button("Update \(skill.name)") {
                            Task { await updateSkill(skill) }
                        }
                    }
                    Button("Move to Category...") {
                        skillToMove = skill
                        showMoveSheet = true
                    }
                    Divider()
                }
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

        // Check for missing push directories before pushing.
        do {
            let missing: [MissingPushDir] = try await client.invoke("agents.checkMissingPushDirs", parameters: AgentNamesParams(agentNames: agentNames))
            if !missing.isEmpty {
                let dirList = missing.map { "• \($0.name): \($0.dir)" }.joined(separator: "\n")
                let alert = NSAlert()
                alert.messageText = "Missing Push Directories"
                alert.informativeText = "The following agents have missing push directories:\n\n\(dirList)\n\nCreate them and continue?"
                alert.alertStyle = .warning
                alert.addButton(withTitle: "Create and Push")
                alert.addButton(withTitle: "Cancel")
                guard alert.runModal() == .alertFirstButtonReturn else {
                    isPushing = false
                    return
                }
            }
        } catch {
            errorMessage = error.localizedDescription
            isPushing = false
            return
        }

        let params = SkillsPushParams(skillIDs: ids, agentNames: agentNames)
        do {
            let conflicts: [PushConflict] = try await client.invoke("skills.push", parameters: params)
            if conflicts.isEmpty {
                statusMessage = "Pushed \(ids.count) skill\(ids.count == 1 ? "" : "s") to \(agentNames.count) agent\(agentNames.count == 1 ? "" : "s")."
                showPushSheet = false
            } else {
                let conflictList = conflicts.map { "• \($0.skillName) → \($0.agentName)" }.joined(separator: "\n")
                let alert = NSAlert()
                alert.messageText = "Push Conflicts"
                alert.informativeText = "\(conflicts.count) conflict\(conflicts.count == 1 ? "" : "s") detected:\n\n\(conflictList)\n\nForce push to overwrite?"
                alert.alertStyle = .warning
                alert.addButton(withTitle: "Force Push")
                alert.addButton(withTitle: "Skip")
                if alert.runModal() == .alertFirstButtonReturn {
                    let _: NativeEmptyResult = try await client.invoke("skills.pushForce", parameters: params)
                    statusMessage = "Force pushed \(ids.count) skill\(ids.count == 1 ? "" : "s")."
                    showPushSheet = false
                } else {
                    statusMessage = "Pushed with \(conflicts.count) conflict\(conflicts.count == 1 ? "" : "s") skipped."
                    showPushSheet = false
                }
            }
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
        isPushing = false
    }

    private func updateSkill(_ skill: InstalledSkill) async {
        errorMessage = nil
        statusMessage = nil
        do {
            let _: NativeEmptyResult = try await client.invoke("skills.updateOne", parameters: SkillsDeleteParams(skillID: skill.id))
            statusMessage = "Updated skill: \(skill.name)"
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func moveSkillCategory(_ skill: InstalledSkill, to category: String) async {
        errorMessage = nil
        statusMessage = nil
        do {
            let _: NativeEmptyResult = try await client.invoke("skills.moveCategory", parameters: SkillsMoveCategoryParams(skillID: skill.id, category: category))
            statusMessage = "Moved \(skill.name) to \(category)."
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func renameCategory(_ oldName: String) async {
        let alert = NSAlert()
        alert.messageText = "Rename Category"
        alert.informativeText = "Enter a new name for \"\(oldName)\":"
        alert.alertStyle = .informational
        alert.addButton(withTitle: "Rename")
        alert.addButton(withTitle: "Cancel")
        let input = NSTextField(frame: NSRect(x: 0, y: 0, width: 250, height: 24))
        input.stringValue = oldName
        alert.accessoryView = input
        guard alert.runModal() == .alertFirstButtonReturn else { return }
        let newName = input.stringValue.trimmingCharacters(in: .whitespaces)
        guard !newName.isEmpty, newName != oldName else { return }
        do {
            let _: NativeEmptyResult = try await client.invoke("skills.categories.rename", parameters: SkillsCategoryRenameParams(oldName: oldName, newName: newName))
            statusMessage = "Renamed category to \(newName)."
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func deleteCategory(_ name: String) async {
        let alert = NSAlert()
        alert.messageText = "Delete Category?"
        alert.informativeText = "Delete \"\(name)\"? Skills in this category will be moved to the default category."
        alert.alertStyle = .warning
        alert.addButton(withTitle: "Delete")
        alert.addButton(withTitle: "Cancel")
        guard alert.runModal() == .alertFirstButtonReturn else { return }
        do {
            let _: NativeEmptyResult = try await client.invoke("skills.categories.delete", parameters: SkillsCategoryNameParams(name: name))
            statusMessage = "Deleted category: \(name)."
            if selectedCategory == name { selectedCategory = nil }
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
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

// MARK: - Add Skill Category Sheet

private struct AddSkillCategorySheet: View {
    let client: DaemonClient
    let onCreated: (String) -> Void

    @State private var name: String = ""
    @State private var isLoading: Bool = false
    @State private var errorMessage: String?
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        VStack(spacing: 16) {
            Text("New Category").font(.headline)
            TextField("Category name", text: $name)
                .textFieldStyle(.roundedBorder)
            if let errorMessage {
                Text(errorMessage).foregroundStyle(.red).font(.caption)
            }
            HStack {
                Button("Cancel") { dismiss() }
                Button("Create") {
                    Task { await createCategory() }
                }
                .disabled(name.isEmpty || isLoading)
                .buttonStyle(.borderedProminent)
            }
        }
        .padding(24)
        .frame(width: 400)
    }

    private func createCategory() async {
        isLoading = true
        errorMessage = nil
        let trimmed = name.trimmingCharacters(in: .whitespaces)
        do {
            let _: NativeEmptyResult = try await client.invoke("skills.categories.create", parameters: SkillsCategoryNameParams(name: trimmed))
            onCreated(trimmed)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }
}

// MARK: - Move Category Sheet

private struct MoveCategorySheet: View {
    let skill: InstalledSkill
    let categories: [String]
    let onMove: (String) -> Void

    @State private var selectedCategory: String = ""
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Move Skill to Category").font(.headline)
            Text("Move \"\(skill.name)\" to a different category:")
                .foregroundStyle(.secondary)
                .font(.subheadline)
            Picker("Category", selection: $selectedCategory) {
                ForEach(categories, id: \.self) { category in
                    Text(category).tag(category)
                }
            }
            .pickerStyle(.menu)
            HStack {
                Button("Cancel") { dismiss() }
                Spacer()
                Button("Move") {
                    onMove(selectedCategory)
                }
                .disabled(selectedCategory.isEmpty)
                .buttonStyle(.borderedProminent)
            }
        }
        .padding(24)
        .frame(width: 400)
        .onAppear {
            if selectedCategory.isEmpty {
                selectedCategory = categories.first ?? skill.category
            }
        }
    }
}
