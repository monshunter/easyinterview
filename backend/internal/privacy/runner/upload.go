package runner

import (
	"context"
	"fmt"

	"github.com/monshunter/easyinterview/backend/internal/upload/store"
)

type UploadFileDeleter interface {
	DeleteFileObjectsForUser(ctx context.Context, userID string) ([]store.DeletedFileObject, error)
}

type Runner struct {
	UploadFiles UploadFileDeleter
}

func (r Runner) DeleteUploadFilesForUser(ctx context.Context, userID string) ([]store.DeletedFileObject, error) {
	if r.UploadFiles == nil {
		return nil, fmt.Errorf("privacy upload file deleter is not configured")
	}
	return r.UploadFiles.DeleteFileObjectsForUser(ctx, userID)
}
