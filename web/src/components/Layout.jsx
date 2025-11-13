import { Box, Flex, IconButton, Heading, useColorMode, Container } from '@chakra-ui/react'
import { MoonIcon, SunIcon } from '@chakra-ui/icons'
import Sidebar from './Sidebar'

function Layout({ children }) {
  const { colorMode, toggleColorMode } = useColorMode()

  return (
    <Flex minH="100vh">
      <Sidebar />
      
      <Box flex="1" ml={{ base: 0, md: 60 }}>
        {/* Header */}
        <Flex
          as="header"
          align="center"
          justify="space-between"
          px={8}
          py={4}
          borderBottomWidth="1px"
          bg={colorMode === 'dark' ? 'gray.700' : 'white'}
        >
          <Heading size="md">NCCL Test Platform</Heading>
          
          <IconButton
            icon={colorMode === 'light' ? <MoonIcon /> : <SunIcon />}
            onClick={toggleColorMode}
            variant="ghost"
            aria-label="Toggle color mode"
          />
        </Flex>

        {/* Main Content */}
        <Container maxW="100%" px={8} py={8}>
          {children}
        </Container>
      </Box>
    </Flex>
  )
}

export default Layout
