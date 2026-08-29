using System.ComponentModel;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using SkillFlow.Daemon;

namespace SkillFlow.Backup;

public sealed partial class BackupPage : Page, INotifyPropertyChanged
{
    private readonly DaemonClient _client;
    private List<RemoteFile> _remoteFiles = new();
    private List<RemoteFile> _lastChanges = new();
    private string _lastCompletedAt = string.Empty;
    private bool _gitConflictPending;
    private bool _isLoading = true;
    private bool _isBackingUp;
    private bool _isRestoring;
    private bool _isListingFiles;
    private string? _errorMessage;
    private string? _statusMessage;

    public BackupPage() : this(new DaemonClient()) { }

    public BackupPage(DaemonClient client)
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

    public bool IsBackingUp
    {
        get => _isBackingUp;
        set => SetProperty(ref _isBackingUp, value);
    }

    public bool IsRestoring
    {
        get => _isRestoring;
        set => SetProperty(ref _isRestoring, value);
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
            _gitConflictPending = await _client.InvokeAsync<bool>("backup.gitConflictPending");
            _lastChanges = await _client.InvokeAsync<List<RemoteFile>>("backup.lastChanges");
            _lastCompletedAt = await _client.InvokeAsync<string>("backup.lastCompletedAt");
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
        ContentPanel.Children.Add(CreateActionsSection());
        if (_gitConflictPending)
        {
            ContentPanel.Children.Add(CreateConflictSection());
        }
        ContentPanel.Children.Add(CreateStatusSection());
        ContentPanel.Children.Add(CreateChangesSection());
        ContentPanel.Children.Add(CreateFilesSection());
    }

    private UIElement CreateActionsSection()
    {
        var panel = new StackPanel { Spacing = 12 };
        panel.Children.Add(new TextBlock
        {
            Text = "Actions",
            Style = (Style)Application.Current.Resources["SubheaderTextBlockStyle"]
        });

        var buttonPanel = new StackPanel { Orientation = Orientation.Horizontal, Spacing = 8 };

        var backupButton = new Button { Content = "Backup Now" };
        backupButton.Click += async (_, _) => await BackupNowAsync();
        buttonPanel.Children.Add(backupButton);

        var restoreButton = new Button { Content = "Restore" };
        restoreButton.Click += async (_, _) => await RestoreAsync();
        buttonPanel.Children.Add(restoreButton);

        var listButton = new Button { Content = "List Remote Files" };
        listButton.Click += async (_, _) => await ListFilesAsync();
        buttonPanel.Children.Add(listButton);

        panel.Children.Add(buttonPanel);
        return WrapInCard(panel);
    }

    private UIElement CreateConflictSection()
    {
        var panel = new StackPanel { Spacing = 12 };
        panel.Children.Add(new TextBlock
        {
            Text = "Git Conflict Detected",
            Style = (Style)Application.Current.Resources["SubheaderTextBlockStyle"],
            Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["SystemCautionTextColorBrush"]
        });
        panel.Children.Add(new TextBlock
        {
            Text = "A merge conflict was detected during the last sync. Choose how to resolve it.",
            Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
        });

        var conflictButtonPanel = new StackPanel { Orientation = Orientation.Horizontal, Spacing = 8 };
        var localButton = new Button { Content = "Use Local" };
        localButton.Click += async (_, _) => await ResolveConflictAsync(true);
        var remoteButton = new Button { Content = "Use Remote" };
        remoteButton.Click += async (_, _) => await ResolveConflictAsync(false);
        conflictButtonPanel.Children.Add(localButton);
        conflictButtonPanel.Children.Add(remoteButton);
        panel.Children.Add(conflictButtonPanel);

        var border = new Border
        {
            Background = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["SystemCautionBrush"],
            CornerRadius = new CornerRadius(14),
            Padding = new Thickness(18),
            Child = panel
        };
        return border;
    }

    private UIElement CreateStatusSection()
    {
        var panel = new StackPanel { Spacing = 8 };
        panel.Children.Add(new TextBlock
        {
            Text = "Status",
            Style = (Style)Application.Current.Resources["SubheaderTextBlockStyle"]
        });
        panel.Children.Add(new TextBlock
        {
            Text = string.IsNullOrEmpty(_lastCompletedAt) ? "No backup has been completed yet." : $"Last backup: {_lastCompletedAt}",
            Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
        });
        return WrapInCard(panel);
    }

    private UIElement CreateChangesSection()
    {
        var panel = new StackPanel { Spacing = 8 };
        panel.Children.Add(new TextBlock
        {
            Text = "Last Backup Changes",
            Style = (Style)Application.Current.Resources["SubheaderTextBlockStyle"]
        });

        if (_lastChanges.Count == 0)
        {
            panel.Children.Add(new TextBlock
            {
                Text = "No changes in the last backup.",
                Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
            });
        }
        else
        {
            foreach (var change in _lastChanges)
            {
                var row = new StackPanel { Orientation = Orientation.Horizontal, Spacing = 8 };
                var icon = new FontIcon
                {
                    Glyph = change.IsDir ? "\xE8D5" : "\xE8A5",
                    FontSize = 14,
                    Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
                };
                row.Children.Add(icon);
                row.Children.Add(new TextBlock { Text = change.Path, FontSize = 13 });
                if (!string.IsNullOrEmpty(change.Action))
                {
                    row.Children.Add(new TextBlock
                    {
                        Text = change.Action,
                        FontSize = 11,
                        Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
                    });
                }
                panel.Children.Add(row);
            }
        }

        return WrapInCard(panel);
    }

    private UIElement CreateFilesSection()
    {
        var panel = new StackPanel { Spacing = 8 };
        panel.Children.Add(new TextBlock
        {
            Text = "Remote Files",
            Style = (Style)Application.Current.Resources["SubheaderTextBlockStyle"]
        });

        if (_remoteFiles.Count == 0)
        {
            panel.Children.Add(new TextBlock
            {
                Text = "Click 'List Remote Files' to load.",
                Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
            });
        }
        else
        {
            foreach (var file in _remoteFiles)
            {
                var row = new StackPanel { Orientation = Orientation.Horizontal, Spacing = 8 };
                var icon = new FontIcon
                {
                    Glyph = file.IsDir ? "\xE8D5" : "\xE8A5",
                    FontSize = 14,
                    Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
                };
                row.Children.Add(icon);
                row.Children.Add(new TextBlock { Text = file.Path, FontSize = 13 });
                row.Children.Add(new TextBlock
                {
                    Text = FormatSize(file.Size),
                    FontSize = 11,
                    Foreground = (Microsoft.UI.Xaml.Media.Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
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

    private async Task BackupNowAsync()
    {
        IsBackingUp = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("backup.now");
            StatusMessage = "Backup completed.";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsBackingUp = false;
    }

    private async Task RestoreAsync()
    {
        var dialog = new ContentDialog
        {
            XamlRoot = XamlRoot,
            Title = "Restore from Cloud?",
            Content = "This will overwrite local data with the cloud backup.",
            PrimaryButtonText = "Restore",
            CloseButtonText = "Cancel",
            DefaultButton = ContentDialogButton.Close
        };

        if (await dialog.ShowAsync() != ContentDialogResult.Primary) return;

        IsRestoring = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("backup.restore");
            StatusMessage = "Restore completed.";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsRestoring = false;
    }

    private async Task ListFilesAsync()
    {
        _isListingFiles = true;
        ErrorMessage = null;
        try
        {
            _remoteFiles = await _client.InvokeAsync<List<RemoteFile>>("backup.listFiles");
            RenderContent();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        _isListingFiles = false;
    }

    private async Task ResolveConflictAsync(bool useLocal)
    {
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("backup.resolveGitConflict",
                new BackupResolveConflictParams { UseLocal = useLocal });
            StatusMessage = useLocal ? "Conflict resolved: kept local changes." : "Conflict resolved: used remote state.";
            await LoadAsync();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private async void OnReload(object sender, RoutedEventArgs e) => await LoadAsync();

    private static string FormatSize(long bytes)
    {
        if (bytes < 1024) return $"{bytes} B";
        if (bytes < 1024 * 1024) return $"{bytes / 1024} KB";
        return $"{bytes / (1024 * 1024)} MB";
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
