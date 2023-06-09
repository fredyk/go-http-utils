package httputils

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/mailru/easyjson"
	"math/rand"
	"net/http"
	"time"
)

func FiberMemoryHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		stats, err := getAllMemoryStats()
		_, err = easyjson.MarshalToWriter(stats, ctx.Response().BodyWriter())
		return err
	}
}

type HttpMemoryHandlerImpl struct {
}

var ProcessHash string

func init() {
	ProcessHash = getRandomProcessHash4bytes()
}

func getRandomProcessHash4bytes() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	init := 0x10000000
	return fmt.Sprintf("%x", init+rnd.Intn(0xdfffffff))
}

func (h *HttpMemoryHandlerImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stats, err := getAllMemoryStats()
	if err != nil {
		fmt.Printf("Error reading memory stats from /proc/meminfo: %v", err)
	}
	bytes, err := easyjson.Marshal(stats)
	if err != nil {
		fmt.Printf("Error marshalling memory stats to json: %v", err)
	}
	// send headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(bytes)))
	// allow all origins
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// send body
	_, err = w.Write(bytes)
	if err != nil {
		fmt.Printf("Error writing memory stats to http response: %v", err)
	}
}

func HttpMemoryHandler() http.Handler {
	return &HttpMemoryHandlerImpl{}
}

func FiberPsHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		//// Try to get result of ps -aux
		//cmd := exec.Command("ps", "-aux")
		//out, err := cmd.Output()
		//if err != nil {
		//	return err
		//}
		//return ctx.SendString(string(out))
		var out []PsEntry
		var err error
		out, err = parseProcessList()
		if err != nil {
			return err
		}
		return ctx.JSON(out)

	}
}
