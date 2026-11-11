package repository

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsSeatAlreadyBookedError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "generated showtime seat constraint",
			err:  &pgconn.PgError{ConstraintName: "booking_seats_showtime_id_seat_id_key"},
			want: true,
		},
		{
			name: "different unique constraint",
			err:  &pgconn.PgError{ConstraintName: "booking_seats_pkey"},
			want: false,
		},
		{
			name: "legacy index name",
			err:  &pgconn.PgError{ConstraintName: "booking_seats_showtime_seat_unique"},
			want: false,
		},
		{
			name: "wrapped PostgreSQL error",
			err:  fmt.Errorf("insert booking seat: %w", &pgconn.PgError{ConstraintName: "booking_seats_showtime_id_seat_id_key"}),
			want: true,
		},
		{
			name: "non-PostgreSQL error",
			err:  errors.New("database unavailable"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSeatAlreadyBookedError(tt.err); got != tt.want {
				t.Fatalf("isSeatAlreadyBookedError() = %t, want %t", got, tt.want)
			}
		})
	}
}
