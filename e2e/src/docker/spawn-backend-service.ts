import { execa } from "execa";
import { spawn, ChildProcess } from "node:child_process";
import path from "node:path";
import { waitOnHttp } from "./wait-on-http.js";

export async function spawnBackendService(
  backendDir: string,
  name: string,
  command: string,
  port: string,
  env: NodeJS.ProcessEnv
): Promise<ChildProcess> {
  const binaryPath = path.join(backendDir, ".bin", `${name}-e2e`);

  await execa("go", ["build", "-o", binaryPath, command], {
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
