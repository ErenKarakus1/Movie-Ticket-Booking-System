package models

import "testing"

func TestRequestNormalization(t *testing.T) {
	t.Run("movie trims title and description", func(t *testing.T) {
		req := AddMovieRequest{Title: "  Example  ", Description: "  A description  "}
		req.Normalize()
		if req.Title != "Example" || req.Description != "A description" {
			t.Fatalf("normalized movie fields = (%q, %q)", req.Title, req.Description)
		}
	})

	t.Run("auditorium trims name", func(t *testing.T) {
		req := CreateAuditoriumRequest{Name: "  Screen 1  "}
		req.Normalize()
		if req.Name != "Screen 1" {
			t.Fatalf("Name = %q, want %q", req.Name, "Screen 1")
		}
	})

	t.Run("seat labels are trimmed and uppercased", func(t *testing.T) {
		seats := NormalizeCreateSeatsRequest([]CreateSeatRequest{
			{RowLabel: "  a ", SeatNumber: 1},
			{RowLabel: " b  ", SeatNumber: 2},
		})
		if seats[0].RowLabel != "A" || seats[1].RowLabel != "B" {
			t.Fatalf("normalized seat labels = %q and %q, want A and B", seats[0].RowLabel, seats[1].RowLabel)
		}
	})
}
