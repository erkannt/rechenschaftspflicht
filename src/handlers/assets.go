package handlers

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func initFs(embeddedFs embed.FS) http.FileSystem {
	sub, err := fs.Sub(embeddedFs, "assets")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

func AssetsHandler(embeddedAssetsFs embed.FS) httprouter.Handle {
	assetsFS := initFs(embeddedAssetsFs)

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.StripPrefix("/assets/", http.FileServer(assetsFS)).ServeHTTP(w, r)
	}
}
