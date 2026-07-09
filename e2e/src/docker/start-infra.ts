import {
  PostgreSqlContainer,
  StartedPostgreSqlContainer,
} from "@testcontainers/postgresql";
import { KafkaContainer, StartedKafkaContainer } from "@testcontainers/kafka";

export interface Infra {
  postgres: StartedPostgreSqlContainer;
  kafka: StartedKafkaContainer;
  dbDsn: string;
  brokers: string;
}

export async function startInfra(): Promise<Infra> {
  const [postgres, kafka] = await Promise.all([
    new PostgreSqlContainer("postgres:17-alpine")
      .withDatabase("e2e")
      .withUsername("e2e")
      .withPassword("e2e")
      .start(),
    new KafkaContainer("confluentinc/cp-kafka:7.6.0").withKraft().start(),
  ]);

  const dbDsn = `${postgres.getConnectionUri()}?sslmode=disable`;
  const brokers = `${kafka.getHost()}:${kafka.getMappedPort(9093)}`;

  return { postgres, kafka, dbDsn, brokers };
}
