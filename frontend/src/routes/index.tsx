import '@/styles/animations.css';
import { Link, createFileRoute } from '@tanstack/react-router';
import { ClipboardList, Kanban, MessageSquareMore, Users, Zap } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { cn } from '../lib/utils';

export const Route = createFileRoute('/')({
  component: App,
});

function App() {
  const [isVisible, setIsVisible] = useState(false);
  const activityRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true);
        }
      },
      { threshold: 0.3 },
    );

    if (activityRef.current) {
      observer.observe(activityRef.current);
    }

    return () => observer.disconnect();
  }, []);

  return (
    <div className="relative min-h-screen bg-slate-50 dark:bg-slate-900">
      <header className="flex items-center justify-between border-b border-slate-200 px-6 py-6 lg:px-8 dark:border-slate-700">
        <div className="flex items-center space-x-3">
          <div className="flex h-7 w-7 items-center justify-center rounded bg-blue-600">
            <span className="text-sm font-bold text-white">T</span>
          </div>
          <span className="text-xl font-semibold text-slate-900 dark:text-slate-100">TaskFlow</span>
        </div>
        <div className="flex items-center space-x-6">
          <Link
            to="/login"
            className="text-sm font-medium text-slate-600 transition-colors hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100"
          >
            Log in
          </Link>
          <Link
            to="/sign-up"
            className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
          >
            Get Started
          </Link>
        </div>
      </header>

      <main>
        <section className="container mx-auto px-6 py-24 lg:px-8 lg:py-32">
          <div className="mx-auto max-w-6xl">
            <div className="mb-16 text-center">
              <h1 className="mb-8 text-5xl leading-tight font-bold text-balance text-slate-900 lg:text-7xl dark:text-slate-100">
                Real-time collaboration for small teams
              </h1>
              <p className="mx-auto mb-12 max-w-3xl text-xl leading-relaxed text-pretty text-slate-600 lg:text-2xl dark:text-slate-400">
                Manage projects and tasks with instant chat, real-time updates, and seamless team coordination.
                Everything happens instantly across your entire team.
              </p>
              <div className="flex flex-col justify-center gap-4 sm:flex-row">
                <Link
                  to="/sign-up"
                  className="rounded-lg bg-blue-600 px-8 py-4 text-lg font-medium text-white transition-colors hover:bg-blue-700"
                >
                  Start collaborating
                </Link>
                <Link
                  to="/login"
                  className="rounded-lg border border-slate-300 px-8 py-4 text-lg font-medium text-slate-900 transition-colors hover:bg-slate-100 dark:border-slate-600 dark:text-slate-100 dark:hover:bg-slate-800"
                >
                  View Features
                </Link>
              </div>
            </div>
          </div>
        </section>

        <section className="container mx-auto px-6 py-24 lg:px-8">
          <div className="mx-auto max-w-6xl">
            <div className="grid items-center gap-16 lg:grid-cols-2">
              <div>
                <div className="mb-8 flex items-center gap-3">
                  <Users />
                  <span className="text-sm font-medium tracking-wider text-slate-500 dark:text-slate-400">
                    COLLABORATION
                  </span>
                </div>
                <h2 className="mb-8 text-5xl leading-tight font-bold text-balance text-slate-900 lg:text-6xl dark:text-slate-100">
                  Instant. Synchronized. Collaborative.
                </h2>
                <p className="text-xl leading-relaxed text-pretty text-slate-600 dark:text-slate-400">
                  Experience instant updates, seamless chat integration, and automatic synchronization that keeps your
                  team in perfect sync.
                </p>
              </div>

              <div className="rounded-xl border border-slate-200 bg-white p-8 dark:border-slate-700 dark:bg-slate-800">
                <h3 className="mb-2 text-2xl font-semibold text-slate-900 dark:text-slate-100">
                  Stay synchronized in real-time.
                </h3>
                <p className="mb-8 text-lg text-slate-600 dark:text-slate-400">
                  Instant messaging, live task updates, and automatic notifications keep everyone on the same page.
                </p>
                <div ref={activityRef} className="space-y-4">
                  <div
                    className={cn(
                      'animate-slide-in animate-slide-in-1 rounded-lg bg-slate-100 p-6 dark:bg-slate-700',
                      isVisible && 'visible',
                    )}
                  >
                    <div className="mb-3 flex items-center gap-3">
                      <div className="animate-pulse-dot h-2 w-2 rounded-full bg-blue-500"></div>
                      <span className="font-medium text-slate-900 dark:text-slate-100">
                        Sarah moved 'Fix login bug' to In Progress
                      </span>
                    </div>
                    <div className="text-sm text-slate-600 dark:text-slate-400">2 seconds ago</div>
                  </div>
                  <div
                    className={cn(
                      'animate-slide-in animate-slide-in-2 rounded-lg bg-slate-100 p-6 dark:bg-slate-700',
                      isVisible && 'visible',
                    )}
                  >
                    <div className="mb-3 flex items-center gap-3">
                      <div className="animate-pulse-dot-green h-2 w-2 rounded-full bg-green-500"></div>
                      <span className="font-medium text-slate-900 dark:text-slate-100">
                        New message in #design-team
                      </span>
                    </div>
                    <div className="text-sm text-slate-600 dark:text-slate-400">12 seconds ago</div>
                  </div>
                  <div
                    className={cn(
                      'animate-slide-in animate-slide-in-3 rounded-lg bg-slate-100 p-6 dark:bg-slate-700',
                      isVisible && 'visible',
                    )}
                  >
                    <div className="mb-2 text-sm text-slate-600 dark:text-slate-400">
                      Mike completed 'Database migration'
                    </div>
                    <div className="text-sm text-slate-500 dark:text-slate-500">1 minute ago</div>
                  </div>
                  <div
                    className={cn(
                      'animate-slide-in animate-slide-in-4 rounded-lg bg-slate-100 p-6 dark:bg-slate-700',
                      isVisible && 'visible',
                    )}
                  >
                    <div className="mb-2 text-sm text-slate-600 dark:text-slate-400">
                      3 new tasks added to Mobile App project
                    </div>
                    <div className="text-sm text-slate-500 dark:text-slate-500">3 minutes ago</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="container mx-auto bg-white px-6 py-24 lg:px-8 dark:bg-slate-800">
          <div className="mx-auto max-w-6xl">
            <div className="mb-16 text-center">
              <div className="mb-8 flex items-center justify-center gap-3">
                <ClipboardList />
                <span className="text-sm font-medium tracking-wider text-slate-500 dark:text-slate-400">FEATURES</span>
              </div>
              <h2 className="mb-6 text-4xl leading-tight font-bold text-balance text-slate-900 lg:text-5xl dark:text-slate-100">
                Built for real-time team collaboration
              </h2>
              <p className="mx-auto max-w-3xl text-xl leading-relaxed text-pretty text-slate-600 dark:text-slate-400">
                Kanban boards, instant messaging, and live updates that work seamlessly together for effortless small
                team coordination.
              </p>
            </div>

            <div className="grid gap-8 md:grid-cols-2 lg:grid-cols-3">
              <div className="rounded-xl border border-slate-200 bg-slate-50 p-8 dark:border-slate-600 dark:bg-slate-700">
                <div className="mb-6 flex h-12 w-12 items-center justify-center rounded-lg bg-blue-100 dark:bg-blue-900">
                  <Kanban />
                </div>
                <h3 className="mb-3 text-xl font-semibold text-slate-900 dark:text-slate-100">Live Kanban Boards</h3>
                <p className="leading-relaxed text-slate-600 dark:text-slate-400">
                  Real-time kanban boards with instant updates. See task movements and changes as they happen across
                  your team.
                </p>
              </div>

              <div className="rounded-xl border border-slate-200 bg-slate-50 p-8 dark:border-slate-600 dark:bg-slate-700">
                <div className="mb-6 flex h-12 w-12 items-center justify-center rounded-lg bg-green-100 dark:bg-green-900">
                  <MessageSquareMore />
                </div>
                <h3 className="mb-3 text-xl font-semibold text-slate-900 dark:text-slate-100">Instant Team Chat</h3>
                <p className="leading-relaxed text-slate-600 dark:text-slate-400">
                  Lightning-fast messaging that's always in sync. Chat seamlessly integrates with your projects and task
                  discussions.
                </p>
              </div>

              <div className="rounded-xl border border-slate-200 bg-slate-50 p-8 dark:border-slate-600 dark:bg-slate-700">
                <div className="mb-6 flex h-12 w-12 items-center justify-center rounded-lg bg-purple-100 dark:bg-purple-900">
                  <Zap />
                </div>
                <h3 className="mb-3 text-xl font-semibold text-slate-900 dark:text-slate-100">Always in Sync</h3>
                <p className="leading-relaxed text-slate-600 dark:text-slate-400">
                  Every action instantly appears for all team members. No refresh needed, no delays, just seamless
                  collaboration.
                </p>
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>
  );
}
