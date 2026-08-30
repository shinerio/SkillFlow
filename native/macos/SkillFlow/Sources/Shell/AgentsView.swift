import SwiftUI

struct AgentsView: View {
    @State private var agents: [AgentConfig] = []
    @State private var selectedAgent: AgentConfig? = nil
    @State private var isLoading: Bool = true
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
                ProgressView("Loading agents...")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else if agents.isEmpty {
                emptyState
            } else {
                content
            }
            Divider()
            footer
        }
        .frame(minWidth: 720, minHeight: 480)
        .task { await load() }
    }

    private var header: some View {
        HStack(spacing: 12) {
            Image(systemName: "person.2")
                .font(.title2)
                .foregroundStyle(.secondary)
            VStack(alignment: .leading, spacing: 2) {
                Text("My Agents")
                    .font(.title2.weight(.semibold))
                Text("\(agents.count) agent\(agents.count == 1 ? "" : "s")")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
            Spacer()
        }
        .padding(16)
    }

    private var emptyState: some View {
        VStack(spacing: 12) {
            Image(systemName: "person.2")
                .font(.system(size: 36))
                .foregroundStyle(.tertiary)
            Text("No agents configured")
                .font(.headline)
            Text("Configure agents in Settings to get started.")
                .font(.body)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private var content: some View {
        HStack(spacing: 0) {
            agentSidebar
            Divider()
            if let selectedAgent {
                AgentDetail(agent: selectedAgent, client: client)
            } else {
                Text("Select an agent")
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
        }
    }

    private var agentSidebar: some View {
        List(selection: $selectedAgent) {
            ForEach(agents) { agent in
                HStack {
                    Image(systemName: agent.enabled ? "checkmark.circle.fill" : "circle")
                        .foregroundStyle(agent.enabled ? .green : .secondary)
                    VStack(alignment: .leading, spacing: 2) {
                        Text(agent.name)
                            .font(.body.weight(.medium))
                        Text(agent.pushDir)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                            .lineLimit(1)
                    }
                    if agent.custom {
                        Text("Custom")
                            .font(.caption2)
                            .padding(.horizontal, 6)
                            .padding(.vertical, 2)
                            .background(.quaternary, in: Capsule())
                    }
                }
                .tag(agent)
            }
        }
        .frame(width: 240)
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

    private func load() async {
        isLoading = true
        errorMessage = nil
        do {
            let agentsList: [AgentConfig] = try await client.invoke("agents.list")
            agents = agentsList
            selectedAgent = agents.first
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }
}

// MARK: - Agent Detail

private struct AgentDetail: View {
    let agent: AgentConfig
    let client: DaemonClient

    @State private var selectedTab: AgentTab = .skills
    @State private var pushedSkills: [AgentSkillEntry] = []
    @State private var scanResults: [AgentSkillCandidate] = []
    @State private var memoryPreview: AgentMemoryPreview?
    @State private var isLoadingSkills: Bool = false
    @State private var isScanning: Bool = false
    @State private var isLoadingMemory: Bool = false
    @State private var errorMessage: String?
    @State private var statusMessage: String?

    enum AgentTab: String, CaseIterable {
        case skills = "Skills"
        case memory = "Memory"
    }

    var body: some View {
        VStack(spacing: 0) {
            agentHeader
            Divider()
            Picker("Tab", selection: $selectedTab) {
                ForEach(AgentTab.allCases, id: \.self) { tab in
                    Text(tab.rawValue).tag(tab)
                }
            }
            .pickerStyle(.segmented)
            .padding(12)

            switch selectedTab {
            case .skills:
                skillsTab
            case .memory:
                memoryTab
            }
        }
        .task(id: agent.name) {
            await loadSkills()
        }
    }

    private var agentHeader: some View {
        HStack(spacing: 12) {
            Image(systemName: "person.crop.circle")
                .font(.title)
                .foregroundStyle(.secondary)
            VStack(alignment: .leading, spacing: 4) {
                Text(agent.name)
                    .font(.title3.weight(.semibold))
                Text("Push: \(agent.pushDir)")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                if !agent.scanDirs.isEmpty {
                    Text("Scan: \(agent.scanDirs.joined(separator: ", "))")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }
            }
            Spacer()
        }
        .padding(16)
    }

    private var skillsTab: some View {
        VStack(spacing: 0) {
            HStack(spacing: 12) {
                Button(action: { Task { await scanSkills() } }) {
                    Label("Scan", systemImage: "magnifyingglass")
                }
                .disabled(isScanning)
                if isScanning {
                    ProgressView().controlSize(.small)
                }
                Spacer()
                if !pushedSkills.isEmpty {
                    Text("\(pushedSkills.count) pushed skill\(pushedSkills.count == 1 ? "" : "s")")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }
            .padding(12)

            Divider()

            if isLoadingSkills {
                ProgressView("Loading skills...")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else if pushedSkills.isEmpty && scanResults.isEmpty {
                VStack(spacing: 8) {
                    Image(systemName: "tray")
                        .font(.system(size: 28))
                        .foregroundStyle(.tertiary)
                    Text("No skills found")
                        .font(.headline)
                    if !agent.scanDirs.isEmpty {
                        Button("Scan for Skills") {
                            Task { await scanSkills() }
                        }
                    }
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                List {
                    if !pushedSkills.isEmpty {
                        Section("Pushed Skills") {
                            ForEach(pushedSkills) { skill in
                                skillRow(skill)
                            }
                        }
                    }
                    if !scanResults.isEmpty {
                        Section("Scan Results") {
                            ForEach(scanResults) { candidate in
                                HStack {
                                    Image(systemName: candidate.installed ? "checkmark.circle.fill" : "circle")
                                        .foregroundStyle(candidate.installed ? .green : .secondary)
                                    VStack(alignment: .leading, spacing: 2) {
                                        Text(candidate.name)
                                        Text(candidate.path)
                                            .font(.caption)
                                            .foregroundStyle(.secondary)
                                            .lineLimit(1)
                                    }
                                    Spacer()
                                    if !candidate.installed {
                                        Button("Pull") {
                                            Task { await pullSkill(candidate) }
                                        }
                                        .buttonStyle(.bordered)
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    private var memoryTab: some View {
        VStack(spacing: 0) {
            if isLoadingMemory {
                ProgressView("Loading memory...")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else if let preview = memoryPreview {
                ScrollView {
                    VStack(alignment: .leading, spacing: 16) {
                        if preview.mainExists {
                            VStack(alignment: .leading, spacing: 8) {
                                Text("Main Memory")
                                    .font(.headline)
                                Text(preview.mainContent.isEmpty ? "(empty)" : preview.mainContent)
                                    .font(.body.monospaced())
                                    .padding(12)
                                    .background(Color(nsColor: .textBackgroundColor), in: RoundedRectangle(cornerRadius: 8))
                            }
                        }
                        if preview.rulesDirExists && !preview.rules.isEmpty {
                            VStack(alignment: .leading, spacing: 8) {
                                Text("Rules")
                                    .font(.headline)
                                ForEach(preview.rules) { rule in
                                    VStack(alignment: .leading, spacing: 4) {
                                        HStack {
                                            Text(rule.name)
                                                .font(.subheadline.weight(.medium))
                                            if rule.managed {
                                                Text("Managed")
                                                    .font(.caption2)
                                                    .padding(.horizontal, 6)
                                                    .padding(.vertical, 2)
                                                    .background(.blue.opacity(0.2), in: Capsule())
                                            }
                                        }
                                        Text(rule.content.isEmpty ? "(empty)" : rule.content)
                                            .font(.caption.monospaced())
                                            .foregroundStyle(.secondary)
                                            .lineLimit(5)
                                    }
                                    .padding(8)
                                    .background(Color(nsColor: .textBackgroundColor), in: RoundedRectangle(cornerRadius: 6))
                                }
                            }
                        }
                        if !preview.mainExists && !preview.rulesDirExists {
                            Text("No memory files found for this agent.")
                                .foregroundStyle(.secondary)
                        }
                    }
                    .padding(16)
                    .frame(maxWidth: 880, alignment: .leading)
                    .frame(maxWidth: .infinity, alignment: .center)
                }
            } else {
                VStack(spacing: 8) {
                    Image(systemName: "brain")
                        .font(.system(size: 28))
                        .foregroundStyle(.tertiary)
                    Text("No memory preview available")
                        .font(.headline)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
        }
        .task(id: selectedTab) {
            if selectedTab == .memory && memoryPreview == nil {
                await loadMemory()
            }
        }
    }

    private func skillRow(_ skill: AgentSkillEntry) -> some View {
        HStack {
            Image(systemName: "doc.text")
                .foregroundStyle(.secondary)
            VStack(alignment: .leading, spacing: 2) {
                Text(skill.name)
                Text(skill.path)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .lineLimit(1)
            }
            Spacer()
            if skill.updatable {
                Image(systemName: "arrow.up.circle.fill")
                    .foregroundStyle(.orange)
            }
            Button(action: { Task { await deleteSkill(skill) } }) {
                Image(systemName: "trash")
            }
            .buttonStyle(.borderless)
        }
    }

    private func loadSkills() async {
        isLoadingSkills = true
        errorMessage = nil
        do {
            let skills: [AgentSkillEntry] = try await client.invoke("agents.listSkills", parameters: AgentNameParams(agentName: agent.name))
            pushedSkills = skills
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoadingSkills = false
    }

    private func scanSkills() async {
        isScanning = true
        errorMessage = nil
        statusMessage = nil
        do {
            let results: [AgentSkillCandidate] = try await client.invoke("agents.scanSkills", parameters: AgentNameParams(agentName: agent.name))
            scanResults = results
            statusMessage = results.isEmpty ? "No new skills found." : "Found \(results.count) skill\(results.count == 1 ? "" : "s")."
        } catch {
            errorMessage = error.localizedDescription
        }
        isScanning = false
    }

    private func loadMemory() async {
        isLoadingMemory = true
        errorMessage = nil
        do {
            let preview: AgentMemoryPreview = try await client.invoke("agents.memoryPreview", parameters: AgentNameParams(agentName: agent.name))
            memoryPreview = preview
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoadingMemory = false
    }

    private func deleteSkill(_ skill: AgentSkillEntry) async {
        errorMessage = nil
        do {
            let _: NativeEmptyResult = try await client.invoke("agents.deleteSkill", parameters: AgentDeleteSkillParams(agentName: agent.name, skillPath: skill.path))
            await loadSkills()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func pullSkill(_ candidate: AgentSkillCandidate) async {
        errorMessage = nil
        let params = AgentPullParams(agentName: agent.name, skillPaths: [candidate.path], category: "Uncategorized")
        do {
            let _: NativeEmptyResult = try await client.invoke("agents.pull", parameters: params)
            statusMessage = "Pulled \(candidate.name)."
            await scanSkills()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

private struct AgentNameParams: Encodable {
    let agentName: String
}
