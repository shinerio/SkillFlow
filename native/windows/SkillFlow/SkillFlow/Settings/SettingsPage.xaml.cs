using System.Collections.ObjectModel;
using System.ComponentModel;
using System.Runtime.CompilerServices;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using SkillFlow.Daemon;
using Windows.Storage.Pickers;

namespace SkillFlow.Settings;

public sealed partial class SettingsPage : Page, INotifyPropertyChanged
{
    private readonly DaemonClient _client;

    private AppSettings _draft = new();
    private ObservableCollection<string> _providerNames = new();
    private bool _isLoading = true;
    private bool _isSaving;
    private bool _isTestingProxy;
    private bool _isCheckingUpdate;
    private string? _errorMessage;
    private string? _statusMessage;
    private AppUpdateInfo? _updateInfo;
    private string? _selectedProviderName;

    public SettingsPage() : this(new DaemonClient()) { }

    public SettingsPage(DaemonClient client)
    {
        _client = client;
        InitializeComponent();
        Loaded += async (_, _) => await LoadAsync();
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    public AppSettings Draft
    {
        get => _draft;
        set => SetProperty(ref _draft, value);
    }

    public ObservableCollection<string> ProviderNames
    {
        get => _providerNames;
        private set => SetProperty(ref _providerNames, value);
    }

    public string? SelectedProviderName
    {
        get => _selectedProviderName;
        set
        {
            if (SetProperty(ref _selectedProviderName, value) && value != null)
            {
                _draft.Cloud.Provider = value;
            }
        }
    }

    public bool IsLoading
    {
        get => _isLoading;
        set => SetProperty(ref _isLoading, value);
    }

    public bool IsSaving
    {
        get => _isSaving;
        set => SetProperty(ref _isSaving, value);
    }

    public bool IsTestingProxy
    {
        get => _isTestingProxy;
        set => SetProperty(ref _isTestingProxy, value);
    }

    public bool IsCheckingUpdate
    {
        get => _isCheckingUpdate;
        set => SetProperty(ref _isCheckingUpdate, value);
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

    public AppUpdateInfo? UpdateInfo
    {
        get => _updateInfo;
        set => SetProperty(ref _updateInfo, value);
    }

    public int LogLevelIndex
    {
        get => _draft.LogLevel switch
        {
            "debug" => 0,
            "error" => 2,
            _ => 1,
        };
        set
        {
            _draft.LogLevel = value switch
            {
                0 => "debug",
                2 => "error",
                _ => "info",
            };
            OnPropertyChanged();
        }
    }

    public int ProxyModeIndex
    {
        get => _draft.Proxy.Mode switch
        {
            "system" => 1,
            "manual" => 2,
            _ => 0,
        };
        set
        {
            _draft.Proxy.Mode = value switch
            {
                1 => "system",
                2 => "manual",
                _ => "none",
            };
            OnPropertyChanged();
            OnPropertyChanged(nameof(ManualProxyVisibility));
        }
    }

    public bool CanInteract => !IsLoading && !IsSaving;
    public bool CanTestProxy => !IsTestingProxy && !IsLoading && !IsSaving;
    public bool CanCheckUpdate => !IsCheckingUpdate && !IsLoading && !IsSaving;

    public Visibility LoadingVisibility => IsLoading ? Visibility.Visible : Visibility.Collapsed;
    public Visibility ContentVisibility => IsLoading ? Visibility.Collapsed : Visibility.Visible;
    public Visibility SavingVisibility => IsSaving ? Visibility.Visible : Visibility.Collapsed;
    public Visibility TestingProxyVisibility => IsTestingProxy ? Visibility.Visible : Visibility.Collapsed;
    public Visibility CheckingUpdateVisibility => IsCheckingUpdate ? Visibility.Visible : Visibility.Collapsed;
    public Visibility ManualProxyVisibility => _draft.Proxy.Mode == "manual" ? Visibility.Visible : Visibility.Collapsed;
    public Visibility ErrorVisibility => !string.IsNullOrEmpty(ErrorMessage) ? Visibility.Visible : Visibility.Collapsed;
    public Visibility StatusVisibility => !string.IsNullOrEmpty(StatusMessage) ? Visibility.Visible : Visibility.Collapsed;
    public Visibility NoAgentsVisibility => _draft.Agents.Count == 0 ? Visibility.Visible : Visibility.Collapsed;
    public Visibility UpdateInfoVisibility => UpdateInfo != null ? Visibility.Visible : Visibility.Collapsed;

    public string UpdateStatusText => UpdateInfo?.HasUpdate == true ? "Update available" : "You are up to date";

    private async Task LoadAsync()
    {
        IsLoading = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            Draft = await _client.InvokeAsync<AppSettings>("settings.get");
            var providerList = await _client.InvokeAsync<List<CloudProviderInfo>>("backup.providers.list");
            ProviderNames = new ObservableCollection<string>(providerList.Select(p => p.Name));
            if (string.IsNullOrEmpty(Draft.Cloud.Provider) && ProviderNames.Count > 0)
            {
                SelectedProviderName = ProviderNames.First();
            }
            else
            {
                SelectedProviderName = Draft.Cloud.Provider;
            }
            RefreshBindings();
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsLoading = false;
    }

    private async Task SaveAsync()
    {
        IsSaving = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            if (!string.IsNullOrEmpty(Draft.Cloud.Provider))
            {
                Draft.CloudProfiles[Draft.Cloud.Provider] = new CloudProviderProfile
                {
                    BucketName = Draft.Cloud.BucketName,
                    RemotePath = Draft.Cloud.RemotePath,
                    Credentials = Draft.Cloud.Credentials,
                };
            }
            await _client.InvokeAsync<object>("settings.save", Draft);
            StatusMessage = "Settings saved.";
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsSaving = false;
    }

    private async Task TestProxyAsync()
    {
        IsTestingProxy = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            var result = await _client.InvokeAsync<ProxyConnectionTestResult>("proxy.test", new ProxyTestParameters
            {
                TargetURL = "https://github.com",
                Proxy = Draft.Proxy,
            });
            StatusMessage = $"{(result.Success ? "Connected" : "Failed")} in {result.ElapsedMs} ms · {result.Message}";
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsTestingProxy = false;
    }

    private async Task CheckUpdateAsync()
    {
        IsCheckingUpdate = true;
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            UpdateInfo = await _client.InvokeAsync<AppUpdateInfo>("app.update.check");
            OnPropertyChanged(nameof(UpdateStatusText));
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
        IsCheckingUpdate = false;
    }

    private async Task OpenLogDirectoryAsync()
    {
        ErrorMessage = null;
        StatusMessage = null;
        try
        {
            await _client.InvokeAsync<object>("logs.openDir");
            StatusMessage = "Log directory opened.";
        }
        catch (Exception ex)
        {
            ErrorMessage = ex.Message;
        }
    }

    private async void OnReload(object sender, RoutedEventArgs e) => await LoadAsync();
    private async void OnSave(object sender, RoutedEventArgs e) => await SaveAsync();
    private async void OnTestProxy(object sender, RoutedEventArgs e) => await TestProxyAsync();
    private async void OnCheckUpdate(object sender, RoutedEventArgs e) => await CheckUpdateAsync();
    private async void OnOpenLogDir(object sender, RoutedEventArgs e) => await OpenLogDirectoryAsync();

    private async void OnChooseRepoCacheDir(object sender, RoutedEventArgs e)
    {
        if (App.MainWindow is null) return;

        var picker = new FolderPicker();
        picker.FileTypeFilter.Add("*");
        picker.SuggestedStartLocation = PickerLocationId.ComputerFolder;

        var hwnd = WinRT.Interop.WindowNative.GetWindowHandle(App.MainWindow);
        WinRT.Interop.InitializeWithWindow.Initialize(picker, hwnd);

        var folder = await picker.PickSingleFolderAsync();
        if (folder != null)
        {
            Draft.RepoCacheDir = folder.Path;
            OnPropertyChanged(nameof(Draft));
        }
    }

    private void RefreshBindings()
    {
        OnPropertyChanged(nameof(Draft));
        OnPropertyChanged(nameof(LogLevelIndex));
        OnPropertyChanged(nameof(ProxyModeIndex));
        OnPropertyChanged(nameof(ManualProxyVisibility));
        OnPropertyChanged(nameof(NoAgentsVisibility));
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
        switch (name)
        {
            case nameof(IsLoading):
                OnPropertyChanged(nameof(LoadingVisibility));
                OnPropertyChanged(nameof(ContentVisibility));
                OnPropertyChanged(nameof(CanInteract));
                OnPropertyChanged(nameof(CanTestProxy));
                OnPropertyChanged(nameof(CanCheckUpdate));
                break;
            case nameof(IsSaving):
                OnPropertyChanged(nameof(SavingVisibility));
                OnPropertyChanged(nameof(CanInteract));
                OnPropertyChanged(nameof(CanTestProxy));
                OnPropertyChanged(nameof(CanCheckUpdate));
                break;
            case nameof(IsTestingProxy):
                OnPropertyChanged(nameof(TestingProxyVisibility));
                OnPropertyChanged(nameof(CanTestProxy));
                break;
            case nameof(IsCheckingUpdate):
                OnPropertyChanged(nameof(CheckingUpdateVisibility));
                OnPropertyChanged(nameof(CanCheckUpdate));
                break;
            case nameof(ErrorMessage):
                OnPropertyChanged(nameof(ErrorVisibility));
                break;
            case nameof(StatusMessage):
                OnPropertyChanged(nameof(StatusVisibility));
                break;
            case nameof(UpdateInfo):
                OnPropertyChanged(nameof(UpdateInfoVisibility));
                OnPropertyChanged(nameof(UpdateStatusText));
                break;
        }
    }
}
