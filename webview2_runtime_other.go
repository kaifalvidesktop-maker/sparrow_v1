//go:build !windows

package main

func prepareBundledWebView2Runtime() error { return nil }

func showWebView2RuntimeError(error) {}
