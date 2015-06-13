package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/gorilla/mux"
)

type JSONError struct {
	Error string `json:"error"`
}

type UploadResult struct {
	MP4URL  string `json:"mp4url"`
	WEBMURL string `json:"webmurl"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}

func main() {
	port := flag.String("port", "9090", "the port to bind to")
	flag.Parse()

	r := mux.NewRouter()

	r.Handle("/", http.HandlerFunc(rootHandler))
	r.Handle("/upload", http.HandlerFunc(uploadHandler))
	r.Handle("/fetch", http.HandlerFunc(fetchHandler))
	http.Handle("/", r)
	fmt.Printf("Starting on port %v...\n", *port)

	err := http.ListenAndServe("0.0.0.0:"+*port, nil)
	log.Fatal(err)
}

func downloadFile(gifURL string) (string, error) {
	outputPath := outputPath(gifURL, "gif")
	if _, err := os.Stat(outputPath); err == nil {
		return outputPath, nil
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	response, err := http.Get(gifURL)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}
	return file.Name(), nil
}

func outputPath(gifURL string, extension string) string {
	h := md5.New()
	io.WriteString(h, gifURL)
	return fmt.Sprintf("%x", h.Sum(nil)) + "." + extension
}

func convertFile(gifURL string, gifPath string, videoExtension string) (string, error) {
	videoPath := outputPath(gifURL, videoExtension)
	if _, err := os.Stat(videoPath); err == nil {
		return videoPath, nil
	}

	o, err := exec.Command(
		"vendor/ffmpeg-2.7-64bit-static/ffmpeg",
		"-i", gifPath,
		videoPath,
	).CombinedOutput()
	if err != nil {
		fmt.Println(string(o))
		return "", err
	}
	return videoPath, nil
}

func getImageDimensions(imagePath string) (int, int, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}
	return image.Width, image.Height, nil
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("gifcitadel"))
}

func fetchHandler(w http.ResponseWriter, r *http.Request) {
	gifURL := r.URL.Query().Get("u")
	if gifURL == "" {
		serveError(w, "please specify a file to download")
		return
	}

	videoExtension := r.URL.Query().Get("t")
	if videoExtension != "webm" && videoExtension != "mp4" {
		serveError(w, "please choose either mp4 or web extension")
		return
	}

	gifPath, err := downloadFile(gifURL)
	if err != nil {
		serveError(w, err.Error())
		return
	}

	videoPath, err := convertFile(gifURL, gifPath, videoExtension)
	if err != nil {
		serveError(w, err.Error())
		return
	}
	http.ServeFile(w, r, videoPath)
}

func serveError(w http.ResponseWriter, e string) {
	b, _ := json.Marshal(JSONError{Error: e})
	w.Header().Set("Content-Type", "application/json")
	log.Println("error: " + e)
	http.Error(w, string(b), http.StatusInternalServerError)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gifURL := r.URL.Query().Get("u")
	if gifURL == "" {
		serveError(w, "please specify a file to download")
		return
	}

	gifPath, err := downloadFile(gifURL)
	if err != nil {
		serveError(w, err.Error())
		return
	}

	fmt.Printf("uploading %q to %q\n", gifURL, gifPath)

	width, height, err := getImageDimensions(gifPath)
	if err != nil {
		serveError(w, "error getting dimensions "+err.Error())
		return
	}

	mp4Query := url.Values{}
	mp4Query.Set("u", gifURL)
	mp4Query.Set("t", "mp4")

	webmQuery := url.Values{}
	webmQuery.Set("u", gifURL)
	webmQuery.Set("t", "webm")

	uploadResult := UploadResult{
		MP4URL:  "/fetch?" + mp4Query.Encode(),
		WEBMURL: "/fetch?" + webmQuery.Encode(),
		Width:   width,
		Height:  height,
	}

	js, err := json.Marshal(uploadResult)
	if err != nil {
		serveError(w, err.Error())
		return
	}

	_, err = w.Write(js)
	if err != nil {
		serveError(w, err.Error())
		return
	}
}
