package main

import (
	"bitkub2-go/core"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("❌ .env NOT loaded:", err)
	} else {
		fmt.Println("✅ .env loaded")
	}

	if err := core.InitDB(os.Getenv("DB_PATH")); err != nil {
		fmt.Printf("Fatal error during DB initialization: %v\n", err)
		return
	}
	defer core.DB.Close()

	username := os.Getenv("BOT_USERNAME")
	password := os.Getenv("BOT_PASSWORD")

	r := gin.Default()
	r.Static("/static", "./templates")
	r.LoadHTMLGlob("templates/layout/*")

	r.GET("/login", func(c *gin.Context) {
		if cookie, _ := c.Cookie("session"); cookie == "authenticated" {
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
		c.HTML(http.StatusOK, "login.html", gin.H{"Error": ""})
	})

	r.GET("/logout", func(c *gin.Context) {
		c.SetCookie("session", "", -1, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/login")
	})

	r.POST("/login", func(c *gin.Context) {
		u := c.PostForm("username")
		p := c.PostForm("password")

		if u == username && p == password {
			c.SetCookie("session", "authenticated", 3600, "/", "", false, true)
			c.Redirect(http.StatusSeeOther, "/")
			return
		}

		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"Error": "Invalid credentials"})
	})

	authRequired := func(c *gin.Context) {
		session, err := c.Cookie("session")
		if err != nil || session != "authenticated" {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		c.Next()
	}

	r.GET("/", authRequired, func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Username": username,
		})
	})

	r.GET("/api/status", func(c *gin.Context) {
		summary := core.CalculatePortfolio()
		core.ConfigMutex.RLock()
		mode := "PRODUCTION"
		if core.IsDryRun {
			mode = "DRY_RUN"
		}
		core.ConfigMutex.RUnlock()

		c.JSON(http.StatusOK, gin.H{
			"status":      "Running",
			"mode":        mode,
			"last_run":    time.Now().Format("15:04:05"),
			"coin_price":  core.RoundFloat(core.LastCoinPrice, 2),
			"total_value": core.RoundFloat(summary.TotalValue, 2),
			"roi":         core.RoundFloat(summary.ROI, 2),
			"portfolio":   summary.Portfolio,
		})
	})

	r.GET("/api/history", func(c *gin.Context) {
		trades, err := core.GetProductionTrades(10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"trades": trades,
		})
	})

	r.POST("/api/mode/:mode", func(c *gin.Context) {
		newMode := c.Param("mode")
		core.ConfigMutex.Lock()
		switch newMode {
		case "dry":
			core.IsDryRun = true
		case "prod":
			core.IsDryRun = false
		}

		currentMode := core.IsDryRun
		core.ConfigMutex.Unlock()

		go core.SendDiscordModeChange(currentMode)
		c.Redirect(http.StatusFound, "/api/status")
	})

	go func() {
		core.SendDiscordStartup()
		core.StartBotLoop()
	}()
	r.Run(":8080")
}
