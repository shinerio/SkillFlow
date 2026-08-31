using System.Collections.ObjectModel;
using System.ComponentModel;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using SkillFlow.Daemon;
using SkillFlow.Skills;

namespace SkillFlow.StarredRepos;

public sealed partial class StarredReposPage : Page, INotifyPropertyChanged
{
    private readonly DaemonClient _client;
    private List<StarRepo> _allRepos = new();
    private List<StarSkillEntry> _allSkills = new();
    private ObservableCollection<StarSkillEntry> _filteredSkills = new();
    private List<string> _categories = new();
    private List<AgentInfo> _agents = new();
    private bool _isLoading = true;
    private bool _isUpdating;
    private bool _isPushing;
    private bool _sortAscending = true;
    private string? _selectedRepoUrl;
    private string _searchText = string.Empty;
    private string? _errorMessage;
    private string? _statusMessage;

    public StarredReposPage() : this(new DaemonClient()) { }

    public StarredReposPage(DaemonClient client)
    {
        _client = client;
        InitializeComponent();
        Loaded += async (_, _) => await LoadAsync();
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    public ObservableCollection<StarSkillEntry> FilteredSkills
    {
        get => _filteredSkills;
        private set { _filteredSkills = value; OnPropertyChanged(); }
    }

    public bool IsLoading
    {
        get => _isLoading;
        set => SetProperty(ref _isLoading, value);
    }

    public bool IsUpdating
    {
        get => _isUpdating;
        set => SetProperty(ref _isUpdating, value);
    }

    public bool IsPushing
    {
        get => _isPushing;
        set => SetProperty(ref _isPushing, value);
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

    public bool CanUpdate => !IsUpdating && !IsLoading && _allRepos.Count > 0;
    public bool HasSelection => SkillListView.SelectedItems.Count > 0 && !IsPushing;

    private async Task LoadAsync()
    {
        IsLoading = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            _allRepos = await _client.InvokeAsync<List<StarRepo>>("starred.listRepos");
            RepoListView.ItemsSource = _allRepos;
            RepoCountText.Text = $"{_allRepos.Count} repo{(_allRepos.Count == 1 ? "" : "s")}";
            _categories = await _client.InvokeAsync<List<string>>("skills.categories.list");
            _agents = await _client.InvokeAsync<List<AgentInfo>>("agents.listEnabled");
            await LoadSkillsAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsLoading = false;
    }

    private async Task LoadSkillsAsync()
    {
        try
        {
            if (!string.IsNullOrEmpty(_selectedRepoUrl))
            {
                _allSkills = await _client.InvokeAsync<List<StarSkillEntry>>("starred.listRepoSkills",
                    new StarredRepoURLParams { RepoURL = _selectedRepoUrl });
            }
            else
            {
                _allSkills = await _client.InvokeAsync<List<StarSkillEntry>>("starred.listAllSkills");
            }
            ApplyFilter();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private void ApplyFilter()
    {
        var result = _allSkills.AsEnumerable();

        if (!string.IsNullOrWhiteSpace(_searchText))
        {
            var search = _searchText;
            result = result.Where(s => s.Name.Contains(search, StringComparison.OrdinalIgnoreCase));
        }

        var sorted = _sortAscending
            ? result.OrderBy(s => s.Name).ToList()
            : result.OrderByDescending(s => s.Name).ToList();

        FilteredSkills = new ObservableCollection<StarSkillEntry>(sorted);
        SkillListView.ItemsSource = FilteredSkills;
    }

    private async Task AddRepoAsync()
    {
        var urlBox = new TextBox { PlaceholderText = "https://github.com/user/repo" };
        var dialog = new ContentDialog
        {
            XamlRoot = XamlRoot,
            Title = "Add Starred Repo",
            Content = new StackPanel
            {
                Spacing = 12,
                Children =
                {
                    new TextBlock { Text = "Repository URL:" },
                    urlBox,
                }
            },
            PrimaryButtonText = "Add",
            CloseButtonText = "Cancel",
            DefaultButton = ContentDialogButton.Primary
        };

        if (await dialog.ShowAsync() != ContentDialogResult.Primary) return;
        if (string.IsNullOrWhiteSpace(urlBox.Text)) return;

        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("starred.addRepo", new StarredRepoURLParams
            {
                RepoURL = urlBox.Text.Trim()
            });
            StatusMessage = $"Added repo: {urlBox.Text.Trim()}";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private async Task RemoveRepoAsync(StarRepo repo)
    {
        var dialog = new ContentDialog
        {
            XamlRoot = XamlRoot,
            Title = "Remove Repo?",
            Content = $"Remove {repo.Name} from starred repos?",
            PrimaryButtonText = "Remove",
            CloseButtonText = "Cancel",
            DefaultButton = ContentDialogButton.Close
        };

        if (await dialog.ShowAsync() != ContentDialogResult.Primary) return;

        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("starred.removeRepo", new StarredRepoURLParams
            {
                RepoURL = repo.Url
            });
            StatusMessage = $"Removed repo: {repo.Name}";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private async Task UpdateRepoAsync(StarRepo repo)
    {
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("starred.updateRepo", new StarredRepoURLParams
            {
                RepoURL = repo.Url
            });
            StatusMessage = $"Updated repo: {repo.Name}";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private async Task UpdateAllAsync()
    {
        IsUpdating = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("starred.updateAll");
            StatusMessage = "All repos updated.";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsUpdating = false;
    }

    private async Task ImportSelectedAsync()
    {
        var selected = SkillListView.SelectedItems.Cast<StarSkillEntry>().ToList();
        if (selected.Count == 0) return;

        var categoryBox = new ComboBox { PlaceholderText = "Uncategorized" };
        foreach (var cat in _categories)
        {
            categoryBox.Items.Add(cat);
        }
        if (_categories.Count > 0) categoryBox.SelectedIndex = 0;

        var dialog = new ContentDialog
        {
            XamlRoot = XamlRoot,
            Title = $"Import {selected.Count} Skill{(selected.Count == 1 ? "" : "s")}",
            Content = new StackPanel
            {
                Spacing = 12,
                Children =
                {
                    new TextBlock { Text = "Select category:" },
                    categoryBox,
                }
            },
            PrimaryButtonText = "Import",
            CloseButtonText = "Cancel",
            DefaultButton = ContentDialogButton.Primary
        };

        if (await dialog.ShowAsync() != ContentDialogResult.Primary) return;

        var category = categoryBox.SelectedItem as string ?? "Uncategorized";
        var repoUrl = _selectedRepoUrl ?? selected.First().RepoUrl;

        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("starred.importSkills", new StarredImportParams
            {
                SkillPaths = selected.Select(s => s.Path).ToList(),
                RepoURL = repoUrl,
                Category = category
            });
            StatusMessage = $"Imported {selected.Count} skill{(selected.Count == 1 ? "" : "s")}.";
            await LoadSkillsAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private async Task PushSelectedAsync()
    {
        var selected = SkillListView.SelectedItems.Cast<StarSkillEntry>().ToList();
        if (selected.Count == 0 || _agents.Count == 0) return;

        var checkboxes = new StackPanel { Spacing = 8 };
        var agentChecks = new Dictionary<string, CheckBox>();
        foreach (var agent in _agents)
        {
            var cb = new CheckBox { Content = agent.Name };
            agentChecks[agent.Name] = cb;
            checkboxes.Children.Add(cb);
        }

        var dialog = new ContentDialog
        {
            XamlRoot = XamlRoot,
            Title = "Push Skills to Agents",
            Content = new StackPanel
            {
                Spacing = 12,
                Children =
                {
                    new TextBlock
                    {
                        Text = $"Pushing {selected.Count} skill{(selected.Count == 1 ? "" : "s")}:",
                        FontWeight = Microsoft.UI.Text.FontWeights.SemiBold
                    },
                    new TextBlock
                    {
                        Text = string.Join(", ", selected.Select(s => s.Name)),
                        Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"],
                        TextWrapping = TextWrapping.Wrap
                    },
                    new TextBlock { Text = "Select target agents:" },
                    checkboxes
                }
            },
            PrimaryButtonText = "Push",
            CloseButtonText = "Cancel",
            DefaultButton = ContentDialogButton.Primary
        };

        if (await dialog.ShowAsync() != ContentDialogResult.Primary) return;

        var targetAgents = agentChecks
            .Where(kvp => kvp.Value.IsChecked == true)
            .Select(kvp => kvp.Key)
            .ToList();

        if (targetAgents.Count == 0) return;

        IsPushing = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("starred.pushToAgents", new StarredPushParams
            {
                SkillPaths = selected.Select(s => s.Path).ToList(),
                AgentNames = targetAgents
            });
            StatusMessage = $"Pushed {selected.Count} skill{(selected.Count == 1 ? "" : "s")} to {targetAgents.Count} agent{(targetAgents.Count == 1 ? "" : "s")}.";
            await LoadSkillsAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsPushing = false;
    }

    private void OnSearchChanged(object sender, TextChangedEventArgs e)
    {
        _searchText = SearchBox.Text;
        ApplyFilter();
    }

    private void OnSortToggle(object sender, RoutedEventArgs e)
    {
        _sortAscending = !_sortAscending;
        SortButton.Content = _sortAscending ? "A-Z" : "Z-A";
        ApplyFilter();
    }

    private async void OnRepoSelectionChanged(object sender, SelectionChangedEventArgs e)
    {
        if (RepoListView.SelectedItem is StarRepo repo)
        {
            _selectedRepoUrl = repo.Url;
        }
        else
        {
            _selectedRepoUrl = null;
        }
        await LoadSkillsAsync();
    }

    private void OnSkillSelectionChanged(object sender, SelectionChangedEventArgs e)
    {
        OnPropertyChanged(nameof(HasSelection));
    }

    private async void OnAddRepo(object sender, RoutedEventArgs e) => await AddRepoAsync();
    private async void OnUpdateAll(object sender, RoutedEventArgs e) => await UpdateAllAsync();
    private async void OnReload(object sender, RoutedEventArgs e) => await LoadAsync();

    private async void OnActions(object sender, RoutedEventArgs e)
    {
        var menu = new MenuFlyout();

        if (RepoListView.SelectedItem is StarRepo repo)
        {
            var updateItem = new MenuFlyoutItem { Text = $"Update {repo.Name}" };
            updateItem.Click += async (_, _) => await UpdateRepoAsync(repo);
            menu.Items.Add(updateItem);

            var removeItem = new MenuFlyoutItem { Text = $"Remove {repo.Name}" };
            removeItem.Click += async (_, _) => await RemoveRepoAsync(repo);
            menu.Items.Add(removeItem);

            menu.Items.Add(new MenuFlyoutSeparator());
        }

        var importItem = new MenuFlyoutItem { Text = "Import Selected" };
        importItem.Click += async (_, _) => await ImportSelectedAsync();
        menu.Items.Add(importItem);

        var pushItem = new MenuFlyoutItem { Text = "Push Selected..." };
        pushItem.Click += async (_, _) => await PushSelectedAsync();
        menu.Items.Add(pushItem);

        menu.ShowAt(ActionsButton, new Windows.Foundation.Point(0, ActionsButton.ActualHeight));
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
        if (name == nameof(IsLoading) || name == nameof(IsUpdating))
        {
            OnPropertyChanged(nameof(CanUpdate));
        }
        if (name == nameof(IsPushing))
        {
            OnPropertyChanged(nameof(HasSelection));
        }
    }
}
