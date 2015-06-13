package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/rlmcpherson/s3gof3r"

	"github.com/gorilla/mux"
)

var bucketName = os.Getenv("S3_BUCKET_NAME")

type JSONError struct {
	Error string `json:"error"`
}

type UploadResult struct {
	MP4URL  string `json:"mp4url"`
	WEBMURL string `json:"webmurl"`
	PNGURL  string `json:"jpgurl"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}

func main() {
	port := flag.String("port", "9090", "the port to bind to")
	flag.Parse()

	r := mux.NewRouter()

	r.Handle("/", http.HandlerFunc(rootHandler))
	r.Handle("/upload", http.HandlerFunc(uploadHandler))
	r.Handle("/{asset}", http.HandlerFunc(assetHandler))
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
	if response.Header.Get("Content-Type") != "image/gif" {
		return "", errors.New(fmt.Sprintf("%q is not an image/gif", gifURL))
	}

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

	if videoExtension == "jpg" {
		o, err := exec.Command(
			"convert",
			gifPath+"[0]",
			videoPath,
		).CombinedOutput()
		if err != nil {
			fmt.Println(string(o))
			return "", err
		}
		return videoPath, nil
	} else if videoExtension == "webm" {
		o, err := exec.Command(
			"vendor/ffmpeg-2.7-64bit-static/ffmpeg",
			"-i", gifPath,
			"-y",
			"-b:v", "5M",
			videoPath,
		).CombinedOutput()
		if err != nil {
			fmt.Println(string(o))
			return "", err
		}
		return videoPath, nil
	} else if videoExtension == "mp4" {
		o, err := exec.Command(
			"vendor/ffmpeg-2.7-64bit-static/ffmpeg",
			"-i", gifPath,
			"-y",
			"-crf", "23",
			"-b:v", "500K",
			videoPath,
		).CombinedOutput()
		if err != nil {
			fmt.Println(string(o))
			return "", err
		}
		return videoPath, nil
	}

	return "", nil
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
	w.Write([]byte("gifs?"))
}

func assetHandler(w http.ResponseWriter, r *http.Request) {
	asset := mux.Vars(r)["asset"]

	http.Redirect(w, r, "https://s3.amazonaws.com/ancientcitadelgifs/"+asset, http.StatusTemporaryRedirect)
	return

	contentTypes := map[string]string{
		".webm": "video/webm",
		".mp4":  "video/mp4",
		".jpg":  "image/jpeg",
	}
	for e, t := range contentTypes {
		if strings.HasSuffix(asset, e) {
			w.Header().Set("Content-Type", t)
			break
		}
	}

	keys, err := s3gof3r.EnvKeys()
	if err != nil {
		serveError(w, err.Error())
		return
	}

	s3 := s3gof3r.New("", keys)
	bucket := s3.Bucket(bucketName)

	reader, _, err := bucket.GetReader(asset, nil)
	if err != nil {
		serveError(w, err.Error())
		return
	}

	_, err = io.Copy(w, reader)
	if err != nil {
		serveError(w, err.Error())
		return
	}

	err = reader.Close()
	if err != nil {
		serveError(w, err.Error())
		return
	}
}

func serveError(w http.ResponseWriter, e string) {
	b, _ := json.Marshal(JSONError{Error: e})
	w.Header().Set("Content-Type", "application/json")
	log.Println("error: " + e)
	http.Error(w, string(b), http.StatusInternalServerError)
}

func putToS3(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	keys, err := s3gof3r.EnvKeys()
	if err != nil {
		return err
	}

	contentTypes := map[string]string{
		".webm": "video/webm",
		".mp4":  "video/mp4",
		".jpg":  "image/jpeg",
	}
	header := http.Header{}
	for e, t := range contentTypes {
		if strings.HasSuffix(path, e) {
			header.Set("Content-Type", t)
			break
		}
	}

	s3 := s3gof3r.New("", keys)
	bucket := s3.Bucket(bucketName)
	writer, err := bucket.PutWriter(path, header, nil)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, file)
	if err != nil {
		return err
	}

	return nil
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gifURL := r.URL.Query().Get("u")
	if gifURL == "" {
		serveError(w, "please specify a file to download")
		return
	}

	fmt.Printf("downloading %q...\n", gifURL)
	gifPath, err := downloadFile(gifURL)
	if err != nil {
		serveError(w, err.Error())
		return
	}
	fi, err := os.Stat(gifPath)
	if err != nil {
		serveError(w, err.Error())
		return
	}
	fmt.Printf("downloaded %d bytes...\n", fi.Size())

	width, height, err := getImageDimensions(gifPath)
	if err != nil {
		serveError(w, "error getting dimensions "+err.Error())
		return
	}

	for _, extension := range []string{"webm", "mp4", "jpg"} {
		fmt.Printf("converting %q to %v...\n", gifPath, extension)

		videoPath, err := convertFile(gifURL, gifPath, extension)
		if err != nil {
			serveError(w, err.Error())
			return
		}

		fmt.Printf("uploading %q to S3...\n", videoPath)
		err = putToS3(videoPath)
		if err != nil {
			serveError(w, err.Error())
			return
		}

		err = os.Remove(videoPath)
		if err != nil {
			serveError(w, err.Error())
			return
		}
	}

	err = os.Remove(gifPath)
	if err != nil {
		serveError(w, err.Error())
		return
	}

	mp4Path := strings.TrimSuffix(gifPath, ".gif") + ".mp4"
	webmPath := strings.TrimSuffix(gifPath, ".gif") + ".webm"
	jpgPath := strings.TrimSuffix(gifPath, ".gif") + ".jpg"

	uploadResult := UploadResult{
		MP4URL:  "http://gifs.ancientcitadel.com/" + mp4Path,
		WEBMURL: "http://gifs.ancientcitadel.com/" + webmPath,
		PNGURL:  "http://gifs.ancientcitadel.com/" + jpgPath,
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
