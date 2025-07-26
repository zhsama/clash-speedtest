# 基于现有架构的测试结果保存方案

## 🎯 设计原则

1. **复用现有数据结构** - 直接使用WebSocket推送的数据格式
2. **最小代码侵入** - 仅在测试完成时进行保存
3. **简单直接** - localStorage直接存储，无复杂转换
4. **渐进增强** - 不影响现有功能，纯增强型功能

## 📊 保存数据结构

```typescript
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
```

## 🔧 核心实现

### 1. 自动保存Hook

```typescript
// hooks/useTestResultSaver.ts
import { useCallback, useEffect } from 'react';
import { toast } from 'sonner';
import type { TestStartData, TestResultData, TestCompleteData } from './useWebSocket';

interface UseTestResultSaverOptions {
  maxSessions?: number;              // 最大保存会话数，默认50
  autoSave?: boolean;                // 是否自动保存，默认true
  onSaved?: (sessionId: string) => void;  // 保存成功回调
}

export function useTestResultSaver(options: UseTestResultSaverOptions = {}) {
  const { maxSessions = 50, autoSave = true, onSaved } = options;
  
  // 保存测试会话
  const saveTestSession = useCallback(async (
    startData: TestStartData,
    results: TestResultData[],
    completeData: TestCompleteData,
    userNotes?: string,
    tags?: string[]
  ) => {
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
          userNotes,
          tags
        }
      };
      
      // 获取现有数据
      const existingData = getStoredSessions();
      
      // 添加新会话，保持数量限制
      const sessions = [session, ...existingData.sessions].slice(0, maxSessions);
      
      // 保存到localStorage
      const storageData: TestSessionStorage = {
        sessions,
        lastCleanup: now,
        version: '1.0.0'
      };
      
      localStorage.setItem('clash-speedtest-sessions', JSON.stringify(storageData));
      
      // 成功通知
      toast.success('测试结果已保存', {
        description: `保存了${results.length}个节点的测试数据`
      });
      
      onSaved?.(sessionId);
      
      return sessionId;
      
    } catch (error) {
      console.error('保存测试结果失败:', error);
      toast.error('保存失败', {
        description: '无法保存测试结果到本地存储'
      });
      throw error;
    }
  }, [maxSessions, onSaved]);
  
  // 获取存储的会话
  const getStoredSessions = useCallback((): TestSessionStorage => {
    try {
      const stored = localStorage.getItem('clash-speedtest-sessions');
      if (!stored) {
        return { sessions: [], lastCleanup: Date.now(), version: '1.0.0' };
      }
      return JSON.parse(stored);
    } catch {
      return { sessions: [], lastCleanup: Date.now(), version: '1.0.0' };
    }
  }, []);
  
  // 获取会话列表（仅摘要信息）
  const getSessionSummaries = useCallback(() => {
    const data = getStoredSessions();
    return data.sessions.map(session => ({
      id: session.id,
      savedAt: session.savedAt,
      configPaths: session.startData.config.config_paths,
      totalProxies: session.startData.total_proxies,
      successfulTests: session.completeData.successful_tests,
      failedTests: session.completeData.failed_tests,
      averageDownloadMbps: session.completeData.average_download_mbps,
      averageLatency: session.completeData.average_latency,
      bestProxy: session.completeData.best_proxy,
      duration: session.meta.duration,
      notes: session.meta.userNotes,
      tags: session.meta.tags
    }));
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
    
    localStorage.setItem('clash-speedtest-sessions', JSON.stringify(storageData));
    
    toast.success('已删除测试记录');
  }, [getStoredSessions]);
  
  // 清空所有会话
  const clearAllSessions = useCallback(() => {
    localStorage.removeItem('clash-speedtest-sessions');
    toast.success('已清空所有测试记录');
  }, []);
  
  return {
    saveTestSession,
    getSessionSummaries,
    getSessionById,
    deleteSession,
    clearAllSessions
  };
}
```

### 2. 在SpeedTest组件中集成

```typescript
// components/SpeedTest.tsx 修改部分
import { useTestResultSaver } from '../hooks/useTestResultSaver';

export default function SpeedTest() {
  // 现有的WebSocket和状态...
  const { saveTestSession } = useTestResultSaver();
  const [showHistory, setShowHistory] = useState(false);
  
  // 监听测试完成事件，自动保存
  useEffect(() => {
    if (testCompleteData && testStartData && testResults.length > 0) {
      // 自动保存测试结果
      saveTestSession(
        testStartData,
        testResults,
        testCompleteData
      ).catch(console.error);
    }
  }, [testCompleteData, testStartData, testResults, saveTestSession]);
  
  return (
    <div className="space-y-md-6">
      {/* 现有UI... */}
      
      {/* 新增：历史记录按钮 */}
      <div className="flex justify-between items-center">
        <h1>Clash SpeedTest</h1>
        <Button 
          onClick={() => setShowHistory(true)}
          className="btn-outlined"
        >
          <History className="h-4 w-4 mr-2" />
          历史记录
        </Button>
      </div>
      
      {/* 现有测试界面... */}
      
      {/* 历史记录组件 */}
      {showHistory && (
        <TestHistoryModal onClose={() => setShowHistory(false)} />
      )}
    </div>
  );
}
```

### 3. 简单的历史记录组件

```typescript
// components/TestHistoryModal.tsx
import { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useTestResultSaver } from '../hooks/useTestResultSaver';
import { Download, Trash2, Eye, Calendar, Globe } from 'lucide-react';
import { toast } from 'sonner';

interface TestHistoryModalProps {
  onClose: () => void;
}

export default function TestHistoryModal({ onClose }: TestHistoryModalProps) {
  const { getSessionSummaries, deleteSession, clearAllSessions } = useTestResultSaver();
  const [sessions, setSessions] = useState<any[]>([]);
  const [selectedSession, setSelectedSession] = useState<string | null>(null);
  
  useEffect(() => {
    setSessions(getSessionSummaries());
  }, [getSessionSummaries]);
  
  const handleDelete = (id: string) => {
    deleteSession(id);
    setSessions(getSessionSummaries());
  };
  
  const handleExport = (session: any) => {
    const exportData = {
      summary: session,
      exportedAt: new Date().toISOString()
    };
    
    const blob = new Blob([JSON.stringify(exportData, null, 2)], {
      type: 'application/json'
    });
    
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `speedtest-${session.id.slice(0, 8)}-${new Date(session.savedAt).toISOString().slice(0, 10)}.json`;
    link.click();
    URL.revokeObjectURL(url);
    
    toast.success('测试结果已导出');
  };
  
  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[80vh] overflow-hidden">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Calendar className="h-5 w-5" />
            测试历史记录
          </DialogTitle>
        </DialogHeader>
        
        <div className="space-y-4 overflow-y-auto">
          {/* 操作栏 */}
          <div className="flex justify-between items-center">
            <div className="text-sm text-muted-foreground">
              共 {sessions.length} 条记录
            </div>
            <Button 
              onClick={() => {
                clearAllSessions();
                setSessions([]);
              }}
              variant="outline"
              size="sm"
              className="btn-outlined text-destructive"
            >
              <Trash2 className="h-4 w-4 mr-1" />
              清空全部
            </Button>
          </div>
          
          {/* 会话列表 */}
          {sessions.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              暂无测试记录
            </div>
          ) : (
            sessions.map((session) => (
              <Card key={session.id} className="card-elevated p-4">
                <div className="flex items-start justify-between">
                  <div className="flex-1 space-y-2">
                    {/* 基本信息 */}
                    <div className="flex items-center gap-2">
                      <Globe className="h-4 w-4 text-muted-foreground" />
                      <span className="font-medium truncate max-w-xs">
                        {session.configPaths || '配置文件'}
                      </span>
                      <Badge className="badge-filled">
                        {session.totalProxies} 节点
                      </Badge>
                    </div>
                    
                    {/* 统计信息 */}
                    <div className="flex items-center gap-4 text-sm text-muted-foreground">
                      <span>✅ {session.successfulTests} 成功</span>
                      <span>❌ {session.failedTests} 失败</span>
                      <span>⚡ {session.averageDownloadMbps.toFixed(1)} Mbps</span>
                      <span>📡 {session.averageLatency.toFixed(0)} ms</span>
                    </div>
                    
                    {/* 时间和最佳节点 */}
                    <div className="flex items-center gap-4 text-xs text-muted-foreground">
                      <span>{new Date(session.savedAt).toLocaleString('zh-CN')}</span>
                      {session.bestProxy && (
                        <span>🏆 {session.bestProxy}</span>
                      )}
                    </div>
                  </div>
                  
                  {/* 操作按钮 */}
                  <div className="flex gap-2">
                    <Button
                      onClick={() => handleExport(session)}
                      size="sm"
                      className="btn-outlined"
                    >
                      <Download className="h-3 w-3" />
                    </Button>
                    <Button
                      onClick={() => handleDelete(session.id)}
                      size="sm"
                      variant="outline"
                      className="btn-outlined text-destructive"
                    >
                      <Trash2 className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
              </Card>
            ))
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
```

## 🚀 集成步骤

### 第一步：创建保存Hook
```bash
# 创建hooks文件
touch frontend/src/hooks/useTestResultSaver.ts
```

### 第二步：修改SpeedTest组件
在现有的`useEffect`中添加测试完成监听：

```typescript
// 在SpeedTest.tsx中添加
useEffect(() => {
  if (testCompleteData && testStartData && testResults.length > 0) {
    saveTestSession(testStartData, testResults, testCompleteData);
  }
}, [testCompleteData, testStartData, testResults, saveTestSession]);
```

### 第三步：添加历史记录按钮
在现有UI中添加历史记录按钮，通过Material 3样式集成。

### 第四步：创建历史记录组件
简单的模态框组件，展示保存的测试会话。

## 📊 存储优化

```typescript
// 存储大小限制和清理策略
const STORAGE_CONFIG = {
  MAX_SESSIONS: 50,           // 最大保存50个测试会话
  MAX_SIZE_MB: 5,            // 最大占用5MB存储空间
  AUTO_CLEANUP_DAYS: 30,     // 30天后自动清理
  CLEANUP_CHECK_HOURS: 24    // 每24小时检查一次清理
};

// 智能清理：优先删除失败的测试、旧的测试
function smartCleanup(sessions: SavedTestSession[]): SavedTestSession[] {
  const now = Date.now();
  const thirtyDaysAgo = now - (30 * 24 * 60 * 60 * 1000);
  
  // 按优先级排序：成功的测试 > 新的测试 > 节点数多的测试
  return sessions
    .filter(s => s.savedAt > thirtyDaysAgo) // 移除超过30天的
    .sort((a, b) => {
      // 成功率高的优先
      const successRateA = a.completeData.successful_tests / a.startData.total_proxies;
      const successRateB = b.completeData.successful_tests / b.startData.total_proxies;
      if (successRateA !== successRateB) return successRateB - successRateA;
      
      // 时间新的优先
      return b.savedAt - a.savedAt;
    })
    .slice(0, STORAGE_CONFIG.MAX_SESSIONS);
}
```

## 🎯 核心优势

1. **零侵入性** - 不修改现有WebSocket逻辑和数据结构
2. **自动保存** - 测试完成后自动触发保存
3. **轻量级** - 直接存储原始数据，无复杂转换
4. **即插即用** - 可以随时启用/禁用功能
5. **渐进增强** - 不影响现有功能，纯增强型

## 📱 用户体验

- **无感保存** - 测试完成自动保存，用户无需操作
- **快速查看** - 历史记录一键查看
- **简单导出** - JSON格式导出，方便分享和分析
- **智能管理** - 自动清理旧数据，不占用过多存储

这个方案完全基于您现有的WebSocket架构，**无需修改后端接口**，只需要在前端监听`test_complete`事件进行保存即可！