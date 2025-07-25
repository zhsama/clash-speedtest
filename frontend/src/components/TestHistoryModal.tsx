import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { useTestResultSaver } from '../hooks/useTestResultSaver';
import { 
  Download, 
  Trash2, 
  Calendar, 
  Globe, 
  Zap, 
  Wifi, 
  Trophy,
  FileText,
  BarChart3
} from 'lucide-react';
import { toast } from 'sonner';
import ClientIcon from './ClientIcon';

interface TestHistoryModalProps {
  onClose: () => void;
}

interface SessionSummary {
  id: string;
  savedAt: number;
  configPaths: string;
  totalProxies: number;
  successfulTests: number;
  failedTests: number;
  averageDownloadMbps: number;
  averageLatency: number;
  bestProxy: string;
  bestDownloadSpeedMbps: number;
  duration: string;
  testType: 'speed' | 'unlock' | 'both';
  notes?: string;
  tags?: string[];
  unlockStats?: any;
}

export default function TestHistoryModal({ onClose }: TestHistoryModalProps) {
  const { 
    getSessionSummaries, 
    deleteSession, 
    deleteSessions,
    clearAllSessions,
    exportSessions,
    getStorageStats 
  } = useTestResultSaver();
  
  const [sessions, setSessions] = useState<SessionSummary[]>([]);
  const [selectedSessions, setSelectedSessions] = useState<string[]>([]);
  const [storageStats, setStorageStats] = useState<any>(null);
  
  // 加载数据
  useEffect(() => {
    refreshData();
  }, []);
  
  const refreshData = () => {
    setSessions(getSessionSummaries());
    setStorageStats(getStorageStats());
  };
  
  // 处理单个删除
  const handleDelete = (id: string) => {
    deleteSession(id);
    refreshData();
    setSelectedSessions(prev => prev.filter(s => s !== id));
  };
  
  // 处理批量删除
  const handleBatchDelete = () => {
    if (selectedSessions.length === 0) return;
    
    deleteSessions(selectedSessions);
    refreshData();
    setSelectedSessions([]);
  };
  
  // 处理全部清空
  const handleClearAll = () => {
    if (window.confirm('确定要清空所有测试记录吗？此操作不可恢复。')) {
      clearAllSessions();
      refreshData();
      setSelectedSessions([]);
    }
  };
  
  // 处理导出
  const handleExport = (sessionIds?: string[], format: 'json' | 'csv' = 'json') => {
    try {
      const content = exportSessions(sessionIds, format);
      const timestamp = new Date().toISOString().slice(0, 10);
      const sessionCount = sessionIds?.length || sessions.length;
      
      const blob = new Blob([content], {
        type: format === 'json' ? 'application/json' : 'text/csv'
      });
      
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `clash-speedtest-${timestamp}-${sessionCount}sessions.${format}`;
      link.click();
      URL.revokeObjectURL(url);
      
      toast.success(`已导出${sessionCount}条测试记录`);
    } catch (error) {
      toast.error('导出失败', {
        description: '无法导出测试记录'
      });
    }
  };
  
  // 处理单选/全选
  const handleSelectSession = (id: string) => {
    setSelectedSessions(prev => 
      prev.includes(id) 
        ? prev.filter(s => s !== id)
        : [...prev, id]
    );
  };
  
  const handleSelectAll = () => {
    setSelectedSessions(
      selectedSessions.length === sessions.length 
        ? [] 
        : sessions.map(s => s.id)
    );
  };
  
  // 格式化速度显示
  const formatSpeed = (mbps: number) => {
    if (mbps >= 1000) {
      return `${(mbps / 1000).toFixed(1)}Gbps`;
    }
    return `${mbps.toFixed(1)}Mbps`;
  };
  
  // 获取成功率颜色
  const getSuccessRateColor = (rate: number) => {
    if (rate >= 0.9) return 'text-success';
    if (rate >= 0.7) return 'text-warning';
    return 'text-destructive';
  };
  
  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="max-w-6xl max-h-[90vh] overflow-hidden">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ClientIcon icon={Calendar} className="h-5 w-5 text-primary" />
            测试历史记录
            {sessions.length > 0 && (
              <Badge className="badge-filled ml-2">
                {sessions.length} 条记录
              </Badge>
            )}
          </DialogTitle>
        </DialogHeader>
        
        <div className="space-y-4 overflow-y-auto">
          {/* 统计信息和操作栏 */}
          <div className="flex justify-between items-center p-4 bg-card rounded-md-md">
            <div className="flex items-center gap-6 text-sm">
              <div className="flex items-center gap-2">
                <ClientIcon icon={BarChart3} className="h-4 w-4 text-primary" />
                <span className="text-muted-foreground">存储使用:</span>
                <span className="font-medium">
                  {storageStats?.sizeMB || 0} MB
                </span>
              </div>
              
              {selectedSessions.length > 0 && (
                <div className="flex items-center gap-2">
                  <Badge className="badge-outlined">
                    已选择 {selectedSessions.length} 项
                  </Badge>
                </div>
              )}
            </div>
            
            <div className="flex items-center gap-2">
              {selectedSessions.length > 0 && (
                <>
                  <Button
                    onClick={() => handleExport(selectedSessions, 'json')}
                    size="sm"
                    className="btn-outlined"
                  >
                    <ClientIcon icon={Download} className="h-3 w-3 mr-1" />
                    导出选中
                  </Button>
                  <Button
                    onClick={handleBatchDelete}
                    size="sm"
                    className="btn-outlined text-destructive"
                  >
                    <ClientIcon icon={Trash2} className="h-3 w-3 mr-1" />
                    删除选中
                  </Button>
                </>
              )}
              
              <Button
                onClick={() => handleExport(undefined, 'json')}
                size="sm"
                className="btn-outlined"
              >
                <ClientIcon icon={Download} className="h-3 w-3 mr-1" />
                导出全部
              </Button>
              
              <Button
                onClick={handleClearAll}
                size="sm"
                className="btn-outlined text-destructive"
                disabled={sessions.length === 0}
              >
                <ClientIcon icon={Trash2} className="h-3 w-3 mr-1" />
                清空全部
              </Button>
            </div>
          </div>
          
          {/* 会话列表 */}
          {sessions.length === 0 ? (
            <div className="text-center py-12">
              <ClientIcon icon={FileText} className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <div className="text-muted-foreground mb-2">暂无测试记录</div>
              <div className="text-sm text-muted-foreground">
                完成速度测试后，结果将自动保存在这里
              </div>
            </div>
          ) : (
            <>
              {/* 全选控制 */}
              <div className="flex items-center gap-2 px-2">
                <input
                  type="checkbox"
                  checked={selectedSessions.length === sessions.length && sessions.length > 0}
                  onChange={handleSelectAll}
                  className="rounded"
                />
                <span className="text-sm text-muted-foreground">
                  全选 ({sessions.length} 项)
                </span>
              </div>
              
              {/* 会话卡片列表 */}
              <div className="space-y-3">
                {sessions.map((session) => {
                  const successRate = session.successfulTests / session.totalProxies;
                  const isSelected = selectedSessions.includes(session.id);
                  
                  return (
                    <Card 
                      key={session.id} 
                      className={`card-elevated p-4 transition-all ${
                        isSelected ? 'ring-2 ring-primary bg-primary/5' : ''
                      }`}
                    >
                      <div className="flex items-start gap-3">
                        {/* 选择框 */}
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => handleSelectSession(session.id)}
                          className="mt-1 rounded"
                        />
                        
                        {/* 主要内容 */}
                        <div className="flex-1 space-y-3">
                          {/* 标题行 */}
                          <div className="flex items-start justify-between">
                            <div className="flex items-center gap-2 flex-1">
                              <ClientIcon icon={Globe} className="h-4 w-4 text-primary flex-shrink-0" />
                              <span className="font-medium truncate max-w-md" title={session.configPaths}>
                                {session.configPaths.split('/').pop() || '配置文件'}
                              </span>
                              <Badge className="badge-filled">
                                {session.totalProxies} 节点
                              </Badge>
                              <Badge className="badge-outlined">
                                {session.testType === 'speed' ? '测速' : 
                                 session.testType === 'unlock' ? '解锁' : 
                                 '测速+解锁'}
                              </Badge>
                            </div>
                            
                            <div className="text-xs text-muted-foreground flex-shrink-0">
                              {new Date(session.savedAt).toLocaleString('zh-CN')}
                            </div>
                          </div>
                          
                          {/* 统计信息网格 - 根据测试类型显示不同内容 */}
                          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                            {/* 成功率 - 始终显示 */}
                            <div className="flex items-center gap-2">
                              <div className={`font-medium ${getSuccessRateColor(successRate)}`}>
                                ✅ {session.successfulTests}
                              </div>
                              <div className="text-xs text-muted-foreground">
                                ❌ {session.failedTests}
                              </div>
                              <div className="text-xs text-muted-foreground">
                                ({(successRate * 100).toFixed(0)}%)
                              </div>
                            </div>
                            
                            {/* 速度测试相关信息 - 仅在speed或both模式显示 */}
                            {(session.testType === 'speed' || session.testType === 'both') && (
                              <>
                                {/* 平均速度 */}
                                <div className="flex items-center gap-2">
                                  <ClientIcon icon={Zap} className="h-3 w-3 text-success" />
                                  <span className="font-medium text-success">
                                    {formatSpeed(session.averageDownloadMbps)}
                                  </span>
                                  <span className="text-xs text-muted-foreground">平均</span>
                                </div>
                                
                                {/* 平均延迟 */}
                                <div className="flex items-center gap-2">
                                  <ClientIcon icon={Wifi} className="h-3 w-3 text-info" />
                                  <span className="font-medium text-info">
                                    {session.averageLatency.toFixed(0)}ms
                                  </span>
                                  <span className="text-xs text-muted-foreground">延迟</span>
                                </div>
                                
                                {/* 最佳节点速度 */}
                                <div className="flex items-center gap-2">
                                  <ClientIcon icon={Trophy} className="h-3 w-3 text-warning" />
                                  <span className="font-medium text-warning truncate max-w-24" title={session.bestProxy}>
                                    {formatSpeed(session.bestDownloadSpeedMbps)}
                                  </span>
                                  <span className="text-xs text-muted-foreground">最快</span>
                                </div>
                              </>
                            )}
                            
                            {/* 解锁测试相关信息 - 仅在unlock或both模式显示 */}
                            {(session.testType === 'unlock' || session.testType === 'both') && session.unlockStats && (
                              <>
                                {/* 解锁平台数 */}
                                <div className="flex items-center gap-2">
                                  <ClientIcon icon={Globe} className="h-3 w-3 text-primary" />
                                  <span className="font-medium text-primary">
                                    {session.unlockStats.total_platforms || 0}
                                  </span>
                                  <span className="text-xs text-muted-foreground">平台</span>
                                </div>
                                
                                {/* 解锁成功数 */}
                                <div className="flex items-center gap-2">
                                  <span className="font-medium text-success">
                                    ✓ {session.unlockStats.unlocked_count || 0}
                                  </span>
                                  <span className="text-xs text-muted-foreground">解锁</span>
                                </div>
                              </>
                            )}
                            
                            {/* 仅解锁模式时显示测试时长 */}
                            {session.testType === 'unlock' && (
                              <div className="flex items-center gap-2">
                                <span className="font-medium text-muted-foreground">
                                  ⏱️ {session.duration}
                                </span>
                              </div>
                            )}
                          </div>
                          
                          {/* 最佳节点和备注 */}
                          <div className="flex items-center justify-between text-xs text-muted-foreground">
                            <div className="flex items-center gap-4">
                              {/* 速度测试时显示最佳节点 */}
                              {session.bestProxy && (session.testType === 'speed' || session.testType === 'both') && (
                                <span>🏆 {session.bestProxy}</span>
                              )}
                              {/* 不同测试类型显示不同的时长标签 */}
                              {session.testType !== 'unlock' && (
                                <span>⏱️ {session.duration}</span>
                              )}
                            </div>
                            
                            {session.notes && (
                              <span className="text-foreground max-w-32 truncate" title={session.notes}>
                                📝 {session.notes}
                              </span>
                            )}
                          </div>
                        </div>
                        
                        {/* 操作按钮 */}
                        <div className="flex flex-col gap-1">
                          <Button
                            onClick={() => handleExport([session.id], 'json')}
                            size="sm"
                            className="btn-outlined"
                            title="导出JSON"
                          >
                            <ClientIcon icon={Download} className="h-3 w-3" />
                          </Button>
                          <Button
                            onClick={() => handleDelete(session.id)}
                            size="sm"
                            className="btn-outlined text-destructive"
                            title="删除记录"
                          >
                            <ClientIcon icon={Trash2} className="h-3 w-3" />
                          </Button>
                        </div>
                      </div>
                    </Card>
                  );
                })}
              </div>
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}