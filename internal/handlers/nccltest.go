package handlers

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
)

// NCCLTestParams 定义 NCCL 测试参数
type NCCLTestParams struct {
	MapBy                  string      `json:"map_by" binding:"required"`
	OOBTCPInterface        string      `json:"oob_tcp_interface" binding:"required"`
	BTLTCPInterface        string      `json:"btl_tcp_interface" binding:"required"`
	NCCLIBGIDIndex         int         `json:"nccl_ib_gid_index" binding:"required"`
	NCCLMinChannels        int         `json:"nccl_min_channels" binding:"required"`
	NCCLIBQPSPerConnection int         `json:"nccl_ib_qps_per_connection" binding:"required"`
	TestSizeBegin          interface{} `json:"test_size_begin" binding:"required"` // 支持 int 或 string (如 "8K", "128M")
	TestSizeEnd            interface{} `json:"test_size_end" binding:"required"`   // 支持 int 或 string (如 "8K", "128M")
	Timeout                int         `json:"timeout"`                            // 超时时间（秒），0 表示不超时
	EnableDebug            bool        `json:"enable_debug"`                       // 是否启用 NCCL DEBUG
	NCCLDebugLevel         string      `json:"nccl_debug_level"`                   // NCCL DEBUG 级别: WARN, INFO, TRACE
}

// NCCLTestResponse 定义测试响应
type NCCLTestResponse struct {
	Status  string `json:"status"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
	Command string `json:"command"`
}

// RunNCCLTest 运行 NCCL 测试
func RunNCCLTest(c *gin.Context) {
	var params NCCLTestParams

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认超时（10分钟）
	timeout := 600
	if params.Timeout > 0 {
		timeout = params.Timeout
	}

	// 构建命令
	cmd := buildNCCLCommand(params)

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// 执行命令
	execCmd := exec.CommandContext(ctx, "bash", "-c", cmd)

	// 捕获输出
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	// 执行
	err := execCmd.Run()

	response := NCCLTestResponse{
		Command: cmd,
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			response.Status = "timeout"
			response.Error = fmt.Sprintf("Command timed out after %d seconds", timeout)
		} else {
			response.Status = "error"
			response.Error = err.Error()
		}
		response.Output = stdout.String() + "\n" + stderr.String()
		c.JSON(http.StatusOK, response)
		return
	}

	response.Status = "success"
	response.Output = stdout.String()
	if stderr.Len() > 0 {
		response.Output += "\n--- STDERR ---\n" + stderr.String()
	}

	c.JSON(http.StatusOK, response)
}

// RunNCCLTestStream 流式返回 NCCL 测试输出
func RunNCCLTestStream(c *gin.Context) {
	var params NCCLTestParams

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置响应头为流式输出
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 构建命令
	cmd := buildNCCLCommand(params)

	// 执行命令，合并 stdout 和 stderr
	execCmd := exec.Command("bash", "-c", cmd+" 2>&1")

	// 获取输出管道
	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		c.SSEvent("error", gin.H{"message": err.Error()})
		return
	}

	// 启动命令
	if err := execCmd.Start(); err != nil {
		c.SSEvent("error", gin.H{"message": err.Error()})
		return
	}

	// 发送命令信息
	c.SSEvent("command", cmd)
	c.Writer.Flush()

	// 读取并流式发送输出
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		c.SSEvent("output", line)
		c.Writer.Flush()
	}

	// 等待命令完成
	err = execCmd.Wait()
	if err != nil {
		c.SSEvent("error", gin.H{"message": err.Error()})
	} else {
		c.SSEvent("done", "Command completed successfully")
	}
	c.Writer.Flush()
}

// GetNCCLTestDefaults 获取默认参数
func GetNCCLTestDefaults(c *gin.Context) {
	defaults := NCCLTestParams{
		MapBy:                  "ppr:8:node",
		OOBTCPInterface:        "bond0",
		BTLTCPInterface:        "bond0",
		NCCLIBGIDIndex:         3,
		NCCLMinChannels:        32,
		NCCLIBQPSPerConnection: 8,
		TestSizeBegin:          1,
		TestSizeEnd:            1,
		Timeout:                600,
		EnableDebug:            false,
		NCCLDebugLevel:         "WARN",
	}

	c.JSON(http.StatusOK, defaults)
}

// buildNCCLCommand 构建 NCCL 测试命令
func buildNCCLCommand(params NCCLTestParams) string {
	// 基础命令
	cmd := fmt.Sprintf(`/usr/local/sihpc/bin/mpirun \
    --allow-run-as-root \
    --hostfile %s \
    --map-by %s \
    --mca oob_tcp_if_include %s \
    --mca pml ^ucx \
    --mca btl self,tcp \
    --mca btl_tcp_if_include %s \
    --mca routed direct \
    --mca plm_rsh_no_tree_spawn 1 \
    -x UCX_TLS=tcp`,
		DataDir+"/"+IPListFileName,
		params.MapBy,
		params.OOBTCPInterface,
		params.BTLTCPInterface,
	)

	// 设置 NCCL_DEBUG 环境变量
	if params.EnableDebug && params.NCCLDebugLevel != "" {
		// 启用 DEBUG 时使用指定的级别
		cmd += fmt.Sprintf(` \
    -x NCCL_DEBUG=%s`, params.NCCLDebugLevel)
	} else {
		// 未启用 DEBUG 时设置为 VERSION，抑制 INFO 级别日志
		cmd += ` \
    -x NCCL_DEBUG=VERSION`
	}

	// 添加其他 NCCL 参数和测试命令
	cmd += fmt.Sprintf(` \
    -x NCCL_IB_GID_INDEX=%d \
    -x NCCL_MIN_NCHANNELS=%d \
    -x NCCL_IB_QPS_PER_CONNECTION=%d \
    /usr/local/sihpc/libexec/nccl-tests/nccl_test -b %v -e %v`,
		params.NCCLIBGIDIndex,
		params.NCCLMinChannels,
		params.NCCLIBQPSPerConnection,
		params.TestSizeBegin,
		params.TestSizeEnd,
	)

	return cmd
}
