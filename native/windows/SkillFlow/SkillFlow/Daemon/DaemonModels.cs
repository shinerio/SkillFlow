using System.Text.Json.Serialization;

namespace SkillFlow.Daemon;

public sealed record DaemonEndpoint(
    [property: JsonPropertyName("address")] string Address,
    [property: JsonPropertyName("token")] string Token,
    [property: JsonPropertyName("pid")] int Pid);

public sealed record NativeApiRequest(
    [property: JsonPropertyName("version")] string Version,
    [property: JsonPropertyName("method")] string Method,
    [property: JsonPropertyName("params")] object? Parameters,
    [property: JsonPropertyName("requestID")] string RequestId);

public sealed record DaemonServiceRequest(
    [property: JsonPropertyName("method")] string Method,
    [property: JsonPropertyName("params")] NativeApiRequest Parameters);

public sealed record NativeApiError(
    [property: JsonPropertyName("code")] string Code,
    [property: JsonPropertyName("message")] string? Message,
    [property: JsonPropertyName("messageKey")] string? MessageKey);

public sealed record NativeApiResponse<T>(
    [property: JsonPropertyName("ok")] bool Ok,
    [property: JsonPropertyName("result")] T? Result,
    [property: JsonPropertyName("error")] NativeApiError? Error);

public sealed record DaemonServiceResponse<T>(
    [property: JsonPropertyName("ok")] bool Ok,
    [property: JsonPropertyName("result")] NativeApiResponse<T>? Result,
    [property: JsonPropertyName("error")] string? Error);

public enum DaemonClientErrorKind
{
    EndpointUnavailable,
    InvalidEndpoint,
    InvalidResponse,
    Transport,
    Unauthorized,
    Service,
    Api,
    MissingResult
}

public sealed class DaemonClientException : Exception
{
    public DaemonClientException(
        DaemonClientErrorKind kind,
        string message,
        string? code = null,
        string? messageKey = null) : base(message)
    {
        Kind = kind;
        Code = code;
        MessageKey = messageKey;
    }

    public DaemonClientErrorKind Kind { get; }

    public string? Code { get; }

    public string? MessageKey { get; }
}
