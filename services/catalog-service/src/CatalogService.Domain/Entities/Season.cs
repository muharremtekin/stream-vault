namespace CatalogService.Domain.Entities;

public class Season
{
    public int SeasonNumber { get; set; }

    public string? Title { get; set; }

    public int ReleaseYear { get; set; }

    public List<Episode> Episodes { get; set; } = new();

    public Season() { }

    public Season(int seasonNumber, int releaseYear, string? title = null)
    {
        SeasonNumber = seasonNumber;
        ReleaseYear = releaseYear;
        Title = title;
    }

    public void AddEpisode(Episode episode)
    {
        if (Episodes.Any(e => e.EpisodeNumber == episode.EpisodeNumber))
            throw new InvalidOperationException(
                $"Episode {episode.EpisodeNumber} already exists in Season {SeasonNumber}.");

        Episodes.Add(episode);
    }
}
