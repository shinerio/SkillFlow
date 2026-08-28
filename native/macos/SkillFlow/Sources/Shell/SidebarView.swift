import SwiftUI

enum SidebarItem: String, Hashable, CaseIterable, Identifiable {
    case settings
    case skills
    case agents
    case starredRepos
    case prompts
    case memory
    case backup

    var id: String { rawValue }

    var title: String {
        switch self {
        case .settings: "Settings"
        case .skills: "My Skills"
        case .agents: "My Agents"
        case .starredRepos: "Starred Repos"
        case .prompts: "My Prompts"
        case .memory: "My Memory"
        case .backup: "Cloud Backup"
        }
    }

    var systemImage: String {
        switch self {
        case .settings: "gearshape"
        case .skills: "square.grid.2x2"
        case .agents: "person.2"
        case .starredRepos: "star"
        case .prompts: "text.badge.star"
        case .memory: "brain"
        case .backup: "externaldrive.badge.icloud"
        }
    }
}

struct SidebarView: View {
    @Binding
    var selection: SidebarItem?

    var body: some View {
        List(selection: $selection) {
            ForEach(SidebarItem.allCases) { item in
                Label(item.title, systemImage: item.systemImage)
                    .tag(item)
            }
        }
        .listStyle(.sidebar)
        .frame(minWidth: 210)
    }
}
