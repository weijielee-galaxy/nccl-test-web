import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Box } from '@chakra-ui/react'
import Layout from './components/Layout'
import IPManagement from './pages/IPManagement'
import Dashboard from './pages/Dashboard'
import NCCLTest from './pages/NCCLTest'

function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/iplist" element={<IPManagement />} />
          <Route path="/nccl-test" element={<NCCLTest />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  )
}

export default App
