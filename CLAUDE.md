# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`github.com/lestrrat-go/gcode` — a Go library for parsing and generating G-code (CNC/3D printer control language). Go 1.26.1.

## Build & Test

```bash
go build ./...
go test ./...
go test -run TestName ./pkg/...   # single test
go vet ./...
```
