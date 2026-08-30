using System.Diagnostics;
using System.Text.Json;

namespace SkillFlow.Daemon;

public sealed class DaemonLauncher
{
    public static DaemonLauncher Shared { get; } = new();

    private readonly string _endpointPath = DaemonClient.DefaultEndpointPath();
    private Process? _daemonProcess;
    private readonly object _lock = new();

    private DaemonLauncher()
    {
        AppDomain.CurrentDomain.ProcessExit += (_, _) => StopIfOwned();
    }

    public void EnsureRunning()
    {
        lock (_lock)
        {
            if (EndpointIsLive())
            {
                return;
            }

            try
            {
                File.Delete(_endpointPath);
            }
            catch (DirectoryNotFoundException)
            {
            }
            catch (FileNotFoundException)
            {
            }

            var daemonPath = Path.Combine(AppContext.BaseDirectory, "skillflowd.exe");
            if (!File.Exists(daemonPath))
            {
                return;
            }

            try
            {
                var process = new Process
                {
                    StartInfo = new ProcessStartInfo
                    {
                        FileName = daemonPath,
                        Arguments = "--daemon-only",
                        UseShellExecute = false,
                        CreateNoWindow = true,
                    },
                    EnableRaisingEvents = true,
                };
                if (!process.Start())
                {
                    return;
                }
                _daemonProcess = process;
            }
            catch (Exception)
            {
                _daemonProcess = null;
                return;
            }

            WaitForEndpoint();
        }
    }

    public void StopIfOwned()
    {
        lock (_lock)
        {
            var process = _daemonProcess;
            _daemonProcess = null;
            if (process is null)
            {
                return;
            }
            try
            {
                if (process.HasExited)
                {
                    return;
                }
            }
            catch (InvalidOperationException)
            {
                return;
            }

            try
            {
                process.Kill(entireProcessTree: true);
            }
            catch (InvalidOperationException)
            {
            }
            catch (Win32Exception)
            {
            }
        }
    }

    private bool EndpointIsLive()
    {
        try
        {
            using var document = JsonDocument.Parse(File.ReadAllText(_endpointPath));
            var root = document.RootElement;
            var pid = root.TryGetProperty("pid", out var pidValue) && pidValue.TryGetInt32(out var value)
                ? value
                : 0;
            if (pid <= 0)
            {
                return false;
            }

            _ = Process.GetProcessById(pid);
            return true;
        }
        catch (Exception)
        {
            return false;
        }
    }

    private void WaitForEndpoint()
    {
        var deadline = DateTime.UtcNow.AddSeconds(8);
        while (DateTime.UtcNow < deadline)
        {
            if (EndpointIsLive())
            {
                return;
            }
            Thread.Sleep(100);
        }
    }
}
