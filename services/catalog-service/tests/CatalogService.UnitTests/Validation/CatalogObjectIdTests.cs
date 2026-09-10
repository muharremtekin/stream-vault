using CatalogService.Application.Validation;
using Xunit;

namespace CatalogService.UnitTests.Validation;

public class CatalogObjectIdTests
{
    [Theory]
    [InlineData("507f1f77bcf86cd799439011")]
    [InlineData("507F1F77BCF86CD799439011")]
    public void IsValid_WithObjectId_ReturnsTrue(string value)
    {
        Assert.True(CatalogObjectId.IsValid(value));
    }

    [Theory]
    [InlineData(null)]
    [InlineData("")]
    [InlineData("507f1f77bcf86cd79943901")]
    [InlineData("507f1f77bcf86cd7994390110")]
    [InlineData("507f1f77bcf86cd79943901g")]
    [InlineData("507f1f77bcf86cd799439011_s1_e1")]
    public void IsValid_WithMalformedValue_ReturnsFalse(string? value)
    {
        Assert.False(CatalogObjectId.IsValid(value));
    }
}
