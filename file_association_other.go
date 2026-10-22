//go:build !windows

package main

func registerSparrowFileAssociations() error   { return nil }
func unregisterSparrowFileAssociations() error { return nil }
