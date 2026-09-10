// StreamVault Catalog Database Initialization

db = db.getSiblingDB(process.env.MONGO_INITDB_DATABASE || 'streamvault_catalog');

// Create collections
db.createCollection('movies');
db.createCollection('series');
db.createCollection('genres');

// Movies indexes
db.movies.createIndex({ genres: 1 });
db.movies.createIndex({ releaseYear: -1 });
db.movies.createIndex({ status: 1, averageRating: -1 });
db.movies.createIndex({ title: "text", description: "text" });

// Series indexes
db.series.createIndex({ genres: 1 });
db.series.createIndex({ status: 1 });

// Genres indexes
db.genres.createIndex({ slug: 1 }, { unique: true });

print('StreamVault catalog database initialized successfully.');
