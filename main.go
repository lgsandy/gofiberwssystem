package main

import (
	"fmt"
	"gofiberhtmx/hardware"
	"gofiberhtmx/manager"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	// Start the WebSocket manager
	wsManager := manager.NewWebSocketManager()
	go wsManager.Run()

	// Create template engine
	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		})
	})

	// WebSocket route
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		wsManager.Register <- c
		defer func() { wsManager.Unregister <- c }()

		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			wsManager.SendBroadcast(msg)
		}
	}))
	go func(ws *manager.WebSocketManager) {
		for {
			systemData, err := hardware.GetSystemSection()
			if err != nil {
				fmt.Println(err)
				continue
			}

			diskData, err := hardware.GetDiskSection()
			if err != nil {
				fmt.Println(err)
				continue
			}
			cpuData, err := hardware.GetCpuSection()
			if err != nil {
				fmt.Println(err)
				continue
			}
			timeStamp := time.Now().Format("2006-01-02 15:04:05")
			msg := []byte(`
      <div hx-swap-oob="innerHTML:#update-timestamp">
        <p><i style="color: green" class="fa fa-circle"></i> ` + timeStamp + `</p>
      </div>
      <div hx-swap-oob="innerHTML:#system-data">` + systemData + `</div>
      <div hx-swap-oob="innerHTML:#cpu-data">` + cpuData + `</div>
      <div hx-swap-oob="innerHTML:#disk-data">` + diskData + `</div>`)
			ws.SendBroadcast(msg)
			time.Sleep(3 * time.Second)
		}
	}(wsManager)

	log.Fatal(app.Listen(":3000"))

	// Uncomment to start getting hardware info
	// go getHardwareInfo()
}
