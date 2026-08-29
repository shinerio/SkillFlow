import SwiftUI

struct RootView: View {
    @State
    private var selection: SidebarItem? = .settings

    var body: some View {
        NavigationSplitView(
            sidebar: {
                SidebarView(selection: $selection)
            },
            detail: {
                if let selection {
                    switch selection {
                    case .settings:
                        SettingsView()
                    case .skills:
                        SkillsView()
                    case .agents:
                        AgentsView()
                    case .memory:
                        MemoryView()
                    case .backup:
                        BackupView()
                    default:
                        PlaceholderView(item: selection)
                    }
                } else {
                    VStack(spacing: 8) {
                        Image(systemName: "sidebar.left")
                            .font(.system(size: 28))
                            .foregroundStyle(.secondary)
                        Text("Select a section")
                            .font(.headline)
                        Text("Choose a SkillFlow section to continue.")
                            .font(.body)
                            .foregroundStyle(.secondary)
                    }
                }
            }
        )
        .navigationTitle("SkillFlow")
        .frame(minWidth: 880, minHeight: 560)
    }
}

private struct PlaceholderView: View {
    let item: SidebarItem

    var body: some View {
        VStack(spacing: 12) {
            Image(systemName: item.systemImage)
                .font(.system(size: 42))
                .foregroundStyle(.secondary)
            Text(item.title)
                .font(.title2.weight(.semibold))
            Text("This section will be migrated to the native client in an upcoming batch.")
                .font(.body)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)
                .padding(.horizontal, 32)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}
