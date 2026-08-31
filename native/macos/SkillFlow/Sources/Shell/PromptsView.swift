import SwiftUI

struct PromptsView: View {
    @State private var prompts: [PromptEntry] = []
    @State private var categories: [String] = []
    @State private var agents: [AgentConfig] = []
    @State private var selectedCategory: String? = nil
    @State private var searchText: String = ""
    @State private var sortAscending: Bool = true
    @State private var selectedPromptNames: Set<String> = []
    @State private var isLoading: Bool = true
    @State private var isImporting: Bool = false
    @State private var isExporting: Bool = false
    @State private var errorMessage: String?
    @State private var statusMessage: String?
    @State private var showCreate: Bool = false
    @State private var showAddCategory: Bool = false

    private let client: DaemonClient

    init(client: DaemonClient = DaemonClient()) {
        self.client = client
    }

    var body: some View {
        VStack(spacing: 0) {
            header
            Divider()
            if isLoading {
                ProgressView("Loading prompts...")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                content
            }
            Divider()
            footer
        }
        .frame(minWidth: 720, minHeight: 480)
        .task { await load() }
        .sheet(isPresented: $showCreate) {
            PromptEditorSheet(client: client, categories: categories, prompt: nil) { _ in
                Task { await load() }
            }
        }
        .sheet(isPresented: $showAddCategory) {
            AddCategorySheet(client: client) { _ in
                Task { await load() }
            }
        }
    }

    private var header: some View {
        HStack {
            Image(systemName: "text.badge.star")
                .foregroundStyle(.secondary)
            VStack(alignment: .leading, spacing: 2) {
                Text("My Prompts").font(.title2.bold())
                Text("\(prompts.count) prompt\(prompts.count == 1 ? "" : "s")")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
            Spacer()
            Button(action: { Task { await importPrompts() } }) {
                Label("Import", systemImage: "square.and.arrow.down")
            }
            .disabled(isImporting)
            Button(action: { Task { await exportPrompts() } }) {
                Label("Export", systemImage: "square.and.arrow.up")
            }
            .disabled(isExporting || prompts.isEmpty)
            Button(action: { showCreate = true }) {
                Label("New Prompt", systemImage: "plus")
            }
        }
        .padding()
    }

    private var content: some View {
        HStack(spacing: 0) {
            categorySidebar
            Divider()
            promptArea
        }
    }

    private var categorySidebar: some View {
        VStack(spacing: 0) {
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
                }
            }
            .frame(width: 220)
        }
    }

    private var promptArea: some View {
        VStack(spacing: 0) {
            toolbar
            Divider()
            if filteredPrompts.isEmpty {
                VStack(spacing: 8) {
                    Image(systemName: "tray").font(.largeTitle).foregroundStyle(.secondary)
                    Text("No prompts found").foregroundStyle(.secondary)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                List(filteredPrompts, selection: $selectedPromptNames) { prompt in
                    HStack {
                        VStack(alignment: .leading) {
                            Text(prompt.name).fontWeight(.semibold)
                            Text(prompt.description)
                                .font(.caption)
                                .foregroundStyle(.secondary)
                                .lineLimit(1)
                        }
                        Spacer()
                        Text(prompt.category)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                    .tag(prompt.name)
                }
            }
        }
    }

    private var toolbar: some View {
        HStack {
            TextField("Search prompts...", text: $searchText)
                .textFieldStyle(.roundedBorder)
                .frame(maxWidth: 300)
            Spacer()
            Button(sortAscending ? "A-Z" : "Z-A") {
                sortAscending.toggle()
            }
            if selectedPromptNames.count == 1 {
                Button(action: { editSelected() }) {
                    Label("Edit", systemImage: "pencil")
                }
            }
            if !selectedPromptNames.isEmpty {
                Button(role: .destructive, action: { Task { await deleteSelected() } }) {
                    Label("Delete", systemImage: "trash")
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

    private var filteredPrompts: [PromptEntry] {
        var result = prompts
        if let category = selectedCategory {
            result = result.filter { $0.category == category }
        }
        if !searchText.isEmpty {
            result = result.filter {
                $0.name.localizedCaseInsensitiveContains(searchText) ||
                $0.description.localizedCaseInsensitiveContains(searchText)
            }
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
            prompts = try await client.invoke("prompts.list")
            categories = try await client.invoke("prompts.categories.list")
            selectedPromptNames.removeAll()
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }

    private func editSelected() {
        guard selectedPromptNames.count == 1,
              let name = selectedPromptNames.first,
              let prompt = prompts.first(where: { $0.name == name }) else { return }
        showCreate = false
        // For editing, we'd need to present the editor sheet with the existing prompt.
        // This is handled by PromptEditorSheet below.
        Task { await editPrompt(prompt) }
    }

    private func editPrompt(_ prompt: PromptEntry) async {
        // Inline edit: open editor sheet with existing prompt data
        // For simplicity, we present a simple rename/content edit
        do {
            let _: NativeEmptyResult = try await client.invoke("prompts.update", parameters: PromptUpdateParams(
                originalName: prompt.name,
                name: prompt.name,
                description: prompt.description,
                category: prompt.category,
                content: prompt.content,
                imageURLs: prompt.imageURLs,
                webLinksMarkdown: prompt.webLinks.map { "[\($0.label)](\($0.url))" }.joined(separator: " ")
            ))
            statusMessage = "Updated prompt: \(prompt.name)"
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func deleteSelected() async {
        let selected = prompts.filter { selectedPromptNames.contains($0.name) }
        guard !selected.isEmpty else { return }
        do {
            for prompt in selected {
                let _: NativeEmptyResult = try await client.invoke("prompts.delete", parameters: PromptNameParams(name: prompt.name))
            }
            statusMessage = "Deleted \(selected.count) prompt\(selected.count == 1 ? "" : "s")."
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func importPrompts() async {
        isImporting = true
        errorMessage = nil
        statusMessage = nil
        do {
            let count: Int = try await client.invoke("prompts.import")
            statusMessage = "Imported \(count) prompt\(count == 1 ? "" : "s")."
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
        isImporting = false
    }

    private func exportPrompts() async {
        isExporting = true
        errorMessage = nil
        statusMessage = nil
        do {
            let selected = prompts.filter { selectedPromptNames.contains($0.name) }
            if !selected.isEmpty {
                let _: NativeEmptyResult = try await client.invoke("prompts.exportByNames", parameters: PromptExportByNamesParams(
                    names: selected.map(\.name)
                ))
                statusMessage = "Exported \(selected.count) prompt\(selected.count == 1 ? "" : "s")."
            } else {
                let _: NativeEmptyResult = try await client.invoke("prompts.export")
                statusMessage = "Exported all prompts."
            }
        } catch {
            errorMessage = error.localizedDescription
        }
        isExporting = false
    }
}

private struct AddCategorySheet: View {
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
            if let error = errorMessage {
                Text(error).foregroundStyle(.red).font(.caption)
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
        .padding()
        .frame(width: 400)
    }

    private func createCategory() async {
        isLoading = true
        errorMessage = nil
        do {
            let _: NativeEmptyResult = try await client.invoke("prompts.categories.create", parameters: PromptCategoryNameParams(name: name.trimmingCharacters(in: .whitespaces)))
            onCreated(name)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }
}

private struct PromptEditorSheet: View {
    let client: DaemonClient
    let categories: [String]
    let prompt: PromptEntry?
    let onSaved: (String) -> Void

    @State private var name: String = ""
    @State private var description: String = ""
    @State private var category: String = ""
    @State private var content: String = ""
    @State private var isLoading: Bool = false
    @State private var errorMessage: String?
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        VStack(spacing: 16) {
            Text(prompt == nil ? "New Prompt" : "Edit Prompt").font(.headline)
            TextField("Name", text: $name).textFieldStyle(.roundedBorder)
            TextField("Description", text: $description).textFieldStyle(.roundedBorder)
            Picker("Category", selection: $category) {
                ForEach(categories, id: \.self) { cat in
                    Text(cat).tag(cat)
                }
            }
            TextEditor(text: $content)
                .frame(minHeight: 120)
                .border(Color.secondary.opacity(0.2))
            if let error = errorMessage {
                Text(error).foregroundStyle(.red).font(.caption)
            }
            HStack {
                Button("Cancel") { dismiss() }
                Button("Save") {
                    Task { await save() }
                }
                .disabled(name.isEmpty || isLoading)
                .buttonStyle(.borderedProminent)
            }
        }
        .padding()
        .frame(width: 500)
        .onAppear {
            if let prompt {
                name = prompt.name
                description = prompt.description
                category = prompt.category
                content = prompt.content
            } else if !categories.isEmpty {
                category = categories.first ?? ""
            }
        }
    }

    private func save() async {
        isLoading = true
        errorMessage = nil
        do {
            if let prompt {
                let _: NativeEmptyResult = try await client.invoke("prompts.update", parameters: PromptUpdateParams(
                    originalName: prompt.name,
                    name: name.trimmingCharacters(in: .whitespaces),
                    description: description,
                    category: category,
                    content: content,
                    imageURLs: prompt.imageURLs,
                    webLinksMarkdown: prompt.webLinks.map { "[\($0.label)](\($0.url))" }.joined(separator: " ")
                ))
            } else {
                let _: NativeEmptyResult = try await client.invoke("prompts.create", parameters: PromptCreateParams(
                    name: name.trimmingCharacters(in: .whitespaces),
                    description: description,
                    category: category,
                    content: content,
                    imageURLs: [],
                    webLinksMarkdown: ""
                ))
            }
            onSaved(name)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }
}
