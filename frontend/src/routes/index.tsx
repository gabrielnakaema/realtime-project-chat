import { createFileRoute, Link } from '@tanstack/react-router'

export const Route = createFileRoute('/')({
  component: App,
})

function App() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
      <div className="container mx-auto px-4 py-16">
        <div className="text-center mb-16">
          <h1 className="text-5xl font-bold text-slate-900 dark:text-slate-100 mb-6 text-balance">
            TaskFlow
          </h1>
          <p className="text-xl text-slate-600 dark:text-slate-400 mb-8 text-pretty max-w-2xl mx-auto">
            Streamline your team's workflow with powerful Kanban boards,
            real-time collaboration, and intelligent project management.
          </p>

          <div>
            <Link
              to="/login"
              className="bg-blue-500 text-white px-4 py-2 rounded-md transition-colors hover:bg-blue-600"
            >
              Login
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}
