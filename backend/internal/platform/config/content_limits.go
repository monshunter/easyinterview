package config

import (
	"fmt"
	"strings"
)

// ContentLimits is the typed A4 view of every runtime content-size boundary.
// Missing YAML keys use DefaultContentLimits; explicit values must be positive
// and satisfy the cross-field relationships validated by ContentLimits.
type ContentLimits struct {
	HTTPMaxRequestBodyBytes        int64
	ResumeUploadBytes              int64
	PrivacyExportUploadBytes       int64
	ResumeMaxActive                int
	ResumeMaxExtractedTextBytes    int64
	ResumeMaxPasteTextBytes        int64
	TargetJobMaxRawTextBytes       int64
	PracticeMaxMessageBytes        int64
	PracticeMaxSessionTextBytes    int64
	ReportMaxFramedInputBytes      int64
	AIProviderMaxResponseBodyBytes int64
}

// DefaultContentLimits returns the code-owned defaults used when a size key is
// absent from the merged configuration. A new value is returned on every call
// so callers cannot mutate shared process state.
func DefaultContentLimits() ContentLimits {
	return ContentLimits{
		HTTPMaxRequestBodyBytes:        10 * 1024 * 1024,
		ResumeUploadBytes:              10 * 1024 * 1024,
		PrivacyExportUploadBytes:       5 * 1024 * 1024,
		ResumeMaxActive:                10,
		ResumeMaxExtractedTextBytes:    384 * 1024,
		ResumeMaxPasteTextBytes:        384 * 1024,
		TargetJobMaxRawTextBytes:       96 * 1024,
		PracticeMaxMessageBytes:        32 * 1024,
		PracticeMaxSessionTextBytes:    256 * 1024,
		ReportMaxFramedInputBytes:      896 * 1024,
		AIProviderMaxResponseBodyBytes: 4 * 1024 * 1024,
	}
}

// ContentLimits resolves configured overrides on top of code defaults and
// rejects explicit zero, negative, or inconsistent values. It never writes
// defaults back into the declarative configuration source.
func (l *Loader) ContentLimits() (ContentLimits, error) {
	if l == nil || l.k == nil {
		return ContentLimits{}, fmt.Errorf("config: loader is nil")
	}

	cfg := DefaultContentLimits()
	l.applyContentLimitInt64("http.maxRequestBodyBytes", &cfg.HTTPMaxRequestBodyBytes)
	l.applyContentLimitInt64("upload.maxBytes.resume", &cfg.ResumeUploadBytes)
	l.applyContentLimitInt64("upload.maxBytes.privacyExport", &cfg.PrivacyExportUploadBytes)
	if l.k.Exists("resume.maxActive") {
		cfg.ResumeMaxActive = l.GetInt("resume.maxActive")
	}
	l.applyContentLimitInt64("resume.maxExtractedTextBytes", &cfg.ResumeMaxExtractedTextBytes)
	l.applyContentLimitInt64("resume.maxPasteTextBytes", &cfg.ResumeMaxPasteTextBytes)
	l.applyContentLimitInt64("targetJob.maxRawTextBytes", &cfg.TargetJobMaxRawTextBytes)
	l.applyContentLimitInt64("practice.maxMessageBytes", &cfg.PracticeMaxMessageBytes)
	l.applyContentLimitInt64("practice.maxSessionTextBytes", &cfg.PracticeMaxSessionTextBytes)
	l.applyContentLimitInt64("report.maxFramedInputBytes", &cfg.ReportMaxFramedInputBytes)
	l.applyContentLimitInt64("ai.maxResponseBodyBytes", &cfg.AIProviderMaxResponseBodyBytes)

	var problems []string
	for _, field := range []struct {
		path  string
		value int64
	}{
		{path: "http.maxRequestBodyBytes", value: cfg.HTTPMaxRequestBodyBytes},
		{path: "upload.maxBytes.resume", value: cfg.ResumeUploadBytes},
		{path: "upload.maxBytes.privacyExport", value: cfg.PrivacyExportUploadBytes},
		{path: "resume.maxActive", value: int64(cfg.ResumeMaxActive)},
		{path: "resume.maxExtractedTextBytes", value: cfg.ResumeMaxExtractedTextBytes},
		{path: "resume.maxPasteTextBytes", value: cfg.ResumeMaxPasteTextBytes},
		{path: "targetJob.maxRawTextBytes", value: cfg.TargetJobMaxRawTextBytes},
		{path: "practice.maxMessageBytes", value: cfg.PracticeMaxMessageBytes},
		{path: "practice.maxSessionTextBytes", value: cfg.PracticeMaxSessionTextBytes},
		{path: "report.maxFramedInputBytes", value: cfg.ReportMaxFramedInputBytes},
		{path: "ai.maxResponseBodyBytes", value: cfg.AIProviderMaxResponseBodyBytes},
	} {
		if field.value <= 0 {
			problems = append(problems, fmt.Sprintf("%s must be positive", field.path))
		}
	}
	if cfg.ResumeMaxPasteTextBytes > cfg.ResumeMaxExtractedTextBytes {
		problems = append(problems, "resume.maxPasteTextBytes must not exceed resume.maxExtractedTextBytes")
	}
	if cfg.PracticeMaxMessageBytes > cfg.PracticeMaxSessionTextBytes {
		problems = append(problems, "practice.maxMessageBytes must not exceed practice.maxSessionTextBytes")
	}
	if len(problems) > 0 {
		return ContentLimits{}, fmt.Errorf("content limits invalid: %s", strings.Join(problems, "; "))
	}
	return cfg, nil
}

func (l *Loader) applyContentLimitInt64(path string, target *int64) {
	if l.k.Exists(path) {
		*target = int64(l.GetInt(path))
	}
}
