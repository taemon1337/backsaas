export default function TestPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md mx-auto bg-white rounded-lg shadow-lg p-8">
        <h1 className="text-2xl font-bold text-center text-gray-900 mb-4">
          🎉 Tenant UI Test Page
        </h1>
        <div className="space-y-3 text-sm text-gray-600">
          <p><strong>Gateway Routing:</strong> ✅ Working</p>
          <p><strong>Next.js App:</strong> ✅ Running</p>
          <p><strong>Base Path:</strong> /ui</p>
          <p><strong>Current URL:</strong> {typeof window !== 'undefined' ? window.location.href : 'SSR'}</p>
        </div>
        <div className="mt-6 text-center">
          <a 
            href="/ui" 
            className="inline-block bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition-colors"
          >
            Go to Dashboard
          </a>
        </div>
      </div>
    </div>
  )
}
