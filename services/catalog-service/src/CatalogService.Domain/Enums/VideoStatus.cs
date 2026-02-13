namespace CatalogService.Domain.Enums;

public enum VideoStatus
{
    NotUploaded = 0,
    Uploading = 1,
    Queued = 2,
    Encoding = 3,
    Ready = 4,
    Error = 5
}
