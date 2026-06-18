package tests

import "time"

// ptrTime is a tiny test helper shared by all *_test.go files in this
// package. We centralise it here to avoid duplicate-declaration errors.
func ptrTime(t time.Time) *time.Time { return &t }
