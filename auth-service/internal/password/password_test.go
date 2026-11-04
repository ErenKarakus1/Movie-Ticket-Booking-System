package password

import "testing"

func TestPasswordHashAndCompare(t *testing.T) {
	password := "CorrectHorseBatteryStaple123"
	hash, err := GeneratePasswordHash(password)
	if err != nil {
		t.Fatalf("GeneratePasswordHash() error = %v", err)
	}
	if hash == password {
		t.Fatal("password hash must not equal the plain-text password")
	}
	if !CompareHashAndPassword(hash, password) {
		t.Fatal("expected correct password to match hash")
	}
	if CompareHashAndPassword(hash, "wrong-password") {
		t.Fatal("expected incorrect password not to match hash")
	}
	if CompareHashAndPassword("not-a-bcrypt-hash", password) {
		t.Fatal("expected malformed hash not to match")
	}
}
