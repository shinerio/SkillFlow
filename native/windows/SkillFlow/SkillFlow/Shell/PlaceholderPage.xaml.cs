using Microsoft.UI.Xaml.Controls;
using Microsoft.UI.Xaml.Navigation;

namespace SkillFlow.Shell;

public sealed partial class PlaceholderPage : Page
{
    public PlaceholderPage()
    {
        InitializeComponent();
    }

    protected override void OnNavigatedTo(NavigationEventArgs e)
    {
        base.OnNavigatedTo(e);
        TitleText.Text = e.Parameter as string ?? "SkillFlow";
    }
}
