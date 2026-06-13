package auth

import "testing"

func TestPasswordHasherHashAndVerify(t *testing.T) {
	t.Parallel()

	hasher := PasswordHasher{
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	}

	hash, err := hasher.Hash("super-secret-password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	ok, err := hasher.Verify("super-secret-password", hash)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !ok {
		t.Fatal("Verify() = false, want true")
	}

	ok, err = hasher.Verify("wrong-password", hash)
	if err != nil {
		t.Fatalf("Verify() wrong password error = %v", err)
	}
	if ok {
		t.Fatal("Verify() = true for wrong password, want false")
	}
}
