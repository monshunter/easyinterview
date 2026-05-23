//go:build integration

package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	"github.com/monshunter/easyinterview/backend/internal/upload/objectstore"
	uploadservice "github.com/monshunter/easyinterview/backend/internal/upload/service"
	uploadstore "github.com/monshunter/easyinterview/backend/internal/upload/store"
)

func TestUploadPresignRegisterPrivacyDeleteLiveRoundtrip(t *testing.T) {
	live := requireUploadLiveConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", live.databaseURL)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	ensureMinIOBucket(t, ctx, live.minio)

	const (
		challengeID      = "018f2a40-0000-7000-9000-00000000a001"
		userID           = "018f2a40-0000-7000-9000-00000000a002"
		sessionID        = "018f2a40-0000-7000-9000-00000000a003"
		privacyRequestID = "018f2a40-0000-7000-9000-00000000a004"
		privacyJobID     = "018f2a40-0000-7000-9000-00000000a005"
		email            = "upload-roundtrip-p0-033@example.test"
		sessionToken     = "upload-roundtrip-session-token"
	)
	cleanupUploadRoundtripRows(t, db, userID, email, challengeID, sessionID, privacyRequestID, privacyJobID)

	loader := uploadRoundtripConfig(t, live)
	uploadRoutes, err := buildUploadRoutes(loader, db)
	if err != nil {
		t.Fatalf("buildUploadRoutes: %v", err)
	}
	authSink := auth.NewDevMailSink(auth.DevMailSinkOptions{VerifyBaseURL: "http://api.test/api/v1/auth/email/verify"})
	authService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:                 auth.NewSQLStore(db),
		Dispatcher:            auth.NewImmediateMailDispatcher(authSink),
		DeliverySecrets:       authSink,
		TokenGenerator:        apiFixedTokenGenerator("upload-roundtrip-magic-token"),
		SessionTokenGenerator: apiFixedTokenGenerator(sessionToken),
		ChallengePepper:       "upload-roundtrip-challenge-pepper",
		SessionCookieSecret:   "upload-roundtrip-session-secret",
		Now:                   func() time.Time { return time.Date(2026, 5, 13, 9, 30, 0, 0, time.UTC) },
		NewID:                 apiFixedIDs(challengeID, userID, sessionID, privacyRequestID, privacyJobID),
	})
	runtime, err := buildTargetJobRuntime(loader, db, slog.New(slog.NewTextHandler(io.Discard, nil)), uploadRoutes.Service, nil)
	if err != nil {
		t.Fatalf("buildTargetJobRuntime: %v", err)
	}
	defer runtime.Close()
	kernel := newTestKernel(runner.NewSQLStore(db), runtime.Handlers)
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second)
		defer shutdownCancel()
		if err := kernel.Shutdown(shutdownCtx); err != nil {
			t.Fatalf("runtime shutdown: %v", err)
		}
	}()
	handler := buildAPIHandlerWithUploadAndHandlers(loader, apiRuntimeFlags{}, authService, runtime.Handler, practiceRoutes{}, uploadRoutes, resumeRoutes{})

	start := httptest.NewRecorder()
	handler.ServeHTTP(start, httptest.NewRequest(http.MethodPost, "/api/v1/auth/email/start", strings.NewReader(`{"email":"`+email+`"}`)))
	if start.Code != http.StatusAccepted {
		t.Fatalf("start auth status = %d body=%s", start.Code, start.Body.String())
	}
	link, ok := authSink.MagicLinkForChallenge(challengeID)
	if !ok {
		t.Fatal("auth magic link was not delivered")
	}
	verify := httptest.NewRecorder()
	handler.ServeHTTP(verify, httptest.NewRequest(http.MethodGet, "/api/v1/auth/email/verify?token="+url.QueryEscape(tokenFromMagicLinkForUploadRoundtrip(t, link)), nil))
	if verify.Code != http.StatusOK {
		t.Fatalf("verify auth status = %d body=%s", verify.Code, verify.Body.String())
	}
	sessionCookie := requireUploadRoundtripSessionCookie(t, verify)

	presignBody := []byte(`{"purpose":"resume","fileName":"resume.pdf","contentType":"application/pdf","byteSize":1024}`)
	firstPresign := serveUploadRoundtripPresign(handler, sessionCookie, "upload-roundtrip-presign-key", presignBody)
	if firstPresign.Code != http.StatusCreated {
		t.Fatalf("presign status = %d body=%s", firstPresign.Code, firstPresign.Body.String())
	}
	var presign api.UploadPresign
	if err := json.Unmarshal(firstPresign.Body.Bytes(), &presign); err != nil {
		t.Fatalf("decode presign: %v", err)
	}
	if presign.FileObjectId == "" || presign.UploadUrl == "" || presign.Method != http.MethodPut {
		t.Fatalf("invalid presign response: %+v", presign)
	}
	t.Cleanup(func() {
		objects := objectstore.NewMinIOStore(live.minio)
		_ = objects.Delete(context.Background(), userID+"/resume/"+presign.FileObjectId+".pdf")
		_, _ = db.Exec(`delete from audit_events where resource_id = $1`, presign.FileObjectId)
	})

	replay := serveUploadRoundtripPresign(handler, sessionCookie, "upload-roundtrip-presign-key", presignBody)
	if replay.Code != http.StatusCreated {
		t.Fatalf("presign replay status = %d body=%s", replay.Code, replay.Body.String())
	}
	var replayPresign api.UploadPresign
	if err := json.Unmarshal(replay.Body.Bytes(), &replayPresign); err != nil {
		t.Fatalf("decode presign replay: %v", err)
	}
	if replayPresign.FileObjectId != presign.FileObjectId || replayPresign.UploadUrl != presign.UploadUrl || replayPresign.ExpiresAt != presign.ExpiresAt {
		t.Fatalf("idempotency replay drift: first=%+v replay=%+v", presign, replayPresign)
	}

	uploadReq, err := http.NewRequestWithContext(ctx, presign.Method, presign.UploadUrl, bytes.NewReader(bytes.Repeat([]byte("x"), 1024)))
	if err != nil {
		t.Fatalf("new signed PUT request: %v", err)
	}
	for key, value := range presign.Headers {
		uploadReq.Header.Set(key, fmt.Sprint(value))
	}
	uploadResp, err := http.DefaultClient.Do(uploadReq)
	if err != nil {
		t.Fatalf("PUT signed URL: %v", err)
	}
	defer uploadResp.Body.Close()
	if uploadResp.StatusCode < 200 || uploadResp.StatusCode >= 300 {
		body, _ := io.ReadAll(uploadResp.Body)
		t.Fatalf("PUT signed URL status = %d body=%s", uploadResp.StatusCode, string(body))
	}

	registered, err := uploadRoutes.Service.RegisterFileObject(ctx, uploadserviceRegisterInput(presign.FileObjectId, userID))
	if err != nil {
		t.Fatalf("RegisterFileObject: %v", err)
	}
	if registered.Status != uploadstore.StatusUploaded || registered.ByteSize != 1024 {
		t.Fatalf("registered record = %+v", registered)
	}

	deleteRec := httptest.NewRecorder()
	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/me", nil)
	deleteReq.AddCookie(sessionCookie)
	deleteReq.Header.Set("Idempotency-Key", "upload-roundtrip-delete-key")
	handler.ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusAccepted {
		t.Fatalf("delete me status = %d body=%s", deleteRec.Code, deleteRec.Body.String())
	}
	if err := prioritizeUploadRoundtripPrivacyJob(ctx, db, privacyJobID); err != nil {
		t.Fatalf("prioritize privacy delete job: %v", err)
	}

	processed, err := kernel.RunOnce(ctx)
	if err != nil {
		t.Fatalf("run privacy delete job: %v", err)
	}
	if !processed {
		t.Fatal("privacy_delete job was not processed")
	}
	assertUploadRoundtripPrivacyDelete(t, db, presign.FileObjectId, privacyRequestID)
}

type uploadRoundtripLiveConfig struct {
	databaseURL string
	minio       objectstore.MinIOConfig
}

func requireUploadLiveConfig(t *testing.T) uploadRoundtripLiveConfig {
	t.Helper()
	cfg := uploadRoundtripLiveConfig{
		databaseURL: os.Getenv("DATABASE_URL"),
		minio: objectstore.MinIOConfig{
			Endpoint:  os.Getenv("OBJECT_STORAGE_ENDPOINT"),
			Bucket:    os.Getenv("OBJECT_STORAGE_BUCKET"),
			AccessKey: os.Getenv("OBJECT_STORAGE_ACCESS_KEY"),
			SecretKey: os.Getenv("OBJECT_STORAGE_SECRET_KEY"),
		},
	}
	missing := []string{}
	if cfg.databaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.minio.Endpoint == "" {
		missing = append(missing, "OBJECT_STORAGE_ENDPOINT")
	}
	if cfg.minio.Bucket == "" {
		missing = append(missing, "OBJECT_STORAGE_BUCKET")
	}
	if cfg.minio.AccessKey == "" {
		missing = append(missing, "OBJECT_STORAGE_ACCESS_KEY")
	}
	if cfg.minio.SecretKey == "" {
		missing = append(missing, "OBJECT_STORAGE_SECRET_KEY")
	}
	if len(missing) > 0 {
		t.Skipf("missing live upload roundtrip env: %s", strings.Join(missing, ", "))
	}
	return cfg
}

func uploadRoundtripConfig(t *testing.T, live uploadRoundtripLiveConfig) *config.Loader {
	t.Helper()
	dir := t.TempDir()
	providersPath := filepath.Join(dir, "ai-providers.yaml")
	profilesPath := filepath.Join(dir, "ai-profiles.yaml")
	writeAPIFile(t, providersPath, `
providers:
  - name: unit-test-stub
    protocol: stub
    capabilities: [chat]
    version: 1.0.0
`)
	writeAPIFile(t, profilesPath, `
profiles:
  - name: target.import.default
    capability: chat
    status: active
    default:
      provider_ref: unit-test-stub
      model: stub-chat
    fallback: []
    timeout_ms: 1000
    max_tokens: 256
    rate_limit:
      rps: 1
      tpm: 1000
    route: target.import
    version: 1.0.0
`)
	promptsDir, rubricsDir := repoConfigPromptsRubrics(t)
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
objectStorage:
  provider: minio
  endpoint: "`+live.minio.Endpoint+`"
  bucket: "`+live.minio.Bucket+`"
  accessKey: "`+live.minio.AccessKey+`"
  secretKey: "`+live.minio.SecretKey+`"
upload:
  presignTTLSeconds: 600
  maxBytes:
    resume: 10485760
    targetJobAttachment: 10485760
    privacyExport: 5242880
ai:
  providerRegistryPath: "`+providersPath+`"
  modelProfilePath: "`+profilesPath+`"
  promptsDir: "`+promptsDir+`"
  rubricsDir: "`+rubricsDir+`"
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return loader
}

func ensureMinIOBucket(t *testing.T, ctx context.Context, cfg objectstore.MinIOConfig) {
	t.Helper()
	endpoint, secure, err := uploadRoundtripMinIOEndpoint(cfg.Endpoint)
	if err != nil {
		t.Fatalf("parse minio endpoint: %v", err)
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: secure,
	})
	if err != nil {
		t.Fatalf("create minio client: %v", err)
	}
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		t.Fatalf("check minio bucket: %v", err)
	}
	if exists {
		return
	}
	if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
		t.Fatalf("create minio bucket %q: %v", cfg.Bucket, err)
	}
}

func uploadRoundtripMinIOEndpoint(raw string) (string, bool, error) {
	u, err := url.Parse(strings.TrimRight(raw, "/"))
	if err != nil {
		return "", false, err
	}
	endpoint := u.Host
	secure := u.Scheme == "https"
	if endpoint == "" {
		endpoint = u.Path
		secure = false
	}
	return endpoint, secure, nil
}

func tokenFromMagicLinkForUploadRoundtrip(t *testing.T, link string) string {
	t.Helper()
	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse magic link: %v", err)
	}
	token := u.Query().Get("token")
	if token == "" {
		t.Fatalf("magic link missing token: %s", link)
	}
	return token
}

func requireUploadRoundtripSessionCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == auth.SessionCookieName && cookie.Value != "" {
			return cookie
		}
	}
	t.Fatalf("session cookie missing: %#v", rec.Result().Cookies())
	return nil
}

func serveUploadRoundtripPresign(handler http.Handler, cookie *http.Cookie, key string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/uploads/presign", bytes.NewReader(body))
	req.AddCookie(cookie)
	req.Header.Set("Idempotency-Key", key)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func uploadserviceRegisterInput(fileObjectID, userID string) uploadservice.RegisterFileObjectInput {
	return uploadservice.RegisterFileObjectInput{
		FileObjectID:    fileObjectID,
		OwnerUserID:     userID,
		ExpectedPurpose: uploadstore.PurposeResume,
	}
}

func prioritizeUploadRoundtripPrivacyJob(ctx context.Context, db *sql.DB, jobID string) error {
	first := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	result, err := db.ExecContext(ctx, `
update async_jobs
set available_at = $1, created_at = $1, updated_at = $1
where id = $2 and job_type = 'privacy_delete' and status = 'queued'`, first, jobID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected one queued privacy_delete job for %s, updated %d", jobID, rows)
	}
	return nil
}

func assertUploadRoundtripPrivacyDelete(t *testing.T, db *sql.DB, fileObjectID string, privacyRequestID string) {
	t.Helper()
	var fileCount int
	if err := db.QueryRow(`select count(*) from file_objects where id = $1`, fileObjectID).Scan(&fileCount); err != nil {
		t.Fatalf("count file object after privacy delete: %v", err)
	}
	if fileCount != 0 {
		t.Fatalf("file object %s still exists after privacy delete", fileObjectID)
	}
	var requestStatus string
	if err := db.QueryRow(`select status from privacy_requests where id = $1`, privacyRequestID).Scan(&requestStatus); err != nil {
		t.Fatalf("select privacy request: %v", err)
	}
	if requestStatus != "completed" {
		var jobStatus, errorCode, errorMessage sql.NullString
		_ = db.QueryRow(`
select status, error_code, error_message
from async_jobs
where resource_id = $1 and job_type = 'privacy_delete'
order by updated_at desc
limit 1`, privacyRequestID).Scan(&jobStatus, &errorCode, &errorMessage)
		t.Fatalf("privacy request status = %q; privacy job status=%q error_code=%q error_message=%q", requestStatus, jobStatus.String, errorCode.String, errorMessage.String)
	}
	var auditMetadata string
	if err := db.QueryRow(`select metadata::text from audit_events where resource_id = $1 and action = 'privacy.file_object_deleted'`, fileObjectID).Scan(&auditMetadata); err != nil {
		t.Fatalf("select privacy file audit tombstone: %v", err)
	}
	if !strings.Contains(auditMetadata, fileObjectID) || strings.Contains(auditMetadata, "objectKey") {
		t.Fatalf("invalid audit tombstone metadata: %s", auditMetadata)
	}
}

func cleanupUploadRoundtripRows(t *testing.T, db *sql.DB, userID, email, challengeID, sessionID, privacyRequestID, privacyJobID string) {
	t.Helper()
	cleanup := func() {
		_, _ = db.Exec(`delete from audit_events where resource_id in ($1, $2) or user_id = $3`, privacyRequestID, userID, userID)
		_, _ = db.Exec(`delete from async_jobs where id = $1 or resource_id = $2`, privacyJobID, privacyRequestID)
		_, _ = db.Exec(`delete from privacy_requests where id = $1 or user_id = $2`, privacyRequestID, userID)
		_, _ = db.Exec(`delete from idempotency_records where user_id = $1`, userID)
		_, _ = db.Exec(`delete from sessions where id = $1 or user_id = $2`, sessionID, userID)
		_, _ = db.Exec(`delete from auth_challenges where id = $1 or email = $2`, challengeID, email)
		_, _ = db.Exec(`delete from file_objects where user_id = $1`, userID)
		_, _ = db.Exec(`delete from user_settings where user_id = $1`, userID)
		_, _ = db.Exec(`delete from users where id = $1 or email = $2`, userID, email)
	}
	cleanup()
	t.Cleanup(cleanup)
}
