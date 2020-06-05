package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"

	//"runtime/debug"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Cfg struct {
	PageSize      string
	Orientation   string
	MarginBottom  string
	MarginTop     string
	MarginLeft    string
	MarginRight   string
	PageShrinking string
	PageZoom      string
}

func getPdfFile(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Print(err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	defer file.Close()
	log.Print("Input file name ", header.Filename)
	if err != nil {
		fmt.Println(err)
	}
	fileBytes, err := ioutil.ReadAll(file)

	cfg := &Cfg{}

	allError := 0
	cfg.PageSize = r.FormValue("page_size")
	if len(cfg.PageSize) == 0 {
		allError++
	}
	cfg.Orientation = r.FormValue("orientation")
	if len(cfg.Orientation) == 0 {
		allError++
	}
	cfg.MarginBottom = r.FormValue("margin_bottom")
	if len(cfg.MarginBottom) == 0 {
		allError++
	}
	cfg.MarginTop = r.FormValue("margin_top")
	if len(cfg.MarginTop) == 0 {
		allError++
	}
	cfg.MarginLeft = r.FormValue("margin_left")
	if len(cfg.MarginLeft) == 0 {
		allError++
	}
	cfg.MarginRight = r.FormValue("margin_right")
	if len(cfg.MarginRight) == 0 {
		allError++
	}
	if r.FormValue("shrink") == "1" {
		cfg.PageShrinking = "--enable-smart-shrinking"
	} else {
		cfg.PageShrinking = "--disable-smart-shrinking"
	}
	if len(r.FormValue("zoom")) > 0 {
		cfg.PageZoom = r.FormValue("zoom")
	} else {
		cfg.PageZoom = "1"
	}
	if allError > 0 {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	tmpName := uuid.New().String()
	log.Print(len(fileBytes))
	fo, err := os.Create("/tmp/" + tmpName + ".html")
	if err != nil {
		log.Print(err)
	}
	defer fo.Close()
	_, err = fo.Write(fileBytes)
	if err != nil {
		log.Print(err)
	}

	_, err = exec.Command("wkhtmltopdf", cfg.PageShrinking, "--dpi", "600", "--page-size", cfg.PageSize, "--orientation", cfg.Orientation, "--zoom", cfg.PageZoom, "--margin-top", cfg.MarginTop, "--margin-bottom", cfg.MarginBottom, "--margin-left", cfg.MarginLeft, "--margin-right", cfg.MarginRight, "/tmp/"+tmpName+".html", "/tmp/"+tmpName+".pdf").Output()
	if err != nil {
		log.Print(err)
		http.Error(w, "application error", http.StatusInternalServerError)
		return
	}

	pdfBodyByte, err := ioutil.ReadFile("/tmp/" + tmpName + ".pdf")
	if err != nil {
		log.Print(err)
		http.Error(w, "application error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	fmt.Fprint(w, string(pdfBodyByte))
	log.Print("Pdf generated successfully, ", time.Since(start), "; ", memUsage())
	//debug.FreeOSMemory()
	//r = nil
	//runtime.GC()
	pprof.WriteHeapProfile(os.Stdout)
	return
}

func saveToFile(file []byte, name string) {
	log.Print(len(file))
	fo, err := os.Create("/tmp/" + name)
	if err != nil {
		log.Print(err)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			log.Print(err)
		}
	}()
	_, err = fo.Write(file)
	if err != nil {
		log.Print(err)
	}
}

func memUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	result := fmt.Sprintf("Alloc = %v MiB; TotalAlloc = %v MiB; Sys = %v MiB; NumGC = %v;", bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC)
	return result
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func main() {
	listenPort := "8080"
	if len(os.Getenv("PORT")) > 0 {
		listenPort = os.Getenv("PORT")
	}
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { getPdfFile(w, r) }).Methods("POST")
	log.Fatal(http.ListenAndServe(":"+listenPort, r))
}
