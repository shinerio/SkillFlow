// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "SkillFlow",
    platforms: [
        .macOS(.v13)
    ],
    targets: [
        .target(
            name: "SkillFlowCore",
            path: "Sources",
            exclude: ["AppDelegate.swift", "SkillFlowApp.swift", "Shell"],
            sources: ["Daemon"]
        ),
        .executableTarget(
            name: "SkillFlow",
            dependencies: ["SkillFlowCore"],
            path: "Sources",
            exclude: ["Daemon"],
            sources: ["AppDelegate.swift", "SkillFlowApp.swift", "Shell"]
        ),
        .testTarget(
            name: "SkillFlowTests",
            dependencies: ["SkillFlowCore"],
            path: "Tests"
        )
    ]
)
