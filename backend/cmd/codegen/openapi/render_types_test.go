package main

import "testing"

func TestTSTypeForNullableStringIncludesNull(t *testing.T) {
	got := tsTypeFor(map[string]any{
		"type":     "string",
		"nullable": true,
	})

	if want := "string | null"; got != want {
		t.Fatalf("tsTypeFor(nullable string) = %q, want %q", got, want)
	}
}

func TestGoTypeForNullableStringUsesPointer(t *testing.T) {
	got := goTypeFor(map[string]any{
		"type":     "string",
		"nullable": true,
	})

	if want := "*string"; got != want {
		t.Fatalf("goTypeFor(nullable string) = %q, want %q", got, want)
	}
}

func TestTSResponseTypeUnionsDeclaredSuccessResponses(t *testing.T) {
	got := tsResponseType(map[string]any{
		"responses": map[string]any{
			"201": map[string]any{
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{"$ref": "#/components/schemas/ResumeVersion"},
					},
				},
			},
			"202": map[string]any{
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{"$ref": "#/components/schemas/BranchResumeVersionAccepted"},
					},
				},
			},
			"default": map[string]any{"$ref": "#/components/responses/ApiErrorResponse"},
		},
	})

	if want := "Types.ResumeVersion | Types.BranchResumeVersionAccepted"; got != want {
		t.Fatalf("tsResponseType(201+202) = %q, want %q", got, want)
	}
}

func TestTSNonOKStatusArgAllowsExplicit501Responses(t *testing.T) {
	got := tsNonOKStatusArg(map[string]any{
		"responses": map[string]any{
			"501": map[string]any{
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{"$ref": "#/components/schemas/ApiErrorResponse"},
					},
				},
			},
			"default": map[string]any{"$ref": "#/components/responses/ApiErrorResponse"},
		},
	})

	if want := "[501]"; got != want {
		t.Fatalf("tsNonOKStatusArg(501) = %q, want %q", got, want)
	}
}
