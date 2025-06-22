// internal/repository/media_repo.go
package repository

import (
	"fmt"
	"lendral3n/ordering-system/internal/database"
	"lendral3n/ordering-system/internal/models"
)

type mediaRepository struct {
	db *database.DB
}

func NewMediaRepository(db *database.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(media *models.MediaFile) error {
	query := `
		INSERT INTO media_files (
			file_type, file_url, thumbnail_url, file_size,
			mime_type, menu_item_id, uploaded_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		query,
		media.FileType,
		media.FileURL,
		media.ThumbnailURL,
		media.FileSize,
		media.MimeType,
		media.MenuItemID,
		media.UploadedBy,
	).Scan(&media.ID, &media.CreatedAt)
}

func (r *mediaRepository) GetByMenuItemID(menuItemID int) ([]models.MediaFile, error) {
	query := `
		SELECT 
			m.id, m.file_type, m.file_url, m.thumbnail_url,
			m.file_size, m.mime_type, m.menu_item_id, m.uploaded_by,
			m.created_at,
			s.id, s.username, s.email, s.full_name, s.role,
			s.is_active, s.created_at
		FROM media_files m
		LEFT JOIN staff s ON m.uploaded_by = s.id
		WHERE m.menu_item_id = $1
		ORDER BY m.created_at DESC
	`

	rows, err := r.db.Query(query, menuItemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaFiles []models.MediaFile
	for rows.Next() {
		var media models.MediaFile
		var staff models.Staff

		err := rows.Scan(
			&media.ID,
			&media.FileType,
			&media.FileURL,
			&media.ThumbnailURL,
			&media.FileSize,
			&media.MimeType,
			&media.MenuItemID,
			&media.UploadedBy,
			&media.CreatedAt,
			&staff.ID,
			&staff.Username,
			&staff.Email,
			&staff.FullName,
			&staff.Role,
			&staff.IsActive,
			&staff.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		media.Staff = &staff
		mediaFiles = append(mediaFiles, media)
	}

	return mediaFiles, rows.Err()
}

func (r *mediaRepository) Delete(id int) error {
	query := `DELETE FROM media_files WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("media file not found")
	}

	return nil
}
