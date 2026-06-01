package views

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestNewEventForm_RendersDatalist_WithExistingTags(t *testing.T) {
	tags := []TagSuggestion{{Tag: "alpha"}, {Tag: "beta"}}

	component := NewEventForm(tags, FormState{})

	var buf bytes.Buffer
	err := component.Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("unexpected error rendering: %v", err)
	}

	html := buf.String()

	// Check input is wired to datalist
	if !strings.Contains(html, `list="tag-suggestions"`) {
		t.Error("expected input to have list=\"tag-suggestions\"")
	}

	// Check datalist element exists
	if !strings.Contains(html, `<datalist id="tag-suggestions">`) {
		t.Error("expected <datalist id=\"tag-suggestions\"> element")
	}

	// Check each tag appears as an option
	for _, tag := range tags {
		if !strings.Contains(html, `<option value="`+tag.Tag+`">`) {
			t.Errorf("expected <option value=\"%s\"> in output", tag.Tag)
		}
	}
}

func TestNewEventForm_RendersWithoutDatalist_WhenNoTags(t *testing.T) {
	component := NewEventForm(nil, FormState{})

	var buf bytes.Buffer
	err := component.Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("unexpected error rendering: %v", err)
	}

	html := buf.String()

	// Datalist should not appear
	if strings.Contains(html, "<datalist") {
		t.Error("expected no <datalist> element when no tags provided")
	}

	// Input should not have a list attribute
	if strings.Contains(html, `list="`) {
		t.Error("expected input to not have a list attribute when no tags provided")
	}

	// Form should still render correctly
	if !strings.Contains(html, `id="tag"`) {
		t.Error("expected tag input to be rendered")
	}
	if !strings.Contains(html, `id="value"`) {
		t.Error("expected value input to be rendered")
	}
	if !strings.Contains(html, `id="comment"`) {
		t.Error("expected comment textarea to be rendered")
	}
}

func TestNewEventFormWithSuccessBanner_RendersDatalist_WithExistingTags(t *testing.T) {
	tags := []TagSuggestion{{Tag: "my-tag"}}

	component := NewEventFormWithSuccessBanner(tags)

	var buf bytes.Buffer
	err := component.Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("unexpected error rendering: %v", err)
	}

	html := buf.String()

	// Check success banner is rendered
	if !strings.Contains(html, "Success") {
		t.Error("expected success banner to be rendered")
	}

	// Check datalist is rendered
	if !strings.Contains(html, `list="tag-suggestions"`) {
		t.Error("expected input to have list=\"tag-suggestions\"")
	}
}

func TestNewEventForm_RendersErrorSummary_WhenErrorsPresent(t *testing.T) {
	formState := FormState{
		Tag:     "BadTag",
		Value:   "10",
		Comment: "test",
		Errors: map[string]string{
			"tag": "Tag must start with a lowercase letter and contain only lowercase letters and hyphens",
		},
	}

	component := NewEventForm(nil, formState)

	var buf bytes.Buffer
	err := component.Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("unexpected error rendering: %v", err)
	}

	html := buf.String()

	// Error summary heading
	if !strings.Contains(html, "There is a problem") {
		t.Error("expected error summary heading 'There is a problem'")
	}

	// Error message in summary
	if !strings.Contains(html, "Tag must start with a lowercase letter") {
		t.Error("expected error message in summary")
	}

	// Error summary links to field
	if !strings.Contains(html, `href="#tag"`) {
		t.Error("expected error summary link to #tag")
	}
}

func TestNewEventForm_PreservesSubmittedValues_OnError(t *testing.T) {
	formState := FormState{
		Tag:     "BadTag",
		Value:   "not-a-number",
		Comment: "test comment",
		Errors: map[string]string{
			"tag":   "bad tag",
			"value": "bad value",
		},
	}

	component := NewEventForm(nil, formState)

	var buf bytes.Buffer
	err := component.Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("unexpected error rendering: %v", err)
	}

	html := buf.String()

	// Tag input preserves value
	if !strings.Contains(html, `value="BadTag"`) {
		t.Error("expected tag input to preserve value 'BadTag'")
	}

	// Value input preserves value
	if !strings.Contains(html, `value="not-a-number"`) {
		t.Error("expected value input to preserve value 'not-a-number'")
	}

	// Comment textarea preserves value
	if !strings.Contains(html, "test comment") {
		t.Error("expected comment textarea to preserve value 'test comment'")
	}

	// Error message appears near the field
	if !strings.Contains(html, "bad tag") {
		t.Error("expected 'bad tag' error message")
	}
	if !strings.Contains(html, "bad value") {
		t.Error("expected 'bad value' error message")
	}
}

func TestNewEventForm_NoErrorSummary_WhenNoErrors(t *testing.T) {
	component := NewEventForm(nil, FormState{})

	var buf bytes.Buffer
	err := component.Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("unexpected error rendering: %v", err)
	}

	html := buf.String()

	if strings.Contains(html, "There is a problem") {
		t.Error("expected no error summary when no errors")
	}
}
