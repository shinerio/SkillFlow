#if canImport(SkillFlowCore)
import SkillFlowCore
#endif
import SwiftUI

@main
struct SkillFlowApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self)
    private var appDelegate

    var body: some Scene {
        WindowGroup {
            RootView()
        }
        .windowToolbarStyle(.unified)
    }
}
