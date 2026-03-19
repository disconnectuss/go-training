package main

import (
	"testing"
)

// --- Contains tests ---

func TestContainsInt(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}

	// Same function works with int...
	if !Contains(numbers, 3) {
		t.Error("expected to find 3")
	}
	if Contains(numbers, 99) {
		t.Error("99 should not be found")
	}
}

func TestContainsString(t *testing.T) {
	cities := []string{"Istanbul", "Ankara", "Izmir"}

	// ...and with string! Same function, no duplication
	if !Contains(cities, "Istanbul") {
		t.Error("expected to find Istanbul")
	}
	if Contains(cities, "Bursa") {
		t.Error("Bursa should not be found")
	}
}

func TestContainsUser(t *testing.T) {
	users := []User{
		{ID: 1, Name: "Fatma"},
		{ID: 2, Name: "Ahmet"},
	}

	// Works with structs too! (User is comparable because all fields are comparable)
	if !Contains(users, User{ID: 1, Name: "Fatma"}) {
		t.Error("expected to find Fatma")
	}
}

// --- Map tests ---

func TestMapIntToString(t *testing.T) {
	numbers := []int{1, 2, 3}

	// Transform int → string using Map
	// Type inference: Go figures out T=int, R=string from the arguments
	strings := Map(numbers, func(n int) string {
		return "num-" + string(rune('0'+n))
	})

	if len(strings) != 3 {
		t.Errorf("expected 3 results, got %d", len(strings))
	}
}

func TestMapUserToName(t *testing.T) {
	users := []User{
		{ID: 1, Name: "Fatma"},
		{ID: 2, Name: "Ahmet"},
	}

	// Extract names from User structs — Map[User, string]
	names := Map(users, func(u User) string {
		return u.Name
	})

	if names[0] != "Fatma" || names[1] != "Ahmet" {
		t.Errorf("expected [Fatma, Ahmet], got %v", names)
	}
}

// --- Filter tests ---

func TestFilterEvenNumbers(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6}

	evens := Filter(numbers, func(n int) bool {
		return n%2 == 0
	})

	if len(evens) != 3 {
		t.Errorf("expected 3 even numbers, got %d", len(evens))
	}
	if evens[0] != 2 || evens[1] != 4 || evens[2] != 6 {
		t.Errorf("expected [2,4,6], got %v", evens)
	}
}

func TestFilterUsers(t *testing.T) {
	users := []User{
		{ID: 1, Name: "Fatma", Age: 25, City: "Istanbul"},
		{ID: 2, Name: "Ahmet", Age: 30, City: "Ankara"},
		{ID: 3, Name: "Elif", Age: 22, City: "Istanbul"},
	}

	// Filter users from Istanbul — works with our User struct!
	istanbulUsers := Filter(users, func(u User) bool {
		return u.City == "Istanbul"
	})

	if len(istanbulUsers) != 2 {
		t.Errorf("expected 2 Istanbul users, got %d", len(istanbulUsers))
	}
}

// --- Sum with type constraint tests ---

func TestSumInts(t *testing.T) {
	result := Sum([]int{1, 2, 3, 4, 5})

	if result != 15 {
		t.Errorf("expected 15, got %d", result)
	}
}

func TestSumFloats(t *testing.T) {
	result := Sum([]float64{1.5, 2.5, 3.0})

	if result != 7.0 {
		t.Errorf("expected 7.0, got %f", result)
	}
}

// Custom type based on int — works because of ~int in Number constraint
type Score int

func TestSumCustomType(t *testing.T) {
	scores := []Score{10, 20, 30}
	result := Sum(scores)

	// ~int in Number constraint allows types BASED ON int
	if result != 60 {
		t.Errorf("expected 60, got %d", result)
	}
}

// --- Pair tests ---

func TestPair(t *testing.T) {
	// int + string pair
	p := NewPair(1, "hello")

	if p.First != 1 {
		t.Errorf("expected First=1, got %d", p.First)
	}
	if p.Second != "hello" {
		t.Errorf("expected Second='hello', got '%s'", p.Second)
	}
}

func TestPairString(t *testing.T) {
	p := NewPair("name", 42)

	expected := "(name, 42)"
	if p.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, p.String())
	}
}

func TestPairWithUser(t *testing.T) {
	// Pair of user ID and User struct
	user := User{ID: 1, Name: "Fatma"}
	p := NewPair(user.ID, user)

	if p.First != 1 {
		t.Errorf("expected First=1, got %d", p.First)
	}
	if p.Second.Name != "Fatma" {
		t.Errorf("expected Second.Name='Fatma', got '%s'", p.Second.Name)
	}
}

// --- Response wrapper tests ---

func TestSuccessResponse(t *testing.T) {
	user := User{ID: 1, Name: "Fatma", Age: 25, City: "Istanbul"}
	resp := NewSuccessResponse(user)

	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Data.Name != "Fatma" {
		t.Errorf("expected data.Name='Fatma', got '%s'", resp.Data.Name)
	}
}

func TestSuccessResponseWithSlice(t *testing.T) {
	users := []User{
		{ID: 1, Name: "Fatma"},
		{ID: 2, Name: "Ahmet"},
	}
	resp := NewSuccessResponse(users)

	if !resp.Success {
		t.Error("expected success=true")
	}
	if len(resp.Data) != 2 {
		t.Errorf("expected 2 users, got %d", len(resp.Data))
	}
}

func TestErrorResponse(t *testing.T) {
	resp := NewErrorResponse("user not found")

	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Message != "user not found" {
		t.Errorf("expected message 'user not found', got '%s'", resp.Message)
	}
}

// --- Keys tests ---

func TestKeysFromMap(t *testing.T) {
	m := map[string]int{
		"fatma": 25,
		"ahmet": 30,
	}

	keys := Keys(m)

	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
	// Keys order is not guaranteed in maps, so check Contains
	if !Contains(keys, "fatma") || !Contains(keys, "ahmet") {
		t.Errorf("expected keys [fatma, ahmet], got %v", keys)
	}
}

func TestKeysFromAPIKeyMap(t *testing.T) {
	// Works with our existing validAPIKeys map!
	keys := Keys(validAPIKeys)

	if len(keys) != 3 {
		t.Errorf("expected 3 API keys, got %d", len(keys))
	}
}
