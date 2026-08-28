import Foundation
import XCTest

#if canImport(SkillFlowCore)
@testable import SkillFlowCore
#else
@testable import SkillFlow
#endif

final class MockURLProtocol: URLProtocol {
    static var handler: ((URLRequest) throws -> (HTTPURLResponse, Data))?

    override class func canInit(with request: URLRequest) -> Bool { true }
    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }

    override func startLoading() {
        guard let handler = MockURLProtocol.handler else {
            client?.urlProtocol(self, didFailWithError: URLError(.unsupportedURL))
            return
        }
        do {
            let (response, data) = try handler(request)
            client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
            client?.urlProtocol(self, didLoad: data)
            client?.urlProtocolDidFinishLoading(self)
        } catch {
            client?.urlProtocol(self, didFailWithError: error)
        }
    }

    override func stopLoading() {}
}

final class DaemonClientTests: XCTestCase {
    override func tearDown() {
        MockURLProtocol.handler = nil
        super.tearDown()
    }

    func testInvokeSendsTokenVersionAndNativeEnvelope() async throws {
        let stateURL = try writeEndpoint(address: "127.0.0.1:49151", token: "test-token")
        let client = DaemonClient(endpointPath: stateURL.path, session: mockedSession())
        let captured = LockedBox<URLRequest?>(nil)

        MockURLProtocol.handler = { request in
            captured.value = request
            let body = """
            {"ok":true,"result":{"ok":true,"result":[{"name":"Codex Skill"}],"error":null}}
            """
            return (
                Self.httpResponse(status: 200),
                Data(body.utf8)
            )
        }

        let skills: [SkillFixture] = try await client.invoke(
            "skills.list",
            requestID: "request-001"
        )

        XCTAssertEqual(skills.map(\.name), ["Codex Skill"])
        let request = try XCTUnwrap(captured.value)
        XCTAssertEqual(request.url?.absoluteString, "http://127.0.0.1:49151/invoke")
        XCTAssertEqual(request.value(forHTTPHeaderField: "X-SkillFlow-Token"), "test-token")
        XCTAssertEqual(request.httpMethod, "POST")

        let body = try XCTUnwrap(request.bodyData)
        let envelope = try XCTUnwrap(
            JSONSerialization.jsonObject(with: body) as? [String: Any]
        )
        XCTAssertEqual(envelope["method"] as? String, "native.api")

        let parameters = try XCTUnwrap(envelope["params"] as? [String: Any])
        XCTAssertEqual(parameters["version"] as? String, NativeAPIVersion.current)
        XCTAssertEqual(parameters["method"] as? String, "skills.list")
        XCTAssertEqual(parameters["requestID"] as? String, "request-001")
        XCTAssertTrue(parameters["params"] is NSNull)
    }

    func testInvokeMapsMethodNotFoundError() async throws {
        let stateURL = try writeEndpoint(address: "127.0.0.1:49151", token: "test-token")
        let client = DaemonClient(endpointPath: stateURL.path, session: mockedSession())

        MockURLProtocol.handler = { _ in
            let body = """
            {"ok":true,"result":{"ok":false,"result":null,"error":{"code":"method_not_found","message":"method not found","messageKey":"nativeapi.error.method_not_found"}}}
            """
            return (
                Self.httpResponse(status: 200),
                Data(body.utf8)
            )
        }

        do {
            let _: [SkillFixture] = try await client.invoke("missing.method")
            XCTFail("Expected method_not_found to be thrown")
        } catch let error as DaemonClientError {
            XCTAssertEqual(
                error,
                .api(
                    code: "method_not_found",
                    messageKey: "nativeapi.error.method_not_found",
                    message: "method not found"
                )
            )
        }
    }

    func testInvokeMapsUnauthorizedResponse() async throws {
        let stateURL = try writeEndpoint(address: "127.0.0.1:49151", token: "expired-token")
        let client = DaemonClient(endpointPath: stateURL.path, session: mockedSession())

        MockURLProtocol.handler = { _ in
            (
                Self.httpResponse(status: 401),
                Data(#"{"ok":false,"error":"unauthorized"}"#.utf8)
            )
        }

        do {
            let _: [SkillFixture] = try await client.invoke("skills.list")
            XCTFail("Expected unauthorized to be thrown")
        } catch let error as DaemonClientError {
            XCTAssertEqual(error, .unauthorized)
        }
    }

    func testMissingEndpointFileMapsToEndpointUnavailable() async {
        let missingPath = NSTemporaryDirectory() + "/skillflow-missing-\(UUID().uuidString).json"
        let client = DaemonClient(endpointPath: missingPath, session: mockedSession())

        do {
            let _: [SkillFixture] = try await client.invoke("skills.list")
            XCTFail("Expected endpointUnavailable to be thrown")
        } catch let error as DaemonClientError {
            XCTAssertEqual(error, .endpointUnavailable)
        }
    }
}

private struct SkillFixture: Codable, Equatable {
    let name: String
}


private final class LockedBox<Value> {
    private let lock = NSLock()
    private var storage: Value

    init(_ value: Value) {
        self.storage = value
    }

    var value: Value {
        get {
            lock.lock()
            defer { lock.unlock() }
            return storage
        }
        set {
            lock.lock()
            defer { lock.unlock() }
            storage = newValue
        }
    }
}

private extension DaemonClientTests {
    func mockedSession() -> URLSession {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [MockURLProtocol.self]
        return URLSession(configuration: configuration)
    }

    func writeEndpoint(address: String, token: String) throws -> URL {
        let url = URL(fileURLWithPath: NSTemporaryDirectory())
            .appendingPathComponent("skillflow-daemon-\(UUID().uuidString).json")
        let endpoint = """
        {"address":"\(address)","token":"\(token)","pid":12345}
        """
        try Data(endpoint.utf8).write(to: url)
        return url
    }

    static func httpResponse(status: Int) -> HTTPURLResponse {
        HTTPURLResponse(
            url: URL(string: "http://127.0.0.1:49151/invoke")!,
            statusCode: status,
            httpVersion: "HTTP/1.1",
            headerFields: ["Content-Type": "application/json"]
        )!
    }
}

private extension URLRequest {
    var bodyData: Data? {
        if let body = httpBody {
            return body
        }
        guard let stream = httpBodyStream else {
            return nil
        }
        stream.open()
        defer { stream.close() }

        var data = Data()
        let bufferSize = 4096
        let buffer = UnsafeMutablePointer<UInt8>.allocate(capacity: bufferSize)
        defer { buffer.deallocate() }

        while stream.hasBytesAvailable {
            let read = stream.read(buffer, maxLength: bufferSize)
            if read <= 0 {
                break
            }
            data.append(buffer, count: read)
        }
        return data
    }
}
