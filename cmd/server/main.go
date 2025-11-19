package main

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/weijielee-galaxy/nccl-test-web/internal/handlers"
	"github.com/weijielee-galaxy/nccl-test-web/web"
)

func main() {
	// 创建 Gin 路由
	r := gin.Default()

	// 注册API路由
	r.GET("/healthz", handlers.HealthCheck)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// IP列表管理接口（增删改查）
		v1.POST("/iplist", handlers.SaveIPList)     // 创建/新增
		v1.GET("/iplist", handlers.GetIPList)       // 查询
		v1.PUT("/iplist", handlers.UpdateIPList)    // 更新/修改
		v1.DELETE("/iplist", handlers.DeleteIPList) // 删除

		// NCCL 测试接口
		v1.GET("/nccl/defaults", handlers.GetNCCLTestDefaults)  // 获取默认参数
		v1.POST("/nccl/run", handlers.RunNCCLTest)              // 运行测试（一次性返回）
		v1.POST("/nccl/run-stream", handlers.RunNCCLTestStream) // 运行测试（流式返回）
		v1.POST("/nccl/stop", handlers.StopNCCLTest)            // 停止当前运行的测试
		v1.GET("/nccl/precheck", handlers.Precheck)             // 检查所有节点的 GPU 进程状态
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
	log.Println("Starting server on :8098")
	if err := r.Run(":8098"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
