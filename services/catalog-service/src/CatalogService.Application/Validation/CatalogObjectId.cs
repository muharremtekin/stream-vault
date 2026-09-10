namespace CatalogService.Application.Validation;

public static class CatalogObjectId
{
    private const int ObjectIdLength = 24;

    public static bool IsValid(string? value)
    {
        if (value is null || value.Length != ObjectIdLength)
            return false;

        foreach (var character in value)
        {
            var isDigit = character is >= '0' and <= '9';
            var isLowercaseHex = character is >= 'a' and <= 'f';
            var isUppercaseHex = character is >= 'A' and <= 'F';

            if (!isDigit && !isLowercaseHex && !isUppercaseHex)
                return false;
        }

        return true;
    }
}
