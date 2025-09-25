export interface ApiClientOptions {
  baseUrl: string;
  token?: string;
}

export class ApiClient {
  private readonly baseUrl: string;
  private readonly token?: string;

  constructor({ baseUrl, token }: ApiClientOptions) {
    this.baseUrl = baseUrl.replace(/\/$/, "");
    this.token = token;
  }

  async get<T>(path: string, init?: RequestInit): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      ...init,
      headers: this.createHeaders(init?.headers),
    });

    if (!response.ok) {
      throw new Error(`Request failed with status ${response.status}`);
    }

    return (await response.json()) as T;
  }

  private createHeaders(headers?: HeadersInit): HeadersInit {
    const merged: Record<string, string> = {
      Accept: "application/json",
      "Content-Type": "application/json",
      ...(headers ? Object.fromEntries(new Headers(headers)) : {}),
    };

    if (this.token) {
      merged.Authorization = `Bearer ${this.token}`;
    }

    return merged;
  }
}

const getBaseUrl = () => {
  if (typeof window !== "undefined") {
    return "";
  }
  return process.env.VERCEL_URL
    ? `https://${process.env.VERCEL_URL}`
    : "http://localhost:3000";
};

export const defaultApiClient = new ApiClient({
  baseUrl: getBaseUrl(),
});
