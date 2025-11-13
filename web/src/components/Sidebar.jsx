import { Box, VStack, Link as ChakraLink, Icon, useColorMode } from '@chakra-ui/react'
import { Link as RouterLink, useLocation } from 'react-router-dom'
import { MdDashboard, MdList, MdPlayArrow } from 'react-icons/md'

const navItems = [
  { name: 'Dashboard', path: '/', icon: MdDashboard },
  { name: 'IP Management', path: '/iplist', icon: MdList },
  { name: 'NCCL Test', path: '/nccl-test', icon: MdPlayArrow },
]

function Sidebar() {
  const location = useLocation()
  const { colorMode } = useColorMode()

  return (
    <Box
      as="nav"
      pos="fixed"
      left={0}
      top={0}
      w={60}
      h="full"
      bg={colorMode === 'dark' ? 'gray.900' : 'gray.100'}
      borderRightWidth="1px"
      display={{ base: 'none', md: 'block' }}
    >
      <VStack align="stretch" spacing={2} p={4}>
        <Box mb={8} px={4} py={6}>
          <Box
            w={12}
            h={12}
            bg="brand.500"
            borderRadius="lg"
            display="flex"
            alignItems="center"
            justifyContent="center"
            color="white"
            fontWeight="bold"
            fontSize="xl"
            mb={2}
          >
            N
          </Box>
          <Box fontSize="lg" fontWeight="bold">
            NCCL Test
          </Box>
        </Box>

        {navItems.map((item) => {
          const isActive = location.pathname === item.path
          
          return (
            <ChakraLink
              as={RouterLink}
              to={item.path}
              key={item.path}
              display="flex"
              alignItems="center"
              px={4}
              py={3}
              borderRadius="lg"
              bg={isActive ? 'brand.500' : 'transparent'}
              color={isActive ? 'white' : undefined}
              _hover={{
                bg: isActive ? 'brand.600' : colorMode === 'dark' ? 'gray.700' : 'gray.200',
                textDecoration: 'none',
              }}
              fontWeight={isActive ? 'semibold' : 'normal'}
            >
              <Icon as={item.icon} boxSize={5} mr={3} />
              {item.name}
            </ChakraLink>
          )
        })}
      </VStack>
    </Box>
  )
}

export default Sidebar
