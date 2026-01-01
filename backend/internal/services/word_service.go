package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslav/elias/internal/models"
)

type WordService struct {
	pool *pgxpool.Pool
}

func NewWordService(pool *pgxpool.Pool) *WordService {
	return &WordService{pool: pool}
}

func (s *WordService) GetRandomWord(ctx context.Context, roomID uuid.UUID, lang string, category string) (*models.Word, error) {
	var word models.Word
	err := s.pool.QueryRow(ctx, `
		SELECT id, word, lang, category FROM words
		WHERE lang = $1
		AND category = $2
		AND id NOT IN (
			SELECT word_id FROM round_words WHERE room_id = $3
		)
		ORDER BY RANDOM()
		LIMIT 1
	`, lang, category, roomID).Scan(&word.ID, &word.Word, &word.Lang, &word.Category)
	if err != nil {
		return nil, err
	}
	return &word, nil
}

func (s *WordService) RecordWordUsed(ctx context.Context, roomID uuid.UUID, wordID int, roundNum int, guessed bool) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO round_words (room_id, word_id, round_num, guessed)
		VALUES ($1, $2, $3, $4)
	`, roomID, wordID, roundNum, guessed)
	return err
}

func (s *WordService) UpdateWordResult(ctx context.Context, roomID uuid.UUID, wordID int, guessed bool) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE round_words SET guessed = $1
		WHERE room_id = $2 AND word_id = $3
	`, guessed, roomID, wordID)
	return err
}

func (s *WordService) GetRoundStats(ctx context.Context, roomID uuid.UUID) ([]*models.RoundStats, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT round_num,
			   COUNT(*) FILTER (WHERE guessed = TRUE) as guessed,
			   COUNT(*) FILTER (WHERE guessed = FALSE) as missed
		FROM round_words
		WHERE room_id = $1
		GROUP BY round_num
		ORDER BY round_num
	`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.RoundStats
	for rows.Next() {
		var s models.RoundStats
		if err := rows.Scan(&s.RoundNum, &s.WordsGuessed, &s.WordsMissed); err != nil {
			return nil, err
		}
		stats = append(stats, &s)
	}
	return stats, nil
}

func (s *WordService) SeedWords(ctx context.Context, words []string, lang string) error {
	for _, word := range words {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO words (word, lang) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, word, lang)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *WordService) GetWordCount(ctx context.Context, lang string) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM words WHERE lang = $1
	`, lang).Scan(&count)
	return count, err
}
