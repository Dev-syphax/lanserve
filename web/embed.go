package web

import "embed"

//go:embed template.html static/*
var FS embed.FS
