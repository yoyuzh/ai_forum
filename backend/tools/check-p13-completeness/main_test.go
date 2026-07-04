package main

import "testing"

func TestCheckCompletenessAcceptsCurrentRepository(t *testing.T) {
	if err := check("../.."); err != nil {
		t.Fatal(err)
	}
}
