package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

type config struct {
	PageSize      string
	Orientation   string
	MarginBottom  string
	MarginTop     string
	MarginLeft    string
	MarginRight   string
	PageShrinking string
	PageZoom      string
}

type status struct {
	All  int64
	Good int64
}

var stat status

func getPdfFile(ctx *fasthttp.RequestCtx) {
	start := time.Now()
	stat.All++
	tmpName := uuid.New().String()
	log.Print("id: ", tmpName, ". Start upload file.")
	fileData, err := ctx.FormFile("file")
	if err != nil {
		log.Print(err)
		sentry.CaptureException(err)
		ctx.SetStatusCode(400)
		fmt.Fprintf(ctx, "can't read body")
		return
	}
	file, _ := fileData.Open()
	defer file.Close()
	//log.Print("Input file name ", header.Filename)
	if err != nil {
		log.Print(err)
		sentry.CaptureException(err)
	}
	fileBytes, err := ioutil.ReadAll(file)

	cfg := &config{}

	allError := 0
	cfg.PageSize = string(ctx.FormValue("page_size"))
	if len(cfg.PageSize) == 0 {
		allError++
	}
	cfg.Orientation = string(ctx.FormValue("orientation"))
	if len(cfg.Orientation) == 0 {
		allError++
	}
	cfg.MarginBottom = string(ctx.FormValue("margin_bottom"))
	if len(cfg.MarginBottom) == 0 {
		allError++
	}
	cfg.MarginTop = string(ctx.FormValue("margin_top"))
	if len(cfg.MarginTop) == 0 {
		allError++
	}
	cfg.MarginLeft = string(ctx.FormValue("margin_left"))
	if len(cfg.MarginLeft) == 0 {
		allError++
	}
	cfg.MarginRight = string(ctx.FormValue("margin_right"))
	if len(cfg.MarginRight) == 0 {
		allError++
	}
	if string(ctx.FormValue("shrink")) == "1" {
		cfg.PageShrinking = "--enable-smart-shrinking"
	} else {
		cfg.PageShrinking = "--disable-smart-shrinking"
	}
	if len(ctx.FormValue("zoom")) > 0 {
		cfg.PageZoom = string(ctx.FormValue("zoom"))
	} else {
		cfg.PageZoom = "1"
	}
	if allError > 0 {
		ctx.SetStatusCode(400)
		fmt.Fprintf(ctx, "can't read body")
		return
	}

	fo, err := os.Create("/tmp/" + tmpName + ".html")
	if err != nil {
		log.Print(err)
		sentry.CaptureException(err)
		ctx.SetStatusCode(500)
		return
	}
	defer fo.Close()
	_, err = fo.Write(fileBytes)
	if err != nil {
		log.Print(err)
		sentry.CaptureException(err)
		ctx.SetStatusCode(500)
		return
	}

	log.Print("id: ", tmpName, ". Start generate pdf.")
	if cfg.PageZoom == "1" {
		_, err = exec.Command("wkhtmltopdf", cfg.PageShrinking, "--dpi", "300", "--page-size", cfg.PageSize, "--orientation", cfg.Orientation, "--margin-top", cfg.MarginTop, "--margin-bottom", cfg.MarginBottom, "--margin-left", cfg.MarginLeft, "--margin-right", cfg.MarginRight, "/tmp/"+tmpName+".html", "/tmp/"+tmpName+".pdf").Output()
	} else {
		_, err = exec.Command("wkhtmltopdf", cfg.PageShrinking, "--dpi", "300", "--page-size", cfg.PageSize, "--orientation", cfg.Orientation, "--zoom", cfg.PageZoom, "--margin-top", cfg.MarginTop, "--margin-bottom", cfg.MarginBottom, "--margin-left", cfg.MarginLeft, "--margin-right", cfg.MarginRight, "/tmp/"+tmpName+".html", "/tmp/"+tmpName+".pdf").Output()
	}
	if err != nil {
		log.Print(err)
		sentry.CaptureException(err)
		ctx.SetStatusCode(500)
		return
	}

	pdfBodyByte, err := ioutil.ReadFile("/tmp/" + tmpName + ".pdf")
	if err != nil {
		log.Print(err)
		sentry.CaptureException(err)
		ctx.SetStatusCode(500)
		return
	}

	err = os.Remove("/tmp/" + tmpName + ".html")
	err = os.Remove("/tmp/" + tmpName + ".pdf")
	if err != nil {
		log.Print(err)
		sentry.CaptureException(err)
		return
	}

	ctx.Response.Header.Set("Content-Type", "application/pdf")
	fmt.Fprint(ctx, string(pdfBodyByte))
	log.Print("id: ", tmpName, ". generated successfully, ", time.Since(start), "; ", memUsage())
	stat.Good++
	return
}

func saveToFile(file []byte, name string) {
	log.Print(len(file))
	fo, err := os.Create("/tmp/" + name)
	if err != nil {
		log.Print(err)
		sentry.CaptureException(err)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			log.Print(err)
			sentry.CaptureException(err)
		}
	}()
	_, err = fo.Write(file)
	if err != nil {
		log.Print(err)
		sentry.CaptureException(err)
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

func getStatus(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "application/json")
	fmt.Fprint(ctx, `{"all_request":`+strconv.FormatInt(stat.All, 10)+`,"good_request":`+strconv.FormatInt(stat.Good, 10)+`}`)
}

func main() {
	listenPort := "8080"
	if len(os.Getenv("PORT")) > 0 {
		listenPort = os.Getenv("PORT")
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRYURL"),
	})
	if err != nil {
		log.Panic(err)
	}
	stat = status{All: 0, Good: 0}
	router := fasthttprouter.New()
	router.POST("/", getPdfFile)
	router.GET("/status", getStatus)
	server := &fasthttp.Server{
		Handler:            router.Handler,
		MaxRequestBodySize: 100 << 20,
	}
	log.Print("App start on port ", listenPort)
	log.Fatal(server.ListenAndServe(":" + listenPort))
}
