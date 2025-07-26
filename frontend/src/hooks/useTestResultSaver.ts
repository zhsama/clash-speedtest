import { useCallback } from 'react';
import { toast } from 'sonner';
import type { TestStartData, TestResultData, TestCompleteData } from './useWebSocket';

// 直接复用现有的WebSocket数据接口
interface SavedTestSession {
  // 基本信息
  id: string;                           // UUID
  savedAt: number;                      // 保存时间戳
  
  // 直接保存WebSocket数据，无需转换
  startData: TestStartData;             // 测试开始配置
  results: TestResultData[];            // 所有节点结果
  completeData: TestCompleteData;       // 测试完成统计
  
  // 简单元数据
  meta: {
    duration: string;                   // 从completeData.total_duration
    testType: 'speed' | 'unlock' | 'both'; // 测试类型
    userNotes?: string;                 // 用户备注
    tags?: string[];                    // 用户标签
  };
}

// localStorage存储格式
interface TestSessionStorage {
  sessions: SavedTestSession[];         // 测试会话列表
  lastCleanup: number;                  // 上次清理时间
  version: string;                      // 数据版本
}

interface UseTestResultSaverOptions {
  maxSessions?: number;              // 最大保存会话数，默认50
  autoSave?: boolean;                // 是否自动保存，默认true
  onSaved?: (sessionId: string) => void;  // 保存成功回调
}

// 存储配置
const STORAGE_CONFIG = {
  KEY: 'clash-speedtest-sessions',
  MAX_SESSIONS: 50,
  MAX_SIZE_MB: 5,
  AUTO_CLEANUP_DAYS: 30
} as const;

export function useTestResultSaver(options: UseTestResultSaverOptions = {}) {
  const { maxSessions = STORAGE_CONFIG.MAX_SESSIONS, autoSave = true, onSaved } = options;
  
  // 获取存储的会话
  const getStoredSessions = useCallback((): TestSessionStorage => {
    try {
      const stored = localStorage.getItem(STORAGE_CONFIG.KEY);
      if (!stored) {
        return { sessions: [], lastCleanup: Date.now(), version: '1.0.0' };
      }
      return JSON.parse(stored);
    } catch {
      return { sessions: [], lastCleanup: Date.now(), version: '1.0.0' };
    }
  }, []);
  
  // 保存数据到localStorage
  const saveToStorage = useCallback((data: TestSessionStorage) => {
    try {
      localStorage.setItem(STORAGE_CONFIG.KEY, JSON.stringify(data));
    } catch (error) {
      console.error('Failed to save to localStorage:', error);
      throw new Error('存储空间不足，请清理部分历史记录');
    }
  }, []);
  
  // 智能清理旧数据
  const performCleanup = useCallback((sessions: SavedTestSession[]): SavedTestSession[] => {
    const now = Date.now();
    const cleanupThreshold = now - (STORAGE_CONFIG.AUTO_CLEANUP_DAYS * 24 * 60 * 60 * 1000);
    
    // 过滤并排序
    return sessions
      .filter(s => s.savedAt > cleanupThreshold) // 移除超过30天的
      .sort((a, b) => {
        // 成功率高的优先
        const successRateA = a.completeData.successful_tests / a.startData.total_proxies;
        const successRateB = b.completeData.successful_tests / b.startData.total_proxies;
        if (successRateA !== successRateB) return successRateB - successRateA;
        
        // 时间新的优先
        return b.savedAt - a.savedAt;
      })
      .slice(0, maxSessions);
  }, [maxSessions]);
  
  // 保存测试会话
  const saveTestSession = useCallback(async (
    startData: TestStartData,
    results: TestResultData[],
    completeData: TestCompleteData,
    testType: 'speed' | 'unlock' | 'both',
    userNotes?: string,
    tags?: string[]
  ) => {
    if (!autoSave) return null;
    
    try {
      const sessionId = crypto.randomUUID();
      const now = Date.now();
      
      const session: SavedTestSession = {
        id: sessionId,
        savedAt: now,
        startData,
        results,
        completeData,
        meta: {
          duration: completeData.total_duration,
          testType,
          userNotes,
          tags
        }
      };
      
      // 获取现有数据并清理
      const existingData = getStoredSessions();
      const cleanedSessions = performCleanup([session, ...existingData.sessions]);
      
      // 保存到localStorage
      const storageData: TestSessionStorage = {
        sessions: cleanedSessions,
        lastCleanup: now,
        version: '1.0.0'
      };
      
      saveToStorage(storageData);
      
      // 成功通知
      toast.success('测试结果已保存', {
        description: `保存了${results.length}个节点的测试数据`
      });
      
      onSaved?.(sessionId);
      
      return sessionId;
      
    } catch (error) {
      console.error('保存测试结果失败:', error);
      toast.error('保存失败', {
        description: error instanceof Error ? error.message : '无法保存测试结果到本地存储'
      });
      throw error;
    }
  }, [autoSave, onSaved, getStoredSessions, performCleanup, saveToStorage]);
  
  // 获取会话摘要列表
  const getSessionSummaries = useCallback(() => {
    const data = getStoredSessions();
    return data.sessions
      .map(session => ({
        id: session.id,
        savedAt: session.savedAt,
        configPaths: session.startData.config.config_paths,
        totalProxies: session.startData.total_proxies,
        successfulTests: session.completeData.successful_tests,
        failedTests: session.completeData.failed_tests,
        averageDownloadMbps: session.completeData.average_download_mbps,
        averageLatency: session.completeData.average_latency,
        bestProxy: session.completeData.best_proxy,
        bestDownloadSpeedMbps: session.completeData.best_download_speed_mbps,
        duration: session.meta.duration,
        testType: session.meta.testType,
        notes: session.meta.userNotes,
        tags: session.meta.tags,
        // 额外的统计信息
        unlockStats: session.completeData.unlock_stats
      }))
      .sort((a, b) => b.savedAt - a.savedAt); // 按时间倒序排序，最新的在前
  }, [getStoredSessions]);
  
  // 获取完整会话数据
  const getSessionById = useCallback((id: string): SavedTestSession | null => {
    const data = getStoredSessions();
    return data.sessions.find(s => s.id === id) || null;
  }, [getStoredSessions]);
  
  // 删除会话
  const deleteSession = useCallback((id: string) => {
    const data = getStoredSessions();
    const sessions = data.sessions.filter(s => s.id !== id);
    
    const storageData: TestSessionStorage = {
      ...data,
      sessions
    };
    
    saveToStorage(storageData);
    toast.success('已删除测试记录');
  }, [getStoredSessions, saveToStorage]);
  
  // 批量删除会话
  const deleteSessions = useCallback((ids: string[]) => {
    const data = getStoredSessions();
    const sessions = data.sessions.filter(s => !ids.includes(s.id));
    
    const storageData: TestSessionStorage = {
      ...data,
      sessions
    };
    
    saveToStorage(storageData);
    toast.success(`已删除${ids.length}条测试记录`);
  }, [getStoredSessions, saveToStorage]);
  
  // 清空所有会话
  const clearAllSessions = useCallback(() => {
    localStorage.removeItem(STORAGE_CONFIG.KEY);
    toast.success('已清空所有测试记录');
  }, []);
  
  // 获取存储统计信息
  const getStorageStats = useCallback(() => {
    const data = getStoredSessions();
    const dataStr = JSON.stringify(data);
    const sizeBytes = new Blob([dataStr]).size;
    const sizeMB = sizeBytes / (1024 * 1024);
    
    return {
      totalSessions: data.sessions.length,
      sizeBytes,
      sizeMB: Number(sizeMB.toFixed(2)),
      lastCleanup: data.lastCleanup,
      oldestSession: data.sessions.length > 0 
        ? Math.min(...data.sessions.map(s => s.savedAt))
        : null,
      newestSession: data.sessions.length > 0
        ? Math.max(...data.sessions.map(s => s.savedAt))
        : null
    };
  }, [getStoredSessions]);
  
  // 导出数据
  const exportSessions = useCallback((ids?: string[], format: 'json' | 'csv' = 'json') => {
    const data = getStoredSessions();
    const sessionsToExport = ids 
      ? data.sessions.filter(s => ids.includes(s.id))
      : data.sessions;
    
    if (format === 'json') {
      const exportData = {
        metadata: {
          exportTime: new Date().toISOString(),
          version: '1.0.0',
          totalSessions: sessionsToExport.length,
          source: 'Clash SpeedTest'
        },
        sessions: sessionsToExport
      };
      
      return JSON.stringify(exportData, null, 2);
    } else {
      // CSV格式导出摘要
      const headers = [
        'ID', '保存时间', '配置路径', '总节点数', '成功节点数', '失败节点数',
        '平均下载速度(Mbps)', '平均延迟(ms)', '最佳节点', '测试时长', '备注'
      ];
      
      const rows = sessionsToExport.map(session => [
        session.id,
        new Date(session.savedAt).toLocaleString('zh-CN'),
        session.startData.config.config_paths,
        session.startData.total_proxies,
        session.completeData.successful_tests,
        session.completeData.failed_tests,
        session.completeData.average_download_mbps.toFixed(2),
        session.completeData.average_latency.toFixed(0),
        session.completeData.best_proxy || 'N/A',
        session.meta.duration,
        session.meta.userNotes || ''
      ]);
      
      return [headers, ...rows]
        .map(row => row.map(cell => `"${cell}"`).join(','))
        .join('\n');
    }
  }, [getStoredSessions]);
  
  return {
    // 核心功能
    saveTestSession,
    getSessionSummaries,
    getSessionById,
    
    // 管理功能
    deleteSession,
    deleteSessions,
    clearAllSessions,
    
    // 工具功能
    getStorageStats,
    exportSessions
  };
}