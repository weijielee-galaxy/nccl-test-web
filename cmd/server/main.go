package main

import (
	"flag"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/weijielee-galaxy/nccl-test-web/internal/handlers"
	"github.com/weijielee-galaxy/nccl-test-web/web"
)

func main() {
	// 命令行参数
	port := flag.String("port", "8098", "Server port")
	flag.Parse()

	// 创建 Gin 路由
	r := gin.Default()

	// 注册API路由
	r.GET("/healthz", handlers.HealthCheck)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// IP列表管理接口（增删改查）
		v1.GET("/iplist/files", handlers.GetIPListFiles)      // 获取IP列表文件列表
		v1.POST("/iplist/:filename", handlers.SaveIPList)     // 创建/更新指定文件
		v1.GET("/iplist/:filename", handlers.GetIPList)       // 读取指定文件
		v1.PUT("/iplist/:filename", handlers.UpdateIPList)    // 更新指定文件
		v1.DELETE("/iplist/:filename", handlers.DeleteIPList) // 删除指定文件

		// NCCL 测试接口
		v1.GET("/nccl/defaults", handlers.GetNCCLTestDefaults)  // 获取默认参数
		v1.POST("/nccl/run", handlers.RunNCCLTest)              // 运行测试（一次性返回）
		v1.POST("/nccl/run-stream", handlers.RunNCCLTestStream) // 运行测试（流式返回）
		v1.POST("/nccl/stop", handlers.StopNCCLTest)            // 停止当前运行的测试
		v1.GET("/nccl/precheck", handlers.Precheck)             // 检查所有节点的 GPU 进程状态

		// 历史记录相关接口
		v1.GET("/history", handlers.GetHistoryList)              // 获取历史记录列表
		v1.GET("/history/:filename", handlers.GetHistoryContent) // 获取指定历史记录内容
		v1.DELETE("/history/:filename", handlers.DeleteHistory)  // 删除指定历史记录
	}

	// 嵌入前端静态文件
	staticFS, err := web.GetDistFS()
	if err != nil {
		log.Fatal("Failed to load embedded web files:", err)
	}

	// 提供静态文件服务
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 对于根路径，返回 index.html
		if path == "/" {
			path = "/index.html"
		}

		// 尝试读取文件
		data, err := fs.ReadFile(staticFS, path[1:])
		if err != nil {
			// 如果文件不存在，返回 index.html（支持 SPA 路由）
			data, err = fs.ReadFile(staticFS, "index.html")
			if err != nil {
				c.String(http.StatusNotFound, "404 page not found")
				return
			}
			c.Data(http.StatusOK, "text/html", data)
			return
		}

		// 根据文件扩展名设置 Content-Type
		contentType := "text/html"
		if len(path) > 3 {
			ext := path[len(path)-3:]
			switch ext {
			case ".js":
				contentType = "application/javascript"
			case "css":
				contentType = "text/css"
			case "png":
				contentType = "image/png"
			case "jpg", "peg":
				contentType = "image/jpeg"
			case "svg":
				contentType = "image/svg+xml"
			}
		}

		c.Data(http.StatusOK, contentType, data)
	})

	// 启动服务器
	log.Printf("Starting server on :%s", *port)
	if err := r.Run(":" + *port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
