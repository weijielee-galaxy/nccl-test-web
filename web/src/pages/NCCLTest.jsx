import { useState, useEffect, useMemo, useRef } from 'react'
import {
  Box,
  Button,
  Heading,
  Text,
  VStack,
  HStack,
  FormControl,
  FormLabel,
  Input,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  useToast,
  useColorMode,
  Code,
  Spinner,
  Badge,
  Textarea,
  Grid,
  GridItem,
  Select,
  Collapse,
  useDisclosure,
  IconButton,
  SimpleGrid,
  Switch,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Tag,
  TagLabel,
  TagCloseButton,
  Wrap,
  WrapItem,
} from '@chakra-ui/react'
import { MdPlayArrow, MdExpandMore, MdExpandLess, MdEdit, MdCheck, MdClose, MdStop } from 'react-icons/md'
import { Line } from 'react-chartjs-2'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
} from 'chart.js'
import { api } from '../api'

// 注册 Chart.js 组件
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
)

// 测试大小单位
const SIZE_UNITS = [
  { value: '', label: 'Bytes' },
  { value: 'K', label: 'KB' },
  { value: 'M', label: 'MB' },
  { value: 'G', label: 'GB' },
]

// NCCL DEBUG 级别
const DEBUG_LEVELS = [
  { value: 'WARN', label: 'WARN' },
  { value: 'INFO', label: 'INFO' },
  { value: 'TRACE', label: 'TRACE' },
]

function NCCLTest() {
  const { colorMode } = useColorMode()
  const toast = useToast()
  const { isOpen: isParamsOpen, onToggle: onToggleParams } = useDisclosure({ defaultIsOpen: true })
  
  const [params, setParams] = useState(null)
  const [loading, setLoading] = useState(true)
  const [running, setRunning] = useState(false)
  const [prechecking, setPrechecking] = useState(false)
  const [result, setResult] = useState(null)
  
  // IP 地址编辑
  const [ipList, setIpList] = useState([])
  const [ipEditMode, setIpEditMode] = useState(false)
  const [ipInputValue, setIpInputValue] = useState('')
  
  // 测试大小的单位
  const [beginUnit, setBeginUnit] = useState('')
  const [endUnit, setEndUnit] = useState('M')  // 默认 MB
  
  // 测试大小和迭代次数的开关
  const [enableTestSize, setEnableTestSize] = useState(true)  // 默认启用测试大小
  const [enableIters, setEnableIters] = useState(true)  // 默认启用迭代次数
  
  // Stream 模式开关
  const [useStream, setUseStream] = useState(true)  // 默认开启
  
  // 用于自动滚动输出
  const outputEndRef = useRef(null)

  useEffect(() => {
    loadDefaults()
    loadIPList()
  }, [])
  
  // 自动滚动到底部
  useEffect(() => {
    if (result && result.output && outputEndRef.current) {
      outputEndRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [result?.output])

  const loadDefaults = async () => {
    setLoading(true)
    const { data, error } = await api.getNCCLDefaults()
    
    if (error) {
      toast({
        title: 'Error loading defaults',
        description: error,
        status: 'error',
        duration: 3000,
      })
    } else {
      setParams(data)
    }
    
    setLoading(false)
  }

  const loadIPList = async () => {
    const { data, error } = await api.getIPList()
    
    if (error) {
      console.error('Error loading IP list:', error)
    } else if (data && data.iplist) {
      // 处理两种可能的数据格式：数组或字符串
      let ips = []
      if (Array.isArray(data.iplist)) {
        // 如果是数组，直接使用
        ips = data.iplist.filter(ip => ip.trim() !== '')
      } else if (typeof data.iplist === 'string') {
        // 如果是字符串，按行分割
        ips = data.iplist.split('\n').filter(ip => ip.trim() !== '')
      }
      setIpList(ips)
      setIpInputValue(ips.join(', '))
    }
  }

  const handleIPEditStart = () => {
    setIpEditMode(true)
  }

  const handleIPEditCancel = () => {
    setIpEditMode(false)
    setIpInputValue(ipList.join(', '))
  }

  const handleIPDelete = async (indexToDelete) => {
    const updatedIps = ipList.filter((_, index) => index !== indexToDelete)
    
    if (updatedIps.length === 0) {
      // 如果删除后列表为空，删除整个 iplist
      const { error } = await api.deleteIPList()
      
      if (error) {
        toast({
          title: 'Error deleting IP',
          description: error,
          status: 'error',
          duration: 3000,
        })
      } else {
        setIpList([])
        setIpInputValue('')
        toast({
          title: 'Success',
          description: 'IP deleted successfully',
          status: 'success',
          duration: 2000,
        })
      }
    } else {
      // 更新列表
      const { error } = await api.updateIPList(updatedIps)
      
      if (error) {
        toast({
          title: 'Error deleting IP',
          description: error,
          status: 'error',
          duration: 3000,
        })
      } else {
        setIpList(updatedIps)
        setIpInputValue(updatedIps.join(', '))
        toast({
          title: 'Success',
          description: 'IP deleted successfully',
          status: 'success',
          duration: 2000,
        })
      }
    }
  }

  const handleIPEditSave = async () => {
    const ips = ipInputValue.split(',').map(ip => ip.trim()).filter(ip => ip !== '')
    
    if (ips.length === 0) {
      toast({
        title: 'Error',
        description: 'Please enter at least one IP address',
        status: 'error',
        duration: 3000,
      })
      return
    }

    // 传递数组格式，和 IP Management 页面保持一致
    const { error } = ipList.length === 0
      ? await api.createIPList(ips)
      : await api.updateIPList(ips)
    
    if (error) {
      toast({
        title: 'Error updating IP list',
        description: error,
        status: 'error',
        duration: 3000,
      })
    } else {
      setIpList(ips)
      setIpEditMode(false)
      toast({
        title: 'Success',
        description: 'IP list updated successfully',
        status: 'success',
        duration: 2000,
      })
    }
  }

  // 生成完整的命令脚本
  const generatedScript = useMemo(() => {
    if (!params) return ''
    
    let script = `/usr/local/sihpc/bin/mpirun \\
    --allow-run-as-root \\
    --hostfile ./data/iplist \\
    --map-by ${params.map_by} \\
    --mca oob_tcp_if_include ${params.oob_tcp_interface} \\
    --mca pml ^ucx \\
    --mca btl self,tcp \\
    --mca btl_tcp_if_include ${params.btl_tcp_interface} \\
    --mca routed direct \\
    --mca plm_rsh_no_tree_spawn 1 \\
    -x UCX_TLS=tcp`
    
    // 如果启用 DEBUG，添加 NCCL_DEBUG
    if (params.enable_debug && params.nccl_debug_level) {
      script += ` \\
    -x NCCL_DEBUG=${params.nccl_debug_level}`
    }
    
    script += ` \\
    -x NCCL_IB_GID_INDEX=${params.nccl_ib_gid_index} \\
    -x NCCL_MIN_NCHANNELS=${params.nccl_min_channels} \\
    -x NCCL_IB_QPS_PER_CONNECTION=${params.nccl_ib_qps_per_connection} \\
    /usr/local/sihpc/libexec/nccl-tests/nccl_test`
    
    // 只在启用时添加测试大小参数
    if (enableTestSize) {
      const beginSize = `${params.test_size_begin}${beginUnit}`
      const endSize = `${params.test_size_end}${endUnit}`
      script += ` -b ${beginSize} -e ${endSize}`
    }
    
    // 只在启用时添加迭代次数参数
    if (enableIters && params.iters) {
      script += ` -n ${params.iters}`
    }
    
    return script
  }, [params, beginUnit, endUnit, enableTestSize, enableIters])

  // 解析 NCCL 测试数据用于图表
  const chartData = useMemo(() => {
    if (!result?.output) return []
    
    const lines = result.output.split('\n')
    const data = []
    let parsing = false
    
    for (const line of lines) {
      // 检测表头开始
      if (line.includes('#       size') && line.includes('count')) {
        parsing = true
        continue
      }
      // 检测数据结束
      if (line.includes('# Out of bounds') || line.includes('# Avg bus bandwidth')) {
        break
      }
      // 解析数据行（不要 trim，直接按空格分割非空部分）
      if (parsing && line.trim() && !line.trim().startsWith('#')) {
        // 使用正则分割，过滤空字符串
        const parts = line.trim().split(/\s+/).filter(p => p)
        if (parts.length >= 12) {
          const size = parseInt(parts[0])
          const outAlgbw = parseFloat(parts[6])
          const outBusbw = parseFloat(parts[7])
          const inAlgbw = parseFloat(parts[10])
          const inBusbw = parseFloat(parts[11])
          
          // 只添加有效数据，过滤 size=0 的数据点
          if (!isNaN(size) && size > 0) {
            data.push({
              size,
              count: parseInt(parts[1]),
              type: parts[2],
              outAlgbw: isNaN(outAlgbw) ? 0 : outAlgbw,
              outBusbw: isNaN(outBusbw) ? 0 : outBusbw,
              inAlgbw: isNaN(inAlgbw) ? 0 : inAlgbw,
              inBusbw: isNaN(inBusbw) ? 0 : inBusbw,
            })
          }
        }
      }
    }
    
    console.log('Parsed chart data:', data)  // 调试用
    return data
  }, [result?.output])

  // Chart.js 配置
  const chartConfig = useMemo(() => {
    if (chartData.length === 0) return null
    
    const labels = chartData.map(d => {
      const size = d.size
      if (size >= 1048576) return `${(size / 1048576).toFixed(1)}M`
      if (size >= 1024) return `${(size / 1024).toFixed(0)}K`
      return size.toString()
    })
    
    return {
      labels,
      datasets: [
        {
          label: 'Out-of-place BusBW',
          data: chartData.map(d => d.outBusbw),
          borderColor: '#3182ce',
          backgroundColor: 'rgba(49, 130, 206, 0.1)',
          borderWidth: 2,
          pointRadius: 3,
          pointHoverRadius: 5,
        },
        {
          label: 'In-place BusBW',
          data: chartData.map(d => d.inBusbw),
          borderColor: '#38a169',
          backgroundColor: 'rgba(56, 161, 105, 0.1)',
          borderWidth: 2,
          pointRadius: 3,
          pointHoverRadius: 5,
        },
        {
          label: 'Out-of-place AlgBW',
          data: chartData.map(d => d.outAlgbw),
          borderColor: '#e53e3e',
          backgroundColor: 'rgba(229, 62, 62, 0.1)',
          borderWidth: 2,
          borderDash: [5, 5],
          pointRadius: 3,
          pointHoverRadius: 5,
        },
        {
          label: 'In-place AlgBW',
          data: chartData.map(d => d.inAlgbw),
          borderColor: '#d69e2e',
          backgroundColor: 'rgba(214, 158, 46, 0.1)',
          borderWidth: 2,
          borderDash: [5, 5],
          pointRadius: 3,
          pointHoverRadius: 5,
        },
      ],
    }
  }, [chartData])

  const chartOptions = useMemo(() => ({
    responsive: true,
    maintainAspectRatio: false,
    interaction: {
      mode: 'index',
      intersect: false,
    },
    plugins: {
      legend: {
        position: 'top',
        labels: {
          color: colorMode === 'dark' ? '#fff' : '#000',
          usePointStyle: true,
          padding: 15,
        },
        onClick: (e, legendItem, legend) => {
          // 默认的图例点击行为（切换显示/隐藏）
          const index = legendItem.datasetIndex
          const chart = legend.chart
          const meta = chart.getDatasetMeta(index)
          meta.hidden = meta.hidden === null ? !chart.data.datasets[index].hidden : null
          chart.update()
        },
      },
      title: {
        display: true,
        text: 'NCCL Bandwidth Performance',
        color: colorMode === 'dark' ? '#fff' : '#000',
        font: {
          size: 16,
          weight: 'bold',
        },
      },
      tooltip: {
        backgroundColor: colorMode === 'dark' ? '#2D3748' : '#fff',
        titleColor: colorMode === 'dark' ? '#fff' : '#000',
        bodyColor: colorMode === 'dark' ? '#fff' : '#000',
        borderColor: colorMode === 'dark' ? '#4A5568' : '#E2E8F0',
        borderWidth: 1,
      },
    },
    scales: {
      x: {
        title: {
          display: true,
          text: 'Size',
          color: colorMode === 'dark' ? '#fff' : '#000',
        },
        ticks: {
          color: colorMode === 'dark' ? '#fff' : '#000',
        },
        grid: {
          color: colorMode === 'dark' ? '#444' : '#ccc',
        },
      },
      y: {
        title: {
          display: true,
          text: 'Bandwidth (GB/s)',
          color: colorMode === 'dark' ? '#fff' : '#000',
        },
        beginAtZero: true,
        ticks: {
          color: colorMode === 'dark' ? '#fff' : '#000',
        },
        grid: {
          color: colorMode === 'dark' ? '#444' : '#ccc',
        },
      },
    },
  }), [colorMode])

  const runTest = async () => {
    setPrechecking(true)
    setResult({ status: 'prechecking', output: 'Checking node status...', command: '' })

    // 先执行 precheck，并传递 iplist_file 参数（必选）
    const iplistFile = params?.iplist_file
    if (!iplistFile) {
      setResult({
        status: 'error',
        output: '',
        error: 'iplist_file parameter is required',
        command: ''
      })
      toast({
        title: 'Error',
        description: 'iplist_file parameter is required',
        status: 'error',
        duration: 5000,
      })
      setPrechecking(false)
      return
    }

    const { data: precheckData, error: precheckError } = await api.precheck(iplistFile)
    
    setPrechecking(false)
    
    if (precheckError) {
      setResult({
        status: 'error',
        output: '',
        error: `Precheck failed: ${precheckError}`,
        command: ''
      })
      toast({
        title: 'Precheck failed',
        description: precheckError,
        status: 'error',
        duration: 5000,
      })
      return
    }

    // 检查是否有繁忙的节点
    if (precheckData && precheckData.busy_count > 0) {
      // 构建繁忙节点信息
      const busyInfo = precheckData.busy_nodes
        .map(node => `${node.ip}: ${node.process_count} process(es)`)
        .join('\n')
      
      setResult({
        status: 'busy',
        output: `Found ${precheckData.busy_count} busy node(s):\n\n${busyInfo}`,
        command: '',
        busyNodes: precheckData.busy_nodes
      })
      
      toast({
        title: 'Nodes are busy',
        description: `${precheckData.busy_count} node(s) have running GPU processes`,
        status: 'warning',
        duration: 5000,
      })
      return
    }

    // Precheck 通过，开始运行测试
    setRunning(true)
    setResult({ status: 'running', output: '', command: '' })

    // 构建带单位的参数
    const testParams = { ...params }
    
    // 只在启用时添加测试大小参数
    if (enableTestSize) {
      testParams.test_size_begin = `${params.test_size_begin}${beginUnit}`
      testParams.test_size_end = `${params.test_size_end}${endUnit}`
    } else {
      // 不启用时删除这些参数
      delete testParams.test_size_begin
      delete testParams.test_size_end
    }
    
    // 只在启用时添加迭代次数参数
    if (!enableIters) {
      delete testParams.iters
    }

    if (useStream) {
      // Stream 模式
      try {
        await api.runNCCLTestStream(testParams, {
          onCommand: (command) => {
            // 接收到命令
            console.log('[Stream] Command:', command)
            setResult(prev => ({ ...prev, command }))
          },
          onOutput: (line) => {
            // 接收到输出行
            console.log('[Stream] Output:', line)
            setResult(prev => ({
              ...prev,
              output: (prev.output || '') + line + '\n'
            }))
          },
          onComplete: () => {
            // 测试完成
            console.log('[Stream] Complete')
            setResult(prev => ({ ...prev, status: 'success' }))
            setRunning(false)
            toast({
              title: 'Test completed successfully',
              status: 'success',
              duration: 3000,
            })
          },
          onError: (error) => {
            // 发生错误
            console.error('[Stream] Error:', error)
            setResult(prev => ({ 
              ...prev, 
              status: 'error',
              error: error
            }))
            setRunning(false)
            toast({
              title: 'Error running test',
              description: error,
              status: 'error',
              duration: 5000,
            })
          }
        })
      } catch (error) {
        setResult(prev => ({ 
          ...prev, 
          status: 'error',
          error: error.message
        }))
        setRunning(false)
        toast({
          title: 'Error running test',
          description: error.message,
          status: 'error',
          duration: 5000,
        })
      }
    } else {
      // 普通模式
      const { data, error } = await api.runNCCLTest(testParams)

      if (error) {
        setResult({
          status: 'error',
          output: '',
          error: error,
          command: ''
        })
        toast({
          title: 'Error running test',
          description: error,
          status: 'error',
          duration: 5000,
        })
      } else {
        setResult(data)
        
        if (data.status === 'success') {
          toast({
            title: 'Test completed successfully',
            status: 'success',
            duration: 3000,
          })
        } else {
          toast({
            title: `Test ${data.status}`,
            description: data.error,
            status: 'warning',
            duration: 5000,
          })
        }
      }

      setRunning(false)
    }
  }

  const stopTest = async () => {
    // 调用后端 API 杀死进程
    // SSE 连接保持打开，等待接收进程结束后的最终输出
    const { data, error } = await api.stopNCCLTest()
    
    if (error) {
      toast({
        title: 'Failed to stop test',
        description: error,
        status: 'error',
        duration: 3000,
      })
      return
    }

    if (data.status === 'stopped') {
      toast({
        title: 'Stopping test...',
        description: 'NCCL test is being stopped',
        status: 'warning',
        duration: 2000,
      })
      // 不立即设置 setRunning(false)
      // 等待 SSE 的 onComplete 或 onError 回调来更新状态
    } else if (data.status === 'no_task') {
      setRunning(false)
      toast({
        title: 'No running test',
        description: 'There is no test running to stop',
        status: 'info',
        duration: 3000,
      })
    }
  }

  const updateParam = (key, value) => {
    setParams(prev => ({ ...prev, [key]: value }))
  }

  if (loading || !params) {
    return (
      <Box textAlign="center" py={12}>
        <Spinner size="xl" color="brand.500" />
        <Text mt={4}>Loading defaults...</Text>
      </Box>
    )
  }

  return (
    <Box position="relative">
      {/* 浮动的停止按钮 */}
      {running && (
        <Box
          position="fixed"
          bottom="20px"
          right="20px"
          zIndex="1000"
        >
          <Button
            leftIcon={<MdStop />}
            colorScheme="red"
            onClick={stopTest}
            size="lg"
            boxShadow="lg"
          >
            Stop Test
          </Button>
        </Box>
      )}
      
      <VStack spacing={4} align="stretch">
        <HStack justify="space-between" align="start">
          <Box>
            <Heading size="lg">NCCL Test</Heading>
            <Text color="gray.500" fontSize="sm">
              Configure and run NCCL tests
            </Text>
          </Box>
          <HStack spacing={3} flexWrap="wrap" justify="flex-end">
            <FormControl display="flex" alignItems="center" width="auto">
              <FormLabel htmlFor="stream-mode" mb={0} fontSize="sm" mr={2} whiteSpace="nowrap">
                Stream
              </FormLabel>
              <Switch
                id="stream-mode"
                size="sm"
                isChecked={useStream}
                onChange={(e) => setUseStream(e.target.checked)}
                isDisabled={running}
              />
            </FormControl>

            <Button
              leftIcon={(running || prechecking) ? <Spinner size="sm" /> : <MdPlayArrow />}
              colorScheme="brand"
              onClick={runTest}
              isLoading={running || prechecking}
              loadingText={prechecking ? "Prechecking..." : "Running..."}
              size="md"
              isDisabled={running || prechecking}
            >
              Run Test
            </Button>
            
            <Button
              variant="outline"
              onClick={loadDefaults}
              isDisabled={running || prechecking}
              size="md"
            >
              Reset
            </Button>

            <IconButton
              icon={isParamsOpen ? <MdExpandLess /> : <MdExpandMore />}
              onClick={onToggleParams}
              variant="outline"
              aria-label="Toggle parameters"
              size="md"
            />
          </HStack>
        </HStack>

      {/* 可折叠的参数配置 + 脚本预览 */}
      <Collapse in={isParamsOpen} animateOpacity>
        <Grid templateColumns="1fr 1fr" gap={4} mb={4}>
          {/* 左侧：紧凑的参数表单（两列布局） */}
          <GridItem>
            <Box
              p={4}
              bg={colorMode === 'dark' ? 'gray.700' : 'white'}
              borderRadius="lg"
              borderWidth="1px"
              h="full"
            >
              <Heading size="sm" mb={3}>Parameters</Heading>

              <SimpleGrid columns={2} spacing={2}>
                <FormControl>
                  <FormLabel fontSize="xs" mb={0.5}>Map By</FormLabel>
                  <Input
                    size="xs"
                    value={params.map_by}
                    onChange={(e) => updateParam('map_by', e.target.value)}
                  />
                </FormControl>

                <FormControl>
                  <FormLabel fontSize="xs" mb={0.5}>OOB TCP</FormLabel>
                  <Input
                    size="xs"
                    value={params.oob_tcp_interface}
                    onChange={(e) => updateParam('oob_tcp_interface', e.target.value)}
                  />
                </FormControl>

                <FormControl>
                  <FormLabel fontSize="xs" mb={0.5}>BTL TCP</FormLabel>
                  <Input
                    size="xs"
                    value={params.btl_tcp_interface}
                    onChange={(e) => updateParam('btl_tcp_interface', e.target.value)}
                  />
                </FormControl>

                <FormControl>
                  <FormLabel fontSize="xs" mb={0.5}>IB GID Index</FormLabel>
                  <NumberInput
                    size="xs"
                    value={params.nccl_ib_gid_index}
                    onChange={(_, val) => updateParam('nccl_ib_gid_index', val)}
                    min={0}
                  >
                    <NumberInputField />
                  </NumberInput>
                </FormControl>

                <FormControl>
                  <FormLabel fontSize="xs" mb={0.5}>Min Channels</FormLabel>
                  <NumberInput
                    size="xs"
                    value={params.nccl_min_channels}
                    onChange={(_, val) => updateParam('nccl_min_channels', val)}
                    min={1}
                  >
                    <NumberInputField />
                  </NumberInput>
                </FormControl>

                <FormControl>
                  <FormLabel fontSize="xs" mb={0.5}>QPS Per Conn</FormLabel>
                  <NumberInput
                    size="xs"
                    value={params.nccl_ib_qps_per_connection}
                    onChange={(_, val) => updateParam('nccl_ib_qps_per_connection', val)}
                    min={1}
                  >
                    <NumberInputField />
                  </NumberInput>
                </FormControl>

                <FormControl gridColumn="span 2">
                  <HStack justify="space-between">
                    <FormLabel fontSize="xs" mb={0}>Enable Test Size (-b / -e)</FormLabel>
                    <Switch
                      size="sm"
                      isChecked={enableTestSize}
                      onChange={(e) => setEnableTestSize(e.target.checked)}
                    />
                  </HStack>
                </FormControl>

                {enableTestSize && (
                  <>
                    <FormControl>
                      <FormLabel fontSize="xs" mb={0.5}>Begin Size</FormLabel>
                      <HStack spacing={1}>
                        <NumberInput
                          size="xs"
                          value={params.test_size_begin}
                          onChange={(_, val) => updateParam('test_size_begin', val)}
                          min={1}
                          flex={1}
                        >
                          <NumberInputField />
                        </NumberInput>
                        <Select
                          size="xs"
                          value={beginUnit}
                          onChange={(e) => setBeginUnit(e.target.value)}
                          w="70px"
                          minW="70px"
                        >
                          {SIZE_UNITS.map(unit => (
                            <option key={unit.value} value={unit.value}>{unit.label}</option>
                          ))}
                        </Select>
                      </HStack>
                    </FormControl>

                    <FormControl>
                      <FormLabel fontSize="xs" mb={0.5}>End Size</FormLabel>
                      <HStack spacing={1}>
                        <NumberInput
                          size="xs"
                          value={params.test_size_end}
                          onChange={(_, val) => updateParam('test_size_end', val)}
                          min={1}
                          flex={1}
                        >
                          <NumberInputField />
                        </NumberInput>
                        <Select
                          size="xs"
                          value={endUnit}
                          onChange={(e) => setEndUnit(e.target.value)}
                          w="70px"
                          minW="70px"
                        >
                          {SIZE_UNITS.map(unit => (
                            <option key={unit.value} value={unit.value}>{unit.label}</option>
                          ))}
                        </Select>
                      </HStack>
                    </FormControl>
                  </>
                )}

                <FormControl gridColumn="span 2">
                  <HStack justify="space-between">
                    <FormLabel fontSize="xs" mb={0}>Enable Iterations (-n)</FormLabel>
                    <Switch
                      size="sm"
                      isChecked={enableIters}
                      onChange={(e) => setEnableIters(e.target.checked)}
                    />
                  </HStack>
                </FormControl>

                {enableIters && (
                  <FormControl gridColumn="span 2">
                    <FormLabel fontSize="xs" mb={0.5}>Iterations</FormLabel>
                    <NumberInput
                      size="xs"
                      value={params.iters}
                      onChange={(_, val) => updateParam('iters', val)}
                      min={1}
                    >
                      <NumberInputField />
                    </NumberInput>
                  </FormControl>
                )}

                <FormControl gridColumn="span 2">
                  <FormLabel fontSize="xs" mb={0.5}>Timeout (seconds)</FormLabel>
                  <NumberInput
                    size="xs"
                    value={params.timeout}
                    onChange={(_, val) => updateParam('timeout', val)}
                    min={0}
                  >
                    <NumberInputField />
                  </NumberInput>
                </FormControl>

                <FormControl gridColumn="span 2">
                  <HStack justify="space-between">
                    <FormLabel fontSize="xs" mb={0}>Enable NCCL Debug</FormLabel>
                    <Switch
                      size="sm"
                      isChecked={params.enable_debug}
                      onChange={(e) => updateParam('enable_debug', e.target.checked)}
                    />
                  </HStack>
                </FormControl>

                {params.enable_debug && (
                  <FormControl gridColumn="span 2">
                    <FormLabel fontSize="xs" mb={0.5}>NCCL Debug Level</FormLabel>
                    <Select
                      size="xs"
                      value={params.nccl_debug_level}
                      onChange={(e) => updateParam('nccl_debug_level', e.target.value)}
                    >
                      {DEBUG_LEVELS.map(level => (
                        <option key={level.value} value={level.value}>{level.label}</option>
                      ))}
                    </Select>
                  </FormControl>
                )}
              </SimpleGrid>
            </Box>
          </GridItem>

          {/* 右侧：终端风格脚本预览 */}
          <GridItem>
            <Box
              p={3}
              bg={colorMode === 'dark' ? '#1a1a1a' : '#0d1117'}
              borderRadius="lg"
              borderWidth="1px"
              borderColor={colorMode === 'dark' ? 'gray.600' : 'gray.700'}
              h="full"
              position="relative"
            >
              {/* Terminal header */}
              <HStack spacing={1.5} mb={3}>
                <Box w="12px" h="12px" borderRadius="full" bg="#ff5f56" />
                <Box w="12px" h="12px" borderRadius="full" bg="#ffbd2e" />
                <Box w="12px" h="12px" borderRadius="full" bg="#27c93f" />
                <Text fontSize="xs" color="gray.500" ml={2}>run_nccl_test.sh</Text>
              </HStack>
              
              {/* Terminal content */}
              <Code
                display="block"
                whiteSpace="pre"
                p={3}
                borderRadius="md"
                fontSize="xs"
                fontFamily="'Fira Code', 'Consolas', 'Monaco', monospace"
                bg="transparent"
                color="#58a6ff"
                overflowX="auto"
                overflowY="auto"
                h="calc(100% - 36px)"
              >
                {generatedScript}
              </Code>
            </Box>
          </GridItem>
        </Grid>
      </Collapse>

      {/* IP 信息展示/编辑 - 独立组件 */}
      <Box
        p={4}
        bg={colorMode === 'dark' ? 'gray.700' : 'white'}
        borderRadius="lg"
        borderWidth="1px"
      >
        <HStack justify="space-between" mb={3}>
          <Heading size="sm">Test IP Addresses</Heading>
          {!ipEditMode ? (
            <IconButton
              size="sm"
              icon={<MdEdit />}
              onClick={handleIPEditStart}
              aria-label="Edit IP list"
              variant="ghost"
            />
          ) : (
            <HStack spacing={2}>
              <IconButton
                size="sm"
                icon={<MdCheck />}
                onClick={handleIPEditSave}
                aria-label="Save IP list"
                colorScheme="green"
                variant="ghost"
              />
              <IconButton
                size="sm"
                icon={<MdClose />}
                onClick={handleIPEditCancel}
                aria-label="Cancel edit"
                variant="ghost"
              />
            </HStack>
          )}
        </HStack>
        
        {ipEditMode ? (
          <Input
            size="sm"
            value={ipInputValue}
            onChange={(e) => setIpInputValue(e.target.value)}
            placeholder="Enter IP addresses separated by commas (e.g., 192.168.1.1, 192.168.1.2)"
            autoFocus
          />
        ) : (
          <Wrap spacing={2}>
            {ipList.length > 0 ? (
              ipList.map((ip, index) => (
                <WrapItem key={index}>
                  <Tag size="sm" colorScheme="blue" variant="subtle">
                    <TagLabel>{ip}</TagLabel>
                    <TagCloseButton onClick={() => handleIPDelete(index)} />
                  </Tag>
                </WrapItem>
              ))
            ) : (
              <Text fontSize="sm" color="gray.400">No IP addresses configured</Text>
            )}
          </Wrap>
        )}
      </Box>

      {/* Results - 主要焦点区域 */}
      {result && (
        <Box
          p={5}
          bg={colorMode === 'dark' ? 'gray.700' : 'white'}
          borderRadius="lg"
          borderWidth="1px"
        >
          <HStack justify="space-between" mb={4}>
            <Heading size="md">Test Result</Heading>
            <HStack spacing={3}>
              {(running || prechecking) && <Spinner size="sm" />}
              <Badge
                colorScheme={
                  result.status === 'success' ? 'green' :
                  result.status === 'running' ? 'blue' :
                  result.status === 'prechecking' ? 'cyan' :
                  result.status === 'timeout' ? 'yellow' :
                  result.status === 'busy' ? 'orange' :
                  'red'
                }
                fontSize="md"
                px={3}
                py={1}
              >
                {result.status.toUpperCase()}
              </Badge>
            </HStack>
          </HStack>

          <VStack align="stretch" spacing={4}>
            {result.error && (
              <Box>
                <Text fontWeight="semibold" mb={2} color="red.500">Error:</Text>
                <Code
                  display="block"
                  whiteSpace="pre-wrap"
                  p={3}
                  borderRadius="md"
                  fontSize="sm"
                  colorScheme="red"
                >
                  {result.error}
                </Code>
              </Box>
            )}

            <Tabs variant="enclosed" colorScheme="brand">
              <TabList>
                <Tab>Output</Tab>
                <Tab isDisabled={chartData.length === 0}>Data</Tab>
                <Tab isDisabled={chartData.length === 0}>Chart</Tab>
              </TabList>
              
              <TabPanels>
                {/* Output Tab */}
                <TabPanel p={0} pt={4}>
                  {result.status === 'busy' && result.busyNodes ? (
                    // 特殊显示繁忙节点
                    <VStack align="stretch" spacing={3}>
                      <Box
                        p={4}
                        bg="orange.50"
                        borderRadius="md"
                        borderWidth="1px"
                        borderColor="orange.300"
                      >
                        <Text fontWeight="bold" color="orange.700" mb={2}>
                          ⚠️ Cannot start test: {result.busyNodes.length} node(s) have running GPU processes
                        </Text>
                        <Text fontSize="sm" color="orange.600">
                          Please wait for these processes to complete or stop them manually.
                        </Text>
                      </Box>
                      
                      <VStack align="stretch" spacing={2}>
                        {result.busyNodes.map((node, idx) => (
                          <Box
                            key={idx}
                            p={3}
                            bg={colorMode === 'dark' ? 'gray.700' : 'white'}
                            borderRadius="md"
                            borderWidth="1px"
                            borderColor="orange.300"
                          >
                            <HStack justify="space-between">
                              <Text fontWeight="semibold">{node.ip}</Text>
                              <Badge colorScheme="orange">
                                {node.process_count} process{node.process_count > 1 ? 'es' : ''}
                              </Badge>
                            </HStack>
                          </Box>
                        ))}
                      </VStack>
                    </VStack>
                  ) : (
                    // 正常显示输出
                    <Box
                      p={3}
                      bg={colorMode === 'dark' ? 'gray.800' : 'gray.50'}
                      borderRadius="md"
                      borderWidth="1px"
                      maxH="600px"
                      overflowY="auto"
                      overflowX="auto"
                      width="0"
                      minWidth="100%"
                      fontFamily="'Fira Code', 'Consolas', 'Monaco', monospace"
                      fontSize="sm"
                      whiteSpace="pre"
                    >
                      {result.output || 'Waiting for output...'}
                      <div ref={outputEndRef} />
                    </Box>
                  )}
                </TabPanel>
                
                {/* Data Tab - 显示解析后的原始数据 */}
                <TabPanel p={0} pt={4}>
                  <Box
                    bg={colorMode === 'dark' ? 'gray.800' : 'gray.50'}
                    borderRadius="md"
                    borderWidth="1px"
                    p={4}
                    maxH="600px"
                    overflowY="auto"
                  >
                    <VStack align="stretch" spacing={2}>
                      <Text fontWeight="bold" fontSize="lg">
                        Parsed Data ({chartData.length} rows)
                      </Text>
                      {chartData.length > 0 ? (
                        <Box
                          as="pre"
                          fontSize="xs"
                          fontFamily="'Fira Code', 'Consolas', 'Monaco', monospace"
                          p={3}
                          bg={colorMode === 'dark' ? 'gray.900' : 'white'}
                          borderRadius="md"
                          overflowX="auto"
                        >
                          {JSON.stringify(chartData, null, 2)}
                        </Box>
                      ) : (
                        <Text color="gray.500">No data parsed yet</Text>
                      )}
                    </VStack>
                  </Box>
                </TabPanel>
                
                {/* Chart Tab */}
                <TabPanel p={0} pt={4}>
                  {chartConfig ? (
                    <Box
                      bg={colorMode === 'dark' ? 'gray.800' : 'gray.50'}
                      borderRadius="md"
                      borderWidth="1px"
                      p={4}
                    >
                      <HStack justify="flex-end" mb={2}>
                        <Text fontSize="sm" color="gray.500">
                          {chartData.length} data points • Click legend to toggle lines
                        </Text>
                      </HStack>
                      <Box height="500px">
                        <Line data={chartConfig} options={chartOptions} />
                      </Box>
                    </Box>
                  ) : (
                    <Text color="gray.500">No chart data available</Text>
                  )}
                </TabPanel>
              </TabPanels>
            </Tabs>
          </VStack>
        </Box>
      )}
      </VStack>
    </Box>
  )
}

export default NCCLTest
