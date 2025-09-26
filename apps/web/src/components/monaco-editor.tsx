"use client"

import { useEffect, useRef } from 'react'
import Editor from '@monaco-editor/react'

interface MonacoEditorProps {
  value: string
  onChange: (value: string) => void
  language: string
  height?: string
  theme?: string
  readOnly?: boolean
}

export function MonacoEditor({
  value,
  onChange,
  language,
  height = '100%',
  theme = 'vs-dark',
  readOnly = false
}: MonacoEditorProps) {
  const editorRef = useRef(null)

  const handleEditorDidMount = (editor: any, monaco: any) => {
    editorRef.current = editor

    // Configure YAML language features
    if (language === 'yaml') {
      monaco.languages.setLanguageConfiguration('yaml', {
        brackets: [
          ['{', '}'],
          ['[', ']'],
          ['(', ')']
        ],
        autoClosingPairs: [
          { open: '{', close: '}' },
          { open: '[', close: ']' },
          { open: '(', close: ')' },
          { open: '"', close: '"' },
          { open: "'", close: "'" }
        ],
        surroundingPairs: [
          { open: '{', close: '}' },
          { open: '[', close: ']' },
          { open: '(', close: ')' },
          { open: '"', close: '"' },
          { open: "'", close: "'" }
        ]
      })

      // Add schema validation for BackSaaS schemas
      monaco.languages.registerCompletionItemProvider('yaml', {
        provideCompletionItems: (model: any, position: any) => {
          const suggestions = [
            {
              label: 'name',
              kind: monaco.languages.CompletionItemKind.Property,
              insertText: 'name: ',
              documentation: 'Schema name'
            },
            {
              label: 'version',
              kind: monaco.languages.CompletionItemKind.Property,
              insertText: 'version: ',
              documentation: 'Schema version'
            },
            {
              label: 'entities',
              kind: monaco.languages.CompletionItemKind.Property,
              insertText: 'entities:\n  - name: \n    fields:\n      - name: \n        type: ',
              documentation: 'Schema entities'
            },
            {
              label: 'fields',
              kind: monaco.languages.CompletionItemKind.Property,
              insertText: 'fields:\n  - name: \n    type: ',
              documentation: 'Entity fields'
            },
            {
              label: 'relationships',
              kind: monaco.languages.CompletionItemKind.Property,
              insertText: 'relationships:\n  - name: \n    type: \n    target: ',
              documentation: 'Entity relationships'
            },
            {
              label: 'functions',
              kind: monaco.languages.CompletionItemKind.Property,
              insertText: 'functions:\n  - name: \n    type: \n    trigger: ',
              documentation: 'Entity functions'
            }
          ]
          return { suggestions }
        }
      })
    }

    // Set editor options
    editor.updateOptions({
      fontSize: 14,
      lineNumbers: 'on',
      roundedSelection: false,
      scrollBeyondLastLine: false,
      minimap: { enabled: false },
      wordWrap: 'on',
      automaticLayout: true,
      tabSize: 2,
      insertSpaces: true
    })
  }

  const handleEditorChange = (value: string | undefined) => {
    if (value !== undefined) {
      onChange(value)
    }
  }

  return (
    <div className="h-full border rounded-md overflow-hidden">
      <Editor
        height={height}
        language={language}
        theme={theme}
        value={value}
        onChange={handleEditorChange}
        onMount={handleEditorDidMount}
        options={{
          readOnly,
          selectOnLineNumbers: true,
          automaticLayout: true,
          scrollBeyondLastLine: false,
          minimap: { enabled: false },
          fontSize: 14,
          lineNumbers: 'on',
          wordWrap: 'on',
          tabSize: 2,
          insertSpaces: true
        }}
        loading={
          <div className="flex items-center justify-center h-full">
            <div className="text-sm text-muted-foreground">Loading editor...</div>
          </div>
        }
      />
    </div>
  )
}
