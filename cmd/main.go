//package main
//
//import (
//	"artifactdooni/api"
//	"log"
//	"net/http"
//)
//
//func main() {
//	mux := http.NewServeMux()
//	mux.HandleFunc("/v2/", api.V2PingHandler)
//
//	log.Println("Starting Artifactdooni registry on :5000")
//	err := http.ListenAndServe(":5000", mux)
//	if err != nil {
//		log.Fatal(err)
//	}
//}

package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// V2PingHandler responds to the /v2/ ping route for Docker clients
func V2PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
	w.WriteHeader(http.StatusOK)
}

// UploadHandler handles uploads to /upload/{name}
func UploadHandler(basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/upload/")
		if name == "" {
			http.Error(w, "Missing artifact name", http.StatusBadRequest)
			return
		}

		path := filepath.Join(basePath, name)
		os.MkdirAll(filepath.Dir(path), 0755)

		f, err := os.Create(path)
		if err != nil {
			http.Error(w, "Failed to create file", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		_, err = io.Copy(f, r.Body)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Upload successful"))
	}
}

// DownloadHandler handles downloads from /download/{name}
func DownloadHandler(basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/download/")
		if name == "" {
			http.Error(w, "Missing artifact name", http.StatusBadRequest)
			return
		}

		path := filepath.Join(basePath, name)
		f, err := os.Open(path)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer f.Close()

		w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(name))
		w.WriteHeader(http.StatusOK)
		io.Copy(w, f)
	}
}

// AuthMiddleware applies a simple token check to secure endpoints
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//token := r.Header.Get("Authorization")
		//if token != "Bearer your-secret-token" {
		//	http.Error(w, "Unauthorized", http.StatusUnauthorized)
		//	return
		//}
		next.ServeHTTP(w, r)
	})
}

func main() {
	basePath := "./storage"
	mux := http.NewServeMux()

	mux.Handle("/v2/", AuthMiddleware(http.HandlerFunc(V2PingHandler)))
	mux.Handle("/upload/", AuthMiddleware(UploadHandler(basePath)))
	mux.Handle("/download/", AuthMiddleware(DownloadHandler(basePath)))

	log.Println("🚀 Starting registry on :5000")
	err := http.ListenAndServe(":5000", mux)
	if err != nil {
		log.Fatal(err)
	}
}
