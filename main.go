package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	glog "github.com/labstack/gommon/log"
)

type (
	rootResponse struct {
		Data interface{} `json:"data"`
	}

	createBoomerangResponse struct {
		Frames []*frame `json:"frames"`
		Fps    string   `json:"fps"`
	}

	frame struct {
		Name   string `json:"name"`
		Source string `json:"src"`
		Picked bool   `json:"picked"`
	}
)

func writeCookie(c echo.Context, sID string) error {
	cookie := new(http.Cookie)
	cookie.Name = "session.id"
	cookie.Value = sID
	cookie.Expires = time.Now().Add(1 * time.Hour)
	c.SetCookie(cookie)
	return nil
}

func readCookie(c echo.Context) string {
	cookie, err := c.Cookie("session.id")
	if err != nil {
		c.Logger().Error("error reading cookie", err)
		return ""
	}
	return cookie.Value
}

func getEnv(name, fallback string) (value string) {
	value = os.Getenv(name)
	if value == "" {
		value = fallback
	}
	return
}

func saveMultipartFile(file *multipart.FileHeader, filename string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	dst, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	_, err = io.Copy(dst, src)

	return err
}

func runCommand(ctx context.Context, cmd *exec.Cmd) error {
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	cmd.Stdout = &stdOut
	err := cmd.Run()
	//log.Printf("Err: %s\nOut: %s\n", stdErr.String(), stdOut.String())

	return err
}

func createFrames(ctx context.Context, vidPath string, fps string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	outputPath := filepath.Join(wd, vidPath)
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", "scripts/create-frames-1/main.sh "+outputPath+"/vid.mp4 output.gif "+fps)
	return runCommand(ctx, cmd)
}

func boomerangFromFrames(ctx context.Context, vidPath string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	outputPath := filepath.Join(wd, vidPath)
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", "scripts/create-loop-gif-1/main.sh "+outputPath+"/output.gif")
	return runCommand(ctx, cmd)
}

func pickFrames(ctx context.Context, vidPath string, frames []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	outputPath := filepath.Join(wd, vidPath)
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", "scripts/pick-frames-1/main.sh "+outputPath+" "+strings.Join(frames, " "))
	return runCommand(ctx, cmd)
}

func listFrames(ctx context.Context, vidPath string, pickedOnly bool) ([]*frame, error) {
	var framesDir string
	if pickedOnly {
		framesDir = "picked-frames"
	} else {
		framesDir = "frames"
	}
	framesPath := filepath.Join(vidPath, framesDir)
	fileInfos, err := ioutil.ReadDir(framesPath)
	if err != nil {
		return nil, err
	}

	var frames []*frame
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() || !strings.HasSuffix(fileInfo.Name(), ".jpg") {
			continue
		}
		frames = append(frames, &frame{Name: fileInfo.Name()})
	}

	return frames, nil
}

func sessionFilePath(sessionID string) string {
	return filepath.Join(".", "sess-data", sessionID)
}

func deleteSessionFiles(sessionID string) {
	outputPath := sessionFilePath(sessionID)
	log.Println("Deleting session files", outputPath)
	err := os.RemoveAll(outputPath)
	if err != nil {
		log.Println("Error deleting session files", err)
	}
}

func handlerGetOutputs(c echo.Context) error {
	sID := readCookie(c)
	if sID == "" {
		return errors.New("No session id")
	}

	outputPath := sessionFilePath(sID)

	staticMiddleware := middleware.Static(outputPath)
	staticMiddlewareFunc := staticMiddleware(func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "This output does not exist")
	})
	return staticMiddlewareFunc(c)
}

func handlerPostInput(c echo.Context) error {
	formFile, err := c.FormFile("file")
	if err != nil && err != http.ErrMissingFile {
		return err
	}

	var sID string
	if formFile != nil {
		sID := readCookie(c)
		if sID != "" {
			// Uploading a new video, delete all previous temporary files
			deleteSessionFiles(sID)
		}
	}

	if sID == "" {
		ID, err := uuid.NewUUID()
		if err != nil {
			return err
		}
		sID = ID.String()
	}

	if err = writeCookie(c, sID); err != nil {
		return err
	}

	vidPath := sessionFilePath(sID)
	if err := os.MkdirAll(vidPath, 0777); err != nil {
		return err
	}

	time.AfterFunc(15*time.Minute, func() {
		deleteSessionFiles(sID)
	})

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Minute)
	defer cancel()

	fps := c.FormValue("fps")
	if fps == "" {
		fps = "2"
	}

	if formFile != nil {
		if err = saveMultipartFile(formFile, vidPath+"/vid.mp4"); err != nil {
			return err
		}
		err = createFrames(ctx, vidPath, fps)
		if err != nil {
			return err
		}
	} else {
		pickedFramesStr := c.FormValue("pickedFrames")
		var framesArr []string
		if err = json.Unmarshal([]byte(pickedFramesStr), &framesArr); err != nil {
			return err
		}

		err = pickFrames(ctx, vidPath, framesArr)
		if err != nil {
			return err
		}
	}
	err = boomerangFromFrames(ctx, vidPath)
	if err != nil {
		return err
	}

	frames, err := listFrames(ctx, vidPath, false)
	if err != nil {
		return err
	}

	pickedFrames, err := listFrames(ctx, vidPath, true)
	if err != nil {
		return err
	}

	pickedFramesByName := make(map[string]*frame, len(pickedFrames))
	for _, pickedFrame := range pickedFrames {
		pickedFramesByName[pickedFrame.Name] = pickedFrame
	}
	for _, frame := range frames {
		if pickedFramesByName[frame.Name] != nil {
			frame.Picked = true
		}
		frame.Source = "frames/" + frame.Name
	}

	return c.JSON(http.StatusOK, rootResponse{Data: createBoomerangResponse{
		Frames: frames,
		Fps:    fps,
	}})
}

func main() {
	e := echo.New()
	e.Logger.SetLevel(glog.DEBUG)

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:4000", "http://video-to-boomerang.seriousben.com"},
		AllowCredentials: true,
	}))

	e.GET("/outputs/*", handlerGetOutputs)
	e.POST("/input", handlerPostInput)

	e.Logger.Fatal(e.Start(":" + getEnv("PORT", "80")))
}
