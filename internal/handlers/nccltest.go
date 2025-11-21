package handlers

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// 全局变量：当前运行的 NCCL 测试进程
var (
	currentCmd   *exec.Cmd
	currentMutex sync.Mutex
)

// NCCLTestParams 定义 NCCL 测试参数
type NCCLTestParams struct {
	MapBy                  string      `json:"map_by" binding:"required"`
	OOBTCPInterface        string      `json:"oob_tcp_interface" binding:"required"`
	BTLTCPInterface        string      `json:"btl_tcp_interface" binding:"required"`
	NCCLIBGIDIndex         int         `json:"nccl_ib_gid_index" binding:"required"`
	NCCLMinChannels        int         `json:"nccl_min_channels" binding:"required"`
	NCCLIBQPSPerConnection int         `json:"nccl_ib_qps_per_connection" binding:"required"`
	TestSizeBegin          interface{} `json:"test_size_begin"`                // 支持 int 或 string (如 "8K", "128M")，可选
	TestSizeEnd            interface{} `json:"test_size_end"`                  // 支持 int 或 string (如 "8K", "128M")，可选
	Iters                  int         `json:"iters"`                          // 迭代次数，可选
	Timeout                int         `json:"timeout"`                        // 超时时间（秒），0 表示不超时
	EnableDebug            bool        `json:"enable_debug"`                   // 是否启用 NCCL DEBUG
	NCCLDebugLevel         string      `json:"nccl_debug_level"`               // NCCL DEBUG 级别: WARN, INFO, TRACE
	IPListFile             string      `json:"iplist_file" binding:"required"` // IP列表文件名，必传
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

	// 设置进程组，以便能够杀死整个进程树
	execCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// 捕获输出
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	// 注册当前运行的命令
	currentMutex.Lock()
	currentCmd = execCmd
	currentMutex.Unlock()

	// 确保执行完成后清理
	defer func() {
		currentMutex.Lock()
		currentCmd = nil
		currentMutex.Unlock()
	}()

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

	// 异步保存历史数据
	SaveHistoryAsync(response.Output)

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

	fmt.Println(execCmd.String())

	// 设置进程组，以便能够杀死整个进程树
	execCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

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

	// 注册当前运行的命令
	currentMutex.Lock()
	currentCmd = execCmd
	currentMutex.Unlock()

	// 确保执行完成后清理
	defer func() {
		currentMutex.Lock()
		currentCmd = nil
		currentMutex.Unlock()
	}()

	// 发送命令信息
	c.SSEvent("command", cmd)
	c.Writer.Flush()

	// 创建 buffer 缓存所有输出
	var outputBuffer bytes.Buffer
	outputBuffer.WriteString(cmd + "\n\n")

	// 读取并流式发送输出
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		c.SSEvent("output", line)
		c.Writer.Flush()
		// 同时写入 buffer
		outputBuffer.WriteString(line + "\n")
	}

	// 等待命令完成
	err = execCmd.Wait()
	if err != nil {
		errorMsg := fmt.Sprintf("Error: %v", err)
		c.SSEvent("error", gin.H{"message": errorMsg})
		outputBuffer.WriteString("\n" + errorMsg)
	} else {
		c.SSEvent("done", "Command completed successfully")
	}
	c.Writer.Flush()

	// 异步保存历史数据
	SaveHistoryAsync(outputBuffer.String())
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
		Iters:                  20,
		Timeout:                600,
		EnableDebug:            false,
		NCCLDebugLevel:         "WARN",
		IPListFile:             "", // 必传，不提供默认值
	}

	c.JSON(http.StatusOK, defaults)
}

// StopNCCLTest 停止当前运行的 NCCL 测试
func StopNCCLTest(c *gin.Context) {
	currentMutex.Lock()
	defer currentMutex.Unlock()

	if currentCmd == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "no_task",
			"message": "No running NCCL test to stop",
		})
		return
	}

	if currentCmd.Process == nil {
		currentCmd = nil
		c.JSON(http.StatusOK, gin.H{
			"status":  "no_process",
			"message": "Command has no process",
		})
		return
	}

	// 杀死整个进程组（包括所有子进程）
	// 使用负的 PID 向整个进程组发送信号
	pgid := currentCmd.Process.Pid
	err := syscall.Kill(-pgid, syscall.SIGKILL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Failed to kill process group: %v", err),
		})
		return
	}

	currentCmd = nil
	c.JSON(http.StatusOK, gin.H{
		"status":  "stopped",
		"message": "NCCL test stopped successfully",
	})
}

// buildNCCLCommand 构建 NCCL 测试命令
func buildNCCLCommand(params NCCLTestParams) string {
	// 使用传入的 iplist 文件名
	// 构建 hostfile 路径
	hostfile := filepath.Join(DataDir, "iplist", params.IPListFile)

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
		hostfile,
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
    /usr/local/sihpc/libexec/nccl-tests/nccl_test`,
		params.NCCLIBGIDIndex,
		params.NCCLMinChannels,
		params.NCCLIBQPSPerConnection,
	)

	// 只在有测试大小参数时才添加 -b 和 -e
	if params.TestSizeBegin != nil && params.TestSizeBegin != "" {
		cmd += fmt.Sprintf(` -b %v`, params.TestSizeBegin)
	}
	if params.TestSizeEnd != nil && params.TestSizeEnd != "" {
		cmd += fmt.Sprintf(` -e %v`, params.TestSizeEnd)
	}

	// 只在有迭代次数参数时才添加 -n
	if params.Iters > 0 {
		cmd += fmt.Sprintf(` -n %d`, params.Iters)
	}

	return cmd
}
