import AppKit
import Foundation

private struct DaemonEndpointRecord: Decodable {
    let address: String
    let token: String
    let pid: Int
}

final class DaemonLauncher {
    static let shared = DaemonLauncher()

    private let endpointPath: String
    private let queue = DispatchQueue(label: "com.shinerio.skillflow.daemon-launcher")
    private var daemonProcess: Process?

    private init(endpointPath: String = DaemonClient.defaultEndpointPath()) {
        self.endpointPath = endpointPath
    }

    func ensureRunning() {
        queue.sync {
            if endpointIsLive {
                return
            }
            try? FileManager.default.removeItem(atPath: endpointPath)
            launchDaemon()
            waitForEndpoint()
        }
    }

    func stopIfOwned() {
        queue.sync {
            guard let process = daemonProcess, process.isRunning else {
                daemonProcess = nil
                return
            }
            process.terminate()
            let deadline = Date().addingTimeInterval(3)
            while process.isRunning && Date() < deadline {
                RunLoop.current.run(until: Date().addingTimeInterval(0.05))
            }
            if process.isRunning {
                process.interrupt()
            }
            daemonProcess = nil
        }
    }

    private var endpointIsLive: Bool {
        guard let data = try? Data(contentsOf: URL(fileURLWithPath: endpointPath)),
              let endpoint = try? JSONDecoder().decode(DaemonEndpointRecord.self, from: data)
        else {
            return false
        }
        return kill(pid_t(endpoint.pid), 0) == 0
    }

    private func launchDaemon() {
        let bundleExecutableURL = Bundle.main.bundleURL
            .appendingPathComponent("Contents", isDirectory: true)
            .appendingPathComponent("MacOS", isDirectory: true)
            .appendingPathComponent("skillflowd", isDirectory: false)
        let executableURL: URL
        if FileManager.default.fileExists(atPath: bundleExecutableURL.path) {
            executableURL = bundleExecutableURL
        } else {
            executableURL = Bundle.main.bundleURL
                .appendingPathComponent("skillflowd", isDirectory: false)
        }

        guard FileManager.default.isExecutableFile(atPath: executableURL.path) else {
            return
        }

        let process = Process()
        process.executableURL = executableURL
        process.arguments = ["--daemon-only"]
        process.standardOutput = FileHandle.nullDevice
        process.standardError = FileHandle.nullDevice
        do {
            try process.run()
            daemonProcess = process
        } catch {
            daemonProcess = nil
        }
    }

    private func waitForEndpoint() {
        let deadline = Date().addingTimeInterval(8)
        while Date() < deadline {
            if endpointIsLive {
                return
            }
            Thread.sleep(forTimeInterval: 0.1)
        }
    }
}
