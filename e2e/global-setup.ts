import path from "node:path";
import { fileURLToPath } from "node:url";
import { ChildProcess } from "node:child_process";
import dotenv from "dotenv";

import { startInfra } from "./src/docker/start-infra.js";
import { runMigrations } from "./src/docker/run-migrations.js";
import { spawnBackendService } from "./src/docker/spawn-backend-service.js";
import { spawnFrontend } from "./src/docker/spawn-frontend.js";
import { startGateway } from "./src/docker/start-gateway.js";

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
  const requestedGatewayPort = process.env.E2E_GATEWAY_PORT;
  const frontendPort = process.env.E2E_FRONTEND_PORT ?? "4173";
  const servicePorts = {
    api: "4333",
    notification: "3335",
    websocket: "3336",
  } as const;
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

  const sharedServiceEnv = {
    DB_DSN: infra.dbDsn,
    PUBSUB_BROKERS: infra.brokers,
    JWT_SECRET: jwtSecret,
    ENV: "test",
    CORS_ORIGINS: corsOrigin,
  };

  const backendProc = await spawnBackendService(
    backendDir,
    "api",
    "./cmd/api",
    servicePorts.api,
    { ...sharedServiceEnv, API_PORT: servicePorts.api }
  );

  const notificationProc = await spawnBackendService(
    backendDir,
    "notification-service",
    "./cmd/notification-service",
    servicePorts.notification,
    {
      ...sharedServiceEnv,
      NOTIFICATION_SERVICE_PORT: servicePorts.notification,
    }
  );

  const websocketProc = await spawnBackendService(
    backendDir,
    "websocket-service",
    "./cmd/websocket-service",
    servicePorts.websocket,
    {
      ...sharedServiceEnv,
      WEBSOCKET_SERVICE_PORT: servicePorts.websocket,
    }
  );

  const gateway = await startGateway(backendDir, requestedGatewayPort);
  const gatewayPort = String(gateway.getMappedPort(80));

  process.env.E2E_RESOLVED_GATEWAY_PORT = gatewayPort;
  const frontendProc = await spawnFrontend(
    frontendDir,
    frontendPort,
    `http://localhost:${gatewayPort}`
  );

  return async () => {
    await killProcess(frontendProc);
    await gateway.stop();
    await killProcess(websocketProc);
    await killProcess(notificationProc);
    await killProcess(backendProc);
    await infra.kafka.stop();
    await infra.postgres.stop();
  };
}
