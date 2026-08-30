// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "SkillFlow",
    platforms: [
        .macOS(.v13)
    ],
    targets: [
        .executableTarget(
            name: "SkillFlow",
            path: "Sources"
        ),
        .testTarget(
            name: "SkillFlowTests",
            dependencies: ["SkillFlow"],
            path: "Tests"
        )
    ]
)
