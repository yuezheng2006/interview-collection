package main

import (
	"ai-writing-assistant/internal/handler"
	"ai-writing-assistant/internal/pkg/utils"
)

func main() {
	r := handler.NewRouter()
	utils.MustStartServer(r)
}
