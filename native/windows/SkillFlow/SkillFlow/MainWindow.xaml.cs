using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using SkillFlow.Agents;
using SkillFlow.Backup;
using SkillFlow.Memory;
using SkillFlow.Prompts;
using SkillFlow.Settings;
using SkillFlow.Shell;
using SkillFlow.Skills;
using SkillFlow.StarredRepos;

namespace SkillFlow;

public sealed partial class MainWindow : Window
{
    public MainWindow()
    {
        InitializeComponent();
        Title = "SkillFlow";
        ContentFrame.Navigate(typeof(SettingsPage));
        Navigation.SelectedItem = Navigation.MenuItems.First();
    }

    private void OnNavigationSelectionChanged(NavigationView sender, NavigationViewSelectionChangedEventArgs args)
    {
        if (args.SelectedItem is not NavigationViewItem item || item.Tag is not string tag)
        {
            return;
        }

        switch (tag)
        {
            case "settings":
                ContentFrame.Navigate(typeof(SettingsPage));
                break;
            case "skills":
                ContentFrame.Navigate(typeof(SkillsPage));
                break;
            case "agents":
                ContentFrame.Navigate(typeof(AgentsPage));
                break;
            case "starredRepos":
                ContentFrame.Navigate(typeof(StarredReposPage));
                break;
            case "prompts":
                ContentFrame.Navigate(typeof(PromptsPage));
                break;
            case "memory":
                ContentFrame.Navigate(typeof(MemoryPage));
                break;
            case "backup":
                ContentFrame.Navigate(typeof(BackupPage));
                break;
            default:
                ContentFrame.Navigate(typeof(SettingsPage));
                break;
        }
    }
}
