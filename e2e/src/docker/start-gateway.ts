import path from "node:path";
import { GenericContainer, type StartedTestContainer } from "testcontainers";
import { waitOnHttp } from "./wait-on-http.js";

export async function startGateway(
  backendDir: string,
  requestedPort?: string
): Promise<StartedTestContainer> {
  const container = new GenericContainer("traefik:v3.1")
    .withCommand([
      "--entrypoints.web.address=:80",
      "--providers.file.filename=/etc/traefik/dynamic.yml",
    ])
    .withBindMounts([
      {
        source: path.join(backendDir, "traefik", "dynamic-dev.yml"),
        target: "/etc/traefik/dynamic.yml",
        mode: "ro",
      },
    ])
    .withExtraHosts([
      { host: "host.docker.internal", ipAddress: "host-gateway" },
    ]);

  if (requestedPort) {
    const port = Number(requestedPort);
    if (!Number.isInteger(port) || port < 1 || port > 65_535) {
      throw new Error(`Invalid E2E_GATEWAY_PORT: ${requestedPort}`);
    }
    container.withExposedPorts({ container: 80, host: port });
  } else {
    container.withExposedPorts(80);
  }

  const gateway = await container.start();
  const port = gateway.getMappedPort(80);

  try {
    await waitOnHttp(`http://localhost:${port}/health`, 30_000);
    return gateway;
  } catch (error) {
    await gateway.stop();
    throw error;
  }
}
