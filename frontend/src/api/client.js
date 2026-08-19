async function parseResponse(response) {
  const contentType = response.headers.get("content-type") || "";
  const isJSON = contentType.includes("application/json");
  const payload = isJSON ? await response.json() : await response.text();

  if (!response.ok) {
    const message =
      typeof payload === "object" && payload !== null && "error" in payload
        ? payload.error
        : "Ошибка запроса к серверу.";

    throw new Error(message);
  }

  return payload;
}

const inFlightGetRequests = new Map();

export const apiClient = {
  async get(path) {
    const existingRequest = inFlightGetRequests.get(path);
    if (existingRequest) {
      return existingRequest;
    }

    const request = fetch(path).then(parseResponse);
    inFlightGetRequests.set(path, request);

    try {
      return await request;
    } finally {
      if (inFlightGetRequests.get(path) === request) {
        inFlightGetRequests.delete(path);
      }
    }
  },

  async post(path, options = {}) {
    const response = await fetch(path, {
      method: "POST",
      ...options
    });

    return parseResponse(response);
  }
};
