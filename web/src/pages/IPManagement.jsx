import { useState, useEffect } from 'react'
import {
  Box,
  Button,
  Heading,
  Text,
  VStack,
  HStack,
  Input,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  IconButton,
  useToast,
  useColorMode,
  Spinner,
  Center,
  Textarea,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  useDisclosure,
  FormControl,
  FormLabel,
} from '@chakra-ui/react'
import { DeleteIcon, EditIcon, AddIcon } from '@chakra-ui/icons'
import { api } from '../api'

function IPManagement() {
  const { colorMode } = useColorMode()
  const toast = useToast()
  const { isOpen, onOpen, onClose } = useDisclosure()
  
  const [ipList, setIpList] = useState([])
  const [loading, setLoading] = useState(true)
  const [newIP, setNewIP] = useState('')
  const [searchTerm, setSearchTerm] = useState('')
  const [batchEdit, setBatchEdit] = useState('')

  useEffect(() => {
    loadIPList()
  }, [])

  const loadIPList = async () => {
    setLoading(true)
    const { data, error } = await api.getIPList()
    
    if (error) {
      toast({
        title: 'Error loading IP list',
        description: error,
        status: 'error',
        duration: 3000,
      })
    } else {
      setIpList(data.iplist || [])
    }
    
    setLoading(false)
  }

  const addIP = async () => {
    if (!newIP.trim()) {
      toast({
        title: 'Please enter an IP address',
        status: 'warning',
        duration: 2000,
      })
      return
    }

    const updatedList = [...ipList, newIP.trim()]
    const { error } = ipList.length === 0 
      ? await api.createIPList(updatedList)
      : await api.updateIPList(updatedList)

    if (error) {
      toast({
        title: 'Error adding IP',
        description: error,
        status: 'error',
        duration: 3000,
      })
    } else {
      toast({
        title: 'IP added successfully',
        status: 'success',
        duration: 2000,
      })
      setNewIP('')
      loadIPList()
    }
  }

  const deleteIP = async (index) => {
    const updatedList = ipList.filter((_, i) => i !== index)
    
    const { error } = await api.updateIPList(updatedList)

    if (error) {
      toast({
        title: 'Error deleting IP',
        description: error,
        status: 'error',
        duration: 3000,
      })
    } else {
      toast({
        title: 'IP deleted successfully',
        status: 'success',
        duration: 2000,
      })
      loadIPList()
    }
  }

  const deleteAll = async () => {
    if (!confirm('Are you sure you want to delete all IPs?')) return

    const { error } = await api.deleteIPList()

    if (error) {
      toast({
        title: 'Error deleting all IPs',
        description: error,
        status: 'error',
        duration: 3000,
      })
    } else {
      toast({
        title: 'All IPs deleted successfully',
        status: 'success',
        duration: 2000,
      })
      loadIPList()
    }
  }

  const handleBatchEdit = async () => {
    const newList = batchEdit
      .split('\n')
      .map(ip => ip.trim())
      .filter(ip => ip !== '')

    const { error } = ipList.length === 0
      ? await api.createIPList(newList)
      : await api.updateIPList(newList)

    if (error) {
      toast({
        title: 'Error updating IP list',
        description: error,
        status: 'error',
        duration: 3000,
      })
    } else {
      toast({
        title: 'IP list updated successfully',
        status: 'success',
        duration: 2000,
      })
      onClose()
      loadIPList()
    }
  }

  const openBatchEdit = () => {
    setBatchEdit(ipList.join('\n'))
    onOpen()
  }

  const filteredIPs = searchTerm
    ? ipList.filter(ip => ip.toLowerCase().includes(searchTerm.toLowerCase()))
    : ipList

  return (
    <Box>
      <HStack justify="space-between" mb={2}>
        <Box>
          <Heading>IP Management</Heading>
          <Text color="gray.500">Manage your IP addresses for NCCL testing</Text>
        </Box>
        <HStack>
          <Button
            leftIcon={<EditIcon />}
            colorScheme="brand"
            variant="outline"
            onClick={openBatchEdit}
          >
            Batch Edit
          </Button>
          {ipList.length > 0 && (
            <Button
              leftIcon={<DeleteIcon />}
              colorScheme="red"
              variant="outline"
              onClick={deleteAll}
            >
              Delete All
            </Button>
          )}
        </HStack>
      </HStack>

      {/* Add New IP */}
      <Box
        mt={8}
        p={6}
        bg={colorMode === 'dark' ? 'gray.700' : 'white'}
        borderRadius="lg"
        borderWidth="1px"
      >
        <Heading size="md" mb={4}>Add New IP Address</Heading>
        <HStack>
          <Input
            placeholder="Enter IP address (e.g., 192.168.1.1)"
            value={newIP}
            onChange={(e) => setNewIP(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && addIP()}
          />
          <Button
            leftIcon={<AddIcon />}
            colorScheme="brand"
            onClick={addIP}
            minW="120px"
          >
            Add IP
          </Button>
        </HStack>
      </Box>

      {/* IP List */}
      <Box
        mt={6}
        bg={colorMode === 'dark' ? 'gray.700' : 'white'}
        borderRadius="lg"
        borderWidth="1px"
      >
        <Box p={6} borderBottomWidth="1px">
          <HStack justify="space-between">
            <Heading size="md">IP List ({ipList.length})</Heading>
            <Input
              placeholder="Search IP addresses..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              maxW="300px"
            />
          </HStack>
        </Box>

        {loading ? (
          <Center py={12}>
            <Spinner size="xl" color="brand.500" />
          </Center>
        ) : filteredIPs.length === 0 ? (
          <Center py={12}>
            <VStack spacing={2}>
              <Text fontSize="4xl">📭</Text>
              <Heading size="md" color="gray.500">
                {searchTerm ? 'No matching IPs found' : 'No IP addresses yet'}
              </Heading>
              <Text color="gray.400">
                {searchTerm ? 'Try a different search term' : 'Add your first IP address to get started'}
              </Text>
            </VStack>
          </Center>
        ) : (
          <Table variant="simple">
            <Thead>
              <Tr>
                <Th>#</Th>
                <Th>IP Address</Th>
                <Th textAlign="right">Actions</Th>
              </Tr>
            </Thead>
            <Tbody>
              {filteredIPs.map((ip, index) => {
                const originalIndex = ipList.indexOf(ip)
                return (
                  <Tr key={originalIndex}>
                    <Td>{originalIndex + 1}</Td>
                    <Td fontFamily="mono" fontWeight="semibold">{ip}</Td>
                    <Td textAlign="right">
                      <IconButton
                        icon={<DeleteIcon />}
                        colorScheme="red"
                        variant="ghost"
                        size="sm"
                        onClick={() => deleteIP(originalIndex)}
                        aria-label="Delete IP"
                      />
                    </Td>
                  </Tr>
                )
              })}
            </Tbody>
          </Table>
        )}
      </Box>

      {/* Batch Edit Modal */}
      <Modal isOpen={isOpen} onClose={onClose} size="xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Batch Edit IP Addresses</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <FormControl>
              <FormLabel>Enter one IP address per line</FormLabel>
              <Textarea
                value={batchEdit}
                onChange={(e) => setBatchEdit(e.target.value)}
                placeholder="192.168.1.1&#10;192.168.1.2&#10;192.168.1.3"
                rows={15}
                fontFamily="mono"
              />
            </FormControl>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onClose}>
              Cancel
            </Button>
            <Button colorScheme="brand" onClick={handleBatchEdit}>
              Save Changes
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Box>
  )
}

export default IPManagement
