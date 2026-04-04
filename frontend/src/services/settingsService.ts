const NTFY_TOPIC_STORAGE_KEY = "ntfyTopic";
const LEGACY_BARK_KEY_STORAGE_KEY = "barkKey";
const CLIENT_ID_STORAGE_KEY = "clientId";

function normalizeNtfyTopic(input: string): string {
  const trimmed = input.trim();
  if (!trimmed) {
    return "";
  }

  if (trimmed.toLowerCase().startsWith("http")) {
    try {
      const parsed = new URL(trimmed);
      const path = parsed.pathname.replace(/^\/+|\/+$/g, "");
      if (!path) return "";

      const parts = path.split("/");
      for (let i = parts.length - 1; i >= 0; i -= 1) {
        const segment = parts[i].trim();
        if (segment) {
          return segment;
        }
      }
      return "";
    } catch {
      const fallback = trimmed.replace(/^\/+|\/+$/g, "");
      const parts = fallback.split("/");
      for (let i = parts.length - 1; i >= 0; i -= 1) {
        const segment = parts[i].trim();
        if (segment) {
          return segment;
        }
      }
      return "";
    }
  }

  return trimmed;
}

function createClientId(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }

  return `client_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`;
}

export const settingsService = {
  // Get ntfy topic from localStorage
  getNtfyTopic: (): string => {
    return (
      localStorage.getItem(NTFY_TOPIC_STORAGE_KEY) ||
      localStorage.getItem(LEGACY_BARK_KEY_STORAGE_KEY) ||
      ""
    );
  },

  // Save ntfy topic to localStorage
  setNtfyTopic: (topic: string): void => {
    const normalizedTopic = normalizeNtfyTopic(topic);

    if (normalizedTopic) {
      localStorage.setItem(NTFY_TOPIC_STORAGE_KEY, normalizedTopic);
      // Keep legacy key for compatibility with existing E2E fixtures.
      localStorage.setItem(LEGACY_BARK_KEY_STORAGE_KEY, normalizedTopic);
    } else {
      localStorage.removeItem(NTFY_TOPIC_STORAGE_KEY);
      localStorage.removeItem(LEGACY_BARK_KEY_STORAGE_KEY);
    }
  },

  // Clear ntfy topic from localStorage
  clearNtfyTopic: (): void => {
    localStorage.removeItem(NTFY_TOPIC_STORAGE_KEY);
    localStorage.removeItem(LEGACY_BARK_KEY_STORAGE_KEY);
  },

  // Get the stable client ID from localStorage
  getClientId: (): string => {
    return localStorage.getItem(CLIENT_ID_STORAGE_KEY) || "";
  },

  // Get or create the stable client ID used as the API user identifier
  getOrCreateClientId: (): string => {
    const existingId = localStorage.getItem(CLIENT_ID_STORAGE_KEY);
    if (existingId) {
      return existingId;
    }

    const clientId = createClientId();
    localStorage.setItem(CLIENT_ID_STORAGE_KEY, clientId);
    return clientId;
  },

  // Clear the stable client ID, primarily for tests
  clearClientId: (): void => {
    localStorage.removeItem(CLIENT_ID_STORAGE_KEY);
  },

  // Get the legacy bark-key-based user ID for one-time backend migration
  getLegacyUserId: (): string => {
    const legacyKey = localStorage.getItem(LEGACY_BARK_KEY_STORAGE_KEY) || "";
    return normalizeNtfyTopic(legacyKey);
  },

  // Normalize ntfy topic for reuse across services
  normalizeNtfyTopic: (input: string): string => {
    return normalizeNtfyTopic(input);
  },

  // Backward-compatible aliases used by existing components/tests.
  getBarkKey: (): string => {
    return settingsService.getNtfyTopic();
  },
  setBarkKey: (key: string): void => {
    settingsService.setNtfyTopic(key);
  },
  clearBarkKey: (): void => {
    settingsService.clearNtfyTopic();
  },
  normalizeBarkKey: (input: string): string => {
    return normalizeNtfyTopic(input);
  },
};
