package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/streamvault/recommendation-service/internal/model"
)

// Repository provides CRUD operations on the recommendation database tables.
type Repository struct {
	pool *pgxpool.Pool
}

// New creates a new Repository.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// --- Interactions ---

// CreateInteraction inserts a new interaction record.
func (r *Repository) CreateInteraction(ctx context.Context, interaction *model.Interaction) error {
	query := `
		INSERT INTO interactions (user_id, content_id, interaction_type, rating, completion_pct, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	if interaction.CreatedAt.IsZero() {
		interaction.CreatedAt = time.Now()
	}

	return r.pool.QueryRow(ctx, query,
		interaction.UserID,
		interaction.ContentID,
		interaction.InteractionType,
		interaction.Rating,
		interaction.CompletionPct,
		interaction.CreatedAt,
	).Scan(&interaction.ID)
}

// GetUserInteractions returns the most recent interactions for a user.
func (r *Repository) GetUserInteractions(ctx context.Context, userID string, limit int) ([]model.Interaction, error) {
	query := `
		SELECT id, user_id, content_id, interaction_type, rating, completion_pct, created_at
		FROM interactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("querying user interactions: %w", err)
	}
	defer rows.Close()

	return scanInteractions(rows)
}

// GetUserContentInteractions returns interactions between a specific user and content item.
func (r *Repository) GetUserContentInteractions(ctx context.Context, userID, contentID string) ([]model.Interaction, error) {
	query := `
		SELECT id, user_id, content_id, interaction_type, rating, completion_pct, created_at
		FROM interactions
		WHERE user_id = $1 AND content_id = $2
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID, contentID)
	if err != nil {
		return nil, fmt.Errorf("querying user-content interactions: %w", err)
	}
	defer rows.Close()

	return scanInteractions(rows)
}

// CountUserInteractions returns the total number of interactions for a user.
func (r *Repository) CountUserInteractions(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM interactions WHERE user_id = $1`, userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting user interactions: %w", err)
	}
	return count, nil
}

// GetAllInteractions returns all interactions (for building the collaborative filtering matrix).
func (r *Repository) GetAllInteractions(ctx context.Context) ([]model.Interaction, error) {
	query := `
		SELECT id, user_id, content_id, interaction_type, rating, completion_pct, created_at
		FROM interactions
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying all interactions: %w", err)
	}
	defer rows.Close()

	return scanInteractions(rows)
}

// GetContentViewCounts returns the number of view/complete interactions per content ID.
func (r *Repository) GetContentViewCounts(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT content_id, COUNT(*) as cnt
		FROM interactions
		WHERE interaction_type IN ('view', 'complete')
		GROUP BY content_id`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying content view counts: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var contentID string
		var cnt int
		if err := rows.Scan(&contentID, &cnt); err != nil {
			return nil, fmt.Errorf("scanning view count: %w", err)
		}
		result[contentID] = cnt
	}
	return result, rows.Err()
}

// GetRecentViewCounts returns view counts per content for interactions since the given time.
func (r *Repository) GetRecentViewCounts(ctx context.Context, since time.Time) (map[string]int, error) {
	query := `
		SELECT content_id, COUNT(*) as cnt
		FROM interactions
		WHERE interaction_type IN ('view', 'complete')
		  AND created_at >= $1
		GROUP BY content_id
		ORDER BY cnt DESC`

	rows, err := r.pool.Query(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("querying recent view counts: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var contentID string
		var cnt int
		if err := rows.Scan(&contentID, &cnt); err != nil {
			return nil, fmt.Errorf("scanning recent view count: %w", err)
		}
		result[contentID] = cnt
	}
	return result, rows.Err()
}

// GetDistinctUserIDs returns all unique user IDs that have at least one interaction.
func (r *Repository) GetDistinctUserIDs(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT user_id FROM interactions`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying distinct user IDs: %w", err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scanning user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, rows.Err()
}

func scanInteractions(rows pgx.Rows) ([]model.Interaction, error) {
	var interactions []model.Interaction
	for rows.Next() {
		var i model.Interaction
		if err := rows.Scan(
			&i.ID, &i.UserID, &i.ContentID, &i.InteractionType,
			&i.Rating, &i.CompletionPct, &i.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning interaction: %w", err)
		}
		interactions = append(interactions, i)
	}
	return interactions, rows.Err()
}

// --- User Profiles ---

// UpsertUserProfile creates or updates a user profile.
func (r *Repository) UpsertUserProfile(ctx context.Context, profile *model.UserProfile) error {
	genreWeightsJSON, err := json.Marshal(profile.GenreWeights)
	if err != nil {
		return fmt.Errorf("marshalling genre weights: %w", err)
	}

	query := `
		INSERT INTO user_profiles (user_id, genre_weights, avg_rating, total_interactions, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			genre_weights = EXCLUDED.genre_weights,
			avg_rating = EXCLUDED.avg_rating,
			total_interactions = EXCLUDED.total_interactions,
			updated_at = EXCLUDED.updated_at`

	_, err = r.pool.Exec(ctx, query,
		profile.UserID,
		genreWeightsJSON,
		profile.AvgRating,
		profile.TotalInteractions,
		time.Now(),
	)
	return err
}

// GetUserProfile retrieves a user profile by user ID. Returns nil if not found.
func (r *Repository) GetUserProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	query := `
		SELECT user_id, genre_weights, avg_rating, total_interactions, updated_at
		FROM user_profiles
		WHERE user_id = $1`

	var p model.UserProfile
	var genreWeightsJSON []byte
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&p.UserID, &genreWeightsJSON, &p.AvgRating, &p.TotalInteractions, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying user profile: %w", err)
	}

	if err := json.Unmarshal(genreWeightsJSON, &p.GenreWeights); err != nil {
		return nil, fmt.Errorf("unmarshalling genre weights: %w", err)
	}
	return &p, nil
}

// --- Content Features ---

// UpsertContentFeatures creates or updates content features.
func (r *Repository) UpsertContentFeatures(ctx context.Context, features *model.ContentFeatures) error {
	featureVectorJSON, err := json.Marshal(features.FeatureVector)
	if err != nil {
		return fmt.Errorf("marshalling feature vector: %w", err)
	}

	query := `
		INSERT INTO content_features (content_id, title, genres, tags, director, release_year, avg_rating, feature_vector, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (content_id) DO UPDATE SET
			title = EXCLUDED.title,
			genres = EXCLUDED.genres,
			tags = EXCLUDED.tags,
			director = EXCLUDED.director,
			release_year = EXCLUDED.release_year,
			avg_rating = EXCLUDED.avg_rating,
			feature_vector = EXCLUDED.feature_vector,
			updated_at = EXCLUDED.updated_at`

	_, err = r.pool.Exec(ctx, query,
		features.ContentID,
		features.Title,
		features.Genres,
		features.Tags,
		features.Director,
		features.ReleaseYear,
		features.AvgRating,
		featureVectorJSON,
		time.Now(),
	)
	return err
}

// GetContentFeatures retrieves content features by content ID. Returns nil if not found.
func (r *Repository) GetContentFeatures(ctx context.Context, contentID string) (*model.ContentFeatures, error) {
	query := `
		SELECT content_id, title, genres, tags, director, release_year, avg_rating, feature_vector, updated_at
		FROM content_features
		WHERE content_id = $1`

	var cf model.ContentFeatures
	var featureVectorJSON []byte
	err := r.pool.QueryRow(ctx, query, contentID).Scan(
		&cf.ContentID, &cf.Title, &cf.Genres, &cf.Tags,
		&cf.Director, &cf.ReleaseYear, &cf.AvgRating,
		&featureVectorJSON, &cf.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying content features: %w", err)
	}

	if err := json.Unmarshal(featureVectorJSON, &cf.FeatureVector); err != nil {
		return nil, fmt.Errorf("unmarshalling feature vector: %w", err)
	}
	return &cf, nil
}

// GetAllContentFeatures returns all content features (for batch computation).
func (r *Repository) GetAllContentFeatures(ctx context.Context) ([]model.ContentFeatures, error) {
	query := `
		SELECT content_id, title, genres, tags, director, release_year, avg_rating, feature_vector, updated_at
		FROM content_features`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying all content features: %w", err)
	}
	defer rows.Close()

	return scanContentFeatures(rows)
}

// GetContentFeaturesByIDs returns content features for a set of content IDs.
func (r *Repository) GetContentFeaturesByIDs(ctx context.Context, contentIDs []string) ([]model.ContentFeatures, error) {
	query := `
		SELECT content_id, title, genres, tags, director, release_year, avg_rating, feature_vector, updated_at
		FROM content_features
		WHERE content_id = ANY($1)`

	rows, err := r.pool.Query(ctx, query, contentIDs)
	if err != nil {
		return nil, fmt.Errorf("querying content features by IDs: %w", err)
	}
	defer rows.Close()

	return scanContentFeatures(rows)
}

func scanContentFeatures(rows pgx.Rows) ([]model.ContentFeatures, error) {
	var result []model.ContentFeatures
	for rows.Next() {
		var cf model.ContentFeatures
		var featureVectorJSON []byte
		if err := rows.Scan(
			&cf.ContentID, &cf.Title, &cf.Genres, &cf.Tags,
			&cf.Director, &cf.ReleaseYear, &cf.AvgRating,
			&featureVectorJSON, &cf.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning content features: %w", err)
		}
		if err := json.Unmarshal(featureVectorJSON, &cf.FeatureVector); err != nil {
			return nil, fmt.Errorf("unmarshalling feature vector: %w", err)
		}
		result = append(result, cf)
	}
	return result, rows.Err()
}

// --- Content Similarity ---

// UpsertContentSimilarity creates or updates a similarity score.
func (r *Repository) UpsertContentSimilarity(ctx context.Context, sim *model.ContentSimilarity) error {
	query := `
		INSERT INTO content_similarity (content_id_a, content_id_b, similarity_score, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (content_id_a, content_id_b) DO UPDATE SET
			similarity_score = EXCLUDED.similarity_score,
			updated_at = EXCLUDED.updated_at`

	_, err := r.pool.Exec(ctx, query,
		sim.ContentIDA, sim.ContentIDB, sim.SimilarityScore, time.Now(),
	)
	return err
}

// BatchUpsertContentSimilarity inserts/updates multiple similarity scores in a batch.
func (r *Repository) BatchUpsertContentSimilarity(ctx context.Context, sims []model.ContentSimilarity) error {
	if len(sims) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO content_similarity (content_id_a, content_id_b, similarity_score, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (content_id_a, content_id_b) DO UPDATE SET
			similarity_score = EXCLUDED.similarity_score,
			updated_at = EXCLUDED.updated_at`

	now := time.Now()
	for _, sim := range sims {
		batch.Queue(query, sim.ContentIDA, sim.ContentIDB, sim.SimilarityScore, now)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range sims {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch upsert content similarity: %w", err)
		}
	}
	return nil
}

// GetSimilarContent returns content similar to the given content ID, ordered by score descending.
func (r *Repository) GetSimilarContent(ctx context.Context, contentID string, limit int) ([]model.ContentSimilarity, error) {
	query := `
		SELECT content_id_a, content_id_b, similarity_score, updated_at
		FROM content_similarity
		WHERE content_id_a = $1 OR content_id_b = $1
		ORDER BY similarity_score DESC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, query, contentID, limit)
	if err != nil {
		return nil, fmt.Errorf("querying similar content: %w", err)
	}
	defer rows.Close()

	var result []model.ContentSimilarity
	for rows.Next() {
		var cs model.ContentSimilarity
		if err := rows.Scan(&cs.ContentIDA, &cs.ContentIDB, &cs.SimilarityScore, &cs.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning content similarity: %w", err)
		}
		result = append(result, cs)
	}
	return result, rows.Err()
}
