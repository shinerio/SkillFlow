import SwiftUI

struct StarredReposView: View {
    @State private var repos: [StarRepo] = []
    @State private var skills: [StarSkillEntry] = []
    @State private var categories: [String] = []
    @State private var agents: [AgentConfig] = []
    @State private var selectedRepo: StarRepo?
    @State private var searchText: String = ""
    @State private var sortAscending: Bool = true
    @State private var selectedSkillPaths: Set<String> = []
    @State private var isLoading: Bool = true
    @State private var isUpdating: Bool = false
    @State private var isPushing: Bool = false
    @State private var errorMessage: String?
    @State private var statusMessage: String?
    @State private var showAddRepo: Bool = false
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
                ProgressView("Loading starred repos...")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                content
            }
            Divider()
            footer
        }
        .frame(minWidth: 720, minHeight: 480)
        .task { await load() }
        .sheet(isPresented: $showAddRepo) {
            AddRepoSheet(client: client) { _ in
                Task { await load() }
            }
        }
        .sheet(isPresented: $showPushSheet) {
            StarredPushSheet(
                skills: filteredSkills.filter { selectedSkillPaths.contains($0.path) },
                agents: agents,
                isPushing: $isPushing,
                onPush: { agentNames in
                    await push(agentNames: agentNames)
                }
            )
        }
    }

    private var header: some View {
        HStack {
            Image(systemName: "star.fill")
                .foregroundStyle(.secondary)
            VStack(alignment: .leading, spacing: 2) {
                Text("Starred Repos").font(.title2.bold())
                Text("\(repos.count) repo\(repos.count == 1 ? "" : "s")")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
            Spacer()
            Button(action: { Task { await updateAll() } }) {
                Label("Update All", systemImage: "arrow.clockwise")
            }
            .disabled(isUpdating || repos.isEmpty)
            Button(action: { showAddRepo = true }) {
                Label("Add Repo", systemImage: "plus")
            }
        }
        .padding()
    }

    private var content: some View {
        HStack(spacing: 0) {
            repoSidebar
            Divider()
            skillArea
        }
    }

    private var repoSidebar: some View {
        VStack(spacing: 0) {
            List(selection: $selectedRepo) {
                ForEach(repos) { repo in
                    VStack(alignment: .leading, spacing: 2) {
                        Text(repo.name).fontWeight(.semibold)
                        Text(repo.source).font(.caption).foregroundStyle(.secondary)
                        Text(repo.syncStatusText).font(.caption2).foregroundStyle(.secondary)
                    }
                    .tag(repo)
                }
            }
            .frame(width: 280)
        }
    }

    private var skillArea: some View {
        VStack(spacing: 0) {
            skillToolbar
            Divider()
            if filteredSkills.isEmpty {
                VStack(spacing: 8) {
                    Image(systemName: "tray").font(.largeTitle).foregroundStyle(.secondary)
                    Text("No skills found").foregroundStyle(.secondary)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                List(filteredSkills, selection: $selectedSkillPaths) { skill in
                    HStack {
                        Image(systemName: "doc.text")
                            .foregroundStyle(.secondary)
                        VStack(alignment: .leading) {
                            Text(skill.name).fontWeight(.semibold)
                            Text(skill.subPath).font(.caption).foregroundStyle(.secondary)
                        }
                        Spacer()
                        Text(skill.statusText)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }
        }
    }

    private var skillToolbar: some View {
        HStack {
            TextField("Search skills...", text: $searchText)
                .textFieldStyle(.roundedBorder)
                .frame(maxWidth: 300)
            Spacer()
            Button(sortAscending ? "A-Z" : "Z-A") {
                sortAscending.toggle()
            }
            if !selectedSkillPaths.isEmpty {
                Button(action: { showPushSheet = true }) {
                    Label("Push Selected", systemImage: "arrow.up.circle")
                }
                Button(action: { Task { await importSelected() } }) {
                    Label("Import Selected", systemImage: "square.and.arrow.down")
                }
            }
        }
        .padding(8)
    }

    private var footer: some View {
        HStack {
            if let status = statusMessage {
                Text(status).foregroundStyle(.secondary).font(.caption)
            } else if let error = errorMessage {
                Text(error).foregroundStyle(.red).font(.caption)
            }
            Spacer()
            Button("Reload") { Task { await load() } }
        }
        .padding()
    }

    private var filteredSkills: [StarSkillEntry] {
        var result = skills
        if !searchText.isEmpty {
            result = result.filter { $0.name.localizedCaseInsensitiveContains(searchText) }
        }
        return sortAscending
            ? result.sorted { $0.name < $1.name }
            : result.sorted { $0.name > $1.name }
    }

    private func load() async {
        isLoading = true
        errorMessage = nil
        statusMessage = nil
        do {
            repos = try await client.invoke("starred.listRepos")
            categories = try await client.invoke("skills.categories.list")
            agents = try await client.invoke("agents.listEnabled")
            await loadSkills()
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }

    private func loadSkills() async {
        do {
            if let repo = selectedRepo {
                skills = try await client.invoke("starred.listRepoSkills", parameters: StarredRepoURLParams(repoURL: repo.url))
            } else {
                skills = try await client.invoke("starred.listAllSkills")
            }
            selectedSkillPaths.removeAll()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func updateAll() async {
        isUpdating = true
        errorMessage = nil
        statusMessage = nil
        do {
            let _: NativeEmptyResult = try await client.invoke("starred.updateAll")
            statusMessage = "All repos updated."
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
        isUpdating = false
    }

    private func importSelected() async {
        let selected = skills.filter { selectedSkillPaths.contains($0.path) }
        guard !selected.isEmpty else { return }
        let category = "Uncategorized"
        let repoUrl = selectedRepo?.url ?? selected.first?.repoUrl ?? ""
        do {
            let _: NativeEmptyResult = try await client.invoke("starred.importSkills", parameters: StarredImportParams(
                skillPaths: selected.map(\.path),
                repoURL: repoUrl,
                category: category
            ))
            statusMessage = "Imported \(selected.count) skill\(selected.count == 1 ? "" : "s")."
            await loadSkills()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func push(agentNames: [String]) async {
        let selected = skills.filter { selectedSkillPaths.contains($0.path) }
        guard !selected.isEmpty, !agentNames.isEmpty else { return }
        isPushing = true
        errorMessage = nil
        statusMessage = nil
        do {
            let _: NativeEmptyResult = try await client.invoke("starred.pushToAgents", parameters: StarredPushParams(
                skillPaths: selected.map(\.path),
                agentNames: agentNames
            ))
            statusMessage = "Pushed \(selected.count) skill\(selected.count == 1 ? "" : "s") to \(agentNames.count) agent\(agentNames.count == 1 ? "" : "s")."
            await loadSkills()
        } catch {
            errorMessage = error.localizedDescription
        }
        isPushing = false
    }
}

private struct AddRepoSheet: View {
    let client: DaemonClient
    let onAdded: (String) -> Void

    @State private var repoURL: String = ""
    @State private var useCredentials: Bool = false
    @State private var username: String = ""
    @State private var password: String = ""
    @State private var isLoading: Bool = false
    @State private var errorMessage: String?
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        VStack(spacing: 16) {
            Text("Add Starred Repo").font(.headline)
            TextField("Repository URL", text: $repoURL)
                .textFieldStyle(.roundedBorder)
            Toggle("Use credentials", isOn: $useCredentials)
            if useCredentials {
                TextField("Username", text: $username)
                    .textFieldStyle(.roundedBorder)
                SecureField("Password", text: $password)
                    .textFieldStyle(.roundedBorder)
            }
            if let error = errorMessage {
                Text(error).foregroundStyle(.red).font(.caption)
            }
            HStack {
                Button("Cancel") { dismiss() }
                Button("Add") {
                    Task { await addRepo() }
                }
                .disabled(repoURL.isEmpty || isLoading)
                .buttonStyle(.borderedProminent)
            }
        }
        .padding()
        .frame(width: 400)
    }

    private func addRepo() async {
        isLoading = true
        errorMessage = nil
        do {
            if useCredentials {
                let _: NativeEmptyResult = try await client.invoke("starred.addRepoWithCredentials", parameters: StarredRepoCredentialsParams(
                    repoURL: repoURL.trimmingCharacters(in: .whitespaces),
                    username: username,
                    password: password
                ))
            } else {
                let _: NativeEmptyResult = try await client.invoke("starred.addRepo", parameters: StarredRepoURLParams(
                    repoURL: repoURL.trimmingCharacters(in: .whitespaces)
                ))
            }
            onAdded(repoURL)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }
}

private struct StarredPushSheet: View {
    let skills: [StarSkillEntry]
    let agents: [AgentConfig]
    @Binding var isPushing: Bool
    let onPush: ([String]) async -> Void

    @State private var selectedAgents: Set<String> = []
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        VStack(spacing: 16) {
            Text("Push Skills to Agents").font(.headline)
            Text("Pushing \(skills.count) skill\(skills.count == 1 ? "" : "s"):")
                .foregroundStyle(.secondary)
            Text(skills.map(\.name).joined(separator: ", "))
                .font(.caption)
                .foregroundStyle(.secondary)
            Divider()
            Text("Select target agents:")
            ForEach(agents) { agent in
                Toggle(agent.name, isOn: Binding(
                    get: { selectedAgents.contains(agent.name) },
                    set: { isSelected in
                        if isSelected { selectedAgents.insert(agent.name) }
                        else { selectedAgents.remove(agent.name) }
                    }
                ))
            }
            HStack {
                Button("Cancel") { dismiss() }
                Button("Push") {
                    Task {
                        await onPush(Array(selectedAgents))
                        dismiss()
                    }
                }
                .disabled(selectedAgents.isEmpty || isPushing)
                .buttonStyle(.borderedProminent)
            }
        }
        .padding()
        .frame(width: 400)
    }
}
