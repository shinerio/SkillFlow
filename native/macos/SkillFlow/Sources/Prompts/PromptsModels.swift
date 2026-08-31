import Foundation

struct PromptEntry: Codable, Identifiable, Equatable {
    let name: String
    let description: String
    let category: String
    let path: String
    let filePath: String
    let content: String
    let imageURLs: [String]
    let webLinks: [PromptLink]
    let createdAt: Date
    let updatedAt: Date

    var id: String { name }
}

struct PromptLink: Codable, Equatable {
    let label: String
    let url: String
}

struct PromptNameParams: Encodable {
    let name: String
}

struct PromptMoveCategoryParams: Encodable {
    let name: String
    let category: String
}

struct PromptCategoryNameParams: Encodable {
    let name: String
}

struct PromptRenameCategoryParams: Encodable {
    let oldName: String
    let newName: String
}

struct PromptCreateParams: Encodable {
    let name: String
    let description: String
    let category: String
    let content: String
    let imageURLs: [String]
    let webLinksMarkdown: String
}

struct PromptUpdateParams: Encodable {
    let originalName: String
    let name: String
    let description: String
    let category: String
    let content: String
    let imageURLs: [String]
    let webLinksMarkdown: String
}

struct PromptExportByNamesParams: Encodable {
    let names: [String]
}
