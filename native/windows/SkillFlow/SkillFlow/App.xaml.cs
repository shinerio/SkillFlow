using Microsoft.UI.Xaml;
using SkillFlow.Daemon;

namespace SkillFlow;

public sealed partial class App : Application
{
    public static MainWindow? MainWindow { get; private set; }

    public App()
    {
        InitializeComponent();
    }

    protected override void OnLaunched(LaunchActivatedEventArgs args)
    {
        DaemonLauncher.Shared.EnsureRunning();
        MainWindow = new MainWindow();
        MainWindow.Activate();
    }
}
