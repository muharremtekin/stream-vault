using CatalogService.Domain.Entities;
using CatalogService.Domain.Enums;
using CatalogService.Domain.ValueObjects;
using CatalogService.Infrastructure.Persistence;
using Microsoft.Extensions.Logging;
using MongoDB.Driver;

namespace CatalogService.Infrastructure.Seed;

public class CatalogSeeder
{
    private readonly CatalogDbContext _context;
    private readonly ILogger<CatalogSeeder> _logger;

    public CatalogSeeder(CatalogDbContext context, ILogger<CatalogSeeder> logger)
    {
        _context = context ?? throw new ArgumentNullException(nameof(context));
        _logger = logger ?? throw new ArgumentNullException(nameof(logger));
    }

    public async Task SeedAsync()
    {
        await SeedGenresAsync();
        await SeedMoviesAsync();
        await SeedSeriesAsync();
        await _context.CreateIndexesAsync();

        _logger.LogInformation("Catalog database seeding completed successfully.");
    }

    private async Task SeedGenresAsync()
    {
        var existingCount = await _context.Genres.CountDocumentsAsync(Builders<Genre>.Filter.Empty);
        if (existingCount > 0)
        {
            _logger.LogInformation("Genres already seeded. Skipping.");
            return;
        }

        var genres = new List<Genre>
        {
            new("Action", "action", "High-energy films with intense physical feats and combat."),
            new("Comedy", "comedy", "Films designed to amuse and provoke laughter."),
            new("Drama", "drama", "Character-driven stories emphasizing realistic emotional themes."),
            new("Science Fiction", "sci-fi", "Speculative stories involving futuristic technology and space."),
            new("Horror", "horror", "Films designed to frighten, scare, and invoke dread."),
            new("Thriller", "thriller", "Suspenseful films that keep audiences on the edge of their seats."),
            new("Romance", "romance", "Stories centered around romantic love between characters."),
            new("Documentary", "documentary", "Non-fictional films that document reality for education or record."),
            new("Animation", "animation", "Films made using animation techniques instead of live action."),
            new("Fantasy", "fantasy", "Stories set in fantastical worlds with magical elements."),
            new("Crime", "crime", "Stories centered around criminal activities and investigations."),
            new("Adventure", "adventure", "Exciting stories involving exploration and daring exploits."),
            new("Mystery", "mystery", "Stories focused on solving a puzzle or uncovering secrets."),
            new("Western", "western", "Stories set in the American Old West frontier.")
        };

        await _context.Genres.InsertManyAsync(genres);
        _logger.LogInformation("Seeded {Count} genres.", genres.Count);
    }

    private async Task SeedMoviesAsync()
    {
        var existingCount = await _context.Movies.CountDocumentsAsync(Builders<Movie>.Filter.Empty);
        if (existingCount > 0)
        {
            _logger.LogInformation("Movies already seeded. Skipping.");
            return;
        }

        var movies = new List<Movie>
        {
            new()
            {
                Title = "Interstellar",
                Description = "A team of explorers travel through a wormhole in space in an attempt to ensure humanity's survival.",
                ReleaseYear = 2014,
                Duration = new Duration(2, 49),
                MaturityRating = MaturityRating.PG13,
                Genres = new List<string> { "Science Fiction", "Drama", "Adventure" },
                Cast = new List<CastMember>
                {
                    new("Matthew McConaughey", "Cooper"),
                    new("Anne Hathaway", "Dr. Amelia Brand"),
                    new("Jessica Chastain", "Murphy Cooper")
                },
                Director = "Christopher Nolan",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/interstellar.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/interstellar.jpg",
                TrailerUrl = "https://cdn.streamvault.io/trailers/interstellar.mp4",
                AverageRating = 8.7,
                RatingCount = 1842,
                Status = ContentStatus.Published,
                Tags = new List<string> { "space", "time-travel", "epic", "nolan" }
            },
            new()
            {
                Title = "The Dark Knight",
                Description = "When the menace known as the Joker wreaks havoc and chaos on the people of Gotham, Batman must accept one of the greatest psychological and physical tests of his ability to fight injustice.",
                ReleaseYear = 2008,
                Duration = new Duration(2, 32),
                MaturityRating = MaturityRating.PG13,
                Genres = new List<string> { "Action", "Crime", "Drama" },
                Cast = new List<CastMember>
                {
                    new("Christian Bale", "Bruce Wayne / Batman"),
                    new("Heath Ledger", "The Joker"),
                    new("Aaron Eckhart", "Harvey Dent")
                },
                Director = "Christopher Nolan",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/the-dark-knight.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/the-dark-knight.jpg",
                TrailerUrl = "https://cdn.streamvault.io/trailers/the-dark-knight.mp4",
                AverageRating = 9.0,
                RatingCount = 2534,
                Status = ContentStatus.Published,
                Tags = new List<string> { "batman", "superhero", "gotham", "joker" }
            },
            new()
            {
                Title = "Inception",
                Description = "A thief who steals corporate secrets through the use of dream-sharing technology is given the inverse task of planting an idea into the mind of a C.E.O.",
                ReleaseYear = 2010,
                Duration = new Duration(2, 28),
                MaturityRating = MaturityRating.PG13,
                Genres = new List<string> { "Action", "Science Fiction", "Thriller" },
                Cast = new List<CastMember>
                {
                    new("Leonardo DiCaprio", "Dom Cobb"),
                    new("Joseph Gordon-Levitt", "Arthur"),
                    new("Elliot Page", "Ariadne")
                },
                Director = "Christopher Nolan",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/inception.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/inception.jpg",
                TrailerUrl = "https://cdn.streamvault.io/trailers/inception.mp4",
                AverageRating = 8.8,
                RatingCount = 2201,
                Status = ContentStatus.Published,
                Tags = new List<string> { "dreams", "heist", "mind-bending" }
            },
            new()
            {
                Title = "The Shawshank Redemption",
                Description = "Two imprisoned men bond over a number of years, finding solace and eventual redemption through acts of common decency.",
                ReleaseYear = 1994,
                Duration = new Duration(2, 22),
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Drama" },
                Cast = new List<CastMember>
                {
                    new("Tim Robbins", "Andy Dufresne"),
                    new("Morgan Freeman", "Ellis 'Red' Redding")
                },
                Director = "Frank Darabont",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/shawshank-redemption.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/shawshank-redemption.jpg",
                AverageRating = 9.3,
                RatingCount = 2890,
                Status = ContentStatus.Published,
                Tags = new List<string> { "prison", "hope", "friendship", "classic" }
            },
            new()
            {
                Title = "Pulp Fiction",
                Description = "The lives of two mob hitmen, a boxer, a gangster and his wife, and a pair of diner bandits intertwine in four tales of violence and redemption.",
                ReleaseYear = 1994,
                Duration = new Duration(2, 34),
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Crime", "Drama" },
                Cast = new List<CastMember>
                {
                    new("John Travolta", "Vincent Vega"),
                    new("Uma Thurman", "Mia Wallace"),
                    new("Samuel L. Jackson", "Jules Winnfield")
                },
                Director = "Quentin Tarantino",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/pulp-fiction.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/pulp-fiction.jpg",
                AverageRating = 8.9,
                RatingCount = 2105,
                Status = ContentStatus.Published,
                Tags = new List<string> { "tarantino", "nonlinear", "crime", "classic" }
            },
            new()
            {
                Title = "The Matrix",
                Description = "When a beautiful stranger leads computer hacker Neo to a forbidding underworld, he discovers the shocking truth: the life he knows is the elaborate deception of an evil cyber-intelligence.",
                ReleaseYear = 1999,
                Duration = new Duration(2, 16),
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Action", "Science Fiction" },
                Cast = new List<CastMember>
                {
                    new("Keanu Reeves", "Neo"),
                    new("Laurence Fishburne", "Morpheus"),
                    new("Carrie-Anne Moss", "Trinity")
                },
                Director = "Lana Wachowski, Lilly Wachowski",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/the-matrix.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/the-matrix.jpg",
                TrailerUrl = "https://cdn.streamvault.io/trailers/the-matrix.mp4",
                AverageRating = 8.7,
                RatingCount = 1987,
                Status = ContentStatus.Published,
                Tags = new List<string> { "cyberpunk", "simulation", "kung-fu", "classic" }
            },
            new()
            {
                Title = "Parasite",
                OriginalTitle = "Gisaengchung",
                Description = "Greed and class discrimination threaten the newly formed symbiotic relationship between the wealthy Park family and the destitute Kim clan.",
                ReleaseYear = 2019,
                Duration = new Duration(2, 12),
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Thriller", "Drama", "Comedy" },
                Cast = new List<CastMember>
                {
                    new("Song Kang-ho", "Ki-taek"),
                    new("Lee Sun-kyun", "Park Dong-ik"),
                    new("Cho Yeo-jeong", "Yeon-gyo")
                },
                Director = "Bong Joon-ho",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/parasite.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/parasite.jpg",
                AverageRating = 8.5,
                RatingCount = 1650,
                Status = ContentStatus.Published,
                Tags = new List<string> { "korean", "social-commentary", "oscar-winner" }
            },
            new()
            {
                Title = "The Godfather",
                Description = "The aging patriarch of an organized crime dynasty in postwar New York City transfers control of his clandestine empire to his reluctant youngest son.",
                ReleaseYear = 1972,
                Duration = new Duration(2, 55),
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Crime", "Drama" },
                Cast = new List<CastMember>
                {
                    new("Marlon Brando", "Don Vito Corleone"),
                    new("Al Pacino", "Michael Corleone"),
                    new("James Caan", "Sonny Corleone")
                },
                Director = "Francis Ford Coppola",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/the-godfather.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/the-godfather.jpg",
                AverageRating = 9.2,
                RatingCount = 2750,
                Status = ContentStatus.Published,
                Tags = new List<string> { "mafia", "family", "classic", "italian" }
            },
            new()
            {
                Title = "Spirited Away",
                OriginalTitle = "Sen to Chihiro no Kamikakushi",
                Description = "During her family's move to the suburbs, a sullen 10-year-old girl wanders into a world ruled by gods, witches, and spirits, where humans are changed into beasts.",
                ReleaseYear = 2001,
                Duration = new Duration(2, 5),
                MaturityRating = MaturityRating.PG,
                Genres = new List<string> { "Animation", "Fantasy", "Adventure" },
                Cast = new List<CastMember>
                {
                    new("Rumi Hiiragi", "Chihiro (voice)"),
                    new("Miyu Irino", "Haku (voice)")
                },
                Director = "Hayao Miyazaki",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/spirited-away.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/spirited-away.jpg",
                AverageRating = 8.6,
                RatingCount = 1420,
                Status = ContentStatus.Published,
                Tags = new List<string> { "anime", "studio-ghibli", "japanese", "oscar-winner" }
            },
            new()
            {
                Title = "Mad Max: Fury Road",
                Description = "In a post-apocalyptic wasteland, a woman rebels against a tyrannical ruler in search for her homeland with the aid of a group of female prisoners, a psychotic worshiper, and a drifter named Max.",
                ReleaseYear = 2015,
                Duration = new Duration(2, 0),
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Action", "Adventure", "Science Fiction" },
                Cast = new List<CastMember>
                {
                    new("Tom Hardy", "Max Rockatansky"),
                    new("Charlize Theron", "Imperator Furiosa")
                },
                Director = "George Miller",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/mad-max-fury-road.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/mad-max-fury-road.jpg",
                TrailerUrl = "https://cdn.streamvault.io/trailers/mad-max-fury-road.mp4",
                AverageRating = 8.1,
                RatingCount = 1580,
                Status = ContentStatus.Published,
                Tags = new List<string> { "post-apocalyptic", "car-chase", "desert" }
            },
            new()
            {
                Title = "Whiplash",
                Description = "A promising young drummer enrolls at a cut-throat music conservatory where his dreams of greatness are mentored by an instructor who will stop at nothing to realize a student's potential.",
                ReleaseYear = 2014,
                Duration = new Duration(1, 47),
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Drama" },
                Cast = new List<CastMember>
                {
                    new("Miles Teller", "Andrew Neiman"),
                    new("J.K. Simmons", "Terence Fletcher")
                },
                Director = "Damien Chazelle",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/whiplash.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/whiplash.jpg",
                AverageRating = 8.5,
                RatingCount = 1350,
                Status = ContentStatus.Published,
                Tags = new List<string> { "music", "jazz", "obsession", "mentor" }
            },
            new()
            {
                Title = "Get Out",
                Description = "A young African-American visits his white girlfriend's parents for the weekend, where his simmering uneasiness about their reception of him eventually reaches a boiling point.",
                ReleaseYear = 2017,
                Duration = new Duration(1, 44),
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Horror", "Thriller", "Mystery" },
                Cast = new List<CastMember>
                {
                    new("Daniel Kaluuya", "Chris Washington"),
                    new("Allison Williams", "Rose Armitage")
                },
                Director = "Jordan Peele",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/get-out.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/get-out.jpg",
                AverageRating = 7.7,
                RatingCount = 1280,
                Status = ContentStatus.Published,
                Tags = new List<string> { "social-thriller", "psychological", "oscar-winner" }
            },
            new()
            {
                Title = "Dune",
                Description = "Feature adaptation of Frank Herbert's science fiction novel about the son of a noble family entrusted with the protection of the most valuable asset in the galaxy.",
                ReleaseYear = 2021,
                Duration = new Duration(2, 35),
                MaturityRating = MaturityRating.PG13,
                Genres = new List<string> { "Science Fiction", "Adventure", "Drama" },
                Cast = new List<CastMember>
                {
                    new("Timothee Chalamet", "Paul Atreides"),
                    new("Rebecca Ferguson", "Lady Jessica"),
                    new("Zendaya", "Chani")
                },
                Director = "Denis Villeneuve",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/dune.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/dune.jpg",
                TrailerUrl = "https://cdn.streamvault.io/trailers/dune.mp4",
                AverageRating = 8.0,
                RatingCount = 1620,
                Status = ContentStatus.Published,
                Tags = new List<string> { "desert", "epic", "adaptation", "spice" }
            },
            new()
            {
                Title = "Everything Everywhere All at Once",
                Description = "An aging Chinese immigrant is swept up in an insane adventure, where she alone can save the world by exploring other universes connecting with the lives she could have led.",
                ReleaseYear = 2022,
                Duration = new Duration(2, 19),
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Action", "Comedy", "Science Fiction" },
                Cast = new List<CastMember>
                {
                    new("Michelle Yeoh", "Evelyn Quan Wang"),
                    new("Stephanie Hsu", "Joy Wang"),
                    new("Ke Huy Quan", "Waymond Wang")
                },
                Director = "Daniel Kwan, Daniel Scheinert",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/eeaao.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/eeaao.jpg",
                AverageRating = 7.8,
                RatingCount = 1150,
                Status = ContentStatus.Published,
                Tags = new List<string> { "multiverse", "family", "oscar-winner", "absurdist" }
            },
            new()
            {
                Title = "The Grand Budapest Hotel",
                Description = "A writer encounters the owner of an aging high-class hotel, who tells him of his early years serving as a lobby boy in the hotel's glorious years under an exceptional concierge.",
                ReleaseYear = 2014,
                Duration = new Duration(1, 39),
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Comedy", "Drama", "Adventure" },
                Cast = new List<CastMember>
                {
                    new("Ralph Fiennes", "M. Gustave"),
                    new("Tony Revolori", "Zero Moustafa"),
                    new("Saoirse Ronan", "Agatha")
                },
                Director = "Wes Anderson",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/grand-budapest-hotel.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/grand-budapest-hotel.jpg",
                AverageRating = 8.1,
                RatingCount = 1320,
                Status = ContentStatus.Published,
                Tags = new List<string> { "wes-anderson", "quirky", "hotel", "europe" }
            },
            new()
            {
                Title = "Blade Runner 2049",
                Description = "Young Blade Runner K's discovery of a long-buried secret leads him to track down former Blade Runner Rick Deckard, who's been missing for thirty years.",
                ReleaseYear = 2017,
                Duration = new Duration(2, 44),
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Science Fiction", "Drama", "Thriller" },
                Cast = new List<CastMember>
                {
                    new("Ryan Gosling", "Officer K"),
                    new("Harrison Ford", "Rick Deckard"),
                    new("Ana de Armas", "Joi")
                },
                Director = "Denis Villeneuve",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/blade-runner-2049.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/blade-runner-2049.jpg",
                TrailerUrl = "https://cdn.streamvault.io/trailers/blade-runner-2049.mp4",
                AverageRating = 8.0,
                RatingCount = 1440,
                Status = ContentStatus.Published,
                Tags = new List<string> { "cyberpunk", "dystopian", "replicants", "neo-noir" }
            },
            new()
            {
                Title = "No Country for Old Men",
                Description = "Violence and mayhem ensue after a hunter stumbles upon a drug deal gone wrong and more than two million dollars in cash near the Rio Grande.",
                ReleaseYear = 2007,
                Duration = new Duration(2, 2),
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Crime", "Drama", "Thriller" },
                Cast = new List<CastMember>
                {
                    new("Javier Bardem", "Anton Chigurh"),
                    new("Josh Brolin", "Llewelyn Moss"),
                    new("Tommy Lee Jones", "Sheriff Ed Tom Bell")
                },
                Director = "Joel Coen, Ethan Coen",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/no-country-for-old-men.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/no-country-for-old-men.jpg",
                AverageRating = 8.2,
                RatingCount = 1380,
                Status = ContentStatus.Published,
                Tags = new List<string> { "coen-brothers", "western", "neo-noir", "oscar-winner" }
            },
            new()
            {
                Title = "The Social Network",
                Description = "As Harvard student Mark Zuckerberg creates the social networking site that would become known as Facebook, he is sued by the twins who claimed he stole their idea.",
                ReleaseYear = 2010,
                Duration = new Duration(2, 0),
                MaturityRating = MaturityRating.PG13,
                Genres = new List<string> { "Drama" },
                Cast = new List<CastMember>
                {
                    new("Jesse Eisenberg", "Mark Zuckerberg"),
                    new("Andrew Garfield", "Eduardo Saverin"),
                    new("Justin Timberlake", "Sean Parker")
                },
                Director = "David Fincher",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/the-social-network.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/the-social-network.jpg",
                AverageRating = 7.8,
                RatingCount = 1290,
                Status = ContentStatus.Published,
                Tags = new List<string> { "tech", "facebook", "startup", "fincher" }
            },
            new()
            {
                Title = "Arrival",
                Description = "A linguist works with the military to communicate with alien lifeforms after twelve mysterious spacecraft appear around the world.",
                ReleaseYear = 2016,
                Duration = new Duration(1, 56),
                MaturityRating = MaturityRating.PG13,
                Genres = new List<string> { "Science Fiction", "Drama", "Mystery" },
                Cast = new List<CastMember>
                {
                    new("Amy Adams", "Dr. Louise Banks"),
                    new("Jeremy Renner", "Ian Donnelly"),
                    new("Forest Whitaker", "Colonel Weber")
                },
                Director = "Denis Villeneuve",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/arrival.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/arrival.jpg",
                AverageRating = 7.9,
                RatingCount = 1310,
                Status = ContentStatus.Published,
                Tags = new List<string> { "aliens", "linguistics", "time", "first-contact" }
            },
            new()
            {
                Title = "Oppenheimer",
                Description = "The story of American scientist J. Robert Oppenheimer and his role in the development of the atomic bomb.",
                ReleaseYear = 2023,
                Duration = new Duration(3, 0),
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Drama" },
                Cast = new List<CastMember>
                {
                    new("Cillian Murphy", "J. Robert Oppenheimer"),
                    new("Emily Blunt", "Kitty Oppenheimer"),
                    new("Robert Downey Jr.", "Lewis Strauss")
                },
                Director = "Christopher Nolan",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/oppenheimer.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/oppenheimer.jpg",
                TrailerUrl = "https://cdn.streamvault.io/trailers/oppenheimer.mp4",
                AverageRating = 8.3,
                RatingCount = 1750,
                Status = ContentStatus.Published,
                Tags = new List<string> { "biography", "manhattan-project", "nuclear", "oscar-winner" }
            }
        };

        await _context.Movies.InsertManyAsync(movies);
        _logger.LogInformation("Seeded {Count} movies.", movies.Count);
    }

    private async Task SeedSeriesAsync()
    {
        var existingCount = await _context.Series.CountDocumentsAsync(Builders<Series>.Filter.Empty);
        if (existingCount > 0)
        {
            _logger.LogInformation("Series already seeded. Skipping.");
            return;
        }

        var seriesList = new List<Series>
        {
            new()
            {
                Title = "Breaking Bad",
                Description = "A high school chemistry teacher diagnosed with inoperable lung cancer turns to manufacturing and selling methamphetamine in order to secure his family's future.",
                ReleaseYear = 2008,
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Crime", "Drama", "Thriller" },
                Cast = new List<CastMember>
                {
                    new("Bryan Cranston", "Walter White"),
                    new("Aaron Paul", "Jesse Pinkman"),
                    new("Anna Gunn", "Skyler White")
                },
                Creator = "Vince Gilligan",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/breaking-bad.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/breaking-bad.jpg",
                Status = ContentStatus.Published,
                Seasons = new List<Season>
                {
                    new(1, 2008, "Season 1") { Episodes = new List<Episode>
                    {
                        new(1, "Pilot", "Walter White, a struggling high school chemistry teacher, is diagnosed with advanced lung cancer.", new Duration(0, 58), "https://cdn.streamvault.io/thumbnails/bb-s1e1.jpg"),
                        new(2, "Cat's in the Bag...", "Walt and Jesse attempt to tie up loose ends.", new Duration(0, 48), "https://cdn.streamvault.io/thumbnails/bb-s1e2.jpg"),
                        new(3, "...And the Bag's in the River", "Walt is faced with a difficult decision.", new Duration(0, 48), "https://cdn.streamvault.io/thumbnails/bb-s1e3.jpg")
                    }}
                }
            },
            new()
            {
                Title = "Stranger Things",
                Description = "When a young boy disappears, his mother, a police chief and his friends must confront terrifying supernatural forces in order to get him back.",
                ReleaseYear = 2016,
                MaturityRating = MaturityRating.PG13,
                Genres = new List<string> { "Drama", "Fantasy", "Horror" },
                Cast = new List<CastMember>
                {
                    new("Millie Bobby Brown", "Eleven"),
                    new("Finn Wolfhard", "Mike Wheeler"),
                    new("Winona Ryder", "Joyce Byers")
                },
                Creator = "The Duffer Brothers",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/stranger-things.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/stranger-things.jpg",
                Status = ContentStatus.Published,
                Seasons = new List<Season>
                {
                    new(1, 2016, "The Vanishing of Will Byers") { Episodes = new List<Episode>
                    {
                        new(1, "Chapter One: The Vanishing of Will Byers", "On his way home from a friend's house, young Will sees something terrifying.", new Duration(0, 49), "https://cdn.streamvault.io/thumbnails/st-s1e1.jpg"),
                        new(2, "Chapter Two: The Weirdo on Maple Street", "Lucas, Mike and Dustin try to talk to the girl they found in the woods.", new Duration(0, 56), "https://cdn.streamvault.io/thumbnails/st-s1e2.jpg")
                    }}
                }
            },
            new()
            {
                Title = "Game of Thrones",
                Description = "Nine noble families fight for control over the lands of Westeros, while an ancient enemy returns after being dormant for millennia.",
                ReleaseYear = 2011,
                MaturityRating = MaturityRating.NC17,
                Genres = new List<string> { "Action", "Adventure", "Drama", "Fantasy" },
                Cast = new List<CastMember>
                {
                    new("Emilia Clarke", "Daenerys Targaryen"),
                    new("Kit Harington", "Jon Snow"),
                    new("Peter Dinklage", "Tyrion Lannister")
                },
                Creator = "David Benioff, D.B. Weiss",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/game-of-thrones.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/game-of-thrones.jpg",
                Status = ContentStatus.Published,
                Seasons = new List<Season>
                {
                    new(1, 2011, "Season 1") { Episodes = new List<Episode>
                    {
                        new(1, "Winter Is Coming", "Eddard Stark is torn between his family and an old friend when asked to serve at the side of King Robert Baratheon.", new Duration(1, 2), "https://cdn.streamvault.io/thumbnails/got-s1e1.jpg"),
                        new(2, "The Kingsroad", "While Bran recovers from his fall, Ned takes only his daughters to King's Landing.", new Duration(0, 56), "https://cdn.streamvault.io/thumbnails/got-s1e2.jpg")
                    }}
                }
            },
            new()
            {
                Title = "The Crown",
                Description = "Follows the political rivalries and romance of Queen Elizabeth II's reign and the events that shaped the second half of the twentieth century.",
                ReleaseYear = 2016,
                MaturityRating = MaturityRating.PG13,
                Genres = new List<string> { "Drama" },
                Cast = new List<CastMember>
                {
                    new("Claire Foy", "Queen Elizabeth II"),
                    new("Olivia Colman", "Queen Elizabeth II"),
                    new("Imelda Staunton", "Queen Elizabeth II")
                },
                Creator = "Peter Morgan",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/the-crown.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/the-crown.jpg",
                Status = ContentStatus.Published,
                Seasons = new List<Season>
                {
                    new(1, 2016, "Season 1") { Episodes = new List<Episode>
                    {
                        new(1, "Wolferton Splash", "The ailing King George VI contemplates the future of the monarchy.", new Duration(0, 57), "https://cdn.streamvault.io/thumbnails/crown-s1e1.jpg")
                    }}
                }
            },
            new()
            {
                Title = "The Mandalorian",
                Description = "The travels of a lone bounty hunter in the outer reaches of the galaxy, far from the authority of the New Republic.",
                ReleaseYear = 2019,
                MaturityRating = MaturityRating.PG13,
                Genres = new List<string> { "Action", "Adventure", "Science Fiction" },
                Cast = new List<CastMember>
                {
                    new("Pedro Pascal", "Din Djarin / The Mandalorian"),
                    new("Giancarlo Esposito", "Moff Gideon")
                },
                Creator = "Jon Favreau",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/the-mandalorian.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/the-mandalorian.jpg",
                Status = ContentStatus.Published,
                Seasons = new List<Season>
                {
                    new(1, 2019, "Season 1") { Episodes = new List<Episode>
                    {
                        new(1, "Chapter 1: The Mandalorian", "A lone gunfighter makes his way through the dangerous galaxy.", new Duration(0, 39), "https://cdn.streamvault.io/thumbnails/mando-s1e1.jpg"),
                        new(2, "Chapter 2: The Child", "Target in hand, the Mandalorian must now contend with scavengers.", new Duration(0, 32), "https://cdn.streamvault.io/thumbnails/mando-s1e2.jpg")
                    }}
                }
            },
            new()
            {
                Title = "Chernobyl",
                Description = "In April 1986, an explosion at the Chernobyl nuclear power plant in the Union of Soviet Socialist Republics becomes one of the world's worst man-made catastrophes.",
                ReleaseYear = 2019,
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Drama", "Thriller" },
                Cast = new List<CastMember>
                {
                    new("Jared Harris", "Valery Legasov"),
                    new("Stellan Skarsgard", "Boris Shcherbina"),
                    new("Emily Watson", "Ulana Khomyuk")
                },
                Creator = "Craig Mazin",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/chernobyl.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/chernobyl.jpg",
                Status = ContentStatus.Published,
                Seasons = new List<Season>
                {
                    new(1, 2019, "Miniseries") { Episodes = new List<Episode>
                    {
                        new(1, "1:23:45", "Plant workers and firefighters put their lives on the line to control a catastrophic explosion at a Soviet nuclear power plant.", new Duration(1, 6), "https://cdn.streamvault.io/thumbnails/chernobyl-s1e1.jpg"),
                        new(2, "Please Remain Calm", "With the plant still burning, Legasov and Shcherbina arrive to assess the damage.", new Duration(1, 4), "https://cdn.streamvault.io/thumbnails/chernobyl-s1e2.jpg")
                    }}
                }
            },
            new()
            {
                Title = "The Last of Us",
                Description = "Joel and Ellie, a pair connected through the harshness of the world they live in, must survive in a post-apocalyptic United States.",
                ReleaseYear = 2023,
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Action", "Drama", "Horror" },
                Cast = new List<CastMember>
                {
                    new("Pedro Pascal", "Joel Miller"),
                    new("Bella Ramsey", "Ellie Williams")
                },
                Creator = "Craig Mazin, Neil Druckmann",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/the-last-of-us.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/the-last-of-us.jpg",
                Status = ContentStatus.Published,
                Seasons = new List<Season>
                {
                    new(1, 2023, "Season 1") { Episodes = new List<Episode>
                    {
                        new(1, "When You're Lost in the Darkness", "Joel is tasked with smuggling a teenage girl, Ellie, out of a quarantine zone.", new Duration(1, 21), "https://cdn.streamvault.io/thumbnails/tlou-s1e1.jpg"),
                        new(2, "Infected", "Joel and Tess take Ellie further into the city.", new Duration(0, 53), "https://cdn.streamvault.io/thumbnails/tlou-s1e2.jpg")
                    }}
                }
            },
            new()
            {
                Title = "Dark",
                OriginalTitle = "Dark",
                Description = "A family saga with a supernatural twist, set in a German town where the disappearance of two young children exposes the relationships among four families.",
                ReleaseYear = 2017,
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Crime", "Drama", "Mystery", "Science Fiction" },
                Cast = new List<CastMember>
                {
                    new("Louis Hofmann", "Jonas Kahnwald"),
                    new("Oliver Masucci", "Ulrich Nielsen"),
                    new("Lisa Vicari", "Martha Nielsen")
                },
                Creator = "Baran bo Odar, Jantje Friese",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/dark.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/dark.jpg",
                Status = ContentStatus.Published,
                Seasons = new List<Season>
                {
                    new(1, 2017, "Season 1") { Episodes = new List<Episode>
                    {
                        new(1, "Secrets", "In 2019, a local boy's disappearance stokes fear in the residents of Winden, a small German town with a strange and tragic history.", new Duration(0, 51), "https://cdn.streamvault.io/thumbnails/dark-s1e1.jpg")
                    }}
                }
            },
            new()
            {
                Title = "Arcane",
                Description = "Set in Riot Games' utopian League of Legends universe, Arcane follows the origins of two iconic League champions and the power that will tear them apart.",
                ReleaseYear = 2021,
                MaturityRating = MaturityRating.PG13,
                Genres = new List<string> { "Animation", "Action", "Adventure", "Fantasy" },
                Cast = new List<CastMember>
                {
                    new("Hailee Steinfeld", "Vi (voice)"),
                    new("Ella Purnell", "Jinx (voice)"),
                    new("Kevin Alejandro", "Jayce (voice)")
                },
                Creator = "Christian Linke, Alex Yee",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/arcane.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/arcane.jpg",
                Status = ContentStatus.Published,
                Seasons = new List<Season>
                {
                    new(1, 2021, "Season 1") { Episodes = new List<Episode>
                    {
                        new(1, "Welcome to the Playground", "Orphaned sisters Vi and Powder bring trouble to the streets of the undercity.", new Duration(0, 44), "https://cdn.streamvault.io/thumbnails/arcane-s1e1.jpg"),
                        new(2, "Some Mysteries Are Better Left Unsolved", "Idealistic inventor Jayce attempts to harness magic through science.", new Duration(0, 41), "https://cdn.streamvault.io/thumbnails/arcane-s1e2.jpg")
                    }}
                }
            },
            new()
            {
                Title = "The Bear",
                Description = "A young chef from the fine dining world returns to Chicago to run his family's sandwich shop after a heartbreaking death in his family.",
                ReleaseYear = 2022,
                MaturityRating = MaturityRating.R,
                Genres = new List<string> { "Comedy", "Drama" },
                Cast = new List<CastMember>
                {
                    new("Jeremy Allen White", "Carmen 'Carmy' Berzatto"),
                    new("Ebon Moss-Bachrach", "Richard 'Richie' Jerimovich"),
                    new("Ayo Edebiri", "Sydney Adamu")
                },
                Creator = "Christopher Storer",
                ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/the-bear.jpg",
                BannerUrl = "https://cdn.streamvault.io/banners/the-bear.jpg",
                Status = ContentStatus.Published,
                Seasons = new List<Season>
                {
                    new(1, 2022, "Season 1") { Episodes = new List<Episode>
                    {
                        new(1, "System", "Carmy returns home to run The Original Beef of Chicagoland.", new Duration(0, 30), "https://cdn.streamvault.io/thumbnails/bear-s1e1.jpg"),
                        new(2, "Hands", "Carmy tries to bring order to the chaos of The Beef.", new Duration(0, 32), "https://cdn.streamvault.io/thumbnails/bear-s1e2.jpg")
                    }}
                }
            }
        };

        await _context.Series.InsertManyAsync(seriesList);
        _logger.LogInformation("Seeded {Count} series.", seriesList.Count);
    }
}
