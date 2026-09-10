using AutoMapper;
using CatalogService.Api.Controllers;
using CatalogService.Application.Interfaces;
using MediatR;
using Microsoft.AspNetCore.Mvc;
using Moq;
using Xunit;

namespace CatalogService.UnitTests.Controllers;

public class CatalogIdValidationTests
{
    private const string EpisodeContentId = "507f1f77bcf86cd799439011_s1_e1";

    [Fact]
    public async Task GetMovieById_WithEpisodeContentId_ReturnsBadRequestWithoutQueryingMongo()
    {
        var repository = new Mock<IMovieRepository>();
        var controller = new MoviesController(
            Mock.Of<IMediator>(),
            repository.Object,
            Mock.Of<IMapper>());

        var result = await controller.GetMovieById(EpisodeContentId);

        Assert.IsType<BadRequestObjectResult>(result);
        repository.Verify(
            item => item.GetByIdAsync(It.IsAny<string>(), It.IsAny<CancellationToken>()),
            Times.Never);
    }

    [Fact]
    public async Task GetMovieStreamingInfo_WithEpisodeContentId_ReturnsBadRequestWithoutDispatchingQuery()
    {
        var mediator = new Mock<IMediator>();
        var controller = new MoviesController(
            mediator.Object,
            Mock.Of<IMovieRepository>(),
            Mock.Of<IMapper>());

        var result = await controller.GetStreamingInfo(EpisodeContentId);

        Assert.IsType<BadRequestObjectResult>(result);
        mediator.Verify(
            item => item.Send(It.IsAny<object>(), It.IsAny<CancellationToken>()),
            Times.Never);
    }

    [Fact]
    public async Task GetSeriesById_WithEpisodeContentId_ReturnsBadRequestWithoutQueryingMongo()
    {
        var repository = new Mock<ISeriesRepository>();
        var controller = new SeriesController(
            Mock.Of<IMediator>(),
            repository.Object,
            Mock.Of<IMapper>());

        var result = await controller.GetSeriesById(EpisodeContentId);

        Assert.IsType<BadRequestObjectResult>(result);
        repository.Verify(
            item => item.GetByIdAsync(It.IsAny<string>(), It.IsAny<CancellationToken>()),
            Times.Never);
    }
}
