import { execa } from "execa";
import { spawn, ChildProcess } from "node:child_process";
import path from "node:path";
import { waitOnHttp } from "./wait-on-http.js";

export interface BackendEnv {
  API_PORT: string;
  DB_DSN: string;
  PUBSUB_BROKERS: string;
  JWT_SECRET: string;
  ENV: string;
  CORS_ORIGINS: string;
}

export async function spawnBackend(
  backendDir: string,
  port: string,
  env: BackendEnv
): Promise<ChildProcess> {
  const binaryPath = path.join(backendDir, ".bin", "api-e2e");

  await execa("go", ["build", "-o", binaryPath, "./cmd/api"], {
    cwd: backendDir,
    stdio: "inherit",
  });

  const proc = spawn(binaryPath, {
    cwd: backendDir,
    env: { ...process.env, ...env },
    stdio: "inherit",
  });

  await waitOnHttp(`http://localhost:${port}/health`, 60_000);

  return proc;
}
