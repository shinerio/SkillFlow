using System.Net;
using System.Text.Json;
using SkillFlow.Daemon;
using Xunit;

namespace SkillFlow.Tests;

public sealed class RecordingHandler : HttpMessageHandler
{
    private readonly Func<HttpRequestMessage, HttpResponseMessage> _handler;

    public RecordingHandler(Func<HttpRequestMessage, HttpResponseMessage> handler)
    {
        _handler = handler;
    }

    public HttpRequestMessage? Request { get; private set; }

    public string? RequestBody { get; private set; }

    protected override async Task<HttpResponseMessage> SendAsync(
        HttpRequestMessage request,
        CancellationToken cancellationToken)
    {
        Request = request;
        RequestBody = request.Content is null
            ? null
            : await request.Content.ReadAsStringAsync(cancellationToken);
        return _handler(request);
    }
}

public sealed class DaemonClientTests
{
    [Fact]
    public async Task InvokeAsyncSendsTokenVersionAndNativeEnvelope()
    {
        var statePath = await WriteEndpointAsync();
        var handler = new RecordingHandler(_ => JsonResponse("""
        {"ok":true,"result":{"ok":true,"result":[{"name":"Codex Skill"}],"error":null}}
        """));
        using var httpClient = new HttpClient(handler);
        var client = new DaemonClient(httpClient, statePath);

        var skills = await client.InvokeAsync<List<SkillFixture>>(
            "skills.list",
            requestId: "request-001");

        Assert.Equal(["Codex Skill"], skills.Select(skill => skill.Name));
        Assert.Equal("http://127.0.0.1:49151/invoke", handler.Request?.RequestUri?.ToString());
        Assert.Equal("POST", handler.Request?.Method.Method);
        Assert.Equal("test-token", handler.Request?.Headers.GetValues("X-SkillFlow-Token").Single());

        using var document = JsonDocument.Parse(handler.RequestBody ?? string.Empty);
        var root = document.RootElement;
        var parameters = root.GetProperty("params");

        Assert.Equal("native.api", root.GetProperty("method").GetString());
        Assert.Equal(DaemonClient.ApiVersion, parameters.GetProperty("version").GetString());
        Assert.Equal("skills.list", parameters.GetProperty("method").GetString());
        Assert.Equal("request-001", parameters.GetProperty("requestID").GetString());
        Assert.Equal(JsonValueKind.Null, parameters.GetProperty("params").ValueKind);
    }

    [Fact]
    public async Task InvokeAsyncMapsMethodNotFoundError()
    {
        var statePath = await WriteEndpointAsync();
        var handler = new RecordingHandler(_ => JsonResponse("""
        {"ok":true,"result":{"ok":false,"result":null,"error":{"code":"method_not_found","message":"method not found","messageKey":"nativeapi.error.method_not_found"}}}
        """));
        using var httpClient = new HttpClient(handler);
        var client = new DaemonClient(httpClient, statePath);

        var exception = await Assert.ThrowsAsync<DaemonClientException>(() =>
            client.InvokeAsync<List<SkillFixture>>("missing.method"));

        Assert.Equal(DaemonClientErrorKind.Api, exception.Kind);
        Assert.Equal("method_not_found", exception.Code);
        Assert.Equal("nativeapi.error.method_not_found", exception.MessageKey);
    }

    [Fact]
    public async Task InvokeAsyncMapsUnauthorizedResponse()
    {
        var statePath = await WriteEndpointAsync();
        var handler = new RecordingHandler(_ => new HttpResponseMessage(HttpStatusCode.Unauthorized));
        using var httpClient = new HttpClient(handler);
        var client = new DaemonClient(httpClient, statePath);

        var exception = await Assert.ThrowsAsync<DaemonClientException>(() =>
            client.InvokeAsync<List<SkillFixture>>("skills.list"));

        Assert.Equal(DaemonClientErrorKind.Unauthorized, exception.Kind);
    }

    [Fact]
    public async Task InvokeAsyncMapsMissingEndpointFile()
    {
        var statePath = Path.Combine(
            Path.GetTempPath(),
            $"skillflow-missing-{Guid.NewGuid():N}.json");
        using var httpClient = new HttpClient(new RecordingHandler(_ => new HttpResponseMessage()));
        var client = new DaemonClient(httpClient, statePath);

        var exception = await Assert.ThrowsAsync<DaemonClientException>(() =>
            client.InvokeAsync<List<SkillFixture>>("skills.list"));

        Assert.Equal(DaemonClientErrorKind.EndpointUnavailable, exception.Kind);
    }

    private static HttpResponseMessage JsonResponse(string json) => new(HttpStatusCode.OK)
    {
        Content = new StringContent(json, System.Text.Encoding.UTF8, "application/json")
    };

    private static async Task<string> WriteEndpointAsync()
    {
        var path = Path.Combine(
            Path.GetTempPath(),
            $"skillflow-daemon-{Guid.NewGuid():N}.json");
        await File.WriteAllTextAsync(path, """
        {"address":"127.0.0.1:49151","token":"test-token","pid":12345}
        """);
        return path;
    }

    private sealed record SkillFixture(string Name);
}
