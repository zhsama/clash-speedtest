import { useEffect, useRef, useState, useCallback } from "react";

export interface WebSocketMessage {
  type: string;
  timestamp: string;
  data: any;
}

export interface TestStartData {
  total_proxies: number;
  config: {
    config_paths: string;
    filter_regex: string;
    server_url: string;
    download_size: number;
    upload_size: number;
    timeout: number;
    concurrent: number;
    max_latency: number;
    min_download_speed: number;
    min_upload_speed: number;
    stash_compatible: boolean;
  };
}

export interface TestProgressData {
  current_proxy: string;
  completed_count: number;
  total_count: number;
  progress_percent: number;
  status: string;
}

export interface TestResultData {
  proxy_name: string;
  proxy_type: string;
  proxy_ip?: string; // 新增代理IP地址
  latency_ms: number;
  jitter_ms: number;
  packet_loss: number;
  download_speed: number;
  upload_speed: number;
  download_speed_mbps: number;
  upload_speed_mbps: number;
  status: string;
  // 新增错误诊断字段
  error_stage?: string;
  error_code?: string;
  error_message?: string;
}

export interface TestCompleteData {
  total_tested: number;
  successful_tests: number;
  failed_tests: number;
  total_duration: string;
  average_latency: number;
  average_download_mbps: number;
  average_upload_mbps: number;
  best_proxy: string;
  best_download_speed_mbps: number;
}

export interface ErrorData {
  message: string;
  code?: string;
}

export interface TestCancelledData {
  message: string;
  completed_tests: number;
  total_tests: number;
  partial_duration: string;
}

export interface UseWebSocketReturn {
  isConnected: boolean;
  sendMessage: (message: any) => boolean;
  connect: () => void;
  disconnect: () => void;
  testStartData: TestStartData | null;
  testProgress: TestProgressData | null;
  testResults: TestResultData[];
  testCompleteData: TestCompleteData | null;
  testCancelledData: TestCancelledData | null;
  error: ErrorData | null;
  clearData: () => void;
}

export const useWebSocket = (url: string): UseWebSocketReturn => {
  const [isConnected, setIsConnected] = useState(false);
  const [testStartData, setTestStartData] = useState<TestStartData | null>(
    null
  );
  const [testProgress, setTestProgress] = useState<TestProgressData | null>(
    null
  );
  const [testResults, setTestResults] = useState<TestResultData[]>([]);
  const [testCompleteData, setTestCompleteData] =
    useState<TestCompleteData | null>(null);
  const [testCancelledData, setTestCancelledData] =
    useState<TestCancelledData | null>(null);
  const [error, setError] = useState<ErrorData | null>(null);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 5;

  const clearData = useCallback(() => {
    setTestStartData(null);
    setTestProgress(null);
    setTestResults([]);
    setTestCompleteData(null);
    setTestCancelledData(null);
    setError(null);
  }, []);

  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const message: WebSocketMessage = JSON.parse(event.data);

      switch (message.type) {
        case "test_start":
          setTestStartData(message.data as TestStartData);
          setTestProgress(null);
          setTestResults([]);
          setTestCompleteData(null);
          setError(null);
          break;

        case "test_progress":
          setTestProgress(message.data as TestProgressData);
          break;

        case "test_result":
          const resultData = message.data as TestResultData;
          setTestResults((prev) => [...prev, resultData]);
          break;

        case "test_complete":
          setTestCompleteData(message.data as TestCompleteData);
          break;

        case "test_cancelled":
          setTestCancelledData(message.data as TestCancelledData);
          break;

        case "error":
          setError(message.data as ErrorData);
          break;

        default:
          console.warn("Unknown WebSocket message type:", message.type);
      }
    } catch (err) {
      console.error("Failed to parse WebSocket message:", err);
    }
  }, []);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    try {
      wsRef.current = new WebSocket(url);

      wsRef.current.onopen = () => {
        console.log("WebSocket connected");
        setIsConnected(true);
        reconnectAttemptsRef.current = 0;

        if (reconnectTimeoutRef.current) {
          clearTimeout(reconnectTimeoutRef.current);
          reconnectTimeoutRef.current = null;
        }
      };

      wsRef.current.onmessage = handleMessage;

      wsRef.current.onclose = (event) => {
        console.log("WebSocket disconnected:", event.code, event.reason);
        setIsConnected(false);

        // Attempt to reconnect if not a normal closure
        if (
          event.code !== 1000 &&
          reconnectAttemptsRef.current < maxReconnectAttempts
        ) {
          const delay = Math.min(
            1000 * Math.pow(2, reconnectAttemptsRef.current),
            30000
          );
          console.log(`Attempting to reconnect in ${delay}ms...`);

          reconnectTimeoutRef.current = setTimeout(() => {
            reconnectAttemptsRef.current++;
            connect();
          }, delay);
        }
      };

      wsRef.current.onerror = (error) => {
        console.error("WebSocket error:", error);
        setError({
          message: "WebSocket connection error",
          code: "WEBSOCKET_ERROR",
        });
      };
    } catch (err) {
      console.error("Failed to create WebSocket connection:", err);
      setError({
        message: "Failed to create WebSocket connection",
        code: "CONNECTION_FAILED",
      });
    }
  }, [url, handleMessage]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    reconnectAttemptsRef.current = maxReconnectAttempts; // Prevent reconnection

    if (wsRef.current) {
      wsRef.current.close(1000, "User disconnected");
      wsRef.current = null;
    }

    setIsConnected(false);
  }, []);

  const sendMessage = useCallback((message: any) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
      return true;
    }
    return false;
  }, []);

  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return {
    isConnected,
    connect,
    disconnect,
    sendMessage,
    testStartData,
    testProgress,
    testResults,
    testCompleteData,
    testCancelledData,
    error,
    clearData,
  };
};
