package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

// getStaticFS 返回静态文件的文件系统
func getStaticFS() http.FileSystem {
	// 从嵌入的文件系统中获取static目录
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(staticFS)
}