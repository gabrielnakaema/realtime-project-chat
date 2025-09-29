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
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900 relative">
      <header className="flex items-center justify-between px-6 lg:px-8 py-6 border-b border-slate-200 dark:border-slate-700">
        <div className="flex items-center space-x-3">
          <div className="w-7 h-7 bg-blue-600 rounded flex items-center justify-center">
            <span className="text-white font-bold text-sm">T</span>
          </div>
          <span className="text-slate-900 dark:text-slate-100 font-semibold text-xl">TaskFlow</span>
        </div>
        <div className="flex items-center space-x-6">
          <Link
            to="/login"
            className="text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 transition-colors text-sm font-medium"
          >
            Log in
          </Link>
          <Link
            to="/sign-up"
            className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors text-sm"
          >
            Get Started
          </Link>
        </div>
      </header>

      <main>
        <section className="container mx-auto px-6 lg:px-8 py-24 lg:py-32">
          <div className="max-w-6xl mx-auto">
            <div className="text-center mb-16">
              <h1 className="text-5xl lg:text-7xl font-bold text-slate-900 dark:text-slate-100 mb-8 leading-tight text-balance">
                Real-time collaboration for small teams
              </h1>
              <p className="text-xl lg:text-2xl text-slate-600 dark:text-slate-400 mb-12 leading-relaxed max-w-3xl mx-auto text-pretty">
                Manage projects and tasks with instant chat, real-time updates, and seamless team coordination.
                Everything happens instantly across your entire team.
              </p>
              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <Link
                  to="/sign-up"
                  className="bg-blue-600 hover:bg-blue-700 text-white px-8 py-4 rounded-lg font-medium transition-colors text-lg"
                >
                  Start collaborating
                </Link>
                <Link
                  to="/login"
                  className="border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-slate-100 px-8 py-4 rounded-lg font-medium hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors text-lg"
                >
                  View Features
                </Link>
              </div>
            </div>
          </div>
        </section>

        <section className="container mx-auto px-6 lg:px-8 py-24">
          <div className="max-w-6xl mx-auto">
            <div className="grid lg:grid-cols-2 gap-16 items-center">
              <div>
                <div className="flex items-center gap-3 mb-8">
                  <Users />
                  <span className="text-slate-500 dark:text-slate-400 text-sm font-medium tracking-wider">
                    COLLABORATION
                  </span>
                </div>
                <h2 className="text-5xl lg:text-6xl font-bold text-slate-900 dark:text-slate-100 mb-8 leading-tight text-balance">
                  Instant. Synchronized. Collaborative.
                </h2>
                <p className="text-slate-600 dark:text-slate-400 text-xl leading-relaxed text-pretty">
                  Experience instant updates, seamless chat integration, and automatic synchronization that keeps your
                  team in perfect sync.
                </p>
              </div>

              <div className="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl p-8">
                <h3 className="text-slate-900 dark:text-slate-100 text-2xl font-semibold mb-2">
                  Stay synchronized in real-time.
                </h3>
                <p className="text-slate-600 dark:text-slate-400 text-lg mb-8">
                  Instant messaging, live task updates, and automatic notifications keep everyone on the same page.
                </p>
                <div ref={activityRef} className="space-y-4">
                  <div
                    className={cn(
                      'bg-slate-100 dark:bg-slate-700 rounded-lg p-6 animate-slide-in animate-slide-in-1',
                      isVisible && 'visible',
                    )}
                  >
                    <div className="flex items-center gap-3 mb-3">
                      <div className="w-2 h-2 bg-blue-500 rounded-full animate-pulse-dot"></div>
                      <span className="text-slate-900 dark:text-slate-100 font-medium">
                        Sarah moved 'Fix login bug' to In Progress
                      </span>
                    </div>
                    <div className="text-slate-600 dark:text-slate-400 text-sm">2 seconds ago</div>
                  </div>
                  <div
                    className={cn(
                      'bg-slate-100 dark:bg-slate-700 rounded-lg p-6 animate-slide-in animate-slide-in-2',
                      isVisible && 'visible',
                    )}
                  >
                    <div className="flex items-center gap-3 mb-3">
                      <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse-dot-green"></div>
                      <span className="text-slate-900 dark:text-slate-100 font-medium">
                        New message in #design-team
                      </span>
                    </div>
                    <div className="text-slate-600 dark:text-slate-400 text-sm">12 seconds ago</div>
                  </div>
                  <div
                    className={cn(
                      'bg-slate-100 dark:bg-slate-700 rounded-lg p-6 animate-slide-in animate-slide-in-3',
                      isVisible && 'visible',
                    )}
                  >
                    <div className="text-slate-600 dark:text-slate-400 text-sm mb-2">
                      Mike completed 'Database migration'
                    </div>
                    <div className="text-slate-500 dark:text-slate-500 text-sm">1 minute ago</div>
                  </div>
                  <div
                    className={cn(
                      'bg-slate-100 dark:bg-slate-700 rounded-lg p-6 animate-slide-in animate-slide-in-4',
                      isVisible && 'visible',
                    )}
                  >
                    <div className="text-slate-600 dark:text-slate-400 text-sm mb-2">
                      3 new tasks added to Mobile App project
                    </div>
                    <div className="text-slate-500 dark:text-slate-500 text-sm">3 minutes ago</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="container mx-auto px-6 lg:px-8 py-24 bg-white dark:bg-slate-800">
          <div className="max-w-6xl mx-auto">
            <div className="text-center mb-16">
              <div className="flex items-center justify-center gap-3 mb-8">
                <ClipboardList />
                <span className="text-slate-500 dark:text-slate-400 text-sm font-medium tracking-wider">FEATURES</span>
              </div>
              <h2 className="text-4xl lg:text-5xl font-bold text-slate-900 dark:text-slate-100 mb-6 leading-tight text-balance">
                Built for real-time team collaboration
              </h2>
              <p className="text-slate-600 dark:text-slate-400 text-xl leading-relaxed max-w-3xl mx-auto text-pretty">
                Kanban boards, instant messaging, and live updates that work seamlessly together for effortless small
                team coordination.
              </p>
            </div>

            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
              <div className="bg-slate-50 dark:bg-slate-700 border border-slate-200 dark:border-slate-600 rounded-xl p-8">
                <div className="w-12 h-12 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center mb-6">
                  <Kanban />
                </div>
                <h3 className="text-slate-900 dark:text-slate-100 text-xl font-semibold mb-3">Live Kanban Boards</h3>
                <p className="text-slate-600 dark:text-slate-400 leading-relaxed">
                  Real-time kanban boards with instant updates. See task movements and changes as they happen across
                  your team.
                </p>
              </div>

              <div className="bg-slate-50 dark:bg-slate-700 border border-slate-200 dark:border-slate-600 rounded-xl p-8">
                <div className="w-12 h-12 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center mb-6">
                  <MessageSquareMore />
                </div>
                <h3 className="text-slate-900 dark:text-slate-100 text-xl font-semibold mb-3">Instant Team Chat</h3>
                <p className="text-slate-600 dark:text-slate-400 leading-relaxed">
                  Lightning-fast messaging that's always in sync. Chat seamlessly integrates with your projects and task
                  discussions.
                </p>
              </div>

              <div className="bg-slate-50 dark:bg-slate-700 border border-slate-200 dark:border-slate-600 rounded-xl p-8">
                <div className="w-12 h-12 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center mb-6">
                  <Zap />
                </div>
                <h3 className="text-slate-900 dark:text-slate-100 text-xl font-semibold mb-3">Always in Sync</h3>
                <p className="text-slate-600 dark:text-slate-400 leading-relaxed">
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
