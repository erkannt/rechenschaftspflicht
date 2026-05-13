package views

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestNewEventForm_RendersDatalist_WithExistingTags(t *testing.T) {
	tags := []string{"alpha", "beta"}

	component := NewEventForm(tags)

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
		if !strings.Contains(html, `<option value="`+tag+`">`) {
			t.Errorf("expected <option value=\"%s\"> in output", tag)
		}
	}
}

func TestNewEventForm_RendersWithoutDatalist_WhenNoTags(t *testing.T) {
	component := NewEventForm(nil)

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
	tags := []string{"my-tag"}

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
