package auth

import "testing"

func TestPasswordService(t *testing.T) {
	service := NewPasswordService(10)
	hash, err := service.Hash("mat-khau-an-toan")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Compare(hash, "mat-khau-an-toan"); err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if err := service.Compare(hash, "sai-mat-khau"); err == nil {
		t.Fatal("Compare() error = nil, want mismatch")
	}
}
