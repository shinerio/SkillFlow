using System.Net.Http.Json;
using System.Text.Json;

namespace SkillFlow.Daemon;

public sealed class DaemonClient
{
    public const string ApiVersion = "2026-04-25";

    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
        PropertyNameCaseInsensitive = true
    };

    private readonly string _endpointPath;
    private readonly HttpClient _httpClient;

    public DaemonClient(HttpClient? httpClient = null, string? endpointPath = null)
    {
        _endpointPath = endpointPath ?? DefaultEndpointPath();
        _httpClient = httpClient ?? new HttpClient();
    }

    public static string DefaultEndpointPath()
    {
        var userProfile = Environment.GetFolderPath(Environment.SpecialFolder.UserProfile);
        return Path.Combine(userProfile, ".skillflow", "runtime", "daemon-service.json");
    }

    public async Task<T> InvokeAsync<T>(
        string method,
        object? parameters = null,
        string? requestId = null,
        CancellationToken cancellationToken = default)
    {
        var endpoint = LoadEndpoint();
        var url = new Uri($"http://{endpoint.Address}/invoke");
        var nativeRequest = new NativeApiRequest(
            ApiVersion,
            method,
            parameters,
            requestId ?? $"win-{Guid.NewGuid():N}");

        var serviceRequest = new DaemonServiceRequest("native.api", nativeRequest);
        using var content = JsonContent.Create(serviceRequest, options: JsonOptions);
        using var request = new HttpRequestMessage(HttpMethod.Post, url)
        {
            Content = content
        };
        request.Headers.Add("X-SkillFlow-Token", endpoint.Token);

        HttpResponseMessage response;
        try
        {
            response = await _httpClient.SendAsync(request, cancellationToken);
        }
        catch (Exception exception) when (exception is HttpRequestException or TaskCanceledException)
        {
            throw new DaemonClientException(
                DaemonClientErrorKind.Transport,
                $"SkillFlow daemon request failed: {exception.Message}");
        }

        await using var disposableResponse = response;
        if ((int)response.StatusCode == 401)
        {
            throw new DaemonClientException(
                DaemonClientErrorKind.Unauthorized,
                "SkillFlow daemon rejected the access token.");
        }

        if (!response.IsSuccessStatusCode)
        {
            throw new DaemonClientException(
                DaemonClientErrorKind.Transport,
                $"SkillFlow daemon returned HTTP {(int)response.StatusCode}.");
        }

        DaemonServiceResponse<T>? serviceResult;
        try
        {
            serviceResult = await response.Content.ReadFromJsonAsync<DaemonServiceResponse<T>>(
                JsonOptions,
                cancellationToken);
        }
        catch (JsonException exception)
        {
            throw new DaemonClientException(
                DaemonClientErrorKind.InvalidResponse,
                $"SkillFlow daemon returned an invalid response: {exception.Message}");
        }

        if (serviceResult is null)
        {
            throw new DaemonClientException(
                DaemonClientErrorKind.InvalidResponse,
                "SkillFlow daemon returned an empty response.");
        }

        if (!serviceResult.Ok)
        {
            throw new DaemonClientException(
                DaemonClientErrorKind.Service,
                serviceResult.Error ?? "SkillFlow daemon service failed.");
        }

        if (serviceResult.Result is null)
        {
            throw new DaemonClientException(
                DaemonClientErrorKind.InvalidResponse,
                "SkillFlow daemon response did not include the native API result.");
        }

        if (!serviceResult.Result.Ok)
        {
            var error = serviceResult.Result.Error;
            throw new DaemonClientException(
                DaemonClientErrorKind.Api,
                error?.Message ?? "The native API request failed.",
                error?.Code ?? "internal_error",
                error?.MessageKey);
        }

        if (serviceResult.Result.Result is null)
        {
            return default!;
        }

        return serviceResult.Result.Result;
    }

    private DaemonEndpoint LoadEndpoint()
    {
        if (!File.Exists(_endpointPath))
        {
            throw new DaemonClientException(
                DaemonClientErrorKind.EndpointUnavailable,
                "SkillFlow daemon endpoint is unavailable.");
        }

        try
        {
            var endpoint = JsonSerializer.Deserialize<DaemonEndpoint>(
                File.ReadAllText(_endpointPath),
                JsonOptions);

            if (endpoint is null ||
                string.IsNullOrWhiteSpace(endpoint.Address) ||
                string.IsNullOrWhiteSpace(endpoint.Token))
            {
                throw new DaemonClientException(
                    DaemonClientErrorKind.InvalidEndpoint,
                    "SkillFlow daemon endpoint is invalid.");
            }

            return endpoint;
        }
        catch (JsonException exception)
        {
            throw new DaemonClientException(
                DaemonClientErrorKind.InvalidEndpoint,
                $"SkillFlow daemon endpoint is invalid: {exception.Message}");
        }
    }
}
