using System.ComponentModel;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using SkillFlow.Daemon;

namespace SkillFlow.Memory;

public sealed partial class MemoryPage : Page, INotifyPropertyChanged
{
    private readonly DaemonClient _client;
    private MainMemory? _mainMemory;
    private List<ModuleMemory> _modules = new();
    private List<MemoryPushConfig> _pushConfigs = new();
    private List<PushStatus> _pushStatuses = new();
    private bool _isLoading = true;
    private bool _isPushingAll;
    private string? _errorMessage;
    private string? _statusMessage;

    public MemoryPage() : this(new DaemonClient()) { }

    public MemoryPage(DaemonClient client)
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

    public bool IsPushingAll
    {
        get => _isPushingAll;
        set => SetProperty(ref _isPushingAll, value);
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
            _mainMemory = await _client.InvokeAsync<MainMemory>("memory.main.get");
            _modules = await _client.InvokeAsync<List<ModuleMemory>>("memory.modules.list");
            _pushConfigs = await _client.InvokeAsync<List<MemoryPushConfig>>("memory.pushConfig.getAll");
            _pushStatuses = await _client.InvokeAsync<List<PushStatus>>("memory.pushStatus.getAll");
            RenderContent();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsLoading = false;
    }

    private void RenderContent()
    {
        ContentPanel.Children.Clear();

        ContentPanel.Children.Add(CreateMainMemorySection());
        ContentPanel.Children.Add(CreateModulesSection());
        ContentPanel.Children.Add(CreatePushConfigSection());
        ContentPanel.Children.Add(CreatePushStatusSection());
    }

    private UIElement CreateMainMemorySection()
    {
        var panel = new StackPanel { Spacing = 12 };
        panel.Children.Add(new TextBlock
        {
            Text = "Main Memory",
            Style = (Style)Application.Current.Resources["SubheaderTextBlockStyle"]
        });

        if (_mainMemory is not null)
        {
            var editor = new TextBox
            {
                Text = _mainMemory.Content,
                AcceptsReturn = true,
                TextWrapping = TextWrapping.Wrap,
                MinHeight = 120,
                FontFamily = new Microsoft.UI.Xaml.Media.FontFamily("Consolas")
            };
            panel.Children.Add(editor);

            var buttonPanel = new StackPanel { Orientation = Orientation.Horizontal, Spacing = 8 };
            var openButton = new Button { Content = "Open in Editor" };
            openButton.Click += async (_, _) =>
            {
                try
                {
                    await _client.InvokeAsync<object>("memory.openInEditor",
                        new MemoryOpenEditorParams { MemoryType = "main", ModuleName = "" });
                }
                catch (Exception ex) { ErrorMessage = ex.Message; }
            };
            var saveButton = new Button { Content = "Save" };
            saveButton.Click += async (_, _) =>
            {
                try
                {
                    await _client.InvokeAsync<object>("memory.main.save",
                        new MemoryContentParams { Content = editor.Text });
                    StatusMessage = "Main memory saved.";
                    await LoadAsync();
                }
                catch (Exception ex) { ErrorMessage = ex.Message; }
            };
            buttonPanel.Children.Add(openButton);
            buttonPanel.Children.Add(saveButton);
            panel.Children.Add(buttonPanel);
        }

        return WrapInCard(panel);
    }

    private UIElement CreateModulesSection()
    {
        var panel = new StackPanel { Spacing = 12 };

        var headerPanel = new StackPanel { Orientation = Orientation.Horizontal, Spacing = 8 };
        headerPanel.Children.Add(new TextBlock
        {
            Text = "Module Memories",
            Style = (Style)Application.Current.Resources["SubheaderTextBlockStyle"]
        });
        var newButton = new Button { Content = "New Module" };
        newButton.Click += async (_, _) => await CreateModuleAsync();
        headerPanel.Children.Add(newButton);
        panel.Children.Add(headerPanel);

        if (_modules.Count == 0)
        {
            panel.Children.Add(new TextBlock
            {
                Text = "No module memories configured.",
                Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
            });
        }
        else
        {
            foreach (var module in _modules)
            {
                panel.Children.Add(CreateModuleCard(module));
            }
        }

        return WrapInCard(panel);
    }

    private UIElement CreateModuleCard(ModuleMemory module)
    {
        var panel = new StackPanel { Spacing = 8, Padding = new Thickness(12) };

        var headerPanel = new StackPanel { Orientation = Orientation.Horizontal, Spacing = 8 };
        var toggle = new ToggleSwitch
        {
            Header = module.Name,
            IsOn = module.Enabled
        };
        toggle.Toggled += async (_, _) =>
        {
            try
            {
                await _client.InvokeAsync<object>("memory.modules.setEnabled",
                    new ModuleMemoryEnabledParams { Name = module.Name, Enabled = toggle.IsOn });
            }
            catch (Exception ex) { ErrorMessage = ex.Message; }
        };
        headerPanel.Children.Add(toggle);

        var actionsPanel = new StackPanel { Orientation = Orientation.Horizontal, Spacing = 8 };
        var openButton = new Button { Content = "Open in Editor" };
        openButton.Click += async (_, _) =>
        {
            try
            {
                await _client.InvokeAsync<object>("memory.openInEditor",
                    new MemoryOpenEditorParams { MemoryType = "module", ModuleName = module.Name });
            }
            catch (Exception ex) { ErrorMessage = ex.Message; }
        };
        var deleteButton = new Button { Content = "Delete" };
        deleteButton.Click += async (_, _) =>
        {
            try
            {
                await _client.InvokeAsync<object>("memory.modules.delete",
                    new ModuleMemoryNameParams { Name = module.Name });
                StatusMessage = $"Deleted module '{module.Name}'.";
                await LoadAsync();
            }
            catch (Exception ex) { ErrorMessage = ex.Message; }
        };
        actionsPanel.Children.Add(openButton);
        actionsPanel.Children.Add(deleteButton);
        headerPanel.Children.Add(actionsPanel);
        panel.Children.Add(headerPanel);

        var editor = new TextBox
        {
            Text = module.Content,
            AcceptsReturn = true,
            TextWrapping = TextWrapping.Wrap,
            MinHeight = 80,
            FontFamily = new Microsoft.UI.Xaml.Media.FontFamily("Consolas")
        };
        panel.Children.Add(editor);

        var saveButton = new Button { Content = "Save" };
        saveButton.Click += async (_, _) =>
        {
            try
            {
                await _client.InvokeAsync<object>("memory.modules.save",
                    new ModuleMemoryContentParams { Name = module.Name, Content = editor.Text });
                StatusMessage = $"Module '{module.Name}' saved.";
                await LoadAsync();
            }
            catch (Exception ex) { ErrorMessage = ex.Message; }
        };
        panel.Children.Add(saveButton);

        var border = new Border
        {
            Background = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["CardBackgroundFillColorDefaultBrush"],
            CornerRadius = new CornerRadius(8),
            Padding = new Thickness(12),
            Child = panel
        };
        return border;
    }

    private UIElement CreatePushConfigSection()
    {
        var panel = new StackPanel { Spacing = 12 };
        panel.Children.Add(new TextBlock
        {
            Text = "Push Configuration",
            Style = (Style)Application.Current.Resources["SubheaderTextBlockStyle"]
        });

        if (_pushConfigs.Count == 0)
        {
            panel.Children.Add(new TextBlock
            {
                Text = "No push configurations found.",
                Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
            });
        }
        else
        {
            foreach (var config in _pushConfigs)
            {
                var row = new StackPanel
                {
                    Orientation = Orientation.Horizontal,
                    Spacing = 16,
                    Padding = new Thickness(12)
                };
                row.Children.Add(new TextBlock { Text = config.AgentType, FontWeight = Microsoft.UI.Text.FontWeights.SemiBold });
                row.Children.Add(new TextBlock { Text = $"Mode: {config.Mode}" });
                row.Children.Add(new TextBlock { Text = config.AutoPush ? "Auto Push: On" : "Auto Push: Off" });
                panel.Children.Add(row);
            }
        }

        return WrapInCard(panel);
    }

    private UIElement CreatePushStatusSection()
    {
        var panel = new StackPanel { Spacing = 12 };
        panel.Children.Add(new TextBlock
        {
            Text = "Push Status",
            Style = (Style)Application.Current.Resources["SubheaderTextBlockStyle"]
        });

        if (_pushStatuses.Count == 0)
        {
            panel.Children.Add(new TextBlock
            {
                Text = "No push status available.",
                Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
            });
        }
        else
        {
            foreach (var status in _pushStatuses)
            {
                var row = new StackPanel
                {
                    Orientation = Orientation.Horizontal,
                    Spacing = 16,
                    Padding = new Thickness(8)
                };
                row.Children.Add(new TextBlock { Text = status.AgentType });
                row.Children.Add(new TextBlock
                {
                    Text = status.Status,
                    FontSize = 12,
                    Foreground = status.Status == "synced"
                        ? (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["SystemFillColorSuccessBrush"]
                        : (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
                });
                panel.Children.Add(row);
            }
        }

        return WrapInCard(panel);
    }

    private UIElement WrapInCard(StackPanel panel)
    {
        return new Border
        {
            Background = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["LayerFillColorDefaultBrush"],
            CornerRadius = new CornerRadius(14),
            Padding = new Thickness(18),
            Child = panel
        };
    }

    private async Task CreateModuleAsync()
    {
        var input = new TextBox { PlaceholderText = "Module name" };
        var dialog = new ContentDialog
        {
            XamlRoot = XamlRoot,
            Title = "New Module Memory",
            Content = input,
            PrimaryButtonText = "Create",
            CloseButtonText = "Cancel",
            DefaultButton = ContentDialogButton.Primary
        };

        if (await dialog.ShowAsync() != ContentDialogResult.Primary) return;
        var name = input.Text.Trim();
        if (string.IsNullOrEmpty(name)) return;

        try
        {
            await _client.InvokeAsync<object>("memory.modules.create",
                new ModuleMemoryContentParams { Name = name, Content = "" });
            StatusMessage = $"Created module '{name}'.";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private async void OnPushAll(object sender, RoutedEventArgs e)
    {
        IsPushingAll = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            var results = await _client.InvokeAsync<List<PushResult>>("memory.pushAll");
            var successCount = results.Count(r => r.Success);
            StatusMessage = $"Pushed to {successCount} of {results.Count} agent{(results.Count == 1 ? "" : "s")}.";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsPushingAll = false;
    }

    private async void OnReload(object sender, RoutedEventArgs e) => await LoadAsync();

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
