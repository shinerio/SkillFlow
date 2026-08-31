using System.Collections.ObjectModel;
using System.ComponentModel;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using SkillFlow.Daemon;
using SkillFlow.Skills;

namespace SkillFlow.Agents;

public sealed partial class AgentsPage : Page, INotifyPropertyChanged
{
    private readonly DaemonClient _client;
    private List<AgentInfo> _agents = new();
    private AgentInfo? _selectedAgent;
    private List<AgentSkillEntry> _pushedSkills = new();
    private List<AgentSkillCandidate> _scanResults = new();
    private AgentMemoryPreview? _memoryPreview;
    private bool _isLoading = true;
    private bool _isScanning;
    private bool _isLoadingMemory;
    private string? _errorMessage;
    private string? _statusMessage;
    private string _activeTab = "skills";

    public AgentsPage() : this(new DaemonClient()) { }

    public AgentsPage(DaemonClient client)
    {
        _client = client;
        InitializeComponent();
        Loaded += async (_, _) => await LoadAsync();
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    public bool IsLoading
    {
        get => _isLoading;
        set => SetProperty(ref _isLoading, value);
    }

    public bool IsScanning
    {
        get => _isScanning;
        set => SetProperty(ref _isScanning, value);
    }

    public string? ErrorMessage
    {
        get => _errorMessage;
        set => SetProperty(ref _errorMessage, value);
    }

    public string? StatusMessage
    {
        get => _statusMessage;
        set => SetProperty(ref _statusMessage, value);
    }

    private async Task LoadAsync()
    {
        IsLoading = true;
        ErrorMessage = null;
        try
        {
            _agents = await _client.InvokeAsync<List<AgentInfo>>("agents.list");
            AgentListView.ItemsSource = _agents;
            AgentCountText.Text = $"{_agents.Count} agent{(_agents.Count == 1 ? "" : "s")}";
            if (_agents.Count > 0)
            {
                AgentListView.SelectedIndex = 0;
            }
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsLoading = false;
    }

    private async void OnAgentSelectionChanged(object sender, SelectionChangedEventArgs e)
    {
        if (AgentListView.SelectedItem is not AgentInfo agent) return;
        _selectedAgent = agent;
        AgentNameText.Text = agent.Name;
        AgentPushDirText.Text = $"Push: {agent.PushDir}";
        AgentScanDirsText.Text = agent.ScanDirs.Count > 0
            ? $"Scan: {string.Join(", ", agent.ScanDirs)}"
            : "";
        _pushedSkills.Clear();
        _scanResults.Clear();
        _memoryPreview = null;
        await LoadSkillsAsync(agent);
    }

    private async Task LoadSkillsAsync(AgentInfo agent)
    {
        try
        {
            _pushedSkills = await _client.InvokeAsync<List<AgentSkillEntry>>("agents.listSkills",
                new AgentNameParams { AgentName = agent.Name });
            RenderSkillsList();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private void RenderSkillsList()
    {
        var items = new ObservableCollection<object>();

        if (_pushedSkills.Count > 0)
        {
            foreach (var skill in _pushedSkills)
            {
                items.Add(new SkillListItem { Group = "Pushed Skills", Name = skill.Name, Path = skill.Path, CanDelete = true });
            }
        }

        if (_scanResults.Count > 0)
        {
            foreach (var candidate in _scanResults)
            {
                items.Add(new SkillListItem
                {
                    Group = "Scan Results",
                    Name = candidate.Name,
                    Path = candidate.Path,
                    CanPull = !candidate.Installed
                });
            }
        }

        var grouped = items.Cast<SkillListItem>()
            .GroupBy(i => i.Group)
            .Select(g => new SkillListGroup(g.Key, g.ToList()))
            .ToList();

        SkillsListView.ItemsSource = grouped;
    }

    private async Task ScanAsync()
    {
        if (_selectedAgent is null) return;
        IsScanning = true;
        ScanProgress.IsActive = true;
        ErrorMessage = null;
        try
        {
            _scanResults = await _client.InvokeAsync<List<AgentSkillCandidate>>("agents.scanSkills",
                new AgentNameParams { AgentName = _selectedAgent.Name });
            StatusMessage = _scanResults.Count == 0 ? "No new skills found." : $"Found {_scanResults.Count} skill{(_scanResults.Count == 1 ? "" : "s")}.";
            RenderSkillsList();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsScanning = false;
        ScanProgress.IsActive = false;
    }

    private async Task LoadMemoryAsync()
    {
        if (_selectedAgent is null) return;
        _isLoadingMemory = true;
        ErrorMessage = null;
        try
        {
            _memoryPreview = await _client.InvokeAsync<AgentMemoryPreview>("agents.memoryPreview",
                new AgentNameParams { AgentName = _selectedAgent.Name });
            RenderMemory();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        _isLoadingMemory = false;
    }

    private void RenderMemory()
    {
        MemoryPanel.Children.Clear();

        if (_memoryPreview is null)
        {
            MemoryPanel.Children.Add(new TextBlock
            {
                Text = "No memory preview available.",
                Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
            });
            return;
        }

        if (_memoryPreview.MainExists)
        {
            var mainPanel = new StackPanel { Spacing = 8 };
            mainPanel.Children.Add(new TextBlock { Text = "Main Memory", FontWeight = Microsoft.UI.Text.FontWeights.SemiBold });
            var content = _memoryPreview.MainContent;
            if (string.IsNullOrEmpty(content)) content = "(empty)";
            mainPanel.Children.Add(new TextBlock
            {
                Text = content,
                FontFamily = new Microsoft.UI.Xaml.Media.FontFamily("Consolas"),
                TextWrapping = TextWrapping.Wrap
            });
            MemoryPanel.Children.Add(mainPanel);
        }

        if (_memoryPreview.RulesDirExists && _memoryPreview.Rules.Count > 0)
        {
            var rulesPanel = new StackPanel { Spacing = 8 };
            rulesPanel.Children.Add(new TextBlock { Text = "Rules", FontWeight = Microsoft.UI.Text.FontWeights.SemiBold });
            foreach (var rule in _memoryPreview.Rules)
            {
                var rulePanel = new StackPanel { Spacing = 4, Padding = new Thickness(8) };
                var headerPanel = new StackPanel { Orientation = Orientation.Horizontal, Spacing = 8 };
                headerPanel.Children.Add(new TextBlock { Text = rule.Name, FontWeight = Microsoft.UI.Text.FontWeights.Medium });
                if (rule.Managed)
                {
                    headerPanel.Children.Add(new TextBlock
                    {
                        Text = "Managed",
                        FontSize = 11,
                        Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["AccentFillColorDefaultBrush"]
                    });
                }
                rulePanel.Children.Add(headerPanel);
                var ruleContent = string.IsNullOrEmpty(rule.Content) ? "(empty)" : rule.Content;
                rulePanel.Children.Add(new TextBlock
                {
                    Text = ruleContent,
                    FontSize = 12,
                    FontFamily = new Microsoft.UI.Xaml.Media.FontFamily("Consolas"),
                    Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"],
                    TextWrapping = TextWrapping.Wrap,
                    MaxLines = 5
                });
                rulesPanel.Children.Add(rulePanel);
            }
            MemoryPanel.Children.Add(rulesPanel);
        }

        if (!_memoryPreview.MainExists && !_memoryPreview.RulesDirExists)
        {
            MemoryPanel.Children.Add(new TextBlock
            {
                Text = "No memory files found for this agent.",
                Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
            });
        }
    }

    private void OnTabSelectionChanged(NavigationView sender, NavigationViewSelectionChangedEventArgs args)
    {
        if (args.SelectedItem is not NavigationViewItem item || item.Tag is not string tag) return;
        _activeTab = tag;

        if (tag == "skills")
        {
            SkillsTab.Visibility = Visibility.Visible;
            MemoryTab.Visibility = Visibility.Collapsed;
        }
        else if (tag == "memory")
        {
            SkillsTab.Visibility = Visibility.Collapsed;
            MemoryTab.Visibility = Visibility.Visible;
            if (_memoryPreview is null && _selectedAgent is not null)
            {
                _ = LoadMemoryAsync();
            }
        }
    }

    private async void OnScan(object sender, RoutedEventArgs e) => await ScanAsync();
    private async void OnReload(object sender, RoutedEventArgs e) => await LoadAsync();

    private async void OnPullSkill(object sender, RoutedEventArgs e)
    {
        if (_selectedAgent is null) return;
        if (sender is not Button btn || btn.Tag is not string skillPath) return;
        ErrorMessage = null;
        try
        {
            await _client.InvokeAsync<NativeEmptyResult>("agents.pull",
                new AgentPullParams
                {
                    AgentName = _selectedAgent.Name,
                    SkillPaths = new List<string> { skillPath },
                    Category = "Uncategorized"
                });
            StatusMessage = "Skill pulled successfully.";
            await ScanAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private async void OnDeleteSkill(object sender, RoutedEventArgs e)
    {
        if (_selectedAgent is null) return;
        if (sender is not Button btn || btn.Tag is not string skillPath) return;

        var dialog = new ContentDialog
        {
            XamlRoot = this.XamlRoot,
            Title = "Delete Skill",
            Content = "Delete this skill from the agent's push directory? This action cannot be undone.",
            PrimaryButtonText = "Delete",
            CloseButtonText = "Cancel",
            DefaultButton = ContentDialogButton.Close
        };
        if (await dialog.ShowAsync() != ContentDialogResult.Primary) return;

        ErrorMessage = null;
        try
        {
            await _client.InvokeAsync<NativeEmptyResult>("agents.deleteSkill",
                new AgentDeleteSkillParams
                {
                    AgentName = _selectedAgent.Name,
                    SkillPath = skillPath
                });
            StatusMessage = "Skill deleted.";
            if (_selectedAgent is not null)
                await LoadSkillsAsync(_selectedAgent);
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private bool SetProperty<T>(ref T field, T value, [CallerMemberName] string? name = null)
    {
        if (EqualityComparer<T>.Default.Equals(field, value))
            return false;
        field = value;
        OnPropertyChanged(name);
        return true;
    }

    private void OnPropertyChanged([CallerMemberName] string? name = null)
    {
        PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(name));
    }
}

public sealed class SkillListItem
{
    public string Group { get; set; } = string.Empty;
    public string Name { get; set; } = string.Empty;
    public string Path { get; set; } = string.Empty;
    public bool CanDelete { get; set; }
    public bool CanPull { get; set; }

    public Visibility CanDeleteVisibility => CanDelete ? Visibility.Visible : Visibility.Collapsed;
    public Visibility CanPullVisibility => CanPull ? Visibility.Visible : Visibility.Collapsed;
}

public sealed class SkillListGroup
{
    public string Key { get; }
    public List<SkillListItem> Items { get; }
    public SkillListGroup(string key, List<SkillListItem> items)
    {
        Key = key;
        Items = items;
    }
}
