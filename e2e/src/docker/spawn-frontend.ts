import { execa } from "execa";
import { spawn, ChildProcess } from "node:child_process";
import { waitOnHttp } from "./wait-on-http.js";

export async function spawnFrontend(
  frontendDir: string,
  port: string,
  apiUrl: string
): Promise<ChildProcess> {
  if (!process.env.SKIP_FRONTEND_BUILD) {
    await execa("pnpm", ["build"], {
      cwd: frontendDir,
      env: { ...process.env, VITE_API_URL: apiUrl },
      stdio: "inherit",
    });
  }

  const proc = spawn(
    "pnpm",
    ["exec", "vite", "preview", "--port", port, "--strictPort"],
    {
      cwd: frontendDir,
      stdio: "inherit",
    }
  );

  await waitOnHttp(`http://localhost:${port}`, 30_000);

  return proc;
}
