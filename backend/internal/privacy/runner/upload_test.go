package runner_test

import (
	"context"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/privacy/runner"
	"github.com/monshunter/easyinterview/backend/internal/upload/store"
)

func TestDeleteUploadFilesForUserDelegatesToUploadService(t *testing.T) {
	deleter := &fakeUploadFileDeleter{deleted: []store.DeletedFileObject{
		{ID: "file-1", ObjectKey: "user-1/resume/file-1.pdf", Purpose: store.PurposeResume},
	}}
	r := runner.Runner{UploadFiles: deleter}

	deleted, err := r.DeleteUploadFilesForUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("DeleteUploadFilesForUser: %v", err)
	}
	if deleter.userID != "user-1" || len(deleted) != 1 || deleted[0].ID != "file-1" {
		t.Fatalf("userID=%q deleted=%+v", deleter.userID, deleted)
	}
}

type fakeUploadFileDeleter struct {
	userID  string
	deleted []store.DeletedFileObject
	err     error
}

func (d *fakeUploadFileDeleter) DeleteFileObjectsForUser(ctx context.Context, userID string) ([]store.DeletedFileObject, error) {
	d.userID = userID
	if d.err != nil {
		return nil, d.err
	}
	return d.deleted, nil
}
