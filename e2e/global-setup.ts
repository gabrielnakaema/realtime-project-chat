import path from "node:path";
import { fileURLToPath } from "node:url";
import { ChildProcess } from "node:child_process";
import dotenv from "dotenv";

import { startInfra } from "./src/docker/start-infra.js";
import { runMigrations } from "./src/docker/run-migrations.js";
import { spawnBackend } from "./src/docker/spawn-backend.js";
import { spawnFrontend } from "./src/docker/spawn-frontend.js";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

dotenv.config({ path: path.join(__dirname, ".env.e2e") });

function killProcess(proc: ChildProcess, graceMs = 5_000): Promise<void> {
  return new Promise((resolve) => {
    if (proc.exitCode !== null || proc.signalCode !== null) {
      resolve();
      return;
    }

    const timer = setTimeout(() => {
      proc.kill("SIGKILL");
    }, graceMs);

    proc.once("exit", () => {
      clearTimeout(timer);
      resolve();
    });

    proc.kill("SIGTERM");
  });
}

export default async function globalSetup(): Promise<() => Promise<void>> {
  const backendPort = process.env.E2E_BACKEND_PORT ?? "4333";
  const frontendPort = process.env.E2E_FRONTEND_PORT ?? "4173";
  const corsOrigin =
    process.env.E2E_CORS_ORIGIN ?? `http://localhost:${frontendPort}`;
  const jwtSecret = process.env.E2E_JWT_SECRET ?? "e2e-test-secret";
  const backendDir = path.resolve(
    __dirname,
    process.env.BACKEND_DIR ?? "../backend"
  );
  const frontendDir = path.resolve(
    __dirname,
    process.env.FRONTEND_DIR ?? "../frontend"
  );

  const infra = await startInfra();

  await runMigrations(backendDir, infra.dbDsn);

  const backendProc = await spawnBackend(backendDir, backendPort, {
    API_PORT: backendPort,
    DB_DSN: infra.dbDsn,
    PUBSUB_BROKERS: infra.brokers,
    JWT_SECRET: jwtSecret,
    ENV: "test",

    CORS_ORIGINS: corsOrigin,
  });

  const frontendProc = await spawnFrontend(
    frontendDir,
    frontendPort,
    `http://localhost:${backendPort}`
  );

  return async () => {
    await killProcess(frontendProc);
    await killProcess(backendProc);
    await infra.kafka.stop();
    await infra.postgres.stop();
  };
}
