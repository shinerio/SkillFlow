using System.Collections.ObjectModel;
using System.ComponentModel;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using SkillFlow.Daemon;

namespace SkillFlow.Prompts;

public sealed partial class PromptsPage : Page, INotifyPropertyChanged
{
    private readonly DaemonClient _client;
    private List<PromptEntry> _allPrompts = new();
    private ObservableCollection<PromptEntry> _filteredPrompts = new();
    private List<string> _categories = new();
    private bool _isLoading = true;
    private bool _isImporting;
    private bool _isExporting;
    private bool _sortAscending = true;
    private string? _selectedCategory;
    private string _searchText = string.Empty;
    private string? _errorMessage;
    private string? _statusMessage;

    public PromptsPage() : this(new DaemonClient()) { }

    public PromptsPage(DaemonClient client)
    {
        _client = client;
        InitializeComponent();
        Loaded += async (_, _) => await LoadAsync();
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    public ObservableCollection<PromptEntry> FilteredPrompts
    {
        get => _filteredPrompts;
        private set { _filteredPrompts = value; OnPropertyChanged(); }
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

    public bool IsExporting
    {
        get => _isExporting;
        set => SetProperty(ref _isExporting, value);
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
    public bool CanExport => !IsExporting && !IsLoading && _allPrompts.Count > 0;
    public bool HasSelection => PromptListView.SelectedItems.Count > 0;

    private async Task LoadAsync()
    {
        IsLoading = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            _allPrompts = await _client.InvokeAsync<List<PromptEntry>>("prompts.list");
            _categories = await _client.InvokeAsync<List<string>>("prompts.categories.list");
            UpdateCategoryList();
            ApplyFilter();
            PromptCountText.Text = $"{_allPrompts.Count} prompt{(_allPrompts.Count == 1 ? "" : "s")}";
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
        var result = _allPrompts.AsEnumerable();

        if (!string.IsNullOrEmpty(_selectedCategory))
        {
            result = result.Where(p => p.Category == _selectedCategory);
        }

        if (!string.IsNullOrWhiteSpace(_searchText))
        {
            var search = _searchText;
            result = result.Where(p => p.Name.Contains(search, StringComparison.OrdinalIgnoreCase)
                || p.Description.Contains(search, StringComparison.OrdinalIgnoreCase));
        }

        var sorted = _sortAscending
            ? result.OrderBy(p => p.Name).ToList()
            : result.OrderByDescending(p => p.Name).ToList();

        FilteredPrompts = new ObservableCollection<PromptEntry>(sorted);
        PromptListView.ItemsSource = FilteredPrompts;
    }

    private async Task AddPromptAsync()
    {
        var nameBox = new TextBox { PlaceholderText = "prompt-name" };
        var descBox = new TextBox { PlaceholderText = "Description (optional)" };
        var contentBox = new TextBox
        {
            PlaceholderText = "Prompt content...",
            AcceptsReturn = true,
            TextWrapping = TextWrapping.Wrap,
            MinHeight = 120
        };
        var categoryBox = new ComboBox { PlaceholderText = "Uncategorized" };
        foreach (var cat in _categories)
        {
            categoryBox.Items.Add(cat);
        }
        if (_categories.Count > 0) categoryBox.SelectedIndex = 0;

        var dialog = new ContentDialog
        {
            XamlRoot = XamlRoot,
            Title = "New Prompt",
            Content = new StackPanel
            {
                Spacing = 12,
                Children =
                {
                    new TextBlock { Text = "Name:" },
                    nameBox,
                    new TextBlock { Text = "Description:" },
                    descBox,
                    new TextBlock { Text = "Category:" },
                    categoryBox,
                    new TextBlock { Text = "Content:" },
                    contentBox,
                }
            },
            PrimaryButtonText = "Create",
            CloseButtonText = "Cancel",
            DefaultButton = ContentDialogButton.Primary
        };

        if (await dialog.ShowAsync() != ContentDialogResult.Primary) return;
        if (string.IsNullOrWhiteSpace(nameBox.Text)) return;

        var category = categoryBox.SelectedItem as string ?? "Uncategorized";
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("prompts.create", new PromptCreateParams
            {
                Name = nameBox.Text.Trim(),
                Description = descBox.Text.Trim(),
                Category = category,
                Content = contentBox.Text
            });
            StatusMessage = $"Created prompt: {nameBox.Text.Trim()}";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private async Task EditPromptAsync(PromptEntry prompt)
    {
        var nameBox = new TextBox { Text = prompt.Name };
        var descBox = new TextBox { Text = prompt.Description };
        var contentBox = new TextBox
        {
            Text = prompt.Content,
            AcceptsReturn = true,
            TextWrapping = TextWrapping.Wrap,
            MinHeight = 120
        };
        var categoryBox = new ComboBox();
        foreach (var cat in _categories)
        {
            categoryBox.Items.Add(cat);
        }
        var idx = _categories.IndexOf(prompt.Category);
        if (idx >= 0) categoryBox.SelectedIndex = idx;
        else if (_categories.Count > 0) categoryBox.SelectedIndex = 0;

        var dialog = new ContentDialog
        {
            XamlRoot = XamlRoot,
            Title = "Edit Prompt",
            Content = new StackPanel
            {
                Spacing = 12,
                Children =
                {
                    new TextBlock { Text = "Name:" },
                    nameBox,
                    new TextBlock { Text = "Description:" },
                    descBox,
                    new TextBlock { Text = "Category:" },
                    categoryBox,
                    new TextBlock { Text = "Content:" },
                    contentBox,
                }
            },
            PrimaryButtonText = "Save",
            CloseButtonText = "Cancel",
            DefaultButton = ContentDialogButton.Primary
        };

        if (await dialog.ShowAsync() != ContentDialogResult.Primary) return;

        var category = categoryBox.SelectedItem as string ?? "Uncategorized";
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("prompts.update", new PromptUpdateParams
            {
                OriginalName = prompt.Name,
                Name = nameBox.Text.Trim(),
                Description = descBox.Text.Trim(),
                Category = category,
                Content = contentBox.Text
            });
            StatusMessage = $"Updated prompt: {nameBox.Text.Trim()}";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private async Task DeleteSelectedAsync()
    {
        var selected = PromptListView.SelectedItems.Cast<PromptEntry>().ToList();
        if (selected.Count == 0) return;

        var dialog = new ContentDialog
        {
            XamlRoot = XamlRoot,
            Title = $"Delete {selected.Count} prompt{(selected.Count == 1 ? "" : "s")}?",
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
            foreach (var prompt in selected)
            {
                await _client.InvokeAsync<object>("prompts.delete", new PromptNameParams { Name = prompt.Name });
            }
            StatusMessage = $"Deleted {selected.Count} prompt{(selected.Count == 1 ? "" : "s")}.";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

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
        try
        {
            await _client.InvokeAsync<object>("prompts.categories.create", new PromptCategoryNameParams
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

    private async Task ImportAsync()
    {
        IsImporting = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("prompts.import");
            StatusMessage = "Import completed.";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsImporting = false;
    }

    private async Task ExportAsync()
    {
        var selected = PromptListView.SelectedItems.Cast<PromptEntry>().ToList();
        IsExporting = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            if (selected.Count > 0)
            {
                await _client.InvokeAsync<object>("prompts.exportByNames", new PromptExportByNamesParams
                {
                    Names = selected.Select(p => p.Name).ToList()
                });
                StatusMessage = $"Exported {selected.Count} prompt{(selected.Count == 1 ? "" : "s")}.";
            }
            else
            {
                await _client.InvokeAsync<object>("prompts.export");
                StatusMessage = "Exported all prompts.";
            }
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsExporting = false;
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

    private void OnCategorySelectionChanged(object sender, SelectionChangedEventArgs e)
    {
        if (CategoryList.SelectedItem is ListViewItem item)
        {
            var tag = item.Tag as string ?? "";
            _selectedCategory = string.IsNullOrEmpty(tag) ? null : tag;
            ApplyFilter();
        }
    }

    private void OnPromptSelectionChanged(object sender, SelectionChangedEventArgs e)
    {
        OnPropertyChanged(nameof(HasSelection));
    }

    private async void OnAddPrompt(object sender, RoutedEventArgs e) => await AddPromptAsync();
    private async void OnAddCategory(object sender, RoutedEventArgs e) => await AddCategoryAsync();
    private async void OnImport(object sender, RoutedEventArgs e) => await ImportAsync();
    private async void OnExport(object sender, RoutedEventArgs e) => await ExportAsync();
    private async void OnReload(object sender, RoutedEventArgs e) => await LoadAsync();

    private async void OnActions(object sender, RoutedEventArgs e)
    {
        var menu = new MenuFlyout();

        var editItem = new MenuFlyoutItem { Text = "Edit Selected" };
        editItem.Click += async (_, _) =>
        {
            if (PromptListView.SelectedItems.Count == 1 &&
                PromptListView.SelectedItem is PromptEntry prompt)
            {
                await EditPromptAsync(prompt);
            }
        };
        menu.Items.Add(editItem);

        var deleteItem = new MenuFlyoutItem { Text = "Delete Selected" };
        deleteItem.Click += async (_, _) => await DeleteSelectedAsync();
        menu.Items.Add(deleteItem);

        menu.ShowAt(ActionsButton, new Windows.Foundation.Point(0, ActionsButton.ActualHeight));
    }

    private async void OnItemDoubleClick(object sender, DoubleTappedRoutedEventArgs e)
    {
        if (sender is FrameworkElement { DataContext: PromptEntry prompt })
        {
            await EditPromptAsync(prompt);
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
        if (name == nameof(IsLoading) || name == nameof(IsImporting))
        {
            OnPropertyChanged(nameof(CanImport));
        }
        if (name == nameof(IsLoading) || name == nameof(IsExporting))
        {
            OnPropertyChanged(nameof(CanExport));
        }
    }
}
