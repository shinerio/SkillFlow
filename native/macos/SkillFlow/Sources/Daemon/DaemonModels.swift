import Foundation

enum NativeAPIVersion {
    static let current = "2026-04-25"
}

struct DaemonEndpoint: Codable, Equatable {
    let address: String
    let token: String
    let pid: Int
}

struct NativeAPIRequest<Parameters: Encodable>: Encodable {
    let version: String
    let method: String
    let parameters: Parameters?
    let requestID: String

    enum CodingKeys: String, CodingKey {
        case version
        case method
        case parameters = "params"
        case requestID
    }

    func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        try container.encode(version, forKey: .version)
        try container.encode(method, forKey: .method)
        if let parameters {
            try container.encode(parameters, forKey: .parameters)
        } else {
            try container.encodeNil(forKey: .parameters)
        }
        try container.encode(requestID, forKey: .requestID)
    }
}

struct NativeAPIError: Codable, Equatable {
    let code: String
    let message: String?
    let messageKey: String?
}

struct NativeAPIResponse<Result: Decodable>: Decodable {
    let ok: Bool
    let result: Result?
    let error: NativeAPIError?
}

struct DaemonServiceRequest<Parameters: Encodable>: Encodable {
    let method: String
    let parameters: NativeAPIRequest<Parameters>

    enum CodingKeys: String, CodingKey {
        case method
        case parameters = "params"
    }
}

struct DaemonServiceResponse<Result: Decodable>: Decodable {
    let ok: Bool
    let result: Result?
    let error: String?
}

public enum DaemonClientError: Error, Equatable {
    case endpointUnavailable
    case invalidEndpoint
    case invalidResponse
    case transport(String)
    case unauthorized
    case service(String)
    case api(code: String, messageKey: String?, message: String?)
    case missingResult

    var localizedDescription: String {
        switch self {
        case .endpointUnavailable:
            return "SkillFlow daemon endpoint is unavailable."
        case .invalidEndpoint:
            return "SkillFlow daemon endpoint is invalid."
        case .invalidResponse:
            return "SkillFlow daemon returned an invalid response."
        case .transport(let message):
            return "SkillFlow daemon request failed: \(message)"
        case .unauthorized:
            return "SkillFlow daemon rejected the access token."
        case .service(let message):
            return "SkillFlow daemon service failed: \(message)"
        case .api(let code, _, let message):
            let detail = message ?? "The request could not be completed."
            return "SkillFlow daemon API error (\(code)): \(detail)"
        case .missingResult:
            return "SkillFlow daemon response did not include a result."
        }
    }
}
