import { createRootRoute, Outlet } from '@tanstack/react-router';
import { lazy, Suspense } from 'react';

const isDev = import.meta.env.DEV;

const devComponents = isDev
  ? {
      TanStackDevtools: lazy(() =>
        import('@tanstack/react-devtools').then((module) => ({ default: module.TanStackDevtools })),
      ),
      TanStackRouterDevtoolsPanel: lazy(() =>
        import('@tanstack/react-router-devtools').then((module) => ({ default: module.TanStackRouterDevtoolsPanel })),
      ),
      ReactQueryDevtoolsPanel: lazy(() =>
        import('@tanstack/react-query-devtools').then((module) => ({ default: module.ReactQueryDevtoolsPanel })),
      ),
    }
  : { TanStackDevtools: null, TanStackRouterDevtoolsPanel: null, ReactQueryDevtoolsPanel: null };

export const Route = createRootRoute({
  component: () => (
    <>
      <Outlet />
      {isDev && devComponents.TanStackDevtools && (
        <Suspense fallback={null}>
          (
          <devComponents.TanStackDevtools
            config={{
              position: 'bottom-left',
              defaultOpen: false,
            }}
            plugins={[
              {
                name: 'Tanstack Router',
                render: <devComponents.TanStackRouterDevtoolsPanel />,
              },
              {
                name: 'Tanstack Query',
                render: <devComponents.ReactQueryDevtoolsPanel />,
              },
            ]}
          />
          )
        </Suspense>
      )}
    </>
  ),
});
