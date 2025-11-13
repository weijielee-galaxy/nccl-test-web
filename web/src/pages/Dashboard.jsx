import { Box, Heading, Text, SimpleGrid, Stat, StatLabel, StatNumber, StatHelpText, useColorMode } from '@chakra-ui/react'
import { useEffect, useState } from 'react'
import { api } from '../api'

function Dashboard() {
  const { colorMode } = useColorMode()
  const [stats, setStats] = useState({
    totalIPs: 0,
    status: 'Loading...',
    lastUpdate: 'Never',
  })

  useEffect(() => {
    loadStats()
  }, [])

  const loadStats = async () => {
    const { data } = await api.getIPList()
    const health = await api.healthCheck()
    
    setStats({
      totalIPs: data?.count || 0,
      status: health.data?.status === 'ok' ? 'Healthy' : 'Unknown',
      lastUpdate: new Date().toLocaleString(),
    })
  }

  return (
    <Box>
      <Heading mb={2}>Dashboard</Heading>
      <Text color="gray.500" mb={8}>
        Overview of your NCCL Test system
      </Text>

      <SimpleGrid columns={{ base: 1, md: 3 }} spacing={6}>
        <Box
          p={6}
          bg={colorMode === 'dark' ? 'gray.700' : 'white'}
          borderRadius="lg"
          borderWidth="1px"
        >
          <Stat>
            <StatLabel>Total IP Addresses</StatLabel>
            <StatNumber fontSize="3xl">{stats.totalIPs}</StatNumber>
            <StatHelpText>Managed IPs</StatHelpText>
          </Stat>
        </Box>

        <Box
          p={6}
          bg={colorMode === 'dark' ? 'gray.700' : 'white'}
          borderRadius="lg"
          borderWidth="1px"
        >
          <Stat>
            <StatLabel>System Status</StatLabel>
            <StatNumber fontSize="3xl">{stats.status}</StatNumber>
            <StatHelpText>All systems operational</StatHelpText>
          </Stat>
        </Box>

        <Box
          p={6}
          bg={colorMode === 'dark' ? 'gray.700' : 'white'}
          borderRadius="lg"
          borderWidth="1px"
        >
          <Stat>
            <StatLabel>Last Update</StatLabel>
            <StatNumber fontSize="lg">{stats.lastUpdate}</StatNumber>
            <StatHelpText>Last data refresh</StatHelpText>
          </Stat>
        </Box>
      </SimpleGrid>
    </Box>
  )
}

export default Dashboard
