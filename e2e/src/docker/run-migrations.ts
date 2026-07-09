import { execa } from "execa";

export async function runMigrations(
  backendDir: string,
  dbDsn: string
): Promise<void> {
  await execa("go", ["run", "./cmd/migrate", "up"], {
    cwd: backendDir,
    env: { ...process.env, DB_DSN: dbDsn },
    stdio: "inherit",
  });
}
