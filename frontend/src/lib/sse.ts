export interface SSEMessage {
  data: unknown;
  event: string;
}

export function connectSSE(
  url: string,
  onMessage: (message: SSEMessage) => void,
  onError?: (error: Error) => void,
  token?: string | null
): () => void {
  // EventSource cannot send custom headers, so pass token as query param
  const sep = url.includes("?") ? "&" : "?";
  const sseUrl = token ? `${url}${sep}token=${encodeURIComponent(token)}` : url;

  const eventSource = new EventSource(sseUrl);

  eventSource.onopen = () => {
    console.log("SSE connection opened");
  };

  eventSource.addEventListener("run_state", (e) => {
    try {
      const data = JSON.parse(e.data);
      onMessage({ event: "run_state", data });
    } catch (err) {
      console.error("Failed to parse SSE message:", err);
    }
  });

  eventSource.addEventListener("steps_state", (e) => {
    try {
      const data = JSON.parse(e.data);
      onMessage({ event: "steps_state", data });
    } catch (err) {
      console.error("Failed to parse SSE message:", err);
    }
  });

  eventSource.addEventListener("complete", (e) => {
    try {
      const data = JSON.parse(e.data);
      onMessage({ event: "complete", data });
      eventSource.close();
    } catch (err) {
      console.error("Failed to parse SSE message:", err);
    }
  });

  eventSource.addEventListener("error", (_e) => {
    const error = new Error("SSE connection error");
    if (onError) {
      onError(error);
    }
    eventSource.close();
  });

  // Return cleanup function
  return () => {
    eventSource.close();
  };
}
