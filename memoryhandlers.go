package httputils

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/mailru/easyjson"
	"net/http"
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

func (h *HttpMemoryHandlerImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stats, err := getAllMemoryStats()
	if err != nil {
		fmt.Printf("Error reading memory stats from /proc/meminfo: %v", err)
	}
	bytes, err := easyjson.Marshal(stats)
	if err != nil {
		fmt.Printf("Error marshalling memory stats to json: %v", err)
	}
	_, err = w.Write(bytes)
	if err != nil {
		fmt.Printf("Error writing memory stats to http response: %v", err)
	}
}

func HttpMemoryHandler() http.Handler {
	return &HttpMemoryHandlerImpl{}
}
