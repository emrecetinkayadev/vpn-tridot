import "server-only";

export type ComponentStatus = {
  name: string;
  status: "operational" | "degraded" | "outage";
  description: string;
  since: string;
  detail?: string;
};

export type IncidentStatus = {
  id: string;
  title: string;
  impact: string;
  timeline: string;
  resolution: string;
  status: "investigating" | "monitoring" | "resolved";
};

export type MaintenanceWindow = {
  id: string;
  title: string;
  window: string;
  scope: string;
  status: "scheduled" | "in-progress" | "complete";
};

export type StatusSnapshot = {
  components: ComponentStatus[];
  incidents: IncidentStatus[];
  maintenance: MaintenanceWindow[];
  lastUpdated: string;
  dataSource: "prometheus" | "static";
  errors?: string[];
};

type PrometheusVectorResult = {
  metric: Record<string, string>;
  value: [string, string];
};

type PrometheusResponse = {
  status: "success" | "error";
  data?: {
    resultType: string;
    result: PrometheusVectorResult[];
  };
  error?: string;
  errorType?: string;
};

const FALLBACK_SNAPSHOT: StatusSnapshot = {
  components: [
    {
      name: "Control plane API",
      status: "operational",
      description: "Auth, billing, ve peer orkestrasyonu uç noktaları",
      since: "2024-10-28T07:45:00Z",
    },
    {
      name: "Node agent fleet",
      status: "degraded",
      description: "EU-FRA node genişlemesi sırasında throttled",
      since: "2024-10-30T12:10:00Z",
    },
    {
      name: "Observability stack",
      status: "operational",
      description: "Prometheus, Grafana, Loki",
      since: "2024-10-31T08:00:00Z",
    },
  ],
  incidents: [
    {
      id: "INC-241030",
      title: "EU-FRA node health flaps",
      impact: "Partial",
      timeline: "08:15Z–08:37Z",
      resolution: "TR-IST'e drain + kernel patch",
      status: "resolved",
    },
  ],
  maintenance: [
    {
      id: "MAIN-241101",
      title: "Database maintenance",
      window: "Nov 1, 01:00–02:00 TRT",
      scope: "Postgres güvenlik yamaları, <2 dk failover",
      status: "scheduled",
    },
  ],
  lastUpdated: "2024-10-31T08:15:00Z",
  dataSource: "static",
};

export async function fetchStatusSnapshot(): Promise<StatusSnapshot> {
  const baseUrl = process.env.STATUS_PROMETHEUS_URL || process.env.NEXT_PUBLIC_STATUS_PROMETHEUS_URL;
  if (!baseUrl) {
    return FALLBACK_SNAPSHOT;
  }

  const errors: string[] = [];
  const now = new Date();

  try {
    const [
      controlPlaneUp,
      nodeAgentUp,
      observabilityUp,
      activePeers,
      handshakeRatio,
    ] = await Promise.all([
      queryVector(baseUrl, 'up{job="control-plane"}'),
      queryVector(baseUrl, 'up{job="node-agent"}'),
      queryVector(baseUrl, 'up{job="observability"}'),
      queryVector(baseUrl, 'sum(node_agent_wireguard_active_peers)'),
      queryVector(baseUrl, 'avg(node_agent_wireguard_handshake_ratio)'),
    ]);

    const components: ComponentStatus[] = [
      buildComponent("Control plane API", controlPlaneUp, now, {
        onAllDown: "Control plane API erişilemez",
        description: "Auth, billing, ve peer orkestrasyonu uç noktaları",
      }),
      buildNodeFleetComponent(nodeAgentUp, handshakeRatio, activePeers, now),
      buildComponent("Observability stack", observabilityUp, now, {
        onAllDown: "Prometheus/Grafana erişilemez",
        description: "Prometheus, Grafana, Loki",
      }),
    ];

    return {
      components,
      incidents: [],
      maintenance: FALLBACK_SNAPSHOT.maintenance,
      lastUpdated: now.toISOString(),
      dataSource: "prometheus",
      errors: errors.length ? errors : undefined,
    };
  } catch (error) {
    errors.push(error instanceof Error ? error.message : "Beklenmeyen Prometheus hatası");
    return {
      ...FALLBACK_SNAPSHOT,
      lastUpdated: now.toISOString(),
      dataSource: "static",
      errors,
    };
  }
}

async function queryVector(baseUrl: string, query: string): Promise<PrometheusVectorResult[]> {
  const url = buildPrometheusURL(baseUrl, "/api/v1/query", { query });
  const response = await fetch(url, {
    cache: "no-store",
    headers: {
      Accept: "application/json",
    },
  });

  if (!response.ok) {
    throw new Error(`Prometheus isteği ${response.status} ile döndü: ${response.statusText}`);
  }

  const payload = (await response.json()) as PrometheusResponse;
  if (payload.status !== "success" || !payload.data) {
    throw new Error(payload.error ?? "Prometheus yanıtı başarısız");
  }

  return payload.data.result ?? [];
}

function buildComponent(
  name: string,
  vector: PrometheusVectorResult[],
  fallbackDate: Date,
  options: { onAllDown: string; description: string },
): ComponentStatus {
  if (vector.length === 0) {
    return {
      name,
      status: "degraded",
      description: options.description,
      since: fallbackDate.toISOString(),
      detail: "Prometheus'tan metrik alınamadı",
    };
  }

  const down = vector.filter((item) => Number(item.value[1]) === 0);
  const allDown = down.length === vector.length;
  const status: ComponentStatus["status"] = allDown ? "outage" : down.length > 0 ? "degraded" : "operational";
  const anySample = vector[0]?.value?.[0];
  const since = anySample ? new Date(Number(anySample) * 1000).toISOString() : fallbackDate.toISOString();

  return {
    name,
    status,
    description: options.description,
    since,
    detail: allDown
      ? options.onAllDown
      : down.length > 0
      ? `${down.length} hedef down: ${down.map((item) => item.metric.instance ?? "unknown").join(", ")}`
      : undefined,
  };
}

function buildNodeFleetComponent(
  vector: PrometheusVectorResult[],
  handshakeRatio: PrometheusVectorResult[],
  activePeers: PrometheusVectorResult[],
  fallbackDate: Date,
): ComponentStatus {
  const base = buildComponent("Node agent fleet", vector, fallbackDate, {
    onAllDown: "Hiçbir node agent ulaşılabilir değil",
    description: "WireGuard node'ları ve agent sağlık metrikleri",
  });

  const peers = parseFloat(activePeers[0]?.value?.[1] ?? "0");
  const ratio = parseFloat(handshakeRatio[0]?.value?.[1] ?? "0");

  let status = base.status;
  if (status === "operational" && Number.isFinite(ratio) && ratio < 0.9) {
    status = "degraded";
  }

  const detailParts: string[] = [];
  if (Number.isFinite(peers)) {
    detailParts.push(`${peers.toFixed(0)} aktif peer`);
  }
  if (Number.isFinite(ratio)) {
    detailParts.push(`handshake oranı ${(ratio * 100).toFixed(1)}%`);
  }
  if (base.detail) {
    detailParts.push(base.detail);
  }

  return {
    ...base,
    status,
    detail: detailParts.length ? detailParts.join(" · ") : base.detail,
  };
}

function buildPrometheusURL(base: string, path: string, params: Record<string, string>): string {
  const url = new URL(path, ensureTrailingSlash(base));
  for (const [key, value] of Object.entries(params)) {
    url.searchParams.set(key, value);
  }
  return url.toString();
}

function ensureTrailingSlash(value: string): string {
  if (/^https?:\/\//i.test(value)) {
    return value;
  }
  return `http://${value.replace(/^\/+/u, "")}`;
}
