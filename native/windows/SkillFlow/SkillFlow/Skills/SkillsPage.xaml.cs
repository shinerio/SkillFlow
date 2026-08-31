using System.Collections.ObjectModel;
using System.ComponentModel;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using SkillFlow.Daemon;
using Windows.Storage.Pickers;

namespace SkillFlow.Skills;

public sealed partial class SkillsPage : Page, INotifyPropertyChanged
{
    private readonly DaemonClient _client;
    private List<InstalledSkill> _allSkills = new();
    private ObservableCollection<InstalledSkill> _filteredSkills = new();
    private List<string> _categories = new();
    private List<AgentInfo> _agents = new();
    private bool _isLoading = true;
    private bool _isImporting;
    private bool _isCheckingUpdates;
    private bool _isPushing;
    private bool _sortAscending = true;
    private string? _selectedCategory;
    private string _searchText = string.Empty;
    private string? _errorMessage;
    private string? _statusMessage;

    public SkillsPage() : this(new DaemonClient()) { }

    public SkillsPage(DaemonClient client)
    {
        _client = client;
        InitializeComponent();
        Loaded += async (_, _) => await LoadAsync();
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    public ObservableCollection<InstalledSkill> FilteredSkills
    {
        get => _filteredSkills;
        private set
        {
            _filteredSkills = value;
            OnPropertyChanged();
        }
    }

    public bool IsLoading
    {
        get => _isLoading;
        set => SetProperty(ref _isLoading, value);
    }

    public bool IsImporting
    {
        get => _isImporting;
        set => SetProperty(ref _isImporting, value);
    }

    public bool IsCheckingUpdates
    {
        get => _isCheckingUpdates;
        set => SetProperty(ref _isCheckingUpdates, value);
    }

    public bool IsPushing
    {
        get => _isPushing;
        set => SetProperty(ref _isPushing, value);
    }

    public bool SortAscending
    {
        get => _sortAscending;
        set => SetProperty(ref _sortAscending, value);
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

    public bool CanImport => !IsImporting && !IsLoading;
    public bool CanCheckUpdates => !IsCheckingUpdates && !IsLoading && _allSkills.Count > 0;
    public bool HasSelection => SkillListView.SelectedItems.Count > 0 && !IsPushing;

    private async Task LoadAsync()
    {
        IsLoading = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            _allSkills = await _client.InvokeAsync<List<InstalledSkill>>("skills.list");
            _categories = await _client.InvokeAsync<List<string>>("skills.categories.list");
            _agents = await _client.InvokeAsync<List<AgentInfo>>("agents.listEnabled");
            UpdateCategoryList();
            ApplyFilter();
            SkillCountText.Text = $"{_allSkills.Count} skill{(_allSkills.Count == 1 ? "" : "s")}";
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsLoading = false;
    }

    private void UpdateCategoryList()
    {
        CategoryList.Items.Clear();
        CategoryList.Items.Add(new ListViewItem { Tag = "", Content = "All", IsSelected = _selectedCategory == null });
        foreach (var category in _categories)
        {
            CategoryList.Items.Add(new ListViewItem
            {
                Tag = category,
                Content = category,
                IsSelected = _selectedCategory == category
            });
        }
    }

    private void ApplyFilter()
    {
        var result = _allSkills.AsEnumerable();

        if (!string.IsNullOrEmpty(_selectedCategory))
        {
            result = result.Where(s => s.Category == _selectedCategory);
        }

        if (!string.IsNullOrWhiteSpace(_searchText))
        {
            var search = _searchText;
            result = result.Where(s => s.Name.Contains(search, StringComparison.OrdinalIgnoreCase));
        }

        var sorted = _sortAscending
            ? result.OrderBy(s => s.Name).ToList()
            : result.OrderByDescending(s => s.Name).ToList();

        FilteredSkills = new ObservableCollection<InstalledSkill>(sorted);
        SkillListView.ItemsSource = FilteredSkills;
    }

    private async Task ImportAsync()
    {
        if (App.MainWindow is null) return;

        var picker = new FolderPicker();
        picker.FileTypeFilter.Add("*");
        picker.SuggestedStartLocation = PickerLocationId.ComputerFolder;

        var hwnd = WinRT.Interop.WindowNative.GetWindowHandle(App.MainWindow);
        WinRT.Interop.InitializeWithWindow.Initialize(picker, hwnd);

        var folder = await picker.PickSingleFolderAsync();
        if (folder is null) return;

        IsImporting = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            var category = _selectedCategory ?? "Uncategorized";
            await _client.InvokeAsync<object>("skills.importLocal", new SkillsImportLocalParams
            {
                Dir = folder.Path,
                Category = category
            });
            StatusMessage = $"Imported skill from {folder.Name}.";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsImporting = false;
    }

    private async Task DeleteSelectedAsync()
    {
        var selected = SkillListView.SelectedItems.Cast<InstalledSkill>().ToList();
        if (selected.Count == 0) return;

        var dialog = new ContentDialog
        {
            XamlRoot = XamlRoot,
            Title = $"Delete {selected.Count} skill{(selected.Count == 1 ? "" : "s")}?",
            Content = "This action cannot be undone.",
            PrimaryButtonText = "Delete",
            CloseButtonText = "Cancel",
            DefaultButton = ContentDialogButton.Close
        };

        if (await dialog.ShowAsync() != ContentDialogResult.Primary) return;

        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            var ids = selected.Select(s => s.Id).ToList();
            if (ids.Count == 1)
            {
                await _client.InvokeAsync<object>("skills.delete", new SkillsDeleteParams { SkillID = ids[0] });
            }
            else
            {
                await _client.InvokeAsync<object>("skills.deleteBatch", new SkillsDeleteBatchParams { SkillIDs = ids });
            }
            StatusMessage = $"Deleted {ids.Count} skill{(ids.Count == 1 ? "" : "s")}.";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private async Task CheckUpdatesAsync()
    {
        IsCheckingUpdates = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("skills.updateCheck");
            var updatable = _allSkills.Count(s => s.Updatable);
            StatusMessage = updatable == 0
                ? "All skills are up to date."
                : $"{updatable} skill{(updatable == 1 ? "" : "s")} can be updated.";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsCheckingUpdates = false;
    }

    private async Task PushAsync(List<string> agentNames)
    {
        var selected = SkillListView.SelectedItems.Cast<InstalledSkill>().ToList();
        if (selected.Count == 0 || agentNames.Count == 0) return;

        IsPushing = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            var ids = selected.Select(s => s.Id).ToList();

            // Check for missing push directories before pushing.
            var missing = await _client.InvokeAsync<List<MissingPushDir>>("agents.checkMissingPushDirs",
                new AgentNamesParams { AgentNames = agentNames });
            if (missing.Count > 0)
            {
                var dirList = string.Join("\n", missing.Select(m => $"- {m.Name}: {m.Dir}"));
                var missingDialog = new ContentDialog
                {
                    XamlRoot = XamlRoot,
                    Title = "Missing Push Directories",
                    Content = $"The following agents have missing push directories:\n\n{dirList}\n\nCreate them and continue?",
                    PrimaryButtonText = "Create and Push",
                    CloseButtonText = "Cancel",
                    DefaultButton = ContentDialogButton.Primary
                };
                if (await missingDialog.ShowAsync() != ContentDialogResult.Primary)
                {
                    IsPushing = false;
                    return;
                }
            }

            var conflicts = await _client.InvokeAsync<List<PushConflict>>("skills.push", new SkillsPushParams
            {
                SkillIDs = ids,
                AgentNames = agentNames
            });

            if (conflicts.Count == 0)
            {
                StatusMessage = $"Pushed {ids.Count} skill{(ids.Count == 1 ? "" : "s")} to {agentNames.Count} agent{(agentNames.Count == 1 ? "" : "s")}.";
            }
            else
            {
                var conflictList = string.Join("\n", conflicts.Select(c => $"- {c.SkillName} -> {c.AgentName}"));
                var conflictDialog = new ContentDialog
                {
                    XamlRoot = XamlRoot,
                    Title = "Push Conflicts",
                    Content = $"{conflicts.Count} conflict{(conflicts.Count == 1 ? "" : "s")} detected:\n\n{conflictList}\n\nForce push to overwrite?",
                    PrimaryButtonText = "Force Push",
                    CloseButtonText = "Skip",
                    DefaultButton = ContentDialogButton.Close
                };
                if (await conflictDialog.ShowAsync() == ContentDialogResult.Primary)
                {
                    await _client.InvokeAsync<NativeEmptyResult>("skills.pushForce", new SkillsPushParams
                    {
                        SkillIDs = ids,
                        AgentNames = agentNames
                    });
                    StatusMessage = $"Force pushed {ids.Count} skill{(ids.Count == 1 ? "" : "s")}.";
                }
                else
                {
                    StatusMessage = $"Pushed with {conflicts.Count} conflict{(conflicts.Count == 1 ? "" : "s")} skipped.";
                }
            }
            await LoadAsync();
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
        SortAscending = !SortAscending;
        SortButton.Content = SortAscending ? "A-Z" : "Z-A";
        ApplyFilter();
    }

    private void OnCategorySelectionChanged(object sender, SelectionChangedEventArgs e)
    {
        if (CategoryList.SelectedItem is ListViewItem item)
        {
            var tag = item.Tag as string ?? "";
            _selectedCategory = string.IsNullOrEmpty(tag) ? null : tag;
            ApplyFilter();
        }
    }

    private async void OnImport(object sender, RoutedEventArgs e) => await ImportAsync();
    private async void OnCheckUpdates(object sender, RoutedEventArgs e) => await CheckUpdatesAsync();
    private async void OnReload(object sender, RoutedEventArgs e) => await LoadAsync();
    private async void OnAddCategory(object sender, RoutedEventArgs e) => await AddCategoryAsync();

    private async Task AddCategoryAsync()
    {
        var nameBox = new TextBox { PlaceholderText = "category-name" };
        var dialog = new ContentDialog
        {
            XamlRoot = XamlRoot,
            Title = "New Category",
            Content = new StackPanel
            {
                Spacing = 12,
                Children = { new TextBlock { Text = "Category name:" }, nameBox }
            },
            PrimaryButtonText = "Create",
            CloseButtonText = "Cancel",
            DefaultButton = ContentDialogButton.Primary
        };

        if (await dialog.ShowAsync() != ContentDialogResult.Primary) return;
        if (string.IsNullOrWhiteSpace(nameBox.Text)) return;

        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("skills.categories.create", new SkillsCategoryNameParams
            {
                Name = nameBox.Text.Trim()
            });
            StatusMessage = $"Created category: {nameBox.Text.Trim()}";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private async Task UpdateSkillAsync(InstalledSkill skill)
    {
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("skills.updateOne", new SkillsDeleteParams
            {
                SkillID = skill.Id
            });
            StatusMessage = $"Updated skill: {skill.Name}";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private async Task MoveSkillCategoryAsync(InstalledSkill skill, string category)
    {
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("skills.moveCategory", new SkillsMoveCategoryParams
            {
                SkillID = skill.Id,
                Category = category
            });
            StatusMessage = $"Moved {skill.Name} to {category}.";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private void OnSkillSelectionChanged(object sender, SelectionChangedEventArgs e)
    {
        OnPropertyChanged(nameof(HasSelection));
    }

    private async void OnActions(object sender, RoutedEventArgs e)
    {
        var menu = new MenuFlyout();

        if (SkillListView.SelectedItems.Count == 1 &&
            SkillListView.SelectedItem is InstalledSkill skill)
        {
            if (skill.Updatable)
            {
                var updateItem = new MenuFlyoutItem { Text = $"Update {skill.Name}" };
                updateItem.Click += async (_, _) => await UpdateSkillAsync(skill);
                menu.Items.Add(updateItem);
            }

            var moveItem = new MenuFlyoutItem { Text = "Move to Category..." };
            moveItem.Click += async (_, _) => await ShowMoveCategoryDialogAsync(skill);
            menu.Items.Add(moveItem);

            menu.Items.Add(new MenuFlyoutSeparator());
        }

        var deleteItem = new MenuFlyoutItem { Text = "Delete Selected" };
        deleteItem.Click += async (_, _) => await DeleteSelectedAsync();
        menu.Items.Add(deleteItem);

        var pushItem = new MenuFlyoutItem { Text = "Push Selected..." };
        pushItem.Click += async (_, _) => await ShowPushDialogAsync();
        menu.Items.Add(pushItem);

        menu.ShowAt(ActionsButton, new Windows.Foundation.Point(0, ActionsButton.ActualHeight));
    }

    private async Task ShowMoveCategoryDialogAsync(InstalledSkill skill)
    {
        if (_categories.Count == 0) return;

        var categoryBox = new ComboBox();
        foreach (var cat in _categories)
        {
            categoryBox.Items.Add(cat);
        }
        var idx = _categories.IndexOf(skill.Category);
        if (idx >= 0) categoryBox.SelectedIndex = idx;
        else if (_categories.Count > 0) categoryBox.SelectedIndex = 0;

        var dialog = new ContentDialog
        {
            XamlRoot = XamlRoot,
            Title = $"Move {skill.Name}",
            Content = new StackPanel
            {
                Spacing = 12,
                Children = { new TextBlock { Text = "Select category:" }, categoryBox }
            },
            PrimaryButtonText = "Move",
            CloseButtonText = "Cancel",
            DefaultButton = ContentDialogButton.Primary
        };

        if (await dialog.ShowAsync() != ContentDialogResult.Primary) return;
        var category = categoryBox.SelectedItem as string;
        if (!string.IsNullOrEmpty(category))
        {
            await MoveSkillCategoryAsync(skill, category);
        }
    }

    private async Task ShowPushDialogAsync()
    {
        var selected = SkillListView.SelectedItems.Cast<InstalledSkill>().ToList();
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

        if (targetAgents.Count > 0)
        {
            await PushAsync(targetAgents);
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
        if (name == nameof(IsLoading) || name == nameof(IsCheckingUpdates))
        {
            OnPropertyChanged(nameof(CanCheckUpdates));
        }
        if (name == nameof(IsLoading) || name == nameof(IsImporting))
        {
            OnPropertyChanged(nameof(CanImport));
        }
        if (name == nameof(IsPushing))
        {
            OnPropertyChanged(nameof(HasSelection));
        }
    }
}
