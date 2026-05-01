export interface SSEMessage {
  event: string;
  data: unknown;
}

export function connectSSE(url: string, onMessage: (message: SSEMessage) => void, onError?: (error: Error) => void): () => void {
  const eventSource = new EventSource(url, {
    withCredentials: true,
  });

  eventSource.onopen = () => {
    console.log('SSE connection opened');
  };

  eventSource.addEventListener('run_state', (e) => {
    try {
      const data = JSON.parse(e.data);
      onMessage({ event: 'run_state', data });
    } catch (err) {
      console.error('Failed to parse SSE message:', err);
    }
  });

  eventSource.addEventListener('complete', (e) => {
    try {
      const data = JSON.parse(e.data);
      onMessage({ event: 'complete', data });
      eventSource.close();
    } catch (err) {
      console.error('Failed to parse SSE message:', err);
    }
  });

  eventSource.addEventListener('error', (e) => {
    const error = new Error('SSE connection error');
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
