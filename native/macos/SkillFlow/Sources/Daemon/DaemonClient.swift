import Foundation

struct DaemonClient {
    private let endpointPath: String
    private let session: URLSession

    init(
        endpointPath: String = DaemonClient.defaultEndpointPath(),
        session: URLSession = .shared
    ) {
        self.endpointPath = endpointPath
        self.session = session
    }

    static func defaultEndpointPath() -> String {
        let applicationSupport = URL(
            fileURLWithPath: NSHomeDirectory(),
            isDirectory: true
        )
        .appendingPathComponent("Library/Application Support", isDirectory: true)
        .appendingPathComponent("SkillFlow", isDirectory: true)
        .appendingPathComponent("runtime", isDirectory: true)
        .appendingPathComponent("daemon-service.json")

        return applicationSupport.path
    }

    func invoke<Result: Decodable, Parameters: Encodable>(
        _ method: String,
        parameters: Parameters? = nil,
        requestID: String = UUID().uuidString
    ) async throws -> Result {
        let endpoint = try loadEndpoint()
        guard let url = URL(string: "http://\(endpoint.address)/invoke") else {
            throw DaemonClientError.invalidEndpoint
        }

        let nativeRequest = NativeAPIRequest(
            version: NativeAPIVersion.current,
            method: method,
            parameters: parameters,
            requestID: requestID
        )
        let serviceRequest = DaemonServiceRequest(
            method: "native.api",
            parameters: nativeRequest
        )

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "POST"
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        urlRequest.setValue(endpoint.token, forHTTPHeaderField: "X-SkillFlow-Token")
        urlRequest.httpBody = try JSONEncoder().encode(serviceRequest)

        let data: Data
        let response: URLResponse
        do {
            (data, response) = try await session.data(for: urlRequest)
        } catch {
            throw DaemonClientError.transport(error.localizedDescription)
        }

        guard let httpResponse = response as? HTTPURLResponse else {
            throw DaemonClientError.invalidResponse
        }
        if httpResponse.statusCode == HTTPStatus.unauthorized {
            throw DaemonClientError.unauthorized
        }
        guard (200..<300).contains(httpResponse.statusCode) else {
            throw DaemonClientError.transport("HTTP \(httpResponse.statusCode)")
        }

        let decoded: DaemonServiceResponse<NativeAPIResponse<Result>>
        do {
            decoded = try JSONDecoder().decode(
                DaemonServiceResponse<NativeAPIResponse<Result>>.self,
                from: data
            )
        } catch {
            throw DaemonClientError.invalidResponse
        }

        guard decoded.ok else {
            throw DaemonClientError.service(decoded.error ?? "unknown service error")
        }

        guard let nativeResponse = decoded.result else {
            throw DaemonClientError.invalidResponse
        }
        guard nativeResponse.ok else {
            let apiError = nativeResponse.error
            throw DaemonClientError.api(
                code: apiError?.code ?? "internal_error",
                messageKey: apiError?.messageKey,
                message: apiError?.message
            )
        }
        guard let result = nativeResponse.result else {
            throw DaemonClientError.missingResult
        }
        return result
    }

    func invoke<Result: Decodable>(
        _ method: String,
        requestID: String = UUID().uuidString
    ) async throws -> Result {
        try await invoke(method, parameters: Optional<NoParameters>.none, requestID: requestID)
    }

    private func loadEndpoint() throws -> DaemonEndpoint {
        let endpointURL = URL(fileURLWithPath: endpointPath)
        guard FileManager.default.fileExists(atPath: endpointURL.path) else {
            throw DaemonClientError.endpointUnavailable
        }

        do {
            let endpoint = try JSONDecoder().decode(
                DaemonEndpoint.self,
                from: Data(contentsOf: endpointURL)
            )
            guard !endpoint.address.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
                  !endpoint.token.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
                throw DaemonClientError.invalidEndpoint
            }
            return endpoint
        } catch let error as DaemonClientError {
            throw error
        } catch {
            throw DaemonClientError.invalidEndpoint
        }
    }
}

private enum HTTPStatus {
    static let unauthorized = 401
}

private struct NoParameters: Encodable {}
