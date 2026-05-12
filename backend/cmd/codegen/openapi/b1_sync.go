package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

const (
	b1BlockBeginMarker = "    # === B1-AUTO-START: synced from shared/conventions.yaml by `make codegen-openapi` ==="
	b1BlockEndMarker   = "    # === B1-AUTO-END ==="
)

// syncB1AutoBlock rewrites the section of openapi.yaml between
// `# === B1-AUTO-START` and `# === B1-AUTO-END ===` with content rendered
// from shared/conventions.yaml. The block contains:
//
//   - ApiErrorCode (errors[].code + OpenAPI-local P0 export exceptions)
//   - ApiError (B1 inner error object), ApiErrorResponse (B2 envelope), PageInfo
//   - JobStatus (B1 jobStatuses)
//   - B1 enums[]
//
// The function is idempotent: invoking it on an in-sync file leaves the file
// byte-identical, so the file participates correctly in `make codegen-check`.
//
// Hand-edits inside the block are overwritten by design; B1 is the truth
// source. To extend or alter the shared shape, edit conventions.yaml.
func syncB1AutoBlock(openapiPath string, conv *Conventions) error {
	data, err := os.ReadFile(openapiPath)
	if err != nil {
		return err
	}
	beginIdx, endIdx, err := findBlockOffsets(data)
	if err != nil {
		return err
	}
	rendered, err := renderB1AutoBlock(conv)
	if err != nil {
		return err
	}
	out := bytes.Buffer{}
	out.Write(data[:beginIdx])
	out.WriteString(b1BlockBeginMarker)
	out.WriteByte('\n')
	out.WriteString(rendered)
	out.WriteString(b1BlockEndMarker)
	out.WriteByte('\n')
	out.Write(data[endIdx:])

	if bytes.Equal(out.Bytes(), data) {
		return nil
	}
	return os.WriteFile(openapiPath, out.Bytes(), 0o644)
}

// findBlockOffsets locates the byte offsets immediately before the begin
// marker and immediately after the end-marker line in the openapi.yaml
// source bytes. Returns offsets such that data[beginIdx:endIdx] is exactly
// the begin-marker line + body + end-marker line.
func findBlockOffsets(data []byte) (int, int, error) {
	source := string(data)
	beginLineIdx := strings.Index(source, b1BlockBeginMarker)
	if beginLineIdx < 0 {
		return 0, 0, fmt.Errorf("missing B1-AUTO-START marker in openapi.yaml")
	}
	bodyStart := beginLineIdx + len(b1BlockBeginMarker)
	if bodyStart < len(source) && source[bodyStart] == '\n' {
		bodyStart++
	}
	endLineIdx := strings.Index(source[bodyStart:], b1BlockEndMarker)
	if endLineIdx < 0 {
		return 0, 0, fmt.Errorf("missing B1-AUTO-END marker in openapi.yaml")
	}
	endLineIdx += bodyStart
	endLineEnd := endLineIdx + len(b1BlockEndMarker)
	if endLineEnd < len(source) && source[endLineEnd] == '\n' {
		endLineEnd++
	}
	return beginLineIdx, endLineEnd, nil
}

// renderB1AutoBlock renders the YAML body that lives between the markers.
// All emitted content is indented under `components.schemas:` (8 spaces for
// schema names, 10 for properties) to match the surrounding hand-authored
// indentation and keep `swagger-cli validate` happy.
func renderB1AutoBlock(conv *Conventions) (string, error) {
	var sb strings.Builder
	sb.WriteString("    # Hand edits inside this block will be overwritten. To extend the contract,\n")
	sb.WriteString("    # update `shared/conventions.yaml` (B1 truth source) and re-run codegen.\n")
	sb.WriteString("    # The block tracks: ApiError inner object, ApiErrorResponse envelope,\n")
	sb.WriteString(fmt.Sprintf("    # ApiErrorCode, PageInfo, JobStatus, and the %d enums catalogued in\n", len(conv.Enums)))
	sb.WriteString("    # `shared/conventions.yaml#enums`.\n")

	// ApiErrorCode
	sb.WriteString("    ApiErrorCode:\n")
	sb.WriteString("      type: string\n")
	sb.WriteString("      description: |\n")
	sb.WriteString("        Documented error codes synced from\n")
	sb.WriteString("        `shared/conventions.yaml#errors[].code`; B1 owns the literal set.\n")
	sb.WriteString("        P0 export exceptions such as `PRIVACY_EXPORT_NOT_AVAILABLE` and\n")
	sb.WriteString("        `RESUME_EXPORT_NOT_AVAILABLE` are included here so the OpenAPI\n")
	sb.WriteString("        contract is self-contained for 501 export responses.\n")
	sb.WriteString("      enum:\n")
	for _, c := range b1OrderedErrorCodes(conv) {
		sb.WriteString("        - ")
		sb.WriteString(c)
		sb.WriteByte('\n')
	}
	sb.WriteByte('\n')

	// ApiError
	sb.WriteString("    ApiError:\n")
	sb.WriteString("      type: object\n")
	sb.WriteString("      required: [code, message, requestId, retryable]\n")
	sb.WriteString("      description: |\n")
	sb.WriteString("        Canonical error object per B1 `structures.ApiError`. B2 wraps it in\n")
	sb.WriteString("        `ApiErrorResponse` for the wire body `{error: ...}`.\n")
	sb.WriteString("      properties:\n")
	sb.WriteString("        code:\n")
	sb.WriteString("          $ref: '#/components/schemas/ApiErrorCode'\n")
	sb.WriteString("        message:\n")
	sb.WriteString("          type: string\n")
	sb.WriteString("          description: Human-readable summary; safe to surface to operators (no secrets, per A4 redaction rules).\n")
	sb.WriteString("        requestId:\n")
	sb.WriteString("          type: string\n")
	sb.WriteString("          description: Request correlation id; equals the X-Request-ID response header for the same request.\n")
	sb.WriteString("        retryable:\n")
	sb.WriteString("          type: boolean\n")
	sb.WriteString("          description: Hint to the client whether retry is safe, as defined by B1 shared-conventions-codified.\n")
	sb.WriteString("        details:\n")
	sb.WriteString("          type: object\n")
	sb.WriteString("          description: Optional structured diagnostic. Must not contain provider, prompt, or secret material per A4 redaction rules.\n")
	sb.WriteString("          additionalProperties: true\n")
	sb.WriteByte('\n')

	// ApiErrorResponse
	sb.WriteString("    ApiErrorResponse:\n")
	sb.WriteString("      type: object\n")
	sb.WriteString("      required: [error]\n")
	sb.WriteString("      description: Wire error response envelope defined by B1 shared-conventions-codified.\n")
	sb.WriteString("      properties:\n")
	sb.WriteString("        error:\n")
	sb.WriteString("          $ref: '#/components/schemas/ApiError'\n")
	sb.WriteByte('\n')

	// PageInfo
	sb.WriteString("    PageInfo:\n")
	sb.WriteString("      type: object\n")
	sb.WriteString("      required: [pageSize, hasMore]\n")
	sb.WriteString("      description: Cursor pagination envelope per spec D-5. Mirrors B1 `structures.PageInfo`.\n")
	sb.WriteString("      properties:\n")
	sb.WriteString("        nextCursor:\n")
	sb.WriteString("          type: string\n")
	sb.WriteString("          nullable: true\n")
	sb.WriteString("          description: Opaque cursor for the next page; `null` on the last page.\n")
	sb.WriteString("        pageSize:\n")
	sb.WriteString("          type: integer\n")
	sb.WriteString("          format: int32\n")
	sb.WriteString("          minimum: 1\n")
	sb.WriteString("          maximum: 100\n")
	sb.WriteString("        hasMore:\n")
	sb.WriteString("          type: boolean\n")
	sb.WriteByte('\n')

	// JobStatus
	sb.WriteString("    JobStatus:\n")
	sb.WriteString("      type: string\n")
	sb.WriteString("      description: Synced from `shared/conventions.yaml#jobStatuses`.\n")
	sb.WriteString("      enum:\n")
	for _, v := range conv.JobStatuses {
		sb.WriteString("        - ")
		sb.WriteString(v)
		sb.WriteByte('\n')
	}
	sb.WriteByte('\n')

	// B1 enums
	for i, e := range conv.Enums {
		sb.WriteString("    ")
		sb.WriteString(e.Name)
		sb.WriteString(":\n")
		sb.WriteString("      type: string\n")
		sb.WriteString(fmt.Sprintf("      description: Synced from `shared/conventions.yaml#enums[%s]` sourceSection %s.\n", e.Name, e.SourceSection))
		sb.WriteString("      enum: [")
		for j, v := range e.Values {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(v)
		}
		sb.WriteString("]\n")
		if i < len(conv.Enums)-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String(), nil
}

// b1OrderedErrorCodes returns the documented error code literals in a stable
// deterministic order: B1 declaration order first, then any extras
// (PRIVACY_EXPORT_NOT_AVAILABLE) appended.
func b1OrderedErrorCodes(conv *Conventions) []string {
	out := make([]string, 0, len(conv.Errors)+1)
	seen := map[string]struct{}{}
	for _, e := range conv.Errors {
		if _, ok := seen[e.Code]; ok {
			continue
		}
		seen[e.Code] = struct{}{}
		out = append(out, e.Code)
	}
	if _, ok := seen["PRIVACY_EXPORT_NOT_AVAILABLE"]; !ok {
		out = append(out, "PRIVACY_EXPORT_NOT_AVAILABLE")
	}
	return out
}
