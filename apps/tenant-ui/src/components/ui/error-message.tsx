import React from 'react'
import { AlertTriangle } from 'lucide-react'
import { Button } from './button'

interface ErrorMessageProps {
  title: string
  message: string
  action?: {
    label: string
    onClick: () => void
  }
}

export function ErrorMessage({ title, message, action }: ErrorMessageProps) {
  return (
    <div className="text-center p-6 max-w-md mx-auto">
      <div className="mx-auto flex items-center justify-center w-12 h-12 rounded-full bg-red-100 mb-4">
        <AlertTriangle className="w-6 h-6 text-red-600" />
      </div>
      <h3 className="text-lg font-semibold text-gray-900 mb-2">{title}</h3>
      <p className="text-gray-600 mb-4">{message}</p>
      {action && (
        <Button onClick={action.onClick} variant="outline">
          {action.label}
        </Button>
      )}
    </div>
  )
}
