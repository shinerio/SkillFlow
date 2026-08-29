using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using SkillFlow.Agents;
using SkillFlow.Settings;
using SkillFlow.Shell;
using SkillFlow.Skills;

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
            default:
                var title = tag switch
                {
                    "starredRepos" => "Starred Repos",
                    "prompts" => "My Prompts",
                    "memory" => "My Memory",
                    "backup" => "Cloud Backup",
                    _ => "Settings",
                };
                ContentFrame.Navigate(typeof(PlaceholderPage), title);
                break;
        }
    }
}
