import { useState } from 'react'
import styled from 'styled-components'
import FileExplorer from './components/FileExplorer'
import DependencyGraph from './components/DependencyGraph'

const AppContainer = styled.div`
  display: flex;
  height: 100vh;
  width: 100%;
`

function App() {
  const [selectedFile, setSelectedFile] = useState<string | undefined>(undefined)

  const handleFileSelect = (filePath: string) => {
    setSelectedFile(filePath)
  }

  return (
    <AppContainer>
      <FileExplorer onFileSelect={handleFileSelect} />
      <DependencyGraph selectedFile={selectedFile} />
    </AppContainer>
  )
}

export default App
